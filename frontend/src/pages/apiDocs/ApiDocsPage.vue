<template>
  <main :class="style.page">
    <header :class="style.header">
      <div>
        <p :class="style.eyebrow">Личный кабинет ННГАСУ · проект контракта v1.1</p>
        <h1>API проекта</h1>
        <p :class="style.lead">
          HTTP и WebSocket для сценариев приложения: от входа и учебных заданий до расписания и
          административной панели.
        </p>
      </div>
      <div :class="style.headerLinks">
        <RouterLink to="/dashboard">Личный кабинет</RouterLink>
        <RouterLink to="/admin">Административная панель</RouterLink>
      </div>
    </header>

    <div :class="style.layout">
      <aside :class="style.sidebar">
        <p :class="style.eyebrow">Содержание</p>
        <nav :class="style.navigation">
          <RouterLink :to="{ name: 'api-docs', hash: '#overview' }">Правила контракта</RouterLink>
          <RouterLink :to="{ name: 'api-docs', hash: '#coverage' }">Покрытие страниц</RouterLink>
          <RouterLink :to="{ name: 'api-docs', hash: '#errors' }">Ошибки и пагинация</RouterLink>
          <RouterLink
            v-for="group in groups"
            :key="group.id"
            :to="{ name: 'api-docs', hash: `#${group.id}` }"
            @click="
              () => {
                search = ''
                method = ''
                stage = ''
              }
            "
            >{{ group.title }} <span>{{ group.endpoints.length }}</span></RouterLink
          >
          <RouterLink :to="{ name: 'api-docs', hash: '#websocket' }"
            >WebSocket · команды и события</RouterLink
          >
          <RouterLink :to="{ name: 'api-docs', hash: '#models' }"
            >Схемы данных <span>{{ schemas.length }}</span></RouterLink
          >
          <RouterLink :to="{ name: 'api-docs', hash: '#plan' }">Порядок реализации</RouterLink>
        </nav>
        <p :class="style.sidebarNote">
          {{ endpointCount }} HTTP-методов · {{ schemas.length }} схем<br />База:
          <code>/api/v1</code>
        </p>
      </aside>

      <div :class="style.content">
        <section id="overview" :class="style.section">
          <div :class="style.sectionHeading">
            <h2>Правила контракта</h2>
            <span>13 ноября 2026</span>
          </div>
          <p :class="style.notice">
            Документация описывает целевое API. Она не отправляет запросы и не добавляет серверные
            обработчики. Требования определяются сценариями приложения. Совместимость с текущим
            бэкендом не требуется.
          </p>
          <div :class="style.legend">
            <span
              v-for="(label, key) in stageLabels"
              :key="key"
              :class="[style.stage, style[key]]"
              >{{ label }}</span
            >
          </div>
          <details
            v-for="(rule, index) in rules"
            :key="rule.title"
            :class="style.rule"
            :open="index < 2"
          >
            <summary>{{ rule.title }}</summary>
            <p>{{ rule.text }}</p>
          </details>
        </section>

        <section id="coverage" :class="style.section">
          <h2>Как API покрывает интерфейс</h2>
          <p :class="style.muted">
            Учтены страницы кабинета, административные формы и ожидаемое поведение пользователей.
            Неиспользуемый код и удалённые кнопки отделены от обязательных сценариев.
          </p>
          <div :class="style.tableScroll">
            <table :class="style.table">
              <thead>
                <tr>
                  <th>Раздел</th>
                  <th>Страницы</th>
                  <th>API</th>
                  <th>Что учесть</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in coverage" :key="row[0]">
                  <td v-for="(cell, index) in row" :key="index">{{ cell }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section id="errors" :class="style.section">
          <h2>Ответы, ошибки и пагинация</h2>
          <p>
            Успешный объект: <code>{"data": T}</code>. Страница:
            <code>{"data": T[], "meta": {"limit": 50, "offset": 0, "total": 120}}</code>. Пустая
            страница: <code>data: []</code>, корректный <code>total</code>. При
            <code>204</code> тела нет.
          </p>
          <p>
            Ошибка: <code>error.code</code> — стабильный машинный код, <code>error.message</code> —
            текст для пользователя. Необязательный <code>error.details</code> содержит
            <code>fields: {"имя_поля": ["описание"]}</code> для валидации либо
            <code>conflicts: ScheduleConflict[]</code> для расписания. Другие ключи
            <code>details</code> описываются в соответствующей карточке.
          </p>
          <pre :class="style.code"><code>{{ errorExample }}</code></pre>
          <div :class="style.tableScroll">
            <table :class="style.table">
              <thead>
                <tr>
                  <th>HTTP</th>
                  <th>Код</th>
                  <th>Поведение</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in errors" :key="row[0]">
                  <td>{{ row[0] }}</td>
                  <td>
                    <code>{{ row[1] }}</code>
                  </td>
                  <td>{{ row[2] }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p :class="style.muted">
            Каждый защищённый метод может вернуть 401/403, метод с path-id — 400/404; любой —
            500/503. В карточках указаны дополнительные бизнес-ошибки. Ошибки не маскируются
            успешным HTTP 200.
          </p>
        </section>

        <section :class="style.section">
          <h2>Каталог HTTP-методов</h2>
          <div :class="style.filters">
            <TextInput
              v-model="search"
              type="search"
              label="Поиск по пути, названию, полям и правилам"
              placeholder="Например: староста, workload, file_ids"
              full-width
            />
            <label>HTTP-метод<Select v-model="method" :options="methodOptions" full-width /></label>
            <label>Категория<Select v-model="stage" :options="stageOptions" full-width /></label>
          </div>
          <p :class="style.muted">
            Найдено {{ visibleEndpointCount }} из {{ endpointCount }}. Откройте метод, чтобы увидеть
            полный контракт. Примеры содержат вымышленные данные.
          </p>
          <p v-if="!visibleEndpointCount" :class="style.notice">
            По заданным фильтрам методов нет.
          </p>
        </section>

        <section
          v-for="group in visibleGroups"
          :id="group.id"
          :key="group.id"
          :class="style.section"
        >
          <div :class="style.sectionHeading">
            <h2>{{ group.title }}</h2>
            <span>{{ group.endpoints.length }} методов</span>
          </div>
          <p>{{ group.description }}</p>
          <p :class="style.source">
            Основание: <code>{{ group.source }}</code>
          </p>
          <details
            v-for="endpoint in group.endpoints"
            :id="endpoint.id"
            :key="endpoint.id"
            :class="style.endpoint"
            :open="$route.hash === `#${endpoint.id}`"
          >
            <summary :class="style.endpointSummary">
              <span :class="[style.method, style[endpoint.method.toLowerCase()]]">{{
                endpoint.method
              }}</span>
              <span :class="style.endpointName"
                ><code>{{ endpoint.path }}</code
                ><span>{{ endpoint.title }}</span></span
              >
              <span :class="[style.stage, style[endpoint.stage]]">{{
                stageLabels[endpoint.stage]
              }}</span>
            </summary>
            <div :class="style.endpointBody">
              <p><strong>Доступ:</strong> {{ accessLabels[endpoint.access] }}.</p>
              <p>
                <strong>URL:</strong> <code>{{ endpoint.url }}</code>
              </p>
              <h3>Headers</h3>
              <ApiFieldsTable :fields="endpoint.requestHeaders" />
              <template v-if="endpoint.pathFields.length"
                ><h3>Path</h3>
                <ApiFieldsTable :fields="endpoint.pathFields"
              /></template>
              <h3>Query</h3>
              <ApiFieldsTable v-if="endpoint.query?.length" :fields="endpoint.query" />
              <p v-else :class="style.muted">Не принимает query-параметры.</p>
              <h3>Body</h3>
              <template v-if="endpoint.bodyModel">
                <p>
                  <RouterLink
                    :to="{ name: 'api-docs', hash: `#schema-${endpoint.bodyModel.name}` }"
                    >{{ endpoint.bodyModel.name }}</RouterLink
                  >
                  — {{ endpoint.bodyModel.description }}
                </p>
                <ApiFieldsTable :fields="endpoint.bodyModel.fields" />
                <details :class="style.example">
                  <summary>
                    Пример тела запроса{{
                      endpoint.body === 'FileUpload' ? ' (поля multipart)' : ''
                    }}
                  </summary>
                  <pre :class="style.code"><code>{{ endpoint.bodyExample }}</code></pre>
                </details>
              </template>
              <p v-else :class="style.muted">Тело запроса отсутствует.</p>
              <h3>Ответ · HTTP {{ endpoint.response.status }}</h3>
              <template v-if="endpoint.response.schema">
                <p>
                  <code>{{ endpoint.responseType }}</code> ·
                  <RouterLink
                    :to="{ name: 'api-docs', hash: `#schema-${endpoint.response.schema}` }"
                    >Поля {{ endpoint.response.schema }}</RouterLink
                  >
                </p>
                <p v-if="endpoint.response.page" :class="style.muted">
                  Возвращает data-массив и meta: limit, offset, total. total учитывает все фильтры.
                </p>
                <details :class="style.example">
                  <summary>Пример ответа</summary>
                  <pre :class="style.code"><code>{{ endpoint.responseExample }}</code></pre>
                </details>
              </template>
              <p v-else>
                {{
                  endpoint.response.upgrade
                    ? 'WebSocket: Switching Protocols, далее JSON-кадры. HTTP-тела нет.'
                    : endpoint.response.binary
                      ? `Бинарный ответ: ${endpoint.response.binary}. Без JSON.`
                      : 'Ответ без тела.'
                }}
              </p>
              <p v-if="endpoint.response.description">{{ endpoint.response.description }}</p>
              <h3>Правила и ограничения</h3>
              <ul>
                <li v-for="note in endpoint.notes" :key="note">{{ note }}</li>
              </ul>
              <template v-if="endpoint.errors?.length"
                ><h3>Бизнес-ошибки</h3>
                <ul>
                  <li v-for="error in endpoint.errors" :key="error">
                    <code>{{ error }}</code>
                  </li>
                </ul></template
              >
              <p :class="style.muted">
                Общие ошибки и формат ответа описаны
                <RouterLink :to="{ name: 'api-docs', hash: '#errors' }">выше</RouterLink>.
              </p>
            </div>
          </details>
        </section>

        <WebSocketDocs :schemas="schemas" />

        <section id="models" :class="style.section">
          <div :class="style.sectionHeading">
            <h2>Схемы данных</h2>
            <span>{{ schemas.length }} DTO</span>
          </div>
          <p>
            Обязательность у ответа означает, что ключ всегда присутствует. У запроса — что клиент
            обязан его передать. Nullable-типы явно допускают <code>null</code>.
            <code>LessonType</code> и <code>SubmissionStatus</code> — перечисления из правил
            контракта.
          </p>
          <div :class="style.schemaIndex">
            <RouterLink
              v-for="schema in schemas"
              :key="schema.name"
              :to="{ name: 'api-docs', hash: `#schema-${schema.name}` }"
              >{{ schema.name }}</RouterLink
            >
          </div>
          <article
            v-for="schema in schemas"
            :id="`schema-${schema.name}`"
            :key="schema.name"
            :class="style.schema"
          >
            <h3>{{ schema.name }}</h3>
            <p>{{ schema.description }}</p>
            <ApiFieldsTable :fields="schema.fields" />
            <details :class="style.example">
              <summary>Пример {{ schema.name }}</summary>
              <pre
                :class="style.code"
              ><code>{{ JSON.stringify(schema.example, null, 2) }}</code></pre>
            </details>
          </article>
        </section>

        <section id="plan" :class="style.section">
          <h2>Порядок реализации и границы интеграции</h2>
          <div :class="style.tableScroll">
            <table :class="style.table">
              <thead>
                <tr>
                  <th>Этап</th>
                  <th>Работа</th>
                  <th>Зависимости и результат</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="row in plan" :key="row[1]">
                  <td>{{ row[0] }}</td>
                  <td>{{ row[1] }}</td>
                  <td>{{ row[2] }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p :class="style.notice">
            Источник контрактов страницы: <code>src/pages/apiDocs/api-docs.endpoints.ts</code>,
            <code>api-docs.schemas.ts</code> , <code>api-docs.websocket.ts</code> и
            <code>api-docs.overview.ts</code>. Их нужно менять вместе с согласованными решениями.
            Эндпоинты в рамках этой задачи не реализованы.
          </p>
        </section>
      </div>
    </div>
  </main>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { RouterLink } from 'vue-router'
import TextInput from '@/components/UI/inputs/TextInput/TextInput.vue'
import Select from '@/components/UI/inputs/Select/Select.vue'
import ApiFieldsTable from './ApiFieldsTable.vue'
import WebSocketDocs from './WebSocketDocs.vue'
import { apiGroups } from './api-docs.endpoints'
import { apiSchemas } from './api-docs.schemas'
import {
  accessLabels,
  stageLabels,
  apiRules,
  apiErrors,
  apiErrorExample,
  coverage,
  implementationPlan,
} from './api-docs.overview'
import type { ApiField } from './api-docs.types'
import style from './ApiDocsPage.module.scss'

export default defineComponent({
  name: 'ApiDocsPage',
  components: { RouterLink, TextInput, Select, ApiFieldsTable, WebSocketDocs },
  data() {
    return {
      style,
      schemas: apiSchemas,
      accessLabels,
      stageLabels,
      rules: apiRules,
      errors: apiErrors,
      errorExample: JSON.stringify(apiErrorExample, null, 2),
      coverage,
      plan: implementationPlan,
      search: '',
      method: '',
      stage: '',
      methodOptions: [
        { value: '', label: 'Все методы' },
        ...['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map((method) => ({
          value: method,
          label: method,
        })),
      ],
      stageOptions: [
        { value: '', label: 'Все сценарии' },
        ...Object.entries(stageLabels).map(([value, label]) => ({ value, label })),
      ],
    }
  },
  computed: {
    groups() {
      const schemasByName = new Map(apiSchemas.map((schema) => [schema.name, schema]))
      return apiGroups.map((group) => ({
        ...group,
        endpoints: group.endpoints.map((endpoint) => {
          const bodyModel = endpoint.body ? schemasByName.get(endpoint.body) : undefined
          const responseModel = endpoint.response.schema
            ? schemasByName.get(endpoint.response.schema)
            : undefined
          const requestHeaders: ApiField[] = endpoint.response.upgrade
            ? [
                ['Upgrade', 'websocket', true, 'Задаёт браузер при HTTP Upgrade.'],
                ['Connection', 'Upgrade', true, 'Задаёт браузер.'],
                ['Sec-WebSocket-Version', '13', true, 'Версия транспорта, задаёт браузер.'],
                ['Sec-WebSocket-Key', 'string', true, 'Случайный nonce, задаёт браузер.'],
                [
                  'Sec-WebSocket-Protocol',
                  'lk-chat.v1',
                  true,
                  'Передать вторым аргументом new WebSocket(url, protocol).',
                ],
                ['Origin', 'string', true, 'Origin приложения; сервер проверяет разрешённый хост.'],
              ]
            : [
                [
                  'Accept',
                  'string',
                  true,
                  endpoint.response.binary ? endpoint.response.binary : 'application/json',
                ],
                ...(!['public', 'refresh', 'ticket'].includes(endpoint.access)
                  ? ([
                      ['Authorization', 'Bearer <access_token>', true, 'Действующий access JWT.'],
                    ] satisfies ApiField[])
                  : []),
                ...(endpoint.body
                  ? ([
                      [
                        'Content-Type',
                        'string',
                        true,
                        endpoint.body === 'FileUpload'
                          ? 'multipart/form-data; boundary=<browser-generated>'
                          : 'application/json',
                      ],
                    ] satisfies ApiField[])
                  : []),
                ...(endpoint.headers ?? []),
              ]
          const payload =
            endpoint.response.page || endpoint.response.array
              ? [responseModel?.example]
              : responseModel?.example
          const responseExample =
            endpoint.response.schema === 'Health'
              ? responseModel?.example
              : {
                  data: payload,
                  ...(endpoint.response.page ? { meta: { limit: 50, offset: 0, total: 1 } } : {}),
                }
          return {
            ...endpoint,
            id: `endpoint-${endpoint.method.toLowerCase()}-${endpoint.path.replace(/[^a-zA-Z0-9]+/g, '-').replace(/^-|-$/g, '')}`,
            url: endpoint.response.schema === 'Health' ? endpoint.path : `/api/v1${endpoint.path}`,
            bodyModel,
            requestHeaders,
            pathFields: [...endpoint.path.matchAll(/:(\w+)/g)].map((match): ApiField => [
              match[1]!,
              'integer',
              true,
              match[1] === 'mid'
                ? 'ID сообщения внутри чата :id.'
                : 'ID ресурса из пути; положительное безопасное целое.',
            ]),
            bodyExample: JSON.stringify(bodyModel?.example, null, 2),
            responseType: endpoint.response.page
              ? `Page<${endpoint.response.schema}>`
              : `{ data: ${endpoint.response.schema}${endpoint.response.array ? '[]' : ''}${endpoint.response.nullable ? ' | null' : ''} }`,
            responseExample: JSON.stringify(responseExample, null, 2),
            searchText: JSON.stringify({
              group: group.title,
              endpoint,
              body: bodyModel?.fields,
              response: responseModel?.fields,
            }).toLocaleLowerCase('ru'),
          }
        }),
      }))
    },
    endpointCount(): number {
      return this.groups.reduce((total, group) => total + group.endpoints.length, 0)
    },
    visibleGroups() {
      const query = this.search.trim().toLocaleLowerCase('ru')
      return this.groups
        .map((group) => ({
          ...group,
          endpoints: group.endpoints.filter(
            (endpoint) =>
              (!this.method || endpoint.method === this.method) &&
              (!this.stage || endpoint.stage === this.stage) &&
              (!query || endpoint.searchText.includes(query)),
          ),
        }))
        .filter((group) => group.endpoints.length)
    },
    visibleEndpointCount(): number {
      return this.visibleGroups.reduce((total, group) => total + group.endpoints.length, 0)
    },
  },
})
</script>
