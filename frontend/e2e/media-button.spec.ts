import { test, expect } from '@playwright/test'

test('media download works from the keyboard and readonly removal is independent', async ({
  page,
}) => {
  await page.goto('/ui-kit')
  const section = page
    .locator('section')
    .filter({ has: page.getByRole('heading', { name: 'Медиафайлы', exact: true }) })
  const downloadButton = section.getByRole('button').filter({ hasText: 'Лекция 1' })
  await downloadButton.focus()
  const downloadPromise = page.waitForEvent('download')
  await page.keyboard.press('Enter')
  const download = await downloadPromise
  expect(download.suggestedFilename()).toBe('material.md')
  await expect(
    section.getByRole('button', { name: 'Недоступный материал Скачивание недоступно' }),
  ).toBeDisabled()
  await expect(section.locator('button button')).toHaveCount(0)
  await section
    .locator('div[class*=mediaButton]')
    .filter({ hasText: 'main.pl' })
    .getByRole('button')
    .click()
  await expect(section.getByText('main.pl', { exact: true })).toHaveCount(0)
  await section.getByRole('button', { name: 'Вернуть файл main.pl' }).click()
  await expect(section.getByText('main.pl', { exact: true })).toBeVisible()
})

test('media rows fit a narrow screen', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 640 })
  await page.goto('/ui-kit')
  const section = page
    .locator('section')
    .filter({ has: page.getByRole('heading', { name: 'Медиафайлы', exact: true }) })
  await section.scrollIntoViewIfNeeded()
  const controls = [
    section.getByRole('button').filter({ hasText: 'Лекция 1' }),
    section.locator('div[class*=mediaButton]').filter({ hasText: 'main.pl' }).getByRole('button'),
  ]
  for (const control of controls) {
    const box = await control.boundingBox()
    expect(box).not.toBeNull()
    expect(box!.x).toBeGreaterThanOrEqual(0)
    expect(box!.x + box!.width).toBeLessThanOrEqual(320)
  }
})
