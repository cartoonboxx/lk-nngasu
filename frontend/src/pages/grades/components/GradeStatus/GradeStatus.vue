<template>
  <StatusBadge :variant="variants[status]">{{ gradeStatuses[status] }}</StatusBadge>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { gradeStatuses } from '../../grades.data'
import type { GradeStatus } from '../../grades.types'
import StatusBadge from '@/components/UI/StatusBadge/StatusBadge.vue'

export default defineComponent({
  name: 'GradeStatus',
  components: { StatusBadge },
  props: { status: { type: String as PropType<GradeStatus>, required: true } },
  data() {
    return {
      gradeStatuses,
      variants: {
        reviewed: 'success',
        pending: 'info',
        'not-submitted': 'warning',
      } satisfies Record<GradeStatus, StatusBadgeVariant>,
    }
  },
})
</script>
