#!/usr/bin/env bash
set -euo pipefail

network="oficina-migrations-test-$$"
database="$network-db"
image="$network:local"
test_files=$(mktemp -d)
# Git Bash must not rewrite the container paths in Docker bind mounts.
if command -v cygpath >/dev/null 2>&1; then
  test_files=$(cygpath -m "$test_files")
  export MSYS_NO_PATHCONV=1
fi
cleanup() {
  docker rm -f "$database" >/dev/null 2>&1 || true
  docker network rm "$network" >/dev/null 2>&1 || true
  docker image rm "$image" >/dev/null 2>&1 || true
  rm -rf "$test_files"
}
trap cleanup EXIT

docker build -f Dockerfile.migrations -t "$image" .
docker network create "$network" >/dev/null
docker run -d --name "$database" --network "$network" --network-alias postgres \
  -e POSTGRES_DB=oficina -e POSTGRES_USER=oficina -e POSTGRES_PASSWORD=test-only \
  postgres:16-alpine >/dev/null

migrate() {
  docker run --rm --network "$network" \
    -e FLYWAY_URL=jdbc:postgresql://postgres:5432/oficina \
    -e FLYWAY_USER=oficina -e FLYWAY_PASSWORD=test-only \
    "$image" "$@"
}
query() {
  docker exec "$database" psql -U oficina -d oficina -At -v ON_ERROR_STOP=1 -c "$1"
}

migrate migrate
expected=$(find db/migrations -maxdepth 1 -name '*.sql' | wc -l | tr -d ' ')
actual=$(query 'SELECT count(*) FROM flyway_schema_history WHERE success AND version IS NOT NULL')
[[ "$actual" == "$expected" ]]
query 'SELECT version FROM mecanico LIMIT 0'
query 'SELECT conteudo_html FROM notificacao LIMIT 0'
migrate migrate
[[ "$(query 'SELECT count(*) FROM flyway_schema_history WHERE success AND version IS NOT NULL')" == "$expected" ]]

# A changed SQL file must fail validation instead of being executed again.
cp db/migrations/V001__schema_inicial.sql "$test_files/changed.sql"
printf '\n-- Checksum regression test\n' >> "$test_files/changed.sql"
if docker run --rm --network "$network" \
  -v "$test_files/changed.sql:/flyway/migrations/V001__schema_inicial.sql:ro" \
  -e FLYWAY_URL=jdbc:postgresql://postgres:5432/oficina \
  -e FLYWAY_USER=oficina -e FLYWAY_PASSWORD=test-only \
  "$image" migrate > "$test_files/checksum.log" 2>&1; then
  echo 'ERROR: changed migration was accepted' >&2
  exit 1
fi
cat "$test_files/checksum.log"
grep -qi 'checksum mismatch' "$test_files/checksum.log"

# A failed migration must roll back its changes and keep the successful history intact.
printf 'CREATE TABLE rollback_probe (id integer); SELECT * FROM missing_table_for_test;\n' \
  > "$test_files/V999__failure.sql"
if docker run --rm --network "$network" \
  -v "$test_files/V999__failure.sql:/flyway/migrations/V999__failure.sql:ro" \
  -e FLYWAY_URL=jdbc:postgresql://postgres:5432/oficina \
  -e FLYWAY_USER=oficina -e FLYWAY_PASSWORD=test-only \
  "$image" migrate > "$test_files/failure.log" 2>&1; then
  echo 'ERROR: invalid migration was accepted' >&2
  exit 1
fi
cat "$test_files/failure.log"
grep -q 'missing_table_for_test' "$test_files/failure.log"
[[ "$(query "SELECT to_regclass('public.rollback_probe') IS NULL")" == t ]]
[[ "$(query 'SELECT count(*) FROM flyway_schema_history WHERE success AND version IS NOT NULL')" == "$expected" ]]
migrate validate
echo "PASS: $expected migrations, repeat execution, checksum protection and rollback."
