import { test, expect } from '@playwright/test'

test('searches chats and marks opened conversations as read', async ({ page }) => {
  await page.goto('/messages')
  const list = page.getByRole('list', { name: 'Список диалогов' })
  await expect(list.getByRole('button')).toHaveCount(4)
  await page.getByPlaceholder('Поиск по чатам').fill('куратор')
  await expect(list.getByRole('button')).toHaveCount(1)
  await page.getByPlaceholder('Поиск по чатам').fill('неизвестный')
  await expect(page.getByText('Чаты не найдены')).toBeVisible()
  await page.getByPlaceholder('Поиск по чатам').fill('')
  await page.getByRole('button', { name: /Непрочитанные/ }).click()
  await expect(list.getByRole('button')).toHaveCount(1)
  await list.getByRole('button', { name: /Группа ИТ-202/ }).click()
  await expect(page.getByRole('log')).toContainText('Кто будет на практике сегодня?')
  await expect(page.getByLabel('Непрочитанных сообщений: 2')).toHaveCount(0)
  await page.reload()
  await expect(page.getByLabel('Непрочитанных сообщений: 2')).toHaveCount(0)
})

test('sends multiline messages, preserves drafts per chat and saves sent messages', async ({
  page,
}) => {
  await page.goto('/messages?chat=group-it-202')
  const field = page.getByRole('textbox', { name: 'Сообщение', exact: true })
  const send = page.getByRole('button', { name: 'Отправить сообщение' })
  await expect(send).toBeDisabled()
  await field.fill('   ')
  await expect(send).toBeDisabled()
  await field.fill('Черновик группы')
  await page.getByRole('button', { name: /Куратор группы/ }).click()
  await expect(field).toHaveValue('')
  await field.fill('Черновик куратору')
  await page.getByRole('button', { name: /Группа ИТ-202/ }).click()
  await expect(field).toHaveValue('Черновик группы')
  await field.fill('Первая строка')
  await field.press('Shift+Enter')
  await field.pressSequentially('Вторая строка')
  await field.press('Enter')
  await expect(page.getByRole('log').getByRole('article').last()).toContainText(
    'Первая строка\nВторая строка',
  )
  await expect(field).toHaveValue('')
  await expect(field).toBeFocused()
  await page.reload()
  await expect(page.getByRole('log').getByRole('article').last()).toContainText(
    'Первая строка\nВторая строка',
  )
})

test('mobile Enter inserts a newline and desktop Enter sends after resizing', async ({ page }) => {
  await page.setViewportSize({ width: 767, height: 900 })
  await page.goto('/messages?chat=curator')
  const field = page.getByRole('textbox', { name: 'Сообщение', exact: true })
  const messages = page.getByRole('log').getByRole('article')
  const initialCount = await messages.count()
  await field.fill('Мобильная строка')
  await field.press('Enter')
  await expect(field).toHaveValue('Мобильная строка\n')
  await expect(messages).toHaveCount(initialCount)
  await field.pressSequentially('Вторая строка')
  await page.getByRole('button', { name: 'Отправить сообщение' }).click()
  await expect(messages.last()).toContainText('Мобильная строка\nВторая строка')
  await expect(field).toHaveValue('')

  await page.setViewportSize({ width: 768, height: 900 })
  await field.fill('Отправка с компьютера')
  await field.press('Enter')
  await expect(messages.last()).toContainText('Отправка с компьютера')
  await expect(field).toHaveValue('')

  await page.setViewportSize({ width: 390, height: 844 })
  await field.fill('Снова мобильная версия')
  await field.press('Enter')
  await expect(field).toHaveValue('Снова мобильная версия\n')
  await expect(messages).toHaveCount(initialCount + 2)
})

test('textarea grows without scrolling until its maximum height and shrinks after clearing', async ({
  page,
}) => {
  await page.goto('/messages?chat=curator')
  const field = page.getByRole('textbox', { name: 'Сообщение', exact: true })
  const initialHeight = (await field.boundingBox())!.height
  await expect(field).toHaveCSS('overflow-y', 'hidden')
  await field.fill('Первая строка\nВторая строка\nТретья строка')
  await expect.poll(async () => (await field.boundingBox())!.height).toBeGreaterThan(initialHeight)
  await expect.poll(async () => (await field.boundingBox())!.height).toBeLessThan(144)
  await expect(field).toHaveCSS('overflow-y', 'hidden')
  expect(await field.evaluate((element) => element.scrollHeight <= element.clientHeight)).toBe(true)

  await field.fill('Длинное сообщение\n'.repeat(30))
  await expect(field).toHaveCSS('height', '144px')
  await expect(field).toHaveCSS('overflow-y', 'auto')
  expect(await field.evaluate((element) => element.scrollHeight > element.clientHeight)).toBe(true)
  await field.evaluate((element) => {
    element.scrollTop = element.scrollHeight
  })
  expect(await field.evaluate((element) => element.scrollTop)).toBeGreaterThan(0)

  await field.fill('')
  await expect.poll(async () => (await field.boundingBox())!.height).toBe(initialHeight)
  await expect(field).toHaveCSS('overflow-y', 'hidden')
  expect(await field.evaluate((element) => element.scrollTop)).toBe(0)
})

test('opens compose links and starts a new conversation without duplicates', async ({ page }) => {
  await page.goto('/messages?compose=1')
  const dialog = page.getByRole('dialog')
  await expect(dialog).toBeVisible()
  await dialog.getByPlaceholder('Имя или фамилия').fill('Мария')
  await dialog.getByRole('button', { name: /Мария Кузнецова/ }).click()
  await expect(page).toHaveURL(/chat=maria/)
  await expect(dialog).toBeHidden()
  await expect(page.getByText('Начните разговор', { exact: true }).last()).toBeVisible()
  await page.getByRole('textbox', { name: 'Сообщение', exact: true }).fill('Привет, Мария!')
  await page.getByRole('button', { name: 'Отправить сообщение' }).click()
  await expect(page.getByRole('log')).toContainText('Привет, Мария!')
  await page.getByRole('button', { name: 'Новый диалог', exact: true }).click()
  await dialog.getByPlaceholder('Имя или фамилия').fill('Мария')
  await dialog.getByRole('button', { name: /Мария Кузнецова/ }).click()
  await expect(page.getByRole('list', { name: 'Список диалогов' }).getByRole('button')).toHaveCount(
    5,
  )
})

test('mobile conversation fits the screen and returns to the chat list', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/messages')
  const list = page.getByRole('list', { name: 'Список диалогов' })
  await list.getByRole('button', { name: /Группа ИТ-202/ }).click()
  await expect(list).toBeHidden()
  const field = page.getByRole('textbox', { name: 'Сообщение', exact: true })
  await expect(field).toBeVisible()
  await field.fill('Сообщение с телефона')
  await page.getByRole('button', { name: 'Отправить сообщение' }).click()
  await expect(page.getByRole('log')).toContainText('Сообщение с телефона')
  await field.fill('Длинное сообщение '.repeat(150))
  await expect.poll(async () => (await field.boundingBox())!.height).toBeLessThanOrEqual(144)
  await page.getByRole('button', { name: 'Отправить сообщение' }).click()
  const log = page.getByRole('log')
  await expect.poll(() => log.evaluate((el) => el.scrollHeight > el.clientHeight)).toBe(true)
  const bounds = await field.boundingBox()
  expect(bounds!.y + bounds!.height).toBeLessThan(772)
  await page.setViewportSize({ width: 320, height: 568 })
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth)).toBe(true)
  await page.getByRole('button', { name: 'Назад к чатам' }).click()
  await expect(list).toBeVisible()
  await expect(page.getByRole('log')).toHaveCount(0)
})

test('keeps drafts after leaving messages and reloading the page', async ({ page }) => {
  await page.goto('/messages?chat=curator')
  await page
    .getByRole('textbox', { name: 'Сообщение', exact: true })
    .fill('Черновик после перехода')
  await page.goto('/subjects')
  await page.goto('/messages?chat=curator')
  await expect(page.getByRole('textbox', { name: 'Сообщение', exact: true })).toHaveValue(
    'Черновик после перехода',
  )
  await page.reload()
  await expect(page.getByRole('textbox', { name: 'Сообщение', exact: true })).toHaveValue(
    'Черновик после перехода',
  )
})

test('synchronizes concurrent messages and read receipts across tabs', async ({
  page,
  context,
}) => {
  await page.goto('/messages?chat=curator')
  const second = await context.newPage()
  await second.goto('/messages?chat=curator')
  await page
    .getByRole('textbox', { name: 'Сообщение', exact: true })
    .fill('Сообщение из первой вкладки')
  await second
    .getByRole('textbox', { name: 'Сообщение', exact: true })
    .fill('Сообщение из второй вкладки')
  await Promise.all([
    page.getByRole('button', { name: 'Отправить сообщение' }).click(),
    second.getByRole('button', { name: 'Отправить сообщение' }).click(),
  ])
  await expect(page.getByRole('log')).toContainText('Сообщение из второй вкладки')
  await expect(second.getByRole('log')).toContainText('Сообщение из первой вкладки')
  await page.reload()
  await expect(page.getByRole('log')).toContainText('Сообщение из первой вкладки')
  await expect(page.getByRole('log')).toContainText('Сообщение из второй вкладки')
  await second.getByRole('button', { name: /Группа ИТ-202/ }).click()
  await expect(
    page.getByRole('list', { name: 'Список диалогов' }).getByLabel('Непрочитанных сообщений: 2'),
  ).toHaveCount(0)
  await page.goto('/dashboard')
  await expect(
    page
      .getByRole('region', { name: 'Чаты', exact: true })
      .getByLabel('Непрочитанных сообщений: 2'),
  ).toHaveCount(0)
  await expect(page.locator('nav').first().getByLabel('Непрочитанных сообщений: 2')).toHaveCount(0)
  await second.close()
})

test('shows a failed message status and retries saving once storage recovers', async ({ page }) => {
  await page.goto('/messages?chat=curator')
  await page.evaluate(() => {
    const original = Storage.prototype.setItem
    Storage.prototype.setItem = function (key, value) {
      if (key.includes(':message:')) throw new DOMException('Quota exceeded', 'QuotaExceededError')
      original.call(this, key, value)
    }
    window.addEventListener(
      'test-enable-save',
      () => {
        Storage.prototype.setItem = original
      },
      { once: true },
    )
  })
  await page.getByRole('textbox', { name: 'Сообщение', exact: true }).fill('Сохранить после ошибки')
  await page.getByRole('button', { name: 'Отправить сообщение' }).click()
  const message = page.getByRole('log').getByRole('article').last()
  await expect(message.getByLabel('Сохранено')).toHaveCount(0)
  await expect(message.getByRole('button', { name: 'Повторить сохранение' })).toBeVisible()
  await expect(page.getByRole('alert')).toBeVisible()
  await page.evaluate(() => window.dispatchEvent(new Event('test-enable-save')))
  await page.getByRole('alert').getByRole('button', { name: 'Повторить', exact: true }).click()
  await expect(message.getByLabel('Сохранено')).toBeVisible()
  await expect(page.getByRole('alert')).toHaveCount(0)
  await page.reload()
  await expect(
    page.getByRole('log').getByText('Сохранить после ошибки', { exact: true }),
  ).toHaveCount(1)
})

test('filters only unread chats and restores keyboard focus on mobile', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await page.goto('/messages')
  const curator = page
    .getByRole('list', { name: 'Список диалогов' })
    .getByRole('button', { name: /Куратор группы/ })
  await curator.click()
  await expect(page.getByRole('heading', { name: 'Куратор группы', exact: true })).toBeFocused()
  await page.getByRole('button', { name: 'Назад к чатам' }).click()
  await expect(curator).toBeFocused()
  await page.getByRole('button', { name: /Непрочитанные/ }).click()
  const list = page.getByRole('list', { name: 'Список диалогов' })
  await expect(list.getByRole('button')).toHaveCount(1)
  await list.getByRole('button', { name: /Группа ИТ-202/ }).click()
  await page.getByRole('button', { name: 'Назад к чатам' }).click()
  await expect(page.getByText('Всё прочитано', { exact: true })).toBeVisible()
  await expect(page.getByRole('heading', { name: /Чаты/ })).toBeFocused()
})

test('sends without the secure-context randomUUID API', async ({ page }) => {
  await page.addInitScript(() => Object.defineProperty(crypto, 'randomUUID', { value: undefined }))
  await page.goto('/messages?chat=curator')
  await page
    .getByRole('textbox', { name: 'Сообщение', exact: true })
    .fill('Отправка по локальной сети')
  await page.getByRole('button', { name: 'Отправить сообщение' }).click()
  await expect(page.getByRole('log')).toContainText('Отправка по локальной сети')
  await page.reload()
  await expect(page.getByRole('log')).toContainText('Отправка по локальной сети')
})

test.describe('touch messages', () => {
  test.use({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true })

  test.beforeEach(async ({ page }) => {
    await page.addInitScript(() => {
      let height: number | undefined
      let top = 0
      const viewport = new EventTarget()
      Object.defineProperties(viewport, {
        height: { get: () => height ?? innerHeight },
        width: { get: () => innerWidth },
        offsetTop: { get: () => top },
        offsetLeft: { get: () => 0 },
        scale: { get: () => 1 },
      })
      Object.defineProperty(window, 'visualViewport', { value: viewport, configurable: true })
      window.addEventListener('test-keyboard-open', () => {
        height = 300
        top = 24
        viewport.dispatchEvent(new Event('resize'))
      })
      window.addEventListener('test-keyboard-close', () => {
        height = undefined
        top = 0
        viewport.dispatchEvent(new Event('resize'))
      })
    })
  })

  test('chat search fills the visible viewport without reserving space above the keyboard', async ({
    page,
  }) => {
    await page.goto('/messages')
    const search = page.getByPlaceholder('Поиск по чатам')
    const messenger = page.getByRole('region', { name: 'Сообщения', exact: true })
    const navigation = page.locator('nav').last()
    await search.tap()
    await search.fill('Групп')
    await page.evaluate(() => window.dispatchEvent(new Event('test-keyboard-open')))
    await expect(page.locator('body')).toHaveAttribute('data-message-keyboard', 'true')
    await expect(navigation).toBeHidden()
    await expect
      .poll(async () => {
        const bounds = (await messenger.boundingBox())!
        return bounds.y + bounds.height
      })
      .toBe(324)
    await expect(
      page.getByRole('list', { name: 'Список диалогов' }).getByRole('button'),
    ).toHaveCount(2)
    await search.fill('Нет такого пользователя')
    await expect(page.getByText('Чаты не найдены', { exact: true })).toBeVisible()

    // A blur can precede the keyboard closing animation; keep the full available height.
    await search.evaluate((element) => (element as HTMLInputElement).blur())
    await expect(navigation).toBeHidden()
    await page.evaluate(() => window.dispatchEvent(new Event('test-keyboard-close')))
    await expect(navigation).toBeVisible()
    await expect(page.locator('body')).not.toHaveAttribute('data-message-keyboard')
    await expect
      .poll(async () => {
        const bounds = (await messenger.boundingBox())!
        return bounds.y + bounds.height
      })
      .toBe((await navigation.boundingBox())!.y)
  })

  test('contact search in a teleported modal uses the keyboard viewport and restores navigation', async ({
    page,
  }) => {
    await page.goto('/messages?compose=1')
    const dialog = page.getByRole('dialog')
    const search = dialog.getByPlaceholder('Имя или фамилия')
    const navigation = page.locator('nav').last()
    await search.tap()
    await search.fill('Мария')
    await page.evaluate(() => window.dispatchEvent(new Event('test-keyboard-open')))
    await expect(page.locator('body')).toHaveAttribute('data-message-keyboard', 'true')
    await expect(navigation).toBeHidden()
    await expect(dialog.getByRole('list').getByRole('button')).toHaveCount(1)
    await expect
      .poll(async () => {
        const bounds = (await dialog.boundingBox())!
        return bounds.y + bounds.height
      })
      .toBe(324)
    await dialog.locator('header').getByRole('button').tap()
    await expect(dialog).toBeHidden()
    await page.evaluate(() => window.dispatchEvent(new Event('test-keyboard-close')))
    await expect(navigation).toBeVisible()
  })

  test('keeps sending accessible with the keyboard open and a long draft', async ({ page }) => {
    await page.goto('/messages?chat=curator')
    const field = page.getByRole('textbox', { name: 'Сообщение', exact: true })
    await field.tap()
    await field.fill('Длинное сообщение с телефона\n'.repeat(16))
    await page.evaluate(() => window.dispatchEvent(new Event('test-keyboard-open')))
    await expect(page.locator('body')).toHaveAttribute('data-message-keyboard', 'true')
    await expect(page.locator('nav').last()).toBeHidden()
    const send = page.getByRole('button', { name: 'Отправить сообщение' })
    await expect
      .poll(async () => {
        const bounds = (await send.boundingBox())!
        return bounds.y + bounds.height
      })
      .toBeLessThanOrEqual(324)
    await send.tap()
    await expect(field).toHaveValue('')
    await expect(field).toBeFocused()
    await expect(page.getByRole('log')).toContainText('Длинное сообщение с телефона')
    await page.evaluate(() => window.dispatchEvent(new Event('test-keyboard-close')))
    await expect(page.locator('nav').last()).toBeVisible()
  })

  test('sends by touch in a short viewport without overlapping navigation', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 300 })
    await page.goto('/messages?chat=curator')
    const field = page.getByRole('textbox', { name: 'Сообщение', exact: true })
    await field.tap()
    await field.fill('Короткое окно\n'.repeat(16))
    const send = page.getByRole('button', { name: 'Отправить сообщение' })
    await expect
      .poll(async () => {
        const button = (await send.boundingBox())!
        const navigation = (await page.locator('nav').last().boundingBox())!
        return button.y + button.height <= navigation.y
      })
      .toBe(true)
    await send.tap()
    await expect(page.getByRole('log')).toContainText('Короткое окно')
  })
})
