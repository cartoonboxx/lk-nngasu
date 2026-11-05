<template>
  <section :class="style.page">
    <SectionHeader header="Расписание" paragraph="Повторяющийся двухнедельный шаблон" nm>
      <template #actions>
        <Button @click="openLesson()">
          <PhPlus :size="18" />
          Добавить занятие
        </Button>
      </template>
    </SectionHeader>
    <div :class="style.toolbar">
      <label :class="style.field">
        Семестр
        <Select
          :model-value="String(semester)"
          :options="semesterOptions"
          @update:model-value="semester = Number($event)"
          full-width
        />
      </label>
      <label :class="style.field">
        Вид
        <Select v-model="scope" :options="scopeTypeOptions" full-width />
      </label>
      <label :class="style.field">
        {{ scope === 'group' ? 'Группа' : scope === 'teacher' ? 'Преподаватель' : 'Аудитория' }}
        <Select
          :model-value="String(scopeId)"
          :options="scopeFilterOptions"
          @update:model-value="scopeId = Number($event)"
          full-width
        />
      </label>
      <label :class="style.field">
        Неделя
        <Select
          :model-value="String(week)"
          :options="weekFilterOptions"
          @update:model-value="week = Number($event)"
          full-width
        />
      </label>
    </div>
    <div :class="style.toolbar">
      <div :class="style.tabs">
        <StyleTab text="Недельная сетка" :active="view === 'grid'" @select="view = 'grid'" />
        <StyleTab text="Список" :active="view === 'list'" @select="view = 'list'" />
      </div>
    </div>
    <div v-if="view === 'grid'" :class="style.panel">
      <div :class="style.tableScroll">
        <table :class="[style.table, style.scheduleTable]">
          <thead>
            <tr>
              <th>Пара</th>
              <th v-for="day in days" :key="day">{{ day }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in scheduleGrid" :key="row.number">
              <th>
                <strong>{{ row.number }} пара</strong>
                <small>{{ row.start }}–{{ row.end }}</small>
              </th>
              <td
                v-for="(lessons, day) in row.days"
                :key="day"
                :class="style.scheduleCell"
                @click="!lessons.length && openLesson(undefined, day + 1, row.number)"
              >
                <button
                  v-for="item in lessons"
                  :key="item.id"
                  :class="[style.lesson, style[`lesson_${item.lesson_type}`]]"
                  @click.stop="openLesson(item)"
                >
                  <strong>{{ item.subject }}</strong>
                  <small :class="style.lessonType">{{ lessonLabels[item.lesson_type] }}</small>
                  <span>{{ item.group }} · {{ item.room }}</span>
                  <small>{{ item.teacher }}</small>
                </button>
                <button
                  v-if="!lessons.length"
                  type="button"
                  :class="style.addLesson"
                  :title="`${days[day]}, ${row.number} пара — добавить занятие`"
                  @click.stop="openLesson(undefined, day + 1, row.number)"
                >
                  <PhPlus :size="18" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
    <AdminTable
      v-else
      :columns="columns"
      :total="filtered.length"
      :reset-key="`${scope}-${scopeId}-${week}-${semester}`"
      v-slot="{ start, end }"
    >
      <tr v-for="item in filtered.slice(start, end)" :key="item.id">
        <td>
          {{ days[item.day_of_week - 1] }}<strong>{{ item.starts_at }}–{{ item.ends_at }}</strong>
        </td>
        <td>
          {{ item.subject }}<br /><small>{{ lessonLabels[item.lesson_type] }}</small>
        </td>
        <td>{{ item.group }}</td>
        <td>{{ item.teacher }}</td>
        <td>{{ item.room }}</td>
        <td>
          <Button size="sm" variant="ghost" @click="openLesson(item)">Изменить</Button>
        </td>
      </tr>
    </AdminTable>
    <p :class="style.muted">
      Пересечения проверяются локально по интервалам времени. Совместные потоки и изменения на
      конкретные даты требуют отдельной модели и пока недоступны.
    </p>
    <AdminDialog v-model="opened" :title="draft.id ? 'Изменить занятие' : 'Новое занятие'">
      <form :class="style.form" @submit.prevent="save">
        <div :class="style.grid">
          <label :class="style.field">
            Группа<Select
              :model-value="String(formGroup)"
              :options="groupOptions"
              @update:model-value="formGroup = Number($event)"
              full-width
              :disabled="Boolean(draft.id)"
            />
          </label>
          <label :class="style.field">
            Семестр
            <Select
              :model-value="String(formSemester)"
              :options="semesterOptions"
              @update:model-value="formSemester = Number($event)"
              full-width
              :disabled="Boolean(draft.id)"
            />
          </label>
        </div>
        <label :class="style.field">
          Дисциплина · преподаватель · тип занятия
          <Select
            :model-value="String(draft.workload_id)"
            :options="lessonOptions"
            @update:model-value="draft.workload_id = Number($event)"
            full-width
            placeholder="Выберите назначенную нагрузку"
            :disabled="Boolean(draft.id)"
          />
        </label>
        <p v-if="!lessonOptions.length" :class="style.notice">
          У группы нет нагрузки в этом семестре. Сначала назначьте дисциплины и преподавателей в
          разделе «Дисциплины и нагрузка».
        </p>
        <div :class="style.grid">
          <label :class="style.field">
            Неделя
            <Select
              :model-value="String(draft.week_number)"
              :options="weekOptions"
              @update:model-value="draft.week_number = Number($event)"
              full-width
            />
          </label>
          <label :class="style.field">
            День
            <Select
              :model-value="String(draft.day_of_week)"
              :options="dayOptions"
              @update:model-value="draft.day_of_week = Number($event)"
              full-width
            />
          </label>
          <label :class="style.field">
            Пара
            <Select v-model="selectedPeriod" :options="periodOptions" full-width />
          </label>
        </div>
        <label :class="style.field">
          Аудитория
          <Select
            :model-value="String(draft.room_id)"
            :options="roomOptions"
            @update:model-value="draft.room_id = Number($event)"
            full-width
            placeholder="Выберите аудиторию"
          />
        </label>
        <p v-if="conflicts.length" :class="style.error">{{ conflicts.join(' ') }}</p>
        <div :class="style.dialogFooter">
          <Button v-if="draft.id" variant="danger" @click="confirmOpen = true">Удалить</Button>
          <Button variant="ghost" @click="opened = false">Отмена</Button>
          <Button
            type="submit"
            :disabled="!draft.workload_id || Boolean(conflicts.length)"
            :loading="$adminStore.saving"
          >
            Сохранить
          </Button>
        </div>
      </form>
    </AdminDialog>
    <AdminDialog v-model="confirmOpen" title="Удалить занятие из шаблона?">
      <p>
        Занятие будет удалено из повторяющегося расписания выбранной недели. Это изменение шаблона,
        а не отмена пары на одну дату.
      </p>
      <div :class="style.dialogFooter">
        <Button variant="ghost" @click="confirmOpen = false">Отмена</Button>
        <Button variant="danger" :loading="$adminStore.saving" @click="remove">
          Удалить занятие
        </Button>
      </div>
    </AdminDialog>
  </section>
</template>
<script lang="ts">
import { periods } from '@/pages/schedule/schedule.periods'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import { defineComponent } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import { adminDays, lessonLabels, type AdminLesson, type LessonKind } from './admin.types'
import { lessonConflicts } from '@/stores/admin'
import style from './Admin.module.scss'
interface ScheduleRow extends AdminLesson {
  subject: string
  group: string
  teacher: string
  room: string
  group_id: number
  teacher_id: number
  semester_id: number
  lesson_type: LessonKind
}
export default defineComponent({
  components: {
    SectionHeader,
    StyleTab,
    Select,
    PhPlus,
    Button,
    AdminTable,
    AdminDialog,
  },
  data() {
    return {
      scopeTypeOptions: [
        { value: 'group', label: 'По группе' },
        { value: 'teacher', label: 'По преподавателю' },
        { value: 'room', label: 'По аудитории' },
      ],
      weekOptions: [
        { value: '1', label: 'Первая' },
        { value: '2', label: 'Вторая' },
      ],
      columns: ['День и время', 'Дисциплина', 'Группа', 'Преподаватель', 'Аудитория', 'Действие'],
      weekFilterOptions: [
        { value: '1', label: 'Первая неделя' },
        { value: '2', label: 'Вторая неделя' },
      ],
      style,
      periodOptions: periods.map((period, index) => ({
        value: String(index + 1),
        label: `${index + 1} пара · ${period.start}–${period.end}`,
      })),
      lessonLabels,
      days: adminDays,
      semester: 1,
      scope: 'group',
      scopeId: 1,
      week: 1,
      view: 'grid',
      opened: false,
      confirmOpen: false,
      formGroup: 1,
      formSemester: 1,
      draft: {
        id: 0,
        workload_id: 0,
        week_number: 1,
        day_of_week: 1,
        starts_at: periods[0]!.start,
        ends_at: periods[0]!.end,
        room_id: 0,
      } as AdminLesson,
    }
  },
  computed: {
    semesterOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.semesters.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },
    groupOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.groups.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },
    dayOptions(): FloatingPopupItem[] {
      return (this.days ?? []).map((day, index) => ({ value: String(index + 1), label: day }))
    },
    roomOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.rooms.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },

    selectedPeriod: {
      get(): string {
        return String(
          periods.findIndex(
            (period, index) =>
              this.draft.starts_at >= period.start &&
              this.draft.starts_at < (periods[index + 1]?.start ?? '24:00'),
          ) + 1,
        )
      },
      set(value: string) {
        const period = periods[Number(value) - 1]
        if (!period) return
        this.draft.starts_at = period.start
        this.draft.ends_at = period.end
      },
    },
    scheduleGrid() {
      return periods.map((period, index) => ({
        ...period,
        number: index + 1,
        days: this.days.map((_, day) =>
          this.filtered.filter(
            (lesson) =>
              lesson.day_of_week === day + 1 &&
              lesson.starts_at >= period.start &&
              lesson.starts_at < (periods[index + 1]?.start ?? '24:00'),
          ),
        ),
      }))
    },
    scopeFilterOptions(): FloatingPopupItem[] {
      const data = this.$adminStore.data
      if (!data) return []
      const entries =
        this.scope === 'group'
          ? data.catalogs.groups
          : this.scope === 'room'
            ? data.catalogs.rooms
            : data.users
                .filter((item) => item.teacher)
                .map((item) => ({ id: item.id, name: `${item.last_name} ${item.first_name}` }))
      return [
        { value: '0', label: 'Все' },
        ...entries.map((item) => ({ value: String(item.id), label: item.name })),
      ]
    },
    filtered(): ScheduleRow[] {
      const data = this.$adminStore.data
      if (!data) return []
      return data.lessons
        .map((item) => {
          const workload = data.workloads.find((value) => value.id === item.workload_id)!
          const course = data.courses.find((value) => value.id === workload.course_id)!
          const teacher = data.users.find((value) => value.id === workload.teacher_id)
          return {
            ...item,
            subject:
              data.catalogs.subjects.find((value) => value.id === course.subject_id)?.name ?? '—',
            group: data.catalogs.groups.find((value) => value.id === course.group_id)?.name ?? '—',
            teacher: `${teacher?.last_name ?? ''} ${teacher?.first_name ?? ''}`,
            room: data.catalogs.rooms.find((value) => value.id === item.room_id)?.name ?? '—',
            group_id: course.group_id,
            teacher_id: workload.teacher_id,
            semester_id: course.semester_id,
            lesson_type: workload.lesson_type,
          }
        })
        .filter(
          (item) =>
            item.semester_id === this.semester &&
            item.week_number === this.week &&
            (!this.scopeId ||
              (this.scope === 'group'
                ? item.group_id
                : this.scope === 'teacher'
                  ? item.teacher_id
                  : item.room_id) === this.scopeId),
        )
        .sort((a, b) => a.day_of_week - b.day_of_week || a.starts_at.localeCompare(b.starts_at))
    },
    lessonOptions(): FloatingPopupItem[] {
      const data = this.$adminStore.data
      if (!data) return []
      return data.workloads
        .filter(
          (item) =>
            data.courses.some(
              (course) =>
                course.id === item.course_id &&
                course.group_id === this.formGroup &&
                course.semester_id === this.formSemester,
            ) && data.users.some((user) => user.id === item.teacher_id && user.is_active),
        )
        .map((item) => {
          const course = data.courses.find((value) => value.id === item.course_id)
          const teacher = data.users.find((value) => value.id === item.teacher_id)
          return {
            value: String(item.id),
            label: `${data.catalogs.subjects.find((value) => value.id === course?.subject_id)?.name} · ${teacher?.last_name} ${teacher?.first_name} · ${lessonLabels[item.lesson_type]}`,
          }
        })
    },
    conflicts(): string[] {
      return this.$adminStore.data && this.draft.workload_id
        ? lessonConflicts(this.$adminStore.data, this.draft)
        : []
    },
  },
  watch: {
    scope() {
      this.scopeId = 0
    },
    formGroup: {
      flush: 'sync',
      handler() {
        this.draft.workload_id = 0
      },
    },
    formSemester: {
      flush: 'sync',
      handler() {
        this.draft.workload_id = 0
      },
    },
  },
  methods: {
    openLesson(item?: AdminLesson, day = 1, periodNumber = 1) {
      const period = periods[periodNumber - 1] ?? periods[0]!
      const workload = this.$adminStore.data?.workloads.find(
        (value) => value.id === item?.workload_id,
      )
      const course = this.$adminStore.data?.courses.find(
        (value) => value.id === workload?.course_id,
      )
      this.formGroup =
        course?.group_id ??
        (this.scope === 'group' && this.scopeId
          ? this.scopeId
          : Number(this.groupOptions[0]?.value ?? 0))
      this.formSemester = course?.semester_id ?? this.semester
      this.draft = item
        ? {
            id: item.id,
            workload_id: item.workload_id,
            week_number: item.week_number,
            day_of_week: item.day_of_week,
            starts_at: item.starts_at,
            ends_at: item.ends_at,
            room_id: item.room_id,
          }
        : {
            id: 0,
            workload_id: 0,
            week_number: this.week,
            day_of_week: day,
            starts_at: period.start,
            ends_at: period.end,
            room_id: this.scope === 'room' ? this.scopeId : 0,
          }
      this.$adminStore.error = ''
      this.opened = true
    },
    async save() {
      const saved = await this.$adminStore.saveLesson(this.draft)
      if (saved) {
        this.opened = false
        this.semester = this.formSemester
        this.week = this.draft.week_number
      }
    },
    async remove() {
      const saved = await this.$adminStore.removeAcademic('lessons', this.draft.id)
      if (saved) {
        this.confirmOpen = false
        this.opened = false
      }
    },
  },
})
</script>
