// Local-first auth (Phase 0): a lightweight email register/login mints the
// same HS256 JWT the middleware already validates (no passwords yet, no
// Supabase libs). We persist the token in localStorage and let the axios
// interceptor attach it to every /v1 request.
//
// When we swap to Supabase (ADR-014), this module will be replaced by
// supabase.auth's session helpers. The call sites in views/components
// should keep working unchanged.

import axios from 'axios'
import { http } from './client'

const TOKEN_KEY = 'fintrack.token'
const USER_ID_KEY = 'fintrack.user_id'
const NAME_KEY = 'fintrack.name'
const EMAIL_KEY = 'fintrack.email'

export interface AuthResult {
  token: string
  user_id: string
  email: string
  name: string
}

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

function unwrap<T>(p: Promise<{ data: Envelope<T> }>): Promise<T> {
  return p
    .then(({ data }) => {
      if (data.data === undefined) throw new Error(data.error?.message ?? 'request failed')
      return data.data
    })
    .catch((err) => {
      if (axios.isAxiosError(err) && err.response?.data?.error?.message) {
        throw new Error(err.response.data.error.message)
      }
      throw err
    })
}

// Public endpoint — no token needed. The request interceptor only adds the
// Authorization header when a token already exists, so these stay unauthed.
export async function register(
  name: string,
  email: string,
  password: string,
): Promise<AuthResult> {
  const result = await unwrap(
    http.post<Envelope<AuthResult>>('/v1/auth/register', { name, email, password }),
  )
  setSession(result.token, result.user_id, result.name, result.email)
  return result
}

// Public endpoint — no token needed.
export async function login(email: string, password: string): Promise<AuthResult> {
  const result = await unwrap(
    http.post<Envelope<AuthResult>>('/v1/auth/login', { email, password }),
  )
  setSession(result.token, result.user_id, result.name, result.email)
  return result
}

// Clear the session and return — callers redirect to /login. Kept as a named
// export so views don't reach into localStorage keys directly.
export function logout(): void {
  clearToken()
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function getUserId(): string | null {
  return localStorage.getItem(USER_ID_KEY)
}

export function getUserName(): string | null {
  return localStorage.getItem(NAME_KEY)
}

export function getUserEmail(): string | null {
  return localStorage.getItem(EMAIL_KEY)
}

// Persist the whole session. name/email are optional because some flows
// (or future callers) may only have the token + user_id on hand; we only
// overwrite the optional keys when a value is actually provided.
export function setSession(token: string, userId: string, name?: string, email?: string): void {
  localStorage.setItem(TOKEN_KEY, token)
  localStorage.setItem(USER_ID_KEY, userId)
  if (name) localStorage.setItem(NAME_KEY, name)
  if (email) localStorage.setItem(EMAIL_KEY, email)
}

export function clearToken(): void {
  localStorage.removeItem(TOKEN_KEY)
  localStorage.removeItem(USER_ID_KEY)
  localStorage.removeItem(NAME_KEY)
  localStorage.removeItem(EMAIL_KEY)
}

// Decode a base64url-encoded JWT segment to a string. JWTs use base64url
// (- and _ instead of + and /, no padding), so we normalize before atob.
function base64UrlDecode(segment: string): string {
  let b64 = segment.replace(/-/g, '+').replace(/_/g, '/')
  // atob requires the input length to be a multiple of 4 — re-pad.
  const pad = b64.length % 4
  if (pad) b64 += '='.repeat(4 - pad)
  return atob(b64)
}

// True only when a token exists AND its exp claim is present and still in the
// future per the browser clock. Any structural/parse problem (or a missing
// exp) is treated as "not authenticated" rather than throwing — the router
// guard and the 401 interceptor will route the user back to /login.
export function isAuthenticated(): boolean {
  const token = getToken()
  if (!token) return false

  try {
    const parts = token.split('.')
    if (parts.length !== 3) return false

    const payload = JSON.parse(base64UrlDecode(parts[1])) as { exp?: number }
    if (typeof payload.exp !== 'number') return false

    // exp is epoch SECONDS; compare against the current epoch in seconds.
    const nowSeconds = Date.now() / 1000
    return payload.exp > nowSeconds
  } catch {
    return false
  }
}
