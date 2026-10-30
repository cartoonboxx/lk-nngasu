#!/usr/bin/env bash
set -Eeuo pipefail

cd "$(dirname "$0")/.."

export FRONTEND_IMAGE="${1:?Передайте образ фронтенда}"
export BACKEND_IMAGE="${2:?Передайте образ бэкенда}"

if [[ ! -f .env ]]; then
  echo 'Создайте серверный .env по образцу deploy/.env.example' >&2
  exit 1
fi

trap 'docker compose -f compose.prod.yaml ps >&2; docker compose -f compose.prod.yaml logs --tail=80 frontend backend db >&2' ERR

docker compose -f compose.prod.yaml config --quiet
docker compose -f compose.prod.yaml pull
docker compose -f compose.prod.yaml up -d --no-build --wait --wait-timeout 180
docker compose -f compose.prod.yaml exec -T frontend nginx -t
docker compose -f compose.prod.yaml exec -T frontend wget -q -O /dev/null http://127.0.0.1/api/v1/schedule/groups

# Save image references only after the deployment passes its health checks.
umask 077
printf 'FRONTEND_IMAGE=%s\nBACKEND_IMAGE=%s\n' "$FRONTEND_IMAGE" "$BACKEND_IMAGE" > .release.env.tmp
mv .release.env.tmp .release.env

docker compose -f compose.prod.yaml ps
