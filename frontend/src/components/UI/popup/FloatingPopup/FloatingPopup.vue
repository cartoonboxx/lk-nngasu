<template>
  <div ref="root" :class="style.root" @focusout="onFocusout" @keydown.esc="close(true)">
    <slot name="trigger" :trigger-props="triggerProps" :open="open">
      <button v-bind="triggerProps" type="button" :class="style.trigger">{{ triggerText }}</button>
    </slot>
    <div v-if="open" :class="style.popup">
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

export interface FloatingPopupItem {
  value: string
  label: string
  disabled?: boolean
}

export default defineComponent({
  name: 'FloatingPopup',
  components: { NavTab, StyleTab },
  props: {
    items: { type: Array as PropType<FloatingPopupItem[]>, required: true },
    modelValue: { type: String, default: undefined },
    triggerText: { type: String, default: 'Вкладки' },
    disabled: Boolean,
    closeOnSelect: Boolean,
    mode: { type: String as PropType<'tabs' | 'select'>, default: 'tabs' },
  },
  emits: { 'update:modelValue': (value: string) => typeof value === 'string' },
  data() {
    return { style, open: false, internalValue: undefined as string | undefined }
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
  },
  mounted() {
    document.addEventListener('pointerdown', this.onPointerdown)
  },
  beforeUnmount() {
    document.removeEventListener('pointerdown', this.onPointerdown)
  },
  methods: {
    focusItem(index: number) {
      ;(this.$refs.list as HTMLElement | undefined)
        ?.querySelectorAll<HTMLButtonElement>('[data-popup-item]')
        [index]?.focus()
    },
    async show(last = false) {
      if (this.disabled) return
      this.open = true
      await this.$nextTick()
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
    onFocusout(event: FocusEvent) {
      if (
        !(this.$refs.root as HTMLElement | undefined)?.contains(event.relatedTarget as Node | null)
      )
        this.close()
    },
  },
})
</script>
