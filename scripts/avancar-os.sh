#!/usr/bin/env bash
# Cobre as transicoes que a API ainda nao expoe.
#
#   ./scripts/avancar-os.sh aprovacao <osId>   EM_DIAGNOSTICO -> AGUARDANDO_APROVACAO
#   ./scripts/avancar-os.sh baixa     <osId>   baixa as reservas para permitir finalizar
#
# A primeira e uma lacuna real: nenhum ponto do codigo escreve AGUARDANDO_APROVACAO.
# A segunda tem rota propria (POST /estoque/saidas), mas exige um payload de
# movimentacao que a colecao ainda nao monta.
set -euo pipefail

ACAO="${1:?informe: aprovacao ou baixa}"
OS_ID="${2:?informe o id da OS}"

case "$ACAO" in
  aprovacao)
    SQL="UPDATE ordem_servico SET status='AGUARDANDO_APROVACAO' WHERE id='${OS_ID}';"
    ;;
  baixa)
    SQL="UPDATE reserva_estoque SET status='BAIXADA'
          WHERE ordem_servico_item_id IN (
            SELECT id FROM ordem_servico_item WHERE ordem_servico_id='${OS_ID}'
          ) AND status='ATIVA';
         UPDATE ordem_servico_servico SET status='CONCLUIDO'
          WHERE ordem_servico_id='${OS_ID}';"
    ;;
  *) echo "acao invalida: use aprovacao ou baixa" >&2; exit 1 ;;
esac

docker compose exec -T postgres psql -U oficina -d oficina -q -v ON_ERROR_STOP=1 -c "${SQL}"
echo "OS ${OS_ID}: ${ACAO} aplicada."
