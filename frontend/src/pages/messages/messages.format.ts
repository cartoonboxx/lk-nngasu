export const messageTimeFormatter = new Intl.DateTimeFormat('ru', {
  hour: '2-digit',
  minute: '2-digit',
})

export function dayLabel(date: string): string {
  const day = new Date(date).toDateString()
  const now = new Date()
  if (day === now.toDateString()) return 'Сегодня'
  now.setDate(now.getDate() - 1)
  if (day === now.toDateString()) return 'Вчера'
  return new Date(date).toLocaleDateString('ru', { day: 'numeric', month: 'long', year: 'numeric' })
}
