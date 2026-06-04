<script setup lang="ts">
import { computed } from 'vue'
import type { BudgetItem, FatigueStatus } from '@/api/budget'

const props = defineProps<{
  items: BudgetItem[]
}>()

const statusOrder: Record<FatigueStatus, number> = { fatigued: 0, warning: 1, fresh: 2 }

// Sort most-stressed first so over-budget bars sit at the top of the chart.
const rows = computed(() =>
  [...props.items].sort((a, b) => statusOrder[a.status] - statusOrder[b.status]),
)

// The track is scaled to the largest of (allocated, spent) across every row, so an
// over-budget bar can visibly blow past its allocated marker without clipping.
const scaleMax = computed(() => {
  let max = 0
  for (const item of props.items) {
    max = Math.max(max, item.allocated_amount, item.spent_amount)
  }
  return max > 0 ? max : 1
})

function pct(value: number): number {
  return Math.max(0, Math.min(100, (value / scaleMax.value) * 100))
}

const fillClass: Record<FatigueStatus, string> = {
  fresh: 'bg-fresh',
  warning: 'bg-warning',
  fatigued: 'bg-fatigued',
}

function formatRp(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}
</script>

<template>
  <section class="space-y-3" data-testid="budget-compare-chart">
    <div class="flex items-baseline justify-between">
      <p class="text-xs uppercase tracking-wider text-muted">Budget vs realisasi</p>
      <!-- Legend -->
      <div class="flex items-center gap-3 text-[10px] text-muted">
        <span class="flex items-center gap-1">
          <span class="inline-block h-2 w-2 rounded-full bg-fresh" />
          fresh
        </span>
        <span class="flex items-center gap-1">
          <span class="inline-block h-2 w-2 rounded-full bg-warning" />
          warning
        </span>
        <span class="flex items-center gap-1">
          <span class="inline-block h-2 w-2 rounded-full bg-fatigued" />
          fatigued
        </span>
      </div>
    </div>

    <p class="flex items-center gap-1.5 text-[10px] text-muted">
      <span class="inline-block h-3 w-[2px] bg-fg/70" />
      penanda = batas alokasi
    </p>

    <div class="space-y-3 rounded-card border border-line bg-surface p-4">
      <div
        v-for="item in rows"
        :key="item.id"
        class="space-y-1"
        :data-testid="`compare-row-${item.category_name}`"
      >
        <div class="flex items-center justify-between gap-2 text-xs">
          <span class="min-w-0 truncate">
            <span class="mr-1">{{ item.category_icon }}</span>{{ item.category_name }}
          </span>
          <span class="shrink-0 font-mono text-[11px] text-muted">
            {{ formatRp(item.spent_amount) }} / {{ formatRp(item.allocated_amount) }}
          </span>
        </div>

        <!-- Compare bar: neutral track, status-colored spent fill, allocated marker. -->
        <div class="relative h-3 w-full overflow-hidden rounded-full bg-bg">
          <div
            class="compare-fill h-full rounded-full"
            :class="fillClass[item.status]"
            :style="{ width: pct(item.spent_amount) + '%' }"
            :data-testid="`compare-fill-${item.category_name}`"
          />
          <!-- Allocated marker overlaid on top of the fill. -->
          <span
            class="pointer-events-none absolute top-0 h-full w-[2px] bg-fg/70"
            :style="{ left: pct(item.allocated_amount) + '%' }"
            :data-testid="`compare-marker-${item.category_name}`"
            aria-hidden="true"
          />
        </div>
      </div>
    </div>
  </section>
</template>

<style scoped>
.compare-fill {
  transition: width 600ms cubic-bezier(0.22, 1, 0.36, 1);
}

@media (prefers-reduced-motion: reduce) {
  .compare-fill {
    transition: none;
  }
}
</style>
