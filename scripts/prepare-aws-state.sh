#!/usr/bin/env bash
# Uso na pipeline: EXPECTED_ACCOUNT_ID=123456789012 bash scripts/prepare-aws-state.sh
# Cria/protege somente o bucket de estados; nao provisiona EKS/RDS nem migra estados locais.
set -euo pipefail

: "${EXPECTED_ACCOUNT_ID:?Informe a conta AWS esperada antes de preparar o backend}"
export AWS_REGION=us-east-1
export AWS_DEFAULT_REGION="$AWS_REGION"
export AWS_PAGER=""
account_id=$(aws sts get-caller-identity --query Account --output text)
if [[ ! "$account_id" =~ ^[0-9]{12}$ || "$account_id" != "$EXPECTED_ACCOUNT_ID" ]]; then
  echo 'Conta autenticada diferente da conta esperada; nenhum bucket foi alterado.' >&2
  exit 1
fi
bucket="oficina-tfstate-$account_id-$AWS_REGION"

if ! bucket_status=$(aws s3api head-bucket --bucket "$bucket" --expected-bucket-owner "$account_id" 2>&1); then
  # 403 tambem pode significar falta de permissao: nunca tratar como bucket inexistente.
  if [[ "$bucket_status" != *'(404)'* && "$bucket_status" != *'NoSuchBucket'* ]]; then
    echo "$bucket_status" >&2
    exit 1
  fi
  aws s3api create-bucket --bucket "$bucket" --region "$AWS_REGION" \
    --object-ownership BucketOwnerEnforced > /dev/null
fi

# O backend precisa existir e estar protegido antes de o Terraform salvar seu proprio estado.
aws s3api put-public-access-block --bucket "$bucket" --expected-bucket-owner "$account_id" \
  --public-access-block-configuration BlockPublicAcls=true,IgnorePublicAcls=true,BlockPublicPolicy=true,RestrictPublicBuckets=true
aws s3api put-bucket-versioning --bucket "$bucket" --expected-bucket-owner "$account_id" \
  --versioning-configuration Status=Enabled
aws s3api put-bucket-encryption --bucket "$bucket" --expected-bucket-owner "$account_id" \
  --server-side-encryption-configuration '{"Rules":[{"ApplyServerSideEncryptionByDefault":{"SSEAlgorithm":"AES256"}}]}'
aws s3api put-bucket-ownership-controls --bucket "$bucket" --expected-bucket-owner "$account_id" \
  --ownership-controls 'Rules=[{ObjectOwnership=BucketOwnerEnforced}]'

terraform -chdir=infra/aws/state init -input=false -reconfigure \
  -backend-config="bucket=$bucket" -backend-config='key=bootstrap/terraform.tfstate' \
  -backend-config="region=$AWS_REGION" -backend-config='encrypt=true' \
  -backend-config='use_lockfile=true' \
  -backend-config="allowed_account_ids=[\"$account_id\"]"
terraform -chdir=infra/aws/state validate

if ! existing_state=$(terraform -chdir=infra/aws/state state list 2>&1); then
  if [[ "$existing_state" == *'No state file was found'* || "$existing_state" == *'No state file found'* ]]; then
    existing_state=''
  else
    echo "$existing_state" >&2
    exit 1
  fi
fi
for type in aws_s3_bucket_public_access_block aws_s3_bucket_versioning \
  aws_s3_bucket_server_side_encryption_configuration aws_s3_bucket_ownership_controls; do
  address="$type.terraform"
  if ! grep -Fxq "$address" <<< "$existing_state"; then
    terraform -chdir=infra/aws/state import -input=false "$address" "$bucket"
  fi
done

terraform -chdir=infra/aws/state plan -input=false -out=state.tfplan
terraform -chdir=infra/aws/state apply -input=false state.tfplan
echo "Backend preparado para conta $account_id: $bucket"
if [[ -n "${GITHUB_OUTPUT:-}" ]]; then
  printf 'account_id=%s\nbucket=%s\n' "$account_id" "$bucket" >> "$GITHUB_OUTPUT"
fi
