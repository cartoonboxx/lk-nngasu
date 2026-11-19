import { defineStore } from 'pinia'
import { markRaw, toRaw } from 'vue'

// Кеш существует только в памяти текущей сессии. Внешние изменения видны
// при следующем открытии раздела после истечения пяти минут или перезагрузке страницы.
export const useDataCacheStore = defineStore('dataCache', {
  state: () => ({
    entries: {} as Record<string, { value: unknown; expiresAt: number }>,
    pending: markRaw(new Map<string, Promise<unknown>>()),
    version: 0,
  }),
  actions: {
    async load<T>(key: string, loader: () => Promise<T>, fresh = false): Promise<T> {
      const cached = this.entries[key]
      if (!fresh && cached && cached.expiresAt > Date.now())
        return structuredClone(toRaw(cached.value)) as T
      if (!fresh && this.pending.has(key)) return structuredClone(await this.pending.get(key)) as T

      const version = this.version
      const request = loader().then((value) => {
        if (version === this.version && this.pending.get(key) === request)
          this.entries[key] = { value: structuredClone(value), expiresAt: Date.now() + 5 * 60_000 }
        return value
      })
      this.pending.set(key, request)
      try {
        // Страницы редактируют свои копии: формы не меняют сохранённый ответ сервера.
        return structuredClone(await request)
      } finally {
        if (this.pending.get(key) === request) this.pending.delete(key)
      }
    },
    invalidate() {
      this.version++
      this.entries = {}
      this.pending.clear()
    },
  },
})
