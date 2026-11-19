export const apiNotifications: { onError: ((error: unknown) => void) | null } = { onError: null }

const messages: Record<string, string> = {
  network_error: 'Не удалось связаться с сервером. Проверьте подключение и повторите попытку.',
  invalid_response: 'Сервер вернул некорректный ответ. Повторите попытку позже.',
  profile_name_admin_only: 'Изменить ФИО может только администратор.',
  request_already_reviewed:
    'Заявка уже рассмотрена другим администратором. Показано актуальное состояние.',
  registration_request_not_found: 'Заявка не найдена.',
  invalid_reference: 'Выбранная связанная запись не существует или находится в архиве.',
  invalid_role: 'Выберите одну учебную роль: студент или преподаватель.',
  record_book_exists: 'Номер зачётной книжки уже занят.',
  last_active_admin: 'Нельзя заблокировать последнего активного администратора.',
  self_lockout: 'Нельзя заблокировать свой аккаунт.',
  workload_exists: 'Такая нагрузка уже существует.',
  teacher_inactive: 'Аккаунт выбранного преподавателя заблокирован.',
  invalid_period: 'Выберите время пары из действующей сетки звонков.',
  current_semester_in_use: 'Нельзя удалить текущий семестр.',
  catalog_in_use: 'Запись используется в учебных данных и не может быть удалена.',
  catalog_entry_exists: 'Запись с такими данными уже существует.',
  news_not_found: 'Новость не найдена или снята с публикации.',
  news_archived: 'Перед редактированием верните новость из архива в черновики.',
  news_is_published: 'Перед удалением снимите новость с публикации.',
  already_reviewed: 'Проверенный ответ нельзя изменить или удалить.',
  already_submitted: 'Ответ уже отправлен. Для сохранения черновика сначала удалите ответ.',
  empty_submission: 'Прикрепите файл или напишите ответ.',
  assignment_has_submissions: 'У задания есть ответы студентов — удалить его нельзя.',
  topic_assignment_exists: 'В этой теме уже есть задание.',
  message_not_found: 'Сообщение уже удалено или недоступно.',
  chat_forbidden: 'Нет доступа к этому чату.',
  group_forbidden: 'Нет доступа к чату этой группы.',
  self_chat: 'Нельзя начать диалог с самим собой.',
  invalid_payload: 'Проверьте текст сообщения и вложения.',
  too_many_files: 'Можно прикрепить не более пяти файлов.',
  no_submissions: 'Отправленных работ пока нет.',
  already_exists: 'Запись с такими данными уже существует.',
  in_use: 'Запись используется в других данных.',
  course_exists: 'Дисциплина уже назначена этой группе в выбранном семестре.',
  schedule_conflict: 'Занятие пересекается с другим занятием группы, преподавателя или аудитории.',
  current_semester_not_set: 'Текущий семестр пока не назначен.',
  academic_calendar_not_configured: 'Учебный календарь пока не настроен.',
  course_forbidden: 'Нет доступа к этому курсу.',
  chat_not_found: 'Чат не найден.',
  chat_with_self: 'Нельзя открыть диалог с самим собой.',
  contact_not_allowed:
    'Нельзя начать диалог с этим пользователем. Доступны одногруппники и преподаватели вашей группы.',
  peer_inactive: 'Аккаунт собеседника заблокирован.',
  empty_answer: 'Прикрепите файл или напишите ответ.',
  password_unchanged: 'Новый пароль должен отличаться от текущего.',
  file_not_found: 'Файл не найден или недоступен.',
  file_too_large: 'Файл превышает допустимый размер.',
  unsupported_file_type: 'Этот тип файла не поддерживается.',
  invalid_file_content: 'Содержимое файла не соответствует его расширению.',
  submission_reviewed: 'Проверенный ответ нельзя изменить.',
  no_personal_schedule: 'У аккаунта нет личного расписания.',
  forbidden: 'Недостаточно прав для выполнения действия.',

  invitations_disabled: 'Для приглашения администраторов нужно настроить SMTP на сервере.',
  owner_protected: 'Изменять аккаунт владельца может только сам владелец.',
  setup_closed: 'Первоначальная настройка отключена или уже завершена.',
  setup_key_invalid: 'Неверный ключ первоначальной настройки.',
  owner_required: 'Управлять администраторами может только владелец системы.',
  wrong_password: 'Неверный текущий пароль.',
  already_admin: 'Пользователь уже является администратором.',
  account_not_ready: 'Нужен активный пользователь с заданным паролем.',
  user_not_found: 'Пользователь не найден.',
  invitation_invalid: 'Ссылка приглашения недействительна или заменена новой.',
  invitation_expired: 'Срок приглашения истёк. Попросите отправить новое.',
  already_activated: 'Пароль уже задан. Перейдите ко входу.',

  invalid_credentials: 'Неверная электронная почта или пароль.',
  account_pending:
    'Заявка ещё рассматривается. Вход станет доступен после одобрения администратором.',
  account_rejected: 'Заявка отклонена. Для доступа подайте новую заявку на регистрацию.',
  account_inactive: 'Аккаунт заблокирован. Обратитесь к администратору.',
  registration_exists: 'Заявка с этой почтой уже ожидает рассмотрения.',
  email_exists: 'Аккаунт с этой почтой уже существует. Попробуйте войти.',
  rate_limited: 'Слишком много попыток. Повторите позже.',
  session_expired: 'Сессия завершена. Войдите снова.',
  refresh_reused: 'Сессия завершена. Войдите снова.',
  csrf_invalid: 'Не удалось подтвердить сессию. Обновите страницу и повторите попытку.',
  invalid_request: 'Проверьте заполнение полей формы.',
}

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    public fields: Record<string, string[]> = {},
  ) {
    super(messages[code] ?? 'Сервис временно недоступен. Попробуйте позже.')
  }
}

export interface ApiPage<T> {
  data: T[]
  meta: { limit: number; offset: number; total: number }
}

export interface ApiRequestOptions extends RequestInit {
  notifyOnError?: boolean
  responseType?: 'blob'
  paginated?: boolean
  // Получить актуальные данные с сервера, заменив сохранённый ответ в Pinia.
  fresh?: boolean
}

export async function apiRequest<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const {
    paginated,
    responseType,
    fresh: _fresh,
    notifyOnError = true,
    ...requestOptions
  } = options
  const headers = new Headers(options.headers)
  if (options.body && !(options.body instanceof FormData))
    headers.set('Content-Type', 'application/json')
  try {
    let response: Response
    try {
      response = await fetch(`/api/v1${path}`, {
        ...requestOptions,
        headers,
        credentials: 'include',
        cache: 'no-store',
        signal: options.signal ?? AbortSignal.timeout(15000),
      })
    } catch (error) {
      if (options.signal?.aborted) throw error
      throw new ApiError(0, 'network_error')
    }
    if (response.status === 204) return undefined as T
    if (response.ok && responseType === 'blob') return (await response.blob()) as T
    const body = await response.json().catch(() => null)
    if (!response.ok) {
      throw new ApiError(
        response.status,
        body?.error?.code ?? 'service_unavailable',
        body?.error?.details?.fields,
      )
    }
    if (!body || !('data' in body)) throw new ApiError(response.status, 'invalid_response')
    return (paginated ? body : body.data) as T
  } catch (error) {
    if (notifyOnError && !options.signal?.aborted) apiNotifications.onError?.(error)
    throw error
  }
}
