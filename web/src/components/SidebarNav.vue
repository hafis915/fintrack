<script setup lang="ts">
// Desktop-only left sidebar (>=lg). Mirrors the mobile bottom-nav destinations
// so the two navs stay in lockstep — same routes, same icons, same labels.
// The bottom nav stays the source of truth for mobile; this is purely additive.
import type { RouteLocationRaw } from 'vue-router'

interface NavItem {
  to: RouteLocationRaw
  testid: string
  icon: string
  label: string
}

const items: NavItem[] = [
  { to: { name: 'home' }, testid: 'sidebar-home', icon: '🏠', label: 'Beranda' },
  { to: { name: 'transactions' }, testid: 'sidebar-transactions', icon: '📒', label: 'Transaksi' },
  { to: { name: 'scan' }, testid: 'sidebar-scan', icon: '📷', label: 'Scan' },
  { to: { name: 'budget' }, testid: 'sidebar-budget', icon: '🎯', label: 'Budget' },
  { to: { name: 'reports' }, testid: 'sidebar-reports', icon: '📊', label: 'Reports' },
]
</script>

<template>
  <aside
    class="fixed inset-y-0 left-0 z-20 hidden w-60 flex-col border-r border-line bg-surface px-4 py-6 lg:flex"
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
        class="flex items-center gap-3 rounded-card px-3 py-2.5 text-sm text-muted transition-colors hover:bg-elevated hover:text-fg motion-reduce:transition-none"
        exact-active-class="text-saffron"
      >
        <span class="text-lg" aria-hidden="true">{{ item.icon }}</span>
        <span>{{ item.label }}</span>
      </router-link>
    </nav>
  </aside>
</template>
