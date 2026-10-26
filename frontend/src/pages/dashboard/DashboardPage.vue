<template>
  <section :class="style.dashboardPage">
    <SectionHeader header="Главная" paragraph="Вся основная информация на дашборде" />
    <section :class="style.content">
      <div :class="style.scheduler">
        <ScheduleDay :lessons="lessons">
          <template #header>
            <div :class="style.headerScheduler">
              <PhCalendar size="24" />
              <h4>Расписание на сегодня</h4>
            </div>
          </template>
        </ScheduleDay>
      </div>
      <aside :class="style.infoPart">
        <NewsComponent />
        <ChatComponent />
      </aside>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './DashboardPage.module.scss'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import ScheduleDay from '@/components/ScheduleDay/ScheduleDay.vue'
import type { ScheduleDayItem } from '@/components/ScheduleDay/schedule-day.types'
import { PhCalendar } from '@phosphor-icons/vue'
import NewsComponent from '@/pages/dashboard/components/NewsComponent/NewsComponent.vue'
import ChatComponent from './components/ChatComponent/ChatComponent.vue'

export default defineComponent({
  name: 'DashboardPage',
  components: { ChatComponent, NewsComponent, PhCalendar, ScheduleDay, SectionHeader },
  data() {
    return {
      style,
      lessons: [
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
        {
          id: 'programming-1',
          start: '10:10',
          end: '11:40',
          subject: 'Программирование',
          type: 'practice',
          typeLabel: 'Практика',
          location: 'Корпус Б, ауд. 415',
          teacher: 'Петров Алексей Викторович',
        },
        {
          id: 'physics-1',
          start: '12:00',
          end: '13:30',
          subject: 'Физика',
          type: 'lab',
          typeLabel: 'Лабораторная',
          location: 'Корпус А, ауд. 201',
          teacher: 'Смирнов Дмитрий Андреевич',
        },
        {
          id: 'english-1',
          start: '14:00',
          end: '15:30',
          subject: 'Английский язык',
          type: 'practice',
          typeLabel: 'Практика',
          location: 'Корпус Б, ауд. 103',
          teacher: 'Кузнецова Мария Олеговна',
        },
        {
          id: 'databases-1',
          start: '16:00',
          end: '17:30',
          subject: 'Базы данных',
          type: 'lecture',
          typeLabel: 'Лекция',
          location: 'Онлайн',
          teacher: 'Соколов Андрей Николаевич',
        },
        {
          id: 'web-1',
          start: '18:00',
          end: '19:30',
          subject: 'Веб-разработка',
          type: 'lab',
          typeLabel: 'Лабораторная',
          location: 'Корпус А, ауд. 305',
          teacher: 'Васильева Ольга Павловна',
        },
      ] satisfies ScheduleDayItem[],
    }
  },
})
</script>
