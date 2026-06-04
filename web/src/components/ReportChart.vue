<script setup lang="ts">
import { computed } from 'vue'

export interface ReportBar {
  category_name: string
  category_icon?: string
  amount: number
}

const props = defineProps<{
  rows: ReportBar[]
}>()

// Bars are sorted by the parent (spend desc), but scale to the largest spend so
// the top bar fills the track and the rest read as proportions of it.
const scaleMax = computed(() => {
  let max = 0
  for (const r of props.rows) max = Math.max(max, r.amount)
  return max > 0 ? max : 1
})

function pct(value: number): number {
  return Math.max(0, Math.min(100, (value / scaleMax.value) * 100))
}

function formatRp(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}
</script>

<template>
  <section class="space-y-3" data-testid="reports-chart">
    <p class="text-xs uppercase tracking-wider text-muted">Distribusi per kategori</p>

    <div class="space-y-3 rounded-card border border-line bg-surface p-4">
      <div
        v-for="row in rows"
        :key="row.category_name"
        class="space-y-1"
        :data-testid="`reports-bar-${row.category_name}`"
      >
        <div class="flex items-center justify-between gap-2 text-xs">
          <span class="min-w-0 truncate">
            <span v-if="row.category_icon" class="mr-1">{{ row.category_icon }}</span
            >{{ row.category_name }}
          </span>
          <span class="shrink-0 font-mono text-[11px] text-muted">
            {{ formatRp(row.amount) }}
          </span>
        </div>

        <!-- Neutral track, saffron-free status-neutral fill: this is reporting,
             not a fatigue state, so we use a muted brand-neutral fill. -->
        <div class="relative h-3 w-full overflow-hidden rounded-full bg-bg">
          <div
            class="report-fill h-full rounded-full bg-fg/70"
            :style="{ width: pct(row.amount) + '%' }"
            :data-testid="`reports-fill-${row.category_name}`"
          />
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.report-fill {
  transition: width 600ms cubic-bezier(0.22, 1, 0.36, 1);
}

@media (prefers-reduced-motion: reduce) {
  .report-fill {
    transition: none;
  }
}
</style>
