import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { isAuthenticated } from '@/api/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    meta: { public: true },
    component: () => import('@/views/LoginView.vue'),
  },
  {
    path: '/register',
    name: 'register',
    meta: { public: true },
    component: () => import('@/views/RegisterView.vue'),
  },
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
  {
    path: '/transactions',
    name: 'transactions',
    component: () => import('@/views/TransactionsView.vue'),
  },
  {
    path: '/budget',
    name: 'budget',
    component: () => import('@/views/BudgetView.vue'),
  },
  {
    path: '/scan',
    name: 'scan',
    component: () => import('@/views/ScanView.vue'),
  },
  {
    path: '/reports',
    name: 'reports',
    component: () => import('@/views/ReportsView.vue'),
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Gate everything except routes flagged `meta.public` (login/register).
// isAuthenticated() verifies the JWT's exp client-side, so an expired or
// garbage token counts as logged-out — the user is bounced to /login,
// preserving the intended destination so we can return them after auth.
router.beforeEach((to) => {
  if (!to.meta.public && !isAuthenticated()) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
})
