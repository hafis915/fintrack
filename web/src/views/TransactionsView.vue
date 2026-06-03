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

const router = useRouter()

const transactions = ref<Transaction[]>([])
const total = ref(0)
const categories = ref<Category[]>([])
const loading = ref(true)
const errorMsg = ref<string | null>(null)

// Edit-in-place state. Only one row is editable at a time.
const editingId = ref<string | null>(null)
const editAmount = ref<number>(0)
const editNote = ref<string>('')

// New-transaction form state. Shown inline at top of the page; submit
// adds to the list without a route change.
const newCategoryId = ref<string>('')
const newAmount = ref<number>(0)
const newNote = ref<string>('')

const categoriesById = computed(() => {
  const m: Record<string, Category> = {}
  for (const c of categories.value) m[c.id] = c
  return m
})

async function refresh() {
  loading.value = true
  errorMsg.value = null
  try {
    const list = await listTransactions({ limit: 100 })
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

async function onCreate() {
  if (!newCategoryId.value || newAmount.value <= 0) {
    errorMsg.value = 'Pilih kategori + isi jumlah dulu.'
    return
  }
  errorMsg.value = null
  try {
    await createTransaction({
      category_id: newCategoryId.value,
      amount: newAmount.value,
      note: newNote.value || undefined,
      transacted_at: new Date().toISOString(),
    })
    newAmount.value = 0
    newNote.value = ''
    await refresh()
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  }
}

function startEdit(tx: Transaction) {
  editingId.value = tx.id
  editAmount.value = tx.amount
  editNote.value = tx.note ?? ''
}

function cancelEdit() {
  editingId.value = null
}

async function saveEdit(tx: Transaction) {
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
    class="mx-auto flex max-w-mobile flex-col gap-6 px-6 py-10"
    data-testid="transactions-view"
  >
    <header class="space-y-1">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Catatan</p>
      <h1 class="font-display text-3xl font-semibold">Transaksi</h1>
      <p class="text-sm text-muted" data-testid="tx-total">
        {{ total }} {{ total === 1 ? 'transaksi' : 'transaksi' }}
      </p>
    </header>

    <!-- Inline create form -->
    <form
      class="space-y-3 rounded-card border border-line bg-surface p-4"
      data-testid="tx-new-form"
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

      <div class="flex gap-2">
        <input
          v-model.number="newAmount"
          type="number"
          min="0"
          step="1000"
          placeholder="Jumlah (Rp)"
          data-testid="tx-new-amount"
          class="flex-1 rounded border border-line bg-bg px-3 py-2 text-right font-mono text-sm focus:border-saffron focus:outline-none"
        />
        <input
          v-model="newNote"
          type="text"
          placeholder="Catatan (opsional)"
          data-testid="tx-new-note"
          class="flex-1 rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
        />
      </div>

      <button
        type="submit"
        data-testid="tx-new-submit"
        class="w-full rounded-card bg-saffron py-2 text-sm font-semibold text-bg"
      >
        Tambah
      </button>
    </form>

    <p v-if="errorMsg" class="text-sm text-fatigued" data-testid="tx-error">{{ errorMsg }}</p>

    <!-- List -->
    <div v-if="loading" class="font-mono text-sm text-muted">memuat…</div>

    <div
      v-else-if="transactions.length === 0"
      class="rounded-card border border-line bg-surface p-6 text-center text-sm text-muted"
      data-testid="tx-empty"
    >
      Belum ada transaksi.
    </div>

    <ul v-else class="space-y-2" data-testid="tx-list">
      <li
        v-for="tx in transactions"
        :key="tx.id"
        class="rounded-card border border-line bg-surface p-3"
        :data-testid="`tx-row-${tx.id}`"
      >
        <template v-if="editingId === tx.id">
          <div class="space-y-2">
            <input
              v-model.number="editAmount"
              type="number"
              min="1"
              data-testid="tx-edit-amount"
              class="w-full rounded border border-line bg-bg px-3 py-2 text-right font-mono text-sm focus:border-saffron focus:outline-none"
            />
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
                {{ tx.category_name ?? categoriesById[tx.category_id]?.name ?? '—' }}
              </p>
              <p v-if="tx.note" class="truncate text-xs text-muted">{{ tx.note }}</p>
              <p class="text-[10px] uppercase tracking-wider text-muted">{{ formatDate(tx.transacted_at) }}</p>
            </div>
            <div class="text-right">
              <p class="font-mono text-sm" :data-testid="`tx-row-${tx.id}-amount`">
                {{ formatRp(tx.amount) }}
              </p>
              <div class="mt-1 flex justify-end gap-2 text-xs">
                <button
                  type="button"
                  class="text-muted hover:text-saffron"
                  :data-testid="`tx-row-${tx.id}-edit`"
                  @click="startEdit(tx)"
                >
                  edit
                </button>
                <button
                  type="button"
                  class="text-muted hover:text-fatigued"
                  :data-testid="`tx-row-${tx.id}-delete`"
                  @click="onDelete(tx)"
                >
                  hapus
                </button>
              </div>
            </div>
          </div>
        </template>
      </li>
    </ul>

    <button
      type="button"
      class="text-xs text-muted hover:text-saffron"
      @click="router.push({ name: 'home' })"
    >
      ← kembali ke beranda
    </button>
  </section>
</template>
