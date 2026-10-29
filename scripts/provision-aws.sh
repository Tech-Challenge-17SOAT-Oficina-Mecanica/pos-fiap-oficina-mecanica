#!/usr/bin/env bash
# Uso pela pipeline manual. Cria/atualiza AWS e platform, sem publicar a API.
set -euo pipefail
: "${TF_VAR_db_password:?Configure o Secret AWS_DB_PASSWORD}"
: "${TF_VAR_jwt_secret:?Configure o Secret AWS_JWT_SECRET}"
[[ "$TF_VAR_db_password" =~ ^[A-Za-z0-9_-]{16,}$ && ${#TF_VAR_jwt_secret} -ge 32 ]] || {
  echo 'Secrets da aplicacao invalidos: senha com 16+ caracteres permitidos e chave JWT com 32+.' >&2
  exit 1
}
if [[ -s infra/aws/cluster/terraform.tfstate || -s infra/aws/platform/terraform.tfstate ]]; then
  echo 'Ha estados locais. Migre-os com backup antes de usar a automacao remota.' >&2
  exit 1
fi
source scripts/discover-aws-infra.sh
bash scripts/prepare-aws-state.sh
source scripts/terraform-safe-apply.sh

for module in cluster platform; do
  terraform -chdir="infra/aws/$module" init -input=false -reconfigure \
    -backend-config="bucket=$STATE_BUCKET" -backend-config="key=$module/terraform.tfstate" \
    -backend-config='region=us-east-1' -backend-config='encrypt=true' \
    -backend-config='use_lockfile=true' \
    -backend-config="allowed_account_ids=[\"$EXPECTED_ACCOUNT_ID\"]"
done
state=$(state_addresses infra/aws/cluster)
if [[ -n "$state" ]]; then verify_state_account infra/aws/cluster "$EXPECTED_ACCOUNT_ID"; fi
clusters=$(aws eks list-clusters --output json)
databases=$(aws rds describe-db-instances --output json)
if { ! grep -Fxq 'aws_eks_cluster.oficina' <<< "$state" && jq -e --arg name "$TF_VAR_cluster_name" '.clusters | index($name) != null' <<< "$clusters" > /dev/null; } ||
   { ! grep -Fxq 'aws_db_instance.oficina' <<< "$state" && jq -e --arg name "$TF_VAR_cluster_name-postgres" 'any(.DBInstances[]; .DBInstanceIdentifier == $name)' <<< "$databases" > /dev/null; }; then
  echo 'Cluster/RDS existente sem estado remoto correspondente. Transfira o estado ou planeje importacao; nao sera recriado.' >&2
  exit 1
fi
repositories=$(aws ecr describe-repositories --output json)
if ! grep -Fxq 'aws_ecr_repository.api' <<< "$state" && jq -e --arg name "$TF_VAR_ecr_repository" 'any(.repositories[]; .repositoryName == $name)' <<< "$repositories" > /dev/null; then
  terraform -chdir=infra/aws/cluster import -input=false aws_ecr_repository.api "$TF_VAR_ecr_repository"
fi
safe_apply infra/aws/cluster
export TF_VAR_db_host
TF_VAR_db_host=$(terraform -chdir=infra/aws/cluster output -raw db_host)
export CLUSTER_NAME="$TF_VAR_cluster_name"
export KUBECONFIG="${RUNNER_TEMP:?Runner temporario necessario}/oficina-kubeconfig"
export TF_VAR_kubeconfig_path="$KUBECONFIG"
aws eks update-kubeconfig --name "$CLUSTER_NAME" --region "$AWS_REGION" --alias oficina-aws --kubeconfig "$KUBECONFIG"
bash scripts/with-eks-access.sh bash scripts/apply-aws-platform.sh
echo 'Infraestrutura preparada. O deploy da aplicacao e um workflow manual separado.'
