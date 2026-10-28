<template>
  <span :class="[style.status, style[status]]">{{ gradeStatuses[status] }}</span>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { gradeStatuses } from '../../grades.data'
import type { GradeStatus } from '../../grades.types'
import style from './GradeStatus.module.scss'

export default defineComponent({
  name: 'GradeStatus',
  props: { status: { type: String as PropType<GradeStatus>, required: true } },
  data() {
    return { style, gradeStatuses }
  },
})
</script>
