import { test, expect } from '@playwright/test'

test('student submits, edits and deletes a separate answer for each topic', async ({ page }) => {
  await page.goto('/subjects/logic')
  await page.getByRole('button', { name: 'Открыть задание', exact: true }).click()
  const dialog = page.getByRole('dialog')
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('button', { name: 'Отправить преподавателю' })).toBeDisabled()
  await dialog.getByLabel('Файлы работы', { exact: true }).setInputFiles({
    name: 'answer.pl',
    mimeType: 'text/plain',
    buffer: Buffer.from('parent(ann, bob).'),
  })
  await dialog.getByLabel('Комментарий к ответу').fill('Первый ответ')
  await dialog.getByRole('button', { name: 'Отправить преподавателю' }).click()
  await expect(dialog.getByRole('status')).toContainText('Ожидает проверки')
  await expect(dialog.getByLabel('Отправленный ответ')).toContainText('answer.pl')
  await dialog.getByRole('button', { name: 'Редактировать ответ' }).click()
  await dialog.getByLabel('Комментарий к ответу').fill('Исправленный ответ')
  await dialog.getByRole('button', { name: 'Сохранить изменения' }).click()
  await page.keyboard.press('Escape')
  await page.getByRole('button', { name: 'Следующая тема' }).click()
  await page.getByRole('button', { name: 'Открыть задание', exact: true }).click()
  await expect(dialog.getByRole('status')).toContainText('Не отправлено')
  await page.keyboard.press('Escape')
  await page.getByRole('button', { name: 'Предыдущая тема' }).click()
  await page.getByRole('button', { name: 'Открыть задание', exact: true }).click()
  await expect(dialog.getByLabel('Отправленный ответ')).toContainText('Исправленный ответ')
  await dialog.getByRole('button', { name: 'Удалить ответ', exact: true }).click()
  await dialog.getByRole('button', { name: 'Подтвердить удаление' }).click()
  await expect(dialog.getByRole('status')).toContainText('Не отправлено')
  await expect(dialog.getByLabel('Комментарий к ответу')).toHaveValue('')
})

test('assignment remains usable on a narrow screen', async ({ page }) => {
  await page.goto('/subjects/logic')
  await page.getByRole('button', { name: 'Открыть задание', exact: true }).click()
  // Resize the open dialog: the existing course layout has a fixed desktop sidebar.
  await page.setViewportSize({ width: 320, height: 600 })
  const dialog = page.getByRole('dialog')
  await dialog.getByLabel('Комментарий к ответу').fill('Ответ без вложений')
  const submit = dialog.getByRole('button', { name: 'Отправить преподавателю' })
  await submit.scrollIntoViewIfNeeded()
  const box = await submit.boundingBox()
  expect(box).not.toBeNull()
  expect(box!.x).toBeGreaterThanOrEqual(0)
  expect(box!.x + box!.width).toBeLessThanOrEqual(320)
  await submit.click()
  await expect(dialog.getByRole('status')).toContainText('Ожидает проверки')
  await dialog.getByRole('button', { name: 'Закрыть задание' }).click()
  await expect(dialog).toBeHidden()
})

test('assignment accepts dropped files and limits the comment', async ({ page }) => {
  await page.goto('/subjects/logic')
  await page.getByRole('button', { name: 'Открыть задание', exact: true }).click()
  const dialog = page.getByRole('dialog')
  await dialog.getByLabel('Комментарий к ответу').fill('а'.repeat(2010))
  await expect(dialog.getByLabel('Комментарий к ответу')).toHaveValue('а'.repeat(2000))
  await expect(dialog.getByText('2000 / 2000', { exact: true })).toBeVisible()
  const transfer = await page.evaluateHandle(() => {
    const data = new DataTransfer()
    data.items.add(new File(['parent(ann, bob).'], 'dropped.pl', { type: 'text/plain' }))
    return data
  })
  await dialog.locator('input[type="file"]').dispatchEvent('drop', { dataTransfer: transfer })
  await expect(dialog.getByLabel('Прикреплённые файлы')).toContainText('dropped.pl')
  await dialog.getByRole('button', { name: 'Отправить преподавателю' }).click()
  await expect(dialog.getByLabel('Файлы ответа')).toContainText('dropped.pl')
})

test('cancelling the file picker keeps the assignment and its draft open', async ({ page }) => {
  await page.goto('/subjects/logic')
  await page.getByRole('button', { name: 'Открыть задание', exact: true }).click()
  const dialog = page.getByRole('dialog')
  await dialog.getByLabel('Комментарий к ответу').fill('Черновик ответа')
  const files = dialog.getByLabel('Файлы работы', { exact: true })
  await files.setInputFiles({
    name: 'answer.pl',
    mimeType: 'text/plain',
    buffer: Buffer.from('facts'),
  })
  await files.dispatchEvent('cancel', { bubbles: true, cancelable: true })
  await expect(dialog).toBeVisible()
  await expect(dialog.getByLabel('Комментарий к ответу')).toHaveValue('Черновик ответа')
  await expect(dialog.getByLabel('Прикреплённые файлы')).toContainText('answer.pl')
  await page.keyboard.press('Escape')
  await expect(dialog).toBeHidden()
})
