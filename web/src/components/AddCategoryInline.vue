<script setup lang="ts">
import { ref } from 'vue'
import { createCategory, type Category, type ExpenseType } from '@/api/categories'

// Inline "create a custom expense category" affordance. Used in two places:
//  - onboarding step 2 (fixed expenses) → defaultType="fixed", no type picker
//  - the add-transaction form           → showTypePicker, defaultType="variable"
// It owns its own open/closed + form state; the parent just listens for @created
// and folds the returned Category into its own list.
const props = withDefaults(
  defineProps<{
    defaultType?: ExpenseType
    showTypePicker?: boolean
  }>(),
  { defaultType: 'variable', showTypePicker: false },
)

const emit = defineEmits<{ created: [category: Category] }>()

const open = ref(false)
const name = ref('')
const icon = ref('')
const type = ref<ExpenseType>(props.defaultType)
const saving = ref(false)
const error = ref<string | null>(null)

// Indonesian labels for the four enum values (matches the backend).
const typeOptions: { value: ExpenseType; label: string }[] = [
  { value: 'fixed', label: 'Tetap' },
  { value: 'variable', label: 'Variabel' },
  { value: 'want', label: 'Keinginan' },
  { value: 'debt', label: 'Utang' },
]

function reset() {
  name.value = ''
  icon.value = ''
  type.value = props.defaultType
  error.value = null
}

function openForm() {
  reset()
  open.value = true
}

function cancel() {
  open.value = false
  reset()
}

async function submit() {
  const trimmed = name.value.trim()
  if (!trimmed) {
    error.value = 'Isi nama kategori dulu.'
    return
  }
  saving.value = true
  error.value = null
  try {
    const created = await createCategory({
      name: trimmed,
      type: type.value,
      icon: icon.value.trim() || undefined,
    })
    emit('created', created)
    open.value = false
    reset()
  } catch (err) {
    error.value = err instanceof Error ? err.message : String(err)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div data-testid="add-category">
    <!-- Collapsed: a dashed "add" affordance that opens the inline form. -->
    <button
      v-if="!open"
      type="button"
      data-testid="add-category-toggle"
      class="w-full rounded-card border-2 border-dashed border-line bg-bg px-3 py-2.5 text-sm font-bold uppercase tracking-wider text-muted active:translate-x-[2px] active:translate-y-[2px] motion-reduce:transform-none"
      @click="openForm"
    >
      + Kategori baru
    </button>

    <!-- Expanded: name + optional icon (+ optional type picker), then save/cancel. -->
    <div
      v-else
      class="space-y-2 rounded-card border-2 border-line bg-surface p-3 shadow-brutal-sm"
    >
      <div class="flex items-center gap-2">
        <input
          v-model="icon"
          type="text"
          maxlength="2"
          placeholder="🏷️"
          aria-label="Ikon kategori"
          data-testid="add-category-icon"
          class="w-12 shrink-0 rounded-card border-2 border-line bg-bg px-2 py-2 text-center text-base focus:border-saffron focus:outline-none"
        />
        <input
          v-model="name"
          type="text"
          placeholder="Nama kategori"
          data-testid="add-category-name"
          class="w-full rounded-card border-2 border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none"
          @keydown.enter.prevent="submit"
        />
      </div>

      <select
        v-if="showTypePicker"
        v-model="type"
        aria-label="Jenis kategori"
        data-testid="add-category-type"
        class="w-full rounded-card border-2 border-line bg-bg px-3 py-2 text-sm font-semibold focus:border-saffron focus:outline-none"
      >
        <option v-for="o in typeOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
      </select>

      <p v-if="error" class="text-xs text-fatigued" data-testid="add-category-error">{{ error }}</p>

      <div class="flex gap-2">
        <button
          type="button"
          data-testid="add-category-submit"
          :disabled="saving"
          class="flex-1 rounded-card border-2 border-line bg-saffron py-2 text-sm font-bold uppercase text-fg shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:opacity-50 motion-reduce:transform-none"
          @click="submit"
        >
          {{ saving ? 'Menyimpan…' : 'Simpan' }}
        </button>
        <button
          type="button"
          data-testid="add-category-cancel"
          class="rounded-card border-2 border-line bg-surface px-4 py-2 text-sm font-bold uppercase text-muted active:translate-x-[2px] active:translate-y-[2px] motion-reduce:transform-none"
          @click="cancel"
        >
          Batal
        </button>
      </div>
    </div>
  </div>
</template>
