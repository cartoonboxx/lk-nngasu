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
    :trigger-attrs="triggerAttrs"
    @select="$emit('select')"
  >
    <template v-if="$slots.default" #default><slot /></template>
    <template v-if="$slots.trailing" #trailing><slot name="trailing" /></template>
    <template v-if="$slots.actions" #actions>
      <slot name="actions" />
    </template>
  </BaseTab>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { baseTabProps } from '../tabProps'

import BaseTab from '@/components/UI/tabs/BaseTab/BaseTab.vue'
import style from './StyleTab.module.scss'

export default defineComponent({
  name: 'StyleTab',
  emits: ['select'],
  components: {
    BaseTab,
  },
  props: baseTabProps,
  data() {
    return {
      style,
    }
  },
})
</script>
