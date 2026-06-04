import { test, expect } from '@playwright/test'
import { mintToken } from './helpers/auth'
import { resetOnboardingDB } from './helpers/db'

/**
 * Responsive nav (ADR-2026-06-04).
 *
 * Mobile-first stays the default: the bottom tab nav renders below the `lg`
 * breakpoint and the desktop sidebar renders at/above it. The two are mutually
 * exclusive (`lg:hidden` on the bottom nav, `hidden lg:flex` on the sidebar).
 *
 * The suite default viewport is mobile (390x844, see playwright.config.ts), so
 * the desktop checks explicitly resize to 1280x800.
 */
test.describe('Responsive nav', () => {
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

  test('mobile viewport shows bottom nav, hides sidebar', async ({ page }) => {
    // Default viewport is mobile.
    await page.goto('/')

    await expect(page.getByTestId('bottom-nav')).toBeVisible()
    await expect(page.getByTestId('sidebar-nav')).toBeHidden()

    // Tapping the Reports tab routes to /reports.
    await page.getByTestId('nav-reports').click()
    await expect(page).toHaveURL(/\/reports$/)
    await expect(page.getByTestId('reports-view')).toBeVisible()
  })

  test('desktop viewport shows sidebar, hides bottom nav', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 800 })
    await page.goto('/')

    await expect(page.getByTestId('sidebar-nav')).toBeVisible()
    await expect(page.getByTestId('bottom-nav')).toBeHidden()

    // The sidebar Reports link routes to /reports.
    await page.getByTestId('sidebar-reports').click()
    await expect(page).toHaveURL(/\/reports$/)
    await expect(page.getByTestId('reports-view')).toBeVisible()
  })
})
