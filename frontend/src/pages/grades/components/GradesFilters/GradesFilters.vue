<template>
  <div :class="style.filters">
    <StyleInput
      :model-value="search"
      @update:model-value="$emit('update:search', $event)"
      aria-label="Поиск задания"
      placeholder="Поиск задания"
      full-width
    >
      <template #prefix><PhMagnifyingGlass :size="20" /></template>
    </StyleInput>
    <StyleSelect
      :model-value="subject"
      @update:model-value="$emit('update:subject', $event)"
      :options="subjectOptions"
      aria-label="Дисциплина"
      full-width
    />
    <StyleSelect
      :model-value="status"
      @update:model-value="$emit('update:status', $event)"
      :options="statusOptions"
      aria-label="Статус"
      full-width
    />
    <StyleSelect
      :model-value="semester"
      @update:model-value="$emit('update:semester', $event)"
      :options="semesterOptions"
      aria-label="Семестр"
      full-width
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhMagnifyingGlass } from '@phosphor-icons/vue'
import StyleInput from '@/components/UI/inputs/Input/Input.vue'
import StyleSelect from '@/components/UI/inputs/Select/Select.vue'
import { gradeStatuses } from '../../grades.data'
import style from './GradesFilters.module.scss'

export default defineComponent({
  name: 'GradesFilters',
  components: { StyleInput, StyleSelect, PhMagnifyingGlass },
  props: {
    search: { type: String, required: true },
    subject: { type: String, required: true },
    status: { type: String, required: true },
    semester: { type: String, required: true },
    subjectOptions: { type: Array as PropType<FloatingPopupItem[]>, required: true },
  },
  emits: ['update:search', 'update:subject', 'update:status', 'update:semester'],
  data() {
    return {
      style,
      semesterOptions: Array.from({ length: 8 }, (_, index) => ({
        label: `${8 - index} семестр`,
        value: String(8 - index),
      })),
      statusOptions: [
        { label: 'Все статусы', value: 'all' },
        ...Object.entries(gradeStatuses).map(([value, label]) => ({ label, value })),
      ],
    }
  },
})
</script>
