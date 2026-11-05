<template>
  <section :class="style.page">
    <SectionHeader
      header="Пользователи"
      paragraph="Учебные роли, персональные данные и доступ к кабинету"
      nm
    >
      <template #actions>
        <Button @click="openUser()"><PhPlus :size="18" />Создать пользователя</Button>
      </template>
    </SectionHeader>
    <div :class="style.toolbar">
      <TextInput
        v-model="search"
        type="search"
        label="Поиск пользователей"
        placeholder="ФИО или электронная почта"
        full-width
      />
      <label :class="style.field">
        Доступ
        <Select v-model="active" :options="accessOptions" full-width />
      </label>
      <label :class="style.field">
        Учебная роль
        <Select v-model="role" :options="roleOptions" full-width />
      </label>
      <label :class="style.field">
        Администратор
        <Select v-model="admin" :options="adminOptions" full-width />
      </label>
    </div>
    <AdminTable
      :columns="columns"
      :total="filtered.length"
      :reset-key="`${search}-${active}-${role}-${admin}`"
      v-slot="{ start, end }"
    >
      <tr v-for="item in filtered.slice(start, end)" :key="item.id">
        <td>
          <Button variant="ghost" size="sm" :class="style.recordButton" @click="openUser(item)">
            <span :class="style.recordCopy">
              <strong>{{ item.last_name }} {{ item.first_name }} {{ item.middle_names }}</strong>
              <small>{{ item.email }}</small>
            </span>
          </Button>
        </td>
        <td>
          <span v-if="item.student">Студент · {{ groupNames[item.student.group_id] }}</span>
          <br v-if="item.student && item.teacher" />
          <span v-if="item.teacher">Преподаватель</span>
          <small v-if="!item.student && !item.teacher">Без учебной роли</small>
          <strong v-if="item.is_admin">Администратор</strong>
        </td>
        <td>
          <span :class="[style.pill, item.is_active ? style.approved : style.rejected]">
            {{ item.is_active ? 'Активен' : 'Заблокирован' }}
          </span>
        </td>
        <td>
          <div :class="style.actions">
            <Button variant="ghost" size="sm" @click="openUser(item)">Карточка</Button>
            <Button
              variant="ghost"
              size="sm"
              @click="
                () => {
                  target = item
                  confirmOpen = true
                }
              "
            >
              {{ item.is_active ? 'Заблокировать' : 'Восстановить' }}
            </Button>
          </div>
        </td>
      </tr>
    </AdminTable>
    <AdminDialog
      v-model="opened"
      :title="draft?.id ? 'Карточка пользователя' : 'Новый пользователь'"
    >
      <form v-if="draft" :class="style.form" @submit.prevent="save">
        <UserFields :key="`${draft.id}-${opened}`" v-model="draft" />
        <AdminImage
          :key="`${draft.id}-${opened}-avatar`"
          v-model="draft.avatar_url"
          label="Аватар"
        />
        <p v-if="!draft.id" :class="style.notice">
          В демо-режиме учётная запись создаётся без пароля и отправки приглашения.
        </p>
        <div :class="style.dialogFooter">
          <Button variant="ghost" @click="opened = false">Отмена</Button>
          <Button type="submit" :loading="$adminStore.saving">Сохранить</Button>
        </div>
      </form>
    </AdminDialog>
    <AdminDialog
      v-model="confirmOpen"
      :title="target?.is_active ? 'Заблокировать пользователя?' : 'Восстановить доступ?'"
    >
      <p>{{ target?.last_name }} {{ target?.first_name }} · {{ target?.email }}</p>
      <p :class="style.muted">
        {{
          target?.is_active
            ? 'Вход в кабинет будет недоступен. Учебные данные, оценки и связи пользователя сохранятся.'
            : 'Пользователь снова сможет входить в личный кабинет.'
        }}
      </p>
      <div :class="style.dialogFooter">
        <Button variant="ghost" @click="confirmOpen = false">Отмена</Button>
        <Button
          :variant="target?.is_active ? 'danger' : 'primary'"
          :loading="$adminStore.saving"
          @click="toggleAccess"
        >
          Подтвердить
        </Button>
      </div>
    </AdminDialog>
  </section>
</template>
<script lang="ts">
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import { defineComponent } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import UserFields from './components/UserFields.vue'
import AdminImage from './components/AdminImage.vue'
import type { AdminUser } from './admin.types'
import style from './Admin.module.scss'
export default defineComponent({
  components: {
    SectionHeader,
    Select,
    PhPlus,
    TextInput,
    Button,
    AdminTable,
    AdminDialog,
    UserFields,
    AdminImage,
  },
  data() {
    return {
      accessOptions: [
        { value: '', label: 'Все' },
        { value: 'active', label: 'Активен' },
        { value: 'inactive', label: 'Заблокирован' },
      ],
      roleOptions: [
        { value: '', label: 'Все роли' },
        { value: 'student', label: 'Студент' },
        { value: 'teacher', label: 'Преподаватель' },
      ],
      adminOptions: [
        { value: '', label: 'Все' },
        { value: 'yes', label: 'Да' },
        { value: 'no', label: 'Нет' },
      ],
      columns: ['Пользователь', 'Принадлежность', 'Доступ', 'Действия'],
      style,
      search: '',
      active: '',
      role: '',
      admin: '',
      opened: false,
      confirmOpen: false,
      draft: null as AdminUser | null,
      target: null as AdminUser | null,
    }
  },
  computed: {
    groupNames(): Record<number, string> {
      return Object.fromEntries(
        (this.$adminStore.data?.catalogs.groups ?? []).map((item) => [item.id, item.name]),
      )
    },

    filtered(): AdminUser[] {
      return (this.$adminStore.data?.users ?? []).filter(
        (item) =>
          `${item.last_name} ${item.first_name} ${item.middle_names} ${item.email}`
            .toLowerCase()
            .includes(this.search.toLowerCase()) &&
          (!this.active || item.is_active === (this.active === 'active')) &&
          (!this.role || Boolean(item[this.role as 'student' | 'teacher'])) &&
          (!this.admin || item.is_admin === (this.admin === 'yes')),
      )
    },
  },
  mounted() {
    const user = this.$adminStore.data?.users.find(
      (item) => item.id === Number(this.$route.query.user),
    )
    if (user) this.openUser(user)
  },
  methods: {
    openUser(item?: AdminUser) {
      this.draft = item
        ? JSON.parse(JSON.stringify(item))
        : {
            id: 0,
            last_name: '',
            first_name: '',
            middle_names: '',
            email: '',
            passport: '',
            snils: '',
            avatar_url: '',
            is_active: true,
            is_admin: false,
            student: null,
            teacher: null,
          }
      this.$adminStore.error = ''
      this.opened = true
    },
    async save() {
      if (!this.draft) return
      const saved = await this.$adminStore.saveUser(this.draft)
      if (saved) {
        this.opened = false
        this.draft = null
      }
    },
    async toggleAccess() {
      if (!this.target) return
      const saved = await this.$adminStore.toggleUser(this.target.id)
      if (saved) {
        this.confirmOpen = false
        this.target = null
      }
    },
  },
})
</script>
