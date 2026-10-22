<template>
  <aside :class="style.courses">
    <div ref="topicList" :class="style.topicList">
      <NavTab
        v-for="(topic, index) in topics"
        :key="topic.id"
        :class="style.topic"
        :text="topic.title"
        :active="index === activeTopicIndex"
        :trigger-attrs="{
          tabindex: index === activeTopicIndex ? 0 : -1,
          onKeydown: (event: KeyboardEvent) => onTopicKeydown(event, index),
        }"
        @select="$emit('select', index)"
      >
        <span :class="style.topicCopy">
          <span :class="style.topicNumber">Тема {{ index + 1 }}.</span>
          <span :class="style.topicTitle">{{ topic.title }}</span>
        </span>
      </NavTab>
    </div>
  </aside>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import NavTab from '@/components/UI/tabs/NavTab/NavTab.vue'
import type { CourseTopic } from '../../course.types.ts'
import style from './AsideCoursePanel.module.scss'

export default defineComponent({
  name: 'AsideCoursePanel',
  components: { NavTab },
  props: {
    topics: {
      type: Array as PropType<CourseTopic[]>,
      required: true,
    },
    activeTopicIndex: {
      type: Number,
      required: true,
    },
  },
  emits: { select: (index: number) => Number.isInteger(index) && index >= 0 },
  data() {
    return { style }
  },
  methods: {
    async onTopicKeydown(event: KeyboardEvent, index: number) {
      let target: number
      switch (event.key) {
        case 'ArrowDown':
          target = (index + 1) % this.topics.length
          break
        case 'ArrowUp':
          target = (index - 1 + this.topics.length) % this.topics.length
          break
        case 'Home':
          target = 0
          break
        case 'End':
          target = this.topics.length - 1
          break
        default:
          return
      }
      event.preventDefault()
      this.$emit('select', target)
      await this.$nextTick()
      ;(this.$refs.topicList as HTMLElement)
        ?.querySelectorAll<HTMLButtonElement>('button')
        [target]?.focus()
    },
  },
})
</script>
