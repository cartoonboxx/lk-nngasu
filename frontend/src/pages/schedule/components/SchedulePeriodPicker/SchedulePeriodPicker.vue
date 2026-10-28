<template>
  <div :class="[style.anchor, location === 'header' ? style.mobileTrigger : style.desktopTrigger]">
    <Tooltip text="Выбрать период" :disabled="opened">
      <IconButton
        :class="[style.trigger, opened ? style.active : null]"
        variant="ghost"
        round
        :size="location === 'header' ? 'lg' : 'md'"
        @click="opened ? close() : open()"
      >
        <PhCalendar />
      </IconButton>
    </Tooltip>
    <Modal
      v-model="opened"
      :anchor="getAnchor"
      :width="304"
      :style="{ '--modal-content-padding': '0px' }"
      @close="restoreTrigger"
    >
      <div ref="calendar" :class="style.content">
        <header :class="style.header">
          <h2>Выбор даты</h2>
          <IconButton size="sm" variant="ghost" round @click="close()"><PhX /></IconButton>
        </header>
        <div :class="style.selectionHeader">
          <span>Выберите неделю</span>
          <Button size="sm" variant="primary" round @click="selectToday">Сегодня</Button>
        </div>
        <div :class="style.monthNavigation">
          <IconButton size="sm" variant="ghost" round @click="shiftMonth(-1)">
            <PhCaretLeft />
          </IconButton>
          <span>{{ monthLabel }}</span>
          <IconButton size="sm" variant="ghost" round @click="shiftMonth(1)">
            <PhCaretRight />
          </IconButton>
        </div>
        <table :class="style.calendar">
          <thead>
            <tr>
              <th v-for="day in weekdays" :key="day" scope="col">{{ day }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(week, index) in calendarWeeks" :key="index">
              <td
                v-for="day in week"
                :key="day.key"
                :class="[
                  day.selected ? style.range : null,
                  day.rangeStart ? style.rangeStart : null,
                  day.rangeEnd ? style.rangeEnd : null,
                ]"
              >
                <button
                  type="button"
                  :class="[
                    style.day,
                    !day.inMonth ? style.outsideMonth : null,
                    day.rangeStart ? style.selectedDay : null,
                    day.today ? style.todayDay : null,
                  ]"
                  @click="selectDate(day.date)"
                  @keydown="onDateKeydown($event, day.date)"
                >
                  {{ day.date.getDate() }}
                </button>
              </td>
            </tr>
          </tbody>
        </table>
        <div :class="style.actions">
          <Button variant="primary" fullWidth @click="apply">Применить</Button>
          <Button variant="secondary" fullWidth @click="close()">Отменить</Button>
        </div>
      </div>
    </Modal>
  </div>
</template>

<script lang="ts">
import type { SchedulePeriodPickerLocation } from '../../schedule.types'
import { defineComponent, type PropType } from 'vue'
import { PhCalendar, PhCaretLeft, PhCaretRight, PhX } from '@phosphor-icons/vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import { useSchedule } from '../../useSchedule'
import style from './SchedulePeriodPicker.module.scss'

function dateKey(date: Date) {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

function weekStart(date: Date) {
  const start = new Date(date)
  start.setHours(0, 0, 0, 0)
  start.setDate(start.getDate() - ((start.getDay() + 6) % 7))
  return start
}

export default defineComponent({
  name: 'SchedulePeriodPicker',
  props: {
    location: { type: String as PropType<SchedulePeriodPickerLocation>, default: 'controls' },
  },
  components: { Modal, IconButton, Button, Tooltip, PhCalendar, PhCaretLeft, PhCaretRight, PhX },
  data() {
    const { weekStart: activeWeek, selectWeek } = useSchedule()
    return {
      style,
      activeWeek,
      selectWeek,
      weekdays: ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс'],
      opened: false,
      selectedWeek: new Date(),
      displayedMonth: new Date(),
      focusedDate: '',
    }
  },
  computed: {
    monthLabel(): string {
      return this.displayedMonth
        .toLocaleDateString('ru-RU', { month: 'long', year: 'numeric' })
        .replace(/^./, (letter) => letter.toUpperCase())
    },
    calendarWeeks() {
      const first = new Date(this.displayedMonth.getFullYear(), this.displayedMonth.getMonth(), 1)
      const start = weekStart(first)
      const last = new Date(first.getFullYear(), first.getMonth() + 1, 0)
      const rowCount = Math.ceil((((first.getDay() + 6) % 7) + last.getDate()) / 7)
      const end = new Date(this.selectedWeek)
      end.setDate(end.getDate() + 6)
      const today = dateKey(new Date())
      return Array.from({ length: rowCount }, (_, row) =>
        Array.from({ length: 7 }, (_, column) => {
          const date = new Date(start)
          date.setDate(date.getDate() + row * 7 + column)
          const key = dateKey(date)
          return {
            date,
            key,
            inMonth: date.getMonth() === this.displayedMonth.getMonth(),
            today: key === today,
            selected: date >= this.selectedWeek && date <= end,
            rangeStart: key === dateKey(this.selectedWeek),
            rangeEnd: key === dateKey(end),
          }
        }),
      )
    },
  },
  methods: {
    getAnchor(): HTMLElement | null {
      const trigger = Array.from(
        document.querySelectorAll<HTMLButtonElement>(`.${style.trigger}`),
      ).find((button) => button.getClientRects().length)
      return trigger?.closest<HTMLElement>(`.${style.anchor}`) ?? null
    },
    async open() {
      this.selectedWeek = new Date(this.activeWeek)
      this.displayedMonth = new Date(this.activeWeek.getFullYear(), this.activeWeek.getMonth(), 1)
      this.focusedDate = dateKey(this.selectedWeek)
      this.opened = true
      await this.$nextTick()
      this.focusDay()
    },
    close() {
      this.opened = false
    },
    restoreTrigger() {
      const trigger = Array.from(
        document.querySelectorAll<HTMLButtonElement>(`.${style.trigger}`),
      ).find((button) => button.getClientRects().length)
      trigger?.focus({ preventScroll: true })
    },
    shiftMonth(direction: number) {
      const date = new Date(this.displayedMonth)
      date.setMonth(date.getMonth() + direction, 1)
      this.displayedMonth = date
      this.focusedDate = dateKey(date)
    },
    selectDate(date: Date) {
      this.selectedWeek = weekStart(date)
      this.focusedDate = dateKey(date)
    },
    selectToday() {
      const today = new Date()
      this.selectDate(today)
      this.displayedMonth = new Date(today.getFullYear(), today.getMonth(), 1)
    },
    apply() {
      this.selectWeek(this.selectedWeek)
      this.close()
    },
    focusDay() {
      const index = this.calendarWeeks.flat().findIndex((day) => day.key === this.focusedDate)
      ;(this.$refs.calendar as HTMLElement)
        .querySelectorAll<HTMLButtonElement>('tbody button')
        [index]?.focus({ preventScroll: true })
    },
    async onDateKeydown(event: KeyboardEvent, current: Date) {
      const date = new Date(current)
      const offset = (date.getDay() + 6) % 7
      switch (event.key) {
        case 'ArrowLeft':
          date.setDate(date.getDate() - 1)
          break
        case 'ArrowRight':
          date.setDate(date.getDate() + 1)
          break
        case 'ArrowUp':
          date.setDate(date.getDate() - 7)
          break
        case 'ArrowDown':
          date.setDate(date.getDate() + 7)
          break
        case 'Home':
          date.setDate(date.getDate() - offset)
          break
        case 'End':
          date.setDate(date.getDate() + 6 - offset)
          break
        case 'PageUp':
        case 'PageDown': {
          const day = date.getDate()
          date.setDate(1)
          date.setMonth(date.getMonth() + (event.key === 'PageUp' ? -1 : 1))
          date.setDate(
            Math.min(day, new Date(date.getFullYear(), date.getMonth() + 1, 0).getDate()),
          )
          break
        }
        default:
          return
      }
      event.preventDefault()
      this.focusedDate = dateKey(date)
      this.displayedMonth = new Date(date.getFullYear(), date.getMonth(), 1)
      await this.$nextTick()
      this.focusDay()
    },
  },
})
</script>
