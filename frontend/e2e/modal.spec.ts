import { test, expect } from '@playwright/test'

test('modal manages focus, background scrolling and closing', async ({ page }) => {
  await page.goto('/ui-kit')
  const trigger = page.getByRole('button', { name: 'Открыть модальное окно', exact: true })
  const dialog = page.locator('dialog')
  await expect(dialog).toBeHidden()
  await trigger.focus()
  await page.keyboard.press('Enter')
  await expect(dialog).toBeVisible()
  await expect(dialog.locator(':scope > div')).toBeFocused()
  await expect(page.locator('body')).toHaveCSS('overflow', 'hidden')

  await page.keyboard.press('Tab')
  await expect(dialog.locator('button').first()).toBeFocused()
  await expect(dialog.locator('button').first()).toHaveCSS('outline-style', 'solid')

  await page.keyboard.press('Tab')
  await expect(dialog.locator('textarea')).toBeFocused()
  await page.keyboard.press('Shift+Tab')
  await expect(dialog.locator('button').first()).toBeFocused()
  await dialog.locator('textarea').fill('Нужна помощь с расписанием')
  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
  await expect(trigger).toBeFocused()
  await expect(page.locator('body')).not.toHaveCSS('overflow', 'hidden')

  await trigger.click()
  await dialog.locator('button').first().click()
  await expect(dialog).toBeHidden()
  await trigger.click()
  await page.mouse.click(5, 5)
  await expect(dialog).toBeHidden()
})

test('modal fits a narrow screen and keeps content clicks inside', async ({ page }) => {
  await page.setViewportSize({ width: 320, height: 480 })
  await page.goto('/ui-kit')
  await page.getByRole('button', { name: 'Открыть модальное окно', exact: true }).click()
  const dialog = page.locator('dialog')
  const heading = dialog.getByRole('heading', { name: 'Обратная связь' })
  await heading.click()
  await expect(dialog).toBeVisible()
  const box = await dialog.locator('textarea').boundingBox()
  expect(box).not.toBeNull()
  expect(box!.x).toBeGreaterThanOrEqual(0)
  expect(box!.x + box!.width).toBeLessThanOrEqual(320)
  expect(box!.y + box!.height).toBeLessThanOrEqual(480)
})

test('desktop modal supports different sizes and becomes a mobile sheet', async ({ page }) => {
  await page.setViewportSize({ width: 1440, height: 900 })
  await page.goto('/ui-kit')
  const trigger = page.getByRole('button', { name: 'Открыть модальное окно', exact: true })
  const dialog = page.getByRole('dialog')
  const panel = dialog.locator(':scope > div')
  await page.getByRole('button', { name: 'Компактная', exact: true }).click()
  await trigger.click()
  await expect(panel).toHaveCSS('width', '360px')
  await expect(panel.locator('[class*=handle]')).toBeHidden()
  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
  await page.getByRole('button', { name: 'Широкая', exact: true }).click()
  await trigger.click()
  await expect(panel).toHaveCSS('width', '800px')
  await expect(panel).toHaveCSS('height', '560px')
  await expect
    .poll(async () => {
      const box = await panel.boundingBox()
      return Math.round(box!.x + box!.width / 2)
    })
    .toBe(720)
  await page.setViewportSize({ width: 768, height: 600 })
  await expect(panel).toHaveCSS('width', '720px')
  await expect(panel).toHaveCSS('height', '552px')
  await page.setViewportSize({ width: 375, height: 700 })
  await expect(panel).toHaveCSS('width', '375px')
  await expect(panel.locator('[class*=handle]')).toBeVisible()
  await expect
    .poll(async () => {
      const box = await panel.boundingBox()
      return Math.round(box!.y + box!.height)
    })
    .toBe(700)
  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
  await expect(page.locator('body')).not.toHaveCSS('overflow', 'hidden')
})

test('mobile sheet snaps between breakpoints, dismisses by drag and restores focus', async ({
  page,
}) => {
  await page.setViewportSize({ width: 375, height: 812 })
  await page.goto('/ui-kit')
  const trigger = page.getByRole('button', { name: 'Открыть модальное окно', exact: true })
  const dialog = page.getByRole('dialog')
  const panel = dialog.locator(':scope > div')
  const handle = dialog.locator('[class*=handle]')
  await trigger.click()
  await expect
    .poll(async () => {
      const box = (await panel.boundingBox())!
      return Math.round(box.y + box.height)
    })
    .toBe(812)
  const full = (await panel.boundingBox())!
  let grip = (await handle.boundingBox())!
  await page.mouse.move(grip.x + grip.width / 2, grip.y + grip.height / 2)
  await page.mouse.down()
  await page.mouse.move(grip.x + grip.width / 2, grip.y + grip.height / 2 + full.height / 2, {
    steps: 12,
  })
  await page.mouse.up()
  await expect(dialog).toBeVisible()
  await expect
    .poll(async () => Math.round((await panel.boundingBox())!.y))
    .toBe(Math.round(full.y + full.height / 2))
  grip = (await handle.boundingBox())!
  await page.mouse.move(grip.x + grip.width / 2, grip.y + grip.height / 2)
  await page.mouse.down()
  await page.mouse.move(grip.x + grip.width / 2, grip.y + grip.height / 2 - full.height / 2, {
    steps: 12,
  })
  await page.mouse.up()
  await expect.poll(async () => Math.round((await panel.boundingBox())!.y)).toBe(Math.round(full.y))
  grip = (await handle.boundingBox())!
  await page.mouse.move(grip.x + grip.width / 2, grip.y + grip.height / 2)
  await page.mouse.down()
  await page.mouse.move(grip.x + grip.width / 2, 800, { steps: 16 })
  await page.mouse.up()
  await expect(dialog).toBeHidden()
  await expect(trigger).toBeFocused()
  await expect(page.locator('body')).not.toHaveCSS('overflow', 'hidden')
  await trigger.click()
  await expect.poll(async () => Math.round((await panel.boundingBox())!.y)).toBe(Math.round(full.y))
})

test('mobile assignment scrolls inside the sheet and reopens after dismissal', async ({ page }) => {
  await page.setViewportSize({ width: 375, height: 667 })
  await page.goto('/subjects/logic')
  const trigger = page.getByRole('button', { name: 'Открыть задание', exact: true })
  await trigger.click()
  const dialog = page.getByRole('dialog')
  const content = dialog.locator('[class*=content]').first()
  const panel = dialog.locator(':scope > div')
  await expect
    .poll(async () => {
      const box = (await panel.boundingBox())!
      return Math.round(box.y + box.height)
    })
    .toBe(667)
  expect(await content.evaluate((element) => element.scrollHeight > element.clientHeight)).toBe(
    true,
  )
  const comment = dialog.getByLabel('Комментарий к ответу')
  await comment.fill('Мобильный черновик')
  await dialog.getByRole('button', { name: 'Отправить преподавателю' }).scrollIntoViewIfNeeded()
  expect(await content.evaluate((element) => element.scrollTop)).toBeGreaterThan(0)
  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
  await trigger.click()
  await expect(comment).toBeVisible()
  await comment.fill('Новый ответ')
  await expect(comment).toHaveValue('Новый ответ')
})

for (const sample of [
  { route: '/ui-kit', trigger: 'button:has-text("Открыть модальное окно")', name: 'base modal' },
  { route: '/schedule', trigger: 'main button[class*=trigger]:visible', name: 'calendar' },
  { route: '/dashboard', trigger: 'nav button:has-text("Ещё")', name: 'navigation menu' },
]) {
  test(`${sample.name} enters from below without scrolling its overlay`, async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 812 })
    await page.goto(sample.route)
    const trigger = page.locator(sample.trigger)
    await trigger.scrollIntoViewIfNeeded()
    const framesPromise = page.evaluate(
      () =>
        new Promise<{ y: number; height: number; scroll: number }[]>((resolve) => {
          const frames: { y: number; height: number; scroll: number }[] = []
          let started = 0
          const record = (time: number) => {
            const dialog = document.querySelector<HTMLDialogElement>('dialog[open]')
            if (dialog) {
              if (!started) started = time
              const panel = dialog.firstElementChild!.getBoundingClientRect()
              frames.push({ y: panel.y, height: panel.height, scroll: dialog.scrollTop })
              if (time - started >= 320) {
                resolve(frames)
                return
              }
            }
            requestAnimationFrame(record)
          }
          requestAnimationFrame(record)
        }),
    )
    await trigger.click()
    const frames = await framesPromise
    expect(frames.length).toBeGreaterThan(3)
    expect(frames[0]!.y).toBeGreaterThanOrEqual(811)
    const restingTop = 812 - frames.at(-1)!.height
    expect(frames.at(-1)!.y).toBeCloseTo(restingTop, 0)
    expect(frames.some((frame) => frame.y > restingTop + 2 && frame.y < 810)).toBe(true)
    for (const [index, frame] of frames.entries()) {
      expect(frame.scroll).toBe(0)
      expect(frame.y).toBeGreaterThanOrEqual(restingTop - 1)
      expect(frame.y).toBeLessThanOrEqual(frames[Math.max(0, index - 1)]!.y + 1)
    }
  })
}

test('touching sheet content scrolls first and dismisses from the top', async ({
  browser,
  browserName,
}) => {
  // eslint-disable-next-line playwright/no-skipped-test -- CDP touch input is Chromium-only.
  test.skip(browserName !== 'chromium', 'Native touch input uses the Chromium protocol')
  const context = await browser.newContext({
    viewport: { width: 375, height: 812 },
    hasTouch: true,
    isMobile: true,
  })
  const page = await context.newPage()
  const session = await context.newCDPSession(page)
  const swipe = async (x: number, y: number, distance: number) => {
    await session.send('Input.dispatchTouchEvent', {
      type: 'touchStart',
      touchPoints: [{ x, y }],
    })
    for (let step = 1; step <= 20; step++) {
      await session.send('Input.dispatchTouchEvent', {
        type: 'touchMove',
        touchPoints: [{ x, y: y + (distance * step) / 20 }],
      })
    }
    await session.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  }
  await page.goto('/subjects/logic')
  await page.getByRole('button', { name: 'Открыть задание', exact: true }).click()
  const dialog = page.getByRole('dialog')
  const panel = dialog.locator(':scope > div')
  const content = dialog.locator('[class*=content]').first()
  await expect.poll(async () => Math.round((await panel.boundingBox())!.y)).toBeGreaterThan(0)
  await expect
    .poll(async () => {
      const box = (await panel.boundingBox())!
      return Math.round(box.y + box.height)
    })
    .toBe(812)
  const top = (await panel.boundingBox())!.y

  // An upward swipe must scroll the content without moving the sheet.
  await swipe(30, 650, -250)
  await expect.poll(() => content.evaluate((element) => element.scrollTop)).toBeGreaterThan(0)
  await expect.poll(async () => (await panel.boundingBox())!.y).toBeCloseTo(top, 0)
  await expect(dialog).toBeVisible()

  // Starting below the top keeps a downward swipe in the scroll container.
  await content.evaluate((element) => {
    element.scrollTop = 400
  })
  await swipe(30, 250, 100)
  await expect(dialog).toBeVisible()
  await expect.poll(async () => (await panel.boundingBox())!.y).toBeCloseTo(top, 0)

  await content.evaluate((element) => {
    element.scrollTop = 0
  })
  await swipe(30, top + 90, 570)
  await expect(dialog).toBeHidden()

  // Short, non-scrollable content also supports the same gesture.
  await page.goto('/ui-kit')
  await page.getByRole('button', { name: 'Открыть модальное окно', exact: true }).click()
  await expect
    .poll(async () => {
      const box = (await panel.boundingBox())!
      return Math.round(box.y + box.height)
    })
    .toBe(812)
  const box = (await panel.boundingBox())!
  await swipe(30, box.y + 50, box.height * 0.7)
  await expect(dialog).toBeHidden()
  await context.close()
})
