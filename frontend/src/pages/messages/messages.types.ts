export interface Message {
  id: string
  text: string
  sender: string
  own: boolean
  sentAt: string
}

export interface Chat {
  id: string
  name: string
  initials: string
  group: boolean
  description: string
  unread: number
  messages: Message[]
}

export interface SendPointer {
  id: number
  x: number
  y: number
}

export type MessageSavedChecker = (chatId: string, message: Message) => boolean
