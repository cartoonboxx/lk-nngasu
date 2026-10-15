<template>
  <div :class="[style.root, fullWidth ? style.fullWidth : null]">
    <div v-if="$slots.prefix" :class="style.prefix">
      <slot name="prefix" />
    </div>
    <BaseInput
      :id="id"
      :class="[
        style[effectiveVariant],
        $slots.prefix ? style.withPrefix : null,
        $slots.actions ? style.withActions : null,
      ]"
      :name="name"
      :autocomplete="autocomplete"
      :required="required"
      :ariaInvalid="ariaInvalid"
      :ariaDescribedby="ariaDescribedby"
      :model-value="currentValue"
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

import BaseInput from '@/components/UI/inputs/BaseInput/BaseInput.vue'
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
    autocomplete: {
      type: String,
      default: undefined,
    },
    required: {
      type: Boolean,
      default: false,
    },
    ariaInvalid: {
      type: Boolean,
      default: false,
    },
    ariaDescribedby: {
      type: String,
      default: undefined,
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
      currentValue: this.modelValue,
    }
  },
  computed: {
    effectiveVariant(): InputVariant {
      return this.variant === 'default' && this.currentValue.length > 0 ? 'filled' : this.variant
    },
  },
  watch: {
    modelValue(value: string) {
      this.currentValue = value
    },
  },
  methods: {
    handleUpdate(value: string) {
      this.currentValue = value
      this.$emit('update:modelValue', value)
    },
  },
})
</script>
