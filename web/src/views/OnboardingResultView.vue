<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { useOnboardingStore } from '@/stores/onboarding'
import { PROGRAM_LABELS } from '@/api/onboarding'

const router = useRouter()
const store = useOnboardingStore()

const plan = computed(() => store.lastPlan)

function formatRp(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}

const summaryRows = computed(() => {
  if (!plan.value) return []
  return [
    { key: 'kebutuhan', label: 'Kebutuhan', bucket: plan.value.summary.kebutuhan },
    { key: 'utang', label: 'Utang', bucket: plan.value.summary.utang },
    { key: 'keinginan', label: 'Keinginan', bucket: plan.value.summary.keinginan },
    { key: 'tabungan', label: 'Tabungan', bucket: plan.value.summary.tabungan },
  ]
})
</script>

<template>
  <section
    class="mx-auto flex max-w-mobile flex-col gap-6 px-6 py-10"
    data-testid="onboarding-result"
  >
    <template v-if="plan">
      <header class="space-y-1">
        <p class="text-xs uppercase tracking-[0.18em] text-muted">Rekomendasi program</p>
        <h1 class="font-display text-3xl font-semibold" data-testid="result-program">
          {{ PROGRAM_LABELS[plan.program] }}
        </h1>
        <p class="text-sm text-muted">Periode {{ plan.period }}</p>
      </header>

      <div class="rounded-card border border-line bg-surface p-4">
        <p class="mb-2 text-xs uppercase tracking-wider text-muted">Pemasukan total</p>
        <p class="font-mono text-2xl" data-testid="result-total-income">
          <span class="text-saffron">Rp </span>
          {{ plan.total_income.toLocaleString('id-ID') }}
        </p>
      </div>

      <div class="rounded-card border border-line bg-surface p-4" data-testid="result-summary">
        <p class="mb-3 text-xs uppercase tracking-wider text-muted">Ringkasan</p>
        <dl class="space-y-2 font-mono text-sm">
          <div
            v-for="row in summaryRows"
            :key="row.key"
            class="flex items-center justify-between"
            :data-testid="`result-bucket-${row.key}`"
          >
            <dt class="text-muted">{{ row.label }}</dt>
            <dd class="text-right">
              {{ formatRp(row.bucket.amount) }}
              <span class="ml-2 text-xs text-muted">{{ row.bucket.percentage.toFixed(2) }}%</span>
            </dd>
          </div>
        </dl>
      </div>

      <div class="rounded-card border border-line bg-surface p-4" data-testid="result-items">
        <p class="mb-3 text-xs uppercase tracking-wider text-muted">Alokasi per kategori</p>
        <ul class="space-y-2 text-sm">
          <li
            v-for="item in plan.items"
            :key="item.category_id"
            class="flex items-center justify-between border-b border-line/40 pb-2 last:border-b-0"
          >
            <span>
              <span class="mr-2">{{ item.icon }}</span>
              {{ item.category_name }}
              <span
                v-if="item.is_debt_focus"
                class="ml-2 rounded-full bg-fatigued/20 px-2 py-[1px] text-[10px] uppercase tracking-wider text-fatigued"
              >
                fokus
              </span>
            </span>
            <span class="font-mono text-sm">{{ formatRp(item.allocated_amount) }}</span>
          </li>
        </ul>
      </div>

      <p
        v-if="plan.warning"
        class="rounded-card border border-warning/40 bg-warning/10 p-3 text-sm text-warning"
        data-testid="result-warning"
      >
        {{ plan.warning }}
      </p>

      <div class="space-y-2">
        <!-- Primary: finish onboarding and go to the live budget dashboard. -->
        <button
          type="button"
          class="w-full rounded-card bg-saffron py-3 font-semibold text-bg"
          data-testid="result-continue"
          @click="router.push({ name: 'budget' })"
        >
          Lanjut ke budget
        </button>
        <!-- Secondary: tweak the answers (preserved) and regenerate. -->
        <button
          type="button"
          class="w-full rounded-card border border-line py-3 text-sm text-muted hover:border-saffron hover:text-saffron"
          data-testid="result-restart"
          @click="router.push({ name: 'onboarding' })"
        >
          Ubah jawaban
        </button>
      </div>
    </template>

    <template v-else>
      <p class="text-sm text-muted" data-testid="result-empty">
        Belum ada hasil onboarding di sesi ini.
      </p>
      <button
        type="button"
        class="rounded-card bg-saffron py-3 font-semibold text-bg"
        @click="router.push({ name: 'onboarding' })"
      >
        Mulai dari awal
      </button>
    </template>
  </section>
</template>
