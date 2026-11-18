<template>
  <section :class="style.page">
    <SectionHeader
      header="Администраторы"
      paragraph="Приглашения и доступ к административной панели"
    />
    <p v-if="success" :class="style.success" role="status">{{ success }}</p>
    <p v-if="error" :class="style.error" role="alert">{{ error }}</p>
    <div :class="style.toolbar">
      <h2>Аккаунты администраторов</h2>
      <Button variant="secondary" :loading="loading" :disabled="loading || submitting" @click="load"
        >Обновить</Button
      >
    </div>
    <div :class="style.panel">
      <div :class="style.tableScroll">
        <table :class="style.table">
          <thead>
            <tr>
              <th>Пользователь</th>
              <th>Доступ</th>
              <th>Состояние</th>
              <th>Приглашение</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in administrators" :key="item.id">
              <td>
                <strong>{{ item.full_name }}</strong
                ><small>{{ item.email }}</small>
              </td>
              <td>
                {{
                  item.can_manage_admins
                    ? 'Владелец'
                    : item.credential_state === 'invited'
                      ? 'Ожидает активации'
                      : 'Администратор'
                }}
              </td>
              <td>
                {{
                  !item.is_active
                    ? 'Заблокирован'
                    : item.credential_state === 'invited'
                      ? 'Пароль не задан'
                      : 'Активен'
                }}
              </td>
              <td>
                <Button
                  v-if="item.credential_state === 'invited' && item.is_active"
                  variant="ghost"
                  :disabled="submitting"
                  @click="prepareResend(item)"
                  >Отправить повторно</Button
                >
                <span v-else>—</span>
                <small v-if="item.invitation_expires_at"
                  >До {{ new Date(item.invitation_expires_at).toLocaleString('ru-RU') }}</small
                >
              </td>
            </tr>
            <tr v-if="!loading && !administrators.length">
              <td colspan="4">Администраторы не загружены.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <form :class="style.form" @submit.prevent="submit">
      <h2>Предоставить доступ</h2>
      <div :class="style.tabs">
        <StyleTab text="Пригласить нового" :active="mode === 'invite'" @select="mode = 'invite'" />
        <StyleTab
          text="Назначить существующего"
          :active="mode === 'promote'"
          @select="mode = 'promote'"
        />
      </div>
      <p :class="style.muted">
        {{
          mode === 'invite'
            ? 'Пользователь получит письмо и самостоятельно задаст пароль. Ссылка действует 24 часа. Повторная отправка отменяет предыдущую ссылку.'
            : 'Укажите почту действующего пользователя. После назначения ему потребуется войти заново.'
        }}
      </p>
      <fieldset :disabled="submitting" :class="pageStyle.fields">
        <div v-if="mode === 'invite'" :class="style.grid">
          <TextInput
            v-model="lastName"
            label="Фамилия"
            name="last-name"
            :maxlength="50"
            required
            full-width
          />
          <TextInput
            v-model="firstName"
            label="Имя"
            name="first-name"
            :maxlength="50"
            required
            full-width
          />
        </div>
        <TextInput
          v-model="email"
          label="Электронная почта пользователя"
          name="email"
          type="email"
          :maxlength="255"
          autocomplete="off"
          required
          full-width
        />
        <TextInput
          v-model="currentPassword"
          label="Ваш текущий пароль для подтверждения"
          name="current-password"
          type="password"
          autocomplete="current-password"
          required
          full-width
        />
      </fieldset>
      <div :class="style.actions">
        <Button type="submit" :loading="submitting" :disabled="submitting">{{
          mode === 'invite' ? 'Отправить приглашение' : 'Назначить администратором'
        }}</Button>
      </div>
    </form>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import { useAuthStore } from '@/stores/auth'
import style from './Admin.module.scss'
import pageStyle from './AdministratorsPage.module.scss'

interface Administrator {
  id: number
  email: string
  full_name: string
  first_name: string
  last_name: string
  can_manage_admins: boolean
  is_active: boolean
  credential_state: 'invited' | 'active'
  invitation_expires_at: string | null
}

export default defineComponent({
  name: 'AdministratorsPage',
  components: { SectionHeader, TextInput, Button, StyleTab },
  data() {
    return {
      style,
      pageStyle,
      auth: useAuthStore(),
      administrators: [] as Administrator[],
      loading: false,
      submitting: false,
      error: '',
      success: '',
      mode: 'invite' as 'invite' | 'promote',
      email: '',
      firstName: '',
      lastName: '',
      currentPassword: '',
    }
  },
  mounted() {
    void this.load()
  },
  methods: {
    async load() {
      if (this.loading) return
      this.loading = true
      this.error = ''
      try {
        this.administrators = await this.auth.request<Administrator[]>('/admin/administrators')
      } catch (error) {
        this.error =
          error instanceof Error ? error.message : 'Не удалось загрузить администраторов.'
      } finally {
        this.loading = false
      }
    },
    prepareResend(item: Administrator) {
      this.mode = 'invite'
      this.email = item.email
      this.firstName = item.first_name
      this.lastName = item.last_name
      this.currentPassword = ''
      this.success = 'Введите свой текущий пароль и подтвердите повторную отправку приглашения.'
    },
    async submit() {
      if (this.submitting) return
      this.submitting = true
      this.error = ''
      this.success = ''
      const mode = this.mode
      try {
        await this.auth.request<void>(
          `/admin/administrators/${mode === 'invite' ? 'invitations' : 'promote'}`,
          {
            method: 'POST',
            body: JSON.stringify({
              email: this.email.trim(),
              current_password: this.currentPassword,
              ...(mode === 'invite'
                ? { first_name: this.firstName.trim(), last_name: this.lastName.trim() }
                : {}),
            }),
          },
        )
        this.success =
          mode === 'invite'
            ? 'Приглашение поставлено в очередь отправки. Права появятся после активации.'
            : 'Пользователь назначен администратором. Для доступа ему нужно войти заново.'
        this.email = ''
        this.firstName = ''
        this.lastName = ''
        await this.load()
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось предоставить доступ.'
      } finally {
        this.currentPassword = ''
        this.submitting = false
      }
    },
  },
})
</script>
