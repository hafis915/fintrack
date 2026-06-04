import { defineStore } from 'pinia'
import type {
  DebtType,
  Goal,
  HousingType,
  LifestyleStyle,
  OnboardingResponse,
} from '@/api/onboarding'

// Snapshot of the intake form so "Ubah jawaban" (and any remount of the
// onboarding view) can rehydrate exactly what the user entered instead of
// resetting to defaults. Per-item amount/enabled is keyed by category_id so
// it can be merged back onto the categories list after it (re)loads.
export interface OnboardingAnswers {
  income: number
  housingType: HousingType
  goal: Goal
  debtTypes: DebtType[]
  emergencyMonths: 0 | 1 | 3 | 6
  lifestyleStyle: LifestyleStyle
  items: Record<string, { amount: number; enabled: boolean }>
}

// Tiny store so the result page can read the plan handed off from the
// submit step without a second fetch, and so the form can restore the
// user's answers on "Ubah jawaban". State doesn't survive a reload —
// the result page falls back to "go back to onboarding" if missing.
export const useOnboardingStore = defineStore('onboarding', {
  state: () => ({
    lastPlan: null as OnboardingResponse | null,
    answers: null as OnboardingAnswers | null,
  }),
  actions: {
    setPlan(plan: OnboardingResponse) {
      this.lastPlan = plan
    },
    setAnswers(answers: OnboardingAnswers) {
      this.answers = answers
    },
    clear() {
      this.lastPlan = null
      this.answers = null
    },
  },
})
