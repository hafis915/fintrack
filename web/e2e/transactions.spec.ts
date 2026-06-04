import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'

/**
 * Phase 2 transactions flow.
 *
 * Empty state → create → see in list → edit amount → delete → empty state.
 * The user is freshly minted; no onboarding required (budget_plan auto-link
 * is nullable, so transactions work standalone).
 */
test.describe('Transactions flow', () => {
  test.beforeEach(async ({ page }) => {
    resetOnboardingDB()
    const { token, userId } = mintToken()
    await page.addInitScript(
      ({ token, userId }) => {
        localStorage.setItem('fintrack.token', token)
        localStorage.setItem('fintrack.user_id', userId)
      },
      { token, userId },
    )
  })

  test('create → list → edit → delete a transaction', async ({ page }) => {
    await page.goto('/transactions')

    // Empty state visible.
    await expect(page.getByTestId('transactions-view')).toBeVisible()
    await expect(page.getByTestId('tx-empty')).toBeVisible()
    await expect(page.getByTestId('tx-total')).toHaveText('0 transaksi')

    // Create — categories load async; wait for the select to have options.
    const select = page.getByTestId('tx-new-category')
    await expect(select.locator('option').first()).toBeAttached({ timeout: 5_000 })
    await page.getByTestId('tx-new-amount').fill('25000')
    await page.getByTestId('tx-new-note').fill('lunch')
    await page.getByTestId('tx-new-submit').click()

    // Row appears.
    const list = page.getByTestId('tx-list')
    await expect(list).toBeVisible()
    await expect(list.locator('li')).toHaveCount(1)
    const row = list.locator('li').first()
    await expect(row).toContainText('Rp 25.000')
    await expect(row).toContainText('lunch')
    await expect(page.getByTestId('tx-total')).toHaveText('1 transaksi')

    // Edit amount via the row's edit affordance.
    await row.getByText('edit').click()
    await page.getByTestId('tx-edit-amount').fill('30000')
    await page.getByTestId('tx-edit-save').click()

    await expect(list.locator('li').first()).toContainText('Rp 30.000')

    // Delete.
    await list.locator('li').first().getByText('hapus').click()
    await expect(page.getByTestId('tx-empty')).toBeVisible()
    await expect(page.getByTestId('tx-total')).toHaveText('0 transaksi')
  })

  test('backend validation surfaces inline', async ({ page }) => {
    await page.goto('/transactions')
    await expect(page.getByTestId('tx-empty')).toBeVisible()

    // Submit zero amount → client-side guard catches it (errorMsg set).
    await page.getByTestId('tx-new-amount').fill('0')
    await page.getByTestId('tx-new-submit').click()
    await expect(page.getByTestId('tx-error')).toBeVisible()
  })

  test('accepts an amount that is NOT a multiple of 1000 (no native step block)', async ({
    page,
  }) => {
    // Regression: the form used to carry step="1000", so a value like 120
    // failed native HTML validation and the submit was silently swallowed.
    // The form is now novalidate + step="1", so an arbitrary amount with a
    // category selected succeeds (201) and the row appears.
    await page.goto('/transactions')
    await expect(page.getByTestId('tx-empty')).toBeVisible()

    // Categories load async — wait for the select to populate so a category
    // is actually selected before submit.
    const select = page.getByTestId('tx-new-category')
    await expect(select.locator('option').first()).toBeAttached({ timeout: 5_000 })

    // Watch for the create call to confirm it really hits the API (201), not
    // blocked client-side.
    const createResponse = page.waitForResponse(
      (resp) =>
        resp.url().includes('/v1/transactions') &&
        resp.request().method() === 'POST',
    )

    await page.getByTestId('tx-new-amount').fill('120')
    await page.getByTestId('tx-new-note').fill('parkir')
    await page.getByTestId('tx-new-submit').click()

    const resp = await createResponse
    expect(resp.status()).toBe(201)

    // No inline error, and the odd amount shows up in the list.
    await expect(page.getByTestId('tx-error')).toHaveCount(0)
    const list = page.getByTestId('tx-list')
    await expect(list.locator('li')).toHaveCount(1)
    await expect(list.locator('li').first()).toContainText('Rp 120')
    await expect(list.locator('li').first()).toContainText('parkir')
  })

  test('month filter scopes the list to the selected calendar month', async ({ page }) => {
    await page.goto('/transactions')
    await expect(page.getByTestId('tx-empty')).toBeVisible()

    // Create a transaction — defaults transacted_at to now (current month).
    const select = page.getByTestId('tx-new-category')
    await expect(select.locator('option').first()).toBeAttached({ timeout: 5_000 })
    await page.getByTestId('tx-new-amount').fill('42000')
    await page.getByTestId('tx-new-note').fill('kopi')
    await page.getByTestId('tx-new-submit').click()

    // It shows in the current month.
    const list = page.getByTestId('tx-list')
    await expect(list.locator('li')).toHaveCount(1)
    await expect(list.locator('li').first()).toContainText('kopi')
    await expect(page.getByTestId('tx-total')).toHaveText('1 transaksi')

    // Step back a month → no transactions there.
    await page.getByTestId('tx-month-prev').click()
    await expect(page.getByTestId('tx-empty')).toBeVisible()
    await expect(page.getByTestId('tx-total')).toHaveText('0 transaksi')

    // Step forward back to the current month → the transaction reappears.
    await page.getByTestId('tx-month-next').click()
    await expect(page.getByTestId('tx-list').locator('li')).toHaveCount(1)
    await expect(page.getByTestId('tx-list').locator('li').first()).toContainText('kopi')
    await expect(page.getByTestId('tx-total')).toHaveText('1 transaksi')
  })

  test('formats the amount with thousand separators and logs a past-dated entry', async ({
    page,
  }) => {
    await page.goto('/transactions')
    const select = page.getByTestId('tx-new-category')
    await expect(select.locator('option').first()).toBeAttached({ timeout: 5_000 })

    // Typing digits shows the Indonesian "." thousand separator live.
    await page.getByTestId('tx-new-amount').fill('2000')
    await expect(page.getByTestId('tx-new-amount')).toHaveValue('2.000')
    await page.getByTestId('tx-new-amount').fill('1500000')
    await expect(page.getByTestId('tx-new-amount')).toHaveValue('1.500.000')

    // Backdate the entry to the 15th of the previous month.
    const now = new Date()
    const prev = new Date(now.getFullYear(), now.getMonth() - 1, 15)
    const yyyy = prev.getFullYear()
    const mm = String(prev.getMonth() + 1).padStart(2, '0')
    await page.getByTestId('tx-new-date').fill(`${yyyy}-${mm}-15`)
    await page.getByTestId('tx-new-note').fill('sewa')
    await page.getByTestId('tx-new-submit').click()

    // The month filter jumped to the entry's month so the past-dated row shows.
    const list = page.getByTestId('tx-list')
    await expect(list.locator('li')).toHaveCount(1)
    const row = list.locator('li').first()
    await expect(row).toContainText('Rp 1.500.000')
    await expect(row).toContainText('sewa')
    await expect(row).toContainText(String(yyyy))
  })
})
