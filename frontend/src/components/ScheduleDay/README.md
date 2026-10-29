`ScheduleDay` — карточка дня со списком занятий. Работает на любой странице и при любой ширине контейнера; не требует store расписания. Цвета типов имеют значения по умолчанию и могут переопределяться переменными `--schedule-lecture`, `--schedule-practice`, `--schedule-lab`.

```vue
<template>
  <ScheduleDay title="Расписание на сегодня" date="2026-10-26" :lessons="lessons">
    <template #header-actions>
      <!-- Дата, кнопки переключения дней или выбора периода -->
    </template>
    <template #lesson-trailing="{ lesson }">
      <!-- Статус или действие для занятия; lesson.id — ключ записи -->
    </template>
  </ScheduleDay>
</template>

<script setup lang="ts">
import ScheduleDay from '@/components/ScheduleDay/ScheduleDay.vue'
import type { ScheduleDayItem } from '@/components/ScheduleDay/schedule-day.types'

const lessons: ScheduleDayItem[] = [
  {
    id: 'math-1',
    start: '08:30',
    end: '10:00',
    subject: 'Математический анализ',
    type: 'lecture',
    typeLabel: 'Лекция',
    location: 'Корпус А, ауд. 302',
    teacher: 'Иванова Елена Сергеевна',
  },
]
</script>
```

Props: `title`, `lessons` обязательны; `date` задаёт ISO-дату для `<time>`, `today` выделяет заголовок.

Слоты: `header` заменяет заголовок и получает `title`, `date`, `today`; `header-actions` дополняет шапку; `lesson-trailing` получает `lesson`; `empty` заменяет сообщение для пустого списка.

`layout="grid"` располагает занятия в две колонки в широком контейнере и переключается в одну на узком. По умолчанию используется `layout="list"`.
