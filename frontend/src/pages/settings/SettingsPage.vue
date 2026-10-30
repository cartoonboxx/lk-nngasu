<template>
  <section :class="style.settingsPage">
    <SectionHeader header="Настройки" paragraph="Основные параметры личного кабинета" />

    <section :class="style.settingsCard">
      <div :class="style.settingsSection">
        <h2>Фото профиля</h2>

        <div :class="style.profileRow">
          <div :class="style.profileAvatar">
            <img
              v-if="photoPreview"
              :src="photoPreview"
              alt="Фото профиля"
              :class="style.profileAvatarImage"
            />
            <span v-else>АИ</span>
          </div>

          <div :class="style.profileInfo">
            <p>Анна Иванова</p>
            <p>Студент</p>
          </div>

          <Button variant="secondary">Изменить фото</Button>
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
                  :aria-label="showCurrentPassword ? 'Скрыть пароль' : 'Показать пароль'"
                  :aria-pressed="showCurrentPassword"
                  @click="showCurrentPassword = !showCurrentPassword"
                >
                  <PhEyeSlash v-if="showCurrentPassword" aria-hidden="true" />
                  <PhEye v-else aria-hidden="true" />
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
                  :aria-label="showNewPassword ? 'Скрыть пароль' : 'Показать пароль'"
                  :aria-pressed="showNewPassword"
                  @click="showNewPassword = !showNewPassword"
                >
                  <PhEyeSlash v-if="showNewPassword" aria-hidden="true" />
                  <PhEye v-else aria-hidden="true" />
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
                  :aria-label="showConfirmPassword ? 'Скрыть пароль' : 'Показать пароль'"
                  :aria-pressed="showConfirmPassword"
                  @click="showConfirmPassword = !showConfirmPassword"
                >
                  <PhEyeSlash v-if="showConfirmPassword" aria-hidden="true" />
                  <PhEye v-else aria-hidden="true" />
                </IconButton>
              </template>
            </TextInput>

            <p v-if="passwordError" :class="style.passwordError" role="alert">
              {{ passwordError }}
            </p>

            <div :class="style.passwordActions">
              <Button variant="secondary" type="button" @click="close"> Отмена </Button>

              <Button type="submit"> Сохранить </Button>
            </div>
          </form>
        </template>
      </Modal>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './SettingsPage.module.scss'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import { PhEye, PhEyeSlash } from '@phosphor-icons/vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
export default defineComponent({
  name: 'SettingsPage',

  components: {
    SectionHeader,
    StyleTab,
    Modal,
    Button,
    TextInput,
    PhEye,
    PhEyeSlash,
  },

  data() {
    return {
      style,
      isPasswordModalOpen: false,
      showCurrentPassword: false,
      showNewPassword: false,
      showConfirmPassword: false,
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
      passwordError: '',
      photoPreview: null as string | null,
      themeOptions: [
        { value: 'system', label: 'Системная' },
        { value: 'light', label: 'Светлая' },
        { value: 'dark', label: 'Тёмная' },
        { value: 'system', label: 'Системная' },
      ],
    }
  },
  methods: {
    openPasswordModal() {
      this.resetPasswordForm()
      this.isPasswordModalOpen = true
    },

    resetPasswordForm() {
      this.currentPassword = ''
      this.newPassword = ''
      this.confirmPassword = ''
      this.passwordError = ''
      this.showCurrentPassword = false
      this.showNewPassword = false
      this.showConfirmPassword = false
    },

    submitPasswordChange() {
      this.passwordError = ''

      if (this.newPassword !== this.confirmPassword) {
        this.passwordError = 'Новые пароли не совпадают.'
        return
      }

      if (this.newPassword === this.currentPassword) {
        this.passwordError = 'Новый пароль должен отличаться от текущего.'
        return
      }

      this.passwordError = 'Сервис смены пароля пока не подключён.'
    },
  },
})
</script>
