<template>
  <AuthLayout>
    <form :class="style.form" @submit.prevent="signIn">
      <header :class="style.header">
        <h2>Вход в аккаунт</h2>
      </header>

      <div :class="style.fields">
        <TextInput
          size="lg"
          v-model="login"
          label="Электронная почта"
          name="username"
          autocomplete="username"
          placeholder="name@example.ru"
          type="email"
          required
          :disabled="submitting"
          full-width
        />
        <TextInput
          size="lg"
          v-model="password"
          :disabled="submitting"
          label="Пароль"
          name="password"
          autocomplete="current-password"
          placeholder="Введи пароль"
          :type="showPassword ? 'text' : 'password'"
          required
          full-width
        >
          <template #actions>
            <IconButton variant="ghost" size="sm" round @click="showPassword = !showPassword">
              <PhEyeSlash v-if="showPassword" />
              <PhEye v-else />
            </IconButton>
          </template>
        </TextInput>
      </div>

      <Checkbox
        v-model="rememberMe"
        :class="style.remember"
        name="remember"
        label="Запомнить меня"
        size="sm"
      />

      <Button
        :class="style.submit"
        variant="primary"
        size="lg"
        type="submit"
        :loading="submitting"
        :disabled="submitting"
        full-width
      >
        <span>Войти в кабинет</span>
      </Button>
      <p :class="style.footer">
        Пока нет аккаунта?
        <RouterLink :class="style.link" to="/registration">Зарегистрироваться</RouterLink>
      </p>
      <p v-if="setupAvailable" :class="style.footer">
        <RouterLink :class="style.link" to="/setup"
          >Настроить первый аккаунт администратора</RouterLink
        >
      </p>
      <p v-if="signInError" :class="style.error" role="alert">{{ signInError }}</p>
      <p v-if="$route.query.reason === 'password-changed'" role="status">
        Пароль изменён. Войдите с новым паролем.
      </p>
      <p v-if="$route.query.reason === 'server-unavailable'" :class="style.error" role="alert">
        Не удалось восстановить сессию. Проверьте подключение к серверу и повторите вход.
      </p>
    </form>
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
import { useAuthStore } from '@/stores/auth'
import { apiRequest } from '@/api/client'

export default defineComponent({
  name: 'AuthorizationPage',
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
      auth: useAuthStore(),
      submitting: false,
      setupAvailable: false,
      login: '',
      signInError: '',
      password: '',
      rememberMe: false,
      showPassword: false,
    }
  },
  async mounted() {
    try {
      const status = await apiRequest<{ available: boolean }>('/auth/setup')
      this.setupAvailable = status.available
    } catch {
      this.setupAvailable = false
    }
  },
  methods: {
    async signIn() {
      if (this.submitting) return
      this.submitting = true
      this.signInError = ''
      try {
        await this.auth.signIn(this.login, this.password, this.rememberMe)
        this.password = ''
        const redirect = this.$route.query.redirect
        await this.$router.replace(
          typeof redirect === 'string' && redirect.startsWith('/') && !redirect.startsWith('//')
            ? redirect
            : { name: this.auth.user?.is_admin ? 'admin-requests' : 'dashboard' },
        )
      } catch (error) {
        this.signInError =
          error instanceof Error ? error.message : 'Не удалось войти. Повторите попытку.'
      } finally {
        this.submitting = false
      }
    },
  },
})
</script>
