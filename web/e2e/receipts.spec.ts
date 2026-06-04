import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'

/**
 * Phase 3 receipt scan flow.
 *
 * In ENV=test the API uses the stub analyzer (Amount:50000, Merchant:"Indomaret",
 * CategoryHint:"Belanja", Confidence:0.95) and real MinIO storage. So a real
 * file upload produces a deterministic draft we can assert against.
 *
 * Flow: pick a photo → stub analyzer fills the draft → confirm → land on
 * /transactions with the new "Indomaret" row.
 *
 * Categories come from the migration's system defaults (user_id is null), so a
 * freshly minted user already has options in scan-category — no seed needed.
 */
const here = path.dirname(fileURLToPath(import.meta.url))
const fixture = path.resolve(here, 'fixtures', 'receipt.jpg')

test.describe('Receipt scan flow', () => {
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

  test('scan receipt → confirm → see transaction', async ({ page }) => {
    await page.goto('/scan')
    await expect(page.getByTestId('scan-view')).toBeVisible()

    // Upload the fixture. The picker input is hidden but setInputFiles works
    // regardless of visibility.
    await page.getByTestId('scan-file-input').setInputFiles(fixture)

    // Stub analyzer returns synchronously-ish; the draft card appears once the
    // upload + analyze round-trips complete.
    const draft = page.getByTestId('scan-draft')
    await expect(draft).toBeVisible({ timeout: 15_000 })

    // Stub values: amount 50000, merchant "Indomaret".
    await expect(page.getByTestId('scan-amount')).toHaveValue('50000')
    await expect(page.getByTestId('scan-merchant')).toHaveValue(/Indomaret/)

    // Category is preselected from the first available system default (the stub
    // returns no category_id). Guard against an empty selection just in case.
    const category = page.getByTestId('scan-category')
    await expect(category.locator('option').first()).toBeAttached({ timeout: 5_000 })
    const selected = await category.inputValue()
    if (!selected) {
      // Pick the first real (non-placeholder) option.
      const firstValue = await category
        .locator('option[value]:not([value=""])')
        .first()
        .getAttribute('value')
      if (firstValue) await category.selectOption(firstValue)
    }

    // Confirm → navigates to transactions.
    await page.getByTestId('scan-confirm').click()

    await page.waitForURL('**/transactions')
    await expect(page.getByTestId('transactions-view')).toBeVisible()

    // The new row shows the merchant from the scan.
    const list = page.getByTestId('tx-list')
    await expect(list).toBeVisible()
    await expect(list.locator('li')).toHaveCount(1)
    await expect(list.locator('li').first()).toContainText('Indomaret')
  })
})
