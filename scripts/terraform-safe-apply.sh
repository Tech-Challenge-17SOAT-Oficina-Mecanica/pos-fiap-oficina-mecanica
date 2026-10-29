#!/usr/bin/env bash
# Helper: aplica somente planos sem remocao/substituicao ou rotacao implicita de credenciais.
state_addresses() {
  local result
  if result=$(terraform -chdir="$1" state list 2>&1); then
    printf '%s\n' "$result"
  elif [[ "$result" != *'No state file was found'* && "$result" != *'No state file found'* ]]; then
    echo "$result" >&2
    return 1
  fi
}

verify_state_account() {
  terraform -chdir="$1" state pull | jq -e --arg account "$2" '
    all(.resources[]?.instances[]?.attributes.arn?;
      if type != "string" then true
      elif test("^arn:[^:]+:[^:]*:[^:]*:[0-9]{12}:") then
        (split(":")[4] == $account)
      else true end)' > /dev/null || {
    echo 'Estado invalido ou com recursos de outra conta; nao sera aplicado.' >&2
    return 1
  }
}

safe_apply() {
  local directory=$1
  terraform -chdir="$directory" validate
  terraform -chdir="$directory" plan -input=false -out=pipeline.tfplan
  if ! terraform -chdir="$directory" show -json pipeline.tfplan | jq -e '
    all(.resource_changes[]?;
      (.change.actions | index("delete") | not) and
      (if .type == "aws_db_instance" and .change.before != null then
        .change.before.password == .change.after.password
       elif .type == "kubernetes_secret_v1" and .change.before != null then
        .change.before.data.DB_PASSWORD == .change.after.data.DB_PASSWORD and
        .change.before.data.JWT_SECRET == .change.after.data.JWT_SECRET
       else true end))' > /dev/null; then
    echo 'Plano bloqueado: remocao/substituicao ou mudanca de credenciais. Revise sem aplicar.' >&2
    return 1
  fi
  terraform -chdir="$directory" apply -input=false pipeline.tfplan
}
