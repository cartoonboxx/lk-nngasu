<template>
  <section :class="style.coursePage">
    <PageLoader v-if="$subjectsStore.loading" />
    <p v-else-if="$subjectsStore.error">
      <button @click="$subjectsStore.fetchSubject(String($route.params.id))">Повторить</button>
    </p>
    <p v-else-if="!subject">Предмет не найден</p>
    <template v-else>
      <SectionHeader
        :header="subject.label"
        :paragraph="`${subject.semesterLabel} · Учебные материалы курса`"
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
            @open-assignment="
              (id: number) => {
                selectedAssignmentId = id
                assignmentOpen = true
              }
            "
            @previous="activeTopicIndex = Math.max(0, activeTopicIndex - 1)"
            @next="activeTopicIndex = Math.min(topics.length - 1, activeTopicIndex + 1)"
          />
        </div>
        <AssignmentModal
          v-if="selectedAssignmentId !== null"
          :key="selectedAssignmentId"
          v-model="assignmentOpen"
          :course-title="subject.label"
          :assignment-id="selectedAssignmentId"
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
import type { CourseTopic } from './course.types'
import style from './CoursePage.module.scss'
import PageLoader from '@/components/PageLoader/PageLoader.vue'
export default defineComponent({
  name: 'CoursePage',
  components: { SectionHeader, AsideCoursePanel, ContentCourseComponent, AssignmentModal,PageLoader },
  data() {
    return {
      style,
      activeTopicIndex: 0,
      assignmentOpen: false,
      selectedAssignmentId: null as number | null,
    }
  },
  beforeUnmount() {
    this.$subjectsStore.clearSession()
  },
  computed: {
    subject(): SubjectDetails | null {
      return this.$subjectsStore.currentSubject
    },
    topics(): CourseTopic[] {
      return this.$subjectsStore.topics
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
        this.selectedAssignmentId = null
        this.$subjectsStore.fetchSubject(String(this.$route.params.id))
      },
    },
  },
})
</script>
