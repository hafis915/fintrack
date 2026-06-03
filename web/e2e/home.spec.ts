import { test, expect } from '@playwright/test'

/**
 * Phase 0 smoke flow.
 *
 * A real user lands on the home page and the app should:
 *   - render the Fintrack heading
 *   - call the backend /health endpoint via the Vite proxy
 *   - render db:ok + status:ok in the status card
 *
 * This is the only user-facing flow in Phase 0. Future phases must add
 * one spec per flow (onboarding, transactions, scan, fatigue) — see DoD in
 * CLAUDE.md.
 */
test.describe('Phase 0 home', () => {
  test('renders heading and live health from the backend', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('heading', { name: 'Fintrack' })).toBeVisible()
    await expect(page.getByText(/Money discipline that feels like training/i)).toBeVisible()

    const card = page.getByTestId('health-card')
    await expect(card).toBeVisible()

    await expect(page.getByTestId('health-loading')).toBeHidden({ timeout: 5_000 })

    await expect(page.getByTestId('health-status')).toHaveText('ok')
    await expect(page.getByTestId('health-db')).toHaveText('ok')
    await expect(page.getByTestId('health-version')).toHaveText('0.0.0')
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
