<template>
  <div ref="root" :class="style.root" @keydown.esc="close(true)">
    <slot name="trigger" :trigger-props="triggerProps" :open="open">
      <button v-bind="triggerProps" type="button" :class="style.trigger">{{ triggerText }}</button>
    </slot>
    <div v-if="open" ref="popup" :class="style.popup" popover="manual">
      <div
        ref="list"
        :class="style.list"
        :tabindex="!enabledIndices.length ? 0 : undefined"
        @keydown="onKeydown"
      >
        <component
          :is="isTabs ? 'NavTab' : 'StyleTab'"
          v-for="(item, index) in items"
          :key="item.value"
          :text="item.label"
          :active="item.value === activeValue"
          :disabled="item.disabled"
          :trigger-attrs="{
            'data-popup-item': '',
            tabindex:
              index === (enabledIndices.includes(activeIndex) ? activeIndex : enabledIndices[0])
                ? 0
                : -1,
          }"
          @select="select(item)"
        >
          <slot name="item" :item="item" :active="item.value === activeValue">{{
            item.label
          }}</slot>
        </component>
        <p v-if="!items.length" :class="style.empty">
          <slot name="empty">Нет доступных вариантов</slot>
        </p>
      </div>
      <template v-if="isTabs">
        <template v-for="item in items" :key="item.value">
          <div
            v-if="$slots[item.value]"
            v-show="item.value === activeValue"
            :class="style.panel"
            tabindex="0"
          >
            <slot :name="item.value" :item="item" />
          </div>
        </template>
      </template>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import style from './FloatingPopup.module.scss'

export default defineComponent({
  name: 'FloatingPopup',
  components: { NavTab, StyleTab },
  props: {
    items: { type: Array as PropType<FloatingPopupItem[]>, required: true },
    modelValue: { type: String, default: undefined },
    triggerText: { type: String, default: 'Вкладки' },
    disabled: Boolean,
    closeOnSelect: Boolean,
    mode: { type: String as PropType<FloatingPopupMode>, default: 'tabs' },
  },
  emits: { 'update:modelValue': (value: string) => typeof value === 'string' },
  data() {
    return {
      style,
      open: false,
      internalValue: undefined as string | undefined,
      resizeObserver: null as ResizeObserver | null,
    }
  },
  computed: {
    isTabs(): boolean {
      return this.mode === 'tabs'
    },
    activeValue(): string | undefined {
      const value = this.modelValue ?? this.internalValue
      if (!this.isTabs) return value
      return (
        this.items.find((item) => item.value === value && !item.disabled)?.value ??
        this.items.find((item) => !item.disabled)?.value
      )
    },
    activeIndex(): number {
      return this.items.findIndex((item) => item.value === this.activeValue)
    },
    enabledIndices(): number[] {
      return this.items.flatMap((item, index) => (item.disabled ? [] : [index]))
    },
    triggerProps() {
      return {
        disabled: this.disabled,
        onClick: this.toggle,
        onKeydown: this.onTriggerKeydown,
      }
    },
  },
  watch: {
    disabled(value: boolean) {
      if (value) this.close()
    },
    items: {
      deep: true,
      handler() {
        void this.$nextTick(this.updatePosition)
      },
    },
  },
  mounted() {
    document.addEventListener('pointerdown', this.onPointerdown)
    document.addEventListener('focusin', this.onFocusin)
    window.addEventListener('scroll', this.updatePosition, true)
    window.addEventListener('resize', this.updatePosition)
    window.visualViewport?.addEventListener('resize', this.updatePosition)
    window.visualViewport?.addEventListener('scroll', this.updatePosition)
    this.resizeObserver = new ResizeObserver(this.updatePosition)
    this.resizeObserver.observe(this.$refs.root as HTMLElement)
  },
  beforeUnmount() {
    document.removeEventListener('pointerdown', this.onPointerdown)
    document.removeEventListener('focusin', this.onFocusin)
    window.removeEventListener('scroll', this.updatePosition, true)
    window.removeEventListener('resize', this.updatePosition)
    window.visualViewport?.removeEventListener('resize', this.updatePosition)
    window.visualViewport?.removeEventListener('scroll', this.updatePosition)
    this.resizeObserver?.disconnect()
  },
  methods: {
    updatePosition() {
      const root = this.$refs.root as HTMLElement | undefined
      const popup = this.$refs.popup as HTMLElement | undefined
      const list = this.$refs.list as HTMLElement | undefined
      if (!this.open || !root || !popup || !list) return
      const anchor = root.getBoundingClientRect()
      const viewport = window.visualViewport
      const left = (viewport?.offsetLeft ?? 0) + 12
      const top = (viewport?.offsetTop ?? 0) + 12
      const right = left + (viewport?.width ?? window.innerWidth) - 24
      const bottom = top + (viewport?.height ?? window.innerHeight) - 24
      const width = Math.min(anchor.width, right - left)
      popup.style.width = `${width}px`
      const height = Math.min(list.scrollHeight, 280) + popup.offsetHeight - list.clientHeight
      const below = Math.max(0, bottom - anchor.bottom - 8)
      const above = Math.max(0, anchor.top - top - 8)
      const upwards = below < height && above > below
      const maxHeight = upwards ? above : below
      popup.style.maxHeight = `${maxHeight}px`
      popup.style.left = `${Math.max(left, Math.min(anchor.left, right - width))}px`
      popup.style.top = `${
        upwards ? Math.max(top, anchor.top - 8 - Math.min(height, maxHeight)) : anchor.bottom + 8
      }px`
    },
    focusItem(index: number) {
      ;(this.$refs.list as HTMLElement | undefined)
        ?.querySelectorAll<HTMLButtonElement>('[data-popup-item]')
        [index]?.focus()
    },
    async show(last = false) {
      if (this.disabled) return
      this.open = true
      await this.$nextTick()
      const popup = this.$refs.popup as HTMLElement | undefined
      if (!this.open || !popup) return
      if (!popup.matches(':popover-open')) popup.showPopover()
      this.updatePosition()
      const index =
        this.activeIndex >= 0 && !this.items[this.activeIndex]?.disabled
          ? this.activeIndex
          : last
            ? this.enabledIndices[this.enabledIndices.length - 1]
            : this.enabledIndices[0]
      if (index === undefined) (this.$refs.list as HTMLElement | undefined)?.focus()
      else this.focusItem(index)
    },
    close(restoreFocus = false) {
      const popup = this.$refs.popup as HTMLElement | undefined
      if (popup?.matches(':popover-open')) popup.hidePopover()
      this.open = false
      if (restoreFocus)
        (this.$refs.root as HTMLElement | undefined)
          ?.querySelector<HTMLButtonElement>('button')
          ?.focus()
    },
    toggle() {
      if (this.open) this.close()
      else void this.show()
    },
    onTriggerKeydown(event: KeyboardEvent) {
      if (event.key !== 'ArrowDown' && event.key !== 'ArrowUp') return
      event.preventDefault()
      void this.show(event.key === 'ArrowUp')
    },
    select(item: FloatingPopupItem) {
      if (item.disabled) return
      this.internalValue = item.value
      this.$emit('update:modelValue', item.value)
      if (this.closeOnSelect) this.close(true)
    },
    onKeydown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault()
        event.stopPropagation()
        this.close(true)
        return
      }
      if (!['ArrowDown', 'ArrowUp', 'Home', 'End'].includes(event.key)) return
      const enabled = this.enabledIndices
      if (!enabled.length) return
      event.preventDefault()
      const buttons = Array.from(
        (this.$refs.list as HTMLElement | undefined)?.querySelectorAll('[data-popup-item]') ?? [],
      )
      const current = enabled.indexOf(buttons.indexOf(document.activeElement as Element))
      const next =
        event.key === 'Home'
          ? enabled[0]
          : event.key === 'End'
            ? enabled[enabled.length - 1]
            : enabled[
                (current + (event.key === 'ArrowDown' ? 1 : -1) + enabled.length) % enabled.length
              ]
      if (next === undefined) return
      this.focusItem(next)
      if (this.isTabs) this.select(this.items[next]!)
    },
    onPointerdown(event: PointerEvent) {
      if (!(this.$refs.root as HTMLElement | undefined)?.contains(event.target as Node))
        this.close()
    },
    onFocusin(event: FocusEvent) {
      // Safari can blur an item before a trigger click without focusing the trigger.
      if (!(this.$refs.root as HTMLElement | undefined)?.contains(event.target as Node))
        this.close()
    },
  },
})
</script>
