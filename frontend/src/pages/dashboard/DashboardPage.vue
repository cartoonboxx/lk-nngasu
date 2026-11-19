<template>
  <section :class="style.dashboardPage">
    <SectionHeader header="Главная" paragraph="Расписание, общение и события университета" />
    <p v-if="cabinet.dashboardError">
      {{ cabinet.dashboardError }} <button @click="cabinet.fetchDashboard()">Повторить</button>
    </p>
    <section :class="style.content">
      <div :class="style.scheduler">
        <p v-if="cabinet.dashboardLoading">Загрузка расписания…</p>
        <ScheduleDay
          v-else-if="!cabinet.dashboardError"
          :class="style.scheduleCard"
          :lessons="lessons"
          layout="list"
        >
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
import { useCabinetStore } from '@/stores/cabinet'
import type { ScheduleLesson, LessonType } from '@/pages/schedule/schedule.types'
import style from './DashboardPage.module.scss'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import ScheduleDay from '@/components/ScheduleDay/ScheduleDay.vue'
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
  async created() {
    await this.cabinet.fetchDashboard()
  },
  computed: {
    cabinet() {
      return useCabinetStore()
    },
    lessons(): ScheduleLesson[] {
      return (this.cabinet.dashboard?.lessons ?? []).map((lesson) => ({
        id: lesson.id,
        start: lesson.starts_at.slice(0, 5),
        end: lesson.ends_at.slice(0, 5),
        subject: lesson.subject,
        type: (lesson.lesson_type === 'laboratory' ? 'lab' : lesson.lesson_type) as LessonType,
        location: lesson.room,
        teacher: lesson.teacher_name,
      }))
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
