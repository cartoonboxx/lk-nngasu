<template>
  <section :class="style.page">
    <SectionHeader
      header="Учебная структура"
      paragraph="Общие справочники университета и учебного процесса"
      nm
    >
      <template #actions>
        <Button @click="openEntry()"><PhPlus :size="18" />Добавить запись</Button>
      </template>
    </SectionHeader>
    <div :class="style.tabs">
      <StyleTab
        :text="label"
        v-for="(label, key) in catalogLabels"
        :key="key"
        :active="kind === key"
        @select="kind = key"
      />
    </div>
    <div :class="style.toolbar">
      <TextInput
        v-model="search"
        type="search"
        :label="`Поиск: ${catalogLabels[kind].toLowerCase()}`"
        :placeholder="kind === 'specializations' ? 'Код и название' : 'Название'"
        full-width
      />
    </div>
    <AdminTable
      :columns="columns"
      :total="filtered.length"
      :reset-key="`${kind}-${search}`"
      v-slot="{ start, end }"
    >
      <tr v-for="item in filtered.slice(start, end)" :key="item.id">
        <td>
          <strong>{{ item.name }}</strong>
        </td>
        <td>
          {{ entryDetails[item.id] }}
        </td>
        <td>
          <StatusBadge>{{ item.is_current ? 'Текущий семестр' : 'Действует' }}</StatusBadge>
        </td>
        <td>
          <div :class="style.actions">
            <Button
              v-if="kind === 'semesters' && !item.is_current"
              size="sm"
              variant="ghost"
              :disabled="$adminStore.saving"
              @click="$adminStore.setCurrentSemester(item.id)"
              >Сделать текущим</Button
            >
            <Button size="sm" variant="ghost" @click="openEntry(item)"> Изменить </Button>
            <Button
              size="sm"
              variant="ghost"
              @click="
                () => {
                  target = item
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
    <AdminDialog v-model="opened" :title="draft.id ? 'Изменить запись' : 'Новая запись'">
      <form :class="style.form" @submit.prevent="save">
        <p :class="style.muted">{{ catalogLabels[kind] }}</p>
        <TextInput
          v-if="kind !== 'semesters'"
          v-model="draft.name"
          :label="kind === 'rooms' ? 'Номер / обозначение аудитории' : 'Название'"
          :maxlength="kind === 'rooms' ? 30 : kind === 'groups' ? 50 : 200"
          required
          full-width
        />
        <TextInput
          v-if="kind === 'specializations'"
          v-model="draft.code"
          label="Код направления"
          :maxlength="20"
          required
          full-width
        />
        <template v-if="kind === 'groups'">
          <label :class="style.field">
            Направление
            <Select
              :model-value="String(draft.specialization_id ?? 0)"
              :options="specializationOptions"
              @update:model-value="draft.specialization_id = Number($event)"
              full-width
              placeholder="Выберите направление"
            />
          </label>
          <label :class="style.field">
            Год поступления<input
              v-model.number="draft.admission_year"
              type="number"
              min="2000"
              max="2100"
              required
            />
          </label>
        </template>
        <div v-if="kind === 'semesters'" :class="style.grid">
          <label :class="style.field">
            Начало учебного года
            <input
              v-model.number="draft.academic_year_start"
              type="number"
              min="2000"
              max="2100"
              required
            />
          </label>
          <label :class="style.field">
            Номер семестра
            <Select
              :model-value="String(draft.number ?? 0)"
              :options="semesterNumberOptions"
              @update:model-value="draft.number = Number($event)"
              full-width
            />
          </label>
        </div>
        <div :class="style.dialogFooter">
          <Button variant="ghost" @click="opened = false">Отмена</Button>
          <Button type="submit" :loading="$adminStore.saving">Сохранить</Button>
        </div>
      </form>
    </AdminDialog>
    <AdminDialog v-model="confirmOpen" title="Удалить запись?">
      <p>{{ target?.name }}</p>
      <p :class="style.muted">
        Запись будет удалена. Удаление доступно только при отсутствии связанных учебных данных.
      </p>
      <div :class="style.dialogFooter">
        <Button variant="ghost" @click="confirmOpen = false">Отмена</Button>
        <Button variant="danger" :loading="$adminStore.saving" @click="removeEntry">
          Удалить
        </Button>
      </div>
    </AdminDialog>
  </section>
</template>
<script lang="ts">
import StatusBadge from '@/components/UI/StatusBadge/StatusBadge.vue'
import SectionHeader from '@/components/SectionHeader/SectionHeader.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import StyleTab from '@/components/UI/tabs/StyleTab/StyleTab.vue'
import { defineComponent } from 'vue'
import { PhPlus } from '@phosphor-icons/vue'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import AdminTable from './components/AdminTable.vue'
import AdminDialog from './components/AdminDialog.vue'
import { catalogLabels, type CatalogKind, type CatalogEntry } from './admin.types'
import style from './Admin.module.scss'
export default defineComponent({
  components: {
    StatusBadge,
    SectionHeader,
    StyleTab,
    Select,
    PhPlus,
    TextInput,
    Button,
    AdminTable,
    AdminDialog,
  },
  data() {
    return {
      semesterNumberOptions: [
        { value: '1', label: '1 · Осенний' },
        { value: '2', label: '2 · Весенний' },
      ],
      columns: ['Название', 'Дополнительные сведения', 'Состояние', 'Действия'],
      style,
      catalogLabels,
      kind: 'groups' as CatalogKind,
      search: '',
      opened: false,
      confirmOpen: false,
      target: null as CatalogEntry | null,
      draft: {
        id: 0,
        name: '',
        code: '',
        specialization_id: 0,
        admission_year: 2026,
        academic_year_start: 2026,
        number: 1,
      } as CatalogEntry,
    }
  },
  computed: {
    entryDetails(): Record<number, string> {
      const data = this.$adminStore.data
      if (!data) return {}
      const specializationCodes = Object.fromEntries(
        data.catalogs.specializations.map((item) => [item.id, item.code]),
      )
      return Object.fromEntries(
        data.catalogs[this.kind].map((item) => {
          if (item.code) return [item.id, item.code]
          if (item.admission_year)
            return [
              item.id,
              `Набор ${item.admission_year} · ${specializationCodes[item.specialization_id ?? 0] ?? '—'}`,
            ]
          if (item.academic_year_start)
            return [
              item.id,
              `${item.academic_year_start}/${item.academic_year_start + 1}, семестр ${item.number}`,
            ]
          return [item.id, '—']
        }),
      )
    },

    specializationOptions(): FloatingPopupItem[] {
      return (this.$adminStore.data?.catalogs.specializations ?? []).map((item) => ({
        value: String(item.id),
        label: `${item.code} · ${item.name}`,
      }))
    },

    filtered(): CatalogEntry[] {
      return (this.$adminStore.data?.catalogs[this.kind] ?? []).filter((item) =>
        `${item.name} ${item.code ?? ''}`.toLowerCase().includes(this.search.toLowerCase()),
      )
    },
  },
  watch: {
    kind() {
      this.search = ''
    },
  },
  methods: {
    openEntry(item?: CatalogEntry) {
      this.draft = item ? { ...item } : { id: 0, name: '' }
      if (this.kind === 'groups')
        Object.assign(this.draft, {
          specialization_id: item?.specialization_id ?? 0,
          admission_year: item?.admission_year ?? 2026,
        })
      if (this.kind === 'semesters')
        Object.assign(this.draft, {
          academic_year_start: item?.academic_year_start ?? 2026,
          number: item?.number ?? 1,
        })
      this.$adminStore.error = ''
      this.opened = true
    },
    async save() {
      const saved = await this.$adminStore.saveCatalog(this.kind, this.draft)
      if (saved) {
        this.opened = false
        this.search = ''
      }
    },
    async removeEntry() {
      if (!this.target) return
      const saved = await this.$adminStore.removeCatalog(this.kind, this.target.id)
      if (saved) {
        this.confirmOpen = false
        this.target = null
      }
    },
  },
})
</script>
