<template>
  <div
    :class="[style.root, fullWidth ? style.fullWidth : null]"
    @keydown="onSuggestionKeydown"
    @compositionstart="composing = true"
    @compositionend="
      () => {
        composing = false
        if (suggestionsEnabled) updateValue(currentValue)
      }
    "
  >
    <div v-if="$slots.prefix" :class="style.prefix">
      <slot name="prefix" />
    </div>
    <BaseInput
      ref="field"
      :max-height="maxHeight"
      :id="id"
      :class="[
        style[effectiveVariant],
        $slots.prefix ? style.withPrefix : null,
        $slots.actions ? style.withActions : null,
      ]"
      :name="name"
      :autocomplete="autocomplete ?? (suggestionsEnabled ? 'off' : undefined)"
      :required="required"
      :model-value="currentValue"
      :placeholder="placeholder"
      :maxlength="maxlength"
      :mask="mask"
      :size="size"
      :type="type"
      :disabled="disabled || variant === 'disabled'"
      :full-width="fullWidth"
      :multiline="multiline"
      :base-height="baseHeight"
      @click="onFieldClick"
      @update:model-value="updateValue"
    />
    <div v-if="$slots.actions" :class="style.actions">
      <slot name="actions" />
    </div>
    <FloatingPopup
      v-if="suggestionsEnabled"
      ref="suggestions"
      :class="style.suggestions"
      :anchor="() => ($refs.field as { $el: HTMLElement } | undefined)?.$el"
      :items="filteredOptions"
      :model-value="options.find((item) => item.label === currentValue)?.value ?? ''"
      mode="select"
      close-on-select
      @update:model-value="selectSuggestion"
    >
      <template #trigger><span /></template>
      <template #item="{ item, active }">
        <slot name="option" :option="item" :selected="active">{{ item.label }}</slot>
      </template>
      <template #empty>Ничего не найдено</template>
    </FloatingPopup>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import FloatingPopup from '@/components/UI/popup/FloatingPopup/FloatingPopup.vue'
import BaseInput from '@/components/UI/inputs/BaseInput/BaseInput.vue'
import style from './Input.module.scss'

export default defineComponent({
  name: 'Input',
  components: {
    BaseInput,
    FloatingPopup,
  },
  props: {
    options: {
      type: Array as PropType<FloatingPopupItem[]>,
      default: () => [],
    },
    mask: {
      type: String,
      default: '',
    },
    id: {
      type: String,
      default: undefined,
    },
    maxHeight: {
      type: Number,
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
  emits: ['update:modelValue', 'select'],
  data() {
    return {
      style,
      composing: false,
      currentValue: this.modelValue,
      filterSuggestions: false,
    }
  },
  computed: {
    suggestionsEnabled(): boolean {
      return (
        this.options.length > 0 &&
        !this.multiline &&
        !this.mask &&
        this.type !== 'password' &&
        !this.disabled &&
        this.variant !== 'disabled'
      )
    },
    filteredOptions(): FloatingPopupItem[] {
      if (!this.filterSuggestions) return this.options
      const query = this.currentValue.trim().toLocaleLowerCase('ru').replace(/ё/g, 'е')
      return this.options.filter((item) =>
        item.label.toLocaleLowerCase('ru').replace(/ё/g, 'е').includes(query),
      )
    },
    effectiveVariant(): InputVariant {
      return this.variant === 'default' && this.currentValue.length > 0 ? 'filled' : this.variant
    },
  },
  watch: {
    modelValue(value: string) {
      if (value === this.currentValue) return
      this.currentValue = value
      this.filterSuggestions = false
      ;(this.$refs.suggestions as InstanceType<typeof FloatingPopup> | undefined)?.close()
    },
  },
  methods: {
    async onFieldClick() {
      if (!this.suggestionsEnabled || this.composing) return
      this.filterSuggestions = false
      await this.$nextTick()
      const field = (this.$refs.field as InstanceType<typeof BaseInput> | undefined)?.$el
      if (document.activeElement !== field) return
      await (this.$refs.suggestions as InstanceType<typeof FloatingPopup> | undefined)?.show(
        false,
        false,
      )
    },
    async updateValue(value: string) {
      this.currentValue = value
      this.filterSuggestions = true
      this.$emit('update:modelValue', value)
      const popup = this.$refs.suggestions as InstanceType<typeof FloatingPopup> | undefined
      if (!this.suggestionsEnabled || this.composing) {
        popup?.close()
        return
      }
      const field = (this.$refs.field as InstanceType<typeof BaseInput> | undefined)?.$el
      if (document.activeElement !== field) return
      await this.$nextTick()
      if (this.currentValue === value && document.activeElement === field)
        await (this.$refs.suggestions as InstanceType<typeof FloatingPopup> | undefined)?.show(
          false,
          false,
        )
    },
    selectSuggestion(value: string) {
      const item = this.filteredOptions.find((option) => option.value === value)
      if (!item || item.disabled) return
      this.currentValue = item.label
      this.$emit('update:modelValue', item.label)
      this.$emit('select', item)
    },
    async onSuggestionKeydown(event: KeyboardEvent) {
      const popup = this.$refs.suggestions as InstanceType<typeof FloatingPopup> | undefined
      const field = (this.$refs.field as InstanceType<typeof BaseInput> | undefined)?.$el
      if (!popup || event.target !== field || event.isComposing || this.composing) return
      if (event.key === 'ArrowDown' || event.key === 'ArrowUp') {
        event.preventDefault()
        if (!popup.open) this.filterSuggestions = false
        await popup.show(event.key === 'ArrowUp')
      } else if (event.key === 'Escape' && popup.open) {
        event.preventDefault()
        event.stopPropagation()
        popup.close()
      } else if (event.key === 'Enter' && popup.open) {
        const item = this.filteredOptions.find((option) => !option.disabled)
        if (!item) return
        event.preventDefault()
        popup.select(item)
      }
    },
    focus() {
      ;(this.$refs.field as InstanceType<typeof BaseInput> | undefined)?.focus()
    },
  },
})
</script>
