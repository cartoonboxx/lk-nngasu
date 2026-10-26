<template>
  <div
    :class="[style.base, active ? style.active : null, disabled ? style.disabled : null]"
    :data-active="active"
    :data-disabled="disabled"
  >
    <button
      v-press
      :class="style.trigger"
      type="button"
      :disabled="disabled"
      v-bind="triggerAttrs"
      @click="$emit('select')"
    >
      <span v-if="icon || imageSrc" :class="style.media">
        <component :is="icon" v-if="icon" :size="iconSize" :weight="active ? 'fill' : iconWeight" />
        <img v-else :class="style.image" :src="imageSrc" :alt="imageAlt" />
      </span>
      <span :class="style.text"
        ><slot>{{ text }}</slot></span
      >
      <slot name="trailing" />
    </button>

    <span v-if="$slots.actions" :class="style.actions">
      <slot name="actions" />
    </span>
  </div>
</template>

<script lang="ts">
import { pressFeedback } from '@/directives/pressFeedback'
import { defineComponent } from 'vue'
import { baseTabProps } from '../tabProps'

import style from './BaseTab.module.scss'

export default defineComponent({
  directives: { press: pressFeedback },
  name: 'BaseTab',
  emits: ['select'],
  props: baseTabProps,
  data() {
    return {
      style,
    }
  },
})
</script>
