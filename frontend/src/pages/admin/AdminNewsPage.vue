<template>
  <section :class="style.page">
    <SectionHeader header="Новости" paragraph="Черновики, публикации и архив" nm>
      <template #actions><Button @click="openEditor()">Добавить новость</Button></template>
    </SectionHeader>
    <div :class="style.toolbar">
      <TextInput v-model="search" label="Поиск новостей" type="search" full-width />
      <Select v-model="status" :options="statusOptions" full-width />
    </div>
    <p v-if="loading">Загрузка новостей…</p>
    <p v-else-if="error">
      {{ error }} <Button variant="ghost" @click="loadNews">Повторить</Button>
    </p>
    <AdminTable
      v-else
      :columns="['Текст', 'Автор', 'Статус', 'Действия']"
      :total="filtered.length"
      :reset-key="`${search}-${status}`"
      v-slot="{ start, end }"
    >
      <tr v-for="item in filtered.slice(start, end)" :key="item.id">
        <td>
          <strong>{{ item.excerpt }}</strong
          ><small>{{ new Date(item.updated_at).toLocaleString('ru-RU') }}</small>
        </td>
        <td>{{ item.author.name }}</td>
        <td>{{ statusLabels[item.status] }}</td>
        <td>
          <div :class="style.actions">
            <Tooltip text="Изменить" v-if="item.status !== 'archived'">
              <IconButton
                size="sm"
                variant="ghost"
                :disabled="saving"
                round
                @click="openEditor(item)"
              >
                <PhPencilSimple />
              </IconButton>
            </Tooltip>
            <Tooltip text="Опубликовать" v-if="item.status !== 'published'">
              <IconButton
                size="sm"
                variant="ghost"
                :disabled="saving"
                round
                @click="changeStatus(item, 'published')"
              >
                <PhPaperPlaneTilt />
              </IconButton>
            </Tooltip>
            <Tooltip text="В черновики" v-if="item.status !== 'draft'">
              <IconButton
                size="sm"
                variant="ghost"
                :disabled="saving"
                round
                @click="changeStatus(item, 'draft')"
              >
                <PhFloppyDisk />
              </IconButton>
            </Tooltip>
            <Tooltip text="В архив" v-if="item.status !== 'archived'">
              <IconButton
                size="sm"
                variant="ghost"
                :disabled="saving"
                round
                @click="changeStatus(item, 'archived')"
              >
                <PhArchive />
              </IconButton>
            </Tooltip>
            <Tooltip text="Удалить" v-if="item.status !== 'published'">
              <IconButton size="sm" variant="ghost" :disabled="saving" round @click="target = item">
                <PhTrash />
              </IconButton>
            </Tooltip>
          </div>
        </td>
      </tr>
    </AdminTable>
    <AdminDialog
      :model-value="opened"
      :title="draft.id ? 'Изменить новость' : 'Новая новость'"
      @update:model-value="!saving && (opened = $event)"
    >
      <form :class="style.form" @submit.prevent="save">
        <TextInput
          v-model="draft.body"
          label="Текст новости"
          multiline
          full-width
          required
          :maxlength="50000"
          :disabled="saving"
        />
        <label :class="style.field"
          >Состояние<Select
            v-model="draft.status"
            :options="editorStatuses"
            :disabled="saving"
            full-width
        /></label>
        <div :class="style.dialogFooter">
          <Button variant="ghost" :disabled="saving" @click="opened = false">Отмена</Button>
          <Button type="submit" :loading="saving" :disabled="saving || !draft.body.trim()"
            >Сохранить</Button
          >
        </div>
      </form>
    </AdminDialog>
    <AdminDialog
      :model-value="!!target"
      title="Удалить новость?"
      @update:model-value="!saving && !$event && (target = null)"
    >
      <p>Новость будет удалена без возможности восстановления.</p>
      <div :class="style.dialogFooter">
        <Button variant="ghost" :disabled="saving" @click="target = null">Отмена</Button>
        <Button variant="danger" :loading="saving" @click="remove">Удалить</Button>
      </div>
    </AdminDialog>
  </section>
</template>
<script lang="ts">
import { defineComponent } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { apiNotifications } from '@/api/client'
import type { AdminNews } from '@/api/cabinet.types'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import AdminDialog from './components/AdminDialog.vue'
import AdminTable from './components/AdminTable.vue'
import style from './Admin.module.scss'
import Tooltip from '@/components/UI/Tooltip/Tooltip.vue'
import IconButton from '@/components/UI/buttons/IconButton/IconButton.vue'
import {
  PhArchive,
  PhFloppyDisk,
  PhPaperPlaneTilt,
  PhPencilSimple,
  PhTrash,
} from '@phosphor-icons/vue'

export default defineComponent({
  name: 'AdminNewsPage',
  components: {
    PhTrash,
    PhArchive,
    PhFloppyDisk,
    PhPaperPlaneTilt,
    PhPencilSimple,
    IconButton,
    Tooltip,
    SectionHeader,
    Button,
    TextInput,
    Select,
    AdminDialog,
    AdminTable,
  },
  data() {
    return {
      style,
      news: [] as AdminNews[],
      loading: false,
      saving: false,
      error: '',
      search: '',
      status: '',
      opened: false,
      target: null as AdminNews | null,
      version: 0,
      draft: { id: 0, body: '', status: 'draft' },
      statusLabels: { draft: 'Черновик', published: 'Опубликовано', archived: 'Архив' },
      statusOptions: [
        { value: '', label: 'Все статусы' },
        { value: 'draft', label: 'Черновики' },
        { value: 'published', label: 'Опубликованные' },
        { value: 'archived', label: 'Архив' },
      ],
      editorStatuses: [
        { value: 'draft', label: 'Черновик' },
        { value: 'published', label: 'Опубликовать' },
      ],
    }
  },
  computed: {
    filtered(): AdminNews[] {
      const query = this.search.trim().toLocaleLowerCase('ru')
      return this.news.filter(
        (item) =>
          (!this.status || item.status === this.status) &&
          `${item.body} ${item.author.name}`.toLocaleLowerCase('ru').includes(query),
      )
    },
  },
  created() {
    void this.loadNews()
  },
  beforeUnmount() {
    this.version++
  },
  methods: {
    async loadNews() {
      const version = ++this.version
      this.loading = true
      this.error = ''
      try {
        const items = await useAuthStore().list<AdminNews>('/admin/news')
        if (version === this.version) this.news = items
      } catch (error) {
        if (version === this.version)
          this.error = error instanceof Error ? error.message : 'Не удалось загрузить новости.'
      } finally {
        if (version === this.version) this.loading = false
      }
    },
    openEditor(item?: AdminNews) {
      this.draft = { id: item?.id ?? 0, body: item?.body ?? '', status: item?.status ?? 'draft' }
      this.$adminStore.error = ''
      this.$adminStore.success = ''
      this.error = ''
      this.opened = true
    },
    async save() {
      if (this.saving || !this.draft.body.trim()) return
      const version = this.version
      this.saving = true
      try {
        const saved = await useAuthStore().request<AdminNews>(
          this.draft.id ? `/admin/news/${this.draft.id}` : '/admin/news',
          {
            method: this.draft.id ? 'PUT' : 'POST',
            body: JSON.stringify({ body: this.draft.body.trim(), status: this.draft.status }),
          },
        )
        if (version !== this.version) return
        this.news = [saved, ...this.news.filter((item) => item.id !== saved.id)]
        this.opened = false
        this.$adminStore.success = 'Новость сохранена.'
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.saving = false
      }
    },
    async changeStatus(item: AdminNews, status: AdminNews['status']) {
      if (this.saving) return
      const version = this.version
      this.saving = true
      try {
        const saved = await useAuthStore().request<AdminNews>(`/admin/news/${item.id}`, {
          method: 'PATCH',
          body: JSON.stringify({ status }),
        })
        if (version !== this.version) return
        this.news = this.news.map((entry) => (entry.id === saved.id ? saved : entry))
        this.$adminStore.success = 'Статус новости изменён.'
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.saving = false
      }
    },
    async remove() {
      if (this.saving || !this.target) return
      const version = this.version
      const id = this.target.id
      this.saving = true
      try {
        await useAuthStore().request(`/admin/news/${id}`, { method: 'DELETE' })
        if (version !== this.version) return
        this.news = this.news.filter((item) => item.id !== id)
        this.target = null
        this.$adminStore.success = 'Новость удалена.'
      } catch (error) {
        apiNotifications.onError?.(error)
      } finally {
        this.saving = false
      }
    },
  },
})
</script>
