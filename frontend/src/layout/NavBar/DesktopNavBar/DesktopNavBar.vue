<template>
  <nav :class="[style.navigation, collapsed ? style.collapsed : null]">
    <div :class="style.upper">
      <div :class="style.header">
        <BrandHeader v-show="!collapsed" compact @click="navigate('dashboard')" />
        <Tooltip :text="collapsed ? 'Развернуть' : 'Свернуть'" position="right">
          <IconButton
            :class="[style.toggle, collapsed && style.compactToggle]"
            variant="ghost"
            @click="
              () => {
                collapsed = !collapsed
                $emit('collapse-change', collapsed)
              }
            "
          >
            <img v-if="collapsed" :class="style.toggleLogo" src="$images/logo.svg" alt="" />
            <PhSidebarSimple :class="style.toggleIcon" />
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
            :to="{ name: 'messages' }"
          >
            <template #trailing>
              <Badge
                v-if="$messageStore.unreadCount"
                :class="style.unreadBadge"
                :value="$messageStore.unreadCount"
                inline
              />
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
        <Tooltip
          v-if="auth.user?.is_admin"
          :class="style.tabTooltip"
          text="Администрирование"
          position="right"
          :is-show="collapsed"
        >
          <NavTab
            :compact="collapsed"
            text="Администрирование"
            :icon="icons.admin"
            :to="{ name: 'admin-requests' }"
          />
        </Tooltip>
      </div>
    </div>

    <Tooltip :class="style.tabTooltip" :isShow="collapsed" text="Аккаунт" position="right">
      <StudentCard
        :image-src="auth.avatarUrl || undefined"
        :compact="collapsed"
        :name="auth.user?.full_name || 'Аккаунт'"
        :role="auth.roleLabel"
        @click="navigate('settings')"
      />
    </Tooltip>
  </nav>
</template>

<script lang="ts">
import { useAuthStore } from '@/stores/auth'
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChartBar,
  PhChatsCircle,
  PhGearSix,
  PhHouse,
  PhSidebarSimple,
  PhSquaresFour,
  PhShieldCheck,
} from '@phosphor-icons/vue'
import { defineComponent, markRaw } from 'vue'

import BrandHeader from '@/components/BrandHeader/BrandHeader.vue'
import StudentCard from '@/components/navigation/StudentCard/StudentCard.vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'
import Badge from '@/components/UI/Badge/Badge.vue'
import style from './DesktopNavBar.module.scss'

export default defineComponent({
  name: 'DesktopNavBar',
  components: { PhSidebarSimple, BrandHeader, StudentCard, NavTab, IconButton, Tooltip, Badge },
  emits: ['collapse-change'],
  data() {
    return { style, collapsed: false }
  },
  methods: {
    async navigate(routeName: string) {
      if (this.$route.name !== routeName) {
        await this.$router.push({ name: routeName })
      }
    },
  },
  computed: {
    auth() {
      return useAuthStore()
    },

    icons() {
      return {
        house: markRaw(PhHouse),
        calendar: markRaw(PhCalendarBlank),
        book: markRaw(PhBookOpen),
        chart: markRaw(PhChartBar),
        chats: markRaw(PhChatsCircle),
        squares: markRaw(PhSquaresFour),
        gear: markRaw(PhGearSix),
        admin: markRaw(PhShieldCheck),
      }
    },
    isSubjectsActive() {
      return this.$route.name === 'subjects' || this.$route.name === 'course'
    },
  },
})
</script>
