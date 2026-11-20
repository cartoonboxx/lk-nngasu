export interface Message {
  id: string
  text: string
  sender: string
  own: boolean
  files: { id: number; name: string }[]
  sentAt: string
  editedAt: string | null
}

export interface Chat {
  unreadCount: number
  lastReadMessageId: number
  id: string
  name: string
  initials: string
  group: boolean
  description: string
  messages: Message[]
  lastMessage?: Message
}

export interface SendPointer {
  id: number
  x: number
  y: number
}
