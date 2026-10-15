import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'
import { createMemoryHistory, createRouter } from 'vue-router'

import StudentCard from '@/components/navigation/StudentCard/StudentCard.vue'
import NavBar from '@/layout/NavBar/NavBar.vue'

describe('Navigation components', () => {
  it('renders student data and initials without an image', () => {
    const wrapper = mount(StudentCard, {
      props: {
        name: 'Анна Иванова',
        role: 'Студентка',
      },
    })

    expect(wrapper.text()).toContain('Анна Иванова')
    expect(wrapper.text()).toContain('Студентка')
    expect(wrapper.text()).toContain('АИ')
    expect(wrapper.find('img').exists()).toBe(false)
  })

  it('opens the selected page and updates the active tab', async () => {
    const EmptyPage = { template: '<main />' }
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/dashboard', name: 'dashboard', component: EmptyPage },
        { path: '/schedule', name: 'schedule', component: EmptyPage },
        { path: '/subjects', name: 'subjects', component: EmptyPage },
        { path: '/grades', name: 'grades', component: EmptyPage },
        { path: '/messages', name: 'messages', component: EmptyPage },
        { path: '/resources', name: 'resources', component: EmptyPage },
        { path: '/settings', name: 'settings', component: EmptyPage },
      ],
    })

    await router.push('/dashboard')
    await router.isReady()

    const wrapper = mount(NavBar, {
      global: {
        plugins: [router],
      },
    })
    const scheduleTab = wrapper
      .findAll('[role="tab"]')
      .find((tab) => tab.text().includes('Расписание'))
    const dashboardTab = wrapper
      .findAll('[role="tab"]')
      .find((tab) => tab.text().includes('Главная'))

    expect(scheduleTab).toBeDefined()
    expect(dashboardTab).toBeDefined()
    expect(dashboardTab!.attributes('aria-selected')).toBe('true')

    await scheduleTab!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('schedule')
    expect(scheduleTab!.attributes('aria-selected')).toBe('true')

    await dashboardTab!.trigger('click')
    await flushPromises()

    expect(router.currentRoute.value.name).toBe('dashboard')
    expect(dashboardTab!.attributes('aria-selected')).toBe('true')
  })
})
