export type GradeStatus = 'reviewed' | 'pending' | 'not-submitted'

export interface GradeAssignment {
  id: number
  title: string
  topic: string
  subject: string
  semester: string
  submittedAt: string | null
  grade: number | null
  status: GradeStatus
  teacherComment: string
  answer: string
}

export const gradeStatuses: Record<GradeStatus, string> = {
  reviewed: 'Проверено',
  pending: 'На проверке',
  'not-submitted': 'Не отправлено',
}

// Demonstration data until the grades API is available.
export const gradeAssignments: GradeAssignment[] = [
  {
    id: 1,
    title: 'Практическая работа №1',
    topic: 'Тема 1 · Факты и правила',
    subject: 'Логическое программирование',
    semester: '7',
    submittedAt: '2026-10-12',
    grade: 5,
    status: 'reviewed',
    teacherComment: 'Все запросы выполнены верно',
    answer:
      'Создана база знаний о родственных связях. Добавлены факты, правила и пять запросов для проверки.',
  },
  {
    id: 2,
    title: 'Лабораторная работа №2',
    topic: 'Тема 2 · Компоненты Vue',
    subject: 'Веб-технологии',
    semester: '7',
    submittedAt: '2026-10-11',
    grade: 4,
    status: 'reviewed',
    teacherComment: 'Добавьте обработку ошибок',
    answer:
      'Реализованы компоненты формы и списка. Настроена передача данных через props и события.',
  },
  {
    id: 3,
    title: 'Практическая работа №3',
    topic: 'Тема 3 · SQL-запросы',
    subject: 'Базы данных',
    semester: '7',
    submittedAt: '2026-10-10',
    grade: 5,
    status: 'reviewed',
    teacherComment: 'Отличная работа',
    answer: 'Подготовлены запросы с объединением таблиц, группировкой и вложенными выборками.',
  },
  {
    id: 4,
    title: 'Лабораторная работа №1',
    topic: 'Тема 1 · Интерфейс приложения',
    subject: 'Мобильная разработка',
    semester: '7',
    submittedAt: '2026-10-09',
    grade: 5,
    status: 'reviewed',
    teacherComment: 'Все требования выполнены',
    answer:
      'Разработан интерфейс приложения с навигацией между экранами и адаптацией под разные размеры экрана.',
  },
  {
    id: 5,
    title: 'Практическая работа №2',
    topic: 'Тема 2 · Тестовые сценарии',
    subject: 'Качество и тестирование ПО',
    semester: '7',
    submittedAt: '2026-10-08',
    grade: 4,
    status: 'reviewed',
    teacherComment: 'Уточните ожидаемые результаты',
    answer:
      'Составлены тестовые сценарии авторизации, включая проверку неверных и пустых значений.',
  },
  {
    id: 6,
    title: 'Практическая работа №2',
    topic: 'Тема 2 · Унификация',
    subject: 'Логическое программирование',
    semester: '7',
    submittedAt: '2026-10-13',
    grade: null,
    status: 'pending',
    teacherComment: '',
    answer:
      'Описана предметная область с помощью фактов и правил. Проверена унификация на примерах запросов.',
  },
  {
    id: 7,
    title: 'Лабораторная работа №3',
    topic: 'Тема 3 · Маршрутизация',
    subject: 'Веб-технологии',
    semester: '7',
    submittedAt: '2026-10-13',
    grade: null,
    status: 'pending',
    teacherComment: '',
    answer:
      'Настроены маршруты приложения, динамические параметры и страница для неизвестного адреса.',
  },
  {
    id: 8,
    title: 'Практическая работа №4',
    topic: 'Тема 4 · Транзакции',
    subject: 'Базы данных',
    semester: '7',
    submittedAt: null,
    grade: null,
    status: 'not-submitted',
    teacherComment: '',
    answer: '',
  },
  {
    id: 9,
    title: 'Практическая работа №2',
    topic: 'Тема 2 · Работа с API',
    subject: 'Мобильная разработка',
    semester: '7',
    submittedAt: null,
    grade: null,
    status: 'not-submitted',
    teacherComment: '',
    answer: '',
  },
]
