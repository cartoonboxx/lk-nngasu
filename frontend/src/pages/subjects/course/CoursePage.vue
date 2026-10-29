<template>
  <section :class="style.coursePage">
    <p v-if="!subject">Предмет не найден</p>
    <template v-else>
      <SectionHeader
        :header="subject.label"
        :paragraph="`${subject.semester} семестр · Учебные материалы курса`"
      />
      <p v-if="!activeTopic">Задания пока не добавлены</p>
      <template v-else>
        <div :class="style.courseContent">
          <AsideCoursePanel
            :topics="topics"
            :active-topic-index="activeTopicIndex"
            @select="activeTopicIndex = $event"
          />
          <ContentCourseComponent
            :topic="activeTopic"
            :topic-index="activeTopicIndex"
            :topic-count="topics.length"
            @open-assignment="assignmentOpen = true"
            @previous="activeTopicIndex = Math.max(0, activeTopicIndex - 1)"
            @next="activeTopicIndex = Math.min(topics.length - 1, activeTopicIndex + 1)"
          />
        </div>
        <AssignmentModal
          v-model="assignmentOpen"
          :course-title="subject.label"
          :topic="activeTopic"
          :topic-index="activeTopicIndex"
          :answer="activeAnswer"
          @submit="submitAnswer"
          @delete="!activeAnswer?.review && delete assignmentAnswers[answerKey]"
        />
      </template>
    </template>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import type { SubjectDetails } from '../subjects.types'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import AsideCoursePanel from './components/AsideCoursePanel/AsideCoursePanel.vue'
import ContentCourseComponent from './components/ContentCourseComponent/ContentCourseComponent.vue'
import AssignmentModal from './components/AssignmentModal/AssignmentModal.vue'
import type { AssignmentAnswer, AssignmentAnswerDraft, CourseTopic } from './course.types'
import { assignmentAnswers } from './assignmentAnswers'
import style from './CoursePage.module.scss'

export default defineComponent({
  name: 'CoursePage',
  components: { SectionHeader, AsideCoursePanel, ContentCourseComponent, AssignmentModal },
  data() {
    return {
      style,
      activeTopicIndex: 0,
      assignmentOpen: false,
      assignmentAnswers,
    }
  },
  beforeUnmount() {
    this.$subjectsStore.currentSubject = null
  },
  computed: {
    subject(): SubjectDetails | null {
      return this.$subjectsStore.currentSubject
    },
    topics(): CourseTopic[] {
      return this.$subjectsStore.assignments
    },
    answerKey(): string {
      return JSON.stringify([this.subject?.id, this.activeTopic?.id])
    },
    activeAnswer(): AssignmentAnswer | null {
      return this.assignmentAnswers[this.answerKey] ?? null
    },
    activeTopic(): CourseTopic | undefined {
      return this.topics[this.activeTopicIndex]
    },
  },
  watch: {
    '$route.params.id': {
      immediate: true,
      handler() {
        if (this.$route.name !== 'course') return
        this.activeTopicIndex = 0
        this.assignmentOpen = false
        this.$subjectsStore.fetchSubject(String(this.$route.params.id))
      },
    },
  },
  methods: {
    submitAnswer(draft: AssignmentAnswerDraft) {
      if (!this.activeTopic || this.activeAnswer?.review) return
      this.assignmentAnswers[this.answerKey] = { ...draft, submittedAt: new Date().toISOString() }
    },
  },
})
</script>
