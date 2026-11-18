<template>
  <NavBar
    @collapse-change="navigationCollapsed = $event"
    @mobile-menu-change="mobileMenuOpen = $event"
  />
  <main
    :inert="mobileMenuOpen"
    :class="[
      style.mainContent,
      navigationCollapsed ? style.collapsed : null,
      $route.name === 'messages' ? style.fullBleed : null,
    ]"
  >
    <p v-if="profileError" role="alert">
      {{ profileError }} <button @click="loadProfile">Повторить загрузку профиля</button>
    </p>
    <RouterView />
  </main>
</template>

<script lang="ts">
import { useAuthStore } from '@/stores/auth'
import { defineComponent } from 'vue'
import NavBar from '@/layout/NavBar/NavBar.vue'
import style from './HomeLayout.module.scss'

export default defineComponent({
  name: 'HomeLayout',
  components: { NavBar },
  created() {
    void this.loadProfile()
  },
  watch: {
    'auth.user.id'(id: number | undefined) {
      if (!id) void this.$router.replace({ name: 'authorization' })
    },
  },
  methods: {
    async loadProfile() {
      this.profileError = ''
      try {
        await this.auth.getMe()
      } catch (error) {
        this.profileError = error instanceof Error ? error.message : 'Не удалось загрузить профиль.'
      }
    },
  },
  data() {
    return {
      style,
      auth: useAuthStore(),
      profileError: '',
      navigationCollapsed: false,
      mobileMenuOpen: false,
    }
  },
})
</script>
