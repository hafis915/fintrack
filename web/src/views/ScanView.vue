<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import imageCompression from 'browser-image-compression'
import { listCategories, type Category } from '@/api/categories'
import { analyzeReceipt, confirmReceipt, type ReceiptDraft } from '@/api/receipts'

const router = useRouter()

const categories = ref<Category[]>([])
const errorMsg = ref<string | null>(null)

// Flow state: idle -> analyzing -> draft.
const analyzing = ref(false)
const draft = ref<ReceiptDraft | null>(null)
const submitting = ref(false)

// The compressed file we both analyze and re-send on confirm.
const compressedFile = ref<File | null>(null)

// Editable draft fields.
const editAmount = ref<number>(0)
const editMerchant = ref<string>('')
const editCategoryId = ref<string>('')
const editNote = ref<string>('')
const confidence = ref<number>(0)

onMounted(async () => {
  try {
    categories.value = await listCategories()
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  }
})

async function onPick(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return

  errorMsg.value = null
  draft.value = null
  analyzing.value = true

  try {
    // Client-side compression to ~1.5MB before upload (server caps at 2MB).
    const compressed = await imageCompression(file, {
      maxSizeMB: 1.5,
      maxWidthOrHeight: 2000,
      useWebWorker: true,
    })
    // imageCompression may return a Blob; normalize to a named File.
    compressedFile.value =
      compressed instanceof File
        ? compressed
        : new File([compressed], file.name, { type: compressed.type || file.type })

    const d = await analyzeReceipt(compressedFile.value)
    draft.value = d
    editAmount.value = d.amount
    editMerchant.value = d.merchant
    editCategoryId.value = d.category_id ?? (categories.value[0]?.id ?? '')
    editNote.value = ''
    confidence.value = d.confidence
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  } finally {
    analyzing.value = false
    // Allow re-picking the same file.
    input.value = ''
  }
}

async function onConfirm() {
  if (!compressedFile.value) {
    errorMsg.value = 'Foto struk dulu ya.'
    return
  }
  if (!editCategoryId.value || editAmount.value <= 0) {
    errorMsg.value = 'Pilih kategori + isi jumlah dulu.'
    return
  }
  errorMsg.value = null
  submitting.value = true
  try {
    await confirmReceipt({
      file: compressedFile.value,
      amount: editAmount.value,
      merchant: editMerchant.value,
      category_id: editCategoryId.value,
      note: editNote.value || undefined,
      transacted_at: new Date().toISOString(),
      ai_confidence: confidence.value,
    })
    router.push({ name: 'transactions' })
  } catch (err) {
    errorMsg.value = err instanceof Error ? err.message : String(err)
  } finally {
    submitting.value = false
  }
}

function reset() {
  draft.value = null
  compressedFile.value = null
  errorMsg.value = null
}

function confidenceLabel(c: number): string {
  if (c >= 0.85) return 'Yakin'
  if (c >= 0.6) return 'Cukup yakin'
  return 'Kurang yakin'
}

function formatPct(c: number): string {
  return Math.round(c * 100) + '%'
}
</script>

<template>
  <section
    class="mx-auto flex max-w-mobile flex-col gap-6 px-6 py-10"
    data-testid="scan-view"
  >
    <header class="space-y-1">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Scan</p>
      <h1 class="font-display text-3xl font-semibold">Foto struk</h1>
      <p class="text-sm text-muted">Jepret struknya, biar dicatat otomatis.</p>
    </header>

    <p v-if="errorMsg" class="text-sm text-fatigued" data-testid="scan-error">{{ errorMsg }}</p>

    <!-- Analyzing: scan-flow animation -->
    <div
      v-if="analyzing"
      class="flex flex-col items-center gap-4 rounded-card border border-line bg-surface p-8"
      data-testid="scan-analyzing"
    >
      <div class="scan-pulse h-16 w-16 rounded-full border-2 border-saffron motion-reduce:animate-none" />
      <p class="font-mono text-sm text-muted">membaca struk…</p>
    </div>

    <!-- Picker (hidden while a draft is being reviewed) -->
    <label
      v-else-if="!draft"
      class="flex cursor-pointer flex-col items-center gap-3 rounded-card border border-dashed border-line bg-surface p-10 text-center"
    >
      <span class="font-mono text-4xl">📷</span>
      <span class="text-sm font-semibold">Ambil / pilih foto struk</span>
      <span class="text-xs text-muted">Maks 2MB · otomatis dikompres</span>
      <input
        type="file"
        accept="image/*"
        capture="environment"
        data-testid="scan-file-input"
        class="hidden"
        @change="onPick"
      />
    </label>

    <!-- Draft review card -->
    <div
      v-else
      class="space-y-5 rounded-card border border-line bg-surface p-5"
      data-testid="scan-draft"
    >
      <div class="flex items-center justify-between">
        <p class="text-xs uppercase tracking-wider text-muted">Cek dulu</p>
        <span
          class="font-mono text-[10px] uppercase tracking-wider text-muted"
          data-testid="scan-confidence"
        >
          {{ confidenceLabel(confidence) }} · {{ formatPct(confidence) }}
        </span>
      </div>

      <!-- Hero amount: saffron Rp prefix + mono digits -->
      <div class="space-y-1">
        <label class="text-xs uppercase tracking-wider text-muted">Jumlah</label>
        <div class="flex items-baseline gap-2">
          <span class="font-mono text-2xl font-semibold text-saffron">Rp</span>
          <input
            v-model.number="editAmount"
            type="number"
            min="0"
            step="1000"
            data-testid="scan-amount"
            class="w-full bg-transparent font-mono text-4xl font-semibold tabular-nums focus:outline-none"
          />
        </div>
      </div>

      <div class="space-y-1">
        <label class="text-xs uppercase tracking-wider text-muted">Merchant</label>
        <input
          v-model="editMerchant"
          type="text"
          placeholder="Nama toko"
          data-testid="scan-merchant"
          class="w-full rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
        />
      </div>

      <div class="space-y-1">
        <label class="text-xs uppercase tracking-wider text-muted">Kategori</label>
        <select
          v-model="editCategoryId"
          data-testid="scan-category"
          class="w-full rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
        >
          <option value="" disabled>Pilih kategori…</option>
          <option v-for="c in categories" :key="c.id" :value="c.id">
            {{ c.icon }} {{ c.name }}
          </option>
        </select>
        <p v-if="draft?.category_hint" class="text-[10px] text-muted">
          AI menebak: {{ draft.category_hint }}
        </p>
      </div>

      <div class="space-y-1">
        <label class="text-xs uppercase tracking-wider text-muted">Catatan (opsional)</label>
        <input
          v-model="editNote"
          type="text"
          placeholder="Catatan"
          data-testid="scan-note"
          class="w-full rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
        />
      </div>

      <div class="flex gap-2">
        <button
          type="button"
          data-testid="scan-confirm"
          :disabled="submitting"
          class="flex-1 rounded-card bg-saffron py-2 text-sm font-semibold text-bg disabled:opacity-60"
          @click="onConfirm"
        >
          {{ submitting ? 'menyimpan…' : 'Simpan transaksi' }}
        </button>
        <button
          type="button"
          class="rounded-card border border-line px-4 py-2 text-sm text-muted"
          @click="reset"
        >
          Ulang
        </button>
      </div>
    </div>

    <button
      type="button"
      class="text-xs text-muted hover:text-saffron"
      @click="router.push({ name: 'transactions' })"
    >
      ← lihat transaksi
    </button>
  </section>
</template>

<style scoped>
.scan-pulse {
  animation: scan-pulse 1.2s ease-in-out infinite;
}

@keyframes scan-pulse {
  0%,
  100% {
    transform: scale(0.85);
    opacity: 0.6;
  }
  50% {
    transform: scale(1.1);
    opacity: 1;
  }
}

@media (prefers-reduced-motion: reduce) {
  .scan-pulse {
    animation: none;
  }
}
</style>
