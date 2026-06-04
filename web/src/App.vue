<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import SidebarNav from '@/components/SidebarNav.vue'

// Shell + responsive nav. The router decides what renders inside.
// Mobile (<lg): fixed bottom tab nav. Desktop (>=lg): fixed left sidebar.
const route = useRoute()

// Hide both navs on flows that own the full screen (onboarding, auth).
const fullScreenRoutes = new Set(['login', 'register'])
const showNav = computed(() => {
  const name = String(route.name ?? '')
  return !name.startsWith('onboarding') && !fullScreenRoutes.has(name)
})
</script>

<template>
  <main class="min-h-screen bg-bg text-fg">
    <!-- Desktop left sidebar (>=lg only); same destinations as the bottom nav. -->
    <SidebarNav v-if="showNav" />

    <!--
      Content wrapper.
      - Mobile: bottom-nav clearance via pb-24; child views own their width.
      - Desktop: clear the sidebar (lg:pl-60), drop the bottom-nav padding,
        and stop constraining to 420px so wide pages (Reports) can breathe.
        We don't force a width here — children control their own max-width,
        capped at lg:max-w-5xl so layouts stay centered, not edge-to-edge.
    -->
    <div :class="showNav ? 'pb-24 lg:pb-0 lg:pl-60' : ''">
      <div class="lg:mx-auto lg:max-w-5xl">
        <router-view />
      </div>
    </div>

    <!-- Mobile bottom tab nav (<lg only). -->
    <nav
      v-if="showNav"
      class="fixed inset-x-0 bottom-0 z-10 border-t-2 border-line bg-surface lg:hidden"
      data-testid="bottom-nav"
    >
      <div class="mx-auto flex max-w-mobile items-center justify-around px-4 py-2">
        <router-link
          :to="{ name: 'home' }"
          data-testid="nav-home"
          class="flex flex-col items-center gap-0.5 px-1.5 py-1 text-[10px] uppercase tracking-wider"
          active-class="text-saffron"
          exact-active-class="text-saffron"
        >
          <span class="text-lg">🏠</span>
          <span>Beranda</span>
        </router-link>

        <router-link
          :to="{ name: 'transactions' }"
          data-testid="nav-transactions"
          class="flex flex-col items-center gap-0.5 px-1.5 py-1 text-[10px] uppercase tracking-wider"
          active-class="text-saffron"
        >
          <span class="text-lg">📒</span>
          <span>Transaksi</span>
        </router-link>

        <!-- Prominent scan action -->
        <router-link
          :to="{ name: 'scan' }"
          data-testid="nav-scan"
          class="flex flex-col items-center gap-1 px-1.5"
        >
          <span
            class="-mt-6 flex h-14 w-14 items-center justify-center rounded-full bg-saffron text-2xl text-bg shadow-lg"
          >
            📷
          </span>
          <span class="text-[10px] uppercase tracking-wider text-saffron">Scan</span>
        </router-link>

        <router-link
          :to="{ name: 'budget' }"
          data-testid="nav-budget"
          class="flex flex-col items-center gap-0.5 px-1.5 py-1 text-[10px] uppercase tracking-wider"
          active-class="text-saffron"
        >
          <span class="text-lg">🎯</span>
          <span>Budget</span>
        </router-link>

        <router-link
          :to="{ name: 'reports' }"
          data-testid="nav-reports"
          class="flex flex-col items-center gap-0.5 px-1.5 py-1 text-[10px] uppercase tracking-wider"
          active-class="text-saffron"
        >
          <span class="text-lg">📊</span>
          <span>Reports</span>
        </router-link>
      </div>
    </nav>
  </main>
</template>
