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

  test('surfaces backend validation errors inline', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    // Income = 0 fails server-side validation.
    await page.getByTestId('onb-income').fill('0')
    await page.getByTestId('onb-submit').click()

    await expect(page.getByTestId('onb-error')).toBeVisible()
    await expect(page.getByTestId('onb-error')).toContainText(/income/i)
    await expect(page).toHaveURL(/\/onboarding$/)
  })
})
