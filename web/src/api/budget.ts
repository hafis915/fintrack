import axios from 'axios'
import { http } from './client'
import type { ExpenseType } from './categories'
import type { Program } from './onboarding'

export type FatigueStatus = 'fresh' | 'warning' | 'fatigued'

export interface BudgetItem {
  id: string
  category_id: string
  category_name: string
  category_icon?: string
  category_type: ExpenseType
  allocated_amount: number
  spent_amount: number
  remaining: number
  percentage_used: number
  status: FatigueStatus
  coaching?: string
  is_debt_focus?: boolean
}

export interface BudgetSummary {
  total_allocated: number
  total_spent: number
  unallocated_spent?: number
  overall_percentage: number
}

export interface CurrentBudget {
  id: string
  period: string
  program: Program
  total_income: number
  items: BudgetItem[]
  summary: BudgetSummary
}

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

export class NoPlanError extends Error {
  code = 'no_active_plan'
}

export async function getCurrentBudget(): Promise<CurrentBudget> {
  try {
    const { data } = await http.get<Envelope<CurrentBudget>>('/v1/budget/current')
    if (!data.data) throw new Error(data.error?.message ?? 'failed to load budget')
    return data.data
  } catch (err) {
    if (axios.isAxiosError(err)) {
      if (err.response?.status === 404 && err.response.data?.error?.code === 'no_active_plan') {
        throw new NoPlanError(err.response.data.error.message ?? 'no plan')
      }
      if (err.response?.data?.error?.message) {
        throw new Error(err.response.data.error.message)
      }
    }
    throw err
  }
}
