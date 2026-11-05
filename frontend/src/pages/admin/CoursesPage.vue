<template>
  <section :class="style.page">
    <header :class="style.heading">
      <div>
        <h1>Дисциплины и нагрузка</h1>
        <p>Назначения по группам, семестрам и преподавателям</p>
      </div>
      <Button @click="openCourse()"><PhPlus :size="18" />Назначить дисциплину</Button>
    </header>
    <div :class="style.grid">
      <label :class="style.field">
        Учебная группа
        <Select
          :model-value="String(group)"
          :options="groupOptions"
          @update:model-value="group = Number($event)"
          full-width
        />
      </label>
      <label :class="style.field">
        Семестр
        <Select
          :model-value="String(semester)"
          :options="semesterOptions"
          @update:model-value="semester = Number($event)"
          full-width
        />
      </label>
    </div>
    <AdminTable
      :columns="columns"
      :total="courses.length"
      :reset-key="`${group}-${semester}`"
      v-slot="{ start, end }"
    >
      <tr v-for="item in courses.slice(start, end)" :key="item.id">
        <td>
          <strong>{{ subjectNames[item.subject_id] }}</strong>
        </td>
        <td>
          {{ workloadCounts[item.id] ?? 0 }}
          назначений
        </td>
        <td>
          <div :class="style.actions">
            <Button
              :variant="selectedId === item.id ? 'secondary' : 'ghost'"
              size="sm"
              @click="selectedId = item.id"
            >
              Нагрузка
            </Button>
            <Button size="sm" variant="ghost" @click="openCourse(item)">Изменить</Button>
            <Button
              size="sm"
              variant="ghost"
              @click="
                () => {
                  removeKind = 'courses'
                  targetId = item.id
                  confirmOpen = true
                }
              "
            >
              Исключить
            </Button>
          </div>
        </td>
      </tr>
    </AdminTable>
    <template v-if="selectedCourse">
      <header :class="style.heading">
        <div>
          <h2>
            Нагрузка:
            {{ subjectNames[selectedCourse.subject_id] }}
          </h2>
          <p>Разные типы занятий могут вести разные преподаватели</p>
        </div>
        <Button variant="secondary" @click="openWorkload()">Добавить нагрузку</Button>
      </header>
      <AdminTable :columns="workloadColumns" :total="workloads.length" v-slot="{ start, end }">
        <tr v-for="item in workloads.slice(start, end)" :key="item.id">
          <td>
            {{ teacherNames[item.teacher_id] }}
          </td>
          <td>{{ lessonLabels[item.lesson_type] }}</td>
          <td>{{ item.planned_hours }} ак. ч.</td>
          <td>{{ scheduledHours[item.id] ?? 0 }} ак. ч.</td>
          <td>
            <div :class="style.actions">
              <Button variant="ghost" size="sm" @click="openWorkload(item)">Изменить</Button>
              <Button
                variant="ghost"
                size="sm"
                @click="
                  () => {
                    removeKind = 'workloads'
                    targetId = item.id
                    confirmOpen = true
                  }
                "
              >
                Удалить
              </Button>
            </div>
          </td>
        </tr>
      </AdminTable>
      <p :class="style.muted">
        Шаблон показывает часы за две недели из расчёта 45 минут на академический час. Итоговая
        сверка с семестровым планом требует дат семестра и исключений.
      </p>
    </template>
    <p v-else-if="courses.length" :class="style.notice">
      Выберите «Нагрузка» у дисциплины, чтобы посмотреть и изменить назначения.
    </p>
    <AdminDialog
      v-model="courseOpen"
      :title="courseDraft.id ? 'Изменить назначение' : 'Назначить дисциплину'"
    >
      <form :class="style.form" @submit.prevent="saveCourse">
        <p :class="style.muted">
          {{ selectedGroupName }} ·
          {{ selectedSemesterName }}
        </p>
        <label :class="style.field">
          Дисциплина
          <Select
            :model-value="String(courseDraft.subject_id)"
            :options="subjectOptions"
            @update:model-value="courseDraft.subject_id = Number($event)"
            full-width
            placeholder="Выберите дисциплину"
          />
        </label>
        <p v-if="courseDraft.id" :class="style.muted">
          Изменение доступно, если у назначения ещё нет связанной нагрузки.
        </p>
        <div :class="style.dialogFooter">
          <Button variant="ghost" @click="courseOpen = false">Отмена</Button>
          <Button type="submit" :loading="$adminStore.saving">Сохранить</Button>
        </div>
      </form>
    </AdminDialog>
    <AdminDialog
      v-model="workloadOpen"
      :title="workloadDraft.id ? 'Изменить нагрузку' : 'Добавить нагрузку'"
    >
      <form :class="style.form" @submit.prevent="saveWorkload">
        <label :class="style.field">
          Преподаватель
          <Select
            :model-value="String(workloadDraft.teacher_id)"
            :options="teacherOptions"
            @update:model-value="workloadDraft.teacher_id = Number($event)"
            full-width
            placeholder="Выберите преподавателя"
          />
        </label>
        <label :class="style.field">
          Тип занятия
          <Select
            :model-value="workloadDraft.lesson_type"
            :options="lessonTypeOptions"
            @update:model-value="workloadDraft.lesson_type = $event as AdminWorkload['lesson_type']"
            full-width
          />
        </label>
        <label :class="style.field">
          Плановые академические часы
          <input
            v-model.number="workloadDraft.planned_hours"
            type="number"
            min="0"
            step="0.25"
            required
          />
        </label>
        <div :class="style.dialogFooter">
          <Button variant="ghost" @click="workloadOpen = false">Отмена</Button>
          <Button type="submit" :loading="$adminStore.saving">Сохранить</Button>
        </div>
      </form>
    </AdminDialog>
    <AdminDialog v-model="confirmOpen" title="Удалить учебную запись?">
      <p>
        Операция доступна только при отсутствии связанных
        {{ removeKind === 'courses' ? 'назначений нагрузки' : 'занятий в расписании' }}.
      </p>
      <div :class="style.dialogFooter">
        <Button variant="ghost" @click="confirmOpen = false">Отмена</Button>
        <Button variant="danger" :loading="$adminStore.saving" @click="remove">Удалить</Button>
      </div>
    </AdminDialog>
  </section>
</template>
<script lang="ts">
import Select from '@/components/UI/inputs/Select/Select.vue'
import { defineComponent } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import { lessonLabels, type CourseOffering, type AdminWorkload } from './admin.types'
import style from './Admin.module.scss'
export default defineComponent({
  components: { Select, PhPlus, Button, AdminTable, AdminDialog },
  data() {
    return {
      columns: ['Дисциплина', 'Нагрузка', 'Действия'],
      workloadColumns: [
        'Преподаватель',
        'Тип занятия',
        'Плановые часы',
        'В шаблоне за 2 недели',
        'Действия',
      ],
      style,
      lessonLabels,
      group: 1,
      semester: 1,
      selectedId: 0,
      courseOpen: false,
      workloadOpen: false,
      confirmOpen: false,
      targetId: 0,
      removeKind: 'courses' as 'courses' | 'workloads',
      courseDraft: { id: 0, group_id: 1, semester_id: 1, subject_id: 0 } as CourseOffering,
      workloadDraft: {
        id: 0,
        course_id: 0,
        teacher_id: 0,
        lesson_type: 'lecture',
        planned_hours: 36,
      } as AdminWorkload,
    }
  },
  computed: {
    subjectNames(): Record<number, string> {
      return Object.fromEntries(
        (this.$adminStore.data?.catalogs.subjects ?? []).map((item) => [item.id, item.name]),
      )
    },
    teacherNames(): Record<number, string> {
      return Object.fromEntries(
        (this.$adminStore.data?.users ?? [])
          .filter((item) => item.teacher)
          .map((item) => [item.id, `${item.last_name} ${item.first_name}`]),
      )
    },
    workloadCounts(): Record<number, number> {
      const counts: Record<number, number> = {}
      for (const item of this.$adminStore.data?.workloads ?? [])
        counts[item.course_id] = (counts[item.course_id] ?? 0) + 1
      return counts
    },
    selectedGroupName(): string {
      return (
        this.$adminStore.data?.catalogs.groups.find((item) => item.id === this.group)?.name ?? ''
      )
    },
    selectedSemesterName(): string {
      return (
        this.$adminStore.data?.catalogs.semesters.find((item) => item.id === this.semester)?.name ??
        ''
      )
    },

    groupOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.groups.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },
    semesterOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.semesters.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },
    subjectOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.subjects.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },
    teacherOptions(): FloatingPopupItem[] {
      return (
        this.$adminStore.data?.users.filter((item) => item.teacher && item.is_active) ?? []
      ).map((item) => ({
        value: String(item.id),
        label: `${item.last_name} ${item.first_name} ${item.middle_names}`,
      }))
    },
    lessonTypeOptions(): FloatingPopupItem[] {
      return Object.entries(this.lessonLabels).map(([key, label]) => ({
        value: String(key),
        label: label,
      }))
    },

    courses(): CourseOffering[] {
      return (
        this.$adminStore.data?.courses.filter(
          (item) => item.group_id === this.group && item.semester_id === this.semester,
        ) ?? []
      )
    },
    selectedCourse(): CourseOffering | undefined {
      return this.courses.find((item) => item.id === this.selectedId)
    },
    workloads(): AdminWorkload[] {
      return (
        this.$adminStore.data?.workloads.filter((item) => item.course_id === this.selectedId) ?? []
      )
    },
    scheduledHours(): Record<number, number> {
      const hours: Record<number, number> = {}
      for (const lesson of this.$adminStore.data?.lessons ?? []) {
        const start = Number(lesson.starts_at.slice(0, 2)) * 60 + Number(lesson.starts_at.slice(3))
        const end = Number(lesson.ends_at.slice(0, 2)) * 60 + Number(lesson.ends_at.slice(3))
        hours[lesson.workload_id] =
          Math.round(((hours[lesson.workload_id] ?? 0) + (end - start) / 45) * 100) / 100
      }
      return hours
    },
  },
  watch: {
    group() {
      this.selectedId = 0
    },
    semester() {
      this.selectedId = 0
    },
  },
  methods: {
    openCourse(item?: CourseOffering) {
      this.courseDraft = item
        ? { ...item }
        : { id: 0, group_id: this.group, semester_id: this.semester, subject_id: 0 }
      this.$adminStore.error = ''
      this.courseOpen = true
    },
    openWorkload(item?: AdminWorkload) {
      this.workloadDraft = item
        ? { ...item }
        : {
            id: 0,
            course_id: this.selectedId,
            teacher_id: 0,
            lesson_type: 'lecture',
            planned_hours: 36,
          }
      this.$adminStore.error = ''
      this.workloadOpen = true
    },
    async saveCourse() {
      const saved = await this.$adminStore.saveCourse(this.courseDraft)
      if (saved) {
        this.courseOpen = false
        this.selectedId = 0
      }
    },
    async saveWorkload() {
      const saved = await this.$adminStore.saveWorkload(this.workloadDraft)
      if (saved) {
        this.workloadOpen = false
        this.workloadDraft.id = 0
      }
    },
    async remove() {
      const saved = await this.$adminStore.removeAcademic(this.removeKind, this.targetId)
      if (saved) {
        this.confirmOpen = false
        if (this.removeKind === 'courses') this.selectedId = 0
      }
    },
  },
})
</script>
