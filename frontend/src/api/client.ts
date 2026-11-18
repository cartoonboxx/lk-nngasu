const messages: Record<string, string> = {
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

export async function apiRequest<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers)
  if (options.body) headers.set('Content-Type', 'application/json')
  let response: Response
  try {
    response = await fetch(`/api/v1${path}`, {
      ...options,
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
  return body.data as T
}
