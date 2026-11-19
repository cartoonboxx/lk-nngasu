# Frontend ННГАСУ

## Регистрация и вход

Формы `/registration` и `/authorization` отправляют запросы в `/api/v1/auth`.
Регистрация создаёт заявку со статусом `pending`; войти можно после одобрения
администратором на сервере. Административная панель получает заявки из API;
одобрение создаёт реальный аккаунт, доступный для входа с паролем из заявки.
Демонстрационные данные и локальные снимки панели удалены. Пользователи,
справочники, курсы и расписание используют существующие API; управление новостями
временно недоступно до появления серверных методов записи.

После входа профиль и access-токен хранятся в памяти. При перезагрузке сессия
восстанавливается через `/auth/csrf` и `/auth/refresh` с HttpOnly-cookie.
«Запомнить меня» задаёт серверный срок сессии: 30 суток вместо одних.
После перезагрузки защищённый маршрут ждёт восстановления через refresh-cookie.
Обновление cookie, вход и выход выполняются под общей Web Locks-блокировкой
между вкладками (HTTPS/localhost), чтобы параллельные обновления не отзывали
сессию как повторное использование refresh-токена. Токены не сохраняются
в localStorage. На браузерах без Web Locks сохраняется блокировка внутри вкладки.
Выход в настройках и в карточке администратора отзывает сессию на сервере.
Административные маршруты требуют `user.is_admin` из ответа API.
`/setup` создаёт первого владельца по серверному ключу `ADMIN_SETUP_KEY`.
Раздел `/admin/administrators` работает с реальным API и доступен владельцу
с `can_manage_admins=true`: приглашения новых и назначение существующих пользователей.
Почтовая ссылка `/activation#token=…` позволяет получателю самостоятельно задать пароль.
[Инструкция настройки](../backend/docs/admin-access.md).

В разработке Vite проксирует `/api` на `http://localhost:8080`; адрес можно
изменить через `API_PROXY_TARGET`. Прокси сохраняет исходный Host для проверки
Origin при обновлении сессии и выходе. Для HTTP задайте бэкенду
`COOKIE_SECURE=false` (в dev Compose уже задано); в production используйте HTTPS
и `COOKIE_SECURE=true`. В `compose.prod.yaml` настройка передаётся из серверного `.env`:
при временной работе сервера по HTTP укажите `COOKIE_SECURE=false` и пересоздайте
контейнер бэкенда. С Secure-cookie на обычном HTTP браузер не сохранит сессию.
`VITE_TERMS_VERSION` должна совпадать с `TERMS_VERSION` бэкенда;
по умолчанию обе версии — `2026-11-13`.

## Личный кабинет и API

Демо-данные и `*.data.ts` удалены. Pinia store используют Options API; запросы
выполняются в actions через `auth.request()` с Bearer-токеном и обновлением сессии.

- `auth.getMe()` — профиль `/me`; `updateMe()` — собственный аватар; `changePassword()` — `/me/password`.
  ФИО пользователей меняет только администратор через админ-панель.
  После смены пароля нужен повторный вход. Аватар загружается с `purpose=avatar`.
- `subjects.fetchSubjects()` — все доступные курсы `/courses` с пагинацией;
  `fetchSubject(id)` — курс, темы, материалы и задания `/courses/:id`.
  Ответ загружается через `/assignments/:id`, отправляется в `/assignments/:id/submission`.
- `schedule.fetchSchedule()` — календарь `/schedule` по датам; список групп берётся
  из `/schedule/groups`. Главная использует агрегат `/dashboard` с серверной датой.
- `cabinet.fetchGrades()` — строки `/grades?semester_id=…` для доступных семестров,
  включая просроченные и несданные задания, оценку и комментарий. Полный ответ
  и файлы загружаются при открытии задания.
- `files.upload()` / `download()` — multipart `/files` и авторизованное скачивание Blob.
  Для работ и сообщений интерфейс использует лимит Compose: 5 файлов по 15 МиБ;
  для аватара — 2 МиБ. Изменение аватара: `PATCH /me` с полем `avatar`.
- `messages` — список `/chats`, контакты `/contacts`, создание личного или группового
  чата и курсорная история `{items, has_more, next_before}`. Отправка, правка и удаление
  идут через WebSocket `lk-chat.v1`, после одноразового билета `/realtime/tickets`.
  Подтверждение `ack` означает сохранение на сервере. При обрыве команды не
  повторяются автоматически; черновик сохраняется, история синхронизируется.
  После переподключения перечитывается весь ранее загруженный диапазон, включая
  правки и удаления. Непрочитанные и отметки чтения текущим API не поддерживаются.
- Новости: публичная лента `/news`, отдельная карточка `/news/:id`, текст `body`
  и серверный `excerpt`. Редактор `/admin/news` поддерживает черновики, публикацию,
  архив, восстановление и удаление. На странице ресурсов сохранены карточки сервисов и контактов университета.
- Учебная структура поддерживает удаление записей без зависимостей и назначение
  текущего семестра. Аудитории используют серверное поле `number`, архивов нет.
- В разделе курсов администратор управляет темами, материалами, заданиями и
  вложениями, просматривает ответы студентов и скачивает ZIP работ. Оценивать
  ответы может только преподаватель курса. Студент сохраняет черновики, отправляет
  и удаляет свои непроверенные ответы; проверенные ответы доступны для чтения.

WebSocket проксируется Vite (`ws: true`) и nginx контейнера. На сервере HTTPS nginx
также должен использовать блок `/api/v1/realtime` из `deploy/nginx-location.conf`;
после изменения системного конфига нужны `nginx -t` и reload. Одноразовые билеты
не записываются в access-лог этого маршрута.

Личные данные и переписки в localStorage/sessionStorage не записываются.
LocalStorage используется только для темы оформления. Черновики сообщений остаются
в памяти до перезагрузки; отправленные сообщения и работы сохраняются на сервере.
При выходе и смене пользователя содержимое store очищается, запоздавшие ответы
предыдущего аккаунта не применяются. WebSocket закрывается при завершении сессии.

## Уведомления

`NotifyBadge` подключён один раз в `App.vue` и показывает только текст сверху,
поверх страниц и модальных окон. Красный фон означает ошибку, зелёный — успех,
серо-синий — нейтральное сообщение. Одновременно отображается одна плашка:
новая заменяет предыдущую и скрывается через 3,5 секунды. Кнопок, иконок,
счётчиков и очереди нет; наведение не останавливает таймер. Появление и исчезновение
плавные; настройка уменьшения анимации в системе учитывается.

Вызов из компонента Options API:

```ts
this.$utilityStore.notify('Изменения сохранены.', 'success')
this.$utilityStore.notify('Данные обновляются.') // neutral по умолчанию
this.$utilityStore.notify('Не удалось загрузить файл.', 'error')
```

Из store или другого TS-модуля: `useUtilityStore().notifyError(error)`.
Ошибки `apiRequest()` и `auth.request()` автоматически показывают текст ошибки
красной плашкой. Отсутствие текущего семестра (`current_semester_not_set`) не
показывает ошибку: расписание остаётся пустым. Успешное обновление токена, отсутствие гостевой сессии и
намеренная отмена запроса уведомлений не создают. Состояние хранится в памяти.

## Документация API

Страница `/api-docs` доступна без входа, маршрут `api-docs`. Она содержит целевой
контракт API: методы и права доступа, path/query/header/body, схемы запросов и ответов,
примеры, ошибки, бизнес-правила, WebSocket-протокол сообщений и порядок внедрения.
Страница не выполняет API-запросы.

Источник документации — `src/pages/apiDocs/api-docs.endpoints.ts`,
`api-docs.schemas.ts`, `api-docs.websocket.ts` и `api-docs.overview.ts`. Это проект контракта, а не описание
полностью реализованного сервера. Контракт основан на сценариях приложения и не
требует совместимости с текущим бэкендом. Целевой сценарий чатов описывает WebSocket;
текущий интерфейс использует реализованный на сервере HTTP API. Расширения интерфейса отмечены отдельно. Обработчики бэкенда в рамках документации не добавлялись.

## Запуск через Docker для разработки

Запустите Docker Desktop. В папке `frontend` выполните:

```sh
docker compose up
```

При первом запуске Compose соберёт образ и установит зависимости.
Откройте http://localhost:1488. Изменения в `src` и `public` доступны
контейнеру сразу; Vite обновляет страницу во время разработки.

Для остановки нажмите Ctrl+C. Чтобы удалить остановленный контейнер и сеть:

```sh
docker compose down
```

После изменения зависимостей (`package.json` или `package-lock.json`),
Dockerfile или других файлов, которые не подключены через `volumes`,
пересоберите образ:

```sh
docker compose up --build
```

This template should help get you started developing with Vue 3 in Vite.

## Recommended IDE Setup

[VS Code](https://code.visualstudio.com/) + [Vue (Official)](https://marketplace.visualstudio.com/items?itemName=Vue.volar) (and disable Vetur).

## Recommended Browser Setup

- Chromium-based browsers (Chrome, Edge, Brave, etc.):
  - [Vue.js devtools](https://chromewebstore.google.com/detail/vuejs-devtools/nhdogjmejiglipccpnnnanhbledajbpd)
  - [Turn on Custom Object Formatter in Chrome DevTools](http://bit.ly/object-formatters)
- Firefox:
  - [Vue.js devtools](https://addons.mozilla.org/en-US/firefox/addon/vue-js-devtools/)
  - [Turn on Custom Object Formatter in Firefox DevTools](https://fxdx.dev/firefox-devtools-custom-object-formatters/)

## Type Support for `.vue` Imports in TS

TypeScript cannot handle type information for `.vue` imports by default, so we replace the `tsc` CLI with `vue-tsc` for type checking. In editors, we need [Volar](https://marketplace.visualstudio.com/items?itemName=Vue.volar) to make the TypeScript language service aware of `.vue` types.

## Customize configuration

See [Vite Configuration Reference](https://vite.dev/config/).

## Project Setup

```sh
npm install
```

### Compile and Hot-Reload for Development

```sh
npm run dev
```

### Type-Check, Compile and Minify for Production

```sh
npm run build
```

### Lint with [ESLint](https://eslint.org/)

```sh
npm run lint
```
