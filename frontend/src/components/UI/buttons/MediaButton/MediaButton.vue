<template>
  <div
    :class="[style.mediaButton, readonly ? style.readonly : null, disabled ? style.disabled : null]"
  >
    <component
      :is="readonly ? 'div' : 'button'"
      :class="style.main"
      :type="readonly ? undefined : type"
      :disabled="!readonly && disabled ? true : undefined"
      :aria-label="readonly ? undefined : ariaLabel"
      @click="handleClick"
    >
      <span :class="style.leftPart">
        <span v-if="$slots.icon" :class="style.icon" aria-hidden="true"><slot name="icon" /></span>
        <span :class="style.content">
          <slot>
            <span :class="style.title">{{ title }}</span>
            <span v-if="meta" :class="style.meta">{{ meta }}</span>
          </slot>
        </span>
      </span>
      <PhDownloadSimple
        v-if="!readonly && !$slots.actions"
        :class="style.download"
        :size="20"
        aria-hidden="true"
      />
    </component>
    <div v-if="$slots.actions" :class="style.actions" :inert="disabled || undefined">
      <slot name="actions" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhDownloadSimple } from '@phosphor-icons/vue'
import style from './MediaButton.module.scss'

export default defineComponent({
  name: 'MediaButton',
  components: { PhDownloadSimple },
  props: {
    title: {
      type: String,
      default: '',
    },
    meta: {
      type: String,
      default: '',
    },
    ariaLabel: {
      type: String,
      default: undefined,
    },
    readonly: {
      type: Boolean,
      default: false,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    type: {
      type: String as PropType<ButtonType>,
      default: 'button',
    },
  },
  emits: { click: (event: MouseEvent) => event instanceof MouseEvent },
  data() {
    return { style }
  },
  methods: {
    handleClick(event: MouseEvent) {
      if (!this.readonly && !this.disabled) this.$emit('click', event)
    },
  },
})
</script>
