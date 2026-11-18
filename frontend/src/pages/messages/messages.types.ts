export interface Message {
  id: string
  text: string
  sender: string
  own: boolean
  files: { id: number; name: string }[]
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
  lastMessage?: Message
}

export interface SendPointer {
  id: number
  x: number
  y: number
}
