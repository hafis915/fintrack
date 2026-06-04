<script setup lang="ts">
// Desktop-only left sidebar (>=lg). Mirrors the mobile bottom-nav destinations
// so the two navs stay in lockstep — same routes, same icons, same labels.
// The bottom nav stays the source of truth for mobile; this is purely additive.
import { computed } from 'vue'
import { useRouter, type RouteLocationRaw } from 'vue-router'
import { getUserName, logout } from '@/api/auth'

interface NavItem {
  to: RouteLocationRaw
  testid: string
  icon: string
  label: string
}

const router = useRouter()

const items: NavItem[] = [
  { to: { name: 'home' }, testid: 'sidebar-home', icon: '🏠', label: 'Beranda' },
  { to: { name: 'transactions' }, testid: 'sidebar-transactions', icon: '📒', label: 'Transaksi' },
  { to: { name: 'scan' }, testid: 'sidebar-scan', icon: '📷', label: 'Scan' },
  { to: { name: 'budget' }, testid: 'sidebar-budget', icon: '🎯', label: 'Budget' },
  { to: { name: 'reports' }, testid: 'sidebar-reports', icon: '📊', label: 'Reports' },
]

const userName = computed(() => getUserName() ?? 'Akun')

function onLogout() {
  logout()
  router.push({ name: 'login' })
}
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-20 hidden w-60 flex-col border-r-2 border-line bg-surface px-4 py-6 lg:flex"
    data-testid="sidebar-nav"
  >
    <!-- Wordmark -->
    <router-link
      :to="{ name: 'home' }"
      class="mb-8 px-2 font-display text-xl font-semibold tracking-tight"
      data-testid="sidebar-wordmark"
    >
      Fin<span class="text-saffron">track</span>
    </router-link>

    <nav class="flex flex-col gap-1">
      <router-link
        v-for="item in items"
        :key="item.testid"
        :to="item.to"
        :data-testid="item.testid"
        class="flex items-center gap-3 rounded-card border-2 border-transparent px-3 py-2.5 text-sm font-semibold text-muted transition-colors hover:bg-elevated hover:text-fg motion-reduce:transition-none"
        exact-active-class="border-line bg-saffron text-fg shadow-brutal-sm"
      >
        <span class="text-lg" aria-hidden="true">{{ item.icon }}</span>
        <span>{{ item.label }}</span>
      </router-link>
    </nav>

    <!-- Account + logout pinned to the bottom of the navbar. -->
    <div class="mt-auto space-y-2 border-t-2 border-line pt-4">
      <p class="truncate px-3 text-xs font-semibold text-muted" data-testid="sidebar-account">
        {{ userName }}
      </p>
      <button
        type="button"
        data-testid="sidebar-logout"
        class="flex w-full items-center gap-3 rounded-card border-2 border-line bg-surface px-3 py-2.5 text-sm font-bold uppercase text-fg shadow-brutal-sm transition-transform hover:bg-elevated active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transition-none motion-reduce:transform-none"
        @click="onLogout"
      >
        <span class="text-lg" aria-hidden="true">⎋</span>
        <span>Keluar</span>
      </button>
    </div>
  </aside>
</template>
