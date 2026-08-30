#!/usr/bin/env bash
# Deixa a OS 1 pronta para ser finalizada.
#
# O seed cria a OS 1 em EM_EXECUCAO, mas com servico pendente, orcamento complementar
# em CRIADO e reservas ATIVAS — tres bloqueios legitimos da regra de negocio. Este
# script resolve os tres para que o fluxo de notificacao possa ser exercitado.
set -euo pipefail

OS_ID="${1:-70000000-0000-0000-0000-000000000001}"

docker compose exec -T postgres psql -U oficina -d oficina -v ON_ERROR_STOP=1 <<SQL
UPDATE ordem_servico SET status = 'EM_EXECUCAO', finalizada_em = NULL, entregue_em = NULL
  WHERE id = '${OS_ID}';

UPDATE ordem_servico_servico SET status = 'CONCLUIDO'
  WHERE ordem_servico_id = '${OS_ID}';

UPDATE orcamento SET status = 'APROVADO', aprovado_em = CURRENT_TIMESTAMP
  WHERE ordem_servico_id = '${OS_ID}' AND tipo_orcamento = 'COMPLEMENTAR' AND status = 'CRIADO';

UPDATE reserva_estoque SET status = 'BAIXADA'
  WHERE ordem_servico_item_id IN (
    SELECT id FROM ordem_servico_item WHERE ordem_servico_id = '${OS_ID}'
  ) AND status = 'ATIVA';
SQL

echo "OS ${OS_ID} pronta para finalizar."
