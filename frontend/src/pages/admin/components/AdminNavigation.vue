<template>
  <nav :class="[desktop.navigation, collapsed && desktop.collapsed]">
    <div :class="desktop.upper">
      <div :class="desktop.header">
        <RouterLink v-show="!collapsed" to="/admin">
          <BrandHeader compact />
        </RouterLink>
        <Tooltip :text="collapsed ? 'Развернуть' : 'Свернуть'" position="right">
          <IconButton
            :class="[desktop.toggle, collapsed && desktop.compactToggle]"
            variant="ghost"
            :title="collapsed ? 'Развернуть навигацию' : 'Свернуть навигацию'"
            @click="
              () => {
                collapsed = !collapsed
                $emit('collapse-change', collapsed)
              }
            "
          >
            <img v-if="collapsed" :class="desktop.toggleLogo" src="$images/logo.svg" alt="" />
            <PhSidebarSimple :class="desktop.toggleIcon" />
          </IconButton>
        </Tooltip>
      </div>
      <div :class="desktop.tabList">
        <Tooltip
          v-for="item in navigation"
          :key="item.path"
          :class="desktop.tabTooltip"
          :text="item.label"
          position="right"
          :is-show="collapsed"
        >
          <NavTab
            :to="`/admin/${item.path}`"
            :text="item.label"
            :icon="item.icon"
            :compact="collapsed"
            :active="$route.name === `admin-${item.path}`"
          >
            <template v-if="item.path === 'requests' && pending && !collapsed" #actions>
              <Badge :value="pending" variant="info" inline />
            </template>
          </NavTab>
        </Tooltip>
      </div>
    </div>
    <Tooltip
      :class="desktop.tabTooltip"
      :is-show="collapsed"
      text="Аккаунт администратора"
      position="right"
    >
      <StudentCard
        :compact="collapsed"
        name="Анна Иванова"
        role="Администратор"
        @click="accountOpen = true"
      />
    </Tooltip>
  </nav>
  <nav :class="mobile.bottomNavigation" :inert="menuOpen">
    <NavigationItem
      v-for="item in primaryNavigation"
      :key="item.path"
      :to="`/admin/${item.path}`"
      :text="item.label"
      :icon="item.icon"
      :active="$route.name === `admin-${item.path}`"
      variant="bottom"
    />
    <NavigationItem
      ref="menuButton"
      text="Ещё"
      :icon="moreIcon"
      :active="isMoreActive"
      variant="bottom"
      @click="menuOpen = true"
    />
  </nav>
  <Modal
    v-if="isMobile"
    v-model="menuOpen"
    :class="style.navigationModal"
    @close="
      isMobile && !accountOpen && ($refs.menuButton as { focus: () => void } | undefined)?.focus()
    "
  >
    <section :class="mobile.moreSheet">
      <header :class="mobile.sheetHeader">
        <h2>Администрирование</h2>
        <IconButton round variant="ghost" title="Закрыть меню" @click="menuOpen = false">
          <PhX :size="22" />
        </IconButton>
      </header>
      <nav :class="mobile.moreLinks">
        <NavTab
          v-for="item in moreNavigation"
          :key="item.path"
          :to="`/admin/${item.path}`"
          :text="item.label"
          :icon="item.icon"
          :active="$route.name === `admin-${item.path}`"
          @select="menuOpen = false"
        />
      </nav>
      <StudentCard
        name="Анна Иванова"
        role="Администратор"
        @click="
          () => {
            menuOpen = false
            accountOpen = true
          }
        "
      />
    </section>
  </Modal>
  <AdminDialog v-model="accountOpen" title="Аккаунт администратора">
    <p>Анна Иванова · Администратор</p>
    <p :class="style.muted">
      Открыт локальный демонстрационный сеанс. Он не предоставляет прав в университетских системах.
    </p>
    <div :class="style.dialogFooter">
      <RouterLink to="/authorization" :class="style.link">Выйти из демо-режима</RouterLink>
    </div>
  </AdminDialog>
</template>
<script lang="ts">
import Badge from '@/components/UI/Badge/Badge.vue'
import { defineComponent, markRaw } from 'vue'
import { RouterLink } from 'vue-router'
import {
  PhUsers,
  PhBuildings,
  PhBooks,
  PhCalendarBlank,
  PhNewspaper,
  PhSidebarSimple,
  PhSquaresFour,
  PhX,
  PhEnvelope,
} from '@phosphor-icons/vue'
import BrandHeader from '@/components/BrandHeader/BrandHeader.vue'
import StudentCard from '@/components/navigation/StudentCard/StudentCard.vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
import NavigationItem from '@/components/UI/navigation/NavigationItem/NavigationItem.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'
import Modal from '@/components/UI/modal/BaseModal/Modal.vue'
import AdminDialog from './AdminDialog.vue'
import desktop from '@/layout/NavBar/DesktopNavBar/DesktopNavBar.module.scss'
import mobile from '@/layout/NavBar/MobileNavBar/MobileNavBar.module.scss'
import breakpoints from '@/styles/breakpoints.module.scss'
import style from '../Admin.module.scss'
export default defineComponent({
  components: {
    Badge,
    RouterLink,
    BrandHeader,
    StudentCard,
    NavTab,
    NavigationItem,
    IconButton,
    Tooltip,
    Modal,
    AdminDialog,
    PhSidebarSimple,
    PhX,
  },
  emits: ['collapse-change', 'mobile-menu-change'],
  data() {
    return {
      desktop,
      mobile,
      style,
      collapsed: false,
      menuOpen: false,
      accountOpen: false,
      isMobile: false,
      mediaQuery: null as MediaQueryList | null,
      moreIcon: markRaw(PhSquaresFour),
      navigation: [
        { path: 'requests', label: 'Заявки', icon: markRaw(PhEnvelope) },
        { path: 'users', label: 'Пользователи', icon: markRaw(PhUsers) },
        { path: 'structure', label: 'Учебная структура', icon: markRaw(PhBuildings) },
        { path: 'courses', label: 'Дисциплины и нагрузка', icon: markRaw(PhBooks) },
        { path: 'schedule', label: 'Расписание', icon: markRaw(PhCalendarBlank) },
        { path: 'news', label: 'Новости', icon: markRaw(PhNewspaper) },
      ],
    }
  },
  computed: {
    primaryNavigation() {
      return this.navigation.slice(0, 3)
    },
    moreNavigation() {
      return this.navigation.slice(3)
    },
    isMoreActive(): boolean {
      return (
        this.menuOpen ||
        this.moreNavigation.some((item) => this.$route.name === `admin-${item.path}`)
      )
    },

    pending(): number {
      return this.$adminStore.data?.requests.filter((item) => item.status === 'pending').length ?? 0
    },
  },
  mounted() {
    this.mediaQuery = window.matchMedia(`(max-width: ${breakpoints.phoneMax})`)
    this.isMobile = this.mediaQuery.matches
    this.mediaQuery.addEventListener('change', this.syncViewport)
  },
  beforeUnmount() {
    this.mediaQuery?.removeEventListener('change', this.syncViewport)
  },
  watch: {
    menuOpen(open: boolean) {
      this.$emit('mobile-menu-change', open)
    },
    '$route.path'() {
      this.menuOpen = false
      window.scrollTo({ top: 0, left: 0 })
    },
  },
  methods: {
    syncViewport() {
      this.isMobile = this.mediaQuery?.matches ?? false
      if (!this.isMobile) this.menuOpen = false
    },
  },
})
</script>
