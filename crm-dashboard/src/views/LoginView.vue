<template>
  <div class="login-page">
    <div class="login-card">
      <!-- Logo / Branding -->
      <div class="login-header">
        <div class="brand-icon">
          <svg viewBox="0 0 40 40" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect width="40" height="40" rx="10" fill="#2563EB" />
            <path d="M10 20L18 12L26 20L18 28L10 20Z" fill="white" fill-opacity="0.9" />
            <path d="M18 12L30 20L18 28" stroke="white" stroke-width="2" stroke-opacity="0.6" />
          </svg>
        </div>
        <h1>G4S CRM</h1>
        <p>Sign in to your account</p>
      </div>

      <!-- Error Banner -->
      <div v-if="authStore.error" class="error-banner">
        <span>{{ authStore.error }}</span>
      </div>

      <!-- Form -->
      <form class="login-form" @submit.prevent="handleLogin">
        <div class="field">
          <label for="email">Email address</label>
          <input
            id="email"
            v-model="form.email"
            type="email"
            placeholder="you@g4s.com"
            autocomplete="email"
            required
            :disabled="authStore.loading"
          />
        </div>

        <div class="field">
          <label for="password">
            Password
            <a href="#" class="forgot-link" @click.prevent>Forgot password?</a>
          </label>
          <div class="password-wrap">
            <input
              id="password"
              v-model="form.password"
              :type="showPassword ? 'text' : 'password'"
              placeholder="••••••••"
              autocomplete="current-password"
              required
              :disabled="authStore.loading"
            />
            <button
              type="button"
              class="toggle-pw"
              tabindex="-1"
              @click="showPassword = !showPassword"
            >
              <svg v-if="!showPassword" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
                <circle cx="12" cy="12" r="3" />
              </svg>
              <svg v-else viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24" />
                <line x1="1" y1="1" x2="23" y2="23" />
              </svg>
            </button>
          </div>
        </div>

        <button type="submit" class="submit-btn" :disabled="authStore.loading">
          <span v-if="authStore.loading" class="spinner" />
          <span>{{ authStore.loading ? 'Signing in…' : 'Sign in' }}</span>
        </button>
      </form>

      <p class="version">G4S CRM v1.0 · Internal Use Only</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const form = ref({ email: '', password: '' })
const showPassword = ref(false)

async function handleLogin() {
  try {
    await authStore.login(form.value.email, form.value.password)
    await router.push({ name: 'dashboard' })
  } catch {
    // error already set in store
  }
}
</script>

<style scoped>
.login-page {
  width: 100%;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #0f172a 0%, #1e3a5f 50%, #0f172a 100%);
  padding: 1rem;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: #1e293b;
  border: 1px solid #334155;
  border-radius: 16px;
  padding: 2.5rem 2rem;
  box-shadow: 0 25px 50px rgba(0, 0, 0, 0.5);
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.brand-icon {
  display: flex;
  justify-content: center;
  margin-bottom: 0.75rem;
}

.brand-icon svg {
  width: 48px;
  height: 48px;
}

.login-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: #f1f5f9;
  margin: 0 0 0.25rem;
  letter-spacing: -0.01em;
}

.login-header p {
  color: #94a3b8;
  font-size: 0.9rem;
  margin: 0;
}

.error-banner {
  background: rgba(239, 68, 68, 0.12);
  border: 1px solid rgba(239, 68, 68, 0.3);
  border-radius: 8px;
  padding: 0.75rem 1rem;
  color: #fca5a5;
  font-size: 0.875rem;
  margin-bottom: 1.25rem;
  text-align: center;
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
}

.field {
  display: flex;
  flex-direction: column;
  gap: 0.4rem;
}

.field label {
  font-size: 0.875rem;
  font-weight: 500;
  color: #cbd5e1;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.forgot-link {
  font-size: 0.8rem;
  color: #60a5fa;
  text-decoration: none;
}

.forgot-link:hover {
  color: #93c5fd;
}

.field input {
  background: #0f172a;
  border: 1px solid #334155;
  border-radius: 8px;
  padding: 0.65rem 0.875rem;
  color: #f1f5f9;
  font-size: 0.9rem;
  outline: none;
  transition: border-color 0.15s;
  width: 100%;
  box-sizing: border-box;
}

.field input:focus {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.field input::placeholder {
  color: #475569;
}

.field input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.password-wrap {
  position: relative;
}

.password-wrap input {
  padding-right: 2.75rem;
}

.toggle-pw {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  cursor: pointer;
  color: #64748b;
  padding: 0;
  display: flex;
  align-items: center;
}

.toggle-pw:hover {
  color: #94a3b8;
}

.toggle-pw svg {
  width: 18px;
  height: 18px;
}

.submit-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  background: #2563eb;
  color: white;
  border: none;
  border-radius: 8px;
  padding: 0.75rem;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s;
  margin-top: 0.25rem;
}

.submit-btn:hover:not(:disabled) {
  background: #1d4ed8;
}

.submit-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.spinner {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
  flex-shrink: 0;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.version {
  text-align: center;
  color: #475569;
  font-size: 0.75rem;
  margin: 1.5rem 0 0;
}
</style>
