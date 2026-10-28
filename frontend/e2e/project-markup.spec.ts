import { test, expect } from '@playwright/test'

test('pages omit explicit attributes and standard router links navigate', async ({ page }) => {
  const errors: string[] = []
  page.on('pageerror', (error) => errors.push(error.message))
  for (const route of [
    '/dashboard',
    '/messages',
    '/messages?chat=curator',
    '/messages?compose=1',
    '/schedule',
    '/subjects',
    '/subjects/logic',
    '/grades',
    '/resources',
    '/settings',
    '/news',
    '/ui-kit',
    '/authorization',
    '/registration',
  ]) {
    await page.goto(route)
    await expect(page.locator('#app')).not.toBeEmpty()
    const attributes = await page.evaluate(() =>
      Array.from(document.querySelectorAll('*')).flatMap((element) =>
        element
          .getAttributeNames()
          .filter(
            (name) =>
              name.startsWith('aria-') &&
              !(
                name === 'aria-current' &&
                element instanceof HTMLAnchorElement &&
                element.classList.contains('router-link-exact-active')
              ),
          ),
      ),
    )
    expect(attributes, route).toEqual([])
  }
  await page.getByRole('link', { name: 'Войти', exact: true }).click()
  await expect(page).toHaveURL(/\/authorization$/)
  await page.getByRole('link', { name: 'Зарегистрироваться', exact: true }).click()
  await expect(page).toHaveURL(/\/registration$/)
  expect(errors).toEqual([])
})
