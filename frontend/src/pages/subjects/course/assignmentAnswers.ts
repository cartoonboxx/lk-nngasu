import { reactive } from 'vue'
import type { AssignmentAnswer } from './course.types'

// Temporary session storage until the course API is available. Files stay in memory.
export const assignmentAnswers = reactive<Record<string, AssignmentAnswer | undefined>>({})
