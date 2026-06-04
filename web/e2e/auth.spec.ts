import { test, expect, type Page } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'

/**
 * Phase 0 local-first auth flow (ADR-014).
 *
 * Covers the route guard, the register/login UI that mints the HS256 JWT the
 * middleware validates, and the regression that the category dropdown only
 * populates once a real token is present (the "auth-cascade dropdown" bug —
 * /v1/categories is auth-protected, so an unauthenticated load left the
 * <select> empty). Also pins bug #3 (cleared numeric field → 400) on both the
 * onboarding income field and the transaction edit-amount field.
 *
 * The DB is reset before each test; register goes through the real backend so
 * the `users` row exists for the subsequent login.
 */

const TOKEN_KEY = 'fintrack.token'
const USER_ID_KEY = 'fintrack.user_id'

// Unique email per invocation so reruns / parallel-ish flows never collide on
// the unique email constraint. resetOnboardingDB clears users anyway, but this
// keeps within-run register+login pairs unambiguous.
function uniqueEmail(): string {
  return `e2e-${Date.now()}-${Math.floor(Math.random() * 1e6)}@local.test`
}

// base64url-encode a string without padding (JWT segment format).
function base64Url(input: string): string {
  return Buffer.from(input, 'utf8')
    .toString('base64')
    .replace(/\+/g, '-')
    .replace(/\//g, '_')
    .replace(/=+$/, '')
}

// Craft a structurally-valid-but-expired JWT. The signature is a placeholder —
// the client guard only decodes the payload's exp, so it doesn't verify it.
function craftExpiredJWT(): string {
  const header = base64Url(JSON.stringify({ alg: 'HS256', typ: 'JWT' }))
  // exp far in the past (2001-09-09), so the client clock always sees it expired.
  const payload = base64Url(
    JSON.stringify({ sub: '00000000-0000-4000-8000-000000000000', exp: 1_000_000_000 }),
  )
  return `${header}.${payload}.not-a-real-signature`
}

// Seed a token into localStorage BEFORE any page script runs so the axios
// interceptor + router guard both see it on the very first navigation.
async function seedToken(page: Page, token: string, userId: string) {
  await page.addInitScript(
    ({ token, userId, TOKEN_KEY, USER_ID_KEY }) => {
      localStorage.setItem(TOKEN_KEY, token)
      localStorage.setItem(USER_ID_KEY, userId)
    },
    { token, userId, TOKEN_KEY, USER_ID_KEY },
  )
}

test.describe('Auth flow', () => {
  test.beforeEach(() => {
    resetOnboardingDB()
  })

  test('unauthenticated guard bounces a protected route to /login', async ({ page }) => {
    // No token seeded. Hitting a protected route must redirect to /login with
    // the intended destination preserved as ?redirect=.
    await page.goto('/transactions')

    await expect(page).toHaveURL(/\/login\?redirect=/)
    await expect(page.getByTestId('login-view')).toBeVisible()
    await expect(page.getByTestId('login-email')).toBeVisible()
    await expect(page.getByTestId('login-submit')).toBeVisible()
  })

  test('an expired token is treated as logged-out and bounces to /login', async ({ page }) => {
    // The router guard decodes the JWT exp client-side (isAuthenticated), so an
    // expired token counts as logged-out — the signature never needs to be
    // valid for this redirect. Craft a 3-part token whose middle segment is
    // base64url-encoded JSON with an exp far in the past.
    const expiredToken = craftExpiredJWT()
    await seedToken(page, expiredToken, '00000000-0000-4000-8000-000000000000')

    await page.goto('/transactions')

    // Guard bounces to /login, preserving the intended destination.
    await expect(page).toHaveURL(/\/login\?redirect=/)
    await expect(page.getByTestId('login-view')).toBeVisible()
  })

  test('register lands on /onboarding and stores a token', async ({ page }) => {
    await page.goto('/register')
    await expect(page.getByTestId('register-view')).toBeVisible()

    const email = uniqueEmail()
    await page.getByTestId('register-name').fill('Hafis E2E')
    await page.getByTestId('register-email').fill(email)
    await page.getByTestId('register-password').fill('supersecret123')
    await page.getByTestId('register-submit').click()

    // New users are dropped straight into onboarding.
    await expect(page).toHaveURL(/\/onboarding$/)
    await expect(page.getByTestId('onboarding-view')).toBeVisible()

    // Token is now persisted — proves the JWT was minted + stored.
    const token = await page.evaluate((k) => localStorage.getItem(k), TOKEN_KEY)
    expect(token, 'token should be in localStorage after register').toBeTruthy()
    const userId = await page.evaluate((k) => localStorage.getItem(k), USER_ID_KEY)
    expect(userId, 'user_id should be in localStorage after register').toBeTruthy()
  })

  test('login authenticates and the category dropdown populates (auth-cascade bug)', async ({
    page,
  }) => {
    // 1. Register through the UI so a real users row exists to log into.
    await page.goto('/register')
    const email = uniqueEmail()
    await page.getByTestId('register-name').fill('Login E2E')
    await page.getByTestId('register-email').fill(email)
    await page.getByTestId('register-password').fill('supersecret123')
    await page.getByTestId('register-submit').click()
    await expect(page).toHaveURL(/\/onboarding$/)

    // 2. Clear the session so we exercise the real login path, not the
    //    register-minted token.
    await page.evaluate(
      ({ TOKEN_KEY, USER_ID_KEY }) => {
        localStorage.removeItem(TOKEN_KEY)
        localStorage.removeItem(USER_ID_KEY)
      },
      { TOKEN_KEY, USER_ID_KEY },
    )

    // 3. Log in with that email + password → lands on home.
    await page.goto('/login')
    await expect(page.getByTestId('login-view')).toBeVisible()
    await page.getByTestId('login-email').fill(email)
    await page.getByTestId('login-password').fill('supersecret123')
    await page.getByTestId('login-submit').click()

    await expect(page).toHaveURL(/\/$/)
    // Home is now a dashboard (the Phase-0 health card is gone). With a fresh
    // login + no plan yet, home renders the greeting + onboarding prompt.
    await expect(page.getByTestId('home-view')).toBeVisible()
    await expect(page.getByTestId('home-greeting')).toBeVisible()

    // Token re-minted by login.
    const token = await page.evaluate((k) => localStorage.getItem(k), TOKEN_KEY)
    expect(token, 'token should be in localStorage after login').toBeTruthy()

    // 4. The cascade: with a valid token, /v1/categories returns data, so the
    //    transactions <select> must be populated (not empty). This is the bug
    //    regression — before the auth fix the dropdown stayed empty.
    await page.goto('/transactions')
    await expect(page.getByTestId('transactions-view')).toBeVisible()

    const select = page.getByTestId('tx-new-category')
    await expect(select.locator('option').first()).toBeAttached({ timeout: 5_000 })
    const optionCount = await select.locator('option').count()
    expect(optionCount, 'category dropdown should be populated').toBeGreaterThan(0)
  })

  test('login with unknown credentials shows a generic inline error', async ({ page }) => {
    await page.goto('/login')
    await expect(page.getByTestId('login-view')).toBeVisible()

    // No such user (DB was reset) → backend returns a GENERIC 401 (it must not
    // reveal whether the email exists) → inline copy, no crash.
    await page.getByTestId('login-email').fill(uniqueEmail())
    await page.getByTestId('login-password').fill('supersecret123')
    await page.getByTestId('login-submit').click()

    await expect(page.getByTestId('login-error')).toBeVisible()
    await expect(page.getByTestId('login-error')).toContainText(/salah/i)

    // Stayed on /login, no token written, no navigation.
    await expect(page).toHaveURL(/\/login$/)
    const token = await page.evaluate((k) => localStorage.getItem(k), TOKEN_KEY)
    expect(token, 'no token should be stored on a failed login').toBeNull()
  })

  test('logout clears the session and returns to /login', async ({ page }) => {
    // Register → authenticated, then visit the home dashboard where Keluar lives.
    await page.goto('/register')
    const email = uniqueEmail()
    await page.getByTestId('register-name').fill('Logout E2E')
    await page.getByTestId('register-email').fill(email)
    await page.getByTestId('register-password').fill('supersecret123')
    await page.getByTestId('register-submit').click()
    await expect(page).toHaveURL(/\/onboarding$/)

    await page.goto('/')
    await expect(page.getByTestId('home-view')).toBeVisible()
    await page.getByTestId('home-logout').click()

    // Back to /login with the session cleared.
    await expect(page).toHaveURL(/\/login$/)
    const token = await page.evaluate((k) => localStorage.getItem(k), TOKEN_KEY)
    expect(token, 'token should be cleared on logout').toBeNull()
  })

  test('desktop: logout lives in the sidebar navbar and clears the session', async ({ page }) => {
    // On a wide viewport the bottom nav is replaced by the left sidebar, which
    // is where Keluar now lives (the home-header button is mobile-only).
    await page.setViewportSize({ width: 1280, height: 800 })

    await page.goto('/register')
    const email = uniqueEmail()
    await page.getByTestId('register-name').fill('Desktop Logout')
    await page.getByTestId('register-email').fill(email)
    await page.getByTestId('register-password').fill('supersecret123')
    await page.getByTestId('register-submit').click()
    await expect(page).toHaveURL(/\/onboarding$/)

    await page.goto('/')
    await expect(page.getByTestId('home-view')).toBeVisible()

    // The sidebar logout is visible on desktop; the header one is hidden.
    await expect(page.getByTestId('sidebar-logout')).toBeVisible()
    await expect(page.getByTestId('home-logout')).toBeHidden()
    await page.getByTestId('sidebar-logout').click()

    await expect(page).toHaveURL(/\/login$/)
    const token = await page.evaluate((k) => localStorage.getItem(k), TOKEN_KEY)
    expect(token, 'token should be cleared on logout').toBeNull()
  })
})

/**
 * Bug #3 regressions: a cleared numeric field used to serialize as "" and the
 * backend rejected the request as invalid_json (HTTP 400). Both views now guard
 * client-side and surface an inline message with NO network call. We assert the
 * inline message AND that no 4xx response left the wire.
 */
test.describe('Bug #3 — cleared numeric field guards', () => {
  test.beforeEach(async ({ page }) => {
    resetOnboardingDB()
    const { token, userId } = mintToken()
    await seedToken(page, token, userId)
  })

  test('onboarding: empty income is caught inline, no 400, no navigation', async ({ page }) => {
    // Track any 4xx from the API while we interact.
    const badResponses: string[] = []
    page.on('response', (resp) => {
      if (resp.url().includes('/v1/') && resp.status() >= 400) {
        badResponses.push(`${resp.status()} ${resp.url()}`)
      }
    })

    await page.goto('/onboarding')
    await expect(page.getByTestId('onb-step-1')).toBeVisible()

    // Step 1: clear the income field entirely, then try to advance. The client
    // guard catches the empty/zero income before any API call.
    await page.getByTestId('onb-income').fill('')
    await page.getByTestId('onb-next').click()

    // Inline validation appears, we stay on step 1 / /onboarding.
    await expect(page.getByTestId('onb-error')).toBeVisible()
    await expect(page.getByTestId('onb-step-1')).toBeVisible()
    await expect(page).toHaveURL(/\/onboarding$/)

    // The guard short-circuits before any suggest/submit request, so no 400 hit the API.
    expect(
      badResponses,
      `unexpected API errors: ${badResponses.join(', ')}`,
    ).toEqual([])
  })

  test('transactions: editing to an empty amount is caught inline, no 400', async ({ page }) => {
    const badResponses: string[] = []
    page.on('response', (resp) => {
      if (
        resp.url().includes('/v1/transactions') &&
        resp.request().method() !== 'GET' &&
        resp.status() >= 400
      ) {
        badResponses.push(`${resp.status()} ${resp.url()}`)
      }
    })

    await page.goto('/transactions')
    await expect(page.getByTestId('transactions-view')).toBeVisible()

    // Create a transaction to edit.
    const select = page.getByTestId('tx-new-category')
    await expect(select.locator('option').first()).toBeAttached({ timeout: 5_000 })
    await page.getByTestId('tx-new-amount').fill('25000')
    await page.getByTestId('tx-new-note').fill('coffee')
    await page.getByTestId('tx-new-submit').click()

    const list = page.getByTestId('tx-list')
    await expect(list.locator('li')).toHaveCount(1)
    const row = list.locator('li').first()

    // Enter edit mode, clear the amount, save.
    await row.getByText('edit').click()
    await page.getByTestId('tx-edit-amount').fill('')
    await page.getByTestId('tx-edit-save').click()

    // Inline validation surfaces. The guard returns before clearing edit mode,
    // so the row stays in the editor (the amount input is still on screen) and
    // never committed the empty value.
    await expect(page.getByTestId('tx-error')).toBeVisible()
    await expect(page.getByTestId('tx-edit-amount')).toBeVisible()

    // No PATCH/PUT 400 should have hit the API — the guard caught it first.
    expect(
      badResponses,
      `unexpected transaction write errors: ${badResponses.join(', ')}`,
    ).toEqual([])
  })
})
