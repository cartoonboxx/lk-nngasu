<template>
  <AdminDialog
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    title="Копирование шаблона"
  >
    <div :class="style.form">
      <label :class="style.field">
        Группа
        <Select
          :model-value="String(selectedGroup)"
          :options="groupOptions"
          @update:model-value="selectedGroup = Number($event)"
          full-width
        />
      </label>
      <p :class="style.muted">
        Источник:
        {{ sourceSemesterName }}, {{ sourceWeek === 1 ? 'первая' : 'вторая' }} неделя.
      </p>
      <label :class="style.field">
        Куда скопировать
        <Select
          :model-value="String(targetSemester)"
          :options="targetSemesterOptions"
          @update:model-value="targetSemester = Number($event)"
          full-width
        />
      </label>
      <p :class="style.notice">
        Существующие занятия сохранятся. Для другого семестра должны быть заранее назначены те же
        дисциплины и нагрузка. Копируется выбранная неделя.
      </p>
      <template v-if="copyPreview">
        <p>Новых занятий: {{ copyPreview.candidates.length }}</p>
        <p
          v-if="!copyPreview.candidates.length && !copyPreview.conflicts.length"
          :class="style.muted"
        >
          В источнике нет занятий.
        </p>
        <p v-for="(conflict, index) in copyPreview.conflicts" :key="index" :class="style.error">
          {{ conflict }}
        </p>
        <ul v-if="copyPreview.candidates.length" :class="style.copyPreviewList">
          <li v-for="item in copyPreview.candidates" :key="item.id">
            {{ days[item.day_of_week - 1] }}, {{ item.starts_at }}–{{ item.ends_at }} · аудитория
            {{ roomNames[item.room_id] }}
          </li>
        </ul>
      </template>
      <div :class="style.dialogFooter">
        <Button variant="ghost" @click="$emit('update:modelValue', false)"> Отмена </Button>
        <Button v-if="!copyPreview" variant="secondary" @click="previewCopy">
          Предварительная проверка
        </Button>
        <Button
          v-else
          :disabled="Boolean(copyPreview.conflicts.length) || !copyPreview.candidates.length"
          :loading="$adminStore.saving"
          @click="applyCopy"
        >
          Подтвердить копирование
        </Button>
      </div>
    </div>
  </AdminDialog>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminDialog from './AdminDialog.vue'
import { adminDays, type AdminLesson } from '../admin.types'
import style from '../Admin.module.scss'
export default defineComponent({
  components: { Select, Button, AdminDialog },
  props: {
    modelValue: Boolean,
    group: {
      type: Number,
      required: true,
    },
    semester: {
      type: Number,
      required: true,
    },
    week: {
      type: Number,
      required: true,
    },
  },
  emits: {
    'update:modelValue': (value: boolean) => typeof value === 'boolean',
    copied: (target: { semester: number; week: number }) => Boolean(target.semester && target.week),
  },
  data() {
    return {
      style,
      days: adminDays,
      selectedGroup: this.group,
      targetSemester: this.semester,
      copyPreview: null as { candidates: AdminLesson[]; conflicts: string[] } | null,
    }
  },
  computed: {
    sourceSemesterName(): string {
      return (
        this.$adminStore.data?.catalogs.semesters.find((item) => item.id === this.semester)?.name ??
        ''
      )
    },
    sourceWeek(): number {
      return this.targetSemester === this.semester ? 1 : this.week
    },
    roomNames(): Record<number, string> {
      return Object.fromEntries(
        (this.$adminStore.data?.catalogs.rooms ?? []).map((item) => [item.id, item.name]),
      )
    },

    groupOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.groups.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },
    targetSemesterOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.semesters.filter((item) => !item.archived) ?? []).map(
        (item) => ({
          value: String(item.id),
          label: item.id === this.semester ? 'На вторую неделю текущего семестра' : item.name,
        }),
      )
    },
  },
  watch: {
    selectedGroup() {
      this.copyPreview = null
    },
    targetSemester() {
      this.copyPreview = null
    },
  },
  methods: {
    async previewCopy() {
      this.copyPreview = await this.$adminStore.copyLessons(
        this.selectedGroup,
        this.semester,
        this.targetSemester,
        this.sourceWeek,
      )
      if (this.copyPreview?.conflicts.length) {
        this.$adminStore.success = ''
        this.$adminStore.error = ''
      }
    },
    async applyCopy() {
      const result = await this.$adminStore.copyLessons(
        this.selectedGroup,
        this.semester,
        this.targetSemester,
        this.sourceWeek,
        false,
      )
      if (result && !result.conflicts.length && result.candidates.length) {
        this.$emit('copied', {
          semester: this.targetSemester,
          week: this.targetSemester === this.semester ? 2 : this.week,
        })
        this.$emit('update:modelValue', false)
      } else if (result) this.copyPreview = result
    },
  },
})
</script>
