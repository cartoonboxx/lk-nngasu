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
  <Modal
    v-if="isMobile"
    v-model="mobileOpen"
    :style="{ '--modal-content-padding': '0px' }"
    @close="restoreMenuFocus"
  >
    <section :class="style.moreSheet">
      <div :class="style.sheetHeader">
        <h2>Ещё</h2>
        <IconButton type="button" variant="ghost" round @click="closeMobileMenu()">
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
      <StudentCard name="Анна Иванова" role="Студент" @click="navigate('settings')" />
    </section>
  </Modal>
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
import breakpoints from '@/styles/breakpoints.module.scss'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import NavigationItem from '@/components/UI/navigation/NavigationItem/NavigationItem.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'

export default defineComponent({
  name: 'MobileNavBar',
  emits: ['mobile-menu-change'],
  components: {
    Modal,
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
      returnFocusOnClose: true,
    }
  },
  mounted() {
    this.mediaQuery = window.matchMedia(`(max-width: ${breakpoints.phoneMax})`)
    this.syncViewport()
    this.mediaQuery.addEventListener('change', this.syncViewport)
  },
  beforeUnmount() {
    this.mediaQuery?.removeEventListener('change', this.syncViewport)
    if (this.mobileOpen) {
      this.$emit('mobile-menu-change', false)
    }
  },
  watch: {
    mobileOpen(value: boolean) {
      this.$emit('mobile-menu-change', value)
    },
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
      this.returnFocusOnClose = true
      this.mobileOpen = true
    },
    closeMobileMenu(restoreFocus = true) {
      if (!this.mobileOpen) return
      this.returnFocusOnClose = restoreFocus
      this.mobileOpen = false
    },
    restoreMenuFocus() {
      if (this.isMobile && this.returnFocusOnClose) {
        ;(this.$refs.menuButton as InstanceType<typeof NavigationItem>)?.focus()
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
