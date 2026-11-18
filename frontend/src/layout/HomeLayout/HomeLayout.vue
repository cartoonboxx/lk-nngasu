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
    <p v-if="profileError">
      <button @click="loadProfile">Повторить загрузку профиля</button>
    </p>
    <RouterView />
  </main>
</template>

<script lang="ts">
import { apiNotifications } from '@/api/client'
import { useAuthStore } from '@/stores/auth'
import { defineComponent } from 'vue'
import NavBar from '@/layout/NavBar/NavBar.vue'
import style from './HomeLayout.module.scss'

export default defineComponent({
  name: 'HomeLayout',
  components: { NavBar },
  async created() {
    await this.loadProfile()
  },
  watch: {
    async 'auth.user.id'(id: number | undefined) {
      if (!id) await this.$router.replace({ name: 'authorization' })
    },
  },
  methods: {
    async loadProfile() {
      this.profileError = ''
      try {
        await this.auth.getMe()
      } catch (error) {
        apiNotifications.onError?.(error)
        this.profileError = error instanceof Error ? error.message : 'Не удалось загрузить профиль.'
      }
    },
  },
  computed: {
    auth() {
      return useAuthStore()
    },
  },
  data() {
    return {
      style,
      profileError: '',
      navigationCollapsed: false,
      mobileMenuOpen: false,
    }
  },
})
</script>
