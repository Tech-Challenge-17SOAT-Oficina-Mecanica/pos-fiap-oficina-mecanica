#!/usr/bin/env bash
set -euo pipefail

: "${SOURCE_SHA:?Published commit is required}"
: "${AWS_REGION:?AWS region is required}"
: "${ECR_REPOSITORY:?ECR repository is required}"
: "${GITHUB_RUN_ID:?Workflow run ID is required}"

# Automatic releases follow main; manual tests follow the explicitly selected branch.
deploy_ref=${DEPLOY_REF:-refs/heads/main}
current_revision=$(git ls-remote origin "$deploy_ref" | cut -f1)
if [[ "$SOURCE_SHA" != "$current_revision" ]]; then
  echo "$deploy_ref advanced; skipping deployment of $SOURCE_SHA."
  exit 0
fi

registry=$(aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" \
  --region "$AWS_REGION" --query 'repositories[0].repositoryUri' --output text)

# Cluster credentials remain in Kubernetes; never generate or overwrite them here.
kubectl get secret oficina-api-secret -n oficina -o name
kubectl get configmap oficina-api-config -n oficina -o name
kubectl auth can-i create jobs.batch -n oficina --quiet
kubectl auth can-i create deployments.apps -n oficina --quiet

(cd k8s/aws && kustomize edit set image "oficina-api=$registry:$SOURCE_SHA")
suffix="-${SOURCE_SHA:0:12}-${GITHUB_RUN_ID}"
job_name="oficina-migrations$suffix"
(
  cd k8s/aws/migrations
  kustomize edit set image "oficina-migrations=$registry:$SOURCE_SHA-migrations"
  kustomize edit set namesuffix -- "$suffix"
)

# Terraform owns namespace, credentials and RDS configuration; do not overwrite them.
# The RDS migration runs from inside EKS and must finish before updating the API.
kubectl delete job "$job_name" -n oficina --ignore-not-found --wait=true
kubectl apply -k k8s/aws/migrations
if ! kubectl wait --for=condition=complete "job/$job_name" -n oficina --timeout=630s; then
  kubectl logs "job/$job_name" -n oficina --all-containers=true --tail=100 || true
  kubectl describe job "$job_name" -n oficina
  exit 1
fi
kubectl logs "job/$job_name" -n oficina --all-containers=true

kubectl apply -k k8s/aws
kubectl rollout status deployment/oficina-api -n oficina --timeout=300s
kubectl port-forward service/oficina-api -n oficina 18080:80 > /tmp/oficina-port-forward.log 2>&1 &
forward_pid=$!
trap 'kill "$forward_pid" 2>/dev/null || true' EXIT
curl --fail --silent --show-error --retry 10 --retry-connrefused --retry-delay 2 http://localhost:18080/health
curl --fail --silent --show-error --retry 10 --retry-connrefused --retry-delay 2 http://localhost:18080/ready
