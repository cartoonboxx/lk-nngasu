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
import CoursePage from '@/pages/subjects/course/CoursePage.vue'
import NewsPage from '@/pages/news/NewsPage.vue'
import NewsDetailsPage from '@/pages/news/page/NewsDetailsPage.vue'
import NotFoundPage from '@/pages/notFound/NotFoundPage.vue'
import AdminPage from '@/pages/admin/AdminPage.vue'
import { adminDemoSession, adminDemoSessionKey } from '@/pages/auth/admin-demo'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/admin',
      component: AdminPage,
      meta: { adminOnly: true },
      children: [
        { path: '', redirect: { name: 'admin-requests' } },
        {
          path: 'requests',
          name: 'admin-requests',
          component: () => import('@/pages/admin/RequestsPage.vue'),
        },
        {
          path: 'users',
          name: 'admin-users',
          component: () => import('@/pages/admin/UsersPage.vue'),
        },
        {
          path: 'structure',
          name: 'admin-structure',
          component: () => import('@/pages/admin/StructurePage.vue'),
        },
        {
          path: 'courses',
          name: 'admin-courses',
          component: () => import('@/pages/admin/CoursesPage.vue'),
        },
        {
          path: 'schedule',
          name: 'admin-schedule',
          component: () => import('@/pages/admin/AdminSchedulePage.vue'),
        },
        {
          path: 'news',
          name: 'admin-news',
          component: () => import('@/pages/admin/AdminNewsPage.vue'),
        },
      ],
    },
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
          path: 'subjects/:id',
          name: 'course',
          component: CoursePage,
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
        {
          path: 'news',
          name: 'news',
          component: NewsPage,
        },
        {
          path: '/news/:idNews',
          name: 'news-details',
          component: NewsDetailsPage,
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
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: NotFoundPage,
    },
  ],
})

router.beforeEach((to, from) => {
  if (to.name === 'authorization' || (from.meta.adminOnly && !to.meta.adminOnly)) {
    adminDemoSession.active = false
    try {
      sessionStorage.removeItem(adminDemoSessionKey)
    } catch {
      /* In-memory access is already cleared. */
    }
  }
  if (to.meta.adminOnly && !adminDemoSession.active) {
    return { name: 'authorization', query: { reason: 'admin-required' } }
  }
})

export default router
