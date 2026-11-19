import { defineStore } from 'pinia'
import { useAuthStore } from './auth'
import type { FileInfo } from '@/api/cabinet.types'

export const useFilesStore = defineStore('files', {
  actions: {
    async upload(file: File, purpose: 'attachment' | 'avatar' = 'attachment') {
      const body = new FormData()
      body.append('file', file)
      body.append('purpose', purpose)
      return useAuthStore().request<FileInfo>('/files', { method: 'POST', body })
    },
    async download(id: number | string, name: string) {
      const blob = await useAuthStore().request<Blob>(
        typeof id === 'number' ? `/files/${id}` : id,
        { responseType: 'blob' },
      )
      const url = URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = name
      link.click()
      // Дать браузеру начать загрузку до освобождения Blob URL.
      window.setTimeout(() => URL.revokeObjectURL(url), 1000)
    },
  },
})
