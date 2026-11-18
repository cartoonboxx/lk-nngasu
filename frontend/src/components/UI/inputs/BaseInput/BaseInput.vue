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
    :value="displayValue"
    :placeholder="placeholder"
    :maxlength="mask ? undefined : maxlength"
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
    :value="displayValue"
    :placeholder="placeholder"
    :maxlength="mask ? undefined : maxlength"
    :disabled="disabled"
    @input="handleInput"
    @change="handleInput"
    :inputmode="mask ? 'numeric' : undefined"
    @beforeinput="handleBeforeInput"
    @animationstart="syncValue"
    @focus="syncValue"
  />
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './BaseInput.module.scss'

export default defineComponent({
  name: 'BaseInput',
  props: {
    mask: {
      type: String,
      default: '',
    },
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
    displayValue(): string {
      return this.formatValue(this.modelValue)
    },
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
    formatValue(value: string): string {
      if (!this.mask) return value
      const digits = value.replace(/[^0-9]/g, '').slice(0, this.mask.split('#').length - 1)
      let formatted = ''
      let index = 0
      for (const character of this.mask) {
        if (index >= digits.length) break
        formatted += character === '#' ? digits[index++] : character
      }
      return formatted
    },
    syncValue() {
      const field = this.$refs.field as InputElement | undefined
      if (!field) return
      const value = this.formatValue(field.value)
      field.value = value
      if (value !== this.modelValue) this.$emit('update:modelValue', value)
    },
    handleBeforeInput(event: InputEvent) {
      if (!this.mask) return
      const field = event.target as HTMLInputElement
      const start = field.selectionStart
      if (start === null || start !== field.selectionEnd) return
      let from = start
      let to = start
      if (
        event.inputType === 'deleteContentBackward' &&
        start > 0 &&
        /[^0-9]/.test(field.value[start - 1]!)
      ) {
        from--
        while (from > 0 && /[^0-9]/.test(field.value[from - 1]!)) from--
        if (from > 0) from--
      } else if (
        event.inputType === 'deleteContentForward' &&
        start < field.value.length &&
        /[^0-9]/.test(field.value[start]!)
      ) {
        while (to < field.value.length && /[^0-9]/.test(field.value[to]!)) to++
        if (to < field.value.length) to++
      } else {
        return
      }
      event.preventDefault()
      field.setRangeText('', from, to, 'start')
      this.handleInput(event)
    },
    handleInput(event: Event) {
      const field = event.target as InputElement
      const position = field.selectionStart
      const digitsBeforeCursor =
        this.mask && position !== null
          ? field.value.slice(0, position).replace(/[^0-9]/g, '').length
          : null
      const value = this.formatValue(field.value)
      if (this.mask) field.value = value
      if (value !== this.modelValue) this.$emit('update:modelValue', value)
      this.$nextTick(() => {
        if (digitsBeforeCursor !== null && document.activeElement === field) {
          let cursor = 0
          let digits = 0
          while (cursor < value.length && digits < digitsBeforeCursor) {
            if (/[0-9]/.test(value[cursor]!)) digits++
            cursor++
          }
          field.setSelectionRange(cursor, cursor)
        }
        this.resizeTextarea()
      })
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
