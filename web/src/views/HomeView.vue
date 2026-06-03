<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { getHealth, type HealthResponse } from '@/api/health'

const health = ref<HealthResponse | null>(null)
const error = ref<string | null>(null)
const loading = ref(true)

onMounted(async () => {
  try {
    health.value = await getHealth()
  } catch (e) {
    error.value = e instanceof Error ? e.message : String(e)
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <section class="mx-auto flex max-w-mobile flex-col items-start gap-8 px-6 py-12">
    <header class="space-y-2">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Phase 0 — Hello</p>
      <h1 class="font-display text-3xl font-semibold">Fintrack</h1>
      <p class="text-sm text-muted">
        Money discipline that feels like training, not bookkeeping.
      </p>
    </header>

    <div data-testid="health-card" class="w-full rounded-card border border-line bg-surface p-5">
      <p class="mb-3 text-xs uppercase tracking-wider text-muted">API status</p>

      <p v-if="loading" data-testid="health-loading" class="font-mono text-sm">checking…</p>

      <div v-else-if="error" data-testid="health-error" class="space-y-1">
        <p class="font-mono text-sm text-fatigued">offline</p>
        <p class="text-xs text-muted">{{ error }}</p>
      </div>

      <dl v-else class="space-y-2 font-mono text-sm">
        <div class="flex justify-between">
          <dt class="text-muted">status</dt>
          <dd data-testid="health-status" class="text-fresh">{{ health?.status }}</dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-muted">db</dt>
          <dd
            data-testid="health-db"
            :class="health?.db === 'ok' ? 'text-fresh' : 'text-fatigued'"
          >
            {{ health?.db }}
          </dd>
        </div>
        <div class="flex justify-between">
          <dt class="text-muted">version</dt>
          <dd data-testid="health-version">{{ health?.version }}</dd>
        </div>
      </dl>
    </div>
  </section>
</template>
