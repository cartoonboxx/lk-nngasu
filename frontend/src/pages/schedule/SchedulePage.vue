<template>
  <section :class="style.schedulePage">
    <SectionHeader header="Расписание" paragraph="Расписание группы ПРИ23/1">
      <template #actions>
        <SchedulePeriodPicker location="header" />
      </template>
    </SectionHeader>
    <ScheduleControls />
    <MobileSchedule />
    <ScheduleTable />
    <ScheduleLegend />
  </section>
</template>

<script lang="ts">
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
  data() {
    return { style }
  },
})
</script>
