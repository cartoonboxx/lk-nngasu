import type { LessonType, ScheduleDay, ScheduleLesson } from './schedule.types'

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
} satisfies Record<string, Omit<ScheduleLesson, 'id' | 'start' | 'end'>>

// Шкала времени для таблицы; занятия хранятся внутри дней.
export const periods = [
  { start: '08:00', end: '09:30' },
  { start: '09:45', end: '11:15' },
  { start: '11:30', end: '13:00' },
  { start: '13:20', end: '14:50' },
  { start: '15:10', end: '16:40' },
  { start: '16:55', end: '18:25' },
  { start: '18:35', end: '19:55' },
]

const monday = new Date()
monday.setHours(0, 0, 0, 0)
monday.setDate(monday.getDate() - ((monday.getDay() + 6) % 7))

// Демо на текущую неделю. Ответ API будет иметь такую же структуру: date и lessons.
export const scheduleDays: ScheduleDay[] = [
  {
    lessons: [
      { id: '1-1', start: '08:00', end: '09:30', ...lessons.mathLecture },
      { id: '1-2', start: '09:45', end: '11:15', ...lessons.programmingPractice },
      { id: '1-3', start: '11:30', end: '13:00', ...lessons.physicsLab },
      { id: '1-4', start: '13:20', end: '14:50', ...lessons.webLecture },
    ],
  },
  {
    lessons: [
      { id: '2-1', start: '08:00', end: '09:30', ...lessons.mathLecture },
      { id: '2-2', start: '09:45', end: '11:15', ...lessons.databasePractice },
      { id: '2-3', start: '11:30', end: '13:00', ...lessons.englishPractice },
      { id: '2-5', start: '15:10', end: '16:40', ...lessons.webPractice },
      { id: '2-6', start: '16:55', end: '18:25', ...lessons.programmingLecture },
      { id: '2-7', start: '18:35', end: '19:55', ...lessons.programmingLecture },
    ],
  },
  {
    lessons: [
      { id: '3-1', start: '08:00', end: '09:30', ...lessons.physicsLecture },
      { id: '3-3', start: '11:30', end: '13:00', ...lessons.programmingLab },
      { id: '3-4', start: '13:20', end: '14:50', ...lessons.databaseLecture },
      { id: '3-5', start: '15:10', end: '16:40', ...lessons.englishPractice },
    ],
  },
  {
    lessons: [
      { id: '4-1', start: '08:00', end: '09:30', ...lessons.programmingLecture },
      { id: '4-2', start: '09:45', end: '11:15', ...lessons.mathPractice },
      { id: '4-4', start: '13:20', end: '14:50', ...lessons.englishLecture },
      { id: '4-5', start: '15:10', end: '16:40', ...lessons.physicsLab },
      { id: '4-6', start: '16:55', end: '18:25', ...lessons.webLab },
      { id: '4-7', start: '18:35', end: '19:55', ...lessons.webLab },
    ],
  },
  {
    lessons: [
      { id: '5-2', start: '09:45', end: '11:15', ...lessons.webPractice },
      { id: '5-3', start: '11:30', end: '13:00', ...lessons.databaseLecture },
      { id: '5-5', start: '15:10', end: '16:40', ...lessons.mathLecture },
      { id: '5-6', start: '16:55', end: '18:25', ...lessons.englishPractice },
      { id: '5-7', start: '18:35', end: '19:55', ...lessons.englishPractice },
    ],
  },
  {
    lessons: [
      { id: '6-1', start: '08:00', end: '09:30', ...lessons.englishLecture },
      { id: '6-3', start: '11:30', end: '13:00', ...lessons.physicsLab },
      { id: '6-4', start: '13:20', end: '14:50', ...lessons.programmingPractice },
    ],
  },
].map((day, index) => {
  const date = new Date(monday)
  date.setDate(date.getDate() + index)
  return {
    date: `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`,
    lessons: day.lessons,
  }
})
