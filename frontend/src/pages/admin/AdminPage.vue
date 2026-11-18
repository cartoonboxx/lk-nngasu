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
        <div v-if="$adminStore.error" :class="style.error" role="alert">
          {{ $adminStore.error }}
          <Button
            v-if="!$adminStore.data"
            variant="ghost"
            @click="$adminStore.initialize(String($route.name))"
          >
            Повторить
          </Button>
        </div>
        <div v-if="$adminStore.loading" :class="style.empty" role="status">Загрузка данных…</div>
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
export default defineComponent({
  components: { RouterView, Button, AdminNavigation },
  data() {
    return { home, layout, style, auth: useAuthStore(), collapsed: false, mobileMenuOpen: false }
  },
  mounted() {
    void this.$adminStore.initialize(String(this.$route.name))
  },
  watch: {
    'auth.user'() {
      if (!this.auth.user) {
        void this.$router.replace({
          name: 'authorization',
          query: { redirect: this.$route.fullPath },
        })
      }
    },
    '$route.name'() {
      void this.$adminStore.initialize(String(this.$route.name))
    },
  },
})
</script>
