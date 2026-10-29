#!/usr/bin/env bash
# Chamado pelo provisionamento depois de configurar o acesso temporario do runner ao EKS.
set -euo pipefail
source scripts/terraform-safe-apply.sh
platform_state=$(state_addresses infra/aws/platform)
if ! grep -Fxq 'kubernetes_namespace_v1.oficina' <<< "$platform_state" && [[ -n $(kubectl --context oficina-aws get namespace oficina --ignore-not-found -o name) ]]; then
  echo 'Namespace existente sem estado platform. Migre esse estado antes de continuar.' >&2
  exit 1
fi
safe_apply infra/aws/platform
