import axios from 'axios'
import { clearToken, getToken } from './auth'

// Same-origin in dev (vite proxies /v1 + /health to localhost:8080),
// configurable via VITE_API_BASE for prod.
export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE ?? '',
  timeout: 10_000,
})

// Attach the locally-minted JWT to every /v1 call. /health is public so
// we don't gate on the token's presence; the backend ignores the header
// for public routes.
http.interceptors.request.use((config) => {
  const tok = getToken()
  if (tok) {
    config.headers = config.headers ?? {}
    config.headers.Authorization = `Bearer ${tok}`
  }
  return config
})

// On 401 the local token is stale/invalid — clear it and bounce to /login so
// the user re-auths instead of hitting a raw "Authorization header required".
// We skip this for the public auth endpoints themselves (a 401 there is the
// caller's concern, surfaced as an inline form error).
http.interceptors.response.use(
  (response) => response,
  (error) => {
    if (axios.isAxiosError(error) && error.response?.status === 401) {
      const url = error.config?.url ?? ''
      const isAuthEndpoint = url.includes('/v1/auth/')
      if (!isAuthEndpoint) {
        clearToken()
        if (window.location.pathname !== '/login') {
          const redirect = encodeURIComponent(
            window.location.pathname + window.location.search,
          )
          window.location.assign(`/login?redirect=${redirect}`)
        }
      }
    }
    return Promise.reject(error)
  },
)
