<template>
  <section :class="style.dashboardPage">
    <SectionHeader header="Главная" paragraph="Расписание, общение и события университета" />
    <section :class="style.content">
      <div :class="style.scheduler">
        <p v-if="$scheduleStore.loading">Загрузка расписания…</p>
        <ScheduleDay
          v-else-if="!$scheduleStore.error"
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
    this.$scheduleStore.groupId = ''
    const today = new Date()
    await this.$scheduleStore.fetchSchedule(today, today)
  },
  computed: {
    lessons() {
      return this.$scheduleStore.days.flatMap((day) => day.lessons)
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
