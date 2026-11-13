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

const unread = { total: 0, version: 17 }
const chat = {
  id: 44,
  kind: 'direct',
  title: 'Петров Алексей Викторович',
  group_id: null,
  peer: { id: 2, name: 'Петров Алексей Викторович' },
  last_message: {
    id: 300,
    sender_id: 12,
    sender_name: 'Иванова Анна Сергеевна',
    body: 'Здравствуйте!',
    has_files: false,
    deleted: false,
    created_at: '2026-11-13T10:00:00Z',
  },
  unread: 0,
  view_version: 8,
}
const message = {
  id: 300,
  chat_id: 44,
  sender_id: 12,
  sender_name: 'Иванова Анна Сергеевна',
  client_message_id: 'bd53ed5a-1c8e-42b9-87b0-ae30e6a84412',
  version: 1,
  body: 'Здравствуйте!',
  files: [],
  created_at: '2026-11-13T10:00:00Z',
  edited_at: null,
  deleted: false,
}

export const websocketSchemas: ApiSchema[] = [
  {
    name: 'WebSocketTicket',
    description:
      'Одноразовый допуск к Upgrade. Не является JWT и не используется как постоянный токен.',
    fields: [
      ['ticket', 'string', true, 'Непредсказуемый одноразовый секрет, TTL 30 секунд.'],
      ['expires_at', 'datetime', true, 'Когда ticket перестаёт приниматься.'],
      ['endpoint', '"/api/v1/realtime"', true, 'Путь Upgrade на том же хосте.'],
      [
        'protocol',
        '"lk-chat.v1"',
        true,
        'Sec-WebSocket-Protocol для браузерного конструктора WebSocket.',
      ],
    ],
    example: {
      ticket: '<one-time-ticket>',
      expires_at: '2026-11-13T10:00:30Z',
      endpoint: '/api/v1/realtime',
      protocol: 'lk-chat.v1',
    },
  },
  {
    name: 'UnreadState',
    description:
      'Общий счётчик пользователя по всем доступным чатам, а не только загруженной странице.',
    fields: [
      ['total', 'integer', true, 'Непрочитанные чужие сообщения >= 0.'],
      [
        'version',
        'integer',
        true,
        'Монотонная версия персонального счётчика. При объединении REST/WS применять только более новую.',
      ],
    ],
    example: unread,
  },
  {
    name: 'ChatSnapshotMeta',
    description: 'Метаданные первой страницы списка в согласованном снимке.',
    fields: [
      ['limit', 'integer', true, 'Запрошенный размер страницы.'],
      ['offset', '0', true, 'Снимок всегда начинается с первой страницы.'],
      ['total', 'integer', true, 'Всего доступных чатов.'],
    ],
    example: { limit: 50, offset: 0, total: 1 },
  },
  {
    name: 'MessagingSnapshot',
    description:
      'Чаты, счётчик и курсор одного согласованного чтения. Опорная точка для stream.resume.',
    fields: [
      ['chats', 'Chat[]', true, 'Первая страница с последними сообщениями без полной истории.'],
      ['meta', 'ChatSnapshotMeta', true, 'Пагинация чатов.'],
      ['unread', 'UnreadState', true, 'Общий счётчик.'],
      [
        'resume_after_event_id',
        'string',
        true,
        'Позиция пользовательского потока, соответствующая этому снимку; десятичная строка, "0" для пустого потока.',
      ],
    ],
    example: {
      chats: [{ ...chat, last_message: null, view_version: 7 }],
      meta: { limit: 50, offset: 0, total: 1 },
      unread,
      resume_after_event_id: '780',
    },
  },
  {
    name: 'WsCommandEnvelope',
    description:
      'Одна команда клиента — один текстовый JSON-кадр. data имеет конкретную схему команды.',
    fields: [
      ['type', 'string', true, 'Тип из каталога команд.'],
      [
        'request_id',
        'UUID',
        true,
        'Корреляция ack/error и повторов. Повторять тот же id при повторе той же команды.',
      ],
      ['data', 'object', true, 'Параметры команды; {} если параметров нет.'],
    ],
    example: {
      type: 'message.send',
      request_id: '7aa2e6f7-7869-41be-8506-b2b6e5c17615',
      data: { chat_id: 44, client_message_id: message.client_message_id, body: 'Здравствуйте!' },
    },
  },
  {
    name: 'WsEventEnvelope',
    description:
      'Долговечное событие персонального потока. Доставка at least once, возможны повторы.',
    fields: [
      ['type', 'string', true, 'Тип события.'],
      [
        'event_id',
        'string',
        true,
        'Строго возрастающая десятичная строка в потоке пользователя. Возможны пропуски чисел. Не преобразовывать в Number.',
      ],
      ['occurred_at', 'datetime', true, 'Время фиксации изменения.'],
      ['data', 'object', true, 'Персонализированная нагрузка события; чужие счётчики не выдаются.'],
    ],
    example: {
      type: 'message.created',
      event_id: '781',
      occurred_at: '2026-11-13T10:00:00Z',
      data: { message, chat, unread },
    },
  },
  {
    name: 'WsAck',
    description:
      'Подтверждение команды. У ack нет event_id: он не является элементом воспроизводимого потока.',
    fields: [
      ['type', '"ack"', true, 'Успешное выполнение команды.'],
      ['request_id', 'UUID', true, 'ID исходной команды.'],
      [
        'data',
        'object',
        true,
        'Результат по resultSchema команды; ack send означает сохранение на сервере, не доставку/прочтение адресатом.',
      ],
    ],
    example: { type: 'ack', request_id: '7aa2e6f7-7869-41be-8506-b2b6e5c17615', data: message },
  },
  {
    name: 'WsError',
    description: 'Ошибка протокола/команды внутри соединения. HTTP-статуса после Upgrade нет.',
    fields: [
      ['type', '"error"', true, 'Тип кадра ошибки.'],
      ['request_id', 'UUID | null', true, 'ID команды либо null, если кадр нельзя разобрать.'],
      [
        'error',
        'WsErrorDetails',
        true,
        'Машинный код, текст, детали и допустимость повторной попытки.',
      ],
    ],
    example: {
      type: 'error',
      request_id: '7aa2e6f7-7869-41be-8506-b2b6e5c17615',
      error: {
        code: 'chat_forbidden',
        message: 'Нет доступа к чату.',
        retryable: false,
        details: null,
      },
    },
  },
  {
    name: 'WsErrorDetails',
    description: 'Ошибка в WebSocket; коды и условия определены для каждой команды.',
    fields: [
      ['code', 'string', true, 'Стабильный код.'],
      ['message', 'string', true, 'Текст для пользователя.'],
      ['retryable', 'boolean', true, 'Можно повторить после устранения временной причины.'],
      [
        'details',
        'object | null',
        true,
        'Для rate_limited: {retry_after_ms:integer}; version_conflict: {current:Message}; resume_required: {action:"reload_snapshot"}; для валидации {fields:{имя_поля:string[]}}.',
      ],
    ],
    example: {
      code: 'rate_limited',
      message: 'Повторите отправку позже.',
      retryable: true,
      details: { retry_after_ms: 3000 },
    },
  },
  {
    name: 'WsReady',
    description: 'Первый служебный кадр сервера после Upgrade, без event_id.',
    fields: [
      ['connection_id', 'UUID', true, 'ID этого подключения.'],
      ['user_id', 'integer', true, 'Пользователь подтверждённой сессии.'],
      [
        'access_expires_at',
        'datetime',
        true,
        'Истечение допуска соединения, не позже access JWT, выдавшего ticket.',
      ],
      ['heartbeat_interval_ms', '25000', true, 'Интервал команды ping.'],
      ['heartbeat_timeout_ms', '10000', true, 'Ожидание ack ping.'],
      ['resume_window_seconds', '86400', true, 'Срок хранения событий для восстановления.'],
      [
        'features',
        'string[]',
        true,
        'Поддерживаемые расширения: message.edit, message.delete, message.attachments. Базовый набор команд доступен всегда; без расширений [].',
      ],
    ],
    example: {
      connection_id: '2dc265bd-d24e-4760-9f6f-99ee65a635c9',
      user_id: 12,
      access_expires_at: '2026-11-13T10:15:00Z',
      heartbeat_interval_ms: 25000,
      heartbeat_timeout_ms: 10000,
      resume_window_seconds: 86400,
      features: [],
    },
  },
  {
    name: 'WsExpiring',
    description: 'Служебное предупреждение об истечении допуска соединения, без event_id.',
    fields: [
      [
        'expires_at',
        'datetime',
        true,
        'Сервер закрывает соединение в этот момент; предупреждение за 30 секунд.',
      ],
    ],
    example: { expires_at: '2026-11-13T10:15:00Z' },
  },
  {
    name: 'StreamResumeInput',
    description: 'Возобновить поток с сохранённой позиции либо позиции согласованного снимка.',
    fields: [
      [
        'after_event_id',
        'string',
        true,
        'Курсор пользователя, после которого нужно воспроизвести события; десятичная строка >= "0".',
      ],
    ],
    example: { after_event_id: '780' },
  },
  {
    name: 'StreamResumeResult',
    description:
      'ack приходит после всех воспроизведённых событий, перед переходом к live-доставке.',
    fields: [
      [
        'through_event_id',
        'string',
        true,
        'Позиция, до которой поток гарантированно восстановлен. Клиент сохраняет её даже если часть событий скрыта правами.',
      ],
      ['replayed_count', 'integer', true, 'Число доставленных кадров воспроизведения.'],
    ],
    example: { through_event_id: '785', replayed_count: 5 },
  },
  {
    name: 'MessageSendCommand',
    description:
      'Базовый сценарий формы — непустой текст до 4000 символов. Вложения — отдельное расширение.',
    fields: [
      ['chat_id', 'integer', true, 'Доступный чат.'],
      [
        'client_message_id',
        'UUID',
        true,
        'Стабильный ID оптимистичного сообщения и всех повторов; уникален на отправителя.',
      ],
      [
        'body',
        'string',
        true,
        'После trim 1–4000 символов. Пустой текст допустим только при включённом message.attachments и непустом file_ids.',
      ],
      [
        'file_ids',
        'integer[]',
        false,
        'Расширение message.attachments: до 5 уникальных ID предварительно загруженных собственных файлов. Базовый клиент поле не отправляет.',
      ],
    ],
    example: { chat_id: 44, client_message_id: message.client_message_id, body: 'Здравствуйте!' },
  },
  {
    name: 'MessageEditCommand',
    description:
      'Необязательное расширение: правка собственного текста с защитой от потерянного изменения.',
    fields: [
      ['chat_id', 'integer', true, 'Чат сообщения.'],
      ['message_id', 'integer', true, 'Собственное сообщение.'],
      ['expected_version', 'integer', true, 'Известная клиенту версия.'],
      [
        'body',
        'string',
        true,
        '1–4000 символов; при разрешённых вложениях может быть пустым, если файлы остаются.',
      ],
    ],
    example: {
      chat_id: 44,
      message_id: 300,
      expected_version: 1,
      body: 'Здравствуйте! Отправляю уточнение.',
    },
  },
  {
    name: 'MessageDeleteCommand',
    description: 'Необязательное расширение: мягкое удаление своего сообщения.',
    fields: [
      ['chat_id', 'integer', true, 'Чат сообщения.'],
      ['message_id', 'integer', true, 'Собственное сообщение.'],
      [
        'expected_version',
        'integer',
        true,
        'Версия перед удалением. Уже удалённое своё сообщение возвращается повторно без ошибки.',
      ],
    ],
    example: { chat_id: 44, message_id: 300, expected_version: 1 },
  },
  {
    name: 'ChatReadCommand',
    description: 'Обновить персональный указатель только до действительно показанного сообщения.',
    fields: [
      ['chat_id', 'integer', true, 'Текущий доступный чат.'],
      [
        'message_id',
        'integer',
        true,
        'ID сообщения этого чата, показанного пользователю; не последний полученный в фоне.',
      ],
    ],
    example: { chat_id: 44, message_id: 300 },
  },
  {
    name: 'ChatReadState',
    description:
      'Результат прочтения для собственных устройств. Индикатор «прочитано адресатом» текущий UI не показывает.',
    fields: [
      ['chat_id', 'integer', true, 'Чат.'],
      ['last_read_message_id', 'integer', true, 'max(старый указатель, новый).'],
      ['read_at', 'datetime', true, 'Время последнего продвижения указателя.'],
      ['chat', 'Chat', true, 'Персональная карточка с актуальным unread/view_version.'],
      ['unread', 'UnreadState', true, 'Общий счётчик пользователя.'],
    ],
    example: {
      chat_id: 44,
      last_read_message_id: 300,
      read_at: '2026-11-13T10:01:00Z',
      chat: { ...chat, view_version: 9 },
      unread: { total: 0, version: 18 },
    },
  },
  {
    name: 'MessageMutation',
    description: 'Сообщение плюс персональные данные чата и счётчик получателя события.',
    fields: [
      ['message', 'Message', true, 'Полное сообщение, новая версия или tombstone при удалении.'],
      [
        'chat',
        'Chat',
        true,
        'Актуальный last_message/unread и view_version для этого пользователя.',
      ],
      [
        'unread',
        'UnreadState',
        true,
        'Авторитетный общий счётчик. Не увеличивать локально на +1 при каждом повторном событии.',
      ],
    ],
    example: { message, chat, unread },
  },
  {
    name: 'ChatMutation',
    description: 'Создание/изменение карточки чата без загрузки полной истории.',
    fields: [
      ['chat', 'Chat', true, 'Карточка для пользователя.'],
      ['unread', 'UnreadState', true, 'Общий счётчик.'],
      ['total_chats', 'integer', true, 'Всего доступных чатов для обновления пагинации.'],
    ],
    example: { chat, unread, total_chats: 1 },
  },
  {
    name: 'ChatRemoval',
    description: 'Утрата доступа к групповому чату, без содержания недоступной переписки.',
    fields: [
      ['chat_id', 'integer', true, 'Скрываемый чат.'],
      ['reason', '"membership_revoked" | "group_archived"', true, 'Причина.'],
      ['unread', 'UnreadState', true, 'Счётчик после исключения чата.'],
      ['total_chats', 'integer', true, 'Новое число доступных чатов.'],
    ],
    example: {
      chat_id: 44,
      reason: 'membership_revoked',
      unread: { total: 0, version: 19 },
      total_chats: 0,
    },
  },
  {
    name: 'PingCommand',
    description: 'Прикладной heartbeat. data всегда пустой объект.',
    fields: [],
    example: {},
  },
  {
    name: 'PongResult',
    description: 'Результат heartbeat в ack, без event_id.',
    fields: [['server_time', 'datetime', true, 'Текущее время сервера.']],
    example: { server_time: '2026-11-13T10:00:25Z' },
  },
]

export const websocketCommands: WebSocketContract[] = [
  {
    type: 'stream.resume',
    title: 'Восстановить поток и перейти к live-доставке',
    direction: 'command',
    stage: 'core',
    dataSchema: 'StreamResumeInput',
    resultSchema: 'StreamResumeResult',
    errors: ['resume_required', 'invalid_cursor', 'already_resumed'],
    notes: [
      'После session.ready: взять курсор локального состояния или GET /messaging/snapshot. Сервер фиксирует верхнюю границу E, воспроизводит доступные события (cursor,E], отправляет ack с through_event_id=E, затем только live >E. Между replay и live не допускаются пропуски.',
      'До ack этой команды разрешён только ping. При слишком старом, чужом или недостоверном курсоре — resume_required, без частичного replay. Клиент заново получает снимок и повторяет resume.',
      'События хранятся 24 часа; актуальные права проверяются и при replay. Закрытая переписка не раскрывается. Event_id может иметь пропуски; клиент не требует последовательности +1.',
    ],
  },
  {
    type: 'message.send',
    title: 'Отправить сообщение или повторить неподтверждённую отправку',
    direction: 'command',
    stage: 'core',
    dataSchema: 'MessageSendCommand',
    resultSchema: 'Message',
    errors: [
      'chat_forbidden',
      'empty_message',
      'message_too_long',
      'client_message_id_reused',
      'rate_limited',
      'unsupported_feature',
    ],
    notes: [
      'Сначала показать локальное сообщение со статусом «отправляется», связать его с client_message_id. Галочка «Сохранено» появляется только после ack либо message.created с тем же client_message_id.',
      'Уникальность sender_id/client_message_id хранить вместе с сообщением, без короткого TTL. Повтор с тем же содержимым возвращает тот же Message; другое содержимое с тем же ID — client_message_id_reused.',
      'ack только после фиксации сообщения и события в одной транзакции. message.created приходит всем активным устройствам участников, включая отправителя. ack и событие могут прийти в любом порядке; объединять по id/client_message_id.',
      'Нет ack за 10 секунд — состояние «отправка не подтверждена» и кнопка повтора. Не считать это доказательством потери. Повторяет те же request_id и client_message_id после восстановления соединения.',
    ],
  },
  {
    type: 'chat.read',
    title: 'Отметить показанные сообщения прочитанными',
    direction: 'command',
    stage: 'core',
    dataSchema: 'ChatReadCommand',
    resultSchema: 'ChatReadState',
    errors: ['chat_forbidden', 'message_chat_mismatch'],
    notes: [
      'Только активный показанный диалог, видимая вкладка и сообщения, которые действительно попали в область чтения. Фоновая доставка сама по себе не обнуляет unread.',
      'Указатель только увеличивается, собственные сообщения не увеличивают unread. Событие chat.read синхронизирует счётчики между устройствами этого пользователя.',
      'Базовый контракт не показывает собеседнику точное время прочтения: текущий UI показывает только сохранение сообщения на сервере.',
    ],
  },
  {
    type: 'ping',
    title: 'Проверить живое соединение',
    direction: 'command',
    stage: 'core',
    dataSchema: 'PingCommand',
    resultSchema: 'PongResult',
    notes: [
      'Каждые 25 секунд, ожидание ack 10 секунд. При таймауте закрыть сокет и перейти к переподключению.',
      'Браузер не позволяет отправить WebSocket control Ping вручную; здесь именно прикладной JSON ping. После сна/смены сети проверять связь сразу. Сервер закрывает соединение 4408, если нет ping или другой команды более 45 секунд; возобновление идёт по курсору.',
    ],
  },
  {
    type: 'message.edit',
    title: 'Изменить собственное сообщение',
    direction: 'command',
    stage: 'extension',
    dataSchema: 'MessageEditCommand',
    resultSchema: 'Message',
    errors: [
      'unsupported_feature',
      'message_not_owned',
      'message_deleted',
      'version_conflict',
      'message_too_long',
    ],
    notes: [
      'В интерфейсе сейчас нет редактирования сообщений. Команда включается только при feature=message.edit и отдельной реализации действия на фронтенде.',
      'Проверить автора и принадлежность чату. При несовпадении expected_version вернуть version_conflict с текущим Message. Увеличить Message.version, обновить edited_at и отправить message.updated.',
    ],
  },
  {
    type: 'message.delete',
    title: 'Удалить собственное сообщение',
    direction: 'command',
    stage: 'extension',
    dataSchema: 'MessageDeleteCommand',
    resultSchema: 'Message',
    errors: ['unsupported_feature', 'message_not_owned', 'version_conflict'],
    notes: [
      'В интерфейсе сейчас нет удаления сообщения. Команда требует feature=message.delete.',
      'Мягко удалить: deleted=true, body="", files=[], версия увеличивается. Участникам отправить message.deleted с полной tombstone-моделью; id/автор/дата остаются.',
      'Повтор для уже удалённого своего сообщения возвращает ту же tombstone; не восстанавливать сообщение повтором edit.',
    ],
  },
]

export const websocketEvents: WebSocketContract[] = [
  {
    type: 'session.ready',
    title: 'Соединение принято',
    direction: 'service',
    stage: 'core',
    dataSchema: 'WsReady',
    notes: [
      'Первый кадр после 101. Без event_id, не сохраняется и не воспроизводится. Пользователь должен совпадать с текущим профилем.',
      'До stream.resume сервер не переводит соединение в live-режим. Пользовательские данные потока не должны уходить до проверки ticket и Origin.',
    ],
  },
  {
    type: 'session.expiring',
    title: 'Допуск соединения истекает',
    direction: 'service',
    stage: 'core',
    dataSchema: 'WsExpiring',
    notes: [
      'За 30 секунд до access_expires_at. Обновить HTTP-сессию, получить новый ticket, переподключиться и выполнить resume с последним применённым event_id.',
      'Без event_id. access_token через команды/URL сокета не передаётся. При истечении — закрытие 4401.',
    ],
  },
  {
    type: 'message.created',
    title: 'Сообщение сохранено и доступно участнику',
    direction: 'event',
    stage: 'core',
    dataSchema: 'MessageMutation',
    notes: [
      'Добавить/объединить сообщение в истории, обновить последнее сообщение и порядок чатов, применить персональные счётчики.',
      'Для собственного client_message_id заменить оптимистичную запись, не создавать второй пузырь. Скрытый диалог обновляет бейдж, а не отмечается прочитанным.',
      'Повторная доставка не создаёт второе сообщение и не увеличивает счётчики повторно.',
    ],
  },
  {
    type: 'chat.created',
    title: 'Новый доступный диалог или групповой чат',
    direction: 'event',
    stage: 'core',
    dataSchema: 'ChatMutation',
    notes: [
      'Вставить чат по id либо объединить с ответом HTTP-открытия диалога. Создание чата из другой вкладки сразу видно здесь.',
    ],
  },
  {
    type: 'chat.updated',
    title: 'Изменились название или другие сведения чата',
    direction: 'event',
    stage: 'core',
    dataSchema: 'ChatMutation',
    notes: [
      'Применять Chat только при не меньшей view_version. Для last_message/unread, уже переданных в message.created/chat.read, отдельный дублирующий chat.updated не требуется.',
    ],
  },
  {
    type: 'chat.read',
    title: 'Прочтение синхронизировано между своими устройствами',
    direction: 'event',
    stage: 'core',
    dataSchema: 'ChatReadState',
    notes: [
      'Событие только пользователю, который прочитал: его вкладкам/устройствам. Применить chat.unread/view_version и UnreadState.version.',
      'Фоновое устройство не должно повторно отправлять chat.read только из-за получения этого события.',
    ],
  },
  {
    type: 'chat.removed',
    title: 'Пользователь утратил доступ к групповому чату',
    direction: 'event',
    stage: 'core',
    dataSchema: 'ChatRemoval',
    notes: [
      'Удалить чат из доступного списка и закрыть активную переписку; не удалять исторические данные на сервере.',
      'Пользователь больше не получает содержимое этого чата, в том числе при replay. Курсор потока продолжает двигаться.',
    ],
  },
  {
    type: 'message.updated',
    title: 'Обновлена версия сообщения',
    direction: 'event',
    stage: 'extension',
    dataSchema: 'MessageMutation',
    notes: [
      'Соответствует расширению message.edit. Обновить по id при более новой версии; повторно в историю не добавлять.',
      'Если сообщение вне загруженной истории, достаточно обновить preview чата из payload; не загружать всю историю.',
    ],
  },
  {
    type: 'message.deleted',
    title: 'Сообщение заменено заглушкой удаления',
    direction: 'event',
    stage: 'extension',
    dataSchema: 'MessageMutation',
    notes: [
      'Соответствует расширению message.delete. Tombstone должна перекрывать старый HTTP-ответ благодаря Message.version.',
      'Сервер возвращает уже пересчитанные unread и last_message; клиент не угадывает счётчики по изменению текста.',
    ],
  },
]

export const websocketRules = [
  {
    title: 'Граница HTTP и WebSocket',
    text: 'HTTP: авторизация, одноразовый ticket, начальный снимок, поиск собеседников, открытие чата, страницы списка, история при прокрутке и загрузка файлов. WebSocket: отправка, подтверждение сохранения, новые сообщения, прочтение и синхронизация устройств. message.edit/delete и вложения описаны как расширения, так как их нет в текущем интерфейсе. Новые сообщения не запрашиваются таймером; отдельных HTTP send/read/edit/delete нет.',
  },
  {
    title: 'Кадры и версионирование',
    text: 'Subprotocol lk-chat.v1. Под кадром далее понимается одно текстовое WebSocket message с JSON-объектом, до 64 KiB UTF-8 после распаковки; транспортная фрагментация frames допустима; содержимое файлов по сокету не передаётся. Команда: {type,request_id,data}. Событие: {type,event_id,occurred_at,data}. Служебный кадр: {type,data}, без event_id. Ack: {type:"ack",request_id,data}. Ошибка: {type:"error",request_id,error}. Неизвестная команда → unknown_command, неверная схема → invalid_payload; соединение при одиночной бизнес-ошибке сохраняется. Поля identity/roles/server timestamps не принимаются от клиента. Ошибки до 101 имеют HTTP-формат, но браузерный WebSocket не предоставляет их body клиенту: onerror переводит интерфейс в состояние восстановления, а причину сессии уточняет HTTP-запрос ticket/refresh.',
  },
  {
    title: 'Первое открытие без потери событий',
    text: '1. GET /messaging/snapshot: получить чаты, unread и курсор B одного согласованного чтения. 2. POST /realtime/tickets и WebSocket Upgrade, дождаться session.ready. 3. Передать stream.resume(after_event_id=B). 4. Применить replay и дождаться ack resume; после этого разрешены send/read и начинается live-поток. История активного чата загружается HTTP в любом моменте, объединение по Message.id/version исключает перезапись свежих событий устаревшим ответом. Chat.view_version и UnreadState.version защищают карточки и счётчики.',
  },
  {
    title: 'Переподключение и курсор',
    text: 'При сетевом обрыве задержки 1,2,4,8,16,30 секунд плюс случайный jitter до 20%; не открывать параллельные сокеты на один стор вкладки. Получить свежий ticket и возобновить после последнего применённого event_id. Курсор и соответствующее локальное состояние сохранять вместе, отдельно на пользователя; если остался только курсор без данных — снова загрузить снимок. События доступны 24 часа. resume_required → остановить live, перечитать снимок и загруженные истории, затем resume с новым курсором. При отсутствии доступа исторический payload не выдаётся; ack resume всё равно сообщает through_event_id.',
  },
  {
    title: 'Порядок, повторы и подтверждения',
    text: 'Порядок сохраняется в одном персональном потоке; event_id — строковое целое, сравнивать BigInt, пропуски допустимы. Транспорт доставляет at least once; повторное event_id не применять. Message.id/client_message_id и версии дополнительно защищают объединение REST, ack и событий. Для изменяющих команд message.send/edit/delete и chat.read сервер хранит результат request_id и хеш запроса 24 часа; иной payload с тем же request_id → request_id_reused. stream.resume и ping выполняются заново на каждом соединении с новым request_id. После resume_required повтор resume с новым курсором также имеет новый request_id. Для send постоянная уникальность client_message_id на отправителя защищает и после истечения этого окна. Ack означает фиксацию на сервере, а не доставку или прочтение адресатом.',
  },
  {
    title: 'Запись и доставка',
    text: 'Сообщение/указатель прочтения, счётчики и событие фиксируются одной транзакцией; после commit событие надёжно публикуется из outbox. Порядок номеров пользовательского потока должен соответствовать порядку зафиксированных изменений. Потеря соединения после commit до ack не создаёт дубля при повторе. Несколько серверных узлов доставляют событие всем активным соединениям получателя. Не считать наличие сокета подтверждением прочтения.',
  },
  {
    title: 'Права и утрата доступа',
    text: 'Ticket привязан к отзываемой HTTP-сессии, пользователь не присылает произвольный user_id для подписки. Все доступные чаты подключаются автоматически. У личного чата доступ только у двух участников; администратор не читает чужие личные диалоги. Учебные связи определяют разрешённые новые контакты и членство групповых чатов; уже созданная личная переписка сохраняется у двух участников. При переводе в другую группу/снятии преподавательской нагрузки групповой доступ пересчитывается, отправляется chat.removed, последующие команды отклоняются. Блокировка/выход/отзыв сессии немедленно закрывают связанные соединения.',
  },
  {
    title: 'Жизненный цикл допуска и нагрузки',
    text: 'Соединение действует не дольше access_expires_at, указанного в session.ready. За 30 секунд сервер отправляет session.expiring; клиент обновляет HTTP-сессию и ticket, затем переподключается с курсором. Команды до ack resume → stream_not_ready. Ограничения: до 5 message.send в секунду и 100 в минуту на пользователя; остальные команды до 20 в секунду, кроме resume — один на соединение. rate_limited содержит retry_after_ms. При слишком медленном клиенте не держать безразмерную очередь: закрыть 1013, продолжить через resume. Лимиты — проектные значения, настраиваемые сервером.',
  },
  {
    title: 'Ожидаемое поведение интерфейса',
    text: 'Во время отсутствия связи сохранять черновик и неподтверждённые отправки с исходными ID. Показывать состояния «подключение/восстановление/нет связи», не писать «сохранено» до подтверждения. Галочка текущего UI означает «сохранено на сервере». Chat unread обновляет список, виджет главной и навигацию сразу. chat.read отправляется при фактическом показе в видимой вкладке, не при открытом в фоне сокете. После восстановления обновить состояние до возобновления очереди отправок. Набор текста, online/last_seen и чужие read receipts в обязательный сценарий не входят.',
  },
]

export const websocketCloseCodes = [
  [
    '1000',
    'Нормальное закрытие',
    'После намеренного выхода не переподключаться. При замене соединения из-за обновления допуска продолжить с новым ticket и курсором.',
  ],
  [
    '1006 / 1011',
    'Обрыв сети / временная ошибка',
    'Переподключение с задержкой и resume; 1006 наблюдается клиентом, сервер не отправляет его в Close frame.',
  ],
  [
    '1009',
    'Сообщение больше 64 KiB',
    'Исправить payload; без бесконечного повтора такого же кадра.',
  ],
  ['1013', 'Перегрузка или медленный клиент', 'Переподключиться с backoff и восстановить поток.'],
  [
    '4400',
    'Неподдерживаемая версия/нарушение протокола',
    'Ошибка клиента; автоматический цикл переподключений остановить.',
  ],
  [
    '4401',
    'Истёк/отозван допуск',
    'Один раз обновить HTTP-сессию и ticket; при отказе очистить состояние и открыть вход. При выходе/блокировке refresh не восстановит сессию.',
  ],
  [
    '4403',
    'Запрещённый Origin или доступ',
    'Не переподключаться автоматически; проверить домен и право подключения.',
  ],
  ['4408', 'Истёк heartbeat', 'Сетевое восстановление с новым ticket и resume.'],
]
