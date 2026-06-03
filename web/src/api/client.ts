import axios from 'axios'

// Same-origin in dev (vite proxies /v1 + /health to localhost:8080),
// configurable via VITE_API_BASE for prod.
export const http = axios.create({
  baseURL: import.meta.env.VITE_API_BASE ?? '',
  timeout: 10_000,
})
