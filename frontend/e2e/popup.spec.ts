import { test, expect } from '@playwright/test'

test('select supports keyboard selection and returns focus', async ({ page }) => {
  await page.goto('/subjects')
  const trigger = page.getByRole('button', { name: '1 семестр', exact: true })
  await trigger.focus()
  await page.keyboard.press('ArrowDown')
  const secondSemester = page.getByRole('button', { name: '2 семестр', exact: true })
  await expect(secondSemester).toBeVisible()
  await page.keyboard.press('ArrowDown')
  await expect(secondSemester).toBeFocused()
  await page.keyboard.press('Enter')
  await expect(secondSemester).toBeFocused()
  await page.keyboard.press('ArrowDown')
  await page.keyboard.press('Escape')
  await expect(secondSemester).toBeFocused()
})

test('course topics support keyboard navigation', async ({ page }) => {
  await page.goto('/subjects/logic')
  const topics = page.locator('aside button')
  await topics.first().focus()
  await page.keyboard.press('ArrowDown')
  await expect(topics.nth(1)).toBeFocused()
  await expect(
    page.getByRole('heading', { name: 'Синтаксис и значение программ языка Пролог' }),
  ).toBeVisible()
})
