# Frontend ННГАСУ

## Регистрация и вход

Формы `/registration` и `/authorization` отправляют запросы в `/api/v1/auth`.
Регистрация создаёт заявку со статусом `pending`; войти можно после одобрения
администратором на сервере. Административная панель пока работает с локальными
данными: её действия с заявками не заменяют серверное одобрение.

После входа профиль и access-токен хранятся в памяти. При перезагрузке сессия
восстанавливается через `/auth/csrf` и `/auth/refresh` с HttpOnly-cookie.
«Запомнить меня» задаёт серверный срок сессии: 30 суток вместо одних.
Выход в настройках и в карточке администратора отзывает сессию на сервере.
Административные маршруты требуют `user.is_admin` из ответа API.

В разработке Vite проксирует `/api` на `http://localhost:8080`; адрес можно
изменить через `API_PROXY_TARGET`. Прокси сохраняет исходный Host для проверки
Origin при обновлении сессии и выходе. Для HTTP задайте бэкенду
`COOKIE_SECURE=false` (в dev Compose уже задано); в production используйте HTTPS.
`VITE_TERMS_VERSION` из `.env.example` должна совпадать с `TERMS_VERSION` бэкенда;
по умолчанию обе версии — `2026-11-13`.

## Документация API

Страница `/api-docs` доступна без входа, маршрут `api-docs`. Она содержит целевой
контракт API: методы и права доступа, path/query/header/body, схемы запросов и ответов,
примеры, ошибки, бизнес-правила, WebSocket-протокол сообщений и порядок внедрения.
Страница не выполняет API-запросы.

Источник документации — `src/pages/apiDocs/api-docs.endpoints.ts`,
`api-docs.schemas.ts`, `api-docs.websocket.ts` и `api-docs.overview.ts`. Это проект контракта, а не описание
полностью реализованного сервера. Контракт основан на сценариях приложения и не
требует совместимости с текущим бэкендом. Основной сценарий чатов использует WebSocket;
HTTP отвечает за исходные данные и историю. Расширения интерфейса отмечены отдельно. Обработчики бэкенда в рамках документации не добавлялись.

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
