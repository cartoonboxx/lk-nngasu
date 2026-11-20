<template>
  <section :class="style.subjectPage">
    <SectionHeader header="Предметы" paragraph="Доступные вам дисциплины" />

    <section :class="style.pageInput">
      <StyleInput v-model="searchQuery" placeholder="Поиск дисциплины" :full-width="true">
        <template #prefix>
          <PhMagnifyingGlass :size="20" />
        </template>
      </StyleInput>

      <StyleSelect v-model="selectedSemester" :options="semesterOptions" />
    </section>

    <PageLoader v-if="$subjectsStore.loading" />
    <p v-else-if="$subjectsStore.error">
      <button @click="$subjectsStore.fetchSubjects()">Повторить</button>
    </p>
    <p v-else-if="!filteredSubjects.length">Предметы не найдены</p>
    <section v-else :class="style.pageSubjectSemester">
      <RouterLink
        v-press
        v-for="subject in filteredSubjects"
        :key="subject.id"
        :class="[baseButtonStyle.base, baseButtonStyle.md, buttonStyle.secondary]"
        :to="{ name: 'course', params: { id: subject.id } }"
      >
        {{ subject.label }}
      </RouterLink>
    </section>
  </section>
</template>

<script lang="ts">
import type { Subject } from './subjects.types'
import { RouterLink } from 'vue-router'
import { defineComponent } from 'vue'
import { pressFeedback } from '@/directives/pressFeedback'
import style from './SubjectPage.module.scss'
import { PhMagnifyingGlass } from '@phosphor-icons/vue'

import StyleInput from '@/components/UI/inputs/Input/Input.vue'
import StyleSelect from '@/components/UI/inputs/Select/Select.vue'
import buttonStyle from '@/components/UI/buttons/Button/Button.module.scss'
import baseButtonStyle from '@/components/UI/buttons/BaseButton/BaseButton.module.scss'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import PageLoader from '@/components/PageLoader/PageLoader.vue'

export default defineComponent({
  name: 'SubjectsPage',
  directives: { press: pressFeedback },

  components: {
    SectionHeader,
    StyleInput,
    StyleSelect,
    PhMagnifyingGlass,
    RouterLink,
    PageLoader,
  },

  data() {
    return {
      style,
      buttonStyle,
      baseButtonStyle,
      searchQuery: '',
      selectedSemester: 'all',
    }
  },
  created() {
    this.$subjectsStore.fetchSubjects()
  },
  computed: {
    semesterOptions() {
      return [
        { label: 'Все семестры', value: 'all' },
        ...new Map(
          this.$subjectsStore.subjects.map((subject) => [
            subject.semester,
            { value: subject.semester, label: subject.semesterLabel },
          ]),
        ).values(),
      ]
    },
    filteredSubjects(): Subject[] {
      const query = this.searchQuery.trim().toLowerCase()

      return this.$subjectsStore.subjects.filter((subject) => {
        const matchesSemester =
          this.selectedSemester === 'all' || subject.semester === this.selectedSemester

        const matchesSearch = subject.label.toLowerCase().includes(query)

        return matchesSemester && matchesSearch
      })
    },
  },
})
</script>
