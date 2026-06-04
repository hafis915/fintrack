<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getCurrentBudget, NoPlanError, type CurrentBudget, type FatigueStatus } from '@/api/budget'
import { PROGRAM_LABELS } from '@/api/onboarding'
import ReduceSuggestions from '@/components/ReduceSuggestions.vue'
import BudgetCompareChart from '@/components/BudgetCompareChart.vue'

const router = useRouter()
const budget = ref<CurrentBudget | null>(null)
const loading = ref(true)
const noPlan = ref(false)
const errorMsg = ref<string | null>(null)

const statusOrder: Record<FatigueStatus, number> = { fatigued: 0, warning: 1, fresh: 2 }

const sortedItems = computed(() => {
  if (!budget.value) return []
  // Surface the most stressed categories first.
  return [...budget.value.items].sort((a, b) => statusOrder[a.status] - statusOrder[b.status])
})

// Brutalist: solid saturated blocks with a black border + black ink.
const statusClass: Record<FatigueStatus, string> = {
  fresh: 'bg-fresh text-fg',
  warning: 'bg-warning text-fg',
  fatigued: 'bg-fatigued text-fg',
}
const barClass: Record<FatigueStatus, string> = {
  fresh: 'bg-fresh',
  warning: 'bg-warning',
  fatigued: 'bg-fatigued',
}

const statusLabel: Record<FatigueStatus, string> = {
  fresh: 'fresh',
  warning: 'warning',
  fatigued: 'fatigued',
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
    data-testid="budget-view"
  >
    <header class="space-y-2">
      <span class="inline-block border-2 border-line bg-fg px-2 py-1 text-[10px] font-bold uppercase tracking-[0.18em] text-bg">
        Fatigue dashboard
      </span>
      <h1 class="font-display text-4xl font-extrabold uppercase leading-none tracking-tight">
        Budget<br />bulan ini
      </h1>
      <p v-if="budget" class="font-mono text-sm" data-testid="budget-period">
        {{ budget.period }} · {{ PROGRAM_LABELS[budget.program] }}
      </p>

      <!-- Re-run the planner to regenerate this month's budget. Safe: the plan
           is only replaced on the final "Mulai program" submit (idempotent
           upsert), so navigating here destroys nothing. -->
      <button
        v-if="budget"
        type="button"
        data-testid="budget-rebudget"
        class="self-start border-2 border-line bg-surface px-3 py-1.5 text-xs font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
        @click="router.push({ name: 'onboarding' })"
      >
        ↻ Atur ulang budget
      </button>
    </header>

    <p v-if="loading" class="font-mono text-sm text-muted">memuat…</p>

    <div
      v-else-if="noPlan"
      class="space-y-3 rounded-card border-2 border-line bg-surface p-5 text-sm shadow-brutal"
      data-testid="budget-no-plan"
    >
      <p>
        Belum ada budget untuk bulan ini. Selesaikan onboarding dulu — kami susun program yang pas.
      </p>
      <button
        type="button"
        class="w-full rounded-card border-2 border-line bg-saffron py-2 font-bold uppercase text-fg shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
        @click="router.push({ name: 'onboarding' })"
      >
        Mulai onboarding
      </button>
    </div>

    <p v-else-if="errorMsg" class="text-sm text-fatigued" data-testid="budget-error">
      {{ errorMsg }}
    </p>

    <template v-else-if="budget">
      <!-- Desktop: summary + recommendations in a left column, the budget-vs-
           realisasi graphic gets the wide right column. Stacks on mobile. -->
      <div class="flex flex-col gap-6 lg:grid lg:grid-cols-3 lg:items-start lg:gap-6">
      <div class="space-y-6 lg:col-span-1">
      <!-- Summary -->
      <div
        class="space-y-3 rounded-card border-2 border-line bg-surface p-4 shadow-brutal"
        data-testid="budget-summary"
      >
        <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Ringkasan</span>
        <dl class="space-y-1 font-mono text-sm">
          <div class="flex justify-between">
            <dt class="text-muted">Total pemasukan</dt>
            <dd>{{ formatRp(budget.total_income) }}</dd>
          </div>
          <div class="flex justify-between" data-testid="summary-spent">
            <dt class="text-muted">Total dipakai</dt>
            <dd>
              {{ formatRp(budget.summary.total_spent) }}
              <span class="ml-2 text-xs text-muted">
                {{ budget.summary.overall_percentage.toFixed(1) }}%
              </span>
            </dd>
          </div>
          <div class="flex justify-between">
            <dt class="text-muted">Total dialokasikan</dt>
            <dd>{{ formatRp(budget.summary.total_allocated) }}</dd>
          </div>
          <div
            v-if="(budget.summary.unallocated_spent ?? 0) > 0"
            class="flex justify-between"
            data-testid="summary-unallocated"
          >
            <dt class="text-muted">Belum masuk plan</dt>
            <dd class="text-warning">{{ formatRp(budget.summary.unallocated_spent!) }}</dd>
          </div>
        </dl>
      </div>

      <!-- Recommendations — most actionable, near the top -->
      <ReduceSuggestions :items="budget.items" />
      </div>
      <!-- /left column -->

      <!-- Budget vs realisasi graphic (wide column on desktop) -->
      <div class="lg:col-span-2">
        <BudgetCompareChart :items="budget.items" />
      </div>
      </div>
      <!-- /summary+chart grid -->

      <!-- Items -->
      <div class="space-y-3" data-testid="budget-items">
        <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Per kategori</span>
        <div class="grid gap-3 lg:grid-cols-2 xl:grid-cols-3">
        <article
          v-for="item in sortedItems"
          :key="item.id"
          class="space-y-2 rounded-card border-2 border-line bg-surface p-4 shadow-brutal"
          :data-testid="`budget-item-${item.category_name}`"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-sm font-semibold">
                <span class="mr-1">{{ item.category_icon }}</span>
                {{ item.category_name }}
                <span
                  v-if="item.is_debt_focus"
                  class="ml-2 border-2 border-line bg-fatigued px-2 py-[1px] text-[10px] font-bold uppercase text-fg"
                >
                  fokus
                </span>
              </p>
              <p class="font-mono text-xs text-muted">
                {{ formatRp(item.spent_amount) }} / {{ formatRp(item.allocated_amount) }}
              </p>
            </div>
            <span
              :class="['border-2 border-line px-2 py-[2px] text-[10px] font-bold uppercase', statusClass[item.status]]"
              :data-testid="`budget-item-${item.category_name}-status`"
            >
              {{ statusLabel[item.status] }}
            </span>
          </div>

          <div class="h-3 w-full overflow-hidden border-2 border-line bg-bg">
            <div
              :class="['h-full transition-all motion-reduce:transition-none', barClass[item.status]]"
              :style="{ width: Math.min(item.percentage_used, 100) + '%' }"
            />
          </div>

          <div class="flex items-center justify-between text-xs">
            <span class="font-mono font-semibold">{{ item.percentage_used.toFixed(1) }}% dipakai</span>
            <span
              :class="item.remaining < 0 ? 'font-mono font-bold text-fatigued' : 'font-mono text-muted'"
            >
              sisa {{ formatRp(item.remaining) }}
            </span>
          </div>

          <p v-if="item.coaching" class="text-xs text-muted">{{ item.coaching }}</p>
        </article>
        </div>
      </div>
    </template>

    <button
      type="button"
      class="self-start border-2 border-line bg-surface px-3 py-2 text-xs font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
      @click="router.push({ name: 'transactions' })"
    >
      → catat transaksi
    </button>
  </section>
</template>
