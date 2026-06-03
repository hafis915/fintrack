import axios from 'axios'
import { getToken } from './auth'

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
