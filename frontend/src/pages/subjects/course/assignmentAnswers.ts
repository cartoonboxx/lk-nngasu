import { reactive } from 'vue'
import type { AssignmentAnswer } from './components/AssignmentModal/assignment.types'

// Temporary session storage until the course API is available. Files stay in memory.
export const assignmentAnswers = reactive<Record<string, AssignmentAnswer | undefined>>({})
