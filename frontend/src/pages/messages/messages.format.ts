import type { Chat } from './messages.types'

export function lastMessage(chat: Chat) {
  return chat.messages[chat.messages.length - 1]
}
export function messageTime(date: string): string {
  return new Date(date).toLocaleTimeString('ru', { hour: '2-digit', minute: '2-digit' })
}
export function dayLabel(date: string): string {
  const day = new Date(date).toDateString()
  const now = new Date()
  if (day === now.toDateString()) return 'Сегодня'
  now.setDate(now.getDate() - 1)
  if (day === now.toDateString()) return 'Вчера'
  return new Date(date).toLocaleDateString('ru', { day: 'numeric', month: 'long', year: 'numeric' })
}
