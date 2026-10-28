<template>
  <textarea
    v-if="multiline"
    rows="1"
    :id="id"
    ref="field"
    :class="[style.base, style.textarea, style[size], fullWidth ? style.fullWidth : null]"
    :name="name"
    :autocomplete="autocomplete"
    :required="required"
    :value="modelValue"
    :placeholder="placeholder"
    :maxlength="maxlength"
    :disabled="disabled"
    :style="{
      minHeight: normalizedBaseHeight,
      maxHeight: maxHeight ? `${maxHeight}px` : undefined,
    }"
    @input="handleInput"
    @change="handleInput"
  />
  <input
    v-else
    ref="field"
    :id="id"
    :class="[style.base, style[size], fullWidth ? style.fullWidth : null]"
    :name="name"
    :autocomplete="autocomplete"
    :required="required"
    :type="type"
    :value="modelValue"
    :placeholder="placeholder"
    :maxlength="maxlength"
    :disabled="disabled"
    @input="handleInput"
    @change="handleInput"
    @animationstart="handleAutofill"
    @focus="syncValue"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './BaseInput.module.scss'

export default defineComponent({
  name: 'BaseInput',
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
    maxlength: {
      type: Number,
      default: undefined,
    },
    placeholder: {
      type: String,
      default: '',
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
    maxHeight: {
      type: Number,
      default: undefined,
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
  computed: {
    normalizedBaseHeight(): string {
      return typeof this.baseHeight === 'number' ? `${this.baseHeight}px` : this.baseHeight
    },
  },
  mounted() {
    this.resizeTextarea()
    this.syncValue()
  },
  updated() {
    this.resizeTextarea()
  },
  methods: {
    focus() {
      ;(this.$refs.field as InputElement | undefined)?.focus()
    },
    syncValue() {
      const field = this.$refs.field as InputElement | undefined
      if (field && field.value !== this.modelValue) {
        this.$emit('update:modelValue', field.value)
      }
    },
    handleAutofill() {
      this.syncValue()
    },
    handleInput(event: Event) {
      const value = (event.target as InputElement).value
      if (value === this.modelValue) return
      this.$emit('update:modelValue', value)
      this.$nextTick(() => this.resizeTextarea())
    },
    resizeTextarea() {
      if (!this.multiline) {
        return
      }

      const field = this.$refs.field as HTMLTextAreaElement | undefined

      if (!field) {
        return
      }

      // Measure without a scrollbar so its width cannot change text wrapping while growing.
      field.style.overflowY = 'hidden'
      field.style.height = this.normalizedBaseHeight
      const borderHeight = field.offsetHeight - field.clientHeight
      const height = Math.max(field.scrollHeight + borderHeight, field.offsetHeight)
      field.style.height = `${this.maxHeight ? Math.min(height, this.maxHeight) : height}px`
      field.style.overflowY = this.maxHeight && height > this.maxHeight ? 'auto' : 'hidden'
    },
  },
})
</script>
