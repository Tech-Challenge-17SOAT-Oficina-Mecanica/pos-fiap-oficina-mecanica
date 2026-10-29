#!/usr/bin/env bash
# Testes isolados do bootstrap: nao acessam AWS nem executam Terraform real.
set -euo pipefail
test_dir=$(mktemp -d)
test_dir=$(cd "$test_dir" && pwd -P)
test_temp_root=$(cd "${TMPDIR:-/tmp}" && pwd -P)
[[ "$test_dir" == "$test_temp_root/"* ]] || exit 1
trap 'rm -rf "$test_dir"' EXIT
mkdir "$test_dir/bin"
export BOOTSTRAP_TEST_LOG="$test_dir/commands.log"
cat > "$test_dir/bin/aws" <<'EOF'
#!/usr/bin/env bash
echo "aws $*" >> "$BOOTSTRAP_TEST_LOG"
if [[ "$1 $2" == 'sts get-caller-identity' ]]; then echo "$FAKE_ACCOUNT"; exit 0; fi
if [[ "$1 $2" == 's3api head-bucket' ]]; then
  if [[ "$BUCKET_MODE" == missing ]]; then echo 'An error occurred (404)' >&2; exit 1; fi
  if [[ "$BUCKET_MODE" == denied ]]; then echo 'An error occurred (403)' >&2; exit 1; fi
fi
EOF
cat > "$test_dir/bin/terraform" <<'EOF'
#!/usr/bin/env bash
echo "terraform $*" >> "$BOOTSTRAP_TEST_LOG"
if [[ "$2 $3" == 'state list' ]]; then
  if [[ "$STATE_MODE" == missing ]]; then echo 'No state file was found!' >&2; exit 1; fi
  if [[ "$STATE_MODE" == broken ]]; then echo 'Unable to read state' >&2; exit 1; fi
  for type in aws_s3_bucket aws_s3_bucket_public_access_block aws_s3_bucket_versioning aws_s3_bucket_server_side_encryption_configuration aws_s3_bucket_ownership_controls; do
    echo "$type.terraform"
  done
fi
if [[ "$2" == plan && "$PLAN_FAIL" == 1 ]]; then exit 1; fi
exit 0
EOF
chmod +x "$test_dir/bin/"*
export PATH="$test_dir/bin:$PATH"
export EXPECTED_ACCOUNT_ID=123456789012 FAKE_ACCOUNT=123456789012
export BUCKET_MODE=missing STATE_MODE=missing PLAN_FAIL=0
# Nao alterar os outputs do job de CI enquanto os testes simulam outras contas.
export GITHUB_OUTPUT="$test_dir/outputs"

bash scripts/prepare-aws-state.sh > /dev/null
grep -q 's3api create-bucket' "$BOOTSTRAP_TEST_LOG"
[[ $(grep -c ' import ' "$BOOTSTRAP_TEST_LOG") -eq 5 ]]
grep -q 'use_lockfile=true' "$BOOTSTRAP_TEST_LOG"
echo 'PASS: bucket novo protegido e cinco recursos importados no estado remoto.'

export BUCKET_MODE=exists STATE_MODE=existing
: > "$BOOTSTRAP_TEST_LOG"
bash scripts/prepare-aws-state.sh > /dev/null
! grep -Eq 'create-bucket| import ' "$BOOTSTRAP_TEST_LOG"
echo 'PASS: reexecucao reconhece bucket e estado existentes.'

export FAKE_ACCOUNT=999999999999
: > "$BOOTSTRAP_TEST_LOG"
if bash scripts/prepare-aws-state.sh > /dev/null 2>&1; then exit 1; fi
! grep -Eq 's3api|terraform' "$BOOTSTRAP_TEST_LOG"
echo 'PASS: conta inesperada bloqueada antes de alterar recursos.'

export FAKE_ACCOUNT=123456789012 BUCKET_MODE=denied
: > "$BOOTSTRAP_TEST_LOG"
if bash scripts/prepare-aws-state.sh > /dev/null 2>&1; then exit 1; fi
! grep -Eq 'create-bucket|put-|terraform' "$BOOTSTRAP_TEST_LOG"
echo 'PASS: acesso negado nao e interpretado como bucket inexistente.'

export BUCKET_MODE=exists STATE_MODE=broken
: > "$BOOTSTRAP_TEST_LOG"
if bash scripts/prepare-aws-state.sh > /dev/null 2>&1; then exit 1; fi
! grep -Eq ' import | apply ' "$BOOTSTRAP_TEST_LOG"
echo 'PASS: erro de leitura do estado interrompe importacao e apply.'

export STATE_MODE=existing PLAN_FAIL=1
: > "$BOOTSTRAP_TEST_LOG"
if bash scripts/prepare-aws-state.sh > /dev/null 2>&1; then exit 1; fi
! grep -q ' apply ' "$BOOTSTRAP_TEST_LOG"
echo 'PASS: falha no plan impede apply.'
