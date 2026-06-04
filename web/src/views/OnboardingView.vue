<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { listCategories, type Category } from '@/api/categories'
import {
  submitOnboarding,
  type DebtType,
  type Goal,
  type HousingType,
  type LifestyleStyle,
  type OnboardingItem,
} from '@/api/onboarding'
import {
  plannerChat,
  suggestPlan,
  type PlannerMessage,
  type SuggestResponse,
} from '@/api/planner'
import { useOnboardingStore } from '@/stores/onboarding'
import PlannerChat from '@/components/PlannerChat.vue'
import AddCategoryInline from '@/components/AddCategoryInline.vue'

const router = useRouter()
const store = useOnboardingStore()

// --- Step machine ---
type Step = 1 | 2 | 3
const step = ref<Step>(1)

// --- Step 1: questions ---
const income = ref<number>(8_000_000)
const housingType = ref<HousingType>('kpr')
const goal = ref<Goal>('debt')
const debtTypes = ref<DebtType[]>(['cc'])
const emergencyMonths = ref<0 | 1 | 3 | 6>(1)
const lifestyleStyle = ref<LifestyleStyle>('balanced')

// --- Categories split by negotiability ---
const categories = ref<Category[]>([])
const loadingCats = ref(true)
const loadError = ref<string | null>(null)

// Fixed/debt = non-negotiable (the user types these in step 2).
const fixedCategories = computed(() =>
  categories.value.filter((c) => c.type === 'fixed' || c.type === 'debt'),
)

// --- Step 2: fixed amounts keyed by category id ---
const fixedAmounts = ref<Record<string, number>>({})

const fixedTotal = computed(() =>
  fixedCategories.value.reduce((sum, c) => sum + (Number(fixedAmounts.value[c.id]) || 0), 0),
)

// --- Step 3: plan (suggested flexible amounts + summary) ---
const plan = ref<SuggestResponse | null>(null)
const suggesting = ref(false)
const suggestError = ref<string | null>(null)
// Editable flexible amounts keyed by category id (seeded from the suggestion,
// then mutated by inline edits + the planner chat).
const flexibleAmounts = ref<Record<string, number>>({})
const savingsTarget = ref<number>(0)

// Live discretionary total the user has allocated across flexible cats.
const flexibleTotal = computed(() =>
  (plan.value?.flexible ?? []).reduce(
    (sum, f) => sum + (Number(flexibleAmounts.value[f.category_id]) || 0),
    0,
  ),
)

// Residual = what actually persists as savings (income - fixed - flexible).
const residualSavings = computed(() => income.value - fixedTotal.value - flexibleTotal.value)

// --- Planner chat thread ---
const messages = ref<PlannerMessage[]>([])
const chatPending = ref(false)

function formatRp(n: number): string {
  return 'Rp ' + Math.round(n).toLocaleString('id-ID')
}

// Parse a "Rp 1.500.000" / "1500000" string back to an integer.
function parseAmount(raw: string): number {
  const digits = raw.replace(/[^\d]/g, '')
  return digits ? Number(digits) : 0
}

// Bind an amount input to a record entry, formatting with thousand separators.
function fixedAmountDisplay(id: string): string {
  const v = fixedAmounts.value[id]
  return v ? v.toLocaleString('id-ID') : ''
}
function onFixedAmountInput(id: string, e: Event) {
  fixedAmounts.value[id] = parseAmount((e.target as HTMLInputElement).value)
}

function flexibleAmountDisplay(id: string): string {
  const v = flexibleAmounts.value[id]
  return v ? v.toLocaleString('id-ID') : ''
}
function onFlexibleAmountInput(id: string, e: Event) {
  flexibleAmounts.value[id] = parseAmount((e.target as HTMLInputElement).value)
}

function incomeDisplay(): string {
  return income.value ? income.value.toLocaleString('id-ID') : ''
}
function onIncomeInput(e: Event) {
  income.value = parseAmount((e.target as HTMLInputElement).value)
}

function toggleDebt(t: DebtType) {
  const i = debtTypes.value.indexOf(t)
  if (i === -1) debtTypes.value.push(t)
  else debtTypes.value.splice(i, 1)
}

onMounted(async () => {
  try {
    categories.value = await listCategories()
    // Seed fixed amounts from any previously-saved answers.
    const saved = store.answers
    if (saved) {
      income.value = saved.income
      housingType.value = saved.housingType
      goal.value = saved.goal
      debtTypes.value = [...saved.debtTypes]
      emergencyMonths.value = saved.emergencyMonths
      lifestyleStyle.value = saved.lifestyleStyle
      for (const c of fixedCategories.value) {
        const item = saved.items[c.id]
        if (item) fixedAmounts.value[c.id] = item.amount
      }
    }
    // Pre-fill plausible defaults for the common non-negotiables.
    for (const c of fixedCategories.value) {
      if (fixedAmounts.value[c.id] === undefined) {
        fixedAmounts.value[c.id] = defaultFixedFor(c)
      }
    }
  } catch (err) {
    loadError.value = err instanceof Error ? err.message : String(err)
  } finally {
    loadingCats.value = false
  }
})

// A freshly-created custom category joins the in-memory list; the fixedCategories
// computed picks it up automatically (it filters on type). Seed a zero amount so
// the new row renders an empty, editable input.
function onFixedCategoryCreated(c: Category) {
  categories.value = [...categories.value, c]
  if (fixedAmounts.value[c.id] === undefined) fixedAmounts.value[c.id] = 0
}

function defaultFixedFor(c: Category): number {
  if (c.name === 'Cicilan KPR') return 1_500_000
  if (c.name === 'Sewa kosan') return 1_200_000
  if (c.name === 'Kartu kredit') return 400_000
  return 0
}

// Persist the wizard answers (incl. fixed amounts) so a remount restores them.
function persistAnswers() {
  const items: Record<string, { amount: number; enabled: boolean }> = {}
  for (const c of fixedCategories.value) {
    items[c.id] = { amount: Number(fixedAmounts.value[c.id]) || 0, enabled: true }
  }
  store.setAnswers({
    income: income.value,
    housingType: housingType.value,
    goal: goal.value,
    debtTypes: [...debtTypes.value],
    emergencyMonths: emergencyMonths.value,
    lifestyleStyle: lifestyleStyle.value,
    items,
  })
}

// --- Navigation ---
function goNext() {
  loadError.value = null
  if (step.value === 1) {
    if (!Number.isFinite(income.value) || income.value < 100_000) {
      loadError.value = 'Isi pemasukan bulanan minimal Rp 100.000.'
      return
    }
    persistAnswers()
    step.value = 2
    return
  }
  if (step.value === 2) {
    persistAnswers()
    step.value = 3
    void loadSuggestion()
  }
}

function goBack() {
  if (step.value === 2) step.value = 1
  else if (step.value === 3) step.value = 2
}

// Call the deterministic suggest endpoint on entering step 3.
async function loadSuggestion() {
  suggesting.value = true
  suggestError.value = null
  try {
    const fixed_items = fixedCategories.value.map((c) => ({
      category_id: c.id,
      name: c.name,
      icon: c.icon,
      type: c.type,
      amount: Number(fixedAmounts.value[c.id]) || 0,
    }))
    const res = await suggestPlan({
      income: income.value,
      housing_type: housingType.value,
      goal: goal.value,
      debt_types: debtTypes.value,
      emergency_months: emergencyMonths.value,
      lifestyle_style: lifestyleStyle.value,
      fixed_items,
    })
    plan.value = res
    savingsTarget.value = res.savings_target
    // Seed the editable flexible amounts from the suggestion.
    const seeded: Record<string, number> = {}
    for (const f of res.flexible) seeded[f.category_id] = f.suggested_amount
    flexibleAmounts.value = seeded
  } catch (err) {
    suggestError.value = err instanceof Error ? err.message : String(err)
  } finally {
    suggesting.value = false
  }
}

// --- Planner chat send: deterministic re-balance happens server-side ---
async function onChatSend(text: string) {
  if (!plan.value) return
  messages.value.push({ role: 'user', content: text })
  chatPending.value = true
  try {
    const res = await plannerChat({
      income: income.value,
      goal: goal.value,
      lifestyle_style: lifestyleStyle.value,
      savings_target: savingsTarget.value,
      fixed_items: fixedCategories.value.map((c) => ({
        category_id: c.id,
        name: c.name,
        amount: Number(fixedAmounts.value[c.id]) || 0,
      })),
      flexible: plan.value.flexible.map((f) => ({
        category_id: f.category_id,
        name: f.name,
        amount: Number(flexibleAmounts.value[f.category_id]) || 0,
      })),
      messages: messages.value,
      user_message: text,
    })
    messages.value.push({ role: 'assistant', content: res.reply })
    // Apply the server's re-balanced numbers (the AI never invents these).
    if (res.changed) {
      for (const f of res.flexible) flexibleAmounts.value[f.category_id] = f.amount
      savingsTarget.value = res.savings_target
    }
  } catch (err) {
    messages.value.push({
      role: 'assistant',
      content: err instanceof Error ? err.message : 'Planner lagi error, coba lagi ya.',
    })
  } finally {
    chatPending.value = false
  }
}

// --- Confirm: persist via the EXISTING finalize endpoint ---
const submitting = ref(false)
const submitError = ref<string | null>(null)

async function onSubmit() {
  submitError.value = null
  if (!plan.value) return
  persistAnswers()
  submitting.value = true
  try {
    const expense_items: OnboardingItem[] = []
    // Fixed/debt items (non-zero only).
    for (const c of fixedCategories.value) {
      const amt = Number(fixedAmounts.value[c.id]) || 0
      if (amt > 0) {
        expense_items.push({
          category_id: c.id,
          name: c.name,
          icon: c.icon,
          type: c.type,
          amount: amt,
        })
      }
    }
    // Final flexible items (non-zero only).
    for (const f of plan.value.flexible) {
      const amt = Number(flexibleAmounts.value[f.category_id]) || 0
      if (amt > 0) {
        expense_items.push({
          category_id: f.category_id,
          name: f.name,
          icon: f.icon,
          type: f.type,
          amount: amt,
        })
      }
    }
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
    <header class="space-y-2">
      <span class="inline-block border-2 border-line bg-fg px-2 py-1 text-[10px] font-bold uppercase tracking-[0.18em] text-bg">
        Langkah {{ step }} / 3
      </span>
      <h1 class="font-display text-4xl font-extrabold uppercase leading-none tracking-tight">
        Susun<br />programmu
      </h1>
      <p class="border-l-4 border-saffron pl-3 text-sm font-medium">
        Bukan form — ini financial planner. Kamu yang pegang kemudi.
      </p>
    </header>

    <p v-if="loadError" class="text-sm text-fatigued" data-testid="onb-error">{{ loadError }}</p>

    <!-- ============ STEP 1 — QUESTIONS ============ -->
    <div v-if="step === 1" class="space-y-5" data-testid="onb-step-1">
      <!-- Q1: income -->
      <fieldset class="space-y-2">
        <label class="text-xs font-bold uppercase tracking-wider text-muted" for="income">
          1. Pemasukan bulanan
        </label>
        <div class="flex items-center rounded-card border-2 border-line bg-surface px-4 py-3 shadow-brutal focus-within:border-saffron">
          <span class="mr-2 font-mono text-lg font-bold text-saffron">Rp</span>
          <input
            id="income"
            :value="incomeDisplay()"
            inputmode="numeric"
            data-testid="onb-income"
            class="w-full bg-transparent font-mono text-lg focus:outline-none"
            @input="onIncomeInput"
          />
        </div>
      </fieldset>

      <!-- Q2: housing -->
      <fieldset class="space-y-2">
        <legend class="text-xs font-bold uppercase tracking-wider text-muted">2. Tempat tinggal</legend>
        <div class="grid grid-cols-3 gap-2">
          <button
            v-for="opt in (['kosan', 'kpr', 'keluarga'] as HousingType[])"
            :key="opt"
            type="button"
            class="rounded-card border-2 border-line px-3 py-2 text-center text-sm font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
            :class="housingType === opt ? 'bg-saffron text-fg' : 'bg-surface'"
            :data-testid="`onb-housing-${opt}`"
            @click="housingType = opt"
          >
            {{ opt }}
          </button>
        </div>
      </fieldset>

      <!-- Q3: goal -->
      <fieldset class="space-y-2">
        <legend class="text-xs font-bold uppercase tracking-wider text-muted">3. Goal utama</legend>
        <select
          v-model="goal"
          data-testid="onb-goal"
          class="w-full rounded-card border-2 border-line bg-surface px-3 py-3 text-sm font-semibold shadow-brutal-sm focus:border-saffron focus:outline-none"
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
        <legend class="text-xs font-bold uppercase tracking-wider text-muted">4. Jenis utang aktif</legend>
        <div class="flex flex-wrap gap-2">
          <button
            v-for="opt in (['none', 'cc', 'paylater', 'multi'] as DebtType[])"
            :key="opt"
            type="button"
            class="rounded-card border-2 border-line px-3 py-1.5 text-sm font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
            :class="debtTypes.includes(opt) ? 'bg-saffron text-fg' : 'bg-surface'"
            :data-testid="`onb-debt-${opt}`"
            @click="toggleDebt(opt)"
          >
            {{ opt === 'none' ? 'tidak ada' : opt }}
          </button>
        </div>
      </fieldset>

      <!-- Q5: emergency months -->
      <fieldset class="space-y-2">
        <legend class="text-xs font-bold uppercase tracking-wider text-muted">5. Dana darurat sekarang</legend>
        <div class="grid grid-cols-4 gap-2">
          <button
            v-for="opt in ([0, 1, 3, 6] as (0 | 1 | 3 | 6)[])"
            :key="opt"
            type="button"
            class="rounded-card border-2 border-line py-2 text-center text-sm font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
            :class="emergencyMonths === opt ? 'bg-saffron text-fg' : 'bg-surface'"
            :data-testid="`onb-emergency-${opt}`"
            @click="emergencyMonths = opt"
          >
            {{ opt }} bln
          </button>
        </div>
      </fieldset>

      <!-- Q6: lifestyle -->
      <fieldset class="space-y-2">
        <legend class="text-xs font-bold uppercase tracking-wider text-muted">6. Gaya hidup</legend>
        <div class="grid grid-cols-3 gap-2">
          <button
            v-for="opt in (['easy', 'balanced', 'strict'] as LifestyleStyle[])"
            :key="opt"
            type="button"
            class="rounded-card border-2 border-line px-3 py-2 text-center text-sm font-bold uppercase shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
            :class="lifestyleStyle === opt ? 'bg-saffron text-fg' : 'bg-surface'"
            :data-testid="`onb-lifestyle-${opt}`"
            @click="lifestyleStyle = opt"
          >
            {{ opt }}
          </button>
        </div>
      </fieldset>
    </div>

    <!-- ============ STEP 2 — FIXED EXPENSES ============ -->
    <div v-else-if="step === 2" class="space-y-4" data-testid="onb-step-2">
      <div class="space-y-2">
        <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Pengeluaran tetap</span>
        <p class="text-sm text-muted">Yang nggak bisa diubah — sewa, listrik, internet, cicilan.</p>
      </div>

      <p v-if="loadingCats" class="font-mono text-sm text-muted">memuat kategori…</p>

      <div v-else class="space-y-2">
        <div
          v-for="c in fixedCategories"
          :key="c.id"
          class="flex items-center gap-3 rounded-card border-2 border-line bg-surface px-3 py-3 shadow-brutal-sm"
        >
          <span class="flex-1 text-sm font-semibold">
            <span class="mr-2">{{ c.icon }}</span>{{ c.name }}
          </span>
          <div class="flex w-40 items-center rounded-card border-2 border-line bg-bg px-2 py-1.5 focus-within:border-saffron">
            <span class="mr-1 font-mono text-xs font-bold text-saffron">Rp</span>
            <input
              :value="fixedAmountDisplay(c.id)"
              inputmode="numeric"
              :data-testid="`onb-fixed-amount-${c.name}`"
              class="w-full bg-transparent text-right font-mono text-sm focus:outline-none"
              @input="(e) => onFixedAmountInput(c.id, e)"
            />
          </div>
        </div>
      </div>

      <!-- Add a fixed expense the catalog doesn't cover (e.g. cicilan motor). -->
      <AddCategoryInline
        v-if="!loadingCats"
        default-type="fixed"
        @created="onFixedCategoryCreated"
      />

      <div
        class="flex items-center justify-between rounded-card border-2 border-line bg-fg px-4 py-3 text-sm text-bg"
        data-testid="onb-fixed-total"
      >
        <span class="font-bold uppercase tracking-wider">Total tetap</span>
        <span class="font-mono">{{ formatRp(fixedTotal) }}</span>
      </div>
    </div>

    <!-- ============ STEP 3 — PLAN + PLANNER CHAT ============ -->
    <div v-else class="space-y-5" data-testid="onb-step-3">
      <p v-if="suggesting" class="font-mono text-sm text-muted">menyusun rencana…</p>
      <p v-else-if="suggestError" class="text-sm text-fatigued" data-testid="onb-suggest-error">
        {{ suggestError }}
      </p>

      <template v-else-if="plan">
        <!-- Top-line numbers -->
        <div class="space-y-2 rounded-card border-2 border-line bg-surface p-4 shadow-brutal" data-testid="onb-plan-summary">
          <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Rencana</span>
          <dl class="space-y-1 font-mono text-sm">
            <div class="flex justify-between">
              <dt class="text-muted">Pemasukan</dt>
              <dd>{{ formatRp(income) }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-muted">Total tetap</dt>
              <dd>{{ formatRp(fixedTotal) }}</dd>
            </div>
            <div class="flex justify-between">
              <dt class="text-muted">Dipakai keinginan</dt>
              <dd>{{ formatRp(flexibleTotal) }}</dd>
            </div>
            <div class="flex justify-between border-t-2 border-line pt-1">
              <dt class="font-bold text-saffron">Target tabungan</dt>
              <dd
                :class="residualSavings < 0 ? 'font-bold text-fatigued' : 'font-bold text-saffron'"
                data-testid="onb-savings-target"
              >
                {{ formatRp(residualSavings) }}
              </dd>
            </div>
          </dl>
        </div>

        <!-- Over-budget coaching -->
        <p
          v-if="plan.warning"
          data-testid="onb-warning"
          class="flex items-start gap-2 rounded-card border-2 border-line bg-warning px-4 py-2 text-sm font-medium text-fg shadow-brutal-sm"
        >
          <span aria-hidden="true">⚠</span><span>{{ plan.warning }}</span>
        </p>
        <p
          v-else-if="residualSavings < 0"
          data-testid="onb-warning"
          class="flex items-start gap-2 rounded-card border-2 border-line bg-fatigued px-4 py-2 text-sm font-medium text-fg shadow-brutal-sm"
        >
          <span aria-hidden="true">⚠</span>
          <span>Pengeluaran lebih dari pemasukan. Kecilin keinginan atau minta planner bantu rapikan.</span>
        </p>

        <!-- Suggested flexible categories — inline editable -->
        <div class="space-y-2" data-testid="onb-flexible">
          <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">Keinginan (saran)</span>
          <p class="text-xs text-muted">Angka saran dari app. Ubah langsung atau ngobrol sama planner di bawah.</p>
          <div
            v-for="f in plan.flexible"
            :key="f.category_id"
            class="flex items-center gap-3 rounded-card border-2 border-line bg-surface px-3 py-3 shadow-brutal-sm"
          >
            <span class="flex-1 text-sm font-semibold">
              <span class="mr-2">{{ f.icon }}</span>{{ f.name }}
            </span>
            <div class="flex w-40 items-center rounded-card border-2 border-line bg-bg px-2 py-1.5 focus-within:border-saffron">
              <span class="mr-1 font-mono text-xs font-bold text-saffron">Rp</span>
              <input
                :value="flexibleAmountDisplay(f.category_id)"
                inputmode="numeric"
                :data-testid="`onb-flex-amount-${f.name}`"
                class="w-full bg-transparent text-right font-mono text-sm focus:outline-none"
                @input="(e) => onFlexibleAmountInput(f.category_id, e)"
              />
            </div>
          </div>
        </div>

        <!-- Planner chat (multi-turn) -->
        <PlannerChat :messages="messages" :pending="chatPending" @send="onChatSend" />
      </template>
    </div>

    <!-- ============ STEP NAV ============ -->
    <div class="flex items-center gap-3">
      <button
        v-if="step > 1"
        type="button"
        data-testid="onb-back"
        class="border-2 border-line bg-surface px-4 py-3 text-sm font-bold uppercase shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none motion-reduce:transform-none"
        @click="goBack"
      >
        ← Kembali
      </button>

      <button
        v-if="step < 3"
        type="button"
        data-testid="onb-next"
        :disabled="loadingCats"
        class="flex-1 rounded-card border-2 border-line bg-saffron py-3 text-sm font-bold uppercase text-fg shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:opacity-50 motion-reduce:transform-none"
        @click="goNext"
      >
        Lanjut →
      </button>

      <button
        v-else
        type="button"
        data-testid="onb-submit"
        :disabled="submitting || suggesting || !plan"
        class="flex-1 rounded-card border-2 border-line bg-saffron py-3 text-sm font-bold uppercase text-fg shadow-brutal active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:opacity-50 motion-reduce:transform-none"
        @click="onSubmit"
      >
        {{ submitting ? 'Menyusun…' : 'Mulai program' }}
      </button>
    </div>

    <p v-if="submitError" data-testid="onb-submit-error" class="text-sm text-fatigued">
      {{ submitError }}
    </p>
  </section>
</template>
