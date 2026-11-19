# university-api

REST API личного кабинета вуза: **Go + Gin + pgx + PostgreSQL**, модульный монолит.
Схема БД `university_cabinet` задаётся миграциями `migrations/NNN_*.sql`; они же — источник правды о схеме.

## Быстрый старт

```bash
cd ..                               # команды из корня репозитория
export JWT_SECRET=$(openssl rand -hex 32)
export ADMIN_SETUP_KEY=$(openssl rand -hex 32) # сохраните ключ для формы /setup
docker compose up -d --build --wait  # фронтенд, API, PostgreSQL и Adminer

# откройте /setup на фронтенде и создайте владельца по ADMIN_SETUP_KEY
# затем войдите с указанными в форме почтой и паролем

curl -s localhost:5173/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"сменить-меня-123"}'
# -> {"data":{"access_token":"...","token_type":"Bearer","expires_at":"...","user":{"id":1,"is_admin":true,...}}}
# дальше: Authorization: Bearer <token>
```

Браузер и HTTP-клиенты обращаются к API через фронтенд: `/api/v1/...`.
Vite проксирует запросы в контейнер `backend:8080`. API применяет недостающие
миграции автоматически и на новой, и на существующей БД.
Загруженные файлы сохраняются в отдельном томе `uploads`.

Проверка работоспособности без авторизации: `GET /api/healthz` возвращает
`{"status":"ok"}` через Vite и продакшен-nginx. Внутренний `GET /healthz`
используется healthcheck контейнера и возвращает тот же ответ.

Без Docker экспортируйте `DATABASE_URL` с адресом PostgreSQL и `JWT_SECRET`
длиной не менее 32 символов, затем запустите `make run`. Для локального HTTP
укажите `COOKIE_SECURE=false`. При необходимости задайте `ADMIN_SETUP_KEY`
и откройте `/setup` на фронтенде. API читает переменные окружения;
файл `.env` самостоятельно не загружается. Остальные настройки и значения
по умолчанию описаны в [config.go](internal/config/config.go).

Первый администратор создаётся через одноразовую настройку; следующих владелец
приглашает из `/admin/administrators`. Подробности и конфигурация SMTP —
[создание владельца и администраторов](docs/admin-access.md).

## Миграции

API применяет новые миграции сам при старте (`MIGRATE_ON_START=true` по умолчанию), поэтому
после деплоя база обновляется автоматически. Применённые версии хранятся в таблице `schema_migrations`;
повторно файл не выполняется. Каждая миграция идёт в своей транзакции: при ошибке база не меняется,
а сервер не стартует и пишет в лог имя файла и ошибку.

- Новая миграция: файл `migrations/011_short_name.sql` со следующим свободным номером.
  `BEGIN`/`COMMIT` писать не нужно. Уже применённые файлы не редактировать — только новый номер.
- Обновить базу без запуска сервера: `make migrate` (или `go run ./cmd/migrate`).
- База, в которую `001`–`010` накатывали вручную через `psql`: при первом запуске они распознаются
  по признакам в схеме и отмечаются применёнными, без повторного выполнения.

## Структура

```
cmd/api            точка входа (при старте применяет миграции)
cmd/migrate        применить миграции без запуска сервера
cmd/createadmin    создание первого администратора
internal/app       сборка зависимостей и роутер (единственное место, где модули знают друг о друге)
internal/platform  apperr, httpx, middleware (JWT, роли), database (DB/InTx/Collect, ошибки PG, Filter),
                   Migrate (миграции), access (права на курс и файл), filekey, filegc (удаление
                   неиспользуемых файлов), validate, mail (очередь писем), ratelimit
internal/modules   auth, users, structure, courses, schedule, content, assignments, files, chat,
                   news, realtime (WebSocket), resources
migrations         001–017 *.sql, встроены в бинарник (embed.go)
```

Модуль = `module.go` (маршруты) → `handler` → `service` (правила и права) → `repository` (SQL).
Модули не импортируют друг друга: нужные им проверки (`Guard`, `Blobs`, `FileOwnership`) описаны интерфейсами в самом модуле,
а реализации подставляет `internal/app`. У справочников (`structure`) сервисного слоя нет.

## Роли и права

Роль у пользователя одна (`role` в профиле и в токене): `admin` (`app_user.is_admin`, профилей нет), `teacher`
(запись в `teacher`), `student` (запись в `student`). Выдать или снять роль `admin` может только владелец (`can_manage_admins`).

- студент видит курсы своей группы; преподаватель — курсы, где у него есть нагрузка; админ — всё;
- студент видит только свои ответы и оценки;
- **админ** управляет справочниками, курсами, нагрузкой, расписанием и пользователями, может создавать и править
  темы, материалы и задания; оценки ставит только преподаватель курса
  (`reviewed_by` ссылается на `teacher`);
- к материалу, заданию или ответу можно прикрепить только свой файл (папка `<user_id>/` в ключе хранения).

## Формат ответов

Успех: `{"data": ...}`; ошибка: `{"error": {"code": "...", "message": "..."}}`; удаление — `204` без тела.

## Эндпоинты

Префикс `/api/v1`. «Все» = любой авторизованный пользователь. Публичные — без токена.
Полный контракт (поля, ошибки, правила) — `frontend/src/pages/apiDocs/api-docs.*.ts`; ниже — краткая карта.

### Аутентификация и профиль
| Метод | Путь | Кто | Описание |
|---|---|---|---|
| POST | /auth/login | публично | вход по email и паролю |
| GET, POST | /auth/setup | публично | первый администратор-владелец по `ADMIN_SETUP_KEY` |
| GET / PATCH | /me | все | профиль (`role`, `student`/`teacher`) / ФИО и аватар (`avatar`) |
| POST | /me/password | все | смена пароля (`old_password`, `new_password`), отзывает все сессии |
| GET / POST | /users | админ | список (`role`, `is_headman`, …) / создание: `role` и профиль; с `password` — сразу активен, без — приглашение |
| GET | /users/:id | админ | карточка с документами |
| PATCH | /users/:id | админ | правка, `role` вместе с профилем, блокировка (`is_active`) |
| POST | /users/:id/invitations | админ | повторное приглашение (пароль ещё не задан) |
| GET, POST | /admin/administrators… | владелец | администраторы, приглашение, назначение |
| GET | /groups/:id/students | преподаватель, админ | студенты группы |

### Структура вуза
Справочники: чтение — все, запись — админ. Архива нет: DELETE удаляет только запись без ссылок (409 `catalog_in_use`).

| Путь | Методы |
|---|---|
| /departments, /catalog/subjects | GET, POST; PUT и DELETE на /:id |
| /specializations | GET, POST; PUT и DELETE на /:id |
| /groups | GET, POST; GET, PUT, DELETE на /:id |
| /semesters | GET, POST; PUT, DELETE на /:id; GET /semesters/current; POST /semesters/:id/current (сделать текущим) |
| /rooms (`number`) | GET, POST; PUT и DELETE на /:id |

### Курсы
| Метод | Путь | Кто | Описание |
|---|---|---|---|
| GET | /courses | все | список по роли; фильтры `group_id`, `semester_id`, `current=true` |
| POST | /courses | админ | курс = предмет + группа + семестр |
| GET / PATCH / DELETE | /courses/:id | все / админ / админ | страница предмета / замена дисциплины / удаление пустого |
| GET | /courses/:id/students | все с доступом | студенты группы курса |
| GET / POST | /courses/:id/workloads | все с доступом / админ | нагрузка: преподаватель, вид занятий, `planned_hours` (null — не задано) |
| PUT / DELETE | /workloads/:id | админ | |

### Расписание
Двухнедельный шаблон: `week_number` 1 — верхняя, 2 — нижняя. Неделя определяется от `WEEK_ONE_DATE` (или по чётности ISO-недели).

| Метод | Путь | Кто | Описание |
|---|---|---|---|
| GET | /schedule/week | публично | номер и границы недели на дату (`date`) |
| GET | /schedule/groups, /schedule/teachers | публично | списки для выбора |
| GET | /schedule/groups/:id, /schedule/teachers/:id | публично | расписание; `week`+`day` или `from`+`to` (до 62 дней, с датами); `semester_id` |
| GET | /schedule/me | все | своё расписание |
| POST / PUT / DELETE | /schedule/lessons[/:id] | админ | занятия (уникальность по нагрузке, неделе, дню, времени) |

### Темы и материалы
| Метод | Путь | Кто |
|---|---|---|
| GET | /courses/:id/topics | все с доступом |
| POST | /courses/:id/topics | преподаватель, админ |
| PUT, DELETE | /topics/:id | преподаватель, админ |
| POST | /topics/:id/materials | преподаватель курса (`file_id` и/или `external_url`; порядок по id) |
| PUT, DELETE | /materials/:id | преподаватель, админ |

### Задания, ответы, оценки
| Метод | Путь | Кто | Описание |
|---|---|---|---|
| GET | /courses/:id/assignments | все с доступом | студент видит `mine`, преподаватель — `stats` |
| POST | /courses/:id/assignments | преподаватель курса | тема, название, описание, дедлайн, до 5 файлов |
| GET / PUT / DELETE | /assignments/:id | все / преподаватель, админ | |
| GET | /assignments/:id/submissions | преподаватель, админ | студент → статус; фильтры `status`, `late`, `has_work` |
| GET | /assignments/:id/submissions/archive | преподаватель, админ | ZIP со всеми работами |
| GET / PUT | /assignments/:id/submission | студент | свой ответ: черновик (`submit=false`) или отправка |
| GET | /submissions/:id | автор, преподаватель, админ | |
| GET | /submissions/:id/archive | автор, преподаватель, админ | ZIP одной работы |
| PUT | /submissions/:id/review | преподаватель курса | оценка 0.0–5.0 (шаг 0.1) и комментарий |
| GET | /grades | студент | свои оценки (`semester_id`, `course_id`, `status`, `q`) |
| GET | /courses/:id/grades | преподаватель, админ | оценки курса |

Статусы для клиента: `not_submitted`, `missing` (дедлайн прошёл, ответа нет), `submitted`, `reviewed`; `is_late` — отправлено после дедлайна.

### Файлы
| Метод | Путь | Описание |
|---|---|---|
| POST | /files | multipart, поле `file`, `purpose=attachment|avatar`; проверка расширения и сигнатуры, лимит `MAX_UPLOAD_MB` |
| GET | /files/:id, /files/:id/meta | скачать / метаданные (по правилам доступа) |
| DELETE | /files/:id | только свой неиспользуемый файл |

### Чаты, главная, ресурсы, WebSocket
| Метод | Путь | Описание |
|---|---|---|
| GET | /contacts | кому можно написать (`q`, `limit`, `offset`) |
| GET | /chats | свои чаты, страница (`q`, `limit`, `offset`) |
| POST | /chats/direct, /chats/group | открыть личный / групповой чат (200) |
| GET | /chats/:id/participants | участники доступного чата: `[{id, name}]` |
| GET | /chats/:id, /chats/:id/messages | чат; история `before`, `limit` → `{items, has_more, next_before}` |
| GET | /dashboard | неделя, пары на сегодня, последние чаты и новости |
| GET | /resources | внешние сервисы и контакты из `RESOURCES_FILE` (по умолчанию — встроенный набор) |
| POST | /realtime/tickets | одноразовый билет на WebSocket (30 с, привязан к сессии и Origin) |
| GET | /realtime?ticket= | WebSocket, подпротокол `lk-chat.v1`; первое событие `session.ready` |

WebSocket `lk-chat.v1`: команды `message.send`, `message.edit`, `message.delete` (`{type, request_id, data}`) →
`ack` с тем же `request_id` или `error {code, message, details}`; после COMMIT всем подключённым участникам
чата — `message.created` / `message.updated` / `message.deleted` с `occurred_at`. Ошибка команды соединение
не закрывает; 4401 — истёк допуск (`expires_at` из `session.ready`) или сессия отозвана. Отправки сообщений через REST нет.

nginx для `/api/v1/realtime` должен пропускать Upgrade и держать долгое соединение:
`proxy_http_version 1.1; proxy_set_header Upgrade $http_upgrade; proxy_set_header Connection "upgrade";
proxy_read_timeout 1h;` и не писать query (`?ticket=`) в access-лог.

## Ключевые правила

- Проверенный ответ изменить нельзя (`409 already_reviewed`); оценить можно только отправленный (`409 not_submitted`).
- Опоздание разрешено: ответ помечается `is_late`.
- Файлы, оставшиеся без ссылок после правки или удаления, стираются из БД и с диска.
- Одна текущая сессия семестра: частичный уникальный индекс `semester_single_current_idx`.
- Пароли — bcrypt; email хранится в нижнем регистре.

## Что проверено и что нет

Проверено: `gofmt`, `go vet`, юнит-тесты (календарь недель, хранилище, ключи файлов, оценки, ZIP, фильтры),
миграция и основные SQL-запросы на реальном PostgreSQL 16 (список заданий, upsert ответа, запрет правки после проверки,
производные статусы, FK темы и курса, диапазон оценки).

Дополнительно проверена настоящая Docker-сборка API и `createadmin` с зависимостями из `go.sum`,
запуск общего Compose с PostgreSQL 17 и применение начальной миграции к новой БД.
Через прокси фронтенда проверены расписание, вход администратора, профиль, запрет доступа без JWT,
загрузка и скачивание файла после перезапуска API. Vue также открывается в браузере из контейнера.

## Не реализовано

Непрочитанные, replay событий после обрыва, идемпотентная отправка (не входят в контракт), уведомления,
сброс пароля по почте, задания для нескольких групп сразу.
