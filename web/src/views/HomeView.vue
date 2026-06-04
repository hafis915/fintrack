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
  fresh: 'text-fresh border-fresh/40 bg-fresh/10',
  warning: 'text-warning border-warning/40 bg-warning/10',
  fatigued: 'text-fatigued border-fatigued/40 bg-fatigued/10',
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
  <section class="mx-auto flex max-w-mobile flex-col gap-6 px-6 py-10" data-testid="home-view">
    <header class="space-y-2">
      <div class="flex items-start justify-between gap-3">
        <h1 class="font-display text-3xl font-semibold" data-testid="home-greeting">
          {{ greeting }}
        </h1>
        <button
          type="button"
          class="shrink-0 rounded-full border border-line px-3 py-1 text-xs text-muted hover:border-fatigued hover:text-fatigued"
          data-testid="home-logout"
          @click="onLogout"
        >
          Keluar
        </button>
      </div>
      <p class="text-sm text-muted">
        Money discipline that feels like training, not bookkeeping.
      </p>
    </header>

    <p v-if="loading" class="font-mono text-sm text-muted">memuat…</p>

    <p v-else-if="errorMsg" class="text-sm text-fatigued" data-testid="home-error">
      {{ errorMsg }}
    </p>

    <!-- No plan yet -->
    <div
      v-else-if="noPlan"
      class="space-y-3 rounded-card border border-line bg-surface p-5 text-sm text-muted"
      data-testid="home-no-plan"
    >
      <p>Belum ada budget bulan ini. Selesaikan onboarding dulu — kami susun program yang pas.</p>
      <button
        type="button"
        class="w-full rounded-card bg-saffron py-2 font-semibold text-bg"
        @click="router.push({ name: 'onboarding' })"
      >
        Mulai onboarding
      </button>
    </div>

    <template v-else-if="budget">
      <!-- This-month snapshot -->
      <div
        class="space-y-3 rounded-card border border-line bg-surface p-5"
        data-testid="home-snapshot"
      >
        <p class="text-xs uppercase tracking-wider text-muted">Bulan ini</p>
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

        <div class="h-1.5 w-full overflow-hidden rounded-full bg-bg">
          <div
            class="h-full bg-saffron transition-all motion-reduce:transition-none"
            :style="{ width: Math.min(overallPercent, 100) + '%' }"
          />
        </div>
      </div>

      <!-- Needs attention -->
      <div v-if="attentionItems.length" class="space-y-3" data-testid="home-attention">
        <p class="text-xs uppercase tracking-wider text-muted">Perlu perhatian</p>
        <article
          v-for="item in attentionItems"
          :key="item.id"
          class="flex items-center justify-between gap-3 rounded-card border border-line bg-surface p-4"
          :data-testid="`home-attention-${item.category_name}`"
        >
          <div class="min-w-0">
            <p class="truncate text-sm">
              <span v-if="item.category_icon" class="mr-1">{{ item.category_icon }}</span>
              {{ item.category_name }}
            </p>
            <p class="font-mono text-xs text-muted">{{ item.percentage_used.toFixed(0) }}% dipakai</p>
          </div>
          <span
            :class="['shrink-0 rounded-full border px-2 py-[2px] text-[10px] uppercase tracking-wider', statusClass[item.status]]"
          >
            {{ item.status }}
          </span>
        </article>
      </div>
    </template>

    <!-- Quick actions: always available so user can log even before a snapshot loads -->
    <div v-if="!loading" class="space-y-3">
      <p class="text-xs uppercase tracking-wider text-muted">Aksi cepat</p>
      <div class="grid grid-cols-2 gap-3">
        <button
          type="button"
          class="rounded-card bg-saffron py-3 text-sm font-semibold text-bg"
          data-testid="home-cta-scan"
          @click="router.push({ name: 'scan' })"
        >
          Scan struk
        </button>
        <button
          type="button"
          class="rounded-card border border-line bg-surface py-3 text-sm font-semibold"
          data-testid="home-cta-transactions"
          @click="router.push({ name: 'transactions' })"
        >
          Catat transaksi
        </button>
      </div>
      <button
        type="button"
        class="text-xs text-muted hover:text-saffron"
        data-testid="home-cta-budget"
        @click="router.push({ name: 'budget' })"
      >
        → Lihat budget lengkap
      </button>
    </div>
  </section>
</template>
