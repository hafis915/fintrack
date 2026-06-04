import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'
import { getDefaultCategoryIDs, seedBudgetPlan, seedTransaction, seedUser } from './helpers/seed'

/**
 * Home dashboard flow.
 *
 * The Phase-0 health card is gone. Home is now a real dashboard:
 *   - with a seeded budget plan → greeting + this-month snapshot + quick actions
 *   - with a token but NO plan → "Mulai onboarding" prompt
 *
 * We seed via SQL (like budget.spec.ts) rather than driving onboarding for
 * every test — that flow has its own coverage and DB seeding is faster +
 * deterministic for the dashboard UI under test here.
 *
 * The pure-API checks (/health envelope, /v1/me without a token) don't depend
 * on the old card and stay.
 */
async function authedSession(page: import('@playwright/test').Page): Promise<{ userId: string }> {
  resetOnboardingDB()
  const { token, userId } = mintToken()
  await page.addInitScript(
    ({ token, userId }) => {
      localStorage.setItem('fintrack.token', token)
      localStorage.setItem('fintrack.user_id', userId)
    },
    { token, userId },
  )
  return { userId }
}

test.describe('Home dashboard', () => {
  test('with a seeded plan shows greeting, snapshot, and quick actions', async ({ page }) => {
    const { userId } = await authedSession(page)

    seedUser(userId)
    const cats = getDefaultCategoryIDs()

    seedBudgetPlan({
      userId,
      income: 8_000_000,
      program: 'seimbang',
      items: [
        { categoryId: cats['Makan & minum'], allocated: 1_000_000, percentage: 12.5 },
        { categoryId: cats['Hiburan'], allocated: 500_000, percentage: 6.25 },
      ],
    })
    seedTransaction({ userId, categoryId: cats['Makan & minum'], amount: 400_000 })

    await page.goto('/')

    await expect(page.getByTestId('home-view')).toBeVisible()
    await expect(page.getByTestId('home-greeting')).toBeVisible()

    // This-month snapshot renders with the seeded numbers.
    const snapshot = page.getByTestId('home-snapshot')
    await expect(snapshot).toBeVisible()
    await expect(snapshot).toContainText('Rp 8.000.000')
    await expect(snapshot).toContainText('Rp 400.000')

    // No-plan prompt must NOT be present when a plan exists.
    await expect(page.getByTestId('home-no-plan')).toHaveCount(0)

    // Quick-action buttons are all present.
    await expect(page.getByTestId('home-cta-scan')).toBeVisible()
    await expect(page.getByTestId('home-cta-transactions')).toBeVisible()
    await expect(page.getByTestId('home-cta-budget')).toBeVisible()
  })

  test('quick-action CTA navigates to its route', async ({ page }) => {
    const { userId } = await authedSession(page)

    seedUser(userId)
    const cats = getDefaultCategoryIDs()
    seedBudgetPlan({
      userId,
      income: 8_000_000,
      program: 'seimbang',
      items: [{ categoryId: cats['Hiburan'], allocated: 500_000, percentage: 6.25 }],
    })

    await page.goto('/')
    await expect(page.getByTestId('home-cta-transactions')).toBeVisible()

    await page.getByTestId('home-cta-transactions').click()
    await expect(page).toHaveURL(/\/transactions$/)
    await expect(page.getByTestId('transactions-view')).toBeVisible()
  })

  test('with a token but no plan shows the onboarding prompt', async ({ page }) => {
    await authedSession(page)

    await page.goto('/')

    await expect(page.getByTestId('home-view')).toBeVisible()
    await expect(page.getByTestId('home-greeting')).toBeVisible()

    const prompt = page.getByTestId('home-no-plan')
    await expect(prompt).toBeVisible()
    await expect(prompt).toContainText('Mulai onboarding')

    // The snapshot belongs to the has-plan branch and must be absent.
    await expect(page.getByTestId('home-snapshot')).toHaveCount(0)
  })

  test('serves the /health JSON envelope directly from the API', async ({ request }) => {
    const res = await request.get('http://127.0.0.1:8088/health')
    expect(res.status()).toBe(200)
    const body = await res.json()
    expect(body).toMatchObject({
      data: { status: 'ok', db: 'ok', version: '0.0.0' },
    })
    expect(res.headers()['x-request-id']).toBeTruthy()
  })

  test('rejects /v1/me without a token via the proxy', async ({ request }) => {
    const res = await request.get('http://127.0.0.1:5173/v1/me')
    expect(res.status()).toBe(401)
    const body = await res.json()
    expect(body.error?.code).toBe('missing_token')
  })
})
