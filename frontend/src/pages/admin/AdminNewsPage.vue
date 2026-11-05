<template>
  <section :class="style.page">
    <SectionHeader header="Новости" paragraph="Черновики, публикации и объявления университета" nm>
      <template #actions>
        <Button @click="openNews()"><PhPlus :size="18" />Создать новость</Button>
      </template>
    </SectionHeader>
    <div :class="style.toolbar">
      <TextInput
        v-model="search"
        type="search"
        label="Поиск новостей"
        placeholder="Заголовок новости"
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
          <Button variant="ghost" size="sm" :class="style.recordButton" @click="openNews(item)">
            <span :class="style.recordCopy">
              <strong>
                <PhPushPin
                  v-if="item.is_pinned"
                  :class="style.pinIcon"
                  :size="16"
                  title="Закреплённая новость"
                />
                {{ item.title }}
              </strong>
              <small>{{ newsDates[item.id]?.published }}</small>
            </span>
          </Button>
        </td>
        <td>
          <span :class="[style.pill, style[item.status]]">{{ statusLabels[item.status] }}</span>
        </td>
        <td>{{ item.author }}</td>
        <td>{{ newsDates[item.id]?.updated }}</td>
        <td>
          <div :class="style.actions">
            <Button variant="ghost" size="sm" @click="openNews(item)">Изменить</Button>
            <Button
              v-if="item.status !== 'archived'"
              variant="ghost"
              size="sm"
              @click="
                () => {
                  pendingNews = { ...item, status: 'archived' }
                  confirmAction = 'archive'
                  confirmOpen = true
                }
              "
            >
              В архив
            </Button>
            <Button
              v-if="item.status !== 'published'"
              variant="ghost"
              size="sm"
              @click="
                () => {
                  pendingNews = item
                  confirmAction = 'delete'
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
    <AdminDialog
      v-model="opened"
      :title="
        preview ? 'Предпросмотр новости' : draft.id ? 'Редактирование новости' : 'Новая новость'
      "
      :width="860"
    >
      <div v-if="preview" :class="style.form">
        <article :class="style.preview">
          <header :class="newsStyle.header">
            <time :class="newsStyle.date">{{ previewDate }}</time>
            <h1>{{ draft.title || 'Заголовок новости' }}</h1>
          </header>
          <img v-if="draft.cover" :src="draft.cover" alt="Обложка новости" />
          <div :class="newsStyle.body">
            <p v-for="(paragraph, index) in previewParagraphs" :key="index">
              {{ paragraph }}
            </p>
          </div>
        </article>
        <div :class="style.dialogFooter">
          <Button variant="secondary" @click="preview = false">Вернуться к редактированию</Button>
        </div>
      </div>
      <form v-else :class="style.form" @submit.prevent="submit">
        <TextInput v-model="draft.title" label="Заголовок" :maxlength="200" required full-width />
        <TextInput
          v-model="draft.body"
          label="Текст новости"
          multiline
          :base-height="240"
          required
          full-width
          placeholder="Разделяйте абзацы пустой строкой"
        />
        <AdminImage :key="`${draft.id}-${opened}`" v-model="draft.cover" />
        <Checkbox v-model="draft.is_pinned" label="Закрепить новость" />
        <p :class="style.muted">
          Разделяйте абзацы пустой строкой. Форматирование текста пока не поддерживается.
        </p>
        <div :class="style.dialogFooter">
          <Button variant="ghost" @click="preview = true">Предпросмотр</Button>
          <Button
            variant="secondary"
            type="submit"
            :loading="$adminStore.saving"
            @click="intent = 'draft'"
          >
            Сохранить черновик
          </Button>
          <Button type="submit" :loading="$adminStore.saving" @click="intent = 'published'">
            {{ draft.status === 'published' ? 'Сохранить публикацию' : 'Опубликовать' }}
          </Button>
        </div>
      </form>
    </AdminDialog>
    <AdminDialog v-model="confirmOpen" :title="confirmation.title">
      <p>{{ pendingNews?.title }}</p>
      <p :class="style.muted">
        {{ confirmation.description }}
      </p>
      <div :class="style.dialogFooter">
        <Button variant="ghost" @click="confirmOpen = false">Отмена</Button>
        <Button
          :variant="confirmAction === 'publish' ? 'primary' : 'danger'"
          :loading="$adminStore.saving"
          @click="confirm"
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
import { PhPlus, PhPushPin } from '@phosphor-icons/vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Checkbox from '@/components/UI/inputs/Checkbox/Checkbox.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import AdminImage from './components/AdminImage.vue'
import { adminDate, statusLabels, type AdminNews } from './admin.types'
import style from './Admin.module.scss'
import newsStyle from '@/pages/news/page/NewsDetailsPage.module.scss'
export default defineComponent({
  components: {
    SectionHeader,
    Select,
    PhPlus,
    PhPushPin,
    TextInput,
    Checkbox,
    Button,
    AdminTable,
    AdminDialog,
    AdminImage,
  },
  data() {
    return {
      statusOptions: [
        { value: '', label: 'Все статусы' },
        { value: 'draft', label: 'Черновики' },
        { value: 'published', label: 'Опубликованные' },
        { value: 'archived', label: 'Архив' },
      ],
      sortOptions: [
        { value: 'updated', label: 'По изменению' },
        { value: 'published', label: 'По публикации' },
        { value: 'title', label: 'По заголовку' },
      ],
      columns: ['Новость', 'Статус', 'Автор', 'Обновлена', 'Действия'],
      style,
      newsStyle,
      statusLabels,
      search: '',
      status: '',
      sort: 'updated',
      opened: false,
      preview: false,
      confirmOpen: false,
      confirmAction: 'publish' as 'publish' | 'draft' | 'archive' | 'delete',
      confirmationCopy: {
        publish: {
          title: 'Опубликовать новость?',
          description:
            'Новость получит статус опубликованной в демо-данных. Публичная лента пока использует собственные примеры.',
        },
        draft: {
          title: 'Снять новость с публикации?',
          description: 'Новость получит выбранный статус и сохранится в панели.',
        },
        archive: {
          title: 'Архивировать новость?',
          description: 'Новость получит выбранный статус и сохранится в панели.',
        },
        delete: {
          title: 'Удалить новость?',
          description: 'Новость будет удалена из локальных данных. Это действие нельзя отменить.',
        },
      },
      pendingNews: null as AdminNews | null,
      intent: 'draft' as 'draft' | 'published',
      draft: {
        id: 0,
        title: '',
        body: '',
        cover: '',
        author: '',
        status: 'draft',
        is_pinned: false,
        created_at: '',
        updated_at: '',
        published_at: null,
      } as AdminNews,
    }
  },
  computed: {
    confirmation(): { title: string; description: string } {
      return this.confirmationCopy[this.confirmAction]
    },
    previewDate(): string {
      return adminDate.format(new Date(this.draft.published_at || Date.now()))
    },
    previewParagraphs(): string[] {
      return this.draft.body.split(/\n\s*\n/)
    },
    newsDates(): Record<number, { published: string; updated: string }> {
      return Object.fromEntries(
        (this.$adminStore.data?.news ?? []).map((item) => [
          item.id,
          {
            published: item.published_at
              ? `Публикация: ${adminDate.format(new Date(item.published_at))}`
              : 'Не опубликована',
            updated: adminDate.format(new Date(item.updated_at)),
          },
        ]),
      )
    },

    filtered(): AdminNews[] {
      return (this.$adminStore.data?.news ?? [])
        .filter(
          (item) =>
            item.title.toLowerCase().includes(this.search.toLowerCase()) &&
            (!this.status || this.status === item.status),
        )
        .sort((a, b) =>
          this.sort === 'title'
            ? a.title.localeCompare(b.title, 'ru')
            : this.sort === 'published'
              ? (b.published_at ?? '').localeCompare(a.published_at ?? '')
              : b.updated_at.localeCompare(a.updated_at),
        )
    },
  },
  methods: {
    openNews(item?: AdminNews) {
      this.draft = item
        ? { ...item }
        : {
            id: 0,
            title: '',
            body: '',
            cover: '',
            author: '',
            status: 'draft',
            is_pinned: false,
            created_at: '',
            updated_at: '',
            published_at: null,
          }
      this.preview = false
      this.intent = 'draft'
      this.$adminStore.error = ''
      this.opened = true
    },
    async submit() {
      if (this.intent === 'published' || this.draft.status === 'published') {
        this.pendingNews = { ...this.draft, status: this.intent }
        this.confirmAction = this.intent === 'published' ? 'publish' : 'draft'
        this.confirmOpen = true
      } else if (await this.$adminStore.saveNews({ ...this.draft, status: 'draft' })) {
        this.opened = false
      }
    },
    async confirm() {
      if (!this.pendingNews) return
      const saved =
        this.confirmAction === 'delete'
          ? await this.$adminStore.deleteNews(this.pendingNews.id)
          : await this.$adminStore.saveNews(this.pendingNews)
      if (saved) {
        this.confirmOpen = false
        this.opened = false
        this.pendingNews = null
      }
    },
  },
})
</script>
