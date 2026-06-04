<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { listTransactions, type Transaction } from '@/api/transactions'
import {
  currentYearMonth,
  monthLabel,
  monthRange,
  monthSlug,
  shiftMonth,
  type YearMonth,
} from '@/utils/period'
import ReportChart, { type ReportBar } from '@/components/ReportChart.vue'

const router = useRouter()

const selected = ref<YearMonth>(currentYearMonth())
const transactions = ref<Transaction[]>([])
const loading = ref(true)
const errorMsg = ref<string | null>(null)

const label = computed(() => monthLabel(selected.value.year, selected.value.month))

function formatRp(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}

// Aggregate the loaded month's transactions by category. Uncategorized rows
// (no category_name) fall back to a stable bucket so totals always reconcile.
interface CategoryAgg {
  category_name: string
  category_icon?: string
  amount: number
  count: number
}

const rows = computed<CategoryAgg[]>(() => {
  const byCat = new Map<string, CategoryAgg>()
  for (const t of transactions.value) {
    const name = t.category_name ?? 'Lainnya'
    const existing = byCat.get(name)
    if (existing) {
      existing.amount += t.amount
      existing.count += 1
    } else {
      byCat.set(name, {
        category_name: name,
        category_icon: t.category_icon,
        amount: t.amount,
        count: 1,
      })
    }
  }
  return [...byCat.values()].sort((a, b) => b.amount - a.amount)
})

const totalSpent = computed(() => rows.value.reduce((sum, r) => sum + r.amount, 0))

const chartRows = computed<ReportBar[]>(() =>
  rows.value.map((r) => ({
    category_name: r.category_name,
    category_icon: r.category_icon,
    amount: r.amount,
  })),
)

function pctOfTotal(amount: number): number {
  const total = totalSpent.value
  return total > 0 ? (amount / total) * 100 : 0
}

async function refresh() {
  loading.value = true
  errorMsg.value = null
  try {
    const { from, to } = monthRange(selected.value.year, selected.value.month)
    const list = await listTransactions({ from, to, limit: 1000 })
    transactions.value = list.items
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

function goPrev() {
  selected.value = shiftMonth(selected.value, -1)
}

function goNext() {
  selected.value = shiftMonth(selected.value, 1)
}

// RFC4180 field escaping: wrap in quotes and double internal quotes when the
// field contains a comma, quote, or newline.
function csvField(value: string): string {
  if (/[",\n\r]/.test(value)) {
    return '"' + value.replace(/"/g, '""') + '"'
  }
  return value
}

// Build the CSV client-side from the in-memory month rows (the axios JWT is
// already applied to the data fetch — no second request, no auth on a raw href).
function exportCsv() {
  const header = 'tanggal,kategori,merchant,jumlah,catatan'
  const lines = transactions.value.map((t) => {
    const tanggal = t.transacted_at
    const kategori = t.category_name ?? 'Lainnya'
    const merchant = t.merchant ?? ''
    const jumlah = String(Math.trunc(t.amount)) // plain integer, no separators
    const catatan = t.note ?? ''
    return [tanggal, kategori, merchant, jumlah, catatan].map(csvField).join(',')
  })
  const csv = [header, ...lines].join('\r\n')

  const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `fintrack-${monthSlug(selected.value.year, selected.value.month)}.csv`
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

watch(selected, refresh)
onMounted(refresh)
</script>

<template>
  <!-- Mobile-first single column; widens to a two-column grid at lg. -->
  <section
    class="mx-auto flex max-w-mobile flex-col gap-6 px-6 py-10 lg:max-w-5xl"
    data-testid="reports-view"
  >
    <header class="space-y-1">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Laporan</p>
      <h1 class="font-display text-3xl font-semibold">Ringkasan bulanan</h1>
    </header>

    <!-- Month filter + export -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-2">
        <button
          type="button"
          class="grid h-9 w-9 place-items-center rounded-card border border-line bg-surface text-muted transition-colors hover:text-fg"
          aria-label="Bulan sebelumnya"
          data-testid="reports-month-prev"
          @click="goPrev"
        >
          ‹
        </button>
        <span
          class="min-w-[8.5rem] text-center font-display text-lg font-semibold"
          data-testid="reports-month-label"
        >
          {{ label }}
        </span>
        <button
          type="button"
          class="grid h-9 w-9 place-items-center rounded-card border border-line bg-surface text-muted transition-colors hover:text-fg"
          aria-label="Bulan berikutnya"
          data-testid="reports-month-next"
          @click="goNext"
        >
          ›
        </button>
      </div>

      <button
        type="button"
        class="rounded-card border border-line bg-surface px-3 py-2 text-xs font-semibold text-fg transition-colors hover:border-saffron disabled:opacity-40"
        data-testid="reports-export"
        :disabled="loading || transactions.length === 0"
        @click="exportCsv"
      >
        Export CSV
      </button>
    </div>

    <p v-if="loading" class="font-mono text-sm text-muted">memuat…</p>

    <p v-else-if="errorMsg" class="text-sm text-fatigued" data-testid="reports-error">
      {{ errorMsg }}
    </p>

    <div
      v-else-if="transactions.length === 0"
      class="rounded-card border border-line bg-surface p-6 text-center text-sm text-muted"
      data-testid="reports-empty"
    >
      Belum ada transaksi di {{ label }}.
    </div>

    <template v-else>
      <!-- Hero: total spent (mono digits + saffron Rp) -->
      <div
        class="space-y-1 rounded-card border border-line bg-surface p-5"
        data-testid="reports-total"
      >
        <p class="text-xs uppercase tracking-wider text-muted">Total dipakai</p>
        <p class="font-mono text-4xl font-semibold tabular-nums">
          <span class="text-saffron">Rp</span>
          {{ totalSpent.toLocaleString('id-ID') }}
        </p>
        <p class="text-xs text-muted">{{ transactions.length }} transaksi · {{ rows.length }} kategori</p>
      </div>

      <!-- Two-column on desktop: table beside chart. Stacked on mobile. -->
      <div class="grid gap-6 lg:grid-cols-2">
        <!-- Table -->
        <div class="overflow-hidden rounded-card border border-line bg-surface">
          <table class="w-full text-sm" data-testid="reports-table">
            <thead>
              <tr class="border-b border-line text-left text-[11px] uppercase tracking-wider text-muted">
                <th class="px-4 py-3 font-medium">Kategori</th>
                <th class="px-4 py-3 text-right font-medium">Dipakai</th>
                <th class="px-4 py-3 text-right font-medium">%</th>
                <th class="px-4 py-3 text-right font-medium">Transaksi</th>
              </tr>
            </thead>
            <tbody>
              <tr
                v-for="row in rows"
                :key="row.category_name"
                class="border-b border-line/60 last:border-0"
                :data-testid="`reports-row-${row.category_name}`"
              >
                <td class="px-4 py-3">
                  <span v-if="row.category_icon" class="mr-1">{{ row.category_icon }}</span
                  >{{ row.category_name }}
                </td>
                <td class="px-4 py-3 text-right font-mono tabular-nums">
                  {{ formatRp(row.amount) }}
                </td>
                <td class="px-4 py-3 text-right font-mono tabular-nums text-muted">
                  {{ pctOfTotal(row.amount).toFixed(1) }}%
                </td>
                <td class="px-4 py-3 text-right font-mono tabular-nums text-muted">
                  {{ row.count }}
                </td>
              </tr>
            </tbody>
            <tfoot>
              <tr class="border-t border-line font-semibold" data-testid="reports-row-total">
                <td class="px-4 py-3">Total</td>
                <td class="px-4 py-3 text-right font-mono tabular-nums">
                  {{ formatRp(totalSpent) }}
                </td>
                <td class="px-4 py-3 text-right font-mono tabular-nums text-muted">100%</td>
                <td class="px-4 py-3 text-right font-mono tabular-nums text-muted">
                  {{ transactions.length }}
                </td>
              </tr>
            </tfoot>
          </table>
        </div>

        <!-- Chart -->
        <ReportChart :rows="chartRows" />
      </div>
    </template>

    <button
      type="button"
      class="text-xs text-muted transition-colors hover:text-saffron"
      @click="router.push({ name: 'transactions' })"
    >
      → catat transaksi
    </button>
  </section>
</template>
