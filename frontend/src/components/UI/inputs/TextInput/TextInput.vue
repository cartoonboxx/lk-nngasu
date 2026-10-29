<template>
  <div :class="[style.root, fullWidth ? style.fullWidth : null]">
    <label v-if="label" :class="style.label" :for="fieldId">
      {{ label }}
    </label>
    <Input
      ref="field"
      :maxlength="maxlength"
      :max-height="maxHeight"
      :id="fieldId"
      :name="name"
      :autocomplete="autocomplete"
      :required="required"
      :model-value="modelValue"
      :placeholder="placeholder"
      :variant="error && variant !== 'disabled' ? 'error' : variant"
      :size="size"
      :type="type"
      :disabled="disabled"
      :full-width="fullWidth"
      :multiline="multiline"
      :base-height="baseHeight"
      @update:model-value="$emit('update:modelValue', $event)"
    >
      <template v-if="$slots.prefix" #prefix>
        <slot name="prefix" />
      </template>
      <template v-if="$slots.actions" #actions>
        <slot name="actions" />
      </template>
    </Input>
    <p v-if="error" :id="`${fieldId}-error`" :class="style.error" role="alert">
      {{ error }}
    </p>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import Input from '@/components/UI/inputs/Input/Input.vue'
import style from './TextInput.module.scss'

let nextTextInputId = 0

export default defineComponent({
  name: 'TextInput',
  components: {
    Input,
  },
  props: {
    id: {
      type: String,
      default: undefined,
    },
    maxlength: {
      type: Number,
      default: undefined,
    },
    maxHeight: {
      type: Number,
      default: undefined,
    },
    label: {
      type: String,
      default: '',
    },
    error: {
      type: String,
      default: '',
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
      generatedId: `text-input-${++nextTextInputId}`,
    }
  },
  computed: {
    fieldId(): string | undefined {
      return this.id ?? (this.label || this.error ? this.generatedId : undefined)
    },
  },
  methods: {
    focus() {
      ;(this.$refs.field as InstanceType<typeof Input> | undefined)?.focus()
    },
  },
})
</script>
