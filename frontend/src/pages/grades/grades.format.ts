const gradeDateFormatter = new Intl.DateTimeFormat('ru', { day: 'numeric', month: 'short' })

export function formatGradeDate(value: string | null): string {
  return value ? gradeDateFormatter.format(new Date(`${value}T12:00:00`)) : '—'
}
