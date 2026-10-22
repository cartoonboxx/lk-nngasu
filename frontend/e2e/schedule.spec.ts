import { test, expect } from '@playwright/test'

test('weekly schedule highlights only the current day and changes weeks', async ({ page }) => {
  await page.clock.install({ time: new Date('2026-10-06T10:00:00') })
  await page.goto('/schedule')
  const table = page.getByRole('table')
  await expect(table.locator('thead th')).toHaveCount(7)
  await expect(table.locator('tbody tr')).toHaveCount(6)
  await expect(table.locator('tbody td')).toHaveCount(36)
  const today = table.getByRole('columnheader', { name: 'Вт, 6 окт', exact: true })
  await expect(today).toHaveClass(/today/)
  await expect(table.getByText('Сегодня', { exact: true })).toHaveCount(0)
  await expect(table.locator('tbody [class*=today]')).toHaveCount(0)
  await expect(table.getByRole('heading', { name: 'Математический анализ' }).first()).toBeVisible()
  const controls = page.locator('main button')
  await controls.nth(1).click()
  await expect(table.getByRole('columnheader', { name: 'Пн, 12 окт', exact: true })).toBeVisible()
  await expect(table.locator('thead [class*=today]')).toHaveCount(0)
  await page.getByRole('button', { name: 'Сегодня', exact: true }).click()
  await expect(today).toHaveClass(/today/)
})

test('mobile schedule scrolls within the table and keeps times visible', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await page.goto('/schedule')
  const table = page.getByRole('table')
  const scroller = table.locator('..')
  const time = table.getByRole('rowheader', { name: /1 пара/ })
  const before = await time.boundingBox()
  await scroller.evaluate((element) => {
    element.scrollLeft = 500
  })
  const after = await time.boundingBox()
  expect(after?.x).toBe(before?.x)
  expect(await scroller.evaluate((element) => element.scrollLeft)).toBeGreaterThan(0)
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(375)
})
