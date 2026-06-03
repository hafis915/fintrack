import { defineStore } from 'pinia'
import type { OnboardingResponse } from '@/api/onboarding'

// Tiny store so the result page can read the plan handed off from the
// submit step without a second fetch. State doesn't survive a reload —
// the result page falls back to "go back to onboarding" if missing.
export const useOnboardingStore = defineStore('onboarding', {
  state: () => ({
    lastPlan: null as OnboardingResponse | null,
  }),
  actions: {
    setPlan(plan: OnboardingResponse) {
      this.lastPlan = plan
    },
    clear() {
      this.lastPlan = null
    },
  },
})
