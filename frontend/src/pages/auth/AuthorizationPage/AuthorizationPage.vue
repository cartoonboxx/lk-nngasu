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
          :label="devPreview ? 'Логин или электронная почта' : 'Электронная почта'"
          name="username"
          autocomplete="username"
          placeholder="name@example.ru"
          :type="devPreview ? 'text' : 'email'"
          required
          full-width
        />
        <TextInput
          size="lg"
          v-model="password"
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

      <Button :class="style.submit" variant="primary" size="lg" type="submit" full-width>
        <span>Войти в кабинет</span>
      </Button>
      <p :class="style.footer">
        Пока нет аккаунта?
        <RouterLink :class="style.link" to="/registration">Зарегистрироваться</RouterLink>
      </p>
      <p v-if="signInError" :class="style.error" role="alert">{{ signInError }}</p>
      <p v-if="devPreview" :class="style.hint">
        В разработке доступен временный вход в локальную административную панель.
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
import { signInAdminDemo } from '../admin-demo'

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
      devPreview: import.meta.env.DEV,
      login: '',
      signInError: '',
      password: '',
      rememberMe: false,
      showPassword: false,
    }
  },
  methods: {
    async signIn() {
      this.signInError = ''
      if (signInAdminDemo(this.login, this.password)) {
        this.password = ''
        await this.$router.replace('/admin')
        return
      }
      this.password = ''
      this.signInError =
        this.devPreview && this.login === 'root'
          ? 'Неверный логин или пароль.'
          : 'Серверный вход пока не подключён.'
    },
  },
})
</script>
