<template>
  <div :class="[style.root, fullWidth ? style.fullWidth : null]">
    <BaseInput
      :id="id"
      :class="[style[variant], $slots.actions ? style.withActions : null]"
      :name="name"
      :model-value="modelValue"
      :placeholder="placeholder"
      :size="size"
      :type="type"
      :disabled="disabled || variant === 'disabled'"
      :full-width="fullWidth"
      :multiline="multiline"
      :base-height="baseHeight"
      @update:model-value="handleUpdate"
    />
    <div v-if="$slots.actions" :class="style.actions">
      <slot name="actions" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import BaseInput from '@/components/inputs/BaseInput/BaseInput.vue'
import style from './Input.module.scss'

export default defineComponent({
  name: 'Input',
  components: {
    BaseInput,
  },
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
      type: String,
      default: '',
    },
    placeholder: {
      type: String,
      default: '',
    },
    variant: {
      type: String as PropType<InputVariant>,
      default: 'default',
    },
    size: {
      type: String as PropType<InputSize>,
      default: 'md',
    },
    type: {
      type: String as PropType<InputType>,
      default: 'text',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    fullWidth: {
      type: Boolean,
      default: false,
    },
    multiline: {
      type: Boolean,
      default: false,
    },
    baseHeight: {
      type: [Number, String],
      default: 96,
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      style,
    }
  },
  methods: {
    handleUpdate(value: string) {
      this.$emit('update:modelValue', value)
    },
  },
})
</script>
