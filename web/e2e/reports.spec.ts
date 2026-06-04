import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'
import { getDefaultCategoryIDs, seedBudgetPlan, seedTransaction, seedUser } from './helpers/seed'

/**
 * Reports view: monthly per-category breakdown + CSV export.
 *
 * Month filtering is a frontend concern (listTransactions({ from, to })), so we
 * seed current-month transactions via SQL and assert the table/chart/total
 * reconcile, then walk the month filter back to an empty month.
 */
async function authedSession(
  page: import('@playwright/test').Page,
): Promise<{ userId: string }> {
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

// Seeds a plan + two current-month transactions in known categories.
function seedCurrentMonth(userId: string) {
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

  // seedTransaction defaults transacted_at to now → current month.
  seedTransaction({ userId, categoryId: cats['Makan & minum'], amount: 300_000 })
  seedTransaction({ userId, categoryId: cats['Makan & minum'], amount: 200_000 })
  seedTransaction({ userId, categoryId: cats['Hiburan'], amount: 150_000 })
  // total = 650.000; Makan & minum = 500.000 (2 tx); Hiburan = 150.000 (1 tx)
}

test.describe('Reports view', () => {
  test('renders the per-category table, total, and chart for the current month', async ({
    page,
  }) => {
    const { userId } = await authedSession(page)
    seedCurrentMonth(userId)

    await page.goto('/reports')

    await expect(page.getByTestId('reports-view')).toBeVisible()

    // Table rows for the seeded categories with the right aggregated spend.
    const makanRow = page.getByTestId('reports-row-Makan & minum')
    const hiburanRow = page.getByTestId('reports-row-Hiburan')
    await expect(makanRow).toBeVisible()
    await expect(makanRow).toContainText('Rp 500.000')
    await expect(hiburanRow).toBeVisible()
    await expect(hiburanRow).toContainText('Rp 150.000')

    // Total reflects the sum of all categories.
    await expect(page.getByTestId('reports-total')).toContainText('Rp 650.000')
    await expect(page.getByTestId('reports-row-total')).toContainText('Rp 650.000')

    // Chart is visible with a bar per category.
    await expect(page.getByTestId('reports-chart')).toBeVisible()
    await expect(page.getByTestId('reports-bar-Makan & minum')).toBeVisible()
    await expect(page.getByTestId('reports-bar-Hiburan')).toBeVisible()
  })

  test('walking to a month with no transactions shows the empty state', async ({ page }) => {
    const { userId } = await authedSession(page)
    seedCurrentMonth(userId)

    await page.goto('/reports')
    await expect(page.getByTestId('reports-table')).toBeVisible()

    // Prior month has no seeded transactions → empty state, no table.
    await page.getByTestId('reports-month-prev').click()
    await expect(page.getByTestId('reports-empty')).toBeVisible()
    await expect(page.getByTestId('reports-table')).toHaveCount(0)

    // Back to the current month → the table reappears.
    await page.getByTestId('reports-month-next').click()
    await expect(page.getByTestId('reports-table')).toBeVisible()
    await expect(page.getByTestId('reports-total')).toContainText('Rp 650.000')
  })

  test('exports the current month as a CSV download', async ({ page }) => {
    const { userId } = await authedSession(page)
    seedCurrentMonth(userId)

    await page.goto('/reports')
    // Export is disabled until the month's data loads.
    await expect(page.getByTestId('reports-export')).toBeEnabled()

    const downloadPromise = page.waitForEvent('download')
    await page.getByTestId('reports-export').click()
    const download = await downloadPromise

    // Filename: fintrack-YYYY-MM.csv
    expect(download.suggestedFilename()).toMatch(/^fintrack-\d{4}-\d{2}\.csv$/)

    // Read the stream and assert the header + a seeded category appear.
    const stream = await download.createReadStream()
    const chunks: Buffer[] = []
    for await (const chunk of stream) chunks.push(chunk as Buffer)
    const csv = Buffer.concat(chunks).toString('utf8')

    expect(csv).toContain('tanggal,kategori,merchant,jumlah,catatan')
    expect(csv).toContain('Makan & minum')
  })
})
