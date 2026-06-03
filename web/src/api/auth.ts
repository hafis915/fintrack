// Local-only auth: the JWT is minted out-of-band via `make token` (dev)
// or by Playwright's globalSetup (e2e). We just persist it in localStorage
// and let the axios interceptor attach it to every request.
//
// When we swap to Supabase (ADR-014), this module will be replaced by
// supabase.auth's session helpers. The call sites in views/components
// should keep working unchanged.

const TOKEN_KEY = 'fintrack.token'
const USER_ID_KEY = 'fintrack.user_id'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string, userId?: string): void {
  localStorage.setItem(TOKEN_KEY, token)
  if (userId) localStorage.setItem(USER_ID_KEY, userId)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_ID_KEY)
}

export function getUserId(): string | null {
  return localStorage.getItem(USER_ID_KEY)
}
