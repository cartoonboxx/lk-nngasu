<template>
  <section :class="style.page">
    <SectionHeader
      header="Заявки на регистрацию"
      paragraph="Проверяйте данные и предоставляйте доступ к личному кабинету"
      nm
    >
      <template #actions>
        <StatusBadge variant="warning"
          >Ожидают решения: {{ $adminStore.pendingRequestsTotal }}</StatusBadge
        >
        <Button variant="secondary" :disabled="$adminStore.requestsLoading" @click="loadPage(page)"
          >Обновить</Button
        >
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
    <p v-if="$adminStore.requestsLoading" role="status">Загрузка заявок…</p>
    <AdminTable
      v-else
      :columns="columns"
      :total="$adminStore.requestsTotal"
      :remote-page="page"
      empty-title="Заявок не найдено"
      empty-description="Новых заявок с выбранными параметрами нет. Измените фильтры или обновите список."
      @page-change="loadPage"
    >
      <tr v-for="item in $adminStore.data?.requests ?? []" :key="item.id">
        <td>
          <Button variant="ghost" size="sm" :class="style.recordButton" @click="openRequest(item)">
            <span :class="style.recordCopy">
              <strong>{{
                [item.last_name, item.first_name].filter(Boolean).join(' ') || `Заявка №${item.id}`
              }}</strong>
              <small>{{ item.email }}</small>
            </span>
          </Button>
        </td>
        <td>{{ requestDates[item.id]?.created }}</td>
        <td>
          <StatusBadge :variant="statusVariants[item.status]">{{
            statusLabels[item.status]
          }}</StatusBadge>
        </td>
        <td>
          <Button variant="ghost" size="sm" @click="openRequest(item)">
            {{ item.status === 'pending' ? 'Рассмотреть' : 'Подробнее' }}
          </Button>
        </td>
      </tr>
    </AdminTable>
    <AdminDialog v-model="opened" :title="selected ? `Заявка №${selected.id}` : 'Заявка'">
      <p v-if="detailLoading" role="status">Загрузка заявки…</p>
      <template v-else-if="selected && user">
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
              <fieldset :class="style.formFields" :disabled="$adminStore.saving">
                <UserFields :key="`${selected.id}-${opened}`" v-model="user" registration />
              </fieldset>
            </template>
            <p v-else :class="style.notice">
              Заявка будет отклонена. Для повторного рассмотрения потребуется новая заявка.
            </p>
            <div :class="style.dialogFooter">
              <Button variant="ghost" :disabled="$adminStore.saving" @click="opened = false"
                >Отмена</Button
              >
              <Button
                :variant="decision === 'reject' ? 'danger' : 'primary'"
                type="submit"
                :loading="$adminStore.saving"
                :disabled="$adminStore.saving"
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
          <dd>{{ selected.reviewed_by?.name || '—' }}</dd>
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
import { apiNotifications } from '@/api/client'
import StatusBadge from '@/components/UI/StatusBadge/StatusBadge.vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { normalizeRequest } from './admin.repository'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import UserFields from './components/UserFields.vue'
import {
  adminDate,
  statusLabels,
  statusVariants,
  type RegistrationRequest,
  type AdminUser,
} from './admin.types'
import style from './Admin.module.scss'
export default defineComponent({
  components: {
    StatusBadge,
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
      statusVariants,
      page: 1,
      reloadTimer: undefined as ReturnType<typeof setTimeout> | undefined,
      detailLoading: false,
      detailVersion: 0,
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
    auth() {
      return useAuthStore()
    },
    requestDates(): Record<number, { created: string; reviewed: string }> {
      return Object.fromEntries(
        [...(this.$adminStore.data?.requests ?? []), ...(this.selected ? [this.selected] : [])].map(
          (item) => [
            item.id,
            {
              created: adminDate.format(new Date(item.created_at)),
              reviewed: item.reviewed_at ? adminDate.format(new Date(item.reviewed_at)) : '—',
            },
          ],
        ),
      )
    },
  },
  async mounted() {
    await this.loadPage(1)
  },
  beforeUnmount() {
    clearTimeout(this.reloadTimer)
    this.detailVersion++
  },
  watch: {
    search() {
      clearTimeout(this.reloadTimer)
      this.reloadTimer = setTimeout(async () => await this.loadPage(1), 300)
    },
    async status() {
      clearTimeout(this.reloadTimer)
      await this.loadPage(1)
    },
    async sort() {
      clearTimeout(this.reloadTimer)
      await this.loadPage(1)
    },
    opened(value: boolean) {
      if (!value) {
        this.detailVersion++
        this.selected = null
        this.user = null
      }
    },
  },
  methods: {
    async loadPage(page: number) {
      clearTimeout(this.reloadTimer)
      this.page = page
      const loaded = await this.$adminStore.loadRequests({
        q: this.search.trim(),
        status: this.status,
        sort: this.sort === 'desc' ? '-created_at' : 'created_at',
        limit: 8,
        offset: (page - 1) * 8,
      })
      if (loaded && page > 1 && (page - 1) * 8 >= this.$adminStore.requestsTotal) {
        await this.loadPage(Math.max(1, Math.ceil(this.$adminStore.requestsTotal / 8)))
      }
    },
    async openRequest(item: RegistrationRequest) {
      const version = ++this.detailVersion
      this.selected = null
      this.user = null
      this.decision = 'approve'
      this.opened = true
      this.detailLoading = true
      this.$adminStore.error = ''
      try {
        const request = normalizeRequest(
          await this.auth.request<RegistrationRequest>(`/registration-requests/${item.id}`),
        )
        if (version !== this.detailVersion) return
        this.selected = request
        this.user = {
          id: 0,
          last_name: request.last_name,
          first_name: request.first_name,
          middle_names: request.middle_names,
          email: request.email,
          passport: request.passport,
          snils: request.snils,
          avatar_url: '',
          is_active: true,
          is_admin: false,
          student: null,
          teacher: null,
        }
      } catch (error) {
        apiNotifications.onError?.(error)
        if (version === this.detailVersion)
          this.$adminStore.error =
            error instanceof Error ? error.message : 'Не удалось загрузить заявку.'
      } finally {
        if (version === this.detailVersion) this.detailLoading = false
      }
    },
    async review() {
      if (!this.selected || this.$adminStore.saving) return
      const id = this.selected.id
      const saved = await this.$adminStore.reviewRequest(
        this.selected.id,
        this.decision === 'approve' ? this.user : null,
      )
      if (saved) {
        this.opened = false
        this.selected = null
        this.user = null
        await this.loadPage(this.page)
      } else {
        try {
          this.selected = normalizeRequest(
            await this.auth.request<RegistrationRequest>(`/registration-requests/${id}`),
          )
          if (this.selected.status !== 'pending') {
            const message = this.$adminStore.error
            await this.loadPage(this.page)
            this.$adminStore.error = message
          }
        } catch {
          // Ошибка решения остаётся видимой, даже если повторное чтение недоступно.
        }
      }
    },
  },
})
</script>
