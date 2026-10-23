<template>
  <FloatingPopup
    v-bind="$attrs"
    :class="[style.root, fullWidth && style.fullWidth]"
    :items="options"
    :model-value="modelValue"
    :disabled="isDisabled"
    mode="select"
    close-on-select
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <template #trigger="{ triggerProps, open }">
      <StyleTab
        :text="selected?.label ?? placeholder"
        :disabled="isDisabled"
        :class="[
          style.trigger,
          style[size],
          style[variant],
          selected && style.filled,
          !selected && style.placeholder,
        ]"
        :trigger-attrs="{
          ...triggerProps,
          class: style.control,
        }"
      >
        <template #trailing>
          <PhCaretDown :class="[style.caret, open && style.open]" :size="18" />
        </template>
      </StyleTab>
    </template>
    <template #item="{ item, active }">
      <slot name="option" :option="item" :selected="active">{{ item.label }}</slot>
    </template>
  </FloatingPopup>
  <input v-if="name" type="hidden" :name="name" :value="modelValue ?? ''" :disabled="isDisabled" />
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhCaretDown } from '@phosphor-icons/vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import FloatingPopup, {
  type FloatingPopupItem,
} from '@/components/UI/popup/FloatingPopup/FloatingPopup.vue'
import style from './Select.module.scss'

export default defineComponent({
  name: 'Select',
  inheritAttrs: false,
  components: { FloatingPopup, StyleTab, PhCaretDown },
  props: {
    modelValue: {
      type: String,
      default: undefined,
    },
    options: {
      type: Array as PropType<FloatingPopupItem[]>,
      required: true,
    },
    placeholder: {
      type: String,
      default: 'Выберите значение',
    },
    name: {
      type: String,
      default: undefined,
    },
    size: {
      type: String as PropType<'sm' | 'md' | 'lg'>,
      default: 'md',
    },
    variant: {
      type: String as PropType<'default' | 'filled' | 'disabled'>,
      default: 'default',
    },
    disabled: Boolean,
    fullWidth: Boolean,
  },
  emits: { 'update:modelValue': (value: string) => typeof value === 'string' },
  data() {
    return { style }
  },
  computed: {
    selected(): FloatingPopupItem | undefined {
      return this.options.find((item) => item.value === this.modelValue)
    },
    isDisabled(): boolean {
      return this.disabled || this.variant === 'disabled'
    },
  },
})
</script>
