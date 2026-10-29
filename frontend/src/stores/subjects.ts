import { defineStore } from 'pinia'
import { subjects, logicTopics } from '@/pages/subjects/subjects.data'
import type { Subject, SubjectDetails } from '@/pages/subjects/subjects.types'

export const useSubjectsStore = defineStore('subjects', {
  state: () => ({
    subjects: [] as Subject[],
    currentSubject: null as SubjectDetails | null,
  }),
  getters: {
    assignments: (state) => state.currentSubject?.topics ?? [],
  },
  actions: {
    fetchSubjects() {
      this.subjects = subjects
    },
    fetchSubject(id: number | string) {
      const key = String(id) === 'logic' ? 'log' : String(id)
      const subject = subjects.find((item) => String(item.id) === key || item.value === key)
      this.currentSubject = subject
        ? { ...subject, topics: subject.value === 'log' ? logicTopics : [] }
        : null
    },
  },
})
