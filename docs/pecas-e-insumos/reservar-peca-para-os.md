---
documento: Refinamento de Requisitos — Reservar Peça para OS
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Reservar Peça para OS

Este documento detalha a tarefa Reservar Peça para OS do contexto de Peças & Insumos.

## 5 · Reservar Peça para OS

### 5.1 Refinamento de Produto

**Persona**
Sistema, acionado pela aprovação do orçamento.
Beneficiário: mecânico, que passa a ter a peça garantida para executar o serviço.

**Objetivo**
Separar as peças de uma OS no estoque, garantindo que não sejam usadas em outro atendimento.

**Problema**
Duas OS aprovadas no mesmo dia podem depender da mesma última peça em estoque. Sem reserva,
quem chegar primeiro na prateleira leva, e o segundo cliente descobre o problema com o
veículo já desmontado. A reserva transforma "tem em estoque" em "está garantido para esta OS".

**Pré-condições**

- A OS deve existir e conter itens de peça registrados.
- O orçamento da OS deve estar aprovado.
- As peças devem estar cadastradas e ativas.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-24 | Permitir reservar todas as peças de uma OS em uma única operação. |
| RF-EST-25 | Validar o saldo disponível de cada peça antes de reservar. |
| RF-EST-26 | Aumentar o saldo reservado da peça e reduzir o saldo disponível. |
| RF-EST-27 | Vincular cada reserva à OS de origem. |
| RF-EST-28 | Informar quais peças não puderam ser reservadas por falta de saldo. |
| RF-EST-29 | Permitir liberar a reserva quando a OS for cancelada ou o orçamento recusado. |
| RF-EST-30 | Registrar a movimentação de reserva no histórico. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-21 | A operação deve ser feita por API RESTful. |
| RNF-EST-22 | A operação deve ser acessível somente por usuário ou serviço autorizado. |
| RNF-EST-23 | A reserva deve ser atômica e protegida contra concorrência — duas OS não podem reservar a mesma unidade simultaneamente. |
| RNF-EST-24 | A operação deve ser transacional — ou todas as peças da OS são reservadas, ou nenhuma é. |
| RNF-EST-25 | A operação deve ser idempotente por OS, para impedir reserva em dobro em caso de reprocessamento da mensagem. |
| RNF-EST-26 | A reserva não altera o saldo físico, apenas o saldo reservado. |

**Fluxo Principal**

1. O sistema recebe o pedido de reserva das peças de uma OS aprovada.
2. O sistema valida que a OS existe e possui orçamento aprovado.
3. O sistema verifica o saldo disponível de cada peça da OS.
4. O sistema aumenta o saldo reservado de cada peça.
5. O sistema vincula as reservas à OS.
6. O sistema registra a movimentação de reserva no histórico.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Peça com saldo disponível insuficiente | Não reserva nenhuma peça da OS, informa quais faltaram e sinaliza a indisponibilidade para que a compra seja solicitada. |
| A2 | Reserva já existente para a OS | Não reserva novamente e retorna a reserva vigente. |
| A3 | OS sem orçamento aprovado | Impede a reserva. |
| A4 | OS cancelada ou orçamento recusado | Libera as reservas vinculadas e devolve o saldo ao disponível. |
| A5 | Peça inativada após o orçamento | Impede a reserva e sinaliza a necessidade de substituição do item na OS. |
| A6 | Usuário ou serviço sem autorização | Impede a operação. |

**Saída**

- Confirmação da reserva com a relação de peças e quantidades reservadas para a OS; **ou**
- Indicação das peças indisponíveis que impediram a reserva.

**Pós-condições**

- O saldo reservado das peças está acrescido das quantidades da OS.
- O saldo disponível está reduzido na mesma proporção.
- O saldo físico permanece inalterado.
- As peças ficam vinculadas à OS e indisponíveis para outros atendimentos.
- A OS está liberada para iniciar a execução.

---

### 5.2 Refinamento Técnico

**Endpoint**

```http
POST   /estoque/reservas
DELETE /estoque/reservas/ordens-servico/{osId}
```

O `DELETE` atende a liberação da reserva (OS cancelada ou orçamento recusado).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `SERVICO` (chamada da política de orçamento aprovado), `MECANICO`, `GESTOR`
- Escopo: `estoque:movimentar`
- O mecânico não reserva diretamente: a reserva é consequência da aprovação do orçamento

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | Obrigatório neste endpoint |
| Body | `ordemServicoId` | uuid   | Obrigatório; OS de origem da reserva |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido |
| Body | `itens[].itemId` | uuid   | Peça a reservar; item do tipo `INSUMO` é rejeitado |
| Body | `itens[].quantidade` | int | Inteiro maior que zero |
| Path (DELETE) | `osId` | uuid   | OS cujas reservas ativas serão liberadas |

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "itens": [
    { "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", "quantidade": 4 },
    { "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90", "quantidade": 1 }
  ]
}
```

**Validações**

*Técnicas*

- `ordemServicoId` obrigatório.
- `itens` não vazio, sem `itemId` repetido.
- `quantidade` inteira maior que zero.
- Todos os itens do tipo `PECA` — insumo não é reservável.

*Negócio*

- A OS existe e possui orçamento aprovado vigente.
- Todas as peças estão ativas.
- `saldoDisponivel >= quantidade` para todas as peças; a reserva é tudo ou nada.
- Não existe reserva `ATIVA` para a mesma OS (idempotência de negócio).

**Processamento**

*Reserva (POST)*

1. Verificar o `Idempotency-Key`; se já processada, retornar a resposta original.
2. Consultar o módulo de OS: a OS existe e tem orçamento aprovado?
3. Verificar reserva `ATIVA` existente para a OS — se houver, retornar a reserva vigente.
4. Abrir transação.
5. Carregar todas as peças com `SELECT ... FOR UPDATE`, ordenadas por `item_id` — a ordem fixa evita deadlock entre transações concorrentes.
6. Para cada peça, conferir `saldo_fisico - saldo_reservado >= quantidade`.
7. Se qualquer peça falhar: rollback, montar a lista de indisponíveis e publicar `PecaIndisponivel`.
8. Se todas passarem: `saldo_reservado += quantidade`, com guarda na cláusula `WHERE`.
9. Inserir `reserva_estoque` com status `ATIVA` por item.
10. Inserir `movimentacao_estoque` do tipo `RESERVA`.
11. Commit.
12. Publicar o evento `PecaReservada`.

*Liberação (DELETE)*

1. Carregar as reservas `ATIVAS` da OS com lock.
2. `saldo_reservado -= quantidade` de cada peça.
3. Marcar as reservas como `LIBERADA`.
4. Inserir movimentação do tipo `LIBERACAO`.
5. Publicar o evento `ReservaLiberada`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `chave_idempotencia`, módulo de OS (externo ao agregado)
- Altera: `item_estoque.saldo_reservado`, `reserva_estoque` (insert e update de status), `movimentacao_estoque` (insert)
- Não altera: `saldo_fisico`
- Isolamento mínimo: `READ COMMITTED` com lock explícito de linha.

**Saída da API**

Sucesso (`POST`):

```json
{
  "reservaId": "8a4f1e07-c923-4d65-91b8-05e7a3c62f14",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "status": "ATIVA",
  "reservadoEm": "2026-08-12T15:20:00-03:00",
  "itens": [
    { "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", "codigo": "PC-0142", "quantidade": 4, "saldoDisponivelApos": 8 },
    { "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90", "codigo": "PC-0311", "quantidade": 1, "saldoDisponivelApos": 2 }
  ]
}
```

Saldo insuficiente (`409`):

```json
{
  "type": "https://api.oficina/errors/saldo-insuficiente",
  "title": "Saldo insuficiente",
  "status": 409,
  "detail": "Não foi possível reservar as peças da ordem de serviço 5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "erros": [{ "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90", "codigo": "PC-0311", "solicitado": 1, "disponivel": 0 }]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Reserva criada |
| `200` | Reserva já existente para a OS; repetição de `Idempotency-Key` |
| `204` | Reserva liberada (`DELETE`) |
| `400` | Body inválido; item do tipo `INSUMO` |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `estoque:movimentar` |
| `404` | OS ou peça não encontrada; nenhuma reserva ativa no `DELETE` |
| `409` | Saldo insuficiente — nenhuma peça foi reservada |
| `422` | OS sem orçamento aprovado; peça inativada após o orçamento |

**Dependências**

- `ItemEstoqueRepository`
- `ReservaEstoqueRepository`
- `MovimentacaoEstoqueRepository`
- Módulo Ordem de Serviço — consulta de status e do orçamento vigente
- Módulo Orçamento — origem do evento `OrcamentoAprovado` que dispara a chamada
- Serviço de idempotência
- Publicador de eventos de domínio

**Testes**

*Unitários*

- Reserva tudo ou nada: uma peça sem saldo impede as demais.
- Cálculo de `saldoDisponivelApos`.
- Rejeita item do tipo `INSUMO`.
- Liberação devolve exatamente a quantidade reservada.

*Integração*

- Reserva válida retorna `201`, `saldo_reservado` sobe e `saldo_fisico` não muda.
- Saldo insuficiente retorna `409` e nenhum saldo é alterado.
- OS sem orçamento aprovado retorna `422`.
- Segunda chamada com a mesma `Idempotency-Key` não duplica a reserva.
- `DELETE` devolve o saldo ao disponível e marca a reserva como `LIBERADA`.
- `DELETE` em OS sem reserva ativa retorna `404`.

*Concorrência (obrigatórios)*

- Duas OS reservando a última peça em paralelo: exatamente uma recebe `201`, a outra `409`.
- Reserva de duas OS com os mesmos itens em ordens diferentes de payload não gera deadlock (validação da ordenação por `item_id`).
- Teste de carga: N requisições simultâneas nunca deixam `saldo_reservado > saldo_fisico`.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `reservar()` na entidade `ItemEstoque` aumentando o saldo reservado
- [ ] Implementar o método `liberarReserva()` na entidade `ItemEstoque`
- [ ] Implementar a invariante de `saldoReservado` nunca maior que `saldoFisico`
- [ ] Implementar a invariante de não reservar acima do saldo disponível
- [ ] Implementar a entidade `ReservaEstoque` com status `ATIVA`, `LIBERADA` e `CONSUMIDA`

**Caso de uso**

- [ ] Implementar `ReservarPecasParaOS` com semântica tudo ou nada
- [ ] Implementar `LiberarReservaDaOS`
- [ ] Implementar o rollback total quando qualquer peça não tiver saldo, sem reservar nenhuma

**Repositório**

- [ ] Implementar `ReservaEstoqueRepository`
- [ ] Registrar `MovimentacaoEstoque` dos tipos `RESERVA` e `LIBERACAO`

**Integrações**

- [ ] Consultar o módulo de Ordem de Serviço para validar existência e orçamento aprovado
- [ ] Assinar o evento `OrcamentoAprovado` e acionar a reserva
- [ ] Assinar os eventos `OrcamentoRecusado` e `OSCancelada` e acionar a liberação

**Handler HTTP**

- [ ] Implementar `POST /estoque/reservas`
- [ ] Implementar `DELETE /estoque/reservas/ordens-servico/{osId}`

**Validações**

- [ ] Validar `ordemServicoId` obrigatório
- [ ] Validar `itens` não vazio e sem repetição
- [ ] Validar `quantidade` inteira maior que zero
- [ ] Rejeitar item do tipo `INSUMO` com `400`
- [ ] Validar que todas as peças estão ativas

**Concorrência e idempotência**

- [ ] Implementar `SELECT ... FOR UPDATE` nas linhas de `item_estoque` dentro da transação
- [ ] Ordenar o carregamento das linhas por `item_id` para evitar deadlock
- [ ] Implementar guarda na cláusula `WHERE` do `UPDATE` (saldo físico menos saldo reservado maior ou igual à quantidade)
- [ ] Tornar a `Idempotency-Key` obrigatória neste endpoint
- [ ] Implementar idempotência de negócio: reserva `ATIVA` existente para a OS retorna a reserva vigente com `200`

**Eventos**

- [ ] Publicar `PecaReservada`
- [ ] Publicar `PecaIndisponivel` no caminho triste
- [ ] Publicar `ReservaLiberada`

**Testes unitários**

- [ ] Semântica tudo ou nada com uma peça sem saldo
- [ ] Cálculo de `saldoDisponivelApos`
- [ ] Rejeição de item do tipo `INSUMO`
- [ ] Liberação devolvendo exatamente a quantidade reservada

**Testes de integração**

- [ ] Reserva válida com saldo reservado subindo e saldo físico inalterado
- [ ] Saldo insuficiente retornando `409` sem alterar nenhum saldo
- [ ] OS sem orçamento aprovado retornando `422`
- [ ] Segunda chamada com a mesma `Idempotency-Key` não duplicando a reserva
- [ ] `DELETE` devolvendo o saldo e marcando a reserva como `LIBERADA`
- [ ] `DELETE` em OS sem reserva ativa retornando `404`

**Testes de concorrência**

- [ ] Duas OS disputando a última peça: exatamente um `201` e um `409`
- [ ] Payloads com itens em ordem invertida sem gerar deadlock
- [ ] Teste de carga garantindo que `saldo_reservado` nunca ultrapassa `saldo_fisico`

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI, com o exemplo de erro `409` de saldo insuficiente

**Review**

- [ ] Code Review aprovado

---
