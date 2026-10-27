import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createMessagesStore } from '@/stores/messages'
import { createChats } from '@/pages/messages/messages.data'

beforeEach(() => {
  vi.restoreAllMocks()
  localStorage.clear()
  sessionStorage.clear()
})

afterEach(() => vi.unstubAllGlobals())

describe('messages persistence', () => {
  it('merges independent writers without dropping either message', () => {
    const first = createMessagesStore(localStorage, sessionStorage)
    const second = createMessagesStore(localStorage, sessionStorage)
    first.send('curator', 'Первая вкладка')
    second.send('curator', 'Вторая вкладка')
    first.refresh()
    const reloaded = createMessagesStore(localStorage, sessionStorage)
    const texts = reloaded.state.chats
      .find((chat) => chat.id === 'curator')!
      .messages.map((message) => message.text)
    expect(texts).toContain('Первая вкладка')
    expect(texts).toContain('Вторая вкладка')
    expect(
      first.state.chats
        .find((chat) => chat.id === 'curator')!
        .messages.map((message) => message.text),
    ).toEqual(texts)
  })

  it('preserves drafts per conversation through reinitialization', () => {
    const store = createMessagesStore(localStorage, sessionStorage)
    store.setDraft('curator', 'Вопрос куратору')
    store.setDraft('alexey', 'Вопрос Алексею')
    const reloaded = createMessagesStore(localStorage, sessionStorage)
    expect(reloaded.state.drafts).toEqual({ curator: 'Вопрос куратору', alexey: 'Вопрос Алексею' })
    reloaded.send('curator', reloaded.state.drafts.curator!)
    expect(reloaded.state.drafts.curator).toBe('')
    expect(reloaded.state.drafts.alexey).toBe('Вопрос Алексею')
  })

  it('keeps failed messages in memory and retries without creating duplicates', () => {
    const store = createMessagesStore(localStorage, sessionStorage)
    const original = Storage.prototype.setItem
    const spy = vi.spyOn(Storage.prototype, 'setItem').mockImplementation(function (
      this: Storage,
      key,
      value,
    ) {
      if (key.includes(':message:')) throw new DOMException('Quota exceeded', 'QuotaExceededError')
      original.call(this, key, value)
    })
    const message = store.send('curator', 'Сообщение для повторного сохранения')!
    expect(store.storageError.value).toBe(true)
    expect(store.isSaved('curator', message)).toBe(false)
    expect(store.state.chats.find((chat) => chat.id === 'curator')!.messages).toContainEqual(
      message,
    )
    spy.mockRestore()
    store.retry()
    expect(store.storageError.value).toBe(false)
    expect(store.isSaved('curator', message)).toBe(true)
    store.retry()
    const reloaded = createMessagesStore(localStorage, sessionStorage)
    expect(
      reloaded.state.chats
        .find((chat) => chat.id === 'curator')!
        .messages.filter((item) => item.id === message.id),
    ).toHaveLength(1)
  })

  it('preserves history and unread counts when migrating the previous storage format', () => {
    const chats = createChats()
    chats[0]!.unread = 0
    chats[0]!.messages.push({
      id: 'legacy-message',
      own: true,
      sender: 'Вы',
      text: 'Старое сообщение',
      sentAt: new Date().toISOString(),
    })
    localStorage.setItem('lk-nngasu-messages-v1', JSON.stringify(chats))
    const store = createMessagesStore(localStorage, sessionStorage)
    expect(store.state.chats[0]!.messages).toContainEqual(
      chats[0]!.messages[chats[0]!.messages.length - 1],
    )
    expect(store.unreadMessages.value).toBe(0)
  })

  it('shares read receipts without rewriting history', () => {
    const first = createMessagesStore(localStorage, sessionStorage)
    const second = createMessagesStore(localStorage, sessionStorage)
    first.markRead('group-it-202')
    second.send('curator', 'Другая переписка')
    const reloaded = createMessagesStore(localStorage, sessionStorage)
    expect(reloaded.unreadMessages.value).toBe(0)
    expect(reloaded.state.chats.find((chat) => chat.id === 'group-it-202')!.messages).toHaveLength(
      6,
    )
  })

  it('sends messages when randomUUID is unavailable on a local HTTP origin', () => {
    const getRandomValues = crypto.getRandomValues.bind(crypto)
    vi.stubGlobal('crypto', { getRandomValues })
    const store = createMessagesStore(localStorage, sessionStorage)
    const message = store.send('curator', 'Сообщение без randomUUID')!
    expect(message.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    )
    expect(store.isSaved('curator', message)).toBe(true)
  })
})
