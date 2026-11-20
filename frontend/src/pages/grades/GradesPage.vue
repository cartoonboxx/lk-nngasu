<template>
  <section :class="style.page">
    <SectionHeader header="Оценки" paragraph="Результаты проверки ваших заданий" />
    <GradesFilters
      v-model:search="search"
      v-model:subject="subject"
      v-model:status="status"
      v-model:semester="semester"
      :subject-options="subjectOptions"
      :semester-options="semesterOptions"
    />
    <PageLoader v-if="cabinet.gradesLoading" />
    <p v-else-if="cabinet.gradesError">
      <button @click="cabinet.fetchGrades()">Повторить</button>
    </p>
    <PageLoader v-if="detailsLoading" />

    <GradesTable
      v-if="!cabinet.gradesLoading && !cabinet.gradesError"
      :assignments="filteredAssignments"
      :total="semesterAssignments.length"
      @select="openDetails"
      @reset="
        () => {
          search = ''
          subject = 'all'
          status = 'all'
          semester = 'all'
        }
      "
    />
    <p :class="style.footnote">
      <PhInfo :size="18" /> Нажмите на название задания, чтобы посмотреть ответ и комментарий
      преподавателя.
    </p>
    <GradeDetailsModal v-model="detailsOpen" :assignment="selectedAssignment" />
  </section>
</template>

<script lang="ts">
import { apiNotifications } from '@/api/client'
import { defineComponent } from 'vue'
import { PhInfo } from '@phosphor-icons/vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import GradesFilters from './components/GradesFilters/GradesFilters.vue'
import GradesTable from './components/GradesTable/GradesTable.vue'
import GradeDetailsModal from './components/GradeDetailsModal/GradeDetailsModal.vue'
import { useCabinetStore } from '@/stores/cabinet'
import type { GradeAssignment } from './grades.types'
import style from './GradesPage.module.scss'
import PageLoader from '@/components/PageLoader/PageLoader.vue'
export default defineComponent({
  name: 'GradesPage',
  components: {
    SectionHeader,
    GradesFilters,
    GradesTable,
    GradeDetailsModal,
    PhInfo,
    PageLoader
  },
  data() {
    return {
      style,
      detailsLoading: false,
      detailsVersion: 0,
      search: '',
      subject: 'all',
      status: 'all',
      semester: 'all',
      detailsOpen: false,
      selectedAssignment: null as GradeAssignment | null,
    }
  },
  async created() {
    await this.cabinet.fetchGrades()
  },
  beforeUnmount() {
    this.detailsVersion++
  },
  methods: {
    async openDetails(row: GradeAssignment) {
      const version = ++this.detailsVersion
      this.detailsLoading = true

      try {
        const assignment = await this.$subjectsStore.getAssignment(row.id)
        if (version !== this.detailsVersion) return
        this.selectedAssignment = {
          ...row,
          answer: assignment.submission?.answer_text ?? '',
          teacherComment: assignment.submission?.review_comment ?? '',
          files: assignment.submission?.files ?? [],
          submittedAt: assignment.submission?.submitted_at ?? null,
          grade: assignment.submission?.grade ?? null,
        }
        this.detailsOpen = true
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        if (version === this.detailsVersion) this.detailsLoading = false
      }
    },
  },
  computed: {
    cabinet() {
      return useCabinetStore()
    },
    semesterOptions() {
      return [
        { label: 'Все семестры', value: 'all' },
        ...[...new Set(this.cabinet.grades.map((row) => row.semester))].map((semester) => ({
          label: semester,
          value: semester,
        })),
      ]
    },
    semesterAssignments(): GradeAssignment[] {
      return this.cabinet.grades.filter(
        (assignment) => this.semester === 'all' || assignment.semester === this.semester,
      )
    },
    subjectOptions() {
      return [
        { label: 'Все дисциплины', value: 'all' },
        ...[...new Set(this.cabinet.grades.map((assignment) => assignment.subject))].map(
          (subject) => ({ label: subject, value: subject }),
        ),
      ]
    },
    filteredAssignments(): GradeAssignment[] {
      const query = this.search.trim().toLocaleLowerCase('ru')
      return this.semesterAssignments.filter(
        (assignment) =>
          (this.subject === 'all' || assignment.subject === this.subject) &&
          (this.status === 'all' || assignment.status === this.status) &&
          `${assignment.title} ${assignment.topic} ${assignment.subject}`
            .toLocaleLowerCase('ru')
            .includes(query),
      )
    },
  },
})
</script>
