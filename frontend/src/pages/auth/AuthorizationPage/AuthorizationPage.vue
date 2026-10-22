<template>
  <AuthLayout title-id="authorization-title">
    <form :class="style.form" @submit.prevent>
      <header :class="style.header">
        <h2 id="authorization-title">Вход в аккаунт</h2>
      </header>

      <div :class="style.fields">
        <TextInput
          v-model="email"
          label="Электронная почта"
          name="username"
          autocomplete="username"
          placeholder="name@example.ru"
          type="email"
          required
          full-width
        />
        <TextInput
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
            <IconButton variant="ghost" size="sm" @click="showPassword = !showPassword">
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
    </form>
  </AuthLayout>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhEye, PhEyeSlash } from '@phosphor-icons/vue'
import AuthLayout from '../AuthLayout/AuthLayout.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Checkbox from '@/components/UI/inputs/Checkbox/Checkbox.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import style from '../AuthForm.module.scss'

export default defineComponent({
  name: 'AuthorizationPage',
  components: {
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
      rememberMe: false,
      showPassword: false,
    }
  },
})
</script>
