# Личный кабинет ННГАСУ

## Локальная разработка

```sh
export JWT_SECRET="$(openssl rand -hex 32)"
docker compose up -d --build --wait
```

Фронтенд доступен на http://localhost:5173, Adminer — на http://localhost:8080.
Изменения исходников фронтенда подхватывает Vite. После изменения Go-кода
выполните `docker compose up -d --build backend`.

Сохраните `JWT_SECRET` (минимум 32 символа) в окружении или корневом `.env`,
чтобы он оставался одинаковым при перезапусках. Не добавляйте `.env` в Git.

Браузер обращается к `/api/v1/...` на том же адресе фронтенда.
Vite перенаправляет эти запросы в `http://backend:8080` внутри сети Compose;
отдельный публичный порт бэкенда и CORS для этой схемы не нужны.
Например:

```sh
curl -fsS http://localhost:5173/api/v1/schedule/week
```

Интерфейс пока использует демонстрационные данные: прокси обеспечивает доступ
к API, подключение экранов к API выполняется отдельно.

Compose ждёт готовности PostgreSQL и API перед запуском фронтенда.
SQL из `backend/migrations` применяется при создании пустого тома БД.
Для существующей БД без таблиц бэкенда примените начальную миграцию вручную:

```sh
docker compose exec -T db psql -U root -d nngasu -v ON_ERROR_STOP=1 --single-transaction < backend/migrations/001_init.sql
```

Повторно запускать начальную миграцию на уже созданной схеме не нужно.
Данные PostgreSQL и загруженные файлы сохраняются в томах `postgres-data`
и `uploads`. Обычный `docker compose down` их сохраняет.

Создание первого администратора:

```sh
docker compose exec -e ADMIN_EMAIL=admin@example.com \
  -e ADMIN_PASSWORD='замените-на-свой-пароль' backend createadmin
```

При запуске Vite вне Docker задайте `API_PROXY_TARGET` с адресом API
(по умолчанию `http://localhost:8080`). Порты Compose при необходимости
меняются через `FRONTEND_PORT` и `ADMINER_PORT`.

## Продакшен и CI/CD

`compose.prod.yaml` запускается самостоятельно и поднимает nginx с фронтендом,
Go API и PostgreSQL 17. База и API доступны только внутри сети Compose.
Внешний HTTPS nginx сервера обращается к `127.0.0.1:8081`.
PostgreSQL и загруженные файлы сохраняются в именованных томах проекта
`lk-nngasu-prod`, отдельно от локальной разработки.

### Подготовка сервера

Нужны Docker Engine с Docker Compose, `bash` и работающий HTTPS nginx.
Пользователь SSH должен иметь доступ к Docker и запись в `/var/www/lk-nngasu`.
Node.js, Go и сборка исходников на сервере не нужны: CI публикует готовые образы.

Перед первым деплоем создайте `/var/www/lk-nngasu/.env` с двумя секретами
(образец — `deploy/.env.example`). Например, в новом каталоге на сервере:

```sh
mkdir -p /var/www/lk-nngasu
cd /var/www/lk-nngasu
umask 077
printf 'POSTGRES_PASSWORD=%s\nJWT_SECRET=%s\n' \
  "$(openssl rand -hex 32)" "$(openssl rand -hex 32)" > .env
```

Эта команда предназначена только для первого запуска. Сохраняйте `.env`
между обновлениями: смена `POSTGRES_PASSWORD` в файле не меняет пароль
в существующей БД, а смена `JWT_SECRET` завершает выданные сессии.
CI не передаёт и не перезаписывает этот файл.

На новой БД SQL из `backend/migrations` выполняется автоматически.
На существующей схеме повторно запускать `001_init.sql` не нужно.
Будущие миграции существующей БД применяются отдельно перед обновлением API;
каталог `docker-entrypoint-initdb.d` обрабатывается PostgreSQL только при
инициализации пустого тома. Том нельзя удалять для применения миграций.

### GitHub Actions

Для репозитория задайте Actions secrets:

- `SSH_HOST` — адрес сервера;
- `SSH_USER` — пользователь с доступом к Docker;
- `SSH_PRIVATE_KEY` — приватный ключ этого пользователя;
- `SSH_KNOWN_HOSTS` — проверенная запись ключа сервера для `known_hosts`.

Workflow `.github/workflows/check.yml` выполняет:

1. Проверки фронтенда (lint и TypeScript), бэкенда (`go vet` и сборка),
   обеих конфигураций Compose и синтаксиса скрипта развёртывания.
2. Сборку обоих `Dockerfile.prod` для PR. В образе фронтенда также выполняется `nginx -t`.
3. На push в `master` — публикацию в GHCR двух образов с одним тегом Git SHA:
   `ghcr.io/<owner>/<repo>:<sha>` и `ghcr.io/<owner>/<repo>-backend:<sha>`.
   Образы собираются на Linux amd64; исходники и секреты приложения в них не передаются.
4. Передачу Compose, скрипта развёртывания, SQL и образца nginx на сервер.
   Для скачивания приватных образов workflow авторизует сервер в GHCR
   через временный `GITHUB_TOKEN` с `packages: read`.
5. Запуск `deploy/deploy.sh`: скачивание образов, запуск БД → API → nginx,
   ожидание healthcheck и запрос к API справочника групп через nginx.

Публикация использует `GITHUB_TOKEN` с `packages: write`. У GHCR-пакетов
должен быть разрешён доступ из этого репозитория (Manage Actions access).
Для первого push в уже существующий пакет без связи с репозиторием может
потребоваться вручную выдать такой доступ в настройках пакета.
Параллельные развёртывания в production запрещены через concurrency.

Скрипт сохраняет ссылки на успешно развёрнутые образы в `.release.env`.
Старый `build-script.sh` на сервере больше не используется.
При ошибке выводятся состояние контейнеров и логи; автоматического отката нет.
Во время пересоздания API возможен короткий перерыв в доступности.

### HTTPS nginx

Схема запросов:

- страницы и файлы: браузер → HTTPS nginx сервера → `127.0.0.1:8081` → nginx контейнера;
- `/api/v1/...`: тот же путь → nginx контейнера → `backend:8080` → PostgreSQL.

В существующем HTTPS `server` замените `location /` содержимым
`deploy/nginx-location.conf`, затем проверьте и примените конфигурацию:

```sh
sudo nginx -t
sudo systemctl reload nginx
```

Сохраняйте существующие `listen`, `server_name` и настройки сертификатов.
Уберите конфликтующие `location /api/` и старую раздачу `/assets/`, если они есть:
все запросы должны доходить до контейнера. Workflow копирует пример конфига,
но не изменяет системный nginx и сертификаты сервера автоматически.

Лимит тела запроса на обоих nginx — 16 МБ: 15 МБ на файл и запас для multipart.
Прокси сохраняет префикс `/api/v1`, параметры запроса и `Authorization`.
Контейнерный nginx передаёт исходный `X-Forwarded-Proto` от HTTPS nginx
и повторно разрешает имя `backend` через Docker DNS при смене IP контейнера.
Вложенные страницы Vue Router открываются напрямую; неизвестные пути API
не подменяются HTML приложения.

Проверки с сервера и через публичный адрес:

```sh
curl -fsS http://127.0.0.1:8081/healthz
curl -fsS http://127.0.0.1:8081/api/healthz
curl -fsS http://127.0.0.1:8081/api/v1/schedule/groups
curl -fsS https://example.org/dashboard
curl -fsS https://example.org/api/v1/schedule/groups
```

`/healthz` проверяет nginx, `/api/healthz` — доступность API.
Healthcheck фронтенда проверяет оба адреса. Скрипт деплоя дополнительно
запрашивает справочник групп, чтобы проверить доступ API к схеме БД.
Публичный HTTPS и сертификат нужно проверять отдельно после настройки
системного nginx, заменив `example.org` своим доменом.

### Ручной запуск и обслуживание

При первом ручном запуске из корня репозитория после подготовки `.env`
передайте ссылки на опубликованные образы одного коммита:

```sh
bash deploy/deploy.sh ghcr.io/<owner>/<repo>:<sha> ghcr.io/<owner>/<repo>-backend:<sha>
```

Для приватных образов сначала выполните `docker login ghcr.io` на сервере
с токеном, имеющим `read:packages`. Авторизация из CI действует ограниченное время.

После успешного деплоя для команд Compose подключайте оба файла окружения:

```sh
docker compose --env-file .env --env-file .release.env -f compose.prod.yaml ps
docker compose --env-file .env --env-file .release.env -f compose.prod.yaml logs --tail=100 -f frontend backend db
```

Создание первого администратора:

```sh
docker compose --env-file .env --env-file .release.env -f compose.prod.yaml exec \
  -e ADMIN_EMAIL=admin@example.com -e ADMIN_PASSWORD='замените-на-свой-пароль' backend createadmin
```

Остановка с сохранением данных:

```sh
docker compose --env-file .env --env-file .release.env -f compose.prod.yaml down
```

Не используйте `down --volumes` для продакшена: это удалит БД и загруженные файлы.
Перед изменением схемы и обновлением PostgreSQL делайте резервные копии обоих томов.
Обновление приложения использует тот же `deploy/deploy.sh` с новой парой образов.
