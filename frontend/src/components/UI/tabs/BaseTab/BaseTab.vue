<template>
  <div
    :class="[style.base, active ? style.active : null, disabled ? style.disabled : null]"
    :data-active="active"
    :data-disabled="disabled"
  >
    <component
      :is="to ? 'RouterLink' : 'button'"
      :to="to"
      v-press
      :class="style.trigger"
      :type="to ? undefined : 'button'"
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
    </component>

    <span v-if="$slots.actions" :class="style.actions">
      <slot name="actions" />
    </span>
  </div>
</template>

<script lang="ts">
import { pressFeedback } from '@/directives/pressFeedback'
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import { baseTabProps } from '../tabProps'

import style from './BaseTab.module.scss'

export default defineComponent({
  directives: { press: pressFeedback },
  name: 'BaseTab',
  components: { RouterLink },
  emits: ['select'],
  props: baseTabProps,
  data() {
    return {
      style,
    }
  },
})
</script>
