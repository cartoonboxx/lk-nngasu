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
    <p v-if="detailLoading" role="status">Загрузка карточки…</p>
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
          <span v-if="item.student">
            {{ item.student.is_headman ? 'Староста' : 'Студент' }} ·
            {{ groupNames[item.student.group_id] }}
          </span>
          <br v-if="item.student && item.teacher" />
          <span v-if="item.teacher">Преподаватель</span>
          <small v-if="!item.student && !item.teacher">Без учебной роли</small>
          <strong v-if="item.is_admin">Администратор</strong>
        </td>
        <td>
          <StatusBadge :variant="item.is_active ? 'success' : 'error'">
            {{ item.is_active ? 'Активен' : 'Заблокирован' }}
          </StatusBadge>
        </td>
        <td>
          <div :class="style.actions">
            <Tooltip text="Карточка">
              <IconButton variant="ghost" size="sm" round @click="openUser(item)">
                <PhIdentificationCard />
              </IconButton>
            </Tooltip>
            <Tooltip :text="item.is_active ? 'Заблокировать' : 'Восстановить'">
              <IconButton
                variant="ghost"
                size="sm"
                round
                @click="
                  () => {
                    target = item
                    confirmOpen = true
                  }
                "
              >
                <PhProhibit v-if="item.is_active" />
                <PhLockOpen v-else />
              </IconButton>
            </Tooltip>
          </div>
        </td>
      </tr>
    </AdminTable>
    <AdminDialog
      v-model="opened"
      :title="draft?.id ? 'Карточка пользователя' : 'Новый пользователь'"
    >
      <form v-if="draft" :class="style.form" @submit.prevent="save">
        <fieldset :class="style.formFields" :disabled="$adminStore.saving">
          <UserFields :key="`${draft.id}-${opened}`" v-model="draft" />
        </fieldset>
        <p v-if="!draft.id" :class="style.notice">
          После создания пользователю будет отправлено приглашение задать пароль.
        </p>
        <div :class="style.dialogFooter">
          <Button
            v-if="draft.id"
            variant="secondary"
            :disabled="$adminStore.saving"
            @click="resendInvitation"
            >Повторить приглашение</Button
          >
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
import { apiNotifications } from '@/api/client'
import StatusBadge from '@/components/UI/StatusBadge/StatusBadge.vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import { defineComponent } from 'vue'
import { PhIdentificationCard, PhLockOpen, PhPlus, PhProhibit } from '@phosphor-icons/vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import UserFields from './components/UserFields.vue'
import { useAuthStore } from '@/stores/auth'
import { normalizeUser, type AdminUserResponse } from './admin.repository'
import type { AdminUser } from './admin.types'
import style from './Admin.module.scss'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
export default defineComponent({
  components: {
    PhLockOpen,
    PhProhibit,
    PhIdentificationCard,
    IconButton,
    Tooltip,
    StatusBadge,
    SectionHeader,
    Select,
    PhPlus,
    TextInput,
    Button,
    AdminTable,
    AdminDialog,
    UserFields,
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
        { value: 'headman', label: 'Староста' },
        { value: 'teacher', label: 'Преподаватель' },
      ],
      adminOptions: [
        { value: '', label: 'Все' },
        { value: 'yes', label: 'Да' },
        { value: 'no', label: 'Нет' },
      ],
      columns: ['Пользователь', 'Принадлежность', 'Доступ', 'Действия'],
      style,
      detailLoading: false,
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
    auth() {
      return useAuthStore()
    },
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
          (!this.role ||
            (this.role === 'headman' && item.student?.is_headman) ||
            (this.role === 'student' && Boolean(item.student)) ||
            (this.role === 'teacher' && Boolean(item.teacher))) &&
          (!this.admin || item.is_admin === (this.admin === 'yes')),
      )
    },
  },
  async mounted() {
    const user = this.$adminStore.data?.users.find(
      (item) => item.id === Number(this.$route.query.user),
    )
    if (user) await this.openUser(user)
    else if (Number(this.$route.query.user) > 0)
      await this.openUser({ id: Number(this.$route.query.user) } as AdminUser)
  },
  methods: {
    async resendInvitation() {
      if (!this.draft?.id || this.$adminStore.saving) return
      const id = this.draft.id
      await this.$adminStore.execute(async () => {
        await this.auth.request(`/users/${id}/invitations`, { method: 'POST' })
      }, 'Приглашение поставлено в очередь отправки.')
    },
    async openUser(item?: AdminUser) {
      if (this.detailLoading) return
      this.$adminStore.error = ''
      this.detailLoading = true
      try {
        this.draft = item
          ? normalizeUser(await this.auth.request<AdminUserResponse>(`/users/${item.id}`))
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
        this.opened = true
      } catch (error) {
        apiNotifications.onError?.(error)
        this.$adminStore.error =
          error instanceof Error ? error.message : 'Не удалось загрузить пользователя.'
      } finally {
        this.detailLoading = false
      }
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
