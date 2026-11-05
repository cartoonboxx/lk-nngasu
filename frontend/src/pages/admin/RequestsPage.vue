<template>
  <section :class="style.page">
    <SectionHeader
      header="Заявки на регистрацию"
      paragraph="Проверяйте данные и предоставляйте доступ к личному кабинету"
      nm
    >
      <template #actions>
        <span :class="[style.pill, style.pending]">Ожидают решения: {{ pending }}</span>
      </template>
    </SectionHeader>
    <div :class="style.toolbar">
      <TextInput
        v-model="search"
        type="search"
        placeholder="Поиск по email или имени"
        label="Поиск заявок"
        full-width
      />
      <label :class="style.field">
        Статус
        <Select v-model="status" :options="statusOptions" full-width />
      </label>
      <label :class="style.field">
        Сортировка
        <Select v-model="sort" :options="sortOptions" full-width />
      </label>
    </div>
    <AdminTable
      :columns="columns"
      :total="filtered.length"
      :reset-key="`${search}-${status}-${sort}`"
      v-slot="{ start, end }"
    >
      <tr v-for="item in filtered.slice(start, end)" :key="item.id">
        <td>
          <Button variant="ghost" size="sm" :class="style.recordButton" @click="openRequest(item)">
            <span :class="style.recordCopy">
              <strong>{{ item.last_name }} {{ item.first_name }}</strong>
              <small>{{ item.email }}</small>
            </span>
          </Button>
        </td>
        <td>{{ requestDates[item.id]?.created }}</td>
        <td>
          <span :class="[style.pill, style[item.status]]">{{ statusLabels[item.status] }}</span>
        </td>
        <td>
          <Button variant="ghost" size="sm" @click="openRequest(item)">
            {{ item.status === 'pending' ? 'Рассмотреть' : 'Подробнее' }}
          </Button>
        </td>
      </tr>
    </AdminTable>
    <AdminDialog v-model="opened" :title="selected ? `Заявка №${selected.id}` : 'Заявка'">
      <template v-if="selected && user">
        <dl :class="style.detail">
          <dt>Получена</dt>
          <dd>{{ requestDates[selected.id]?.created }}</dd>
          <dt>Паспорт (серия и номер)</dt>
          <dd>{{ selected.passport || 'Не указан' }}</dd>
          <dt>СНИЛС</dt>
          <dd>{{ selected.snils || 'Не указан' }}</dd>
          <dt>Комментарий</dt>
          <dd>{{ selected.comment || 'Не указан' }}</dd>
        </dl>
        <template v-if="selected.status === 'pending'">
          <div :class="style.tabs">
            <StyleTab
              text="Одобрить"
              :active="decision === 'approve'"
              @select="decision = 'approve'"
            />
            <StyleTab
              text="Отклонить"
              :active="decision === 'reject'"
              @select="decision = 'reject'"
            />
          </div>
          <form :class="style.form" @submit.prevent="review">
            <template v-if="decision === 'approve'">
              <p :class="style.muted">
                Определите роль вручную. Email не назначает роль автоматически.
              </p>
              <UserFields :key="`${selected.id}-${opened}`" v-model="user" registration />
            </template>
            <p v-else :class="style.notice">
              Заявка будет отклонена. Для повторного рассмотрения потребуется новая заявка.
            </p>
            <div :class="style.dialogFooter">
              <Button variant="ghost" @click="opened = false">Отмена</Button>
              <Button
                :variant="decision === 'reject' ? 'danger' : 'primary'"
                type="submit"
                :loading="$adminStore.saving"
              >
                {{
                  decision === 'reject'
                    ? 'Подтвердить отклонение'
                    : 'Одобрить и создать пользователя'
                }}
              </Button>
            </div>
          </form>
        </template>
        <dl v-else :class="style.detail">
          <dt>Решение</dt>
          <dd>{{ statusLabels[selected.status] }}</dd>
          <dt>Рассмотрел</dt>
          <dd>{{ selected.reviewed_by }}</dd>
          <dt>Дата решения</dt>
          <dd>
            {{ requestDates[selected.id]?.reviewed }}
          </dd>
          <dt v-if="selected.user_id">Пользователь</dt>
          <dd v-if="selected.user_id">
            <RouterLink :class="style.link" :to="`/admin/users?user=${selected.user_id}`">
              Открыть пользователя №{{ selected.user_id }}
            </RouterLink>
          </dd>
        </dl>
      </template>
    </AdminDialog>
  </section>
</template>
<script lang="ts">
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import UserFields from './components/UserFields.vue'
import { adminDate, statusLabels, type RegistrationRequest, type AdminUser } from './admin.types'
import style from './Admin.module.scss'
export default defineComponent({
  components: {
    SectionHeader,
    StyleTab,
    Select,
    RouterLink,
    TextInput,
    Button,
    AdminTable,
    AdminDialog,
    UserFields,
  },
  data() {
    return {
      statusOptions: [
        { value: '', label: 'Все статусы' },
        { value: 'pending', label: 'Ожидают решения' },
        { value: 'approved', label: 'Одобренные' },
        { value: 'rejected', label: 'Отклонённые' },
      ],
      sortOptions: [
        { value: 'desc', label: 'Сначала новые' },
        { value: 'asc', label: 'Сначала старые' },
      ],
      columns: ['Заявитель', 'Дата заявки', 'Статус', 'Действие'],
      style,
      statusLabels,
      search: '',
      status: 'pending',
      sort: 'desc',
      selected: null as RegistrationRequest | null,
      user: null as AdminUser | null,
      opened: false,
      decision: 'approve',
    }
  },
  computed: {
    requestDates(): Record<number, { created: string; reviewed: string }> {
      return Object.fromEntries(
        (this.$adminStore.data?.requests ?? []).map((item) => [
          item.id,
          {
            created: adminDate.format(new Date(item.created_at)),
            reviewed: item.reviewed_at ? adminDate.format(new Date(item.reviewed_at)) : '—',
          },
        ]),
      )
    },

    pending(): number {
      return this.$adminStore.data?.requests.filter((item) => item.status === 'pending').length ?? 0
    },
    filtered(): RegistrationRequest[] {
      return (this.$adminStore.data?.requests ?? [])
        .filter(
          (item) =>
            (!this.status || item.status === this.status) &&
            `${item.email} ${item.last_name} ${item.first_name}`
              .toLowerCase()
              .includes(this.search.toLowerCase()),
        )
        .sort((a, b) => (this.sort === 'desc' ? -1 : 1) * a.created_at.localeCompare(b.created_at))
    },
  },
  methods: {
    openRequest(item: RegistrationRequest) {
      this.selected = item
      this.user = {
        id: 0,
        last_name: item.last_name,
        first_name: item.first_name,
        middle_names: item.middle_names,
        email: item.email,
        passport: item.passport,
        snils: item.snils,
        avatar_url: '',
        is_active: true,
        is_admin: false,
        student: null,
        teacher: null,
      }
      this.decision = 'approve'
      this.$adminStore.error = ''
      this.opened = true
    },
    async review() {
      if (!this.selected) return
      const saved = await this.$adminStore.reviewRequest(
        this.selected.id,
        this.decision === 'approve' ? this.user : null,
      )
      if (saved) {
        this.opened = false
        this.selected = null
        this.user = null
      }
    },
  },
})
</script>
