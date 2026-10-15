import { createRouter, createWebHistory } from 'vue-router'
import UiKitPage from '@/pages/UiKitPage/UiKitPage.vue'
import AuthorizationPage from '@/pages/auth/AuthorizationPage/AuthorizationPage.vue'
import RegistrationComponent from '@/pages/auth/RegistrationPage/RegistrationComponent.vue'
import GradesPage from '@/pages/grades/GradesPage.vue'
import DashboardPage from '@/pages/dashboard/DashboardPage.vue'
import HomeLayout from '@/layout/HomeLayout/HomeLayout.vue'
import MessagesPage from '@/pages/messages/MessagesPage.vue'
import ResourcesPage from '@/pages/resources/ResourcesPage.vue'
import SchedulePage from '@/pages/schedule/SchedulePage.vue'
import SettingsPage from '@/pages/settings/SettingsPage.vue'
import SubjectsPage from '@/pages/subjects/SubjectsPage.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: HomeLayout,
      children: [
        {
          path: '',
          redirect: { name: 'dashboard' },
        },
        {
          path: 'dashboard',
          name: 'dashboard',
          component: DashboardPage,
        },
        {
          path: 'schedule',
          name: 'schedule',
          component: SchedulePage,
        },
        {
          path: 'subjects',
          name: 'subjects',
          component: SubjectsPage,
        },
        {
          path: 'grades',
          name: 'grades',
          component: GradesPage,
        },
        {
          path: 'messages',
          name: 'messages',
          component: MessagesPage,
        },
        {
          path: 'resources',
          name: 'resources',
          component: ResourcesPage,
        },
        {
          path: 'settings',
          name: 'settings',
          component: SettingsPage,
        },
      ],
    },
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
