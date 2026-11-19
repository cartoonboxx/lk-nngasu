<template>
  <section :class="style.settingsPage">
    <SectionHeader header="Настройки" paragraph="Основные параметры личного кабинета" />

    <div :class="style.settingsSections">
      <section :class="style.settingsSection">
        <h2>Личные данные</h2>
        <div :class="style.settingsPanel">
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
              <h3>Фото профиля</h3>
              <p>JPG, PNG или WebP · до 2 МБ</p>
            </div>
            <input
              ref="avatarInput"
              type="file"
              accept="image/jpeg,image/png,image/webp"
              hidden
              @change="uploadAvatar"
            />
            <div :class="style.profileActions">
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
                >Удалить</Button
              >
            </div>
          </div>
          <dl :class="style.accountFields">
            <div>
              <dt>ФИО</dt>
              <dd>{{ auth.user?.full_name || 'Не указано' }}</dd>
            </div>
            <div>
              <dt>Электронная почта</dt>
              <dd>{{ auth.user?.email || 'Не указана' }}</dd>
            </div>
            <div v-if="auth.user?.student">
              <dt>Группа</dt>
              <dd>{{ auth.user.student.group_name || 'Не указана' }}</dd>
            </div>
            <div v-else-if="auth.user?.teacher">
              <dt>Кафедра</dt>
              <dd>{{ auth.user.teacher.department_name || 'Не указана' }}</dd>
            </div>
          </dl>
        </div>
      </section>

      <section :class="style.settingsSection">
        <h2>Вход и безопасность</h2>
        <div :class="style.settingsPanel">
          <div :class="style.settingRow">
            <div>
              <h3>Пароль</h3>
              <p>Измените пароль для входа в личный кабинет</p>
            </div>
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
          <div :class="style.settingRow">
            <div>
              <h3>Текущий сеанс</h3>
              <p>Вы вошли как {{ auth.user?.email }}</p>
            </div>
            <Button variant="ghost" :loading="loggingOut" :disabled="loggingOut" @click="signOut">
              <PhSignOut :size="18" /> Выйти
            </Button>
          </div>
        </div>
      </section>

      <section :class="style.settingsSection">
        <h2>Оформление</h2>
        <div :class="style.settingsPanel">
          <div :class="[style.settingRow, style.appearanceRow]">
            <div>
              <h3>Тема интерфейса</h3>
              <p>Выберите удобное оформление</p>
            </div>
            <fieldset :class="style.themeControl">
              <legend :class="style.visuallyHidden">Тема интерфейса</legend>
              <label v-for="item in themeOptions" :key="item.value" :class="style.themeOption">
                <input
                  v-model="$utilityStore.themePreference"
                  type="radio"
                  name="theme"
                  :value="item.value"
                />
                <span>
                  <PhSun v-if="item.value === 'light'" :size="18" />
                  <PhMoon v-else-if="item.value === 'dark'" :size="18" />
                  <PhDesktop v-else :size="18" />
                  {{ item.label }}
                </span>
              </label>
            </fieldset>
          </div>
        </div>
        <p :class="style.sectionNote">Системная тема подстраивается под настройки устройства.</p>
      </section>
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
</template>

<script lang="ts">
import { apiNotifications } from '@/api/client'
import { useFilesStore } from '@/stores/files'
import { useAuthStore } from '@/stores/auth'
import { defineComponent } from 'vue'
import { PhEye, PhEyeSlash, PhSun, PhMoon, PhDesktop, PhSignOut } from '@phosphor-icons/vue'

import style from './SettingsPage.module.scss'
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
    Modal,
    Button,
    IconButton,
    TextInput,
    PhEye,
    PhEyeSlash,
    PhSun,
    PhMoon,
    PhDesktop,
    PhSignOut,
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
        { value: 'light', label: 'Светлая' },
        { value: 'dark', label: 'Тёмная' },
        { value: 'system', label: 'Системная' },
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
        await this.auth.updateMe({ avatar: uploaded.id })
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
        await this.auth.updateMe({ avatar: null })
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
