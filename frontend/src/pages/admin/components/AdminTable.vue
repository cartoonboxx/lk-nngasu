<template>
  <div :class="style.panel">
    <div v-if="total" :class="style.tableScroll">
      <table :class="style.table">
        <thead>
          <tr>
            <th v-for="column in columns" :key="column">{{ column }}</th>
          </tr>
        </thead>
        <tbody>
          <slot :start="start" :end="end" />
        </tbody>
      </table>
    </div>
    <div v-else :class="style.empty">
      <PhMagnifyingGlass :size="32" /><strong>{{ emptyTitle }}</strong>
      <p>Измените поиск или фильтры, либо добавьте запись.</p>
    </div>
    <footer v-if="total" :class="style.pagination">
      <span>{{ start + 1 }}–{{ end }} из {{ total }}</span>
      <div :class="style.actions">
        <Button
          variant="ghost"
          size="sm"
          :disabled="currentPage <= 1"
          @click="page = currentPage - 1"
          >Назад</Button
        >
        <span>{{ currentPage }} / {{ pages }}</span>
        <Button
          variant="ghost"
          size="sm"
          :disabled="currentPage >= pages"
          @click="page = currentPage + 1"
          >Далее</Button
        >
      </div>
    </footer>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { PhMagnifyingGlass } from '@phosphor-icons/vue'
import Button from '@/components/UI/buttons/Button/Button.vue'
import style from '../Admin.module.scss'
export default defineComponent({
  components: { Button, PhMagnifyingGlass },
  props: {
    columns: {
      type: Array as PropType<string[]>,
      required: true,
    },
    total: {
      type: Number,
      required: true,
    },
    resetKey: {
      type: String,
      default: '',
    },
    emptyTitle: {
      type: String,
      default: 'Ничего не найдено',
    },
  },
  data() {
    return { style, page: 1, pageSize: 8 }
  },
  computed: {
    end(): number {
      return Math.min(this.start + this.pageSize, this.total)
    },
    pages(): number {
      return Math.max(1, Math.ceil(this.total / this.pageSize))
    },
    currentPage(): number {
      return Math.min(this.page, this.pages)
    },
    start(): number {
      return (this.currentPage - 1) * this.pageSize
    },
  },
  watch: {
    resetKey() {
      this.page = 1
    },
  },
})
</script>
