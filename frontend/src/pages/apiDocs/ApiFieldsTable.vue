<template>
  <div :class="style.tableScroll">
    <table :class="style.table">
      <thead>
        <tr>
          <th>Поле</th>
          <th>Тип</th>
          <th>Обязательно</th>
          <th>Описание</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="field in fields" :key="field[0]">
          <td>
            <code>{{ field[0] }}</code>
          </td>
          <td>
            <code>{{ field[1] }}</code>
          </td>
          <td>{{ field[2] ? 'Да' : 'Нет' }}</td>
          <td>{{ field[3] }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import type { ApiField } from './api-docs.types'
import style from './ApiDocsPage.module.scss'

export default defineComponent({
  name: 'ApiFieldsTable',
  props: {
    fields: { type: Array as PropType<ApiField[]>, required: true },
  },
  data() {
    return { style }
  },
})
</script>
