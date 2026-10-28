#!/usr/bin/env bash
set -euo pipefail

test_dir=$(mktemp -d)
trap 'rm -rf "$test_dir"' EXIT
mkdir "$test_dir/bin"
export DEPLOY_TEST_LOG="$test_dir/commands.log"
export SOURCE_SHA=aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa
export REMOTE_SHA="$SOURCE_SHA"
export DEPLOY_REF=refs/heads/main
export AWS_REGION=us-east-1 ECR_REPOSITORY=oficina-mecanica GITHUB_RUN_ID=123
export MIGRATION_FAIL=0

cat > "$test_dir/bin/git" <<'EOF'
#!/usr/bin/env bash
printf '%s\t%s\n' "$REMOTE_SHA" "$3"
EOF
cat > "$test_dir/bin/aws" <<'EOF'
#!/usr/bin/env bash
echo example.dkr.ecr.us-east-1.amazonaws.com/oficina-mecanica
EOF
cat > "$test_dir/bin/kubectl" <<'EOF'
#!/usr/bin/env bash
echo "kubectl $*" >> "$DEPLOY_TEST_LOG"
if [[ "$1" == wait && "$MIGRATION_FAIL" == 1 ]]; then exit 1; fi
EOF
cat > "$test_dir/bin/kustomize" <<'EOF'
#!/usr/bin/env bash
if [[ "$1 $2 $3" == 'edit set namesuffix' ]]; then
  [[ "$4" == -- && "$5" == -* ]] || exit 1
fi
exit 0
EOF
printf '#!/usr/bin/env bash\nexit 0\n' > "$test_dir/bin/curl"
chmod +x "$test_dir/bin/"*
export PATH="$test_dir/bin:$PATH"

export MIGRATION_FAIL=1
if bash scripts/deploy-eks.sh; then
  echo 'ERROR: deploy accepted a failed migration' >&2
  exit 1
fi
if grep -Fxq 'kubectl apply -k k8s' "$DEPLOY_TEST_LOG"; then
  echo 'ERROR: application applied after migration failure' >&2
  exit 1
fi

export MIGRATION_FAIL=0
: > "$DEPLOY_TEST_LOG"
bash scripts/deploy-eks.sh
grep -Fxq 'kubectl apply -k k8s' "$DEPLOY_TEST_LOG"

export DEPLOY_REF=refs/heads/codex/eks-migrations-kustomize
: > "$DEPLOY_TEST_LOG"
bash scripts/deploy-eks.sh
grep -Fxq 'kubectl apply -k k8s' "$DEPLOY_TEST_LOG"

export REMOTE_SHA=bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb
: > "$DEPLOY_TEST_LOG"
bash scripts/deploy-eks.sh
[[ ! -s "$DEPLOY_TEST_LOG" ]]
echo 'PASS: migration failure blocks deploy; main/manual branch deploy; stale release skipped.'
