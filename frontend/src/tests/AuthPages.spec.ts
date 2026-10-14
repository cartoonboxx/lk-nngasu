import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createMemoryHistory, createRouter } from 'vue-router'
import App from '../App.vue'
import AuthorizationPage from '../pages/auth/AuthorizationPage/AuthorizationPage.vue'
import RegistrationComponent from '../pages/auth/RegistrationPage/RegistrationComponent.vue'

async function mountPage(path: string) {
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/authorization', component: AuthorizationPage },
      { path: '/registration', component: RegistrationComponent },
    ],
  })
  await router.push(path)
  await router.isReady()
  return mount(App, { global: { plugins: [router] } })
}

describe('Authentication forms', () => {
  it('preserves a password when toggling its visibility', async () => {
    const wrapper = await mountPage('/authorization')
    const password = wrapper.get('input[name="password"]')
    await password.setValue('saved-password')
    await wrapper.get('button[aria-label="Показать пароль"]').trigger('click')
    expect(password.attributes('type')).toBe('text')
    expect((password.element as HTMLInputElement).value).toBe('saved-password')
    await wrapper.get('button[aria-label="Скрыть пароль"]').trigger('click')
    expect(password.attributes('type')).toBe('password')
    expect((password.element as HTMLInputElement).value).toBe('saved-password')
    wrapper.unmount()
  })

  it('exposes autocomplete and required attributes on the native fields', async () => {
    const login = await mountPage('/authorization')
    expect(login.get('input[name="username"]').attributes('autocomplete')).toBe('username')
    expect(login.get('input[name="password"]').attributes('autocomplete')).toBe('current-password')
    expect((login.get('input[name="password"]').element as HTMLInputElement).required).toBe(true)
    login.unmount()

    const registration = await mountPage('/registration')
    expect(registration.get('input[name="email"]').attributes('autocomplete')).toBe('email')
    expect(registration.get('input[name="password"]').attributes('autocomplete')).toBe(
      'new-password',
    )
    expect(registration.get('input[name="password-confirmation"]').attributes('autocomplete')).toBe(
      'new-password',
    )
    registration.unmount()
  })

  it('keeps registration unavailable until consent and matching passwords', async () => {
    const wrapper = await mountPage('/registration')
    const submit = wrapper.get('button[type="submit"]')
    const confirmation = wrapper.get('input[name="password-confirmation"]')
    expect((submit.element as HTMLButtonElement).disabled).toBe(true)
    await wrapper.get('input[name="password"]').setValue('my-password')
    await confirmation.setValue('another-password')
    await wrapper.get('input[name="terms"]').setValue(true)
    expect((submit.element as HTMLButtonElement).disabled).toBe(true)
    expect(confirmation.attributes('aria-invalid')).toBe('true')
    expect(confirmation.attributes('aria-describedby')).toBe('password-error')
    expect(wrapper.get('#password-error').text()).toContain('Пароли не совпадают')
    await confirmation.setValue('my-password')
    expect((submit.element as HTMLButtonElement).disabled).toBe(false)
    expect(wrapper.find('#password-error').exists()).toBe(false)
    expect(confirmation.attributes('aria-invalid')).toBeUndefined()
    wrapper.unmount()
  })
})
