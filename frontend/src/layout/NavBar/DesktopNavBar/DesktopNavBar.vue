<template>
  <nav :class="[style.navigation, collapsed ? style.collapsed : null]">
    <div :class="style.upper">
      <div :class="style.header">
        <BrandHeader v-show="!collapsed" compact @click="navigate('dashboard')" />
        <Tooltip :text="collapsed ? 'Развернуть' : 'Свернуть'" position="right">
          <IconButton
            :class="[style.toggle, collapsed && style.compactToggle]"
            variant="ghost"
            @click="toggleCollapsed"
          >
            <img v-if="collapsed" :class="style.toggleLogo" src="$images/logo.svg" alt="" />
            <PhSquareHalf :class="style.toggleIcon" />
          </IconButton>
        </Tooltip>
      </div>

      <div :class="style.tabList">
        <Tooltip :class="style.tabTooltip" text="Главная" position="right" :isShow="collapsed">
          <NavTab
            :compact="collapsed"
            text="Главная"
            :icon="icons.house"
            :active="$route.name === 'dashboard'"
            @select="navigate('dashboard')"
          />
        </Tooltip>
        <Tooltip :class="style.tabTooltip" text="Расписание" position="right" :isShow="collapsed">
          <NavTab
            :compact="collapsed"
            text="Расписание"
            :icon="icons.calendar"
            :active="$route.name === 'schedule'"
            @select="navigate('schedule')"
          />
        </Tooltip>
        <Tooltip :class="style.tabTooltip" text="Предметы" position="right" :isShow="collapsed">
          <NavTab
            :compact="collapsed"
            text="Предметы"
            :icon="icons.book"
            :active="isSubjectsActive"
            @select="navigate('subjects')"
          />
        </Tooltip>
        <Tooltip :class="style.tabTooltip" text="Оценки" position="right" :isShow="collapsed">
          <NavTab
            :compact="collapsed"
            text="Оценки"
            :icon="icons.chart"
            :active="$route.name === 'grades'"
            @select="navigate('grades')"
          />
        </Tooltip>
        <Tooltip :class="style.tabTooltip" text="Сообщения" position="right" :isShow="collapsed">
          <NavTab
            :compact="collapsed"
            text="Сообщения"
            :icon="icons.chats"
            :active="$route.name === 'messages'"
            @select="navigate('messages')"
          >
            <template v-if="!collapsed" #actions>
              <span :class="style.tabBadge">2</span>
            </template>
          </NavTab>
        </Tooltip>
        <Tooltip
          :class="style.tabTooltip"
          text="Ресурсы и сервисы"
          position="right"
          :isShow="collapsed"
        >
          <NavTab
            :compact="collapsed"
            text="Ресурсы и сервисы"
            :icon="icons.squares"
            :active="$route.name === 'resources'"
            @select="navigate('resources')"
          />
        </Tooltip>
        <Tooltip :class="style.tabTooltip" text="Настройки" position="right" :isShow="collapsed">
          <NavTab
            :compact="collapsed"
            text="Настройки"
            :icon="icons.gear"
            :active="$route.name === 'settings'"
            @select="navigate('settings')"
          />
        </Tooltip>
      </div>
    </div>

    <Tooltip :class="style.tabTooltip" :isShow="collapsed" text="Аккаунт" position="right">
      <StudentCard
        :compact="collapsed"
        name="Анна Иванова"
        role="Студентка"
        @click="navigate('settings')"
      />
    </Tooltip>
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
  PhSquareHalf,
  PhSquaresFour,
} from '@phosphor-icons/vue'
import { defineComponent, markRaw } from 'vue'

import BrandHeader from '@/components/BrandHeader/BrandHeader.vue'
import StudentCard from '@/components/navigation/StudentCard/StudentCard.vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'
import style from './DesktopNavBar.module.scss'

export default defineComponent({
  name: 'DesktopNavBar',
  components: { BrandHeader, StudentCard, NavTab, IconButton, Tooltip, PhSquareHalf },
  emits: ['collapse-change'],
  data() {
    return { style, collapsed: false }
  },
  methods: {
    toggleCollapsed() {
      this.collapsed = !this.collapsed
      this.$emit('collapse-change', this.collapsed)
    },
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
