<template>
  <div :class="[style.root, fullWidth ? style.fullWidth : null]">
    <label v-if="label" :class="style.label" :for="fieldId">
      {{ label }}
    </label>
    <Input
      :id="fieldId"
      :name="name"
      :autocomplete="autocomplete"
      :required="required"
      :ariaInvalid="ariaInvalid"
      :ariaDescribedby="ariaDescribedby"
      :model-value="modelValue"
      :placeholder="placeholder"
      :variant="variant"
      :size="size"
      :type="type"
      :disabled="disabled"
      :full-width="fullWidth"
      :multiline="multiline"
      :base-height="baseHeight"
      @update:model-value="handleUpdate"
    >
      <template v-if="$slots.actions" #actions>
        <slot name="actions" />
      </template>
    </Input>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import Input from '@/components/inputs/Input/Input.vue'
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
    label: {
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
      generatedId: `text-input-${++nextTextInputId}`,
    }
  },
  computed: {
    fieldId(): string {
      return this.id ?? this.generatedId
    },
  },
  methods: {
    handleUpdate(value: string) {
      this.$emit('update:modelValue', value)
    },
  },
})
</script>
