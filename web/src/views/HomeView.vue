<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getCurrentBudget, NoPlanError, type CurrentBudget, type FatigueStatus } from '@/api/budget'
import { getUserName, logout } from '@/api/auth'

const router = useRouter()

function onLogout() {
  logout()
  router.push({ name: 'login' })
}
const budget = ref<CurrentBudget | null>(null)
const loading = ref(true)
const noPlan = ref(false)
const errorMsg = ref<string | null>(null)

const greeting = computed(() => {
  const name = getUserName()
  return name ? `Halo, ${name}` : 'Halo!'
})

// Surface only the categories that actually need attention, most stressed first.
const attentionItems = computed(() => {
  if (!budget.value) return []
  const order: Record<FatigueStatus, number> = { fatigued: 0, warning: 1, fresh: 2 }
  return budget.value.items
    .filter((i) => i.status === 'warning' || i.status === 'fatigued')
    .sort((a, b) => order[a.status] - order[b.status])
    .slice(0, 2)
})

const overallPercent = computed(() => budget.value?.summary.overall_percentage ?? 0)

const statusClass: Record<FatigueStatus, string> = {
  fresh: 'bg-fresh text-fg',
  warning: 'bg-warning text-fg',
  fatigued: 'bg-fatigued text-fg',
}

function formatRp(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}

async function refresh() {
  loading.value = true
  errorMsg.value = null
  noPlan.value = false
  try {
    budget.value = await getCurrentBudget()
  } catch (err) {
    if (err instanceof NoPlanError) {
      noPlan.value = true
    } else {
      errorMsg.value = err instanceof Error ? err.message : String(err)
    }
  } finally {
    loading.value = false
  }
}

onMounted(refresh)
</script>

<template>
  <section
    class="mx-auto flex w-full max-w-mobile flex-col gap-6 px-6 py-10 lg:max-w-none lg:px-10"
    data-testid="home-view"
  >
    <header class="space-y-2">
      <div class="flex items-start justify-between gap-3">
        <h1 class="font-display text-4xl font-extrabold uppercase leading-none tracking-tight lg:text-5xl" data-testid="home-greeting">
          {{ greeting }}
        </h1>
        <!-- Mobile only: desktop puts logout in the sidebar navbar. -->
        <button
          type="button"
          class="shrink-0 border-2 border-line bg-surface px-3 py-1 text-xs font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none lg:hidden"
          data-testid="home-logout"
          @click="onLogout"
        >
          Keluar
        </button>
      </div>
      <p class="border-l-4 border-saffron pl-3 text-sm font-medium">
        Money discipline that feels like training, not bookkeeping.
      </p>
    </header>

    <!-- On desktop: main snapshot/attention column + a quick-actions right rail. -->
    <div class="flex flex-col gap-6 lg:grid lg:grid-cols-3 lg:items-start lg:gap-6">

      <!-- Main column: snapshot + what needs attention. -->
      <div class="space-y-6 lg:col-span-2">
        <p v-if="loading" class="font-mono text-sm text-muted">memuat…</p>

        <p v-else-if="errorMsg" class="text-sm text-fatigued" data-testid="home-error">
          {{ errorMsg }}
        </p>

        <!-- No plan yet -->
    <div
      v-else-if="noPlan"
      class="space-y-3 rounded-card border-2 border-line bg-surface p-5 text-sm shadow-brutal"
      data-testid="home-no-plan"
    >
      <p>Belum ada budget bulan ini. Selesaikan onboarding dulu — kami susun program yang pas.</p>
      <button
        type="button"
        class="w-full rounded-card border-2 border-line bg-saffron py-2 font-bold uppercase text-fg shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
        @click="router.push({ name: 'onboarding' })"
      >
        Mulai onboarding
      </button>
    </div>

    <template v-else-if="budget">
      <!-- This-month snapshot -->
      <div
        class="space-y-3 rounded-card border-2 border-line bg-surface p-5 shadow-brutal"
        data-testid="home-snapshot"
      >
        <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Bulan ini</span>
        <dl class="space-y-1 font-mono text-sm">
          <div class="flex justify-between">
            <dt class="text-muted">Pemasukan</dt>
            <dd>{{ formatRp(budget.total_income) }}</dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-muted">Dipakai</dt>
            <dd>
              {{ formatRp(budget.summary.total_spent) }}
              <span class="ml-2 text-xs text-muted">{{ overallPercent.toFixed(1) }}%</span>
            </dd>
          </div>
        </dl>

        <div class="h-3 w-full overflow-hidden border-2 border-line bg-bg">
          <div
            class="h-full bg-saffron transition-all motion-reduce:transition-none"
            :style="{ width: Math.min(overallPercent, 100) + '%' }"
          />
        </div>
      </div>

      <!-- Needs attention -->
      <div v-if="attentionItems.length" class="space-y-3" data-testid="home-attention">
        <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Perlu perhatian</span>
        <article
          v-for="item in attentionItems"
          :key="item.id"
          class="flex items-center justify-between gap-3 rounded-card border-2 border-line bg-surface p-4 shadow-brutal"
          :data-testid="`home-attention-${item.category_name}`"
        >
          <div class="min-w-0">
            <p class="truncate text-sm font-semibold">
              <span v-if="item.category_icon" class="mr-1">{{ item.category_icon }}</span>
              {{ item.category_name }}
            </p>
            <p class="font-mono text-xs text-muted">{{ item.percentage_used.toFixed(0) }}% dipakai</p>
          </div>
          <span
            :class="['shrink-0 border-2 border-line px-2 py-[2px] text-[10px] font-bold uppercase', statusClass[item.status]]"
          >
            {{ item.status }}
          </span>
        </article>
      </div>
        </template>
      </div>
      <!-- /main column -->

      <!-- Quick actions: prominent right-rail panel on desktop, always available
           so the user can log even before a snapshot loads. -->
      <aside
        v-if="!loading"
        class="space-y-3 lg:col-span-1 lg:rounded-card lg:border-2 lg:border-line lg:bg-surface lg:p-5 lg:shadow-brutal"
        data-testid="home-quick-actions"
      >
        <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Aksi cepat</span>
        <div class="grid grid-cols-2 gap-3 lg:grid-cols-1">
          <button
            type="button"
            class="flex items-center justify-center gap-2 rounded-card border-2 border-line bg-saffron py-4 text-sm font-bold uppercase text-fg shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
            data-testid="home-cta-scan"
            @click="router.push({ name: 'scan' })"
          >
            <span class="text-xl" aria-hidden="true">📷</span>
            <span>Scan struk</span>
          </button>
          <button
            type="button"
            class="flex items-center justify-center gap-2 rounded-card border-2 border-line bg-fg py-4 text-sm font-bold uppercase text-bg shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
            data-testid="home-cta-transactions"
            @click="router.push({ name: 'transactions' })"
          >
            <span class="text-xl" aria-hidden="true">📒</span>
            <span>Catat transaksi</span>
          </button>
        </div>
        <button
          type="button"
          class="flex w-full items-center justify-center gap-2 border-2 border-line bg-surface px-3 py-2.5 text-xs font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
          data-testid="home-cta-budget"
          @click="router.push({ name: 'budget' })"
        >
          <span aria-hidden="true">🎯</span>
          <span>Lihat budget lengkap</span>
        </button>
      </aside>
    </div>
  </section>
</template>
