#!/usr/bin/env bash
# Libera apenas o IPv4 /32 do runner durante o comando e remove essa entrada ao terminar.
# Requer endpoint publico EKS; nao usa 0.0.0.0/0 nem altera o acesso privado.
set -euo pipefail
: "${CLUSTER_NAME:?Informe o cluster}"
: "${EXPECTED_ACCOUNT_ID:?Informe a conta esperada}"
[[ $(aws sts get-caller-identity --query Account --output text) == "$EXPECTED_ACCOUNT_ID" ]] || exit 1
runner_ip=$(curl --fail --silent --show-error -4 https://checkip.amazonaws.com)
[[ "$runner_ip" =~ ^([0-9]{1,3}\.){3}[0-9]{1,3}$ ]] || { echo 'IPv4 do runner invalido.' >&2; exit 1; }
runner_cidr="$runner_ip/32"
get_config() { aws eks describe-cluster --name "$CLUSTER_NAME" --region us-east-1 --query cluster.resourcesVpcConfig --output json; }
update_cidrs() {
  local update_id status
  update_id=$(aws eks update-cluster-config --name "$CLUSTER_NAME" --region us-east-1 \
    --resources-vpc-config "$(jq -cn --argjson cidrs "$1" '{publicAccessCidrs:$cidrs}')" --query update.id --output text)
  for ((attempt=0; attempt<120; attempt++)); do
    status=$(aws eks describe-update --name "$CLUSTER_NAME" --region us-east-1 --update-id "$update_id" --query update.status --output text)
    case "$status" in
      Successful) return 0 ;;
      Failed|Cancelled) echo "Atualizacao EKS terminou em $status" >&2; return 1 ;;
    esac
    sleep 5
  done
  echo 'Tempo esgotado ao atualizar acesso EKS.' >&2
  return 1
}
config=$(get_config)
jq -e '.endpointPublicAccess == true' <<< "$config" > /dev/null || { echo 'Endpoint EKS nao e publico; use runner dentro da VPC.' >&2; exit 1; }
cidrs=$(jq -c '.publicAccessCidrs' <<< "$config")
added=0
cleanup() {
  local result=$? current remaining
  trap - EXIT
  if [[ "$added" == 1 ]]; then
    if current=$(get_config) && remaining=$(jq -ce --arg cidr "$runner_cidr" '.publicAccessCidrs | map(select(. != $cidr)) | select(length > 0)' <<< "$current") && update_cidrs "$remaining"; then
      echo 'Acesso temporario do runner removido.'
    else
      echo 'Nao foi possivel confirmar a remocao do CIDR temporario. Revise o endpoint EKS antes de continuar.' >&2
      result=1
    fi
  fi
  exit "$result"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
if ! jq -e --arg cidr "$runner_cidr" 'index($cidr) != null' <<< "$cidrs" > /dev/null; then
  [[ $(jq length <<< "$cidrs") -lt 50 ]] || { echo 'Limite de CIDRs atingido.' >&2; exit 1; }
  added=1
  update_cidrs "$(jq -c --arg cidr "$runner_cidr" '. + [$cidr]' <<< "$cidrs")"
fi
"$@"
