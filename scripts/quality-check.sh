#!/bin/sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)"

run_step() {
  name="$1"
  shift

  printf '\n==> %s\n' "$name"
  "$@"
}

package_manager_cmd() {
  dir="$1"

  if [ -f "$dir/package-lock.json" ]; then
    printf 'npm'
  elif [ -f "$dir/pnpm-lock.yaml" ]; then
    printf 'pnpm'
  elif [ -f "$dir/yarn.lock" ]; then
    printf 'yarn'
  else
    printf 'npm'
  fi
}

has_npm_script() {
  dir="$1"
  script="$2"

  [ -f "$dir/package.json" ] || return 1

  node -e "
    const fs = require('fs');
    const pkg = JSON.parse(fs.readFileSync(process.argv[1], 'utf8'));
    process.exit(pkg.scripts && pkg.scripts[process.argv[2]] ? 0 : 1);
  " "$dir/package.json" "$script"
}

run_npm_script_if_present() {
  dir="$1"
  script="$2"
  name="$3"
  shift 3

  if ! has_npm_script "$dir" "$script"; then
    printf '\n-- Skipping %s: script "%s" not found in %s/package.json\n' "$name" "$script" "${dir#$ROOT_DIR/}"
    return 0
  fi

  pm="$(package_manager_cmd "$dir")"

  case "$pm" in
    npm)
      run_step "$name" npm --prefix "$dir" run "$script" -- "$@"
      ;;
    pnpm)
      run_step "$name" pnpm --dir "$dir" run "$script" -- "$@"
      ;;
    yarn)
      run_step "$name" sh -c 'cd "$1" && yarn "$2" "$@"' sh "$dir" "$script" "$@"
      ;;
  esac
}

run_node_project_checks() {
  dir="$1"
  label="$2"

  [ -f "$dir/package.json" ] || return 0

  if has_npm_script "$dir" lint:check; then
    run_npm_script_if_present "$dir" lint:check "$label lint"
  else
    run_npm_script_if_present "$dir" lint "$label lint"
  fi

  if has_npm_script "$dir" format:check; then
    run_npm_script_if_present "$dir" format:check "$label prettier"
  elif has_npm_script "$dir" prettier:check; then
    run_npm_script_if_present "$dir" prettier:check "$label prettier"
  else
    printf '\n-- Skipping %s prettier: script "format:check" or "prettier:check" not found in %s/package.json\n' "$label" "${dir#$ROOT_DIR/}"
  fi

  if has_npm_script "$dir" test:unit:run; then
    run_npm_script_if_present "$dir" test:unit:run "$label unit tests"
  elif has_npm_script "$dir" test:unit; then
    run_npm_script_if_present "$dir" test:unit "$label unit tests" --run
  elif has_npm_script "$dir" test; then
    run_npm_script_if_present "$dir" test "$label tests"
  else
    printf '\n-- Skipping %s tests: script "test:unit:run", "test:unit" or "test" not found in %s/package.json\n' "$label" "${dir#$ROOT_DIR/}"
  fi
}

run_backend_checks() {
  dir="$ROOT_DIR/backend"

  if [ ! -d "$dir" ]; then
    printf '\n-- Skipping backend checks: backend directory not found\n'
    return 0
  fi

  run_node_project_checks "$dir" backend

  if [ -f "$dir/pyproject.toml" ] || [ -f "$dir/pytest.ini" ]; then
    run_step "backend tests" sh -c 'cd "$1" && python -m pytest' sh "$dir"
  elif [ ! -f "$dir/package.json" ]; then
    printf '\n-- Skipping backend checks: no supported backend project files found\n'
  fi
}

run_frontend_checks() {
  run_node_project_checks "$ROOT_DIR/frontend" frontend
}

run_backend_checks
run_frontend_checks

printf '\nAll quality checks passed.\n'
