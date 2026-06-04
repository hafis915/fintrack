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

export interface CreateCategoryInput {
  name: string
  type: ExpenseType
  icon?: string
}

// Creates a user-scoped custom expense category (e.g. an onboarding item the
// default seed doesn't cover). It then shows up in listCategories() like any
// default, so transactions/budget/scan can use it too.
export async function createCategory(input: CreateCategoryInput): Promise<Category> {
  const { data } = await http.post<Envelope<Category>>('/v1/categories', input)
  if (!data.data) throw new Error(data.error?.message ?? 'failed to create category')
  return data.data
}
