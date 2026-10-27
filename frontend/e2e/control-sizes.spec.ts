import { test, expect, type Locator } from '@playwright/test'

async function bounds(locator: Locator) {
  const box = await locator.boundingBox()
  expect(box).not.toBeNull()
  return box!
}

for (const { width, alignedControls } of [
  { width: 1280, alignedControls: 6 },
  { width: 375, alignedControls: 3 },
]) {
  test(`controls share heights at ${width}px`, async ({ page }) => {
    await page.setViewportSize({ width, height: 900 })
    await page.goto('/ui-kit')
    const scale = page.locator('section').filter({
      has: page.getByRole('heading', { name: 'Общая шкала размеров', exact: true }),
    })
    for (const [name, height] of [
      ['Small', 36],
      ['Medium', 46],
      ['Large', 54],
    ] as const) {
      const row = scale
        .locator('article')
        .filter({ has: page.getByRole('heading', { name, exact: true }) })
      const buttons = row.getByRole('button')
      const controls = [
        buttons.nth(0),
        buttons.nth(1),
        buttons.nth(2),
        row.getByRole('textbox'),
        row.getByRole('button', { name: 'Выбор', exact: true }).locator('..'),
        row.locator('label'),
      ]
      const boxes = await Promise.all(controls.map(bounds))
      for (const box of boxes) expect(box.height).toBe(height)
      for (const index of [1, 2]) expect(boxes[index]!.width).toBe(height)
      for (const box of boxes.slice(0, alignedControls)) expect(box.y).toBe(boxes[0]!.y)
    }

    await page.evaluate(() =>
      document.documentElement.style.setProperty('--control-height-md', '50px'),
    )
    const medium = scale
      .locator('article')
      .filter({ has: page.getByRole('heading', { name: 'Medium', exact: true }) })
    expect((await bounds(medium.getByRole('button').nth(0))).height).toBe(50)
    expect((await bounds(medium.getByRole('button').nth(1))).height).toBe(50)
    expect((await bounds(medium.getByRole('button').nth(2))).height).toBe(50)
    expect((await bounds(medium.getByRole('textbox'))).height).toBe(50)
    expect(
      (await bounds(medium.getByRole('button', { name: 'Выбор', exact: true }).locator('..')))
        .height,
    ).toBe(50)
    expect((await bounds(medium.locator('label'))).height).toBe(50)
  })
}

test('auth fields and submit button use the large height', async ({ page }) => {
  for (const path of ['/authorization', '/registration']) {
    await page.goto(path)
    const form = page.locator('form')
    const fields = form.locator('input:not([type=checkbox])')
    for (const field of await fields.all()) expect((await bounds(field)).height).toBe(54)
    expect((await bounds(form.locator('button[type=submit]'))).height).toBe(54)
    expect(
      (await bounds(form.locator('input[type=checkbox]').locator('..'))).height,
    ).toBeGreaterThanOrEqual(36)
  }
})

test('default tabs and popup trigger use the medium height', async ({ page }) => {
  await page.goto('/ui-kit')
  const tabs = page
    .locator('section')
    .filter({ has: page.getByRole('heading', { name: 'Табы', exact: true }) })
  for (const name of ['Базовый таб', 'Только текст', 'Изображение без actions', 'Активный таб']) {
    const tab = tabs.getByRole('button', { name, exact: true }).locator('..')
    expect((await bounds(tab)).height).toBe(46)
  }
  const navTab = tabs
    .locator('article')
    .filter({ has: page.getByRole('heading', { name: 'NavTab — навигация', exact: true }) })
    .getByRole('button')
    .first()
    .locator('..')
  expect((await bounds(navTab)).height).toBe(46)
  const popup = page
    .locator('section')
    .filter({ has: page.getByRole('heading', { name: 'FloatingPopup', exact: true }) })
  expect(
    (await bounds(popup.getByRole('button', { name: 'Учебная информация', exact: true }))).height,
  ).toBe(46)
})
