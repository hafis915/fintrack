<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { getCurrentBudget, NoPlanError, type CurrentBudget, type FatigueStatus } from '@/api/budget'
import { PROGRAM_LABELS } from '@/api/onboarding'

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

const statusClass: Record<FatigueStatus, string> = {
  fresh: 'text-fresh border-fresh/40 bg-fresh/10',
  warning: 'text-warning border-warning/40 bg-warning/10',
  fatigued: 'text-fatigued border-fatigued/40 bg-fatigued/10',
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
  <section class="mx-auto flex max-w-mobile flex-col gap-6 px-6 py-10" data-testid="budget-view">
    <header class="space-y-1">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Fatigue dashboard</p>
      <h1 class="font-display text-3xl font-semibold">Budget bulan ini</h1>
      <p v-if="budget" class="text-sm text-muted" data-testid="budget-period">
        Periode {{ budget.period }} · {{ PROGRAM_LABELS[budget.program] }}
      </p>
    </header>

    <p v-if="loading" class="font-mono text-sm text-muted">memuat…</p>

    <div
      v-else-if="noPlan"
      class="space-y-3 rounded-card border border-line bg-surface p-5 text-sm text-muted"
      data-testid="budget-no-plan"
    >
      <p>
        Belum ada budget untuk bulan ini. Selesaikan onboarding dulu — kami susun program yang pas.
      </p>
      <button
        type="button"
        class="w-full rounded-card bg-saffron py-2 font-semibold text-bg"
        @click="router.push({ name: 'onboarding' })"
      >
        Mulai onboarding
      </button>
    </div>

    <p v-else-if="errorMsg" class="text-sm text-fatigued" data-testid="budget-error">
      {{ errorMsg }}
    </p>

    <template v-else-if="budget">
      <!-- Summary -->
      <div
        class="space-y-2 rounded-card border border-line bg-surface p-4"
        data-testid="budget-summary"
      >
        <p class="text-xs uppercase tracking-wider text-muted">Ringkasan</p>
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

      <!-- Items -->
      <div class="space-y-3" data-testid="budget-items">
        <p class="text-xs uppercase tracking-wider text-muted">Per kategori</p>
        <article
          v-for="item in sortedItems"
          :key="item.id"
          class="space-y-2 rounded-card border border-line bg-surface p-4"
          :data-testid="`budget-item-${item.category_name}`"
        >
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0">
              <p class="text-sm">
                <span class="mr-1">{{ item.category_icon }}</span>
                {{ item.category_name }}
                <span
                  v-if="item.is_debt_focus"
                  class="ml-2 rounded-full bg-fatigued/20 px-2 py-[1px] text-[10px] uppercase tracking-wider text-fatigued"
                >
                  fokus
                </span>
              </p>
              <p class="font-mono text-xs text-muted">
                {{ formatRp(item.spent_amount) }} / {{ formatRp(item.allocated_amount) }}
              </p>
            </div>
            <span
              :class="['rounded-full border px-2 py-[2px] text-[10px] uppercase tracking-wider', statusClass[item.status]]"
              :data-testid="`budget-item-${item.category_name}-status`"
            >
              {{ statusLabel[item.status] }}
            </span>
          </div>

          <div class="h-1.5 w-full overflow-hidden rounded-full bg-bg">
            <div
              :class="['h-full transition-all', {
                'bg-fresh': item.status === 'fresh',
                'bg-warning': item.status === 'warning',
                'bg-fatigued': item.status === 'fatigued',
              }]"
              :style="{ width: Math.min(item.percentage_used, 100) + '%' }"
            />
          </div>

          <div class="flex items-center justify-between text-xs">
            <span class="font-mono text-muted">{{ item.percentage_used.toFixed(1) }}% dipakai</span>
            <span
              :class="item.remaining < 0 ? 'text-fatigued font-mono' : 'text-muted font-mono'"
            >
              sisa {{ formatRp(item.remaining) }}
            </span>
          </div>

          <p v-if="item.coaching" class="text-xs text-muted">{{ item.coaching }}</p>
        </article>
      </div>
    </template>

    <button
      type="button"
      class="text-xs text-muted hover:text-saffron"
      @click="router.push({ name: 'transactions' })"
    >
      → catat transaksi
    </button>
  </section>
</template>
