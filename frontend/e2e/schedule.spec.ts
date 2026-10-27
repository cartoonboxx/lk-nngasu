import { test, expect } from '@playwright/test'

test('weekly schedule highlights only the current day and changes weeks', async ({ page }) => {
  await page.clock.install({ time: new Date('2026-10-06T10:00:00') })
  await page.goto('/schedule')
  const table = page.getByRole('table')
  await expect(table.locator('thead th')).toHaveCount(7)
  await expect(table.locator('tbody tr')).toHaveCount(7)
  await expect(table.locator('tbody td')).toHaveCount(42)
  const today = table.getByRole('columnheader', { name: 'Вт, 6 окт', exact: true })
  await expect(today).toHaveClass(/today/)
  await expect(table.getByText('Сегодня', { exact: true })).toHaveCount(0)
  await expect(table.locator('tbody [class*=today]')).toHaveCount(0)
  await expect(table.getByRole('heading', { name: 'Математический анализ' }).first()).toBeVisible()
  await page.locator('[class*=dailyPart] > button').nth(1).click()
  await expect(table.getByRole('columnheader', { name: 'Пн, 12 окт', exact: true })).toBeVisible()
  await expect(table.locator('thead [class*=today]')).toHaveCount(0)
  await page.locator('main button[class*=trigger]:visible').click()
  await page.getByRole('dialog').getByRole('button', { name: 'Сегодня', exact: true }).click()
  await page.getByRole('dialog').getByRole('button', { name: 'Применить', exact: true }).click()
  await expect(today).toHaveClass(/today/)
})

test('mobile schedule shows the full week with day headings and a type legend', async ({
  page,
}) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await page.clock.install({ time: new Date('2026-10-23T10:00:00') })
  await page.goto('/schedule')
  const schedule = page.locator('section[class*=mobileSchedule]')
  await expect(page.getByRole('table')).toBeHidden()
  await expect(schedule.getByRole('heading', { level: 2 })).toHaveText([
    'Понедельник, 19 октября',
    'Вторник, 20 октября',
    'Среда, 21 октября',
    'Четверг, 22 октября',
    'Пятница, 23 октября',
    'Суббота, 24 октября',
  ])
  const monday = schedule
    .locator('section')
    .filter({ has: page.getByRole('heading', { name: 'Понедельник, 19 октября', exact: true }) })
  await expect(monday.getByRole('listitem')).toHaveCount(4)
  await expect(monday.locator('time')).toHaveAttribute('datetime', '2026-10-19')
  await expect(schedule.getByRole('heading', { name: /Воскресенье/ })).toHaveCount(0)
  await expect(schedule.getByText(/^(Скоро|Идёт|Прошло)$/)).toHaveCount(0)
  await page.locator('[class*=dailyPart] > button').nth(1).click()
  await expect(schedule.getByRole('heading', { level: 2 }).first()).toHaveText(
    'Понедельник, 26 октября',
  )
  await expect(schedule.getByRole('heading', { level: 2 }).last()).toHaveText('Суббота, 31 октября')
  await page.setViewportSize({ width: 1440, height: 900 })
  await expect(
    page.getByRole('table').getByRole('columnheader', { name: 'Пн, 26 окт', exact: true }),
  ).toBeVisible()
  await page.setViewportSize({ width: 375, height: 812 })
  await expect(schedule.getByRole('heading', { level: 2 }).first()).toHaveText(
    'Понедельник, 26 октября',
  )
  await page.locator('[class*=dailyPart] > button').first().click()
  await expect(monday).toBeVisible()
  await page.locator('[class*=dailyPart] > button').first().click()
  await expect(schedule.getByRole('heading', { level: 2 }).first()).toHaveText(
    'Понедельник, 12 октября',
  )
  await page.locator('main button[class*=trigger]:visible').click()
  await page.getByRole('dialog').getByRole('button', { name: 'Сегодня', exact: true }).click()
  await page.getByRole('dialog').getByRole('button', { name: 'Применить', exact: true }).click()
  await expect(monday).toBeVisible()
  const legend = page.locator('ul[class*=legend]')
  await legend.scrollIntoViewIfNeeded()
  await expect(legend.getByRole('listitem')).toHaveText(['Лекция', 'Практика', 'Лабораторная'])
  expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(375)
})

test('weekly mobile list fits narrow and wide mobile screens', async ({ page }) => {
  await page.clock.install({ time: new Date('2026-10-23T10:00:00') })
  await page.goto('/schedule')
  const schedule = page.locator('section[class*=mobileSchedule]')
  for (const width of [320, 485, 767]) {
    await page.setViewportSize({ width, height: 900 })
    await expect(schedule).toBeVisible()
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width)
    const saturday = schedule.getByRole('heading', { name: 'Суббота, 24 октября' })
    await saturday.scrollIntoViewIfNeeded()
    await expect(saturday).toBeInViewport()
    const legend = page.locator('ul[class*=legend]')
    await legend.scrollIntoViewIfNeeded()
    await expect(legend).toBeInViewport()
  }
})
