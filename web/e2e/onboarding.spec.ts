import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'

/**
 * Phase 1.5 onboarding — the goal-first FINANCIAL PLANNER flow.
 *
 * Onboarding is no longer a "fill every expense" form. It's a 3-step planner:
 *   STEP 1  six intake questions (income, housing, goal, debt, EF, lifestyle)
 *   STEP 2  the user types ONLY their FIXED expenses (rent, debt, utilities)
 *   STEP 3  the app DETERMINISTICALLY suggests amounts for the FLEXIBLE
 *           (keinginan) categories, shows a savings target + summary, and lets
 *           the user refine — either by inline-editing a number or by chatting
 *           with the planner in natural language. The LLM only reads intent +
 *           narrates; the actual numbers are re-balanced by app code.
 *
 * In ENV=test the API selects the deterministic STUB LLM (no OPEN_ROUTER_API_KEY),
 * so "naikin makan jadi 1500000" reliably sets Makan & minum to Rp 1.500.000.
 *
 * Confirm reuses the EXISTING POST /v1/onboarding finalize: at submit the
 * frontend sends fixed + flexible items, so income - sum(items) == savings.
 */
test.describe('Onboarding planner flow', () => {
  test.beforeEach(async ({ page }) => {
    resetOnboardingDB()
    const { token, userId } = mintToken()
    // Drop the token into localStorage BEFORE any page script runs so the axios
    // interceptor sees it on the very first request.
    await page.addInitScript(
      ({ token, userId }) => {
        localStorage.setItem('fintrack.token', token)
        localStorage.setItem('fintrack.user_id', userId)
      },
      { token, userId },
    )
  })

  // Helper: drive step 1 → step 2 → step 3 with a sensible answer set, leaving
  // the wizard parked on step 3 with a generated plan + planner chat ready.
  async function reachPlanStep(page: import('@playwright/test').Page) {
    await page.goto('/onboarding')
    await expect(page.getByTestId('onboarding-view')).toBeVisible()
    await expect(page.getByTestId('onb-step-1')).toBeVisible()

    // STEP 1 — intake questions.
    await page.getByTestId('onb-income').fill('8000000')
    await page.getByTestId('onb-goal').selectOption('balance')
    await page.getByTestId('onb-lifestyle-balanced').click()
    await page.getByTestId('onb-next').click()

    // STEP 2 — fixed expenses. Wait for the catalog to load, then set a couple.
    await expect(page.getByTestId('onb-step-2')).toBeVisible()
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })
    await page.getByTestId('onb-fixed-amount-Sewa kosan').fill('1200000')
    await page.getByTestId('onb-fixed-amount-Internet & pulsa').fill('300000')
    await page.getByTestId('onb-next').click()

    // STEP 3 — the deterministic plan renders.
    await expect(page.getByTestId('onb-step-3')).toBeVisible()
    await expect(page.getByText('menyusun rencana…')).toBeHidden({ timeout: 5_000 })
  }

  test('happy path: suggest flexible amounts, refine via planner chat, submit', async ({
    page,
  }) => {
    await reachPlanStep(page)

    // The app suggested non-zero amounts for the flexible (keinginan) cats.
    const makanInput = page.getByTestId('onb-flex-amount-Makan & minum')
    await expect(makanInput).toBeVisible()
    await expect(makanInput).not.toHaveValue('')
    await expect(makanInput).not.toHaveValue('0')

    // A savings target + plan summary is shown.
    await expect(page.getByTestId('onb-plan-summary')).toBeVisible()
    await expect(page.getByTestId('onb-savings-target')).toBeVisible()

    // Refine via the planner chat. The stub LLM parses the last user message for
    // a rupiah number + a flexible-category substring ("makan"), and the API
    // re-balances Makan & minum to exactly 1.500.000 deterministically.
    await page.getByTestId('planner-input').fill('naikin makan jadi 1500000')
    await page.getByTestId('planner-send').click()

    // The assistant reply lands in the transcript...
    await expect(
      page.getByTestId('planner-message').filter({ hasText: /makan/i }).last(),
    ).toBeVisible({ timeout: 10_000 })

    // ...and the deterministic re-balance reflects in the editable input.
    await expect(makanInput).toHaveValue('1.500.000')

    // Confirm → finalize via the existing POST /v1/onboarding, land on result.
    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)
    await expect(page.getByTestId('onboarding-result')).toBeVisible()
    await expect(page.getByTestId('result-program')).toBeVisible()
    // The refined Makan & minum allocation persisted into the plan.
    await expect(page.getByTestId('result-items')).toContainText('Makan & minum')
    await expect(page.getByTestId('result-items')).toContainText('Rp 1.500.000')
  })

  test('inline-editing a flexible amount carries into the finalized plan', async ({ page }) => {
    await reachPlanStep(page)

    // Edit a flexible category directly (no chat) — the wizard owns the number.
    const hiburan = page.getByTestId('onb-flex-amount-Hiburan')
    await expect(hiburan).toBeVisible()
    await hiburan.fill('250000')
    await expect(hiburan).toHaveValue('250.000')

    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)
    await expect(page.getByTestId('result-items')).toContainText('Hiburan')
    await expect(page.getByTestId('result-items')).toContainText('Rp 250.000')
  })

  test('result page continues to the live budget dashboard', async ({ page }) => {
    await reachPlanStep(page)
    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)

    await page.getByTestId('result-continue').click()
    await expect(page).toHaveURL(/\/budget$/)
    await expect(page.getByTestId('budget-view')).toBeVisible()
    // The plan just created renders (not the no-plan CTA).
    await expect(page.getByTestId('budget-summary')).toBeVisible({ timeout: 5_000 })
  })

  // "Ubah jawaban" must restore the user's previous answers (income/goal/fixed)
  // instead of resetting the wizard to defaults — the store persists them.
  test('Ubah jawaban restores previous answers', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByTestId('onb-step-1')).toBeVisible()

    await page.getByTestId('onb-income').fill('5500000')
    await page.getByTestId('onb-goal').selectOption('invest')
    await page.getByTestId('onb-next').click()

    // Step 2: set a fixed amount so we can assert it survives the round-trip.
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })
    await page.getByTestId('onb-fixed-amount-Sewa kosan').fill('1000000')
    await page.getByTestId('onb-next').click()

    await expect(page.getByTestId('onb-step-3')).toBeVisible()
    await expect(page.getByText('menyusun rencana…')).toBeHidden({ timeout: 5_000 })
    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)

    // Go back to edit answers.
    await page.getByTestId('result-restart').click()
    await expect(page).toHaveURL(/\/onboarding$/)
    await expect(page.getByTestId('onb-step-1')).toBeVisible()

    // Step 1 answers are preserved, not reset to the 8.000.000 / debt defaults.
    await expect(page.getByTestId('onb-income')).toHaveValue('5.500.000')
    await expect(page.getByTestId('onb-goal')).toHaveValue('invest')

    // Step 2 fixed amount is preserved too.
    await page.getByTestId('onb-next').click()
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })
    await expect(page.getByTestId('onb-fixed-amount-Sewa kosan')).toHaveValue('1.000.000')
  })

  // Step 1 income guard: empty/too-low income is caught inline before any API
  // call; the wizard stays on step 1.
  test('empty income is caught inline on step 1', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByTestId('onb-step-1')).toBeVisible()

    await page.getByTestId('onb-income').fill('')
    await page.getByTestId('onb-next').click()

    await expect(page.getByTestId('onb-error')).toBeVisible()
    await expect(page.getByTestId('onb-error')).toContainText(/pemasukan/i)
    await expect(page.getByTestId('onb-step-1')).toBeVisible()
    await expect(page).toHaveURL(/\/onboarding$/)
  })

  // The planner chat is multi-turn: an ambiguous message (no category/number)
  // gets a clarifying reply and changes nothing.
  test('planner asks for clarification when the message is ambiguous', async ({ page }) => {
    await reachPlanStep(page)

    const makanInput = page.getByTestId('onb-flex-amount-Makan & minum')
    const before = await makanInput.inputValue()

    await page.getByTestId('planner-input').fill('halo planner')
    await page.getByTestId('planner-send').click()

    // The stub returns the clarifying prompt and no adjustments.
    await expect(
      page.getByTestId('planner-message').filter({ hasText: /kategori mana/i }).last(),
    ).toBeVisible({ timeout: 10_000 })

    // Nothing changed.
    await expect(makanInput).toHaveValue(before)
  })

  // Step 2 lets the user add a fixed expense the catalog doesn't cover; it must
  // flow into the plan and persist into the finalized budget.
  test('add a custom fixed expense in step 2 and it carries into the plan', async ({ page }) => {
    await page.goto('/onboarding')
    await expect(page.getByTestId('onb-step-1')).toBeVisible()
    await page.getByTestId('onb-income').fill('8000000')
    await page.getByTestId('onb-next').click()

    await expect(page.getByTestId('onb-step-2')).toBeVisible()
    await expect(page.getByText('memuat kategori…')).toBeHidden({ timeout: 5_000 })

    // Create a custom fixed expense (onboarding step 2 has no type picker — it's
    // fixed by definition).
    await page.getByTestId('add-category-toggle').click()
    await page.getByTestId('add-category-name').fill('Cicilan motor')
    await page.getByTestId('add-category-submit').click()

    // It appears as an editable fixed row; give it an amount.
    const motor = page.getByTestId('onb-fixed-amount-Cicilan motor')
    await expect(motor).toBeVisible()
    await motor.fill('850000')

    await page.getByTestId('onb-next').click()
    await expect(page.getByTestId('onb-step-3')).toBeVisible()
    await expect(page.getByText('menyusun rencana…')).toBeHidden({ timeout: 5_000 })

    // Finalize and confirm the custom fixed expense persisted into the plan.
    await page.getByTestId('onb-submit').click()
    await expect(page).toHaveURL(/\/onboarding\/result$/)
    await expect(page.getByTestId('result-items')).toContainText('Cicilan motor')
    await expect(page.getByTestId('result-items')).toContainText('Rp 850.000')
  })
})
