<template>
  <AuthLayout>
    <form v-if="!receipt" :class="style.form" @submit.prevent="register">
      <header :class="style.header">
        <h2>Регистрация</h2>
      </header>

      <div :class="style.fields">
        <TextInput
          size="lg"
          v-model="email"
          :error="fieldErrors.email"
          :disabled="submitting"
          label="Электронная почта"
          name="email"
          autocomplete="email"
          placeholder="name@example.ru"
          type="email"
          required
          full-width
        />
        <TextInput
          size="lg"
          v-model="snils"
          :error="fieldErrors.snils"
          :disabled="submitting"
          label="СНИЛС"
          name="snils"
          autocomplete="snils"
          placeholder="112-233-445 95"
          type="text"
          required
          full-width
        />
        <TextInput
          size="lg"
          v-model="passport"
          :error="fieldErrors.passport"
          :disabled="submitting"
          label="Паспорт"
          name="passport"
          autocomplete="passport"
          placeholder="45 09 123456"
          type="text"
          required
          full-width
        />
        <TextInput
          size="lg"
          v-model="password"
          :error="fieldErrors.password"
          :disabled="submitting"
          label="Пароль"
          name="password"
          autocomplete="new-password"
          placeholder="Придумай пароль"
          :type="showPassword ? 'text' : 'password'"
          required
          full-width
        >
          <template #actions>
            <IconButton variant="ghost" size="sm" @click="showPassword = !showPassword">
              <PhEyeSlash v-if="showPassword" />
              <PhEye v-else />
            </IconButton>
          </template>
        </TextInput>
        <TextInput
          size="lg"
          v-model="confirmPassword"
          :disabled="submitting"
          label="Повтори пароль"
          name="password-confirmation"
          autocomplete="new-password"
          placeholder="Введи пароль ещё раз"
          :type="showConfirmation ? 'text' : 'password'"
          :error="passwordMismatch ? 'Пароли не совпадают. Проверь введённые значения.' : ''"
          required
          full-width
        >
          <template #actions>
            <IconButton variant="ghost" size="sm" @click="showConfirmation = !showConfirmation">
              <PhEyeSlash v-if="showConfirmation" />
              <PhEye v-else />
            </IconButton>
          </template>
        </TextInput>
      </div>

      <Checkbox
        v-model="acceptedTerms"
        :disabled="submitting"
        :class="style.consent"
        name="terms"
        label="Принимаю условия использования личного кабинета"
        size="sm"
      />
      <Button
        :class="style.submit"
        variant="primary"
        size="lg"
        type="submit"
        :disabled="submitting || !acceptedTerms || passwordMismatch"
        :loading="submitting"
        full-width
      >
        <span>Отправить заявку</span>
      </Button>
      <p :class="style.hint">Вход станет доступен после одобрения заявки администратором.</p>
      <p v-if="registrationError" :class="style.error" role="alert">{{ registrationError }}</p>
      <p :class="style.footer">
        Уже есть аккаунт? <RouterLink :class="style.link" to="/authorization">Войти</RouterLink>
      </p>
    </form>
    <section v-else :class="style.form" role="status">
      <header :class="style.header"><h2>Заявка отправлена</h2></header>
      <p>
        Заявка №{{ receipt.id }} ожидает рассмотрения администратором. После одобрения вы сможете
        войти с указанной почтой и паролем.
      </p>
      <RouterLink :class="style.link" to="/authorization">Перейти ко входу</RouterLink>
    </section>
  </AuthLayout>
</template>

<script lang="ts">
import { RouterLink } from 'vue-router'
import { defineComponent } from 'vue'
import { PhEye, PhEyeSlash } from '@phosphor-icons/vue'
import AuthLayout from '../AuthLayout/AuthLayout.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Checkbox from '@/components/UI/inputs/Checkbox/Checkbox.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import style from '../AuthForm.module.scss'
import { ApiError, apiRequest } from '@/api/client'

export default defineComponent({
  name: 'RegistrationComponent',
  components: {
    RouterLink,
    AuthLayout,
    Button,
    Checkbox,
    IconButton,
    PhEye,
    PhEyeSlash,
    TextInput,
  },
  data() {
    return {
      style,
      submitting: false,
      registrationError: '',
      fieldErrors: {} as Record<string, string>,
      receipt: null as { id: number; status: 'pending'; created_at: string } | null,
      email: '',
      passport: '',
      snils: '',
      password: '',
      confirmPassword: '',
      acceptedTerms: false,
      showPassword: false,
      showConfirmation: false,
    }
  },
  methods: {
    async register() {
      if (this.submitting || this.receipt) return
      this.registrationError = ''
      this.fieldErrors = {}
      const passport = this.passport.replace(/[^0-9]/g, '')
      const snils = this.snils.replace(/[^0-9]/g, '')
      if (passport.length !== 10) this.fieldErrors.passport = 'Паспорт должен содержать 10 цифр.'
      if (snils.length !== 11) this.fieldErrors.snils = 'СНИЛС должен содержать 11 цифр.'
      if ([...this.password].length < 8 || new TextEncoder().encode(this.password).length > 72) {
        this.fieldErrors.password = 'Пароль должен содержать от 8 символов до 72 байт UTF-8.'
      }
      if (
        Object.keys(this.fieldErrors).length ||
        this.password !== this.confirmPassword ||
        !this.acceptedTerms
      )
        return
      this.submitting = true
      try {
        this.receipt = await apiRequest<{ id: number; status: 'pending'; created_at: string }>(
          '/auth/register',
          {
            method: 'POST',
            body: JSON.stringify({
              email: this.email.trim(),
              password: this.password,
              passport,
              snils,
              accepted_terms: this.acceptedTerms,
              terms_version: import.meta.env.VITE_TERMS_VERSION || '2026-11-13',
            }),
          },
        )
        this.password = ''
        this.confirmPassword = ''
        this.passport = ''
        this.snils = ''
      } catch (error) {
        this.registrationError =
          error instanceof Error ? error.message : 'Не удалось отправить заявку.'
        if (error instanceof ApiError) {
          const fieldMessages: Record<string, string> = {
            email: 'Введите корректную электронную почту длиной до 255 символов.',
            passport: 'Паспорт должен содержать 10 цифр.',
            snils: 'СНИЛС должен содержать 11 цифр.',
            password: 'Пароль должен содержать от 8 символов до 72 байт UTF-8.',
          }
          for (const field of Object.keys(error.fields)) {
            if (fieldMessages[field]) this.fieldErrors[field] = fieldMessages[field]
          }
          if (error.fields.terms_version) {
            this.registrationError =
              'Условия использования обновились. Обновите страницу или обратитесь к администратору.'
          }
        }
      } finally {
        this.submitting = false
      }
    },
  },
  computed: {
    passwordMismatch(): boolean {
      return this.confirmPassword.length > 0 && this.password !== this.confirmPassword
    },
  },
})
</script>
