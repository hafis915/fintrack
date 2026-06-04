<script setup lang="ts">
import { nextTick, ref, watch } from 'vue'
import type { PlannerMessage } from '@/api/planner'

// Multi-turn planner chat. The parent owns the message thread + the actual
// money state; this component is a dumb brutalist transcript + input. On send
// it emits the raw text and lets the parent fire the API call, append the
// assistant reply, and re-balance the deterministic numbers.
const props = defineProps<{
  messages: PlannerMessage[]
  pending: boolean
}>()

const emit = defineEmits<{ (e: 'send', text: string): void }>()

const draft = ref('')
const listEl = ref<HTMLElement | null>(null)

function onSend() {
  const text = draft.value.trim()
  if (!text || props.pending) return
  emit('send', text)
  draft.value = ''
}

// Keep the transcript pinned to the latest message as it grows.
watch(
  () => props.messages.length,
  async () => {
    await nextTick()
    if (listEl.value) listEl.value.scrollTop = listEl.value.scrollHeight
  },
)
</script>

<template>
  <div
    class="space-y-3 rounded-card border-2 border-line bg-surface p-4 shadow-brutal"
    data-testid="planner-chat"
  >
    <span class="inline-block bg-fg px-2 py-[2px] text-[10px] font-bold uppercase tracking-wider text-bg">
      Ngobrol sama planner
    </span>

    <p class="text-xs text-muted">
      Minta apa aja: "naikin makan jadi 1.5jt", "ambil dari tabungan", "kecilin hiburan".
    </p>

    <div
      ref="listEl"
      class="flex max-h-72 flex-col gap-2 overflow-y-auto"
      data-testid="planner-thread"
    >
      <p
        v-if="messages.length === 0"
        class="py-6 text-center font-mono text-xs text-muted"
      >
        Belum ada obrolan. Mulai ngetik di bawah.
      </p>

      <div
        v-for="(m, i) in messages"
        :key="i"
        class="flex"
        :class="m.role === 'user' ? 'justify-end' : 'justify-start'"
      >
        <p
          :data-testid="`planner-message`"
          :data-role="m.role"
          class="max-w-[80%] whitespace-pre-wrap break-words rounded-card border-2 border-line px-3 py-2 text-sm"
          :class="
            m.role === 'user'
              ? 'bg-saffron text-fg shadow-brutal-sm'
              : 'bg-elevated text-fg shadow-brutal-sm'
          "
        >
          {{ m.content }}
        </p>
      </div>

      <div v-if="pending" class="flex justify-start">
        <p class="rounded-card border-2 border-line bg-elevated px-3 py-2 font-mono text-xs text-muted shadow-brutal-sm">
          planner ngetik…
        </p>
      </div>
    </div>

    <div class="flex items-stretch gap-2">
      <input
        v-model="draft"
        type="text"
        placeholder="Tulis permintaanmu…"
        data-testid="planner-input"
        :disabled="pending"
        class="min-w-0 flex-1 rounded-card border-2 border-line bg-bg px-3 py-2 text-sm focus:border-saffron focus:outline-none disabled:opacity-50"
        @keydown.enter.prevent="onSend"
      />
      <button
        type="button"
        data-testid="planner-send"
        :disabled="pending || draft.trim().length === 0"
        class="shrink-0 rounded-card border-2 border-line bg-fg px-4 py-2 text-xs font-bold uppercase text-bg shadow-brutal-sm active:translate-x-[2px] active:translate-y-[2px] active:shadow-none disabled:opacity-50 motion-reduce:transform-none"
        @click="onSend"
      >
        Kirim
      </button>
    </div>
  </div>
</template>
