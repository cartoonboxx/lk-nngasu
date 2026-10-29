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
            :active="item.value === theme"
            @select="theme = item.value"
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

            <label :class="style.passwordField">
              <span>Текущий пароль</span>
              <input
                v-model="currentPassword"
                type="password"
                autocomplete="current-password"
                required
              />
            </label>

            <label :class="style.passwordField">
              <span>Новый пароль</span>
              <input v-model="newPassword" type="password" autocomplete="new-password" required />
            </label>

            <label :class="style.passwordField">
              <span>Повторите новый пароль</span>
              <input
                v-model="confirmPassword"
                type="password"
                autocomplete="new-password"
                required
              />
            </label>

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
export default defineComponent({
  name: 'SettingsPage',

  components: {
    SectionHeader,
    StyleTab,
    Modal,
    Button,
  },

  data() {
    return {
      style,
      isPasswordModalOpen: false,
      currentPassword: '',
      newPassword: '',
      confirmPassword: '',
      passwordError: '',
      photoPreview: null as string | null,
      theme: 'dark',
      themeOptions: [
        { value: 'light', label: 'Светлая' },
        { value: 'dark', label: 'Тёмная' },
        { value: 'system', label: 'Системная' },
      ],
    }
  },
  methods: {
    resetPasswordForm() {
      this.currentPassword = ''
      this.newPassword = ''
      this.confirmPassword = ''
      this.passwordError = ''
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
