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
      :id="fieldId"
      :class="style.input"
      :name="name"
      type="checkbox"
      :checked="modelValue"
      :disabled="disabled || state === 'disabled'"
      :aria-label="label ? undefined : ariaLabel"
      @change="handleChange"
    />
    <span :class="style.control" aria-hidden="true" />
    <span v-if="label" :class="style.label">{{ label }}</span>
  </label>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import style from './Checkbox.module.scss'

let nextCheckboxId = 0

export default defineComponent({
  name: 'Checkbox',
  props: {
    id: {
      type: String,
      default: undefined,
    },
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
    ariaLabel: {
      type: String,
      default: undefined,
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
      generatedId: `checkbox-${++nextCheckboxId}`,
    }
  },
  computed: {
    fieldId(): string {
      return this.id ?? this.generatedId
    },
  },
  methods: {
    handleChange(event: Event) {
      this.$emit('update:modelValue', (event.target as HTMLInputElement).checked)
    },
  },
})
</script>
