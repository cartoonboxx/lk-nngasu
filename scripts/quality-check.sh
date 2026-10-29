#!/bin/sh
set -eu

PROJECT_ROOT="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

npm --prefix "$PROJECT_ROOT/frontend" run lint:check
npm --prefix "$PROJECT_ROOT/frontend" run type-check

printf '\nAll quality checks passed.\n'
