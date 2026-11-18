const messages: Record<string, string> = {
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
  current_semester_in_use: 'Нельзя архивировать текущий семестр.',
  already_exists: 'Запись с такими данными уже существует.',
  in_use: 'Запись используется в других данных.',
  course_exists: 'Дисциплина уже назначена этой группе в выбранном семестре.',
  schedule_conflict: 'Занятие пересекается с другим занятием группы, преподавателя или аудитории.',
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
  paginated?: boolean
}

export async function apiRequest<T>(path: string, options: ApiRequestOptions = {}): Promise<T> {
  const { paginated, ...requestOptions } = options
  const headers = new Headers(options.headers)
  if (options.body) headers.set('Content-Type', 'application/json')
  let response: Response
  try {
    response = await fetch(`/api/v1${path}`, {
      ...requestOptions,
      headers,
      credentials: 'include',
      cache: 'no-store',
      signal: options.signal ?? AbortSignal.timeout(15000),
    })
  } catch {
    throw new Error('Не удалось связаться с сервером. Проверьте подключение и повторите попытку.')
  }
  if (response.status === 204) return undefined as T
  const body = await response.json().catch(() => null)
  if (!response.ok) {
    throw new ApiError(
      response.status,
      body?.error?.code ?? 'service_unavailable',
      body?.error?.details?.fields,
    )
  }
  if (!body || !('data' in body)) {
    throw new Error('Сервер вернул некорректный ответ. Повторите попытку позже.')
  }
  return (paginated ? body : body.data) as T
}
