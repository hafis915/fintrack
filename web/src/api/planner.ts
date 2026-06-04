import axios from 'axios'
import { http } from './client'
import type { ExpenseType } from './categories'
import type { DebtType, Goal, HousingType, LifestyleStyle, Program } from './onboarding'

interface Envelope<T> {
  data?: T
  error?: { code: string; message: string }
}

// --- /v1/onboarding/suggest ---

export interface SuggestFixedItem {
  category_id: string
  name: string
  icon?: string
  type: ExpenseType
  amount: number
}

export interface SuggestRequest {
  income: number
  housing_type: HousingType
  goal: Goal
  debt_types: DebtType[]
  emergency_months: number
  lifestyle_style: LifestyleStyle
  fixed_items: SuggestFixedItem[]
}

export interface SuggestBucket {
  amount: number
  percentage: number
}

export interface SuggestFlexibleItem {
  category_id: string
  name: string
  icon?: string
  type: ExpenseType
  suggested_amount: number
}

export interface SuggestResponse {
  program: Program
  savings_target: number
  fixed_total: number
  discretionary: number
  summary: {
    kebutuhan: SuggestBucket
    utang: SuggestBucket
    keinginan: SuggestBucket
    tabungan: SuggestBucket
  }
  flexible: SuggestFlexibleItem[]
  warning?: string
}

export async function suggestPlan(body: SuggestRequest): Promise<SuggestResponse> {
  try {
    const { data } = await http.post<Envelope<SuggestResponse>>('/v1/onboarding/suggest', body)
    if (!data.data) throw new Error(data.error?.message ?? 'gagal menyusun rencana')
    return data.data
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.data?.error?.message) {
      throw new Error(err.response.data.error.message)
    }
    throw err
  }
}

// --- /v1/planner/chat ---

export type PlannerRole = 'user' | 'assistant'

export interface PlannerMessage {
  role: PlannerRole
  content: string
}

export interface PlannerChatItem {
  category_id: string
  name: string
  amount: number
}

export interface PlannerChatRequest {
  income: number
  goal: Goal
  lifestyle_style: LifestyleStyle
  savings_target: number
  fixed_items: PlannerChatItem[]
  flexible: PlannerChatItem[]
  messages: PlannerMessage[]
  user_message: string
}

export interface PlannerChatResponse {
  reply: string
  flexible: PlannerChatItem[]
  savings_target: number
  changed: boolean
}

export async function plannerChat(body: PlannerChatRequest): Promise<PlannerChatResponse> {
  try {
    const { data } = await http.post<Envelope<PlannerChatResponse>>('/v1/planner/chat', body)
    if (!data.data) throw new Error(data.error?.message ?? 'planner gagal merespons')
    return data.data
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.data?.error?.message) {
      throw new Error(err.response.data.error.message)
    }
    throw err
  }
}
