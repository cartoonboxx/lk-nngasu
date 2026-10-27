export type LessonType = 'lecture' | 'practice' | 'lab'

export interface ScheduleLesson {
  subject: string
  type: LessonType
  location: string
  teacher: string
}

export const lessonTypes: { type: LessonType; label: string }[] = [
  { type: 'lecture', label: 'Лекция' },
  { type: 'practice', label: 'Практика' },
  { type: 'lab', label: 'Лабораторная' },
]

// Демонстрационное наполнение до подключения API расписания.
const lessons = {
  mathLecture: {
    subject: 'Математический анализ',
    type: 'lecture',
    location: 'Корпус А · ауд. 302',
    teacher: 'Иванова Е. С.',
  },
  mathPractice: {
    subject: 'Математический анализ',
    type: 'practice',
    location: 'Корпус А · ауд. 305',
    teacher: 'Иванова Е. С.',
  },
  programmingLecture: {
    subject: 'Программирование',
    type: 'lecture',
    location: 'Корпус Б · ауд. 415',
    teacher: 'Петров А. В.',
  },
  programmingPractice: {
    subject: 'Программирование',
    type: 'practice',
    location: 'Корпус Б · ауд. 415',
    teacher: 'Петров А. В.',
  },
  programmingLab: {
    subject: 'Программирование',
    type: 'lab',
    location: 'Корпус Б · ауд. 415',
    teacher: 'Петров А. В.',
  },
  physicsLecture: {
    subject: 'Физика',
    type: 'lecture',
    location: 'Корпус А · ауд. 201',
    teacher: 'Смирнов Д. А.',
  },
  physicsLab: {
    subject: 'Физика',
    type: 'lab',
    location: 'Корпус А · ауд. 201',
    teacher: 'Смирнов Д. А.',
  },
  databaseLecture: {
    subject: 'Базы данных',
    type: 'lecture',
    location: 'Корпус Б · ауд. 410',
    teacher: 'Соколова А. Н.',
  },
  databasePractice: {
    subject: 'Базы данных',
    type: 'practice',
    location: 'Корпус Б · ауд. 410',
    teacher: 'Соколова А. Н.',
  },
  englishLecture: {
    subject: 'Английский язык',
    type: 'lecture',
    location: 'Корпус Б · ауд. 103',
    teacher: 'Кузнецова М. О.',
  },
  englishPractice: {
    subject: 'Английский язык',
    type: 'practice',
    location: 'Корпус Б · ауд. 103',
    teacher: 'Кузнецова М. О.',
  },
  webLecture: {
    subject: 'Веб-разработка',
    type: 'lecture',
    location: 'Корпус Б · ауд. 407',
    teacher: 'Орлова К. В.',
  },
  webPractice: {
    subject: 'Веб-разработка',
    type: 'practice',
    location: 'Корпус Б · ауд. 407',
    teacher: 'Орлова К. В.',
  },
  webLab: {
    subject: 'Веб-разработка',
    type: 'lab',
    location: 'Корпус Б · ауд. 407',
    teacher: 'Орлова К. В.',
  },
} satisfies Record<string, ScheduleLesson>

type LessonKey = keyof typeof lessons
const week: (LessonKey | null)[][] = [
  ['mathLecture', 'mathLecture', 'physicsLecture', 'programmingLecture', null, 'englishLecture'],
  ['programmingPractice', 'databasePractice', null, 'mathPractice', 'webPractice', null],
  ['physicsLab', 'englishPractice', 'programmingLab', null, 'databaseLecture', 'physicsLab'],
  ['webLecture', null, 'databaseLecture', 'englishLecture', null, 'programmingPractice'],
  [null, 'webPractice', 'englishPractice', 'physicsLab', 'mathLecture', null],
  [null, 'programmingLecture', null, 'webLab', 'englishPractice', null],
  [null, 'programmingLecture', null, 'webLab', 'englishPractice', null],
]

export const periods = [
  { start: '08:00', end: '09:30' },
  { start: '09:45', end: '11:15' },
  { start: '11:30', end: '13:00' },
  { start: '13:20', end: '14:50' },
  { start: '15:10', end: '16:40' },
  { start: '16:55', end: '18:25' },
  { start: '18:35', end: '19:55' },
].map((period, index) => ({
  ...period,
  lessons: week[index]!.map((key) => (key ? lessons[key] : null)),
}))
