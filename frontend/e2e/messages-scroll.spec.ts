import { test, expect, type Locator, type Page } from '@playwright/test'
import { createChats } from '../src/pages/messages/messages.data'

async function distanceToEnd(history: Locator) {
  return history.evaluate(
    (element) => element.scrollHeight - element.clientHeight - element.scrollTop,
  )
}

async function receiveMessage(page: Page, id: string) {
  await page.evaluate((messageId) => {
    const key = `lk-nngasu-messages-v2:message:group-it-202:${messageId}`
    const value = JSON.stringify({
      chatId: 'group-it-202',
      message: {
        id: messageId,
        text: 'Новое входящее сообщение',
        sender: 'Мария',
        own: false,
        sentAt: new Date().toISOString(),
      },
    })
    localStorage.setItem(key, value)
    window.dispatchEvent(new StorageEvent('storage', { key, newValue: value }))
  }, id)
}

test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true })

test.beforeEach(async ({ page }) => {
  const chats = createChats()
  chats[0]!.messages = Array.from({ length: 80 }, (_, index) => ({
    id: `history-${index}`,
    text: `Сообщение ${index + 1}. Длинная история переписки для проверки прокрутки.`,
    sender: 'Мария',
    own: false,
    sentAt: new Date(2026, 9, 28, 10, index).toISOString(),
  }))
  for (let index = 0; index < 25; index++) {
    chats.push({ ...chats[1]!, id: `contact-${index}`, name: `Собеседник ${index}`, messages: [] })
  }
  await page.addInitScript((data) => {
    localStorage.setItem('lk-nngasu-messages-v2:base', JSON.stringify(data))
  }, chats)
})

test('incoming messages preserve the reading position and follow only at the end', async ({
  page,
}) => {
  await page.goto('/messages?chat=group-it-202')
  const history = page.getByRole('log')
  await expect.poll(() => distanceToEnd(history)).toBeLessThanOrEqual(1)
  await history.evaluate((element) => {
    element.scrollTop = 120
    element.dispatchEvent(new Event('scroll'))
  })
  await receiveMessage(page, 'incoming-reading')
  await expect(history.getByRole('article')).toHaveCount(81)
  await expect.poll(() => history.evaluate((element) => element.scrollTop)).toBe(120)

  await history.evaluate((element) => {
    element.scrollTop = element.scrollHeight
    element.dispatchEvent(new Event('scroll'))
  })
  await receiveMessage(page, 'incoming-following')
  await expect(history.getByRole('article')).toHaveCount(82)
  await expect.poll(() => distanceToEnd(history)).toBeLessThanOrEqual(1)
})

test('resizing and growing the composer keep the latest messages visible', async ({ page }) => {
  await page.goto('/messages?chat=group-it-202')
  const history = page.getByRole('log')
  await expect.poll(() => distanceToEnd(history)).toBeLessThanOrEqual(1)
  const field = page.getByRole('textbox', { name: 'Сообщение', exact: true })
  await field.fill('Многострочный черновик\n'.repeat(20))
  await expect(field).toHaveCSS('height', '144px')
  await expect.poll(() => distanceToEnd(history)).toBeLessThanOrEqual(1)
  await page.setViewportSize({ width: 390, height: 500 })
  await expect.poll(() => distanceToEnd(history)).toBeLessThanOrEqual(1)

  await history.evaluate((element) => {
    element.scrollTop = 120
    element.dispatchEvent(new Event('scroll'))
  })
  await field.fill('Короткий черновик')
  await expect.poll(() => history.evaluate((element) => element.scrollTop)).toBe(120)
  await page.getByRole('button', { name: 'Отправить сообщение' }).tap()
  await expect(history.getByRole('article').last()).toContainText('Короткий черновик')
  await expect.poll(() => distanceToEnd(history)).toBeLessThanOrEqual(1)
})

test('chat list scrolls independently and other pages retain document scrolling', async ({
  page,
}) => {
  await page.goto('/messages')
  const list = page.getByRole('list', { name: 'Список диалогов' })
  const header = page.getByRole('heading', { name: /Чаты/ })
  const headerTop = (await header.boundingBox())!.y
  await list.evaluate((element) => {
    element.scrollTop = element.scrollHeight
  })
  await expect.poll(() => list.evaluate((element) => element.scrollTop)).toBeGreaterThan(0)
  expect((await header.boundingBox())!.y).toBe(headerTop)
  expect(await page.evaluate(() => window.scrollY)).toBe(0)
  // The page must stay locked even when a gesture reaches a list boundary.
  await expect(page.locator('body')).toHaveCSS('position', 'fixed')
  await page.locator('nav').last().getByRole('link', { name: 'Главная', exact: true }).tap()
  await expect(page).toHaveURL(/\/dashboard$/)
  await expect(page.locator('body')).not.toHaveCSS('position', 'fixed')
  await page.evaluate(() => window.scrollTo(0, document.documentElement.scrollHeight))
  await expect.poll(() => page.evaluate(() => window.scrollY)).toBeGreaterThan(0)
})

test('native touch scrolls the history without dragging the page or its header', async ({
  page,
  browserName,
}) => {
  // eslint-disable-next-line playwright/no-skipped-test -- Native touch input uses the Chromium protocol.
  test.skip(browserName !== 'chromium', 'Native touch sequences require CDP')
  await page.goto('/messages?chat=group-it-202')
  const history = page.getByRole('log')
  const heading = page.getByRole('heading', { name: 'Группа ИТ-202', exact: true })
  await expect.poll(() => distanceToEnd(history)).toBeLessThanOrEqual(1)
  const initialTop = await history.evaluate((element) => element.scrollTop)
  const headerTop = (await heading.boundingBox())!.y
  const bounds = (await history.boundingBox())!
  const session = await page.context().newCDPSession(page)
  const x = bounds.x + bounds.width / 2
  const y = bounds.y + 60
  await session.send('Input.dispatchTouchEvent', {
    type: 'touchStart',
    touchPoints: [{ x, y }],
  })
  for (let step = 1; step <= 16; step++) {
    await session.send('Input.dispatchTouchEvent', {
      type: 'touchMove',
      touchPoints: [{ x, y: y + step * 10 }],
    })
  }
  await session.send('Input.dispatchTouchEvent', { type: 'touchEnd', touchPoints: [] })
  await expect.poll(() => history.evaluate((element) => element.scrollTop)).toBeLessThan(initialTop)
  expect((await heading.boundingBox())!.y).toBe(headerTop)
  expect(await page.evaluate(() => window.scrollY)).toBe(0)
})
