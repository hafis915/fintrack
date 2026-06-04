<script setup lang="ts">
import { computed } from 'vue'
import type { BudgetItem } from '@/api/budget'

const props = defineProps<{
  items: BudgetItem[]
}>()

interface Suggestion {
  item: BudgetItem
  overspend: number
  suggested: number
}

// A category needs attention if it spent past its allocation OR the fatigue model
// already flagged it. Over-budget categories are the actionable ones to cut.
const suggestions = computed<Suggestion[]>(() => {
  return props.items
    .filter((item) => item.spent_amount > item.allocated_amount || item.status === 'fatigued')
    .map((item) => {
      const overspend = Math.max(0, item.spent_amount - item.allocated_amount)
      // If genuinely over budget, suggest trimming the overspend back to plan.
      // If only "fatigued" but not yet over, suggest a gentle ~15% portion.
      const suggested = overspend > 0 ? overspend : Math.round(item.spent_amount * 0.15)
      return { item, overspend, suggested }
    })
    .sort((a, b) => b.overspend - a.overspend)
    .slice(0, 3)
})

function formatRp(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}
</script>

<template>
  <section
    class="space-y-3 rounded-card border-2 border-line bg-surface p-4 shadow-brutal"
    data-testid="budget-recommendations"
  >
    <span class="inline-block border-2 border-line bg-saffron px-2 py-[2px] text-[10px] font-bold uppercase text-fg">Saran bulan ini</span>

    <p
      v-if="suggestions.length === 0"
      class="text-sm font-semibold text-fresh"
      data-testid="reco-on-track"
    >
      Mantap — semua kategori masih on track bulan ini.
    </p>

    <ul v-else class="space-y-2">
      <li
        v-for="{ item, overspend, suggested } in suggestions"
        :key="item.id"
        class="text-sm leading-snug"
        :data-testid="`reco-item-${item.category_name}`"
      >
        <span class="mr-1">{{ item.category_icon }}</span>
        <span class="font-bold">{{ item.category_name }}</span>
        <template v-if="overspend > 0">
          — lebih
          <span class="font-mono text-fatigued">{{ formatRp(overspend) }}</span
          >, coba kurangi ~<span class="font-mono">{{ formatRp(suggested) }}</span>
        </template>
        <template v-else>
          — mulai terasa berat, coba kurangi ~<span class="font-mono">{{
            formatRp(suggested)
          }}</span>
        </template>
      </li>
    </ul>
  </section>
</template>
