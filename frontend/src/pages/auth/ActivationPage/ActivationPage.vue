<template>
  <AuthLayout>
    <section v-if="activated" :class="style.form" role="status">
      <header :class="style.header"><h2>Аккаунт активирован</h2></header>
      <p>Пароль сохранён. Войдите с почтой, на которую получили приглашение.</p>
      <RouterLink :class="style.link" to="/authorization">Перейти ко входу</RouterLink>
    </section>
    <form v-else :class="style.form" @submit.prevent="submit">
      <header :class="style.header"><h2>Активация аккаунта</h2></header>
      <template v-if="token">
        <p>Задайте пароль для входа в личный кабинет.</p>
        <TextInput
          v-model="password"
          label="Пароль"
          name="password"
          type="password"
          autocomplete="new-password"
          required
          full-width
          :disabled="submitting"
        />
        <TextInput
          v-model="confirmation"
          label="Повторите пароль"
          name="password-confirmation"
          type="password"
          autocomplete="new-password"
          required
          full-width
          :disabled="submitting"
        />
        <Button type="submit" size="lg" full-width :loading="submitting" :disabled="submitting"
          >Активировать аккаунт</Button
        >
      </template>
      <p v-else>Откройте ссылку из письма с приглашением.</p>
      <p v-if="error" :class="style.error" role="alert">{{ error }}</p>
      <RouterLink :class="style.link" to="/authorization">Перейти ко входу</RouterLink>
    </form>
  </AuthLayout>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import AuthLayout from '../AuthLayout/AuthLayout.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import { apiRequest } from '@/api/client'
import style from '../AuthForm.module.scss'

export default defineComponent({
  name: 'ActivationPage',
  components: { AuthLayout, TextInput, Button, RouterLink },
  data() {
    const token =
      new URLSearchParams(this.$route.hash.slice(1)).get('token') ?? this.$route.query.token
    return {
      style,
      token: typeof token === 'string' ? token : '',
      password: '',
      confirmation: '',
      error: '',
      submitting: false,
      activated: false,
    }
  },
  mounted() {
    // Сохраняем токен в памяти и убираем его из адресной строки и истории.
    if (this.token) void this.$router.replace({ name: 'activation' })
  },
  methods: {
    async submit() {
      if (this.submitting || !this.token) return
      this.error = ''
      if (this.password !== this.confirmation) {
        this.error = 'Пароли не совпадают.'
        return
      }
      if ([...this.password].length < 8 || new TextEncoder().encode(this.password).length > 72) {
        this.error = 'Пароль должен содержать от 8 символов до 72 байт UTF-8.'
        return
      }
      this.submitting = true
      try {
        await apiRequest<void>('/auth/activate', {
          method: 'POST',
          body: JSON.stringify({ token: this.token, password: this.password }),
        })
        this.activated = true
        this.token = ''
        this.password = ''
        this.confirmation = ''
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось активировать аккаунт.'
      } finally {
        this.submitting = false
      }
    },
  },
})
</script>
