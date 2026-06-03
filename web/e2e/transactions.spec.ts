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
})
