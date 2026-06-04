<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { createCategory, listCategories, type Category, type ExpenseType } from '@/api/categories'
import {
  submitOnboarding,
  type DebtType,
  type Goal,
  type HousingType,
  type LifestyleStyle,
  type OnboardingItem,
} from '@/api/onboarding'
import { useOnboardingStore, type OnboardingAnswers } from '@/stores/onboarding'

interface DraftItem {
  category: Category
  amount: number
  enabled: boolean
}

const router = useRouter()
const store = useOnboardingStore()

const categories = ref<Category[]>([])
const loadingCats = ref(true)
const submitting = ref(false)
const submitError = ref<string | null>(null)

// Form state — defaults chosen to be safe to submit immediately in tests.
const income = ref<number>(8_000_000)
const housingType = ref<HousingType>('kpr')
const goal = ref<Goal>('debt')
const debtTypes = ref<DebtType[]>(['cc'])
const emergencyMonths = ref<0 | 1 | 3 | 6>(1)
const lifestyleStyle = ref<LifestyleStyle>('balanced')

const drafts = ref<DraftItem[]>([])

const groupedDrafts = computed(() => {
  const groups: Record<ExpenseType, DraftItem[]> = { fixed: [], variable: [], debt: [], want: [] }
  for (const d of drafts.value) groups[d.category.type].push(d)
  return groups
})

// Live running total of declared expenses, and how far it exceeds income.
// Surfaced as coaching feedback — overspending is allowed (the result page
// coaches it), so this never blocks submission.
const totalExpenses = computed(() =>
  drafts.value
    .filter((d) => d.enabled && d.amount > 0)
    .reduce((sum, d) => sum + (Number(d.amount) || 0), 0),
)
const overBy = computed(() => totalExpenses.value - (Number(income.value) || 0))

function formatRp(n: number): string {
  return 'Rp ' + Math.round(n).toLocaleString('id-ID')
}

onMounted(async () => {
  try {
    categories.value = await listCategories()
    // Rehydrate the user's previous answers (e.g. after "Ubah jawaban")
    // instead of resetting to defaults. Scalars first, then per-item state
    // merged onto the freshly loaded categories by category_id.
    const saved = store.answers
    if (saved) {
      income.value = saved.income
      housingType.value = saved.housingType
      goal.value = saved.goal
      debtTypes.value = [...saved.debtTypes]
      emergencyMonths.value = saved.emergencyMonths
      lifestyleStyle.value = saved.lifestyleStyle
    }
    drafts.value = categories.value.map((c) => {
      const savedItem = saved?.items[c.id]
      return {
        category: c,
        // Restore the user's amount/enabled if we have it; otherwise pre-fill
        // plausible type-specific defaults for a quick first-time path.
        amount: savedItem ? savedItem.amount : defaultAmountFor(c),
        enabled: savedItem ? savedItem.enabled : defaultEnabledFor(c),
      }
    })
  } catch (err) {
    submitError.value = err instanceof Error ? err.message : String(err)
  } finally {
    loadingCats.value = false
  }
})

// Snapshot the current form so the result page's "Ubah jawaban" can restore it.
function snapshotAnswers(): OnboardingAnswers {
  const items: Record<string, { amount: number; enabled: boolean }> = {}
  for (const d of drafts.value) {
    items[d.category.id] = { amount: Number(d.amount) || 0, enabled: d.enabled }
  }
  return {
    income: income.value,
    housingType: housingType.value,
    goal: goal.value,
    debtTypes: [...debtTypes.value],
    emergencyMonths: emergencyMonths.value,
    lifestyleStyle: lifestyleStyle.value,
    items,
  }
}

function defaultAmountFor(c: Category): number {
  if (c.name === 'Cicilan KPR') return 1_500_000
  if (c.name === 'Sewa kosan') return 1_200_000
  if (c.name === 'Makan & minum') return 1_200_000
  if (c.name === 'Kartu kredit') return 400_000
  if (c.name === 'Hiburan') return 500_000
  return 0
}

function defaultEnabledFor(c: Category): boolean {
  return ['Cicilan KPR', 'Makan & minum', 'Kartu kredit', 'Hiburan'].includes(c.name)
}

// --- Add a custom expense the default seed doesn't cover ---
const newExpName = ref('')
const newExpType = ref<ExpenseType>('variable')
const newExpAmount = ref<number>(0)
const addingExp = ref(false)
const addError = ref<string | null>(null)

async function addCustomExpense() {
  addError.value = null
  const name = newExpName.value.trim()
  if (!name) {
    addError.value = 'Isi nama pengeluaran dulu.'
    return
  }
  const amt = Number(newExpAmount.value)
  if (!Number.isFinite(amt) || amt <= 0) {
    addError.value = 'Isi jumlah lebih dari 0.'
    return
  }
  addingExp.value = true
  try {
    const cat = await createCategory({ name, type: newExpType.value })
    categories.value.push(cat)
    drafts.value.push({ category: cat, amount: amt, enabled: true })
    newExpName.value = ''
    newExpAmount.value = 0
    newExpType.value = 'variable'
  } catch (err) {
    addError.value = err instanceof Error ? err.message : String(err)
  } finally {
    addingExp.value = false
  }
}

function toggleDebt(t: DebtType) {
  const i = debtTypes.value.indexOf(t)
  if (i === -1) debtTypes.value.push(t)
  else debtTypes.value.splice(i, 1)
}

async function onSubmit() {
  submitError.value = null

  // Guard income before hitting the API. A cleared number field serializes
  // as "" and the backend rejects it as invalid_json (bug #3) — catch it
  // here with a friendly message instead. Number() normalizes NaN/empty.
  const incomeValue = Number(income.value)
  if (!Number.isFinite(incomeValue) || incomeValue < 100_000) {
    income.value = incomeValue
    submitError.value = 'Isi pemasukan bulanan minimal Rp 100.000.'
    return
  }
  income.value = incomeValue

  // Persist answers up front so "Ubah jawaban" restores them even if the
  // request fails or the user navigates back without resubmitting.
  store.setAnswers(snapshotAnswers())

  submitting.value = true
  try {
    const expense_items: OnboardingItem[] = drafts.value
      .filter((d) => d.enabled && d.amount > 0)
      .map((d) => ({
        category_id: d.category.id,
        name: d.category.name,
        icon: d.category.icon,
        type: d.category.type,
        amount: d.amount,
      }))

    const resp = await submitOnboarding({
      income: income.value,
      housing_type: housingType.value,
      goal: goal.value,
      debt_types: debtTypes.value,
      emergency_months: emergencyMonths.value,
      lifestyle_style: lifestyleStyle.value,
      expense_items,
    })
    store.setPlan(resp)
    router.push({ name: 'onboarding-result' })
  } catch (err) {
    submitError.value = err instanceof Error ? err.message : String(err)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section
    class="mx-auto flex max-w-mobile flex-col gap-6 px-6 py-10"
    data-testid="onboarding-view"
  >
    <header class="space-y-1">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Onboarding</p>
      <h1 class="font-display text-3xl font-semibold">Mulai program-mu</h1>
      <p class="text-sm text-muted">Jawab 6 pertanyaan singkat. Kami susun budget yang pas.</p>
    </header>

    <form novalidate class="space-y-5" @submit.prevent="onSubmit">
      <!-- Q1: income -->
      <fieldset class="space-y-2">
        <label class="text-xs uppercase tracking-wider text-muted" for="income">
          1. Pemasukan bulanan
        </label>
        <input
          id="income"
          v-model.number="income"
          type="number"
          min="100000"
          step="100000"
          data-testid="onb-income"
          class="w-full rounded-card border border-line bg-surface px-4 py-3 font-mono text-lg focus:border-saffron focus:outline-none"
        />
      </fieldset>

      <!-- Q2: housing -->
      <fieldset class="space-y-2">
        <legend class="text-xs uppercase tracking-wider text-muted">2. Tempat tinggal</legend>
        <div class="grid grid-cols-3 gap-2">
          <label
            v-for="opt in (['kosan', 'kpr', 'keluarga'] as HousingType[])"
            :key="opt"
            class="cursor-pointer rounded-card border border-line bg-surface px-3 py-2 text-center text-sm capitalize hover:border-saffron"
            :class="housingType === opt ? 'border-saffron text-saffron' : ''"
          >
            <input
              v-model="housingType"
              type="radio"
              :value="opt"
              class="sr-only"
              :data-testid="`onb-housing-${opt}`"
            />
            {{ opt }}
          </label>
        </div>
      </fieldset>

      <!-- Q3: goal -->
      <fieldset class="space-y-2">
        <legend class="text-xs uppercase tracking-wider text-muted">3. Goal utama</legend>
        <select
          v-model="goal"
          data-testid="onb-goal"
          class="w-full rounded-card border border-line bg-surface px-3 py-3 capitalize focus:border-saffron focus:outline-none"
        >
          <option value="emergency">Bangun dana darurat</option>
          <option value="debt">Bebas dari utang</option>
          <option value="goal">Nabung tujuan spesifik</option>
          <option value="invest">Mulai investasi</option>
          <option value="balance">Kontrol pengeluaran umum</option>
        </select>
      </fieldset>

      <!-- Q4: debt types (multi) -->
      <fieldset class="space-y-2">
        <legend class="text-xs uppercase tracking-wider text-muted">4. Jenis utang aktif</legend>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="opt in (['none', 'cc', 'paylater', 'multi'] as DebtType[])"
            :key="opt"
            type="button"
            class="rounded-full border border-line bg-surface px-3 py-1 text-sm capitalize hover:border-saffron"
            :class="debtTypes.includes(opt) ? 'border-saffron bg-saffron/10 text-saffron' : ''"
            :data-testid="`onb-debt-${opt}`"
            @click="toggleDebt(opt)"
          >
            {{ opt === 'none' ? 'tidak ada' : opt }}
          </button>
        </div>
      </fieldset>

      <!-- Q5: emergency months -->
      <fieldset class="space-y-2">
        <legend class="text-xs uppercase tracking-wider text-muted">
          5. Dana darurat sekarang
        </legend>
        <div class="grid grid-cols-4 gap-2">
          <label
            v-for="opt in [0, 1, 3, 6]"
            :key="opt"
            class="cursor-pointer rounded-card border border-line bg-surface py-2 text-center text-sm hover:border-saffron"
            :class="emergencyMonths === opt ? 'border-saffron text-saffron' : ''"
          >
            <input
              v-model.number="emergencyMonths"
              type="radio"
              :value="opt"
              class="sr-only"
              :data-testid="`onb-emergency-${opt}`"
            />
            {{ opt === 0 ? '0 bln' : `${opt} bln` }}
          </label>
        </div>
      </fieldset>

      <!-- Q6: lifestyle -->
      <fieldset class="space-y-2">
        <legend class="text-xs uppercase tracking-wider text-muted">6. Gaya hidup</legend>
        <div class="grid grid-cols-3 gap-2">
          <label
            v-for="opt in (['easy', 'balanced', 'strict'] as LifestyleStyle[])"
            :key="opt"
            class="cursor-pointer rounded-card border border-line bg-surface px-3 py-2 text-center text-sm capitalize hover:border-saffron"
            :class="lifestyleStyle === opt ? 'border-saffron text-saffron' : ''"
          >
            <input
              v-model="lifestyleStyle"
              type="radio"
              :value="opt"
              class="sr-only"
              :data-testid="`onb-lifestyle-${opt}`"
            />
            {{ opt }}
          </label>
        </div>
      </fieldset>

      <!-- Expense items -->
      <fieldset class="space-y-3" data-testid="onb-items">
        <legend class="text-xs uppercase tracking-wider text-muted">
          7. Pengeluaran bulanan kamu
        </legend>
        <p v-if="loadingCats" class="font-mono text-sm text-muted">memuat kategori…</p>

        <div v-else class="space-y-4">
          <div v-for="(items, type) in groupedDrafts" :key="type" class="space-y-2">
            <p class="text-[10px] uppercase tracking-[0.2em] text-muted">{{ type }}</p>
            <div
              v-for="d in items"
              :key="d.category.id"
              class="flex items-center gap-3 rounded-card border border-line bg-surface px-3 py-2"
            >
              <input
                v-model="d.enabled"
                type="checkbox"
                :data-testid="`onb-item-enable-${d.category.name}`"
                class="h-4 w-4 accent-saffron"
              />
              <span class="flex-1 text-sm">
                <span class="mr-2">{{ d.category.icon }}</span>
                {{ d.category.name }}
              </span>
              <input
                v-model.number="d.amount"
                type="number"
                min="0"
                step="50000"
                :disabled="!d.enabled"
                :data-testid="`onb-item-amount-${d.category.name}`"
                class="w-32 rounded border border-line bg-bg px-2 py-1 text-right font-mono text-sm disabled:opacity-40 focus:border-saffron focus:outline-none"
              />
            </div>
          </div>

          <!-- Add an expense the defaults don't cover -->
          <div
            class="space-y-2 rounded-card border border-dashed border-line p-3"
            data-testid="onb-add-expense"
          >
            <p class="text-[10px] uppercase tracking-[0.2em] text-muted">Tambah pengeluaran lain</p>
            <input
              v-model="newExpName"
              type="text"
              placeholder="Nama (mis. Kursus online)"
              data-testid="onb-add-name"
              class="w-full rounded border border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
            />
            <div class="flex gap-2">
              <select
                v-model="newExpType"
                data-testid="onb-add-type"
                class="rounded border border-line bg-bg px-2 py-2 text-sm capitalize focus:border-saffron focus:outline-none"
              >
                <option value="fixed">fixed</option>
                <option value="variable">variable</option>
                <option value="debt">debt</option>
                <option value="want">want</option>
              </select>
              <input
                v-model.number="newExpAmount"
                type="number"
                min="0"
                step="50000"
                placeholder="Jumlah"
                data-testid="onb-add-amount"
                class="w-32 flex-1 rounded border border-line bg-bg px-2 py-2 text-right font-mono text-sm focus:border-saffron focus:outline-none"
              />
            </div>
            <button
              type="button"
              :disabled="addingExp"
              data-testid="onb-add-submit"
              class="w-full rounded-card border border-saffron py-2 text-sm font-semibold text-saffron disabled:opacity-50"
              @click="addCustomExpense"
            >
              {{ addingExp ? 'Menambah…' : '+ Tambah pengeluaran' }}
            </button>
            <p v-if="addError" data-testid="onb-add-error" class="text-xs text-fatigued">
              {{ addError }}
            </p>
          </div>
        </div>
      </fieldset>

      <!-- Live expenses total vs income — coaching feedback, never blocks submit. -->
      <div
        v-if="!loadingCats"
        class="flex items-center justify-between rounded-card border border-line bg-surface px-4 py-3 text-sm"
        data-testid="onb-expense-total"
      >
        <span class="text-muted">Total pengeluaran</span>
        <span class="font-mono">{{ formatRp(totalExpenses) }}</span>
      </div>
      <p
        v-if="overBy > 0"
        data-testid="onb-overspend"
        class="flex items-start gap-2 rounded-card border border-warning/40 bg-warning/10 px-4 py-2 text-sm text-warning"
      >
        <span aria-hidden="true">⚠</span>
        <span>Pengeluaran melebihi pemasukan — lebih {{ formatRp(overBy) }}. Tetap bisa lanjut, nanti kami bantu rapikan.</span>
      </p>

      <p v-if="submitError" data-testid="onb-error" class="text-sm text-fatigued">
        {{ submitError }}
      </p>

      <button
        type="submit"
        :disabled="submitting || loadingCats"
        data-testid="onb-submit"
        class="w-full rounded-card bg-saffron py-3 font-semibold text-bg disabled:opacity-50"
      >
        {{ submitting ? 'Menyusun program…' : 'Mulai program' }}
      </button>
    </form>
  </section>
</template>
