import axios from 'axios'
import { http } from './client'
import type { ExpenseType } from './categories'

export interface Transaction {
  id: string
  category_id: string
  category_name?: string
  category_icon?: string
  category_type?: ExpenseType
  amount: number
  note?: string
  transacted_at: string
  budget_plan_id?: string
  ai_categorized: boolean
  created_at: string
  updated_at: string
}

export interface TransactionsList {
  items: Transaction[]
  total: number
}

export interface CreateTransactionInput {
  category_id: string
  amount: number
  note?: string
  transacted_at: string // RFC3339
}

export interface UpdateTransactionInput {
  category_id?: string
  amount?: number
  note?: string | null
  transacted_at?: string
}

export interface ListFilter {
  category_id?: string
  from?: string
  to?: string
  limit?: number
  offset?: number
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

export function createTransaction(input: CreateTransactionInput): Promise<Transaction> {
  return unwrap(http.post<Envelope<Transaction>>('/v1/transactions', input))
}

export function listTransactions(filter: ListFilter = {}): Promise<TransactionsList> {
  return unwrap(
    http.get<Envelope<TransactionsList>>('/v1/transactions', { params: filter }),
  )
}

export function updateTransaction(id: string, input: UpdateTransactionInput): Promise<Transaction> {
  return unwrap(http.patch<Envelope<Transaction>>(`/v1/transactions/${id}`, input))
}

export async function deleteTransaction(id: string): Promise<void> {
  try {
    await http.delete(`/v1/transactions/${id}`)
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.data?.error?.message) {
      throw new Error(err.response.data.error.message)
    }
    throw err
  }
}
