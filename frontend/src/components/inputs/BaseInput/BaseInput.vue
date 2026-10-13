<template>
  <textarea
    v-if="multiline"
    :id="id"
    ref="field"
    :class="[style.base, style.textarea, style[size], fullWidth ? style.fullWidth : null]"
    :name="name"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    :style="{ minHeight: normalizedBaseHeight }"
    @input="handleInput"
  />
  <input
    v-else
    :id="id"
    :class="[style.base, style[size], fullWidth ? style.fullWidth : null]"
    :name="name"
    :type="type"
    :value="modelValue"
    :placeholder="placeholder"
    :disabled="disabled"
    @input="handleInput"
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
  },
  updated() {
    this.resizeTextarea()
  },
  methods: {
    handleInput(event: Event) {
      this.$emit(
        'update:modelValue',
        (event.target as HTMLInputElement | HTMLTextAreaElement).value,
      )
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

      field.style.height = this.normalizedBaseHeight
      field.style.height = `${Math.max(field.scrollHeight, field.clientHeight)}px`
    },
  },
})
</script>
