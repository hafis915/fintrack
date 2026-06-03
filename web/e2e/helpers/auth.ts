import { execSync } from 'node:child_process'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const here = path.dirname(fileURLToPath(import.meta.url))

// Mints a JWT via the Go helper for a fresh user UUID. Used in e2e specs
// so the test acts as an authenticated caller without depending on any
// frontend "login" UI (none exists in MVP local-first mode — see ADR-014).
export function mintToken(opts: { sub?: string; ttl?: string } = {}): {
  token: string
  userId: string
} {
  const userId = opts.sub ?? crypto.randomUUID()
  const ttl = opts.ttl ?? '24h'
  const repoRoot = path.resolve(here, '..', '..', '..')

  // Same TEST_ENV the Playwright webServer uses — secret/issuer must match
  // what the API is started with.
  const env = {
    ...process.env,
    JWT_SECRET: 'test-secret-do-not-use-in-prod-test-secret-do-not-use-in-prod',
    JWT_ISSUER: 'fintrack-test',
    DATABASE_URL: 'postgres://fintrack:fintrack@localhost:55432/fintrack_test?sslmode=disable',
    INCOME_ENCRYPTION_KEY:
      '00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff',
  }

  const out = execSync(`go run ./cmd/mint-jwt -sub ${userId} -ttl ${ttl}`, {
    cwd: repoRoot,
    env,
    encoding: 'utf8',
    stdio: ['ignore', 'pipe', 'ignore'],
  }).trim()

  return { token: out, userId }
}
