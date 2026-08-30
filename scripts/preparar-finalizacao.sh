#!/usr/bin/env bash
# Deixa uma OS pronta para ser finalizada e imprime o id.
#
# A regra de negocio exige: status EM_EXECUCAO, todos os servicos CONCLUIDO, nenhum
# orcamento complementar em CRIADO e nenhuma reserva ATIVA. O seed nao deixa nenhuma OS
# nesse estado, e depois de um teste a OS fica ENTREGUE — por isso este script existe.
#
#   ./scripts/preparar-finalizacao.sh              # usa a OS 1 do seed
#   ./scripts/preparar-finalizacao.sh <uuid-da-os> # usa outra
set -euo pipefail

OS_ID="${1:-70000000-0000-0000-0000-000000000001}"

docker compose exec -T postgres psql -U oficina -d oficina -q -v ON_ERROR_STOP=1 <<SQL
UPDATE ordem_servico
   SET status = 'EM_EXECUCAO', finalizada_em = NULL, entregue_em = NULL,
       valor_final = NULL, observacoes_finalizacao = NULL, observacoes_entrega = NULL
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

echo "OS pronta para finalizar: ${OS_ID}"
echo "Use este id na variavel osId do Bruno."
