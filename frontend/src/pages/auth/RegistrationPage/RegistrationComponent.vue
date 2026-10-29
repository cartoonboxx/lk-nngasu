<template>
  <AuthLayout>
    <form :class="style.form" @submit.prevent>
      <header :class="style.header">
        <h2>Регистрация</h2>
      </header>

      <div :class="style.fields">
        <TextInput
          size="lg"
          v-model="email"
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
          v-model="password"
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
        :disabled="!acceptedTerms || passwordMismatch"
        full-width
      >
        <span>Создать аккаунт</span>
      </Button>
      <p :class="style.footer">
        Уже есть аккаунт? <RouterLink :class="style.link" to="/authorization">Войти</RouterLink>
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
      email: '',
      password: '',
      confirmPassword: '',
      acceptedTerms: false,
      showPassword: false,
      showConfirmation: false,
    }
  },
  computed: {
    passwordMismatch(): boolean {
      return this.confirmPassword.length > 0 && this.password !== this.confirmPassword
    },
  },
})
</script>
