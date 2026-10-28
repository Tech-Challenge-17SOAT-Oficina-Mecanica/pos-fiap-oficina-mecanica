#!/usr/bin/env bash
set -euo pipefail

: "${SOURCE_SHA:?Published commit is required}"
: "${AWS_REGION:?AWS region is required}"
: "${ECR_REPOSITORY:?ECR repository is required}"
: "${GITHUB_RUN_ID:?Workflow run ID is required}"

# Ignore delayed releases when main has already advanced.
current_main=$(git ls-remote origin refs/heads/main | cut -f1)
if [[ "$SOURCE_SHA" != "$current_main" ]]; then
  echo "Main advanced; skipping deployment of $SOURCE_SHA."
  exit 0
fi

registry=$(aws ecr describe-repositories --repository-names "$ECR_REPOSITORY" \
  --region "$AWS_REGION" --query 'repositories[0].repositoryUri' --output text)

# Cluster credentials remain in Kubernetes; never generate or overwrite them here.
kubectl get secret oficina-api-secret -n oficina -o name
kubectl auth can-i create jobs.batch -n oficina --quiet
kubectl auth can-i create deployments.apps -n oficina --quiet

(cd k8s && kustomize edit set image "oficina-api=$registry:$SOURCE_SHA")
suffix="-${SOURCE_SHA:0:12}-${GITHUB_RUN_ID}"
job_name="oficina-migrations$suffix"
(
  cd k8s/migrations
  kustomize edit set image "oficina-migrations=$registry:$SOURCE_SHA-migrations"
  kustomize edit set namesuffix "$suffix"
)

# Prepare the database first; the application manifests are applied only after migration.
kubectl apply -f k8s/namespace.yaml -f k8s/configmap.yaml \
  -f k8s/postgres-storageclass.yaml -f k8s/postgres.yaml
kubectl rollout status statefulset/oficina-postgres -n oficina --timeout=300s
kubectl delete job "$job_name" -n oficina --ignore-not-found --wait=true
kubectl apply -k k8s/migrations
if ! kubectl wait --for=condition=complete "job/$job_name" -n oficina --timeout=630s; then
  kubectl logs "job/$job_name" -n oficina --all-containers=true --tail=100 || true
  kubectl describe job "$job_name" -n oficina
  exit 1
fi
kubectl logs "job/$job_name" -n oficina --all-containers=true

kubectl apply -k k8s
kubectl rollout status deployment/oficina-api -n oficina --timeout=300s
kubectl port-forward service/oficina-api -n oficina 18080:80 > /tmp/oficina-port-forward.log 2>&1 &
forward_pid=$!
trap 'kill "$forward_pid" 2>/dev/null || true' EXIT
curl --fail --silent --show-error --retry 10 --retry-connrefused --retry-delay 2 http://localhost:18080/health
curl --fail --silent --show-error --retry 10 --retry-connrefused --retry-delay 2 http://localhost:18080/ready
