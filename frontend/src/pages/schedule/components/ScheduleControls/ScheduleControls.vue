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
      <label for="schedule-group-search" :class="style.inputLabel">Группа расписания</label>
      <Input
        id="schedule-group-search"
        v-model="groupSearch"
        :options="groupOptions"
        :class="style.input"
        placeholder="Название группы"
        full-width
        @select="
          (option) => {
            if ($scheduleStore.groupId !== option.value) {
              $scheduleStore.groupId = option.value
              $scheduleStore.fetchSchedule()
            }
            groupQuery = null
          }
        "
      >
        <template #prefix><PhMagnifyingGlass :size="20" /></template>
      </Input>
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
import Input from '@/components/UI/inputs/Input/Input.vue'
import { PhCaretLeft, PhCaretRight, PhMagnifyingGlass } from '@phosphor-icons/vue'
import SchedulePeriodPicker from '../SchedulePeriodPicker/SchedulePeriodPicker.vue'

export default defineComponent({
  name: 'ScheduleControls',
  components: {
    SchedulePeriodPicker,
    IconButton,
    Input,
    PhMagnifyingGlass,
    PhCaretLeft,
    PhCaretRight,
  },
  async created() {
    await this.$scheduleStore.fetchGroups()
  },
  computed: {
    groupSearch: {
      get(): string {
        return (
          this.groupQuery ??
          this.groupOptions.find((option) => option.value === this.$scheduleStore.groupId)?.label ??
          ''
        )
      },
      set(value: string) {
        this.groupQuery = value
      },
    },
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
      groupQuery: null as string | null,
    }
  },
})
</script>
