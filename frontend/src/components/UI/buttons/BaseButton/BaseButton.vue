<template>
  <button
    :class="[
      style.base,
      style[size],
      state !== 'default' ? style[state] : null,
      fullWidth ? style.fullWidth : null,
    ]"
    :type="type"
    :disabled="disabled || loading || state === 'disabled'"
    :aria-busy="loading"
  >
    <span v-if="loading" :class="style.loader" aria-hidden="true" />
    <span :class="style.content">
      <slot />
    </span>
  </button>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import style from './BaseButton.module.scss'

export default defineComponent({
  name: 'BaseButton',
  props: {
    size: {
      type: String as PropType<ButtonSize>,
      default: 'md',
    },
    type: {
      type: String as PropType<ButtonType>,
      default: 'button',
    },
    state: {
      type: String as PropType<ButtonState>,
      default: 'default',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    fullWidth: {
      type: Boolean,
      default: false,
    },
    loading: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      style,
    }
  },
})
</script>
