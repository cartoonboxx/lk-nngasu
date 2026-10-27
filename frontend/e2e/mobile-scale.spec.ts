import { test, expect } from '@playwright/test'

for (const width of [320, 440, 767]) {
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
        await expect(page.locator('body')).toHaveCSS('font-size', '17px')
        expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width)
      })
    }
    await page.goto('/schedule')
    const schedule = page.locator('section[class*=mobileSchedule]')
    await expect(schedule).toBeVisible()
    const lesson = schedule.getByRole('listitem').first()
    await expect(lesson.getByRole('heading')).toHaveCSS('font-size', '17px')
    await expect(lesson.locator('[class*=lessonTime]')).toHaveCSS('font-size', '14px')
    await expect(lesson.locator('[class*=lessonMeta]')).toHaveCSS('font-size', '14px')
    expect((await lesson.boundingBox())!.height).toBeGreaterThanOrEqual(128)
    expect(await page.evaluate(() => document.documentElement.scrollWidth)).toBe(width)
  })
}
