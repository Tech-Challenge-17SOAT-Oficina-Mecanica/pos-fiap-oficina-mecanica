#!/usr/bin/env bash
# Executa o ciclo completo de uma OS, do cadastro do cliente ate a entrega.
# Gera cliente e veiculo novos a cada execucao, entao pode rodar quantas vezes quiser.
set -euo pipefail
BASE=${BASE:-http://localhost:8080}

api() { # metodo caminho [corpo]
  local m=$1 p=$2 b=${3:-}
  if [ -n "$b" ]; then
    curl -s -X "$m" "$BASE$p" -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' -d "$b"
  else
    curl -s -X "$m" "$BASE$p" -H "Authorization: Bearer $TOKEN"
  fi
}
campo() { python3 -c "import sys,json;print(json.load(sys.stdin).get('$1',''))" 2>/dev/null; }

TOKEN=$(curl -s -X POST "$BASE/autenticacao/login" -H 'Content-Type: application/json' \
  -d '{"email":"mecanico@oficina.local","senha":"mecanico123"}' | campo accessToken)
[ -n "$TOKEN" ] || { echo "login falhou"; exit 1; }

CPF=$(python3 -c "
import random
n=[random.randint(0,9) for _ in range(9)]
def dv(b):
    s=sum(v*(len(b)+1-i) for i,v in enumerate(b)); r=(s*10)%11
    return 0 if r==10 else r
d1=dv(n); d2=dv(n+[d1]); print(''.join(map(str,n+[d1,d2])))")
PLACA=$(python3 -c "
import random,string
L=lambda: random.choice(string.ascii_uppercase); N=lambda: str(random.randint(0,9))
print(L()+L()+L()+N()+L()+N()+N())")

CLI=$(api POST /clientes "{\"nome\":\"Cliente $CPF\",\"documento\":\"$CPF\",\"tipoDocumento\":\"CPF\",\"telefone\":\"11988887777\",\"email\":\"c$CPF@example.com\"}" | campo id)
VEI=$(api POST "/clientes/$CLI/veiculos" "{\"placa\":\"$PLACA\",\"marca\":\"VW\",\"modelo\":\"Gol\",\"ano\":2020}" | campo id)
OS=$(api POST /ordens-servico "{\"clienteId\":\"$CLI\",\"veiculoId\":\"$VEI\"}" | campo ordemServicoId)
echo "  cliente=$CLI"
echo "  veiculo=$VEI"
echo "  os=$OS"

api POST "/ordens-servico/$OS/problema-relatado" '{"descricao":"Barulho na suspensao"}' >/dev/null
api POST "/ordens-servico/$OS/problemas" '{"descricao":"Amortecedor gasto","observacoes":"Troca"}' >/dev/null
api POST "/ordens-servico/$OS/servicos" '{"servicos":[{"servicoId":"40000000-0000-0000-0000-000000000001"}]}' >/dev/null
ORC=$(api POST "/ordens-servico/$OS/pecas" '{"itens":[{"itemId":"50000000-0000-0000-0000-000000000001","quantidade":2}]}' | campo orcamentoId)
echo "  orcamento=$ORC"

api POST "/orcamentos/$ORC/calcular" >/dev/null
./scripts/avancar-os.sh aprovacao "$OS" >/dev/null
api POST "/orcamentos/$ORC/aprovar" "{\"clienteId\":\"$CLI\"}" >/dev/null
api POST "/ordens-servico/$OS/execucao/iniciar" >/dev/null
./scripts/avancar-os.sh baixa "$OS" >/dev/null
api POST "/ordens-servico/$OS/finalizar" '{"observacoes":"Concluido"}' >/dev/null
api POST "/ordens-servico/$OS/entrega" '{"observacoes":"Retirado"}' >/dev/null

STATUS=$(docker compose exec -T postgres psql -U oficina -d oficina -tAc "SELECT status FROM ordem_servico WHERE id='$OS';" | tr -d ' \r')
echo "  status final: $STATUS"
