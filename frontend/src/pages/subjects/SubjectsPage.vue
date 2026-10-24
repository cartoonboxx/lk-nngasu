<template>
  <section :class="style.subjectPage">
    <SectionHeader header="Предметы" paragraph="Дисциплины текущего семестра" />

    <section :class="style.pageInput">
      <StyleInput v-model="searchQuery" placeholder="Поиск дисциплины" :full-width="true">
        <template #prefix>
          <PhMagnifyingGlass :size="20" />
        </template>
      </StyleInput>

      <StyleSelect v-model="selectedSemester" :options="semesterOptions" />
    </section>

    <section :class="style.pageSubjectSemester">
      <Button
        v-for="subject in filteredSubjects"
        :key="subject.id"
        variant="secondary"
        @click="openSubject(subject.id)"
      >
        {{ subject.label }}
      </Button>
    </section>
  </section>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import style from './SubjectPage.module.scss'
import { PhMagnifyingGlass } from '@phosphor-icons/vue'

import StyleInput from '@/components/UI/inputs/Input/Input.vue'
import StyleSelect from '@/components/UI/inputs/Select/Select.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'

interface Subject {
  id: number
  label: string
  value: string
  semester: string
}

export default defineComponent({
  name: 'SubjectsPage',

  components: {
    SectionHeader,
    StyleInput,
    StyleSelect,
    PhMagnifyingGlass,
    Button,
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
          id: 1,
          label: 'Программирование',
          value: 'programming',
          semester: '1',
        },
        {
          id: 2,
          label: 'Математика',
          value: 'math',
          semester: '1',
        },
        {
          id: 3,
          label: 'Базы данных',
          value: 'database',
          semester: '1',
        },
        {
          id: 4,
          label: 'Веб-разработка',
          value: 'web',
          semester: '2',
        },
        {
          id: 5,
          label: 'Компьютерные сети',
          value: 'cet',
          semester: '2',
        },
        {
          id: 6,
          label: 'ОС',
          value: 'oc',
          semester: '1',
        },
        {
          id: 7,
          label: 'Тестирование',
          value: 'test',
          semester: '1',
        },
        {
          id: 8,
          label: 'Качество ПО',
          value: 'po',
          semester: '1',
        },
        {
          id: 9,
          label: 'Логическое программирование',
          value: 'log',
          semester: '2',
        },
        {
          id: 10,
          label: 'Веб',
          value: 'web2',
          semester: '2',
        },
      ] as Subject[],
    }
  },
  methods: {
    openSubject(idSubject: number) {
      this.$router.push(`/subject/${idSubject}`)
    },
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
