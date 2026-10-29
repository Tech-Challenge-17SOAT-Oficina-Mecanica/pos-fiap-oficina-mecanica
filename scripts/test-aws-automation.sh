#!/usr/bin/env bash
# Testes sem AWS: APIs e Terraform simulados; jq real verifica JSON e regras de protecao.
set -euo pipefail
test_dir=$(mktemp -d)
test_dir=$(cd "$test_dir" && pwd -P)
test_temp_root=$(cd "${TMPDIR:-/tmp}" && pwd -P)
[[ "$test_dir" == "$test_temp_root/"* ]] || exit 1
trap 'rm -rf "$test_dir"' EXIT
mkdir "$test_dir/bin"
export AWS_TEST_LOG="$test_dir/calls" AWS_TEST_CONFIG="$test_dir/config.json" AWS_TEST_PLAN="$test_dir/plan.json"
export FAKE_ACCOUNT=123456789012 EXPECTED_ACCOUNT_ID=123456789012 ADMIN_CIDR=203.0.113.10/32
export AWS_TEST_MODE=normal PLAN_FAIL=0 CLUSTER_NAME=oficina-academy
cat > "$test_dir/bin/aws" <<'EOF'
#!/usr/bin/env bash
echo "aws $*" >> "$AWS_TEST_LOG"
case "$1 $2" in
  'sts get-caller-identity') echo "$FAKE_ACCOUNT" ;;
  'ec2 describe-vpcs') echo 'vpc-test' ;;
  'iam list-roles')
    if [[ "$AWS_TEST_MODE" == ambiguous ]]; then
      echo '{"Roles":[{"RoleName":"a-LabEksClusterRole-x"},{"RoleName":"b-LabEksClusterRole-y"}]}'
    else
      echo '{"Roles":[{"RoleName":"a-LabEksClusterRole-x"},{"RoleName":"a-LabEksNodeRole-x"}]}'
    fi ;;
  'ec2 describe-subnets') echo '{"Subnets":[{"SubnetId":"subnet-a","State":"available","MapPublicIpOnLaunch":true,"AvailableIpAddressCount":100,"AvailabilityZone":"zone-a"},{"SubnetId":"subnet-b","State":"available","MapPublicIpOnLaunch":true,"AvailableIpAddressCount":100,"AvailabilityZone":"zone-b"}]}' ;;
  'ec2 describe-route-tables')
    if [[ "$AWS_TEST_MODE" == no-route ]]; then echo '{"RouteTables":[]}'
    else echo '{"RouteTables":[{"Associations":[{"Main":true}],"Routes":[{"State":"active","DestinationCidrBlock":"0.0.0.0/0","GatewayId":"igw-test"}]}]}'
    fi ;;
  'eks describe-cluster') cat "$AWS_TEST_CONFIG" ;;
  'eks update-cluster-config')
    while [[ "$1" != --resources-vpc-config ]]; do shift; done
    jq --argjson update "$2" '.publicAccessCidrs=$update.publicAccessCidrs' "$AWS_TEST_CONFIG" > "$AWS_TEST_CONFIG.next"
    mv "$AWS_TEST_CONFIG.next" "$AWS_TEST_CONFIG"
    echo update-test ;;
  'eks describe-update') echo Successful ;;
  'eks list-clusters')
    if [[ "$AWS_TEST_MODE" == existing-rds ]]; then echo '{"clusters":[]}'
    else echo '{"clusters":["oficina-academy"]}'; fi ;;
  'rds describe-db-instances')
    if [[ "$AWS_TEST_MODE" == existing-rds ]]; then echo '{"DBInstances":[{"DBInstanceIdentifier":"oficina-academy-postgres"}]}'
    else echo '{"DBInstances":[]}'; fi ;;
  'ecr describe-repositories') echo '{"repositories":[]}' ;;
esac
exit 0
EOF
cat > "$test_dir/bin/terraform" <<'EOF'
#!/usr/bin/env bash
echo "terraform $*" >> "$AWS_TEST_LOG"
if [[ "$2" == show ]]; then cat "$AWS_TEST_PLAN"; fi
if [[ "$2 $3" == 'state pull' ]]; then echo "$AWS_TEST_STATE_JSON"; fi
if [[ "$2" == plan && "$PLAN_FAIL" == 1 ]]; then exit 1; fi
exit 0
EOF
printf '#!/usr/bin/env bash\necho 198.51.100.42\n' > "$test_dir/bin/curl"
cat > "$test_dir/bin/kubectl" <<'EOF'
#!/usr/bin/env bash
echo "kubectl $*" >> "$AWS_TEST_LOG"
if [[ "$*" == *'auth can-i'* ]]; then
  [[ "${KUBE_TEST_MODE:-existing}" != denied ]]
  exit $?
fi
if [[ "${KUBE_TEST_MODE:-existing}" == lookup-failed ]]; then exit 1; fi
if [[ "${KUBE_TEST_MODE:-existing}" == absent ]]; then exit 0; fi
echo namespace/oficina
EOF
chmod +x "$test_dir/bin/"*
export PATH="$test_dir/bin:$PATH"

(source scripts/discover-aws-infra.sh; [[ "$TF_VAR_vpc_id" == vpc-test && "$TF_VAR_cluster_role_name" == a-LabEksClusterRole-x && "$TF_VAR_subnet_ids" == '["subnet-a","subnet-b"]' ]])
echo 'PASS: descoberta usa IDs da conta, sem valores pessoais fixos.'
export AWS_TEST_MODE=ambiguous
if bash scripts/discover-aws-infra.sh > /dev/null 2>&1; then exit 1; fi
export AWS_TEST_MODE=no-route
if bash scripts/discover-aws-infra.sh > /dev/null 2>&1; then exit 1; fi
export AWS_TEST_MODE=normal FAKE_ACCOUNT=999999999999
if bash scripts/discover-aws-infra.sh > /dev/null 2>&1; then exit 1; fi
export FAKE_ACCOUNT=123456789012
echo 'PASS: conta incorreta, funcoes ambiguas e rede inadequada sao rejeitadas.'

for plan in \
  '{"resource_changes":[{"type":"aws_eks_cluster","change":{"actions":["delete","create"]}}]}' \
  '{"resource_changes":[{"type":"aws_db_instance","change":{"actions":["update"],"before":{"password":"old"},"after":{"password":"new"}}}]}' \
  '{"resource_changes":[{"type":"kubernetes_secret_v1","change":{"actions":["update"],"before":{"data":{"DB_PASSWORD":"same","JWT_SECRET":"old"}},"after":{"data":{"DB_PASSWORD":"same","JWT_SECRET":"new"}}}}]}'; do
  echo "$plan" > "$AWS_TEST_PLAN"
  : > "$AWS_TEST_LOG"
  if bash -ec 'source scripts/terraform-safe-apply.sh; safe_apply fixture' > /dev/null 2>&1; then exit 1; fi
  ! grep -q ' apply ' "$AWS_TEST_LOG"
done
echo 'PASS: remocao e rotacao de credenciais bloqueadas antes de apply.'
echo '{"resource_changes":[{"type":"aws_eks_cluster","change":{"actions":["create"]}}]}' > "$AWS_TEST_PLAN"
bash -ec 'source scripts/terraform-safe-apply.sh; safe_apply fixture'
export PLAN_FAIL=1
: > "$AWS_TEST_LOG"
if bash -ec 'source scripts/terraform-safe-apply.sh; safe_apply fixture' > /dev/null 2>&1; then exit 1; fi
! grep -q ' apply ' "$AWS_TEST_LOG"
export PLAN_FAIL=0
echo 'PASS: plano seguro aplicado e falha de plan interrompe fluxo.'

export AWS_TEST_STATE_JSON='{"resources":[{"instances":[{"attributes":{"arn":"arn:aws:eks:us-east-1:999999999999:cluster/other-account"}}]}]}'
if bash -ec 'source scripts/terraform-safe-apply.sh; verify_state_account fixture 123456789012' > /dev/null 2>&1; then exit 1; fi
export AWS_TEST_STATE_JSON='{"resources":[{"instances":[{"attributes":{"arn":"arn:aws:eks:us-east-1:123456789012:cluster/same-account"}},{"attributes":{}},{"attributes":{"arn":null}},{"attributes":{"arn":"arn:aws:iam::aws:policy/ReadOnlyAccess"}}]}]}'
bash -ec 'source scripts/terraform-safe-apply.sh; verify_state_account fixture 123456789012'
echo 'PASS: estado de outra conta e rejeitado mesmo antes do refresh.'

echo '{"endpointPublicAccess":true,"publicAccessCidrs":["203.0.113.10/32"]}' > "$AWS_TEST_CONFIG"
bash scripts/with-eks-access.sh bash -ec 'jq -e '\''.publicAccessCidrs | index("198.51.100.42/32") != null'\'' "$AWS_TEST_CONFIG" > /dev/null'
jq -e '.publicAccessCidrs == ["203.0.113.10/32"]' "$AWS_TEST_CONFIG" > /dev/null
if bash scripts/with-eks-access.sh bash -c 'exit 7'; then exit 1; else [[ $? == 7 ]]; fi
jq -e '.publicAccessCidrs == ["203.0.113.10/32"]' "$AWS_TEST_CONFIG" > /dev/null
echo 'PASS: acesso /32 temporario removido no sucesso e na falha do comando.'

echo '{"endpointPublicAccess":true,"publicAccessCidrs":["203.0.113.10/32","198.51.100.42/32"]}' > "$AWS_TEST_CONFIG"
: > "$AWS_TEST_LOG"
bash scripts/with-eks-access.sh bash -c 'exit 0'
! grep -q 'update-cluster-config' "$AWS_TEST_LOG"
echo 'PASS: CIDR previamente autorizado nao e removido pelo runner.'

mkdir -p "$test_dir/fixture/scripts" "$test_dir/fixture/infra/aws/cluster" "$test_dir/fixture/infra/aws/platform"
cp scripts/{provision-aws,discover-aws-infra,prepare-aws-state,terraform-safe-apply}.sh "$test_dir/fixture/scripts/"
export TF_VAR_db_password=fake-password-123456 TF_VAR_jwt_secret=fake-jwt-key-only-for-tests-1234567890
: > "$AWS_TEST_LOG"
if (export TF_VAR_db_password=short; cd "$test_dir/fixture"; bash scripts/provision-aws.sh) > /dev/null 2>&1; then exit 1; fi
[[ ! -s "$AWS_TEST_LOG" ]]
echo 'PASS: credenciais da aplicacao invalidas sao rejeitadas antes de chamadas AWS.'
: > "$AWS_TEST_LOG"
if (cd "$test_dir/fixture"; bash scripts/provision-aws.sh) > /dev/null 2>&1; then exit 1; fi
! grep -q 'terraform -chdir=infra/aws/cluster apply' "$AWS_TEST_LOG"
grep -q 'eks list-clusters' "$AWS_TEST_LOG"
echo 'PASS: cluster existente sem estado remoto impede provisionamento duplicado.'
export AWS_TEST_MODE=existing-rds
: > "$AWS_TEST_LOG"
if (cd "$test_dir/fixture"; bash scripts/provision-aws.sh) > /dev/null 2>&1; then exit 1; fi
! grep -q 'terraform -chdir=infra/aws/cluster apply' "$AWS_TEST_LOG"
echo 'PASS: RDS existente sem estado remoto tambem bloqueia provisionamento.'

: > "$AWS_TEST_LOG"
if bash scripts/apply-aws-platform.sh > /dev/null 2>&1; then exit 1; fi
! grep -q ' apply ' "$AWS_TEST_LOG"
echo 'PASS: namespace existente sem estado platform bloqueia recriacao.'

for mode in denied lookup-failed; do
  export KUBE_TEST_MODE=$mode
  : > "$AWS_TEST_LOG"
  if bash scripts/apply-aws-platform.sh > /dev/null 2>&1; then exit 1; fi
  ! grep -Eq ' plan | apply ' "$AWS_TEST_LOG"
done
echo 'PASS: falha de autorizacao ou consulta Kubernetes interrompe antes do plano.'

export KUBE_TEST_MODE=absent
: > "$AWS_TEST_LOG"
bash scripts/apply-aws-platform.sh
grep -q 'terraform -chdir=infra/aws/platform apply' "$AWS_TEST_LOG"
echo 'PASS: acesso autorizado e namespace ausente permitem aplicar platform.'
