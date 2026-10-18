<template :class="style.page">
  <section :class="style.pageHeader">
    <h1>Предметы</h1>
    <p :class="style.pageDescription">Дисциплины текущего семестра</p>
  </section>

  <section :class="style.pageInput">
    <StyleInput v-model="searchQuery" placeholder="Поиск дисциплины" :full-width="true">
      <template #prefix>
        <PhMagnifyingGlass :size="20" aria-hidden="true" />
      </template>
    </StyleInput>

    <StyleSelect v-model="selectedSemester" :options="semesterOptions" />
  </section>

  <section :class="style.pageSubjectSemestr">
    <Button
      v-for="subject in filteredSubjects"
      :key="subject.value"
      :variant="selectedSubject === subject.value ? 'primary' : 'secondary'"
      @click="selectedSubject = subject.value"
    >
      {{ subject.label }}
    </Button>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './SubjectPage.module.scss'
import { PhMagnifyingGlass } from '@phosphor-icons/vue'

import StyleInput from '@/components/UI/inputs/Input/Input.vue'
import StyleSelect from '@/components/UI/inputs/Select/Select.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'

interface Subject {
  label: string
  value: string
  semester: string
}

export default defineComponent({
  name: 'SubjectsPage',

  components: {
    StyleInput,
    StyleSelect,
    PhMagnifyingGlass,
  },

  data() {
    return {
      style,

      searchQuery: '',
      selectedSemester: '1',
      selectedSubject: 'programming',

      semesterOptions: [
        { label: '1 семестр', value: '1' },
        { label: '2 семестр', value: '2' },
      ],

      subjects: [
        {
          label: 'Программирование',
          value: 'programming',
          semester: '1',
        },
        {
          label: 'Математика',
          value: 'math',
          semester: '1',
        },
        {
          label: 'Базы данных',
          value: 'database',
          semester: '1',
        },
        {
          label: 'Веб-разработка',
          value: 'web',
          semester: '2',
        },
        {
          label: 'Компьютерные сети',
          value: 'cet',
          semester: '2',
        },
        {
          label: 'ОС',
          value: 'oc',
          semester: '1',
        },
        {
          label: 'Тестирование',
          value: 'test',
          semester: '1',
        },
        {
          label: 'Качество ПО',
          value: 'po',
          semester: '1',
        },
        {
          label: 'Логическое программирование',
          value: 'log',
          semester: '2',
        },
        {
          label: 'Веб',
          value: 'web2',
          semester: '2',
        },
      ] as Subject[],
    }
  },
  computed: {
    filteredSubjects(): Subject[] {
      const query = this.searchQuery.trim().toLowerCase()

      return this.subjects.filter((subject) => {
        const matchesSemester = subject.semester === this.selectedSemester

        const matchesSearch = subject.label.toLowerCase().includes(query)

        return matchesSemester && matchesSearch
      })
    },
  },

  watch: {
    filteredSubjects(subjects: Subject[]) {
      const selectedExists = subjects.some((subject) => subject.value === this.selectedSubject)

      if (!selectedExists) {
        this.selectedSubject = subjects[0]?.value ?? ''
      }
    },
  },
})
</script>
