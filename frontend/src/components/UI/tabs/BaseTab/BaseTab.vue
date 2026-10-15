<template>
  <div
    :class="[style.base, active ? style.active : null, disabled ? style.disabled : null]"
    :data-active="active"
    :data-disabled="disabled"
  >
    <button
      :class="style.trigger"
      type="button"
      role="tab"
      :aria-selected="active"
      :disabled="disabled"
      @click="$emit('select')"
    >
      <span v-if="icon || imageSrc" :class="style.media">
        <component
          :is="icon"
          v-if="icon"
          :size="iconSize"
          :weight="active ? 'fill' : iconWeight"
          aria-hidden="true"
        />
        <img v-else :class="style.image" :src="imageSrc" :alt="imageAlt" />
      </span>
      <span :class="style.text">{{ text }}</span>
    </button>

    <span v-if="$slots.actions" :class="style.actions">
      <slot name="actions" />
    </span>
  </div>
</template>

<script lang="ts">
import { defineComponent, type Component, type PropType } from 'vue'

import style from './BaseTab.module.scss'

export default defineComponent({
  name: 'BaseTab',
  emits: ['select'],
  props: {
    text: {
      type: String,
      required: true,
    },
    imageSrc: {
      type: String,
      default: undefined,
    },
    imageAlt: {
      type: String,
      default: '',
    },
    icon: {
      type: [Object, Function] as PropType<Component>,
      default: undefined,
    },
    iconSize: {
      type: Number,
      default: 22,
    },
    iconWeight: {
      type: String,
      default: 'regular',
    },
    active: {
      type: Boolean,
      default: false,
    },
    disabled: {
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
