<template>
  <AuthLayout>
    <section v-if="created" :class="style.form" role="status">
      <header :class="style.header"><h2>Владелец создан</h2></header>
      <p>Теперь вы можете войти и приглашать администраторов из раздела «Администраторы».</p>
      <RouterLink :class="style.link" to="/authorization">Перейти ко входу</RouterLink>
    </section>
    <form v-else :class="style.form" @submit.prevent="submit">
      <header :class="style.header"><h2>Первоначальная настройка</h2></header>
      <p v-if="loading" role="status">Проверяем доступность настройки…</p>
      <template v-if="available">
        <p>Создайте аккаунт владельца системы. Он сможет приглашать других администраторов.</p>
        <div :class="style.fields">
          <TextInput
            v-model="setupKey"
            label="Ключ настройки"
            name="setup-key"
            type="password"
            autocomplete="off"
            :maxlength="256"
            required
            full-width
            :disabled="submitting"
          />
          <TextInput
            v-model="lastName"
            label="Фамилия"
            name="last-name"
            autocomplete="family-name"
            :maxlength="50"
            required
            full-width
            :disabled="submitting"
          />
          <TextInput
            v-model="firstName"
            label="Имя"
            name="first-name"
            autocomplete="given-name"
            :maxlength="50"
            required
            full-width
            :disabled="submitting"
          />
          <TextInput
            v-model="email"
            label="Электронная почта"
            name="email"
            type="email"
            autocomplete="email"
            :maxlength="255"
            required
            full-width
            :disabled="submitting"
          />
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
        </div>
        <Button type="submit" size="lg" full-width :loading="submitting" :disabled="submitting"
          >Создать владельца</Button
        >
      </template>
      <p v-else-if="!loading && !error">Первоначальная настройка отключена или уже завершена.</p>

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
import { apiNotifications, ApiError, apiRequest } from '@/api/client'
import style from '../AuthForm.module.scss'

export default defineComponent({
  name: 'SetupPage',
  components: { AuthLayout, TextInput, Button, RouterLink },
  data() {
    return {
      style,
      loading: true,
      available: false,
      submitting: false,
      created: false,
      error: '',
      setupKey: '',
      firstName: '',
      lastName: '',
      email: '',
      password: '',
      confirmation: '',
    }
  },
  async mounted() {
    try {
      const status = await apiRequest<{ available: boolean }>('/auth/setup')
      this.available = status.available
    } catch (error) {
      apiNotifications.onError?.(error)
      this.error =
        error instanceof Error ? error.message : 'Не удалось проверить доступность настройки.'
    } finally {
      this.loading = false
    }
  },
  methods: {
    async submit() {
      if (this.submitting || !this.available) return
      this.error = ''
      if (this.password !== this.confirmation) {
        this.error = 'Пароли не совпадают.'
        this.$utilityStore.notifyError('Пароли не совпадают.')
        return
      }
      if ([...this.password].length < 8 || new TextEncoder().encode(this.password).length > 72) {
        this.error = 'Пароль должен содержать от 8 символов до 72 байт UTF-8.'
        this.$utilityStore.notifyError('Пароль должен содержать от 8 символов до 72 байт UTF-8.')
        return
      }
      this.submitting = true
      try {
        await apiRequest('/auth/setup', {
          method: 'POST',
          body: JSON.stringify({
            setup_key: this.setupKey,
            first_name: this.firstName.trim(),
            last_name: this.lastName.trim(),
            email: this.email.trim(),
            password: this.password,
          }),
        })
        this.created = true
        this.setupKey = ''
        this.password = ''
        this.confirmation = ''
      } catch (error) {
        apiNotifications.onError?.(error)
        this.error = error instanceof Error ? error.message : 'Не удалось создать владельца.'
        if (error instanceof ApiError && error.code === 'setup_closed') this.available = false
      } finally {
        this.submitting = false
      }
    },
  },
})
</script>
