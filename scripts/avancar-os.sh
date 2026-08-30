#!/usr/bin/env bash
# Atalho para a baixa de reservas no teste manual.
#
#   ./scripts/avancar-os.sh baixa <osId>
#
# A transicao para AGUARDANDO_APROVACAO saiu daqui: agora existe a rota
# POST /orcamentos/{orcamentoId}/enviar, que faz isso e avisa o cliente.
#
# A baixa tem rota propria (POST /estoque/saidas), mas exige um payload de
# movimentacao que a colecao ainda nao monta. Enquanto isso, este atalho serve.
set -euo pipefail

ACAO="${1:?informe: baixa}"
OS_ID="${2:?informe o id da OS}"

case "$ACAO" in
  baixa)
    SQL="UPDATE reserva_estoque SET status='BAIXADA'
          WHERE ordem_servico_item_id IN (
            SELECT id FROM ordem_servico_item WHERE ordem_servico_id='${OS_ID}'
          ) AND status='ATIVA';
         UPDATE ordem_servico_servico SET status='CONCLUIDO'
          WHERE ordem_servico_id='${OS_ID}';"
    ;;
  aprovacao)
    echo "essa transicao agora tem rota propria: POST /orcamentos/{orcamentoId}/enviar" >&2
    exit 1
    ;;
  *) echo "acao invalida: use baixa" >&2; exit 1 ;;
esac

docker compose exec -T postgres psql -U oficina -d oficina -q -v ON_ERROR_STOP=1 -c "${SQL}"
echo "OS ${OS_ID}: ${ACAO} aplicada."
