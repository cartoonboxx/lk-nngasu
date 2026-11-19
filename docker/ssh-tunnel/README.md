# Серверная база через SSH

Локальный frontend → backend → ssh-tunnel → SSH сервера → 127.0.0.1:15432 → PostgreSQL.
Нужен Docker Compose 2.24.4+ (поддержка `!override`).

## Однократная настройка

1. Проверьте вход: `ssh -i ~/.ssh/lk-nngasu-db root@77.83.244.134`.
   Для другого SSH-порта добавьте `-p НОМЕР`. Если ключа нет, создайте его
   командой `ssh-keygen -t ed25519 -f ~/.ssh/lk-nngasu-db -N ''`, затем установите
   публичный ключ через `ssh-copy-id -i ~/.ssh/lk-nngasu-db.pub root@77.83.244.134`.
   При первом входе проверьте отпечаток сервера через доверенный канал и сохраните
   его в `known_hosts`. Туннель требует ключ без парольной фразы.
2. В серверном `compose.prod.yaml` уже предусмотрена публикация PostgreSQL
   на `127.0.0.1:15432`. Примените актуальный файл на сервере из каталога проекта:

   ```sh
   docker compose --env-file .env --env-file .release.env -f compose.prod.yaml up -d db
   ```

   БД кратковременно пересоздаётся с сохранением тома. Открывать порт 15432
   в интернет не нужно. SSH-сервер должен разрешать TCP forwarding
   к `127.0.0.1:15432`.
3. В корневом `.env` задайте параметры подключения, сохранив существующий
   `JWT_SECRET`:

   | Переменная | Значение |
   | --- | --- |
   | `COMPOSE_FILE` | `compose.yaml:compose.ssh.yaml` для macOS/Linux |
   | `SSH_HOST` | Адрес сервера |
   | `SSH_USER` | Пользователь SSH |
   | `SSH_PORT` | Порт SSH, по умолчанию `22` |
   | `SSH_DB_PORT` | Порт базы на loopback сервера, по умолчанию `15432` |
   | `SSH_KEY_PATH` | Абсолютный путь к приватному ключу на компьютере |
   | `SSH_KNOWN_HOSTS_PATH` | Абсолютный путь к файлу `~/.ssh/known_hosts` |
   | `REMOTE_DB_USER` | Пользователь PostgreSQL, по умолчанию `nngasu` |
   | `REMOTE_DB_NAME` | Имя базы, по умолчанию `nngasu` |
   | `REMOTE_DATABASE_PASSWORD` | Пароль PostgreSQL серверного контейнера |

   Пароль REMOTE_DATABASE_PASSWORD — значение POSTGRES_PASSWORD из серверного
   контейнера, не пароль SSH. Пароли с `$` или `#` заключайте в одинарные
   кавычки. Приватный ключ и .env не коммитьте.

## Запуск

Запустите Docker Desktop, затем:

```sh
docker compose up -d --build --wait
```

После первой сборки достаточно `docker compose up -d`. COMPOSE_FILE в .env
автоматически подключает compose.ssh.yaml. Локальные db и adminer в этом режиме
не запускаются. Порт туннеля доступен только в сети Compose. SSH шифрует
соединение с сервером; отдельный терминал не требуется. При разрыве связи
SSH завершится и Docker перезапустит контейнер.

```sh
curl -fsS http://localhost:5173/api/v1/schedule/groups
docker compose logs --tail=100 ssh-tunnel backend
```

Healthcheck туннеля проверяет ответ PostgreSQL; пароль и совместимость схемы
проверяет бэкенд при подключении. Если SSH зависает с VPN, проверяйте без VPN.

Все изменения через приложение записываются в серверную БД. Миграции локального
API отключены, поэтому серверная схема должна быть совместима с локальным кодом.
Серверные файлы не копируются в локальный том uploads. Почтовые функции используют
локальные настройки SMTP и APP_URL.

Чтобы вернуться к локальной БД, сначала выполните `docker compose down` без `-v`,
затем удалите COMPOSE_FILE из .env и запустите `docker compose up -d --build --wait`.
Тома не удаляются. REMOTE_DATABASE_URL из старой инструкции в SSH-режиме не используется.
