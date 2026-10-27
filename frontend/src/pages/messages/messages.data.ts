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

export const contacts = [
  { id: 'curator', name: 'Куратор группы', initials: 'КГ', description: 'Куратор · ИТ-202' },
  { id: 'alexey', name: 'Алексей Смирнов', initials: 'АС', description: 'Студент · ИТ-202' },
  { id: 'maria', name: 'Мария Кузнецова', initials: 'МК', description: 'Студент · ИТ-202' },
  { id: 'ivan', name: 'Иван Петров', initials: 'ИП', description: 'Студент · ИТ-202' },
]

function sentAt(hour: number, minute: number, yesterday = false): string {
  const date = new Date()
  if (yesterday) date.setDate(date.getDate() - 1)
  date.setHours(hour, minute, 0, 0)
  return date.toISOString()
}

export function createChats(): Chat[] {
  return [
    {
      id: 'group-it-202',
      name: 'Группа ИТ-202',
      initials: 'ИТ',
      group: true,
      description: 'Учебный чат группы',
      unread: 2,
      messages: [
        {
          id: 'g1',
          sender: 'Мария Кузнецова',
          text: 'Всем привет! Напоминаю, завтра сдаём лабораторную по программированию.',
          own: false,
          sentAt: sentAt(16, 20, true),
        },
        {
          id: 'g2',
          sender: 'Вы',
          text: 'Спасибо за напоминание! А отчёт нужно распечатать?',
          own: true,
          sentAt: sentAt(16, 24, true),
        },
        {
          id: 'g3',
          sender: 'Мария Кузнецова',
          text: 'Нет, достаточно загрузить файл в личный кабинет.',
          own: false,
          sentAt: sentAt(16, 26, true),
        },
        {
          id: 'g4',
          sender: 'Алексей Смирнов',
          text: 'Доброе утро! Практика сегодня в 415-й аудитории корпуса Б.',
          own: false,
          sentAt: sentAt(10, 12),
        },
        {
          id: 'g5',
          sender: 'Вы',
          text: 'Отлично, буду. Возьму с собой ноутбук.',
          own: true,
          sentAt: sentAt(10, 18),
        },
        {
          id: 'g6',
          sender: 'Иван Петров',
          text: 'Кто будет на практике сегодня?',
          own: false,
          sentAt: sentAt(10, 24),
        },
      ],
    },
    {
      id: 'curator',
      name: 'Куратор группы',
      initials: 'КГ',
      group: false,
      description: 'Куратор · ИТ-202',
      unread: 0,
      messages: [
        {
          id: 'c1',
          sender: 'Куратор группы',
          text: 'Доброе утро! Пришлите, пожалуйста, список участников конференции от вашей группы.',
          own: false,
          sentAt: sentAt(9, 30),
        },
        {
          id: 'c2',
          sender: 'Вы',
          text: 'Доброе утро! Соберу список и отправлю сегодня до 15:00.',
          own: true,
          sentAt: sentAt(9, 38),
        },
        {
          id: 'c3',
          sender: 'Куратор группы',
          text: 'Хорошо, спасибо!',
          own: false,
          sentAt: sentAt(9, 42),
        },
      ],
    },
    {
      id: 'general',
      name: 'Общий чат',
      initials: 'ОЧ',
      group: true,
      description: 'Общение студентов университета',
      unread: 0,
      messages: [
        {
          id: 'a1',
          sender: 'Алексей Смирнов',
          text: 'Здесь можно обсудить учёбу и поделиться полезными материалами.',
          own: false,
          sentAt: sentAt(14, 10, true),
        },
        {
          id: 'a2',
          sender: 'Мария Кузнецова',
          text: 'Привет! Есть вопросы по заданию?',
          own: false,
          sentAt: sentAt(14, 35, true),
        },
      ],
    },
    {
      id: 'alexey',
      name: 'Алексей Смирнов',
      initials: 'АС',
      group: false,
      description: 'Студент · ИТ-202',
      unread: 0,
      messages: [
        {
          id: 's1',
          sender: 'Вы',
          text: 'Привет! Давай вместе разберём задание по базам данных?',
          own: true,
          sentAt: sentAt(18, 5, true),
        },
        {
          id: 's2',
          sender: 'Алексей Смирнов',
          text: 'Да, давай. Мне удобно после занятий.',
          own: false,
          sentAt: sentAt(18, 12, true),
        },
        {
          id: 's3',
          sender: 'Вы',
          text: 'Тогда завтра после последней пары?',
          own: true,
          sentAt: sentAt(18, 14, true),
        },
        {
          id: 's4',
          sender: 'Алексей Смирнов',
          text: 'Договорились, созвонимся завтра.',
          own: false,
          sentAt: sentAt(18, 16, true),
        },
      ],
    },
  ]
}

export function isChat(value: unknown): value is Chat {
  if (!value || typeof value !== 'object') return false
  const chat = value as Chat
  return (
    typeof chat.id === 'string' &&
    typeof chat.name === 'string' &&
    typeof chat.initials === 'string' &&
    typeof chat.group === 'boolean' &&
    typeof chat.description === 'string' &&
    Number.isInteger(chat.unread) &&
    chat.unread >= 0 &&
    Array.isArray(chat.messages) &&
    chat.messages.every(
      (message: Message) =>
        message &&
        typeof message.id === 'string' &&
        typeof message.text === 'string' &&
        typeof message.sender === 'string' &&
        typeof message.own === 'boolean' &&
        typeof message.sentAt === 'string' &&
        Number.isFinite(Date.parse(message.sentAt)),
    )
  )
}
