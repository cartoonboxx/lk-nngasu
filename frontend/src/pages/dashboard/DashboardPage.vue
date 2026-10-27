<template>
  <section :class="style.dashboardPage">
    <SectionHeader header="Главная" paragraph="Расписание, общение и события университета" />
    <section :class="style.content">
      <div :class="style.scheduler">
        <ScheduleDay :class="style.scheduleCard" :lessons="lessons" layout="list">
          <template #header>
            <div :class="style.headerScheduler">
              <PhCalendar size="24" />
              <div>
                <h2>Расписание на сегодня</h2>
              </div>
            </div>
          </template>
        </ScheduleDay>
      </div>
      <ChatComponent :class="style.chatPanel" />
      <NewsComponent :class="style.newsPanel" />
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
import { pressFeedback } from '@/directives/pressFeedback'
import NewsComponent from '@/pages/dashboard/components/NewsComponent/NewsComponent.vue'
import ChatComponent from './components/ChatComponent/ChatComponent.vue'

export default defineComponent({
  name: 'DashboardPage',
  components: {
    ChatComponent,
    NewsComponent,
    PhCalendar,
    ScheduleDay,
    SectionHeader,
  },
  directives: { press: pressFeedback },
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
