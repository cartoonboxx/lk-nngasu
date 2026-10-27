import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, disposePinia, type Pinia } from 'pinia'
import { useMessageStore, messagesStoragePrefix } from '@/stores/messages'
import { createApp, defineComponent, h, nextTick } from 'vue'
import stores from '@/stores'
import { createChats } from '@/pages/messages/messages.data'

const instances: Pinia[] = []
function createStore() {
  const pinia = createPinia()
  instances.push(pinia)
  const store = useMessageStore(pinia)
  store.initialize()
  return store
}

beforeEach(() => {
  vi.restoreAllMocks()
  localStorage.clear()
  sessionStorage.clear()
})

afterEach(() => {
  vi.unstubAllGlobals()
  instances.splice(0).forEach(disposePinia)
})

describe('messages persistence', () => {
  it('merges independent writers without dropping either message', () => {
    const first = createStore()
    const second = createStore()
    first.send('curator', 'Первая вкладка')
    second.send('curator', 'Вторая вкладка')
    first.refresh()
    const reloaded = createStore()
    const texts = reloaded.chats
      .find((chat) => chat.id === 'curator')!
      .messages.map((message) => message.text)
    expect(texts).toContain('Первая вкладка')
    expect(texts).toContain('Вторая вкладка')
    expect(
      first.chats.find((chat) => chat.id === 'curator')!.messages.map((message) => message.text),
    ).toEqual(texts)
  })

  it('preserves drafts per conversation through reinitialization', () => {
    const store = createStore()
    store.setDraft('curator', 'Вопрос куратору')
    store.setDraft('alexey', 'Вопрос Алексею')
    const reloaded = createStore()
    expect(reloaded.drafts).toEqual({ curator: 'Вопрос куратору', alexey: 'Вопрос Алексею' })
    reloaded.send('curator', reloaded.drafts.curator!)
    expect(reloaded.drafts.curator).toBe('')
    expect(reloaded.drafts.alexey).toBe('Вопрос Алексею')
  })

  it('keeps failed messages in memory and retries without creating duplicates', () => {
    const store = createStore()
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
    expect(store.storageError).toBe(true)
    expect(store.isSaved('curator', message)).toBe(false)
    expect(store.chats.find((chat) => chat.id === 'curator')!.messages).toContainEqual(message)
    spy.mockRestore()
    store.retry()
    expect(store.storageError).toBe(false)
    expect(store.isSaved('curator', message)).toBe(true)
    store.retry()
    const reloaded = createStore()
    expect(
      reloaded.chats
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
    const store = createStore()
    expect(store.chats[0]!.messages).toContainEqual(
      chats[0]!.messages[chats[0]!.messages.length - 1],
    )
    expect(store.unreadMessages).toBe(0)
  })

  it('shares read receipts without rewriting history', () => {
    const first = createStore()
    const second = createStore()
    first.markRead('group-it-202')
    second.send('curator', 'Другая переписка')
    const reloaded = createStore()
    expect(reloaded.unreadMessages).toBe(0)
    expect(reloaded.chats.find((chat) => chat.id === 'group-it-202')!.messages).toHaveLength(6)
  })

  it('sends messages when randomUUID is unavailable on a local HTTP origin', () => {
    const getRandomValues = crypto.getRandomValues.bind(crypto)
    vi.stubGlobal('crypto', { getRandomValues })
    const store = createStore()
    const message = store.send('curator', 'Сообщение без randomUUID')!
    expect(message.id).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/,
    )
    expect(store.isSaved('curator', message)).toBe(true)
  })
})

describe('message store registration', () => {
  it('exposes the reactive Pinia store through $messageStore and cleans up synchronization', async () => {
    const app = createApp(
      defineComponent({
        render() {
          return h('span', String(this.$messageStore.unreadMessages))
        },
      }),
    )
    app.use(stores)
    const target = document.createElement('div')
    app.mount(target)
    const store = app.config.globalProperties.$messageStore
    try {
      expect(store.$id).toBe('messages')
      expect(target.textContent).toBe('2')
      store.markRead('group-it-202')
      await nextTick()
      expect(target.textContent).toBe('0')
      const refresh = vi.spyOn(store, 'refresh')
      window.dispatchEvent(new StorageEvent('storage', { key: 'unrelated-setting' }))
      expect(refresh).not.toHaveBeenCalled()
      window.dispatchEvent(
        new StorageEvent('storage', { key: `${messagesStoragePrefix}read:curator` }),
      )
      expect(refresh).toHaveBeenCalledOnce()
      window.dispatchEvent(new StorageEvent('storage', { key: null }))
      expect(refresh).toHaveBeenCalledTimes(2)
      app.unmount()
      window.dispatchEvent(new StorageEvent('storage', { key: null }))
      expect(refresh).toHaveBeenCalledTimes(2)
    } finally {
      if (target.hasChildNodes()) app.unmount()
    }
  })
})
