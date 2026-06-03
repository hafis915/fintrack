import { http } from './client'

export type ExpenseType = 'fixed' | 'variable' | 'debt' | 'want'

export interface Category {
  id: string
  name: string
  icon?: string
  type: ExpenseType
  is_default: boolean
  sort_order: number
}

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

export async function listCategories(): Promise<Category[]> {
  const { data } = await http.get<Envelope<Category[]>>('/v1/categories')
  if (!data.data) throw new Error(data.error?.message ?? 'failed to load categories')
  return data.data
}
