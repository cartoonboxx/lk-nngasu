<template>
  <section :class="style.page">
    <SectionHeader header="Оценки" paragraph="Результаты проверки ваших заданий" />
    <GradesFilters
      v-model:search="search"
      v-model:subject="subject"
      v-model:status="status"
      v-model:semester="semester"
      :subject-options="subjectOptions"
    />
    <GradesTable
      :assignments="filteredAssignments"
      :total="semesterAssignments.length"
      @select="openAssignment"
      @reset="resetFilters"
    />
    <p :class="style.footnote">
      <PhInfo :size="18" /> Нажмите на название задания, чтобы посмотреть ответ и комментарий
      преподавателя.
    </p>
    <GradeDetailsModal v-model="detailsOpen" :assignment="selectedAssignment" />
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { PhInfo } from '@phosphor-icons/vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import GradesFilters from './components/GradesFilters/GradesFilters.vue'
import GradesTable from './components/GradesTable/GradesTable.vue'
import GradeDetailsModal from './components/GradeDetailsModal/GradeDetailsModal.vue'
import { gradeAssignments, type GradeAssignment } from './grades.data'
import style from './GradesPage.module.scss'

export default defineComponent({
  name: 'GradesPage',
  components: {
    SectionHeader,
    GradesFilters,
    GradesTable,
    GradeDetailsModal,
    PhInfo,
  },
  data() {
    return {
      style,
      search: '',
      subject: 'all',
      status: 'all',
      semester: '7',
      detailsOpen: false,
      selectedAssignment: null as GradeAssignment | null,
    }
  },
  computed: {
    semesterAssignments(): GradeAssignment[] {
      return gradeAssignments.filter((assignment) => assignment.semester === this.semester)
    },
    subjectOptions() {
      return [
        { label: 'Все дисциплины', value: 'all' },
        ...[...new Set(gradeAssignments.map((assignment) => assignment.subject))].map(
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
  methods: {
    openAssignment(assignment: GradeAssignment) {
      this.selectedAssignment = assignment
      this.detailsOpen = true
    },
    resetFilters() {
      this.search = ''
      this.subject = 'all'
      this.status = 'all'
      this.semester = '7'
    },
  },
})
</script>
