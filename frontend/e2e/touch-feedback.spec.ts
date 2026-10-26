import { test, expect, type Locator } from '@playwright/test'

function section(page: import('@playwright/test').Page, name: string) {
  return page.locator('section').filter({
    has: page.getByRole('heading', { name, exact: true }),
  })
}

async function center(control: Locator) {
  await control.scrollIntoViewIfNeeded()
  const box = (await control.boundingBox())!
  return { x: box.x + box.width / 2, y: box.y + box.height / 2 }
}

test('desktop hover and keyboard focus remain available', async ({ page }) => {
  await page.goto('/ui-kit')
  const button = section(page, 'Общая шкала размеров').getByRole('button').first()
  const background = await button.evaluate((element) => getComputedStyle(element).backgroundColor)
  await button.hover()
  await expect(button).not.toHaveCSS('background-color', background)
  await button.focus()
  await expect(button).toHaveCSS('outline-style', 'solid')
  const tooltip = section(page, 'Tooltip').getByRole('button', { name: 'top', exact: true })
  await tooltip.hover()
  await expect(page.getByText('Подсказка: top', { exact: true })).toBeVisible()
})

test('touch does not leave hover or tooltips and fields show focus', async ({
  browser,
  browserName,
}) => {
  // eslint-disable-next-line playwright/no-skipped-test -- Firefox has no touch context support.
  test.skip(browserName === 'firefox', 'Touch emulation requires Chromium or WebKit')
  const context = await browser.newContext({
    viewport: { width: 375, height: 812 },
    hasTouch: true,
  })
  const page = await context.newPage()
  await page.goto('/ui-kit')
  const scale = section(page, 'Общая шкала размеров')
  const button = scale.getByRole('button').first()
  const background = await button.evaluate((element) => getComputedStyle(element).backgroundColor)
  await button.tap()
  await expect(button).toHaveCSS('background-color', background)
  const input = scale.getByRole('textbox').first()
  await input.tap()
  await expect(input).toBeFocused()
  await expect(input).toHaveCSS('font-size', '16px')
  await expect(input).toHaveCSS('border-top-color', 'rgb(89, 181, 255)')
  const tooltip = section(page, 'Tooltip').getByRole('button', { name: 'top', exact: true })
  await tooltip.tap()
  await expect(page.getByText('Подсказка: top', { exact: true })).toHaveCount(0)
  await context.close()
})

test('touch press releases on lift, cancellation and scrolling and respects reduced motion', async ({
  browser,
  browserName,
}) => {
  // eslint-disable-next-line playwright/no-skipped-test -- Native touch input requires CDP.
  test.skip(browserName !== 'chromium', 'Native touch sequences use the Chromium protocol')
  const context = await browser.newContext({
    viewport: { width: 375, height: 812 },
    hasTouch: true,
  })
  const page = await context.newPage()
  const session = await context.newCDPSession(page)
  await page.goto('/ui-kit')
  const scale = section(page, 'Общая шкала размеров')
  for (const [index, pressedScale] of [
    [0, '0.97'],
    [1, '0.92'],
    [2, '0.92'],
    [3, '0.98'],
  ] as const) {
    const button = scale.getByRole('button').nth(index)
    const point = await center(button)
    await session.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point] })
    await expect(button).toHaveCSS('scale', pressedScale)
    await session.send('Input.dispatchTouchEvent', { type: 'touchCancel', touchPoints: [] })
    await expect(button).toHaveCSS('scale', 'none')
  }

  const button = scale.getByRole('button').first()
  let point = await center(button)
  await session.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point] })
  await expect(button).toHaveCSS('scale', '0.97')
  await session.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  await expect(button).toHaveCSS('scale', 'none')
  point = await center(button)
  const scrollBefore = await page.evaluate(() => window.scrollY)
  await session.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point] })
  for (let step = 1; step <= 8; step++) {
    await session.send('Input.dispatchTouchEvent', {
      type: 'touchMove',
      touchPoints: [{ x: point.x, y: point.y - step * 20 }],
    })
  }
  await session.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  await expect(button).toHaveCSS('scale', 'none')
  await expect.poll(() => page.evaluate(() => window.scrollY)).toBeGreaterThan(scrollBefore)

  await page.emulateMedia({ reducedMotion: 'reduce' })
  point = await center(button)
  await session.send('Input.dispatchTouchEvent', { type: 'touchStart', touchPoints: [point] })
  await expect(button).toHaveCSS('scale', 'none')
  await expect(button).toHaveCSS('transition-duration', '0s')
  await session.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  await context.close()
})
