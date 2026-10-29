#!/usr/bin/env bash
# Chamado pelo provisionamento depois de configurar o acesso temporario do runner ao EKS.
set -euo pipefail
source scripts/terraform-safe-apply.sh
if ! kubectl --context oficina-aws auth can-i create namespaces --quiet; then
  echo 'A identidade AWS nao conseguiu autorizar a criacao de namespaces no EKS. Terraform nao sera aplicado.' >&2
  exit 1
fi
platform_state=$(state_addresses infra/aws/platform)
if ! grep -Fxq 'kubernetes_namespace_v1.oficina' <<< "$platform_state"; then
  # Uma falha de autenticacao nao pode ser interpretada como namespace ausente.
  namespace=$(kubectl --context oficina-aws get namespace oficina --ignore-not-found -o name)
  if [[ -n "$namespace" ]]; then
    echo 'Namespace existente sem estado platform. Migre esse estado antes de continuar.' >&2
    exit 1
  fi
fi
safe_apply infra/aws/platform
