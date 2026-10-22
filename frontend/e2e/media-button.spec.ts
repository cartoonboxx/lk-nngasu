import { test, expect } from '@playwright/test'

test('media download works from the keyboard and readonly removal is independent', async ({
  page,
}) => {
  await page.goto('/ui-kit')
  const section = page.getByRole('region', { name: 'Медиафайлы' })
  const downloadButton = section.getByRole('button', { name: 'Скачать лекцию' })
  await downloadButton.focus()
  const downloadPromise = page.waitForEvent('download')
  await page.keyboard.press('Enter')
  const download = await downloadPromise
  expect(download.suggestedFilename()).toBe('material.md')
  await expect(
    section.getByRole('button', { name: 'Недоступный материал Скачивание недоступно' }),
  ).toBeDisabled()
  await expect(section.locator('button button')).toHaveCount(0)
  await section.getByRole('button', { name: 'Удалить файл main.pl' }).click()
  await expect(section.getByText('main.pl', { exact: true })).toHaveCount(0)
  await section.getByRole('button', { name: 'Вернуть файл main.pl' }).click()
  await expect(section.getByText('main.pl', { exact: true })).toBeVisible()
})

test('media rows fit a narrow screen', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 640 })
  await page.goto('/ui-kit')
  const section = page.getByRole('region', { name: 'Медиафайлы' })
  await section.scrollIntoViewIfNeeded()
  for (const name of ['Скачать лекцию', 'Удалить файл main.pl']) {
    const box = await section.getByRole('button', { name }).boundingBox()
    expect(box).not.toBeNull()
    expect(box!.x).toBeGreaterThanOrEqual(0)
    expect(box!.x + box!.width).toBeLessThanOrEqual(320)
  }
})
