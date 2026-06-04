<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import axios from 'axios'
import { login } from '@/api/auth'

const router = useRouter()
const route = useRoute()

const email = ref('')
const password = ref('')
const submitting = ref(false)
const errorMsg = ref<string | null>(null)

async function onSubmit() {
  errorMsg.value = null

  const value = email.value.trim()
  if (!value) {
    errorMsg.value = 'Isi email kamu dulu ya.'
    return
  }
  if (!password.value) {
    errorMsg.value = 'Isi password kamu dulu ya.'
    return
  }

  submitting.value = true
  try {
    await login(value, password.value)
    const redirect = route.query.redirect
    // Only honour an in-app, single-leading-slash path so a crafted ?redirect=
    // can't bounce the user off-site or to a protocol-relative URL.
    if (typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')) {
      router.push(redirect)
    } else {
      router.push({ name: 'home' })
    }
  } catch (err) {
    // Backend returns a generic 401 ("Email atau password salah.") — surface it.
    if (axios.isAxiosError(err) && err.response?.data?.error?.message) {
      errorMsg.value = err.response.data.error.message
    } else {
      errorMsg.value = err instanceof Error ? err.message : String(err)
    }
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <section
    class="mx-auto flex min-h-screen max-w-mobile flex-col justify-center gap-8 px-6 py-10"
    data-testid="login-view"
  >
    <header class="space-y-2 text-center">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Fintrack</p>
      <h1 class="font-display text-3xl font-semibold">Latih duitmu</h1>
      <p class="text-sm text-muted">Disiplin keuangan yang rasanya kayak latihan, bukan pembukuan.</p>
    </header>

    <form novalidate class="space-y-4" @submit.prevent="onSubmit">
      <fieldset class="space-y-2">
        <label class="text-xs uppercase tracking-wider text-muted" for="login-email">
          Email
        </label>
        <input
          id="login-email"
          v-model="email"
          type="email"
          inputmode="email"
          autocomplete="email"
          placeholder="kamu@email.com"
          data-testid="login-email"
          class="w-full rounded-card border border-line bg-surface px-4 py-3 text-sm focus:border-saffron focus:outline-none"
        />
      </fieldset>

      <fieldset class="space-y-2">
        <label class="text-xs uppercase tracking-wider text-muted" for="login-password">
          Password
        </label>
        <input
          id="login-password"
          v-model="password"
          type="password"
          autocomplete="current-password"
          placeholder="••••••••"
          data-testid="login-password"
          class="w-full rounded-card border border-line bg-surface px-4 py-3 text-sm focus:border-saffron focus:outline-none"
        />
      </fieldset>

      <p v-if="errorMsg" data-testid="login-error" class="text-sm text-fatigued">
        {{ errorMsg }}
      </p>

      <button
        type="submit"
        :disabled="submitting"
        data-testid="login-submit"
        class="w-full rounded-card bg-saffron py-3 font-semibold text-bg disabled:opacity-50"
      >
        {{ submitting ? 'Masuk…' : 'Masuk' }}
      </button>
    </form>

    <p class="text-center text-sm text-muted">
      Belum punya akun?
      <router-link
        :to="{ name: 'register' }"
        data-testid="login-to-register"
        class="text-saffron hover:underline"
      >
        Daftar
      </router-link>
    </p>
  </section>
</template>
