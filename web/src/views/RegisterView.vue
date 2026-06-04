<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import axios from 'axios'
import { register } from '@/api/auth'

const router = useRouter()

const name = ref('')
const email = ref('')
const password = ref('')
const submitting = ref(false)
const errorMsg = ref<string | null>(null)

async function onSubmit() {
  errorMsg.value = null

  const nameValue = name.value.trim()
  const emailValue = email.value.trim()
  if (!nameValue) {
    errorMsg.value = 'Isi nama kamu dulu ya.'
    return
  }
  if (!emailValue) {
    errorMsg.value = 'Isi email kamu dulu ya.'
    return
  }
  if (password.value.length < 8) {
    errorMsg.value = 'Password minimal 8 karakter.'
    return
  }

  submitting.value = true
  try {
    await register(nameValue, emailValue, password.value)
    // New users need a budget — drop them straight into onboarding.
    router.push({ name: 'onboarding' })
  } catch (err) {
    if (axios.isAxiosError(err) && err.response?.status === 409) {
      errorMsg.value = 'Email sudah terdaftar, masuk aja.'
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
    data-testid="register-view"
  >
    <header class="space-y-2 text-center">
      <p class="text-xs uppercase tracking-[0.18em] text-muted">Fintrack</p>
      <h1 class="font-display text-3xl font-semibold">Mulai latihan</h1>
      <p class="text-sm text-muted">Buat akun dan susun budget pertamamu.</p>
    </header>

    <form novalidate class="space-y-4" @submit.prevent="onSubmit">
      <fieldset class="space-y-2">
        <label class="text-xs uppercase tracking-wider text-muted" for="register-name">
          Nama
        </label>
        <input
          id="register-name"
          v-model="name"
          type="text"
          autocomplete="name"
          placeholder="Nama kamu"
          data-testid="register-name"
          class="w-full rounded-card border border-line bg-surface px-4 py-3 text-sm focus:border-saffron focus:outline-none"
        />
      </fieldset>

      <fieldset class="space-y-2">
        <label class="text-xs uppercase tracking-wider text-muted" for="register-email">
          Email
        </label>
        <input
          id="register-email"
          v-model="email"
          type="email"
          inputmode="email"
          autocomplete="email"
          placeholder="kamu@email.com"
          data-testid="register-email"
          class="w-full rounded-card border border-line bg-surface px-4 py-3 text-sm focus:border-saffron focus:outline-none"
        />
      </fieldset>

      <fieldset class="space-y-2">
        <label class="text-xs uppercase tracking-wider text-muted" for="register-password">
          Password
        </label>
        <input
          id="register-password"
          v-model="password"
          type="password"
          autocomplete="new-password"
          placeholder="Minimal 8 karakter"
          data-testid="register-password"
          class="w-full rounded-card border border-line bg-surface px-4 py-3 text-sm focus:border-saffron focus:outline-none"
        />
      </fieldset>

      <p v-if="errorMsg" data-testid="register-error" class="text-sm text-fatigued">
        {{ errorMsg }}
      </p>

      <button
        type="submit"
        :disabled="submitting"
        data-testid="register-submit"
        class="w-full rounded-card bg-saffron py-3 font-semibold text-bg disabled:opacity-50"
      >
        {{ submitting ? 'Mendaftar…' : 'Daftar' }}
      </button>
    </form>

    <p class="text-center text-sm text-muted">
      Sudah punya akun?
      <router-link
        :to="{ name: 'login' }"
        data-testid="register-to-login"
        class="text-saffron hover:underline"
      >
        Masuk
      </router-link>
    </p>
  </section>
</template>
