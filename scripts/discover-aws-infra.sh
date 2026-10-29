#!/usr/bin/env bash
# Source pela preparacao de infraestrutura. Descobre IDs na conta autenticada, sem alterar AWS.
set -euo pipefail
: "${EXPECTED_ACCOUNT_ID:?Informe o ID da conta esperada}"
: "${ADMIN_CIDR:?Informe um IP administrativo autorizado com /32}"
[[ "$ADMIN_CIDR" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}/32$ ]] || { echo 'Use um IPv4 administrativo /32.' >&2; exit 1; }
jq -en --arg cidr "$ADMIN_CIDR" '$cidr | split("/")[0] | split(".") | all(.[]; (tonumber <= 255))' > /dev/null || exit 1
export AWS_REGION=us-east-1 AWS_DEFAULT_REGION=us-east-1 AWS_PAGER=""
account_id=$(aws sts get-caller-identity --query Account --output text)
[[ "$account_id" =~ ^[0-9]{12}$ && "$account_id" == "$EXPECTED_ACCOUNT_ID" ]] || {
  echo 'Credenciais de outra conta; interrompido.' >&2; exit 1;
}
export TF_VAR_cluster_name="${CLUSTER_NAME:-oficina-academy}"
export TF_VAR_ecr_repository="${ECR_REPOSITORY:-oficina-mecanica}"
[[ "$TF_VAR_cluster_name" =~ ^[a-z][a-z0-9-]{0,39}$ && "$TF_VAR_cluster_name" != *--* && "$TF_VAR_cluster_name" != *- ]] || exit 1
[[ "$TF_VAR_ecr_repository" =~ ^[a-z0-9][a-z0-9._/-]*$ ]] || exit 1
export STATE_BUCKET="oficina-tfstate-$account_id-$AWS_REGION"
export TF_VAR_vpc_id="${AWS_VPC_ID:-$(aws ec2 describe-vpcs --filters Name=is-default,Values=true --query 'Vpcs[0].VpcId' --output text)}"
[[ "$TF_VAR_vpc_id" == vpc-* ]] || { echo 'VPC padrao nao encontrada.' >&2; exit 1; }

roles=$(aws iam list-roles --output json)
export TF_VAR_cluster_role_name="${AWS_EKS_CLUSTER_ROLE:-$(jq -er '[.Roles[].RoleName | select(test("(^|-)LabEksClusterRole(-|$)"))] | if length == 1 then .[0] else error("Esperada uma funcao de cluster Academy") end' <<< "$roles")}"
export TF_VAR_node_role_name="${AWS_EKS_NODE_ROLE:-$(jq -er '[.Roles[].RoleName | select(test("(^|-)LabEksNodeRole(-|$)"))] | if length == 1 then .[0] else error("Esperada uma funcao de nos Academy") end' <<< "$roles")}"
[[ -n "$TF_VAR_cluster_role_name" && -n "$TF_VAR_node_role_name" ]] || exit 1

subnets=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$TF_VAR_vpc_id" --output json)
export TF_VAR_subnet_ids
TF_VAR_subnet_ids=$(jq -cer '[.Subnets[] | select(.State == "available" and .MapPublicIpOnLaunch and .AvailableIpAddressCount >= 16)] | sort_by(.AvailabilityZone, .SubnetId) | unique_by(.AvailabilityZone) | if length >= 2 then .[:2] | map(.SubnetId) else error("Duas zonas com sub-redes publicas sao necessarias") end' <<< "$subnets")
routes=$(aws ec2 describe-route-tables --filters "Name=vpc-id,Values=$TF_VAR_vpc_id" --output json)
while read -r subnet; do
  jq -e --arg subnet "$subnet" '
    .RouteTables | ([.[] | select(any(.Associations[]; .SubnetId == $subnet))][0] //
    [.[] | select(any(.Associations[]; .Main == true))][0]) |
    any(.Routes[]; .State == "active" and .DestinationCidrBlock == "0.0.0.0/0" and ((.GatewayId // "") | startswith("igw-")))
  ' <<< "$routes" > /dev/null || { echo "Sub-rede sem rota publica: $subnet" >&2; exit 1; }
done < <(jq -r '.[]' <<< "$TF_VAR_subnet_ids")
export TF_VAR_admin_cidrs
TF_VAR_admin_cidrs=$(jq -cn --arg cidr "$ADMIN_CIDR" '[$cidr]')
echo "Infraestrutura selecionada na conta $account_id; IDs obtidos da AWS."
