<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listCategories, type Category } from '@/api/categories'
import {
  createTransaction,
  deleteTransaction,
  listTransactions,
  updateTransaction,
  type Transaction,
} from '@/api/transactions'
import AddCategoryInline from '@/components/AddCategoryInline.vue'

const router = useRouter()

const transactions = ref<Transaction[]>([])
const total = ref(0)
const categories = ref<Category[]>([])
const loading = ref(true)
const errorMsg = ref<string | null>(null)

// Edit-in-place state. Only one row is editable at a time.
const editingId = ref<string | null>(null)
const editAmount = ref<number>(0)
const editAmountDisplay = ref<string>('') // formatted "2.000" shown in the input
const editNote = ref<string>('')

// New-transaction form state. Shown inline at top of the page; submit
// adds to the list without a route change.
const newCategoryId = ref<string>('')
const newAmount = ref<number>(0)
const newAmountDisplay = ref<string>('') // formatted "2.000" shown in the input
const newNote = ref<string>('')
// Date the transaction happened — defaults to today, but the user can log a
// past entry (yesterday, last month). Stored as YYYY-MM-DD (local).
const newDate = ref<string>(toDateInputValue(new Date()))
const todayStr = toDateInputValue(new Date())

// --- amount formatting: show Indonesian thousand separators ("." divider) so
// 2.000 vs 20.000 is unambiguous, while the API still receives a raw integer.
function digitsToNumber(raw: string): number {
  const digits = raw.replace(/\D/g, '')
  return digits ? parseInt(digits, 10) : 0
}
function formatThousands(n: number): string {
  return n > 0 ? n.toLocaleString('id-ID') : ''
}
function onNewAmountInput(e: Event) {
  newAmount.value = digitsToNumber((e.target as HTMLInputElement).value)
  newAmountDisplay.value = formatThousands(newAmount.value)
}
function onEditAmountInput(e: Event) {
  editAmount.value = digitsToNumber((e.target as HTMLInputElement).value)
  editAmountDisplay.value = formatThousands(editAmount.value)
}

// --- date helpers
function toDateInputValue(d: Date): string {
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  return `${y}-${m}-${day}`
}
function dateToRFC3339(dateStr: string): string {
  const [y, m, d] = dateStr.split('-').map(Number)
  // Local noon: keeps the entry on the user's intended calendar day regardless
  // of timezone when the backend buckets it by month.
  return new Date(y, m - 1, d, 12, 0, 0).toISOString()
}

// Month filter state. `selectedMonth` is the first day of the active month at
// local midnight; prev/next shift it by one month. Bounds are derived inline.
const now = new Date()
const selectedMonth = ref<Date>(new Date(now.getFullYear(), now.getMonth(), 1))

const monthLabel = computed(() =>
  selectedMonth.value.toLocaleDateString('id-ID', { month: 'long', year: 'numeric' }),
)

function prevMonth() {
  const d = selectedMonth.value
  selectedMonth.value = new Date(d.getFullYear(), d.getMonth() - 1, 1)
  refresh()
}

function nextMonth() {
  const d = selectedMonth.value
  selectedMonth.value = new Date(d.getFullYear(), d.getMonth() + 1, 1)
  refresh()
}

const categoriesById = computed(() => {
  const m: Record<string, Category> = {}
  for (const c of categories.value) m[c.id] = c
  return m
})

async function refresh() {
  loading.value = true
  errorMsg.value = null
  try {
    // [from, to) bounds: first day of selected month → first day of next month.
    const d = selectedMonth.value
    const from = new Date(d.getFullYear(), d.getMonth(), 1).toISOString()
    const to = new Date(d.getFullYear(), d.getMonth() + 1, 1).toISOString()
    const list = await listTransactions({ from, to, limit: 100 })
    transactions.value = list.items
    total.value = list.total
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  try {
    const cats = await listCategories()
    categories.value = cats
    if (cats.length > 0) newCategoryId.value = cats[0].id
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  }
  await refresh()
})

// A newly-created category joins the dropdown and becomes the active selection,
// so the user can log against it immediately.
function onCategoryCreated(c: Category) {
  categories.value = [...categories.value, c]
  newCategoryId.value = c.id
}

async function onCreate() {
  if (!newCategoryId.value || newAmount.value <= 0) {
    errorMsg.value = 'Pilih kategori + isi jumlah dulu.'
    return
  }
  if (!newDate.value) {
    errorMsg.value = 'Pilih tanggal transaksi dulu.'
    return
  }
  errorMsg.value = null
  try {
    await createTransaction({
      category_id: newCategoryId.value,
      amount: newAmount.value,
      note: newNote.value || undefined,
      transacted_at: dateToRFC3339(newDate.value),
    })
    // Jump the month filter to the entry's month so a past-dated transaction
    // is still visible after it's saved (otherwise it'd be filtered out).
    const [cy, cm] = newDate.value.split('-').map(Number)
    selectedMonth.value = new Date(cy, cm - 1, 1)
    newAmount.value = 0
    newAmountDisplay.value = ''
    newNote.value = ''
    await refresh()
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  }
}

function startEdit(tx: Transaction) {
  editingId.value = tx.id
  editAmount.value = tx.amount
  editAmountDisplay.value = formatThousands(tx.amount)
  editNote.value = tx.note ?? ''
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit(tx: Transaction) {
  // Guard amount before hitting the API. A cleared number field serializes
  // as "" and the backend rejects it as invalid_json (bug #3) — mirror the
  // create path and surface a friendly inline message instead.
  const amountValue = Number(editAmount.value)
  if (!Number.isFinite(amountValue) || amountValue <= 0) {
    editAmount.value = amountValue
    errorMsg.value = 'Jumlah harus lebih dari 0.'
    return
  }
  editAmount.value = amountValue

  errorMsg.value = null
  try {
    await updateTransaction(tx.id, {
      amount: editAmount.value,
      note: editNote.value,
    })
    editingId.value = null
    await refresh()
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  }
}

async function onDelete(tx: Transaction) {
  errorMsg.value = null
  try {
    await deleteTransaction(tx.id)
    await refresh()
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  }
}

function formatRp(n: number): string {
  return 'Rp ' + n.toLocaleString('id-ID')
}

function formatDate(iso: string): string {
  const d = new Date(iso)
  return d.toLocaleDateString('id-ID', { day: '2-digit', month: 'short', year: 'numeric' })
}
</script>

<template>
  <section
    class="mx-auto flex w-full max-w-mobile flex-col gap-6 px-6 py-10 lg:max-w-none lg:px-10"
    data-testid="transactions-view"
  >
    <header class="space-y-1">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Catatan</p>
      <h1 class="font-display text-3xl font-semibold lg:text-4xl">Transaksi</h1>
      <p class="text-sm text-muted" data-testid="tx-total">
        {{ total }} transaksi
      </p>
    </header>

    <!-- Desktop: the create form sits in a sticky left rail, the list takes the
         wide right column. Stacks on mobile. -->
    <div class="flex flex-col gap-6 lg:grid lg:grid-cols-3 lg:items-start lg:gap-6">
    <!-- Left rail: create + add category. -->
    <div class="space-y-4 lg:col-span-1 lg:sticky lg:top-6">
    <!-- Inline create form -->
    <form
      class="space-y-3 rounded-card border-2 border-line bg-surface p-4 shadow-brutal"
      data-testid="tx-new-form"
      novalidate
      @submit.prevent="onCreate"
    >
      <p class="text-xs uppercase tracking-wider text-muted">Catat transaksi baru</p>

      <select
        v-model="newCategoryId"
        data-testid="tx-new-category"
        class="w-full rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
      >
        <option v-for="c in categories" :key="c.id" :value="c.id">
          {{ c.icon }} {{ c.name }}
        </option>
      </select>

      <!-- Add a category on the fly when the one you need isn't listed. -->
      <AddCategoryInline show-type-picker default-type="variable" @created="onCategoryCreated" />

      <!-- Date the transaction happened (default today; past dates allowed) -->
      <label class="block">
        <span class="mb-1 block text-[10px] uppercase tracking-wider text-muted">Tanggal</span>
        <input
          v-model="newDate"
          type="date"
          :max="todayStr"
          data-testid="tx-new-date"
          class="w-full rounded border border-line bg-bg px-3 py-2 font-mono text-sm focus:border-saffron focus:outline-none"
        />
      </label>

      <!-- Amount with a saffron Rp prefix + live "." thousand separators -->
      <div
        class="flex items-center gap-2 rounded border border-line bg-bg px-3 focus-within:border-saffron"
      >
        <span class="font-mono text-sm text-saffron">Rp</span>
        <input
          :value="newAmountDisplay"
          type="text"
          inputmode="numeric"
          placeholder="0"
          data-testid="tx-new-amount"
          class="w-full bg-transparent py-2 text-right font-mono text-sm focus:outline-none"
          @input="onNewAmountInput"
        />
      </div>
      <input
        v-model="newNote"
        type="text"
        placeholder="Catatan (opsional)"
        data-testid="tx-new-note"
        class="w-full rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
      />

      <button
        type="submit"
        data-testid="tx-new-submit"
        class="flex w-full items-center justify-center gap-2 rounded-card border-2 border-line bg-saffron py-3 text-sm font-bold uppercase text-fg shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
      >
        <span aria-hidden="true">＋</span>
        <span>Tambah transaksi</span>
      </button>
    </form>

    <p v-if="errorMsg" class="text-sm text-fatigued" data-testid="tx-error">{{ errorMsg }}</p>
    </div>
    <!-- /left rail -->

    <!-- Right column: month filter + list. -->
    <div class="space-y-6 lg:col-span-2">
    <!-- Month filter -->
    <div
      class="flex items-center justify-between rounded-card border-2 border-line bg-surface px-3 py-2 shadow-brutal-sm"
      data-testid="tx-month-filter"
    >
      <button
        type="button"
        data-testid="tx-month-prev"
        aria-label="Bulan sebelumnya"
        class="rounded px-2 py-1 text-muted hover:text-saffron"
        @click="prevMonth"
      >
        ‹
      </button>
      <span
        class="font-mono text-sm capitalize"
        data-testid="tx-month-label"
      >
        {{ monthLabel }}
      </span>
      <button
        type="button"
        data-testid="tx-month-next"
        aria-label="Bulan berikutnya"
        class="rounded px-2 py-1 text-muted hover:text-saffron"
        @click="nextMonth"
      >
        ›
      </button>
    </div>

    <!-- List -->
    <div v-if="loading" class="font-mono text-sm text-muted">memuat…</div>

    <div
      v-else-if="transactions.length === 0"
      class="rounded-card border border-line bg-surface p-6 text-center text-sm text-muted"
      data-testid="tx-empty"
    >
      Belum ada transaksi.
    </div>

    <ul v-else class="grid gap-2 lg:grid-cols-2" data-testid="tx-list">
      <li
        v-for="tx in transactions"
        :key="tx.id"
        class="rounded-card border-2 border-line bg-surface p-3 shadow-brutal-sm"
        :data-testid="`tx-row-${tx.id}`"
      >
        <template v-if="editingId === tx.id">
          <div class="space-y-2">
            <div
              class="flex items-center gap-2 rounded border border-line bg-bg px-3 focus-within:border-saffron"
            >
              <span class="font-mono text-sm text-saffron">Rp</span>
              <input
                :value="editAmountDisplay"
                type="text"
                inputmode="numeric"
                data-testid="tx-edit-amount"
                class="w-full bg-transparent py-2 text-right font-mono text-sm focus:outline-none"
                @input="onEditAmountInput"
              />
            </div>
            <input
              v-model="editNote"
              type="text"
              data-testid="tx-edit-note"
              class="w-full rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
            />
            <div class="flex gap-2">
              <button
                type="button"
                data-testid="tx-edit-save"
                class="flex-1 rounded bg-saffron py-1 text-sm font-semibold text-bg"
                @click="saveEdit(tx)"
              >
                Simpan
              </button>
              <button
                type="button"
                class="flex-1 rounded border border-line py-1 text-sm text-muted"
                @click="cancelEdit"
              >
                Batal
              </button>
            </div>
          </div>
        </template>

        <template v-else>
          <div class="flex items-start justify-between gap-3">
            <div class="min-w-0 flex-1">
              <p class="text-sm" :data-testid="`tx-row-${tx.id}-name`">
                <span class="mr-1">{{ categoriesById[tx.category_id]?.icon ?? tx.category_icon }}</span>
                {{ tx.merchant || tx.category_name || categoriesById[tx.category_id]?.name || '—' }}
              </p>
              <p
                v-if="tx.merchant"
                class="text-[10px] uppercase tracking-wider text-muted"
              >
                {{ tx.category_name ?? categoriesById[tx.category_id]?.name ?? '—' }}
              </p>
              <p v-if="tx.note" class="truncate text-xs text-muted">{{ tx.note }}</p>
              <p class="text-[10px] uppercase tracking-wider text-muted">{{ formatDate(tx.transacted_at) }}</p>
            </div>
            <div class="text-right">
              <p class="font-mono text-sm" :data-testid="`tx-row-${tx.id}-amount`">
                {{ formatRp(tx.amount) }}
              </p>
              <div class="mt-2 flex justify-end gap-2">
                <button
                  type="button"
                  class="flex items-center gap-1 rounded-card border-2 border-line bg-surface px-2.5 py-1 text-[11px] font-bold uppercase shadow-brutal-sm transition-colors hover:bg-saffron active:translate-x-[1px] active:translate-y-[1px] active:shadow-none motion-reduce:transform-none"
                  :data-testid="`tx-row-${tx.id}-edit`"
                  @click="startEdit(tx)"
                >
                  <span aria-hidden="true">✏️</span><span>Edit</span>
                </button>
                <button
                  type="button"
                  class="flex items-center gap-1 rounded-card border-2 border-line bg-surface px-2.5 py-1 text-[11px] font-bold uppercase shadow-brutal-sm transition-colors hover:bg-fatigued hover:text-fg active:translate-x-[1px] active:translate-y-[1px] active:shadow-none motion-reduce:transform-none"
                  :data-testid="`tx-row-${tx.id}-delete`"
                  @click="onDelete(tx)"
                >
                  <span aria-hidden="true">🗑️</span><span>Hapus</span>
                </button>
              </div>
            </div>
          </div>
        </template>
      </li>
    </ul>
    </div>
    <!-- /right column -->
    </div>
    <!-- /grid -->

    <button
      type="button"
      class="text-xs text-muted hover:text-saffron"
      @click="router.push({ name: 'home' })"
    >
      ← kembali ke beranda
    </button>
  </section>
</template>
