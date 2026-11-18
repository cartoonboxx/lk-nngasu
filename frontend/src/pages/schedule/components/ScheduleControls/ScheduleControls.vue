<template>
  <div :class="style.filterBlock">
    <div :class="style.leftPart">
      <div :class="style.dailyPart">
        <IconButton variant="ghost" round @click="$scheduleStore.shiftWeek(-1)">
          <PhCaretLeft />
        </IconButton>
        <span :class="style.weekLabel">{{ weekLabel }}</span>
        <IconButton variant="ghost" round @click="$scheduleStore.shiftWeek(1)">
          <PhCaretRight />
        </IconButton>
        <SchedulePeriodPicker />
      </div>
    </div>
    <div :class="style.inputGroup">
      <Select
        v-model="$scheduleStore.groupId"
        :options="groupOptions"
        :class="style.input"
        @update:model-value="$scheduleStore.fetchSchedule()"
      />
      <p v-if="$scheduleStore.groupsError">
        <button @click="$scheduleStore.fetchGroups()">Повторить</button>
      </p>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './ScheduleControls.module.scss'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import { PhCaretLeft, PhCaretRight } from '@phosphor-icons/vue'
import SchedulePeriodPicker from '../SchedulePeriodPicker/SchedulePeriodPicker.vue'

export default defineComponent({
  name: 'ScheduleControls',
  components: {
    SchedulePeriodPicker,
    IconButton,
    Select,
    PhCaretLeft,
    PhCaretRight,
  },
  async created() {
    await this.$scheduleStore.fetchGroups()
  },
  computed: {
    groupOptions() {
      return [
        { label: 'Личное расписание', value: '' },
        ...this.$scheduleStore.groups.map((group) => ({
          label: group.name,
          value: String(group.id),
        })),
      ]
    },
    weekLabel() {
      const start = this.$scheduleStore.weekStart
      const end = new Date(start)
      end.setDate(end.getDate() + 6)
      return new Intl.DateTimeFormat('ru-RU', {
        day: 'numeric',
        month: 'long',
        year: 'numeric',
      }).formatRange(start, end)
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
