<template>
  <AdminNavigation
    @collapse-change="collapsed = $event"
    @mobile-menu-change="mobileMenuOpen = $event"
  />
  <main :class="[home.mainContent, collapsed && home.collapsed]" :inert="mobileMenuOpen">
    <div :class="layout.content">
      <RouterView v-if="$route.meta.ownerOnly" />
      <template v-else>
        <p v-if="$adminStore.success" :class="style.success" role="status">
          {{ $adminStore.success }}
        </p>
        <div v-if="$adminStore.error" :class="style.error">
          <p>{{ $adminStore.error }}</p>
          <Button
            v-if="!$adminStore.data"
            variant="ghost"
            @click="$adminStore.initialize(String($route.name))"
          >
            Повторить
          </Button>
        </div>
        <!-- <div v-if="$adminStore.loading" :class="style.empty" role="status">Загрузка данных…</div> -->
         <PageLoader v-if="$adminStore.loading" />
        <RouterView v-else-if="$adminStore.data" />
      </template>
    </div>
  </main>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { RouterView } from 'vue-router'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminNavigation from './components/AdminNavigation.vue'
import home from '@/layout/HomeLayout/HomeLayout.module.scss'
import layout from './AdminLayout.module.scss'
import style from './Admin.module.scss'
import PageLoader from '@/components/PageLoader/PageLoader.vue'
export default defineComponent({
  components: { RouterView, Button, AdminNavigation, PageLoader },
  computed: {
    auth() {
      return useAuthStore()
    },
  },
  data() {
    return { home, layout, style, collapsed: false, mobileMenuOpen: false }
  },
  async mounted() {
    await this.$adminStore.initialize(String(this.$route.name))
  },
  watch: {
    async 'auth.user'() {
      if (!this.auth.user) {
        await this.$router.replace({
          name: 'authorization',
          query: { redirect: this.$route.fullPath },
        })
      }
    },
    async '$route.name'() {
      await this.$adminStore.initialize(String(this.$route.name))
    },
  },
})
</script>
