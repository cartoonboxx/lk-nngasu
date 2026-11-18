<template>
  <section :class="style.settingsPage">
    <SectionHeader header="Настройки" paragraph="Основные параметры личного кабинета" />

    <section :class="style.settingsCard">
      <div :class="style.settingsSection">
        <h2>Фото профиля</h2>

        <div :class="style.profileRow">
          <div :class="style.profileAvatar">
            <img
              v-if="auth.avatarUrl"
              :src="auth.avatarUrl"
              alt="Фото профиля"
              :class="style.profileAvatarImage"
            />
            <span v-else
              >{{ auth.user?.first_name.slice(0, 1) }}{{ auth.user?.last_name.slice(0, 1) }}</span
            >
          </div>

          <div :class="style.profileInfo">
            <p>{{ auth.user?.full_name }}</p>
            <p>{{ auth.roleLabel }}</p>
          </div>

          <input
            ref="avatarInput"
            type="file"
            accept="image/jpeg,image/png,image/webp"
            hidden
            @change="uploadAvatar"
          />
          <Button
            variant="secondary"
            :disabled="savingAvatar"
            :loading="savingAvatar"
            @click="($refs.avatarInput as HTMLInputElement).click()"
            >Изменить фото</Button
          >
          <Button
            v-if="auth.user?.avatar"
            variant="ghost"
            :disabled="savingAvatar"
            @click="removeAvatar"
            >Удалить фото</Button
          >
        </div>
      </div>

      <div :class="style.settingsSectionRow">
        <div>
          <h2>Тема интерфейса</h2>
          <p>Выберите удобное оформление</p>
        </div>

        <div :class="style.tabColumn">
          <StyleTab
            v-for="item in themeOptions"
            :key="item.value"
            :text="item.label"
            :active="item.value === $utilityStore.themePreference"
            @select="$utilityStore.themePreference = item.value"
          />
        </div>
      </div>

      <div :class="style.settingsSectionRow">
        <div>
          <h2>Пароль</h2>
          <p>Изменить пароль для входа</p>
        </div>

        <div>
          <Button
            variant="secondary"
            @click="
              () => {
                resetPasswordForm()
                isPasswordModalOpen = true
              }
            "
            >Изменить пароль</Button
          >
        </div>
      </div>

      <div :class="style.settingsSectionRow">
        <div>
          <h2>Аккаунт</h2>
          <p>{{ auth.user?.email }}</p>
        </div>
        <Button variant="secondary" :loading="loggingOut" :disabled="loggingOut" @click="signOut"
          >Выйти из аккаунта</Button
        >
      </div>

      <Modal v-model="isPasswordModalOpen" :width="460" @close="resetPasswordForm">
        <template #default="{ close }">
          <form :class="style.passwordForm" @submit.prevent="submitPasswordChange">
            <h2 id="password-modal-title">Изменить пароль</h2>

            <TextInput
              v-model="currentPassword"
              label="Текущий пароль"
              :type="showCurrentPassword ? 'text' : 'password'"
              autocomplete="current-password"
              fullWidth
              size="lg"
              required
            >
              <template #actions>
                <IconButton
                  size="sm"
                  variant="ghost"
                  type="button"
                  :title="showCurrentPassword ? 'Скрыть пароль' : 'Показать пароль'"
                  @click="showCurrentPassword = !showCurrentPassword"
                >
                  <PhEyeSlash v-if="showCurrentPassword" />
                  <PhEye v-else />
                </IconButton>
              </template>
            </TextInput>

            <TextInput
              v-model="newPassword"
              label="Новый пароль"
              :type="showNewPassword ? 'text' : 'password'"
              autocomplete="new-password"
              fullWidth
              size="lg"
              required
            >
              <template #actions>
                <IconButton
                  size="sm"
                  variant="ghost"
                  type="button"
                  :title="showNewPassword ? 'Скрыть пароль' : 'Показать пароль'"
                  @click="showNewPassword = !showNewPassword"
                >
                  <PhEyeSlash v-if="showNewPassword" />
                  <PhEye v-else />
                </IconButton>
              </template>
            </TextInput>

            <TextInput
              v-model="confirmPassword"
              label="Повторите новый пароль"
              :type="showConfirmPassword ? 'text' : 'password'"
              autocomplete="new-password"
              fullWidth
              size="lg"
              required
            >
              <template #actions>
                <IconButton
                  size="sm"
                  variant="ghost"
                  type="button"
                  :title="showConfirmPassword ? 'Скрыть пароль' : 'Показать пароль'"
                  @click="showConfirmPassword = !showConfirmPassword"
                >
                  <PhEyeSlash v-if="showConfirmPassword" />
                  <PhEye v-else />
                </IconButton>
              </template>
            </TextInput>

            <div :class="style.passwordActions">
              <Button variant="secondary" type="button" :disabled="savingPassword" @click="close">
                Отмена
              </Button>

              <Button type="submit" :disabled="savingPassword" :loading="savingPassword">
                Сохранить
              </Button>
            </div>
          </form>
        </template>
      </Modal>
    </section>
  </section>
</template>

<script lang="ts">
import { apiNotifications } from '@/api/client'
import { useFilesStore } from '@/stores/files'
import { useAuthStore } from '@/stores/auth'
import { defineComponent } from 'vue'
import { PhEye, PhEyeSlash } from '@phosphor-icons/vue'

import style from './SettingsPage.module.scss'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import type { ThemePreference } from '@/stores/utility'

export default defineComponent({
  name: 'SettingsPage',

  components: {
    SectionHeader,
    StyleTab,
    Modal,
    Button,
    IconButton,
    TextInput,
    PhEye,
    PhEyeSlash,
  },

  computed: {
    auth() {
      return useAuthStore()
    },
  },
  data() {
    return {
      style,
      loggingOut: false,
      savingAvatar: false,
      savingPassword: false,

      isPasswordModalOpen: false,
      showCurrentPassword: false,
      showNewPassword: false,
      showConfirmPassword: false,
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
      themeOptions: [
        { value: 'system', label: 'Системная' },
        { value: 'light', label: 'Светлая' },
        { value: 'dark', label: 'Тёмная' },
      ] satisfies { value: ThemePreference; label: string }[],
    }
  },

  methods: {
    async uploadAvatar(event: Event) {
      const input = event.target as HTMLInputElement
      const file = input.files?.[0]
      input.value = ''
      if (!file || this.savingAvatar) return

      if (
        file.size > 2 * 1024 * 1024 ||
        !['image/jpeg', 'image/png', 'image/webp'].includes(file.type)
      ) {
        this.$utilityStore.notifyError('Выберите JPEG, PNG или WebP размером до 2 МБ.')
        return
      }
      const userId = this.auth.user?.id
      this.savingAvatar = true
      try {
        const uploaded = await useFilesStore().upload(file, 'avatar')
        if (this.auth.user?.id !== userId) return
        await this.auth.updateMe({ avatar_file_id: uploaded.id })
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.savingAvatar = false
      }
    },
    async removeAvatar() {
      if (this.savingAvatar) return
      this.savingAvatar = true

      try {
        await this.auth.updateMe({ avatar_file_id: null })
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.savingAvatar = false
      }
    },
    async signOut() {
      if (this.loggingOut) return
      this.loggingOut = true

      try {
        await this.auth.signOut()
        await this.$router.replace({ name: 'authorization' })
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.loggingOut = false
      }
    },
    resetPasswordForm() {
      this.currentPassword = ''
      this.newPassword = ''
      this.confirmPassword = ''

      this.showCurrentPassword = false
      this.showNewPassword = false
      this.showConfirmPassword = false
    },

    async submitPasswordChange() {
      if (this.savingPassword) return

      if (this.newPassword !== this.confirmPassword) {
        this.$utilityStore.notifyError('Новые пароли не совпадают.')
        return
      }

      if (this.newPassword === this.currentPassword) {
        this.$utilityStore.notifyError('Новый пароль должен отличаться от текущего.')
        return
      }

      if (
        [...this.newPassword].length < 8 ||
        new TextEncoder().encode(this.newPassword).length > 72
      ) {
        this.$utilityStore.notifyError('Пароль должен содержать от 8 символов до 72 байт UTF-8.')
        return
      }
      this.savingPassword = true
      try {
        await this.auth.changePassword(this.currentPassword, this.newPassword)
        this.resetPasswordForm()
        await this.$router.replace({ name: 'authorization', query: { reason: 'password-changed' } })
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.savingPassword = false
      }
    },
  },
})
</script>
