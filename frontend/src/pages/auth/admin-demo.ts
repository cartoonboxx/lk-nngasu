import { reactive } from 'vue'

// TODO: remove this dev-only adapter when server authentication is connected.
// This marker opens local mock screens only; it never authorizes API requests.
export const adminDemoSessionKey = 'lk-nngasu-admin-demo-session'
let active = false
if (import.meta.env.DEV) {
  try {
    active = sessionStorage.getItem(adminDemoSessionKey) === 'active'
  } catch {
    /* Without storage the session lasts until reload. */
  }
}
export const adminDemoSession = reactive({ active })

export function signInAdminDemo(login: string, password: string): boolean {
  if (!import.meta.env.DEV || login !== 'root@mail' || password !== 'admin') return false
  adminDemoSession.active = true
  try {
    sessionStorage.setItem(adminDemoSessionKey, 'active')
  } catch {
    /* Local preview still works until reload. */
  }
  return true
}
