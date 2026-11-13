<template>
  <section id="websocket" :class="style.section">
    <div :class="style.sectionHeading">
      <h2>WebSocket · сообщения</h2>
      <span>{{ commands.length }} команд · {{ events.length }} типов событий/служебных кадров</span>
    </div>
    <p :class="style.notice">
      Сообщения работают через WebSocket с первого этапа. HTTP загружает исходное состояние и
      историю. Протокол <code>lk-chat.v1</code>:
      <code>wss://&lt;host&gt;/api/v1/realtime?ticket=&lt;ticket&gt;</code>. Примеры ниже —
      документация: страница не устанавливает соединение.
    </p>
    <p>
      <RouterLink :to="{ name: 'api-docs', hash: '#realtime' }"
        >Получение ticket и Upgrade</RouterLink
      >
      ·
      <RouterLink :to="{ name: 'api-docs', hash: '#chat' }"
        >HTTP-снимок, список чатов и история</RouterLink
      >
    </p>
    <details v-for="rule in rules" :key="rule.title" :class="style.rule" open>
      <summary>{{ rule.title }}</summary>
      <p>{{ rule.text }}</p>
    </details>
    <div v-for="group in contractGroups" :key="group.title" :class="style.schema">
      <h3>{{ group.title }}</h3>
      <details v-for="item in group.items" :key="item.type" :class="style.endpoint">
        <summary :class="style.endpointSummary">
          <span :class="[style.method, item.direction === 'command' ? style.post : style.get]">{{
            item.direction === 'command' ? 'C → S' : 'S → C'
          }}</span>
          <span :class="style.endpointName"
            ><code>{{ item.type }}</code
            ><span>{{ item.title }}</span></span
          >
          <span :class="[style.stage, style[item.stage]]">{{ stageLabels[item.stage] }}</span>
        </summary>
        <div :class="style.endpointBody">
          <p>
            <strong>Тип кадра:</strong>
            {{
              item.direction === 'command'
                ? 'Команда клиента с request_id.'
                : item.direction === 'event'
                  ? 'Событие сервера с event_id и occurred_at; подлежит воспроизведению.'
                  : 'Служебный кадр сервера; без event_id и без воспроизведения.'
            }}
          </p>
          <p>
            <strong>data:</strong>
            <RouterLink :to="{ name: 'api-docs', hash: `#schema-${item.dataSchema}` }">{{
              item.dataSchema
            }}</RouterLink>
          </p>
          <ApiFieldsTable v-if="item.fields.length" :fields="item.fields" />
          <p v-else>Параметров нет: <code>data: {}</code>.</p>
          <details :class="style.example">
            <summary>Пример кадра</summary>
            <pre :class="style.code"><code>{{ item.frameExample }}</code></pre>
          </details>
          <template v-if="item.resultSchema">
            <p>
              <strong>Успех:</strong> <code>ack</code> с исходным <code>request_id</code>;
              <code>data</code> —
              <RouterLink :to="{ name: 'api-docs', hash: `#schema-${item.resultSchema}` }">{{
                item.resultSchema
              }}</RouterLink
              >.
            </p>
            <details :class="style.example">
              <summary>Пример подтверждения</summary>
              <pre :class="style.code"><code>{{ item.resultExample }}</code></pre>
            </details>
          </template>
          <h3>Поведение</h3>
          <ul>
            <li v-for="note in item.notes" :key="note">{{ note }}</li>
          </ul>
          <template v-if="item.errors?.length"
            ><h3>Ошибки внутри WebSocket</h3>
            <p>
              <code>{{ item.errors.join(', ') }}</code>
            </p></template
          >
        </div>
      </details>
    </div>
    <article :class="style.schema">
      <h3>Форматы кадров и ошибок</h3>
      <p>
        В WebSocket нет HTTP body/query/header у каждой команды: после Upgrade данные передаются в
        JSON-кадрах. HTTP-статусы применяются только до ответа <code>101</code>.
      </p>
      <div :class="style.schemaIndex">
        <RouterLink
          v-for="name in envelopeNames"
          :key="name"
          :to="{ name: 'api-docs', hash: `#schema-${name}` }"
          >{{ name }}</RouterLink
        >
      </div>
      <pre
        :class="style.code"
      ><code>{{ JSON.stringify(schemas.find((schema) => schema.name === 'WsError')?.example, null, 2) }}</code></pre>
      <p>
        Общие ошибки команд: <code>invalid_payload</code>, <code>unknown_command</code>,
        <code>stream_not_ready</code>, <code>request_id_reused</code>, <code>rate_limited</code>,
        <code>internal_error</code>. Ошибка одной команды не закрывает исправное соединение; отзыв
        сессии закрывает.
      </p>
      <h3>Коды закрытия соединения</h3>
      <div :class="style.tableScroll">
        <table :class="style.table">
          <thead>
            <tr>
              <th>Код</th>
              <th>Причина</th>
              <th>Действие клиента</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in closeCodes" :key="row[0]">
              <td>
                <code>{{ row[0] }}</code>
              </td>
              <td>{{ row[1] }}</td>
              <td>{{ row[2] }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p :class="style.source">
        Транспорт и браузерный handshake:
        <a href="https://websockets.spec.whatwg.org/" target="_blank" rel="noreferrer"
          >WHATWG WebSockets</a
        >
        и
        <a href="https://www.rfc-editor.org/rfc/rfc6455.html" target="_blank" rel="noreferrer"
          >RFC 6455</a
        >. Команды, события, ticket, TTL и лимиты выше — прикладной контракт проекта, а не
        требования этих стандартов.
      </p>
    </article>
  </section>
</template>
<script lang="ts">
import { defineComponent, type PropType } from 'vue'
import { RouterLink } from 'vue-router'
import ApiFieldsTable from './ApiFieldsTable.vue'
import type { ApiSchema } from './api-docs.types'
import { stageLabels } from './api-docs.overview'
import {
  websocketCommands,
  websocketEvents,
  websocketRules,
  websocketCloseCodes,
} from './api-docs.websocket'
import style from './ApiDocsPage.module.scss'

export default defineComponent({
  name: 'WebSocketDocs',
  components: { RouterLink, ApiFieldsTable },
  props: { schemas: { type: Array as PropType<ApiSchema[]>, required: true } },
  data() {
    return {
      style,
      stageLabels,
      commands: websocketCommands,
      events: websocketEvents,
      rules: websocketRules,
      closeCodes: websocketCloseCodes,
      envelopeNames: ['WsCommandEnvelope', 'WsEventEnvelope', 'WsAck', 'WsError', 'WsErrorDetails'],
    }
  },
  computed: {
    contractGroups() {
      const byName = new Map(this.schemas.map((schema) => [schema.name, schema]))
      return [
        { title: 'Команды клиента → сервер', items: this.commands },
        { title: 'События и служебные кадры сервера → клиент', items: this.events },
      ].map((group) => ({
        ...group,
        items: group.items.map((item) => ({
          ...item,
          fields: byName.get(item.dataSchema)?.fields ?? [],
          frameExample: JSON.stringify(
            {
              type: item.type,
              ...(item.direction === 'command'
                ? { request_id: '7aa2e6f7-7869-41be-8506-b2b6e5c17615' }
                : {}),
              ...(item.direction === 'event'
                ? { event_id: '781', occurred_at: '2026-11-13T10:00:00Z' }
                : {}),
              data: byName.get(item.dataSchema)?.example,
            },
            null,
            2,
          ),
          resultExample: JSON.stringify(
            {
              type: 'ack',
              request_id: '7aa2e6f7-7869-41be-8506-b2b6e5c17615',
              data: item.resultSchema ? byName.get(item.resultSchema)?.example : undefined,
            },
            null,
            2,
          ),
        })),
      }))
    },
  },
})
</script>
