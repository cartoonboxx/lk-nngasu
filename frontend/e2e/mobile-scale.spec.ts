import { test, expect } from '@playwright/test'

for (const { width, body, small, caption, rowHeight } of [
  { width: 320, body: 16, small: 14, caption: 12, rowHeight: 104 },
  { width: 374, body: 16, small: 14, caption: 12, rowHeight: 104 },
  { width: 375, body: 16, small: 15, caption: 13, rowHeight: 116 },
  { width: 390, body: 16, small: 15, caption: 13, rowHeight: 116 },
  { width: 413, body: 16, small: 15, caption: 13, rowHeight: 116 },
  { width: 414, body: 17, small: 16, caption: 14, rowHeight: 128 },
  { width: 440, body: 17, small: 16, caption: 14, rowHeight: 128 },
  { width: 767, body: 17, small: 16, caption: 14, rowHeight: 128 },
]) {
  test(`mobile pages remain readable and fit at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 956 })
    for (const path of [
      '/dashboard',
      '/subjects',
      '/subjects/1',
      '/resources',
      '/settings',
      '/news',
      '/authorization',
      '/registration',
      '/messages?chat=group-it-202',
    ]) {
      await test.step(path, async () => {
        await page.goto(path)
        await expect(page.locator('body')).toHaveCSS('font-size', `${body}px`)
        expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width)
      })
    }
    await page.goto('/messages')
    const chat = page.getByRole('list', { name: 'Список диалогов' }).getByRole('button').first()
    await expect(chat.locator('[class*=chatName]')).toHaveCSS('font-size', `${small}px`)
    await page.goto('/schedule')
    const schedule = page.locator('section[class*=mobileSchedule]')
    await expect(schedule).toBeVisible()
    const lesson = schedule.getByRole('listitem').first()
    await expect(lesson.getByRole('heading')).toHaveCSS('font-size', `${body}px`)
    await expect(lesson.locator('[class*=lessonTime]')).toHaveCSS('font-size', `${caption}px`)
    await expect(lesson.locator('[class*=lessonMeta]')).toHaveCSS('font-size', `${caption}px`)
    expect((await lesson.boundingBox())!.height).toBeGreaterThanOrEqual(rowHeight)
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width)
  })
}
