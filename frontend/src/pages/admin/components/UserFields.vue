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
        mask="## ## ######"
        :maxlength="12"
        required
        full-width
      />
      <TextInput
        v-model="user.snils"
        label="СНИЛС"
        name="snils"
        placeholder="112-233-445 95"
        mask="###-###-### ##"
        :maxlength="14"
        required
        full-width
      />
    </div>
    <label v-if="!user.is_admin" :class="style.field">
      Учебная роль
      <Select v-model="role" :options="roleOptions" full-width />
    </label>
    <div v-if="user.student" :class="style.grid">
      <TextInput
        v-model="groupSearch"
        label="Учебная группа"
        :options="groupOptions"
        @update:model-value="user.student.group_id = 0"
        @select="user.student.group_id = Number($event.value)"
        full-width
        required
        placeholder="Введите название группы"
      />
      <TextInput
        v-model="user.student.record_book_number"
        label="Номер зачётной книжки"
        :maxlength="50"
        full-width
      />
      <Checkbox v-model="user.student.is_headman" label="Староста группы" />
    </div>
    <TextInput
      v-if="user.teacher"
      v-model="departmentSearch"
      label="Кафедра"
      :options="departmentOptions"
      @update:model-value="user.teacher.department_id = 0"
      @select="user.teacher.department_id = Number($event.value)"
      full-width
      required
      placeholder="Введите название кафедры"
    />
    <p v-if="user.student && !user.student.group_id" :class="style.muted">
      Выберите учебную группу из списка.
    </p>
    <p v-if="user.teacher && !user.teacher.department_id" :class="style.muted">
      Выберите кафедру из списка.
    </p>
    <p v-if="!registration" :class="style.muted">
      Административный доступ: {{ user.is_admin ? 'есть' : 'нет' }}. Управляется владельцем в
      разделе «Администраторы».
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
      groupSearch:
        this.$adminStore.data?.catalogs.groups.find(
          (item) => item.id === this.modelValue.student?.group_id,
        )?.name ?? '',
      departmentSearch:
        this.$adminStore.data?.catalogs.departments.find(
          (item) => item.id === this.modelValue.teacher?.department_id,
        )?.name ?? '',
    }
  },
  computed: {
    roleOptions(): FloatingPopupItem[] {
      return [
        { value: '', label: 'Выберите роль' },
        { value: 'student', label: 'Студент' },
        { value: 'teacher', label: 'Преподаватель' },
      ]
    },
    role: {
      get(): string {
        return this.user.is_admin
          ? 'admin'
          : this.user.student
            ? 'student'
            : this.user.teacher
              ? 'teacher'
              : ''
      },
      set(value: string) {
        if (value === this.role) return
        this.user.student =
          value === 'student' ? { group_id: 0, record_book_number: '', is_headman: false } : null
        this.user.teacher = value === 'teacher' ? { department_id: 0 } : null
        this.groupSearch = ''
        this.departmentSearch = ''
      },
    },
    groupOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.groups ?? []).map((item) => ({
        value: String(item.id),
        label: item.name,
      }))
    },
    departmentOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.departments ?? []).map((item) => ({
        value: String(item.id),
        label: item.name,
      }))
    },

    user(): AdminUser {
      return this.modelValue
    },
  },
})
</script>
