import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/views/HomeView.vue'),
  },
  {
    path: '/onboarding',
    name: 'onboarding',
    component: () => import('@/views/OnboardingView.vue'),
  },
  {
    path: '/onboarding/result',
    name: 'onboarding-result',
    component: () => import('@/views/OnboardingResultView.vue'),
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})
