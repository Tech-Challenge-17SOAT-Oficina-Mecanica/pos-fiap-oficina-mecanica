---
documento: Refinamento de Requisitos — Registrar Entrada de Insumos
dono: A definir
versao: 0.4
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Registrar Entrada de Insumos

Este documento detalha a tarefa Registrar Entrada de Insumos do contexto de Insumos.

> **Escopo deste documento.** O recebimento é **uma rota só**, compartilhada com peças. Este
> documento descreve a parte de **insumos**; a de peças está em [registrar-entrada-de-pecas.md](../pecas/registrar-entrada-de-pecas.md).
> As duas compartilham a rota, o recurso `pedido_compra` e a mesma regra de liberação de OS — o
> que muda são as validações do item.

## 8 · Registrar Entrada de Insumos

### 8.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar o recebimento de insumos, aumentando o saldo físico do estoque e liberando para execução as
Ordens de Serviço que estavam paradas esperando esses itens.

**Problema**

Sem registro de entrada, o saldo do sistema diverge do saldo real da prateleira. A consequência é
dupla: o mecânico deixa de reservar peça que existe, ou reserva peça que já acabou — e a oficina
só descobre no momento de executar o serviço. Pior ainda, a OS que estava parada esperando aquela
peça continua marcada como aguardando recursos mesmo depois de a peça chegar, e ninguém percebe
que o serviço já pode começar.

**Pré-condições**

- O insumo recebido deve estar cadastrado e ativo.
- Deve existir um documento de origem do recebimento, nota fiscal ou pedido de compra.
- O usuário deve estar autorizado a movimentar estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-INS-97 | Atualizar o custo unitário do insumo com o custo da entrada mais recente. |
| RF-INS-66 | Permitir registrar entrada informando item, quantidade, custo unitário e documento de origem. |
| RF-INS-67 | Permitir registrar a entrada de vários itens em um mesmo recebimento. |
| RF-INS-68 | Validar que a quantidade informada é maior que zero. |
| RF-INS-69 | Vincular a entrada ao pedido de compra correspondente, quando houver. |
| RF-INS-70 | Atualizar o saldo físico do item. |
| RF-INS-71 | Efetivar as reservas vinculadas ao pedido de compra recebido. |
| RF-INS-72 | Registrar a movimentação no histórico de estoque. |
| RF-INS-73 | Atualizar a situação do pedido de compra conforme o recebimento seja parcial ou total. |
| RF-INS-74 | Identificar as Ordens de Serviço vinculadas ao pedido recebido. |
| RF-INS-75 | Alterar o status dessas OS de `AGUARDANDO_RECURSOS` para `AGUARDANDO_EXECUCAO` quando todos os seus itens pendentes estiverem atendidos. |
| RF-INS-76 | Manter em `AGUARDANDO_RECURSOS` as OS que ainda possuem itens pendentes. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-INS-43 | A operação deve ser feita por API RESTful. |
| RNF-INS-44 | A operação deve ser acessível somente por usuário autorizado com permissão de estoque. |
| RNF-INS-45 | A entrada deve ser transacional — ou todos os itens do recebimento entram, ou nenhum entra. |
| RNF-INS-46 | A mudança de status das OS deve ocorrer na mesma operação da entrada. |
| RNF-INS-47 | A operação deve ser idempotente em relação ao documento de origem, para impedir dupla contagem em caso de reenvio. |
| RNF-INS-48 | A movimentação deve ser auditável e o histórico imutável. |
| RNF-INS-49 | A operação deve ser protegida contra concorrência, sem perda de atualização de saldo. |

**Fluxo Principal**

1. O mecânico informa o documento de origem do recebimento.
2. O mecânico informa os itens, as quantidades e os custos unitários.
3. O sistema valida os itens e as quantidades.
4. O sistema atualiza o saldo físico de cada item.
5. O sistema efetiva as reservas vinculadas ao pedido de compra recebido.
6. O sistema registra a movimentação de entrada no histórico.
7. O sistema atualiza a situação do pedido de compra vinculado.
8. O sistema identifica as OS vinculadas ao pedido.
9. O sistema verifica, para cada OS, se ainda restam itens pendentes.
10. O sistema altera para `AGUARDANDO_EXECUCAO` as OS sem itens pendentes.
11. O sistema confirma a entrada e devolve os saldos atualizados e as OS liberadas.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Item não encontrado ou inativo | Impede a entrada e informa qual item está irregular. |
| A2 | Quantidade menor ou igual a zero | Impede a entrada. |
| A3 | Documento de origem já registrado | Informa que o recebimento já foi lançado e não altera o saldo. |
| A4 | Recebimento parcial | Registra a quantidade recebida, mantém o pedido em aberto com o saldo pendente e mantém as OS em `AGUARDANDO_RECURSOS`. |
| A5 | OS com itens de mais de um pedido | A OS só passa para `AGUARDANDO_EXECUCAO` quando todos os itens pendentes forem atendidos. |
| A6 | OS em status diferente de `AGUARDANDO_RECURSOS` | Não altera o status dessa OS. |
| A7 | Entrada sem pedido de compra vinculado | Registra a entrada e não altera nenhuma OS. |
| A8 | Quantidade recebida maior que a pedida | Alerta a divergência e exige confirmação antes de gravar. |
| A9 | Usuário sem autorização | Impede a operação. |

**Saída**

- Confirmação da entrada com o saldo físico atualizado de cada item.
- Situação atualizada do pedido de compra vinculado.
- Relação das OS liberadas para execução e das que seguem aguardando recursos.
- Ou indicação do motivo pelo qual a entrada foi recusada.

**Pós-condições**

- O saldo físico dos itens recebidos está acrescido da quantidade informada.
- As reservas vinculadas ao pedido recebido estão efetivadas sobre o saldo físico.
- A movimentação de entrada está registrada no histórico.
- O pedido de compra vinculado está atualizado, concluído ou parcialmente atendido.
- As OS sem itens pendentes estão em `AGUARDANDO_EXECUCAO` e as que ainda dependem de itens
  continuam em `AGUARDANDO_RECURSOS`.

---

### 8.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/entradas
```

> **Decisão de projeto.** `Idempotency-Key` é **obrigatório**, e não recomendado. A entrada é a
> operação em que a repetição faz mais estrago: reenviar a mesma nota soma o saldo duas vezes.
> Vale para todas as operações que movimentam saldo (D-02).

> **Decisão de projeto.** O recebimento tem **uma rota só**, `POST /estoque/entradas`, para peça
> e insumo. Chegou a existir uma rota por tipo, criada junto com a divisão dos contextos, mas ela
> foi descartada: a nota fiscal do fornecedor costuma trazer os dois tipos, e obrigar duas
> chamadas para o mesmo recebimento quebraria a transação e a idempotência por
> `documentoOrigem`. Este documento descreve a parte de **insumos** do recebimento; a de peças está em
> [registrar-entrada-de-pecas.md](../pecas/registrar-entrada-de-pecas.md).

> **Decisão de projeto.** A entrada não é só um lançamento de saldo: ela é o gatilho que
> **destrava as Ordens de Serviço**. Por isso a mudança de status das OS acontece na mesma
> transação do recebimento, e não por um processo separado. A alternativa — atualizar as OS depois,
> por evento — deixaria uma janela em que a peça já chegou e a OS ainda aparece parada.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `estoque:movimentar`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | **Obrigatório**; impede dupla contagem em caso de reenvio. |
| Body | `documentoOrigem` | string | Obrigatório; nota fiscal ou documento do recebimento. |
| Body | `fornecedorId` | uuid | Fornecedor do recebimento. |
| Body | `pedidoCompraId` | uuid | Pedido de compra vinculado, quando houver. |
| Body | `confirmarDivergencia` | boolean | Obrigatório quando a quantidade recebida excede a pedida. |
| Body | `itens[]` | array | Obrigatório, não vazio, máximo 200 linhas. |
| Body | `itens[].itemId` | uuid | Item recebido; não pode repetir no mesmo payload. |
| Body | `itens[].quantidade` | decimal | Maior que zero; casas decimais conforme a `unidadeMedida` quando o item for insumo. |
| Body | `itens[].custoUnitario` | decimal | Maior que zero. |

```json
{
  "documentoOrigem": "NF-88421",
  "fornecedorId": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
  "pedidoCompraId": "6b2e9f47-3a15-4c80-9d62-7e10b4f83a95",
  "confirmarDivergencia": false,
  "itens": [
    {
      "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "quantidade": 60.0,
      "custoUnitario": 22.75
    },
    {
      "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
      "quantidade": 20.0,
      "custoUnitario": 30.10
    }
  ]
}
```

**Validações**

*Técnicas*

- `documentoOrigem` obrigatório.
- `itens` não vazio, com no máximo 200 linhas.
- `quantidade` maior que zero, com as casas decimais compatíveis com a `unidadeMedida` nas linhas do tipo `INSUMO`.
- `custoUnitario` maior que zero.

> **Decisão de projeto — D-14.** O custo do insumo é o **último custo de entrada**: cada
> recebimento sobrescreve `custo_unitario` com o valor daquela nota. Média ponderada fica para
> quando existir histórico que justifique o cálculo.
- Sem `itemId` repetido no mesmo payload.

*Negócio*

- Todos os `itemId` existem e estão ativos. A rota aceita peça e insumo no mesmo recebimento; as validações abaixo tratam das linhas do tipo `INSUMO`.
- `documentoOrigem` ainda não registrado — chave única em `movimentacao_estoque`.
- Quando há `pedidoCompraId`, os itens devem pertencer ao pedido.
- Quantidade recebida maior que a pedida exige `confirmarDivergencia: true` no body.
- Só mudam de status as OS vinculadas ao pedido que estejam em `AGUARDANDO_RECURSOS`.
- A OS só passa para `AGUARDANDO_EXECUCAO` quando não restar nenhum item pendente nela.

**Processamento**

1. Verificar o `Idempotency-Key`; se já processada, retornar a resposta original.
2. Abrir transação.
3. Validar o payload e carregar todos os itens com `SELECT ... FOR UPDATE`, ordenados por `item_id`.
4. Verificar duplicidade de `documentoOrigem`.
5. Havendo `pedidoCompraId`, carregar o pedido e conferir divergência de quantidade.
6. Para cada linha: `saldo_fisico += quantidade` e `custo_unitario = custoUnitario recebido`.
7. Efetivar as reservas vinculadas ao pedido: `saldo_reservado += quantidade reservada recebida`.
8. Inserir uma `movimentacao_estoque` do tipo `ENTRADA` por linha.
9. Atualizar `quantidade_recebida` em `pedido_compra_item`.
10. Recalcular o status do pedido: `ABERTO` para `PARCIAL` ou `CONCLUIDO`.
11. Carregar as OS vinculadas ao pedido que estejam em `AGUARDANDO_RECURSOS`.
12. Para cada OS, verificar se ainda existem itens pendentes de recebimento.
13. Alterar para `AGUARDANDO_EXECUCAO` as OS sem itens pendentes.
14. Commit.
15. Registrar a entrada na trilha de auditoria.

**Persistência**

- Consulta: `item_estoque`, `pedido_compra`, `pedido_compra_item`, `reserva_estoque`,
  `ordem_servico`, `chave_idempotencia`.
- Altera: `item_estoque.saldo_fisico`, `item_estoque.saldo_reservado` (efetivação das reservas),
  `movimentacao_estoque` (insert), `pedido_compra_item.quantidade_recebida`,
  `pedido_compra.status`, `ordem_servico.status`, `chave_idempotencia` (insert).
- Não altera: orçamento da OS.
- Tudo em uma transação — ou todas as linhas entram e as OS são atualizadas, ou nada é gravado.

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
      "codigo": "INS-000077",
      "unidadeMedida": "L",
      "quantidade": 60.0,
      "saldoFisicoAnterior": 12.5,
      "saldoFisicoAtual": 72.5,
      "saldoReservado": 60.0,
      "saldoDisponivel": 12.5
    },
    {
      "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
      "codigo": "INS-000031",
      "unidadeMedida": "L",
      "quantidade": 20.0,
      "saldoFisicoAnterior": 27.5,
      "saldoFisicoAtual": 47.5,
      "saldoReservado": 20.0,
      "saldoDisponivel": 27.5
    }
  ],
  "pedidoCompra": {
    "id": "6b2e9f47-3a15-4c80-9d62-7e10b4f83a95",
    "numero": "2026/0117",
    "status": "CONCLUIDO"
  },
  "ordensServico": [
    {
      "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
      "statusAnterior": "AGUARDANDO_RECURSOS",
      "status": "AGUARDANDO_EXECUCAO",
      "itensPendentes": 0
    },
    {
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "statusAnterior": "AGUARDANDO_RECURSOS",
      "status": "AGUARDANDO_RECURSOS",
      "itensPendentes": 2
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Entrada registrada. |
| `200` | Requisição repetida com a mesma `Idempotency-Key` — retorna a resposta original. |
| `400` | Body inválido; quantidade menor ou igual a zero; item repetido no payload; `Idempotency-Key` ausente. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `estoque:movimentar`. |
| `404` | Item ou pedido de compra não encontrado. |
| `409` | `documentoOrigem` já registrado. |
| `409` | Item inativo; quantidade recebida maior que a pedida sem `confirmarDivergencia`. |

**Dependências**

- `ItemEstoqueRepository`.
- `MovimentacaoEstoqueRepository`.
- `PedidoCompraRepository`.
- `ReservaEstoqueRepository`, para a efetivação das reservas.
- `OrdemDeServicoRepository`, para a verificação de pendências e a mudança de status.
- Serviço de idempotência.
- Trilha de auditoria.
- Caso de uso Solicitar Compra, origem do `pedidoCompraId`.

**Testes**

*Unitários*

- Soma correta do saldo por linha.
- Rejeita quantidade zero e negativa.
- Detecta `itemId` duplicado no payload.
- Cálculo de status do pedido: parcial versus concluído.
- Efetivação da reserva vinculada ao pedido recebido.
- OS sem itens pendentes marcada para `AGUARDANDO_EXECUCAO`.
- OS com itens pendentes mantida em `AGUARDANDO_RECURSOS`.
- OS em status diferente de `AGUARDANDO_RECURSOS` não é alterada.

*Integração*

- Entrada de 2 itens atualiza os dois saldos e cria 2 movimentações.
- Mesma `Idempotency-Key` duas vezes: o saldo sobe uma vez só.
- `documentoOrigem` repetido retorna `409`.
- Item inativo retorna `409` e nenhum saldo é alterado, com rollback total.
- Recebimento parcial mantém o pedido em `PARCIAL` e a OS em `AGUARDANDO_RECURSOS`.
- Recebimento total move o pedido para `CONCLUIDO` e libera a OS para `AGUARDANDO_EXECUCAO`.
- OS com itens de dois pedidos só é liberada após o segundo recebimento.
- Entrada sem `pedidoCompraId` não altera nenhuma OS.
- Saldo reservado atualizado conforme as reservas efetivadas.
- Orçamento da OS permanece inalterado.

*Concorrência*

- Duas entradas simultâneas do mesmo item somam corretamente, sem perda de atualização.
- Entradas concorrentes que liberam a mesma OS não geram dupla transição de status.

---

### 8.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `registrarEntrada()` na entidade `ItemEstoque` somando o saldo físico
- [ ] Implementar a invariante de quantidade maior que zero
- [ ] Implementar a movimentação de estoque do tipo `ENTRADA` como histórico imutável
- [ ] Implementar a efetivação das reservas vinculadas ao pedido recebido
- [ ] Implementar o recálculo de status do pedido de compra: `ABERTO`, `PARCIAL`, `CONCLUIDO`
- [ ] Implementar a regra de liberação da OS: sem itens pendentes, passa para `AGUARDANDO_EXECUCAO`
- [ ] Garantir que a entrada não altera o orçamento da OS

**Caso de uso**

- [ ] Implementar `RegistrarEntradaEstoque` com múltiplos itens
- [ ] Implementar a regra de divergência: quantidade recebida maior que a pedida exige `confirmarDivergencia`
- [ ] Identificar as OS vinculadas ao pedido recebido
- [ ] Verificar, para cada OS, se ainda restam itens pendentes
- [ ] Não alterar OS em status diferente de `AGUARDANDO_RECURSOS`

**Repositório**

- [ ] Implementar `MovimentacaoEstoqueRepository`
- [ ] Atualizar `quantidadeRecebida` em `PedidoCompraItem`
- [ ] Criar constraint de unicidade de `documentoOrigem` em `movimentacao_estoque`
- [ ] Implementar a consulta de itens pendentes por Ordem de Serviço

**Handler HTTP**

- [ ] Implementar `POST /estoque/entradas`, compartilhado com o outro contexto
- [ ] Criar DTO/request de entrada e DTO/response com saldos, pedido e OS afetadas
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar `documentoOrigem` obrigatório
- [ ] Validar `itens` não vazio e sem `itemId` repetido
- [ ] Validar `quantidade` e `custoUnitario` maiores que zero
- [ ] Sobrescrever o custo unitário do insumo com o custo da entrada
- [ ] Validar que todos os itens existem e estão ativos

**Transação e idempotência**

- [ ] Executar a operação inteira em uma única transação, incluindo a mudança de status das OS
- [ ] Implementar a tabela `chave_idempotencia` e o fluxo do header `Idempotency-Key`
- [ ] Gravar a chave de idempotência dentro da mesma transação da operação

**Concorrência**

- [ ] Implementar `SELECT ... FOR UPDATE` ordenado por `item_id`
- [ ] Garantir que entradas concorrentes não gerem dupla transição de status da mesma OS

**Auditoria**

- [ ] Registrar a entrada na trilha de auditoria
- [ ] Atualizar diretamente, na mesma transação, o status das OS liberadas

**Testes unitários**

- [ ] Soma correta do saldo por linha
- [ ] Rejeição de quantidade zero e negativa
- [ ] Detecção de `itemId` duplicado no payload
- [ ] Cálculo de status do pedido: parcial versus concluído
- [ ] Efetivação da reserva do pedido recebido
- [ ] OS sem pendência liberada e OS com pendência mantida
- [ ] OS em outro status não alterada

**Testes de integração**

- [ ] Entrada com 2 itens atualizando saldos e criando 2 movimentações
- [ ] `Idempotency-Key` repetida somando o saldo uma única vez
- [ ] `documentoOrigem` repetido retornando `409`
- [ ] Item inativo retornando `409` com rollback total
- [ ] Recebimento parcial e recebimento total
- [ ] OS com itens de dois pedidos liberada apenas no segundo recebimento
- [ ] Entrada sem pedido não alterando nenhuma OS
- [ ] Orçamento da OS inalterado

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI, incluindo o header `Idempotency-Key`

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite da task

---
