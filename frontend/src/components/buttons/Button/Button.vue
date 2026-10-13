<template>
  <BaseButton
    :class="[style[variant], state !== 'default' ? style[state] : null]"
    :size="size"
    :type="type"
    :state="state"
    :disabled="disabled"
    :full-width="fullWidth"
    :loading="loading"
  >
    <slot />
  </BaseButton>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'

import BaseButton from '@/components/buttons/BaseButton/BaseButton.vue'
import style from './Button.module.scss'

export default defineComponent({
  name: 'Button',
  components: {
    BaseButton,
  },
  props: {
    variant: {
      type: String as PropType<ButtonVariant>,
      default: 'primary',
    },
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
