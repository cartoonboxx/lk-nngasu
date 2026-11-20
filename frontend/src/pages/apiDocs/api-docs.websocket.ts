import type { ApiSchema, ApiStage } from './api-docs.types'

export interface WebSocketContract {
  type: string
  title: string
  direction: 'command' | 'event' | 'service'
  stage: ApiStage
  dataSchema: string
  resultSchema?: string
  errors?: string[]
  notes: string[]
}

const message = {
  id: 300,
  chat_id: 44,
  sender_id: 12,
  sender_name: 'Иванова Анна Сергеевна',
  body: 'Здравствуйте!',
  files: [],
  created_at: '2026-11-16T10:00:00Z',
  edited_at: null,
}

export const websocketSchemas: ApiSchema[] = [
  {
    name: 'WebSocketTicket',
    description:
      'Одноразовый допуск к Upgrade; инфраструктура авторизации вне предметных таблиц DBML.',
    fields: [
      ['ticket', 'string', true, 'Одноразовый секрет, TTL 30 секунд; хранится в серверном кеше.'],
      ['expires_at', 'datetime', true, 'Срок действия.'],
      ['endpoint', '"/api/v1/realtime"', true, 'Путь Upgrade.'],
      ['protocol', '"lk-chat.v1"', true, 'Sec-WebSocket-Protocol.'],
    ],
    example: {
      ticket: '<one-time-ticket>',
      expires_at: '2026-11-16T10:00:30Z',
      endpoint: '/api/v1/realtime',
      protocol: 'lk-chat.v1',
    },
  },
  {
    name: 'WsCommandEnvelope',
    description:
      'Текстовый JSON-кадр клиента. request_id служит корреляции в текущем соединении и не хранится в message.',
    fields: [
      ['type', 'string', true, 'Имя команды.'],
      [
        'request_id',
        'UUID',
        true,
        'ID команды для сопоставления с ack/error. Не гарантирует идемпотентность.',
      ],
      ['data', 'object', true, 'Поля конкретной команды.'],
    ],
    example: {
      type: 'message.send',
      request_id: '7aa2e6f7-7869-41be-8506-b2b6e5c17615',
      data: { chat_id: 44, body: 'Здравствуйте!', file_ids: [] },
    },
  },
  {
    name: 'WsEventEnvelope',
    description:
      'Оперативное событие после фиксации транзакции. Воспроизведение после обрыва не предусмотрено схемой.',
    fields: [
      ['type', 'string', true, 'Имя события.'],
      ['occurred_at', 'datetime', true, 'Время изменения.'],
      ['data', 'object', true, 'Нагрузка события.'],
    ],
    example: { type: 'message.created', occurred_at: message.created_at, data: { message } },
  },
  {
    name: 'WsAck',
    description: 'Успешная запись команды; не означает прочтение или доставку адресату.',
    fields: [
      ['type', '"ack"', true, 'Успех.'],
      ['request_id', 'UUID', true, 'ID исходной команды.'],
      ['data', 'object', true, 'Результат по resultSchema.'],
    ],
    example: { type: 'ack', request_id: '7aa2e6f7-7869-41be-8506-b2b6e5c17615', data: message },
  },
  {
    name: 'WsError',
    description: 'Ошибка внутри WebSocket; HTTP-статуса после Upgrade нет.',
    fields: [
      ['type', '"error"', true, 'Ошибка.'],
      ['request_id', 'UUID | null', true, 'ID команды либо null у неразбираемого кадра.'],
      ['error', 'WsErrorDetails', true, 'Причина ошибки.'],
    ],
    example: {
      type: 'error',
      request_id: null,
      error: { code: 'invalid_payload', message: 'Некорректный кадр.', details: null },
    },
  },
  {
    name: 'WsErrorDetails',
    description: 'Машинный код и текст для клиента.',
    fields: [
      ['code', 'string', true, 'Стабильный код.'],
      ['message', 'string', true, 'Текст без внутренних данных.'],
      ['details', 'object | null', true, 'fields для валидации; retry_after_ms для rate_limited.'],
    ],
    example: { code: 'chat_forbidden', message: 'Нет доступа к чату.', details: null },
  },
  {
    name: 'WsReady',
    description: 'Соединение авторизовано, команды разрешены.',
    fields: [
      ['user_id', 'integer', true, 'app_user.id.'],
      ['expires_at', 'datetime', true, 'Срок текущего допуска; затем получить новый ticket.'],
    ],
    example: { user_id: 12, expires_at: '2026-11-16T10:15:00Z' },
  },
  {
    name: 'MessageSendCommand',
    description: 'Запись message и message_file одной транзакцией. sender_id и даты задаёт сервер.',
    fields: [
      ['chat_id', 'integer', true, 'Чат, в котором пользователь есть в chat_participant.'],
      ['body', 'string', true, 'До 4000 символов; после trim нужен текст или хотя бы один файл.'],
      ['file_ids', 'integer[]', true, 'До 5 уникальных доступных файлов; [] без вложений.'],
    ],
    example: { chat_id: 44, body: 'Здравствуйте!', file_ids: [] },
  },
  {
    name: 'MessageMutation',
    description: 'Новое сообщение; last_message затем вычисляется по истории.',
    fields: [['message', 'Message', true, 'Полное сообщение.']],
    example: { message },
  },
]

export const websocketCommands: WebSocketContract[] = [
  {
    type: 'message.send',
    title: 'Отправить текст и вложения',
    direction: 'command',
    stage: 'core',
    dataSchema: 'MessageSendCommand',
    resultSchema: 'Message',
    errors: ['invalid_payload', 'chat_forbidden', 'file_not_found'],
    notes: [
      'Проверить chat_participant, ограничения body и до 5 вложений. Отправка блокируется до исправления типа и уникального индекса message.sender_id в DBML.',
      'После COMMIT вернуть ack и доставить message.created всем подключённым участникам, включая отправителя. Клиент объединяет ack и событие по message.id.',
      'Автоматически повторять неподтверждённую отправку нельзя: client_message_id отсутствует, повтор может создать дубль.',
    ],
  },
]
export const websocketEvents: WebSocketContract[] = [
  {
    type: 'session.ready',
    title: 'Соединение готово',
    direction: 'service',
    stage: 'core',
    dataSchema: 'WsReady',
    notes: [
      'Сервер проверяет активность пользователя и членство в чате при каждой команде и доставке.',
    ],
  },
  {
    type: 'message.created',
    title: 'Сообщение сохранено',
    direction: 'event',
    stage: 'core',
    dataSchema: 'MessageMutation',
    notes: ['Добавить по message.id; обновить карточку чата. Событие отправляется после COMMIT.'],
  },
]
export const websocketRules = [
  {
    title: 'Основание — активные поля DBML',
    text: 'Протокол использует chat, chat_participant, message и message_file. client_message_id, version, is_deleted, last_read_message_id и joined_at закомментированы и не входят в основной контракт. has_files вычисляется по message_file; last_message — по максимальному message.id.',
  },
  {
    title: 'Восстановление соединения',
    text: 'После обрыва получить новый ticket, подключиться, дождаться session.ready и загрузить GET /chats и историю активного чата. События, полученные во время HTTP-чтения, буферизовать; новые сообщения объединять по id. Редактирование и удаление отправленных сообщений недоступны. История — источник сохранённых данных; гарантий replay и доставки событий во время отключения нет.',
  },
  {
    title: 'Отдельные расширения хранения',
    text: 'Непрочитанные, chat.read, stream.resume, event_id, outbox, версии карточек и идемпотентная отправка требуют новых полей или хранилищ. Они не обещаются текущим контрактом. request_id — только корреляция; не заменяет уникальность сохранённой команды.',
  },
  {
    title: 'Ошибка в message.sender_id',
    text: 'В DBML sender_id имеет UUID, но ссылается на app_user.id типа int. UNIQUE(sender_id) допускает всего одно сообщение от пользователя. До реализации нужно согласовать целочисленный внешний ключ и обычный неуникальный индекс. Документация использует integer как тип ID связанного пользователя; это явное ожидающее исправление схемы.',
  },
]
export const websocketCloseCodes = [
  ['1000', 'Штатное завершение', 'Переподключаться при следующем открытии.'],
  [
    '1006',
    'Обрыв транспорта (локальный код браузера)',
    'Новый ticket, соединение и повторная загрузка истории; отправку автоматически не повторять.',
  ],
  [
    '4401',
    'Допуск истёк или аккаунт заблокирован',
    'Обновить авторизацию; при отказе открыть вход.',
  ],
  ['4403', 'Доступ запрещён', 'Закрыть чат и обновить список доступных чатов.'],
]
