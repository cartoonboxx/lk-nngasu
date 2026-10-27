<template>
  <div :class="style.filterBlock">
    <div :class="style.leftPart">
      <div :class="style.dailyPart">
        <IconButton variant="ghost" round @click="shiftWeek(-1)">
          <PhCaretLeft />
        </IconButton>
        <span :class="style.weekLabel">{{ weekLabel }}</span>
        <IconButton variant="ghost" round @click="shiftWeek(1)">
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
  <IconButton
    variant="primary"
    round
    :class="[style.buttonUp, isScrolledPage ? style.isVisible : null]"
    @click="scrollToStart"
  >
    <PhCaretUp />
  </IconButton>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './ScheduleControls.module.scss'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Input from '@/components/UI/inputs/Input/Input.vue'
import { PhCaretLeft, PhCaretRight, PhCaretUp, PhMagnifyingGlass } from '@phosphor-icons/vue'
import { useSchedule } from '../../useSchedule'
import SchedulePeriodPicker from '../SchedulePeriodPicker/SchedulePeriodPicker.vue'

export default defineComponent({
  name: 'ScheduleControls',
  components: {
    PhCaretUp,
    SchedulePeriodPicker,
    IconButton,
    Input,
    PhCaretLeft,
    PhCaretRight,
    PhMagnifyingGlass,
  },
  data() {
    const { weekLabel, shiftWeek, resetWeek } = useSchedule()
    return {
      style,
      weekLabel,
      shiftWeek,
      resetWeek,
      isScrolledPage: false,
    }
  },
  mounted() {
    document.addEventListener('scroll', this.isScrolling)
  },
  beforeUnmount() {
    document.removeEventListener('scroll', this.isScrolling)
  },
  methods: {
    isScrolling() {
      if (window.scrollY > 300) {
        this.isScrolledPage = true
        return
      }
      this.isScrolledPage = false
    },
    scrollToStart() {
      window.scrollTo({ top: 0, left: 0, behavior: 'smooth' })
    },
  },
})
</script>
