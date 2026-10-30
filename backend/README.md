# university-api

REST API личного кабинета вуза: **Go + Gin + pgx + PostgreSQL**, модульный монолит.
Схема БД — упрощённая `university_cabinet` (см. `migrations/001_init.sql`, 19 таблиц).

## Быстрый старт

```bash
cd ..                               # команды из корня репозитория
export JWT_SECRET=$(openssl rand -hex 32)
docker compose up -d --build --wait  # фронтенд, API, PostgreSQL и Adminer

# первый администратор (через API создать нельзя: POST /users требует админа)
docker compose exec -e ADMIN_EMAIL=admin@example.com \
  -e ADMIN_PASSWORD='сменить-меня-123' \
  backend createadmin

curl -s localhost:5173/api/v1/auth/login -H 'Content-Type: application/json' \
  -d '{"email":"admin@example.com","password":"сменить-меня-123"}'
# -> {"data":{"access_token":"...","token_type":"Bearer","expires_at":"...","role":"admin"}}
# дальше: Authorization: Bearer <token>
```

Браузер и HTTP-клиенты обращаются к API через фронтенд: `/api/v1/...`.
Vite проксирует запросы в контейнер `backend:8080`. SQL применяется автоматически
только к пустому тому PostgreSQL; для существующей БД см. корневой README.
Загруженные файлы сохраняются в отдельном томе `uploads`.

Проверка работоспособности без авторизации: `GET /api/healthz` возвращает
`{"status":"ok"}` через Vite и продакшен-nginx. Внутренний `GET /healthz`
используется healthcheck контейнера и возвращает тот же ответ.

Без Docker: `cp .env.example .env`, задать переменные, `make tidy` (создаст `go.sum`), `make migrate`, `make admin`, `make run`.

## Структура

```
cmd/api            точка входа
cmd/createadmin    создание первого администратора
internal/app       сборка зависимостей и роутер (единственное место, где модули знают друг о друге)
internal/platform  apperr, httpx, middleware (JWT, роли), database (DB/InTx/Collect, ошибки PG, Filter),
                   access (права на курс и файл), filekey, filegc (удаление неиспользуемых файлов), validate
internal/modules   auth, users, structure, courses, schedule, content, assignments, files
migrations         001_init.sql
```

Модуль = `module.go` (маршруты) → `handler` → `service` (правила и права) → `repository` (SQL).
Модули не импортируют друг друга: нужные им проверки (`Guard`, `Blobs`, `FileOwnership`) описаны интерфейсами в самом модуле,
а реализации подставляет `internal/app`. У справочников (`structure`) сервисного слоя нет.

## Роли и права

Роль берётся из токена: `admin` (`app_user.is_admin`), `teacher` (есть запись в `teacher`), иначе `student`.

- студент видит курсы своей группы; преподаватель — курсы, где у него есть нагрузка; админ — всё;
- студент видит только свои ответы и оценки;
- `created_by` и `reviewed_by` в БД ссылаются на `teacher`, поэтому **админ** управляет справочниками, курсами, нагрузкой,
  расписанием, пользователями и может править/удалять задания и материалы, но **не** создаёт их и не ставит оценки — это делает преподаватель курса;
- к материалу, заданию или ответу можно прикрепить только свой файл (папка `<user_id>/` в ключе хранения).

## Формат ответов

Успех: `{"data": ...}`; ошибка: `{"error": {"code": "...", "message": "..."}}`; удаление — `204` без тела.

## Эндпоинты

Префикс `/api/v1`. «Все» = любой авторизованный пользователь. Публичные — без токена.

### Аутентификация и профиль
| Метод | Путь | Кто | Описание |
|---|---|---|---|
| POST | /auth/login | публично | вход по email и паролю |
| GET / PATCH | /me | все | профиль / имя, отчество, аватар (`avatar_file_id`) |
| POST | /me/password | все | смена пароля (`old_password`, `new_password`) |
| POST | /me/email | все | смена email (нужен текущий пароль) |
| GET / POST | /users | админ | список (фильтры) / создание пользователя (роль student, teacher, admin) |
| GET | /users/:id | преподаватель, админ | карточка |
| PATCH | /users/:id | админ | правка, блокировка (`is_active`), сброс пароля (`password`), группа, кафедра |
| GET | /groups/:id/students | преподаватель, админ | студенты группы |

### Структура вуза
Справочники: чтение — все, запись — админ.

| Путь | Методы |
|---|---|
| /departments, /subjects | GET, POST; PUT и DELETE на /:id |
| /specializations | GET, POST; PUT и DELETE на /:id |
| /groups | GET, POST; GET, PUT, DELETE на /:id |
| /semesters | GET, POST; PUT, DELETE на /:id; GET /semesters/current; POST /semesters/:id/current (сделать текущим) |
| /rooms | GET, POST; PUT и DELETE на /:id |

### Курсы
| Метод | Путь | Кто | Описание |
|---|---|---|---|
| GET | /courses | все | список по роли; фильтры `group_id`, `semester_id`, `current=true` |
| POST | /courses | админ | курс = предмет + группа + семестр |
| GET / DELETE | /courses/:id | все / админ | |
| GET | /courses/:id/students | все с доступом | студенты группы курса |
| POST | /courses/:id/workloads | админ | нагрузка: преподаватель + вид занятий |
| DELETE | /workloads/:id | админ | |

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
| POST | /topics/:id/materials | преподаватель курса (файл и/или ссылка) |
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
| GET | /me/grades | студент | свои оценки (`course_id`) |
| GET | /courses/:id/grades | преподаватель, админ | оценки курса |

Статусы для клиента: `not_submitted`, `missing` (дедлайн прошёл, ответа нет), `submitted`, `reviewed`; `is_late` — отправлено после дедлайна.

### Файлы
| Метод | Путь | Описание |
|---|---|---|
| POST | /files | multipart, поле `file`, `purpose=attachment|avatar`; проверка расширения и сигнатуры, лимит `MAX_UPLOAD_MB` |
| GET | /files/:id, /files/:id/meta | скачать / метаданные (по правилам доступа) |
| DELETE | /files/:id | только свой неиспользуемый файл |

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

Чаты, новости, уведомления, сброс пароля и подтверждение email по почте, ведомость, задания для нескольких групп сразу, CORS, refresh-токены.
