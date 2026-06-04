import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'
import { getDefaultCategoryIDs, seedBudgetPlan, seedTransaction, seedUser } from './helpers/seed'

/**
 * Phase 3 fatigue dashboard flow.
 *
 * Two scenarios:
 *   - no plan yet → empty-state CTA back to onboarding
 *   - plan + transactions → fatigue badges (fresh/warning/fatigued)
 *
 * We pre-seed via SQL rather than driving the onboarding wizard for
 * every test — onboarding has its own e2e coverage and seed-via-DB is
 * faster + more deterministic for the fatigue UI under test here.
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

test.describe('Budget / fatigue dashboard', () => {
  test('shows no-plan CTA when user has no budget for this month', async ({ page }) => {
    await authedSession(page)
    await page.goto('/budget')
    await expect(page.getByTestId('budget-view')).toBeVisible()
    await expect(page.getByTestId('budget-no-plan')).toBeVisible()
    await expect(page.getByTestId('budget-no-plan')).toContainText('Selesaikan onboarding')
  })

  test('renders fresh / warning / fatigued indicators per category', async ({ page }) => {
    const { userId } = await authedSession(page)

    seedUser(userId)
    const cats = getDefaultCategoryIDs()

    seedBudgetPlan({
      userId,
      income: 8_000_000,
      program: 'seimbang',
      items: [
        { categoryId: cats['Makan & minum'], allocated: 1_000_000, percentage: 12.5 }, // warning
        { categoryId: cats['Hiburan'], allocated: 500_000, percentage: 6.25 },         // fresh
        { categoryId: cats['Kartu kredit'], allocated: 400_000, percentage: 5 },       // fatigued
      ],
    })

    // 80% of 1M → warning
    seedTransaction({ userId, categoryId: cats['Makan & minum'], amount: 800_000 })
    // 20% of 500k → fresh
    seedTransaction({ userId, categoryId: cats['Hiburan'], amount: 100_000 })
    // 125% of 400k → fatigued
    seedTransaction({ userId, categoryId: cats['Kartu kredit'], amount: 500_000 })

    await page.goto('/budget')

    await expect(page.getByTestId('budget-view')).toBeVisible()
    await expect(page.getByTestId('budget-period')).toContainText('Program Seimbang')

    // Summary totals.
    await expect(page.getByTestId('summary-spent')).toContainText('Rp 1.400.000')
    await expect(page.getByTestId('summary-spent')).toContainText('17.5%')

    // Per-item badges.
    await expect(page.getByTestId('budget-item-Makan & minum-status')).toHaveText('warning')
    await expect(page.getByTestId('budget-item-Hiburan-status')).toHaveText('fresh')
    await expect(page.getByTestId('budget-item-Kartu kredit-status')).toHaveText('fatigued')

    // Fatigued item shown first (sorted by stress).
    const items = page.getByTestId('budget-items').locator('article')
    await expect(items.first()).toContainText('Kartu kredit')
  })

  test('over-budget category surfaces a recommendation + the compare chart', async ({ page }) => {
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

    // Makan & minum is driven OVER budget (1.3M spent vs 1M allocated).
    seedTransaction({ userId, categoryId: cats['Makan & minum'], amount: 1_300_000 })
    // Hiburan stays comfortably under.
    seedTransaction({ userId, categoryId: cats['Hiburan'], amount: 100_000 })

    await page.goto('/budget')
    await expect(page.getByTestId('budget-view')).toBeVisible()

    // Recommendations block renders with a reco-item for the over-budget category.
    const recos = page.getByTestId('budget-recommendations')
    await expect(recos).toBeVisible()
    await expect(page.getByTestId('reco-item-Makan & minum')).toBeVisible()
    // The over-budget reco quantifies the overspend (1.3M - 1M = 300k).
    await expect(page.getByTestId('reco-item-Makan & minum')).toContainText('Rp 300.000')

    // The budget-vs-realisasi chart is visible with a row for the category.
    await expect(page.getByTestId('budget-compare-chart')).toBeVisible()
    await expect(page.getByTestId('compare-row-Makan & minum')).toBeVisible()
  })

  // An already-onboarded user can re-run the planner from the budget dashboard.
  test('re-budget button sends an onboarded user back into the planner', async ({ page }) => {
    const { userId } = await authedSession(page)

    seedUser(userId)
    const cats = getDefaultCategoryIDs()
    seedBudgetPlan({
      userId,
      income: 8_000_000,
      program: 'seimbang',
      items: [{ categoryId: cats['Makan & minum'], allocated: 1_000_000, percentage: 12.5 }],
    })

    await page.goto('/budget')
    await expect(page.getByTestId('budget-summary')).toBeVisible()

    // The re-budget affordance is only shown once a plan exists.
    const rebudget = page.getByTestId('budget-rebudget')
    await expect(rebudget).toBeVisible()
    await rebudget.click()

    // Lands back on step 1 of the onboarding planner.
    await expect(page).toHaveURL(/\/onboarding$/)
    await expect(page.getByTestId('onb-step-1')).toBeVisible()
  })
})
