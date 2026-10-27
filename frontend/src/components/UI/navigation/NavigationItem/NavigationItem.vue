<template>
  <component
    v-press
    :is="to ? 'RouterLink' : 'button'"
    :to="to"
    :type="to ? undefined : 'button'"
    :class="[style.item, style[variant], active ? style.active : null]"
  >
    <span :class="style.icon">
      <component
        :is="icon"
        :size="24"
        :weight="active && variant === 'bottom' ? 'fill' : 'regular'"
      />
      <span v-if="notification" :class="style.notification" />
    </span>
    <span>{{ text }}</span>
    <span v-if="badge !== undefined" :class="style.badge">{{ badge }}</span>
    <PhCaretRight v-else-if="variant === 'list'" :class="style.arrow" :size="18" />
  </component>
</template>

<script lang="ts">
import { pressFeedback } from '@/directives/pressFeedback'
import { defineComponent, type Component, type PropType } from 'vue'
import { RouterLink, type RouteLocationRaw } from 'vue-router'
import { PhCaretRight } from '@phosphor-icons/vue'
import style from './NavigationItem.module.scss'

export default defineComponent({
  directives: { press: pressFeedback },
  name: 'NavigationItem',
  components: { RouterLink, PhCaretRight },
  props: {
    text: { type: String, required: true },
    icon: { type: [Object, Function] as PropType<Component>, required: true },
    to: { type: [String, Object] as PropType<RouteLocationRaw>, default: undefined },
    variant: { type: String as PropType<'bottom' | 'list'>, default: 'list' },
    active: Boolean,
    notification: Boolean,
    badge: { type: Number, default: undefined },
  },
  data() {
    return { style }
  },
  methods: {
    focus() {
      ;(this.$el as HTMLElement).focus()
    },
  },
})
</script>
