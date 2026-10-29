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
      <Input placeholder="Введите группу" :class="style.input">
        <template #prefix>
          <PhMagnifyingGlass />
        </template>
      </Input>
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
    PhCaretLeft,
    PhCaretRight,
    PhMagnifyingGlass,
  },
  computed: {
    weekLabel() {
      const start = this.$scheduleStore.weekStart
      const end = new Date(start)
      end.setDate(end.getDate() + 5)
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
