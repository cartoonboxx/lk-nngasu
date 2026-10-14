import { createRouter, createWebHistory } from 'vue-router'
import UiKitPage from '@/pages/UiKitPage/UiKitPage.vue'
import AuthorizationPage from '@/pages/auth/AuthorizationPage/AuthorizationPage.vue'
import RegistrationComponent from '@/pages/auth/RegistrationPage/RegistrationComponent.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/authorization',
      name: 'authorization',
      component: AuthorizationPage,
    },
    {
      path: '/registration',
      name: 'registration',
      component: RegistrationComponent,
    },
    {
      path: '/ui-kit',
      name: 'ui-kit',
      component: UiKitPage,
    },
  ],
})

export default router
