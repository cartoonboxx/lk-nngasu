<template>
  <section :class="style.content">
    <header :class="style.topicHeader">
      <p :class="style.eyebrow">Тема {{ topicIndex + 1 }}</p>
      <h2>{{ topic.title }}</h2>
      <p :class="style.description">{{ topic.description }}</p>
    </header>

    <section :class="style.materials">
      <h3>Материалы</h3>
      <div :class="style.materialList">
        <button
          v-for="material in materials"
          :key="material.kind"
          :class="style.material"
          type="button"
          :aria-label="`Скачать: ${material.title}`"
          @click="downloadMaterial(material)"
        >
          <span :class="[style.fileIcon, style[material.kind]]" aria-hidden="true">
            <component :is="material.icon" :size="30" weight="fill" />
          </span>
          <span :class="style.materialCopy">
            <span :class="style.materialTitle">{{ material.title }}</span>
            <span :class="style.materialMeta">{{ material.meta }}</span>
          </span>
          <PhDownloadSimple :class="style.downloadIcon" :size="20" aria-hidden="true" />
        </button>
      </div>
    </section>

    <section :class="style.assignmentSection">
      <div :class="style.assignment">
        <p :class="style.assignmentLabel">Практическая работа №{{ topicIndex + 1 }}</p>
        <p :class="style.assignmentText">{{ topic.task }}</p>

        <div :class="style.assignmentFooter">
          <p :class="style.deadline">
            <PhCalendarBlank :size="19" aria-hidden="true" />
            <span>Срок сдачи: {{ topic.deadline }}, 23:59</span>
          </p>

          <Button>Открыть задание</Button>
        </div>
      </div>
    </section>

    <footer :class="style.footer">
      <Button variant="secondary" :disabled="topicIndex === 0" @click="$emit('previous')">
        <PhCaretLeft :size="18" aria-hidden="true" />
        <span>Предыдущая тема</span>
      </Button>
      <Button variant="secondary" :disabled="topicIndex === topicCount - 1" @click="$emit('next')">
        <span>Следующая тема</span>
        <PhCaretRight :size="18" aria-hidden="true" />
      </Button>
    </footer>
  </section>
</template>

<script lang="ts">
import { defineComponent, markRaw, type Component, type PropType } from 'vue'
import {
  PhCalendarBlank,
  PhCaretLeft,
  PhCaretRight,
  PhDownloadSimple,
  PhFileCode,
  PhFileText,
  PhPresentationChart,
} from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import type { CourseTopic } from '../../course.types.ts'
import style from './ContentCourseComponent.module.scss'

interface CourseMaterial {
  kind: 'lecture' | 'presentation' | 'practice'
  title: string
  meta: string
  icon: Component
}

export default defineComponent({
  name: 'ContentCourseComponent',
  components: { Button, PhCalendarBlank, PhCaretLeft, PhCaretRight, PhDownloadSimple },
  props: {
    topic: { type: Object as PropType<CourseTopic>, required: true },
    topicIndex: { type: Number, required: true },
    topicCount: { type: Number, required: true },
  },
  emits: ['previous', 'next'],
  data() {
    return { style }
  },
  computed: {
    materials(): CourseMaterial[] {
      return [
        {
          kind: 'lecture',
          title: `Лекция ${this.topicIndex + 1}`,
          meta: 'Конспект · MD',
          icon: markRaw(PhFileText),
        },
        {
          kind: 'presentation',
          title: 'План темы',
          meta: 'Учебный материал · MD',
          icon: markRaw(PhPresentationChart),
        },
        {
          kind: 'practice',
          title: `Практическая работа №${this.topicIndex + 1}`,
          meta: 'Задание · MD',
          icon: markRaw(PhFileCode),
        },
      ]
    },
  },

  methods: {
    downloadMaterial(material: CourseMaterial) {
      const topic = this.topic
      const content = `# ${material.title}\n\n## ${topic.title}\n\n${topic.description}\n\n${material.kind === 'practice' ? topic.task : 'План изучения темы'}\n\n${topic.steps.map((step, index) => `${index + 1}. ${step}`).join('\n')}\n`
      const url = URL.createObjectURL(new Blob([content], { type: 'text/markdown;charset=utf-8' }))
      const link = document.createElement('a')
      link.href = url
      link.download = `${topic.id}-${material.kind}.md`
      link.click()
      setTimeout(() => URL.revokeObjectURL(url), 1000)
    },
  },
})
</script>
