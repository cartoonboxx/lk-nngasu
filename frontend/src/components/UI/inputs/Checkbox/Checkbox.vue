<template>
  <label
    :class="[
      style.root,
      style[size],
      style[variant],
      disabled || state === 'disabled' ? style.disabled : null,
      state !== 'default' && state !== 'disabled' ? style[state] : null,
    ]"
  >
    <input
      :class="style.input"
      :name="name"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled || state === 'disabled'"
      @change="handleChange"
    />
    <span :class="style.control" />
    <span v-if="label" :class="style.label">{{ label }}</span>
  </label>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import style from './Checkbox.module.scss'

export default defineComponent({
  name: 'Checkbox',
  props: {
    name: {
      type: String,
      default: undefined,
    },
    modelValue: {
      type: Boolean,
      default: false,
    },
    label: {
      type: String,
      default: '',
    },
    size: {
      type: String as PropType<CheckboxSize>,
      default: 'md',
    },
    variant: {
      type: String as PropType<CheckboxVariant>,
      default: 'default',
    },
    state: {
      type: String as PropType<CheckboxState>,
      default: 'default',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      style,
    }
  },
  methods: {
    handleChange(event: Event) {
      this.$emit('update:modelValue', (event.target as HTMLInputElement).checked)
    },
  },
})
</script>
