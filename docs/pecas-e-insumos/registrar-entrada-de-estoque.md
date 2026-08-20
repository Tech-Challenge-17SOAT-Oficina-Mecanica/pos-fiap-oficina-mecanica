---
documento: Refinamento de Requisitos — Registrar Entrada de Estoque
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Registrar Entrada de Estoque

Este documento detalha a tarefa Registrar Entrada de Estoque do contexto de Peças & Insumos.

## 4 · Registrar Entrada de Estoque

### 4.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Registrar o recebimento de peças e insumos, aumentando o saldo físico do estoque.

**Problema**
Sem registro de entrada, o saldo do sistema diverge do saldo real da prateleira. A
consequência é dupla: o mecânico deixa de reservar peça que existe, ou reserva peça que já
acabou — e a oficina só descobre no momento de executar o serviço.

**Pré-condições**

- O item (peça ou insumo) deve estar cadastrado e ativo.
- Deve existir um documento de origem do recebimento (nota fiscal ou pedido de compra).
- O usuário deve estar autorizado a movimentar estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-17 | Permitir registrar entrada informando item, quantidade, custo unitário e documento de origem. |
| RF-EST-18 | Permitir registrar a entrada de vários itens em um mesmo recebimento. |
| RF-EST-19 | Validar que a quantidade informada é maior que zero. |
| RF-EST-20 | Vincular a entrada ao pedido de compra correspondente, quando houver. |
| RF-EST-21 | Atualizar o saldo físico do item. |
| RF-EST-22 | Registrar a movimentação no histórico de estoque. |
| RF-EST-23 | Atualizar a situação do pedido de compra quando o recebimento for total. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-16 | A operação deve ser feita por API RESTful. |
| RNF-EST-17 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-EST-18 | A entrada deve ser transacional — ou todos os itens do recebimento entram, ou nenhum entra. |
| RNF-EST-19 | A operação deve ser idempotente em relação ao documento de origem, para impedir dupla contagem em caso de reenvio. |
| RNF-EST-20 | A movimentação deve ser auditável e o histórico imutável. |

**Fluxo Principal**

1. O mecânico informa o documento de origem do recebimento.
2. O mecânico informa os itens, as quantidades e os custos unitários.
3. O sistema valida os itens e as quantidades.
4. O sistema atualiza o saldo físico de cada item.
5. O sistema registra a movimentação de entrada no histórico.
6. O sistema atualiza a situação do pedido de compra vinculado.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Item não encontrado ou inativo | Impede a entrada e informa qual item está irregular. |
| A2 | Quantidade menor ou igual a zero | Impede a entrada. |
| A3 | Documento de origem já registrado | Informa que o recebimento já foi lançado e não altera o saldo. |
| A4 | Recebimento parcial | Registra a quantidade recebida e mantém o pedido de compra em aberto com o saldo pendente. |
| A5 | Quantidade recebida maior que a pedida | Alerta a divergência e exige confirmação antes de gravar. |
| A6 | Usuário sem autorização | Impede a operação. |

**Saída**

- Confirmação da entrada com o saldo físico atualizado de cada item; **ou**
- Indicação do motivo pelo qual a entrada foi recusada.

**Pós-condições**

- O saldo físico dos itens recebidos está acrescido da quantidade informada.
- O saldo reservado permanece inalterado.
- A movimentação de entrada está registrada no histórico.
- O pedido de compra vinculado está atualizado — concluído ou parcialmente atendido.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/entradas
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:movimentar`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | Recomendado; impede dupla contagem em caso de reenvio |
| Body | `documentoOrigem` | string | Obrigatório; nota fiscal ou documento do recebimento |
| Body | `fornecedorId` | uuid   | Fornecedor do recebimento |
| Body | `pedidoCompraId` | uuid   | Pedido de compra vinculado, quando houver |
| Body | `itens[]` | array | Obrigatório, não vazio, máximo 200 linhas |
| Body | `itens[].itemId` | uuid   | Item recebido; não pode repetir no mesmo payload |
| Body | `itens[].quantidade` | decimal | Maior que zero |
| Body | `itens[].custoUnitario` | decimal | Maior que zero |
| Body | `confirmarDivergencia` | boolean | Obrigatório quando a quantidade recebida excede a pedida |

```json
{
  "documentoOrigem": "NF-88421",
  "fornecedorId": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
  "pedidoCompraId": "6b2e9f47-3a15-4c80-9d62-7e10b4f83a95",
  "itens": [
    { "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", "quantidade": 10, "custoUnitario": 118.40 },
    { "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48", "quantidade": 20.0, "custoUnitario": 30.10 }
  ]
}
```

**Validações**

*Técnicas*

- `documentoOrigem` obrigatório.
- `itens` não vazio, com no máximo 200 linhas.
- `quantidade` maior que zero em cada linha.
- `custoUnitario` maior que zero.
- Sem `itemId` repetido no mesmo payload.

*Negócio*

- Todos os `itemId` existem e estão ativos.
- `documentoOrigem` ainda não registrado — chave única em `movimentacao_estoque`.
- Quando há `pedidoCompraId`, os itens devem pertencer ao pedido.
- Quantidade recebida maior que a pedida exige `confirmarDivergencia: true` no body.

**Processamento**

1. Verificar o `Idempotency-Key` — se já processada, retornar a resposta original.
2. Abrir transação.
3. Validar o payload e carregar todos os itens com `SELECT ... FOR UPDATE`.
4. Verificar duplicidade de `documentoOrigem`.
5. Se houver `pedidoCompraId`, carregar o pedido e conferir divergência de quantidade.
6. Para cada linha: `saldoFisico += quantidade`.
7. Inserir uma `movimentacao_estoque` do tipo `ENTRADA` por linha.
8. Atualizar `quantidade_recebida` em `pedido_compra_item`.
9. Recalcular o status do pedido: `ABERTO` para `PARCIAL` ou `CONCLUIDO`.
10. Commit.
11. Publicar o evento `EntradaRegistrada`.

**Persistência**

- Consulta: `item_estoque`, `pedido_compra`, `pedido_compra_item`, `chave_idempotencia`
- Altera: `item_estoque.saldo_fisico`, `movimentacao_estoque` (insert), `pedido_compra_item.quantidade_recebida`, `pedido_compra.status`, `chave_idempotencia` (insert)
- Não altera: `saldo_reservado`
- Tudo em uma transação — ou todas as linhas entram, ou nenhuma entra.

**Saída da API**

```json
{
  "entradaId": "2c7e5b91-4d38-4a67-b052-8f13e9a4d760",
  "documentoOrigem": "NF-88421",
  "registradoEm": "2026-08-12T15:02:00-03:00",
  "registradoPor": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29",
  "itens": [
    {
      "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "codigo": "PC-0142",
      "quantidade": 10,
      "saldoFisicoAnterior": 6,
      "saldoFisicoAtual": 16,
      "saldoDisponivel": 12
    },
    {
      "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
      "codigo": "IN-0031",
      "quantidade": 20.0,
      "saldoFisicoAnterior": 27.5,
      "saldoFisicoAtual": 47.5,
      "saldoDisponivel": 47.5
    }
  ],
  "pedidoCompra": { "id": "6b2e9f47-3a15-4c80-9d62-7e10b4f83a95", "numero": "2026/0117", "status": "CONCLUIDO" }
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Entrada registrada |
| `200` | Requisição repetida com a mesma `Idempotency-Key` — retorna a resposta original |
| `400` | Body inválido; quantidade menor ou igual a zero; item repetido no payload |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `estoque:movimentar` |
| `404` | Item ou pedido de compra não encontrado |
| `409` | `documentoOrigem` já registrado |
| `422` | Item inativo; quantidade recebida maior que a pedida sem `confirmarDivergencia` |

**Dependências**

- `ItemEstoqueRepository`
- `MovimentacaoEstoqueRepository`
- `PedidoCompraRepository`
- Serviço de idempotência
- Publicador de eventos de domínio
- Caso de uso Solicitar Compra (origem do `pedidoCompraId`)

**Testes**

*Unitários*

- Soma correta do saldo por linha.
- Rejeita quantidade zero e negativa.
- Detecta `itemId` duplicado no payload.
- Cálculo de status do pedido: parcial versus concluído.

*Integração*

- Entrada de 2 itens atualiza os dois saldos e cria 2 movimentações.
- Mesma `Idempotency-Key` duas vezes: o saldo sobe uma vez só.
- `documentoOrigem` repetido retorna `409`.
- Item inativo retorna `422` e nenhum saldo é alterado (rollback).
- Recebimento parcial mantém o pedido em `PARCIAL`.
- Recebimento total move o pedido para `CONCLUIDO`.

*Concorrência*

- Duas entradas simultâneas do mesmo item somam corretamente, sem perda de atualização.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `registrarEntrada()` na entidade `ItemEstoque` somando o saldo físico
- [ ] Implementar a invariante de quantidade maior que zero
- [ ] Implementar a entidade `MovimentacaoEstoque` do tipo `ENTRADA` como histórico imutável
- [ ] Implementar o recálculo de status do `PedidoCompra` (`ABERTO`, `PARCIAL`, `CONCLUIDO`)

**Caso de uso**

- [ ] Implementar `RegistrarEntradaEstoque` com múltiplos itens
- [ ] Implementar a regra de divergência: quantidade recebida maior que a pedida exige `confirmarDivergencia`

**Repositório**

- [ ] Implementar `MovimentacaoEstoqueRepository`
- [ ] Atualizar `quantidadeRecebida` em `PedidoCompraItem`
- [ ] Criar constraint de unicidade de `documentoOrigem` em `movimentacao_estoque`

**Handler HTTP**

- [ ] Implementar `POST /estoque/entradas`

**Validações**

- [ ] Validar `documentoOrigem` obrigatório
- [ ] Validar `itens` não vazio e sem `itemId` repetido
- [ ] Validar `quantidade` maior que zero em cada linha
- [ ] Validar `custoUnitario` maior que zero
- [ ] Validar que todos os itens existem e estão ativos

**Transação e idempotência**

- [ ] Executar a operação inteira em uma única transação (todas as linhas entram ou nenhuma entra)
- [ ] Implementar a tabela `chave_idempotencia` e o fluxo do header `Idempotency-Key`
- [ ] Gravar a chave de idempotência dentro da mesma transação da operação

**Eventos**

- [ ] Publicar `EntradaRegistrada`

**Testes unitários**

- [ ] Soma correta do saldo por linha
- [ ] Rejeição de quantidade zero e negativa
- [ ] Detecção de `itemId` duplicado no payload
- [ ] Cálculo de status do pedido: parcial versus concluído

**Testes de integração**

- [ ] Entrada com 2 itens atualizando os dois saldos e criando 2 movimentações
- [ ] Mesma `Idempotency-Key` duas vezes somando o saldo uma única vez
- [ ] `documentoOrigem` repetido retornando `409`
- [ ] Item inativo retornando `422` com rollback total dos saldos
- [ ] Recebimento parcial mantendo o pedido em `PARCIAL`
- [ ] Recebimento total movendo o pedido para `CONCLUIDO`

**Testes de concorrência**

- [ ] Duas entradas simultâneas do mesmo item somando corretamente, sem perda de atualização

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, incluindo o header `Idempotency-Key`

**Review**

- [ ] Code Review aprovado

---
