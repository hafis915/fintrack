import { http } from './client'

export interface HealthResponse {
  status: string
  db: string
  version: string
}

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

export async function getHealth(): Promise<HealthResponse> {
  const { data } = await http.get<Envelope<HealthResponse>>('/health')
  if (!data.data) throw new Error(data.error?.message ?? 'unknown error')
  return data.data
}
