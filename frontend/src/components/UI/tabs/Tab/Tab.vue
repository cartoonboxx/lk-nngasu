<template>
  <BaseTab
    :class="style.tab"
    :text="text"
    :image-src="imageSrc"
    :image-alt="imageAlt"
    :icon="icon"
    :icon-size="iconSize"
    :icon-weight="iconWeight"
    :active="active"
    :disabled="disabled"
    @select="$emit('select')"
  >
    <template v-if="$slots.actions" #actions>
      <slot name="actions" />
    </template>
  </BaseTab>
</template>

<script lang="ts">
import { defineComponent, type Component, type PropType } from 'vue'

import BaseTab from '@/components/UI/tabs/BaseTab/BaseTab.vue'
import style from './Tab.module.scss'

export default defineComponent({
  name: 'Tab',
  emits: ['select'],
  components: {
    BaseTab,
  },
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
