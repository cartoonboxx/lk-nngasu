<template>
  <nav :class="style.navigation" aria-label="Основная навигация">
    <div :class="style.upper">
      <div :class="style.header">
        <BrandHeader compact />
      </div>

      <div :class="style.tabList">
        <NavTab
          text="Главная"
          :icon="icons.house"
          :active="$route.name === 'dashboard'"
          @select="navigate('dashboard')"
        />
        <NavTab
          text="Расписание"
          :icon="icons.calendar"
          :active="$route.name === 'schedule'"
          @select="navigate('schedule')"
        />
        <NavTab
          text="Предметы"
          :icon="icons.book"
          :active="isSubjectsActive"
          @select="navigate('subjects')"
        />
        <NavTab
          text="Оценки"
          :icon="icons.chart"
          :active="$route.name === 'grades'"
          @select="navigate('grades')"
        />
        <NavTab
          text="Сообщения"
          :icon="icons.chats"
          :active="$route.name === 'messages'"
          @select="navigate('messages')"
        >
          <template #actions>
            <span :class="style.tabBadge" aria-label="2 новых">2</span>
          </template>
        </NavTab>
        <NavTab
          text="Ресурсы и сервисы"
          :icon="icons.squares"
          :active="$route.name === 'resources'"
          @select="navigate('resources')"
        />
        <NavTab
          text="Настройки"
          :icon="icons.gear"
          :active="$route.name === 'settings'"
          @select="navigate('settings')"
        />
      </div>
    </div>

    <StudentCard name="Анна Иванова" role="Студентка" />
  </nav>
</template>

<script lang="ts">
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChartBar,
  PhChatsCircle,
  PhGearSix,
  PhHouse,
  PhSquaresFour,
} from '@phosphor-icons/vue'
import { defineComponent, markRaw } from 'vue'

import BrandHeader from '@/components/BrandHeader/BrandHeader.vue'
import StudentCard from '@/components/navigation/StudentCard/StudentCard.vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
import style from './NavBar.module.scss'

export default defineComponent({
  name: 'NavBar',
  components: {
    BrandHeader,
    StudentCard,
    NavTab,
  },
  data() {
    return {
      style,
    }
  },
  methods: {
    navigate(routeName: string) {
      if (this.$route.name !== routeName) {
        void this.$router.push({ name: routeName })
      }
    },
  },
  computed: {
    icons() {
      return {
        house: markRaw(PhHouse),
        calendar: markRaw(PhCalendarBlank),
        book: markRaw(PhBookOpen),
        chart: markRaw(PhChartBar),
        chats: markRaw(PhChatsCircle),
        squares: markRaw(PhSquaresFour),
        gear: markRaw(PhGearSix),
      }
    },
    isSubjectsActive() {
      return this.$route.name === 'subjects' || this.$route.name === 'course'
    },
  },
})
</script>
