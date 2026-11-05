<template>
  <div :class="style.form">
    <div :class="style.grid">
      <TextInput v-model="user.last_name" label="Фамилия" required :maxlength="50" full-width />
      <TextInput v-model="user.first_name" label="Имя" required :maxlength="50" full-width />
      <TextInput v-model="user.middle_names" label="Отчество" :maxlength="50" full-width />
      <TextInput
        v-model="user.email"
        label="Электронная почта"
        type="email"
        required
        :disabled="registration"
        full-width
      />
      <TextInput
        v-model="user.passport"
        label="Паспорт (серия и номер)"
        name="passport"
        placeholder="45 09 123456"
        :maxlength="12"
        required
        full-width
      />
      <TextInput
        v-model="user.snils"
        label="СНИЛС"
        name="snils"
        placeholder="112-233-445 95"
        :maxlength="14"
        required
        full-width
      />
    </div>
    <template v-if="registration">
      <label :class="style.field">
        Учебная роль
        <Select v-model="role" :options="roleOptions" full-width />
      </label>
    </template>
    <div v-else :class="style.actions">
      <Checkbox v-model="isStudent" label="Студент" />
      <Checkbox v-model="isTeacher" label="Преподаватель" />
    </div>
    <div v-if="user.student" :class="style.grid">
      <label :class="style.field">
        Учебная группа
        <Select
          :model-value="String(user.student.group_id)"
          :options="groupOptions"
          @update:model-value="user.student.group_id = Number($event)"
          full-width
          placeholder="Выберите группу"
        />
      </label>
      <TextInput
        v-model="user.student.record_book_number"
        label="Номер зачётной книжки"
        :maxlength="50"
        required
        full-width
      />
      <Checkbox v-model="user.student.is_headman" label="Староста группы" />
    </div>
    <label v-if="user.teacher" :class="style.field">
      Кафедра
      <Select
        :model-value="String(user.teacher.department_id)"
        :options="departmentOptions"
        @update:model-value="user.teacher.department_id = Number($event)"
        full-width
        placeholder="Выберите кафедру"
      />
    </label>
    <p v-if="!registration" :class="style.muted">
      Административный доступ: {{ user.is_admin ? 'есть' : 'нет' }}. Назначается вне панели.
    </p>
  </div>
</template>
<script lang="ts">
import Select from '@/components/UI/inputs/Select/Select.vue'
import { defineComponent, type PropType } from 'vue'
import type { AdminUser } from '../admin.types'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Checkbox from '@/components/UI/inputs/Checkbox/Checkbox.vue'
import style from '../Admin.module.scss'
export default defineComponent({
  components: { Select, TextInput, Checkbox },
  props: {
    modelValue: { type: Object as PropType<AdminUser>, required: true },
    registration: Boolean,
  },
  data() {
    return {
      style,
      roleOptions: [
        { value: '', label: 'Выберите роль вручную' },
        { value: 'student', label: 'Студент' },
        { value: 'teacher', label: 'Преподаватель' },
      ],
    }
  },
  computed: {
    role: {
      get(): string {
        return this.user.student ? 'student' : this.user.teacher ? 'teacher' : ''
      },
      set(value: string) {
        this.user.student =
          value === 'student' ? { group_id: 0, record_book_number: '', is_headman: false } : null
        this.user.teacher = value === 'teacher' ? { department_id: 0 } : null
      },
    },
    isStudent: {
      get(): boolean {
        return Boolean(this.user.student)
      },
      set(value: boolean) {
        this.user.student = value
          ? { group_id: 0, record_book_number: '', is_headman: false }
          : null
      },
    },
    isTeacher: {
      get(): boolean {
        return Boolean(this.user.teacher)
      },
      set(value: boolean) {
        this.user.teacher = value ? { department_id: 0 } : null
      },
    },

    groupOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.groups.filter((item) => !item.archived) ?? []).map(
        (item) => ({ value: String(item.id), label: item.name }),
      )
    },
    departmentOptions(): FloatingPopupItem[] {
      return (
        this.$adminStore.data?.catalogs.departments.filter((item) => !item.archived) ?? []
      ).map((item) => ({ value: String(item.id), label: item.name }))
    },

    user(): AdminUser {
      return this.modelValue
    },
  },
})
</script>
