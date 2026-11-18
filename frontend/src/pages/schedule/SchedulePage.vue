<template>
  <section :class="style.schedulePage">
    <SectionHeader
      header="Расписание"
      :paragraph="
        $scheduleStore.groupId
          ? `Расписание группы ${$scheduleStore.groups.find((group) => String(group.id) === $scheduleStore.groupId)?.name ?? ''}`
          : auth.user?.student
            ? `Расписание группы ${auth.user.student.group_name}`
            : 'Личное расписание'
      "
    >
      <template #actions>
        <SchedulePeriodPicker location="header" />
      </template>
    </SectionHeader>
    <ScheduleControls />
    <p v-if="$scheduleStore.loading">Загрузка расписания…</p>
    <p v-else-if="$scheduleStore.error" role="alert">
      {{ $scheduleStore.error }} <button @click="$scheduleStore.fetchSchedule()">Повторить</button>
    </p>
    <MobileSchedule v-if="!$scheduleStore.loading && !$scheduleStore.error" />
    <ScheduleTable v-if="!$scheduleStore.loading && !$scheduleStore.error" />
    <ScheduleLegend />
  </section>
</template>

<script lang="ts">
import { useAuthStore } from '@/stores/auth'
import { defineComponent } from 'vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import ScheduleControls from './components/ScheduleControls/ScheduleControls.vue'
import MobileSchedule from './components/MobileSchedule/MobileSchedule.vue'
import ScheduleTable from './components/ScheduleTable/ScheduleTable.vue'
import ScheduleLegend from './components/ScheduleLegend/ScheduleLegend.vue'
import style from './SchedulePage.module.scss'
import SchedulePeriodPicker from './components/SchedulePeriodPicker/SchedulePeriodPicker.vue'

export default defineComponent({
  name: 'SchedulePage',
  components: {
    SchedulePeriodPicker,
    SectionHeader,
    ScheduleControls,
    MobileSchedule,
    ScheduleTable,
    ScheduleLegend,
  },
  created() {
    void this.$scheduleStore.fetchSchedule()
  },
  data() {
    return { style, auth: useAuthStore() }
  },
})
</script>
