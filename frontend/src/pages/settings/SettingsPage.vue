
<template>
  <!-- первая секция -->
  <section :class="style.settingsPage">
    <h1>Настройки</h1>
    <p :class="style.settingsPageDescription">
      Основные параметры личного кабинета
    </p>
  </section>
<!-- основа -->
  <section :class="style.settingsCard">
    <!-- Фото -->
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
          <p :class="style.profileInfoName">Анна Иванова</p>
          <p :class="style.profileInfoRole">Студентка</p>
        </div>

        <button class="_base_qfg6r_1 _md_qfg6r_28 _secondary_5ij5q_12"
          type="button">Изменить фото</button>
      </div>
    </div>

    <div :class="style.settingsDivider"></div>

<!-- Тема -->
    <div :class="style.settingsSectionRow">
      <div >
        <h2 :class="style.settingsSectionTitle">Тема интерфейса</h2>
        <p :class="style.settingsSectionDescription">
          Выберите удобное оформление
        </p>
      </div>

      <div :class="style.tabColumn" role="listbox" aria-label="Выбор темы">
        <StyleTab 
          v-for="item in themeOptions"
          :key="item.value"
          :text="item.label"
          :active="item.value === theme"
          @select="theme = item.value"
        />
      </div>
    </div>

    <div :class="style.settingsDivider"></div>

<!-- Пароль -->
 <div :class="style.settingsSectionRow">

    <div>
      <h2 :class="style.settingsSectionTitle">Пароль</h2>
      <p :class="style.settingsSectionDescription">Изменить пароль для входа</p>
    </div>

    <div>
      <button
        class="_base_qfg6r_1 _md_qfg6r_28 _secondary_5ij5q_12"
        type="button">Изменить пароль
      </button>
    </div>
 </div>

  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './SettingsPage.module.scss'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'

export default defineComponent({
  name: 'SettingsPage',

  components: {
    StyleTab,
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
