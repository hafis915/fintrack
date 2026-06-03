import { defineConfig, devices } from '@playwright/test'

// Playwright config — starts the full local stack (Go API + Vite) before
// running tests. Assumes `docker compose up -d` is already running (postgres
// + minio) since spinning those up from here would be slow and noisy.
//
// Run with: `make test-e2e` (from repo root) or `npm run test:e2e` (from web/).

const repoRoot = '..'

// Real test secrets — match testhelper.TestJWTSecret / TestEncryptionKey
// so a JWT minted by `go run ./cmd/mint-jwt` against the same env works.
const TEST_ENV = {
  ENV: 'test',
  LOG_LEVEL: 'warn',
  HTTP_HOST: '127.0.0.1',
  HTTP_PORT: '8088',
  CORS_ALLOWED_ORIGINS: 'http://127.0.0.1:5173',
  DATABASE_URL: 'postgres://fintrack:fintrack@localhost:55432/fintrack_test?sslmode=disable',
  JWT_SECRET: 'test-secret-do-not-use-in-prod-test-secret-do-not-use-in-prod',
  JWT_ISSUER: 'fintrack-test',
  INCOME_ENCRYPTION_KEY: '00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff',
  STORAGE_ENDPOINT: 'localhost:9000',
  STORAGE_ACCESS_KEY: 'fintrack',
  STORAGE_SECRET_KEY: 'fintrack-dev-secret',
  STORAGE_BUCKET: 'receipts',
  STORAGE_USE_SSL: 'false',
}

export default defineConfig({
  testDir: './e2e',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'list',

  use: {
    baseURL: 'http://127.0.0.1:5173',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],

  // webServer accepts an array — both processes must be reachable before tests run.
  webServer: [
    {
      command: 'go run ./apps/api',
      cwd: repoRoot,
      url: 'http://127.0.0.1:8088/health',
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
      stdout: 'pipe',
      stderr: 'pipe',
      env: TEST_ENV,
    },
    {
      command: 'npm run dev -- --host 127.0.0.1',
      url: 'http://127.0.0.1:5173',
      reuseExistingServer: !process.env.CI,
      timeout: 60_000,
      stdout: 'pipe',
      stderr: 'pipe',
      env: {
        VITE_API_PROXY_TARGET: 'http://127.0.0.1:8088',
      },
    },
  ],
})
