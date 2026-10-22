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
            <p>Студентка</p>
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
          <Button variant="secondary">Изменить пароль</Button>
        </div>
      </div>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './SettingsPage.module.scss'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'

export default defineComponent({
  name: 'SettingsPage',

  components: {
    SectionHeader,
    StyleTab,
    Button,
  },

  data() {
    return {
      style,
      photoPreview: null as string | null,
      theme: 'dark',
      themeOptions: [
        { value: 'light', label: 'Светлая' },
        { value: 'dark', label: 'Тёмная' },
        { value: 'system', label: 'Системный' },
      ],
    }
  },
})
</script>
