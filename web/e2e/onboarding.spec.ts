import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'

/**
 * Phase 1 onboarding flow.
 *
 * A returning user with a fresh JWT lands on /onboarding, accepts the
 * pre-filled defaults (or tweaks them), submits, and is shown the
 * generated Plan on /onboarding/result.
 *
 * The defaults in OnboardingView pre-select a Bebas Utang answer set
 * (cc debt + emergency_months=1 + goal=debt) so the flow has a meaningful
 * outcome with zero clicks.
 */
test.describe('Onboarding flow', () => {
  test.beforeEach(async ({ page }) => {
    resetOnboardingDB()
    const { token, userId } = mintToken()
    // Drop the token into localStorage BEFORE any page script runs so
    // the axios interceptor sees it on first request.
    await page.addInitScript(
      ({ token, userId }) => {
        localStorage.setItem('fintrack.token', token)
        localStorage.setItem('fintrack.user_id', userId)
      },
      { token, userId },
    )
  })

  test('accept defaults → land on result with Bebas Utang plan', async ({ page }) => {
    await page.goto('/onboarding')

    await expect(page.getByTestId('onboarding-view')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Mulai program-mu' })).toBeVisible()

    // Wait for categories to load (loading copy disappears).
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    // Submit unchanged — defaults already match Bebas Utang.
    await page.getByTestId('onb-submit').click()

    await expect(page).toHaveURL(/\/onboarding\/result$/)
    await expect(page.getByTestId('onboarding-result')).toBeVisible()
    await expect(page.getByTestId('result-program')).toHaveText(/Program Bebas Utang/)

    // Summary buckets all present with sensible values.
    await expect(page.getByTestId('result-bucket-kebutuhan')).toContainText('Rp 2.700.000')
    await expect(page.getByTestId('result-bucket-utang')).toContainText('Rp 400.000')
    await expect(page.getByTestId('result-bucket-keinginan')).toContainText('Rp 500.000')
    await expect(page.getByTestId('result-bucket-tabungan')).toContainText('Rp 4.400.000')

    // Items list rendered.
    const items = page.getByTestId('result-items').locator('li')
    await expect(items).toHaveCount(4)
  })

  test('switching goal to invest with full EF yields Tumbuh plan', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    // Flip goal → invest, EF → 6, drop the cc debt selection + the cc item.
    await page.getByTestId('onb-goal').selectOption('invest')
    // emergency-6 is a sr-only radio inside a label, so click via force.
    await page.getByTestId('onb-emergency-6').check({ force: true })
    // Toggle "none" on, then "cc" off so debt_types = ['none'].
    await page.getByTestId('onb-debt-none').click()
    await page.getByTestId('onb-debt-cc').click()
    // Uncheck the kartu kredit row so the request body has no debt items.
    await page.getByTestId('onb-item-enable-Kartu kredit').uncheck()

    await page.getByTestId('onb-submit').click()
    await expect(page.getByTestId('result-program')).toHaveText(/Program Tumbuh/)
    await expect(page.getByTestId('result-bucket-utang')).toContainText('Rp 0')
  })

  // Bug #1: "Ubah jawaban" must restore the user's previous answers instead
  // of resetting the form to defaults.
  test('Ubah jawaban restores previous answers', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    await page.getByTestId('onb-income').fill('5500000')
    await page.getByTestId('onb-goal').selectOption('invest')
    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)

    // Go back to edit answers.
    await page.getByTestId('result-restart').click()
    await expect(page).toHaveURL(/\/onboarding$/)
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    // Answers are preserved, not reset to the 8.000.000 / debt defaults.
    await expect(page.getByTestId('onb-income')).toHaveValue('5500000')
    await expect(page.getByTestId('onb-goal')).toHaveValue('invest')
  })

  // Bug #2: an overspending budget shows a coaching chip but still submits
  // (no disabled/blocked "next").
  test('overspending shows a coaching chip and still submits', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    // Push a want item far above income (8.000.000 default) → total overspends.
    await page.getByTestId('onb-item-amount-Hiburan').fill('9000000')

    await expect(page.getByTestId('onb-overspend')).toBeVisible()
    await expect(page.getByTestId('onb-overspend')).toContainText(/melebihi pemasukan/i)
    // Submit is NOT disabled by overspending.
    await expect(page.getByTestId('onb-submit')).toBeEnabled()

    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)
    await expect(page.getByTestId('result-warning')).toBeVisible()
  })

  // New: a user can add an expense the default seed doesn't cover, and it
  // flows into the generated plan.
  test('can add a custom expense not in the defaults', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    await page.getByTestId('onb-add-name').fill('Kursus online')
    await page.getByTestId('onb-add-type').selectOption('want')
    await page.getByTestId('onb-add-amount').fill('300000')
    await page.getByTestId('onb-add-submit').click()

    // The new item appears as a normal, enabled expense row.
    await expect(page.getByTestId('onb-item-enable-Kursus online')).toBeVisible()

    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)
    await expect(page.getByTestId('result-items')).toContainText('Kursus online')
  })

  // The result page must finish onboarding, not dead-end. "Lanjut" takes the
  // user to their live budget dashboard.
  test('result page continues to the budget dashboard', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })
    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)

    await page.getByTestId('result-continue').click()
    await expect(page).toHaveURL(/\/budget$/)
    await expect(page.getByTestId('budget-view')).toBeVisible()
    // The plan just created renders (not the no-plan CTA).
    await expect(page.getByTestId('budget-summary')).toBeVisible({ timeout: 5_000 })
  })

  test('surfaces backend validation errors inline', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    // Income below the floor is now caught by the client guard (bug #3) before
    // it reaches the API, surfacing a friendly inline message and keeping us on
    // /onboarding with no navigation.
    await page.getByTestId('onb-income').fill('0')
    await page.getByTestId('onb-submit').click()

    await expect(page.getByTestId('onb-error')).toBeVisible()
    await expect(page.getByTestId('onb-error')).toContainText(/pemasukan/i)
    await expect(page).toHaveURL(/\/onboarding$/)
  })
})
