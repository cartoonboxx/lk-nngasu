<template>
  <section :class="style.content">
    <header :class="style.topicHeader">
      <p :class="style.eyebrow">Тема {{ topicIndex + 1 }}</p>
      <h2>{{ topic.title }}</h2>
      <p :class="style.description">{{ topic.description }}</p>
    </header>
    <section :class="style.materials">
      <h3>Материалы</h3>
      <p v-if="!topic.materials.length">Материалы пока не добавлены</p>
      <div :class="style.materialList">
        <div v-for="material in topic.materials" :key="material.id">
          <MediaButton
            v-if="material.file"
            :title="material.title"
            :filename="material.file.name"
            :meta="`${material.file.name} · ${(material.file.size_bytes / 1024).toFixed(1)} КБ`"
            @click="download(material.file.id, material.file.name)"
          >
            <template #icon><PhFileText /></template>
          </MediaButton>
          <a
            v-if="safeLink(material.external_url)"
            :href="material.external_url!"
            target="_blank"
            rel="noopener noreferrer"
            >{{ material.title }}</a
          >
        </div>
      </div>
      <p v-if="error" role="alert">{{ error }}</p>
    </section>
    <section :class="style.assignmentSection">
      <p v-if="!topic.assignments.length">Задания пока не добавлены</p>
      <div v-for="assignment in topic.assignments" :key="assignment.id" :class="style.assignment">
        <h3>{{ assignment.title }}</h3>
        <p :class="style.assignmentText">{{ assignment.description }}</p>
        <div :class="style.assignmentFooter">
          <p :class="style.deadline">
            Срок сдачи:
            {{
              assignment.deadline_at
                ? new Date(assignment.deadline_at).toLocaleString('ru-RU')
                : 'не указан'
            }}
          </p>
          <Button @click="$emit('open-assignment', assignment.id)">Открыть задание</Button>
        </div>
      </div>
    </section>
    <footer :class="style.footer">
      <Button variant="secondary" :disabled="topicIndex === 0" @click="$emit('previous')"
        >Предыдущая тема</Button
      >
      <Button variant="secondary" :disabled="topicIndex === topicCount - 1" @click="$emit('next')"
        >Следующая тема</Button
      >
    </footer>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import MediaButton from '@/components/UI/buttons/MediaButton/MediaButton.vue'
import { PhFileText } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import { useFilesStore } from '@/stores/files'
import type { CourseTopic } from '../../course.types'
import style from './ContentCourseComponent.module.scss'
export default defineComponent({
  name: 'ContentCourseComponent',
  components: { Button, MediaButton, PhFileText },
  props: {
    topic: { type: Object as PropType<CourseTopic>, required: true },
    topicIndex: { type: Number, required: true },
    topicCount: { type: Number, required: true },
  },
  emits: ['previous', 'next', 'open-assignment'],
  data() {
    return { style, error: '' }
  },
  methods: {
    safeLink(url: string | null): boolean {
      if (!url) return false
      try {
        return ['http:', 'https:'].includes(new URL(url).protocol)
      } catch {
        return false
      }
    },
    async download(id: number, name: string) {
      this.error = ''
      try {
        await useFilesStore().download(id, name)
      } catch (error) {
        this.error = error instanceof Error ? error.message : 'Не удалось скачать файл.'
      }
    },
  },
})
</script>
