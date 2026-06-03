import axios from 'axios'
import { http } from './client'
import type { ExpenseType } from './categories'

export type HousingType = 'kosan' | 'kpr' | 'keluarga'
export type Goal = 'emergency' | 'debt' | 'goal' | 'invest' | 'balance'
export type DebtType = 'none' | 'cc' | 'paylater' | 'multi'
export type LifestyleStyle = 'easy' | 'balanced' | 'strict'
export type Program = 'pondasi' | 'bebas_utang' | 'goal_chaser' | 'tumbuh' | 'seimbang'

export interface OnboardingItem {
  category_id: string
  name: string
  icon?: string
  type: ExpenseType
  amount: number
}

export interface OnboardingRequest {
  income: number
  housing_type: HousingType
  goal: Goal
  debt_types: DebtType[]
  emergency_months: 0 | 1 | 3 | 6
  lifestyle_style: LifestyleStyle
  expense_items: OnboardingItem[]
}

export interface OnboardingBucket {
  amount: number
  percentage: number
}

export interface OnboardingResponseItem {
  category_id: string
  category_name: string
  icon?: string
  type: ExpenseType
  allocated_amount: number
  percentage: number
  is_debt_focus?: boolean
}

export interface OnboardingResponse {
  program: Program
  budget_plan_id: string
  period: string
  total_income: number
  summary: {
    kebutuhan: OnboardingBucket
    utang: OnboardingBucket
    keinginan: OnboardingBucket
    tabungan: OnboardingBucket
  }
  items: OnboardingResponseItem[]
  warning?: string
}

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

export async function submitOnboarding(body: OnboardingRequest): Promise<OnboardingResponse> {
  try {
    const { data } = await http.post<Envelope<OnboardingResponse>>('/v1/onboarding', body)
    if (!data.data) throw new Error(data.error?.message ?? 'onboarding failed')
    return data.data
  } catch (err) {
    // Surface the backend's validation message, not "Request failed with status code 400".
    if (axios.isAxiosError(err) && err.response?.data?.error?.message) {
      throw new Error(err.response.data.error.message)
    }
    throw err
  }
}

export const PROGRAM_LABELS: Record<Program, string> = {
  pondasi: 'Program Pondasi',
  bebas_utang: 'Program Bebas Utang',
  goal_chaser: 'Program Goal Chaser',
  tumbuh: 'Program Tumbuh',
  seimbang: 'Program Seimbang',
}
