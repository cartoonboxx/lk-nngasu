import { createRouter, createWebHistory } from 'vue-router'
import UiKitPage from '@/pages/UiKitPage/UiKitPage.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/ui-kit',
      name: 'ui-kit',
      component: UiKitPage,
    },
  ],
})

export default router
