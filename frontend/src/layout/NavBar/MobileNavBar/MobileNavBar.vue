<template>
  <nav :class="style.bottomNavigation" :inert="mobileOpen">
    <NavigationItem
      v-for="tab in mobileTabs"
      :key="tab.route"
      :to="{ name: tab.route }"
      :text="tab.text"
      :icon="tab.icon"
      :active="tab.active"
      variant="bottom"
    />
    <NavigationItem
      ref="menuButton"
      text="Ещё"
      :icon="icons.squares"
      :active="isMoreActive || mobileOpen"
      variant="bottom"
      notification
      @click="openMobileMenu"
    />
  </nav>
  <div v-if="isMobile && mobileOpen" :class="style.backdrop" @click="closeMobileMenu()" />
  <section
    v-if="isMobile && mobileOpen"
    ref="navigation"
    :class="style.moreSheet"
    @keydown="handleMenuKeydown"
  >
    <div :class="style.sheetHandle" />
    <div :class="style.sheetHeader">
      <h2>Ещё</h2>
      <IconButton type="button" round @click="closeMobileMenu()">
        <PhX :size="22" />
      </IconButton>
    </div>
    <div :class="style.moreLinks">
      <NavigationItem
        v-for="item in moreItems"
        :key="item.route"
        :text="item.text"
        :icon="item.icon"
        :active="$route.name === item.route"
        :badge="item.route === 'messages' ? 2 : undefined"
        @click="navigate(item.route)"
      />
    </div>
    <StudentCard name="Анна Иванова" role="Студентка" @click="navigate('settings')" />
  </section>
</template>

<script lang="ts">
import {
  PhBookOpen,
  PhCalendarBlank,
  PhChartBar,
  PhChatsCircle,
  PhGearSix,
  PhHouse,
  PhX,
  PhSquaresFour,
} from '@phosphor-icons/vue'
import { defineComponent, markRaw } from 'vue'

import StudentCard from '@/components/navigation/StudentCard/StudentCard.vue'
import style from './MobileNavBar.module.scss'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import NavigationItem from '@/components/UI/navigation/NavigationItem/NavigationItem.vue'

export default defineComponent({
  name: 'MobileNavBar',
  emits: ['mobile-menu-change'],
  components: {
    IconButton,
    NavigationItem,
    PhX,
    StudentCard,
  },
  data() {
    return {
      style,
      isMobile: false,
      mobileOpen: false,
      mediaQuery: null as MediaQueryList | null,
      previousOverflow: '',
    }
  },
  mounted() {
    this.mediaQuery = window.matchMedia('(max-width: 767px)')
    this.syncViewport()
    this.mediaQuery.addEventListener('change', this.syncViewport)
  },
  beforeUnmount() {
    this.mediaQuery?.removeEventListener('change', this.syncViewport)
    if (this.mobileOpen) {
      document.body.style.overflow = this.previousOverflow
      this.$emit('mobile-menu-change', false)
    }
  },
  watch: {
    '$route.fullPath'() {
      this.closeMobileMenu()
    },
  },
  methods: {
    syncViewport() {
      this.isMobile = this.mediaQuery?.matches ?? false
      if (!this.isMobile) this.closeMobileMenu(false)
    },
    openMobileMenu() {
      if (!this.isMobile || this.mobileOpen) return
      this.previousOverflow = document.body.style.overflow
      document.body.style.overflow = 'hidden'
      this.mobileOpen = true
      this.$emit('mobile-menu-change', true)
      void this.$nextTick(() => {
        ;(this.$refs.navigation as HTMLElement)?.querySelector<HTMLButtonElement>('button')?.focus()
      })
    },
    closeMobileMenu(restoreFocus = true) {
      if (!this.mobileOpen) return
      this.mobileOpen = false
      document.body.style.overflow = this.previousOverflow
      this.$emit('mobile-menu-change', false)
      if (restoreFocus) {
        void this.$nextTick(() =>
          (this.$refs.menuButton as InstanceType<typeof NavigationItem>)?.focus(),
        )
      }
    },
    handleMenuKeydown(event: KeyboardEvent) {
      if (!this.mobileOpen) return
      if (event.key === 'Escape') {
        event.preventDefault()
        this.closeMobileMenu()
      }
      if (event.key !== 'Tab') return
      const buttons = (this.$refs.navigation as HTMLElement).querySelectorAll<HTMLButtonElement>(
        'button:not(:disabled)',
      )
      const first = buttons[0]
      const last = buttons[buttons.length - 1]
      if (event.shiftKey && document.activeElement === first) {
        event.preventDefault()
        last?.focus()
      } else if (!event.shiftKey && document.activeElement === last) {
        event.preventDefault()
        first?.focus()
      }
    },
    navigate(routeName: string) {
      this.closeMobileMenu()
      if (this.$route.name !== routeName) {
        void this.$router.push({ name: routeName })
      }
    },
  },
  computed: {
    mobileTabs() {
      return [
        {
          route: 'dashboard',
          text: 'Главная',
          icon: this.icons.house,
          active: this.$route.name === 'dashboard',
        },
        {
          route: 'schedule',
          text: 'Расписание',
          icon: this.icons.calendar,
          active: this.$route.name === 'schedule',
        },
        {
          route: 'subjects',
          text: 'Предметы',
          icon: this.icons.book,
          active: this.isSubjectsActive,
        },
      ]
    },
    moreItems() {
      return [
        { route: 'grades', text: 'Оценки', icon: this.icons.chart },
        { route: 'messages', text: 'Сообщения', icon: this.icons.chats },
        { route: 'resources', text: 'Ресурсы и сервисы', icon: this.icons.squares },
        { route: 'settings', text: 'Настройки', icon: this.icons.gear },
      ]
    },
    isMoreActive() {
      return this.moreItems.some((item) => item.route === this.$route.name)
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
      }
    },
    isSubjectsActive() {
      return this.$route.name === 'subjects' || this.$route.name === 'course'
    },
  },
})
</script>
