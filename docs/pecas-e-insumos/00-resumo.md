---
documento: Resumo do Contexto — Peças & Insumos
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Peças & Insumos

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. O que ainda não está resolvido fica em
[`pontos-em-aberto.md`](pontos-em-aberto.md).

## O que este contexto cobre

O catálogo de peças e insumos, os saldos de estoque, a reserva para Ordens de Serviço, o ciclo de
compra e o retorno de itens quando o serviço não acontece. É o contexto que responde "tem na
prateleira?" e "está garantido para esta OS?".

Compras faz parte deste contexto: `pedido_compra` pertence a Peças & Insumos.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 10 | Cadastrar Peça | `POST /estoque/pecas` | `estoque:escrever` | [cadastrar-peca.md](cadastrar-peca.md) |
| 2 | Atualizar Peça | `PUT /estoque/pecas/{pecaId}` | `estoque:escrever` | [atualizar-peca.md](atualizar-peca.md) |
| 11 | Deletar Peça | `DELETE /estoque/pecas/{pecaId}` | `estoque:escrever` | [deletar-peca.md](deletar-peca.md) |
| 11 | Consultar Peças | `GET /estoque/pecas` | escopo ausente no documento | [consultar-estoque.md](consultar-estoque.md) |
| 12 | Cadastrar Insumo | `POST /estoque/insumos` | `estoque:escrever` | [cadastrar-insumo.md](cadastrar-insumo.md) |
| 3 | Atualizar Insumo | `PUT /estoque/insumos/{insumoId}` | `estoque:escrever` | [atualizar-insumo.md](atualizar-insumo.md) |
| 13 | Deletar Insumo | `DELETE /estoque/insumos/{insumoId}` | `estoque:escrever` | [deletar-insumo.md](deletar-insumo.md) |
| 14 | Consultar Insumos | `GET /estoque/insumos` e `GET /estoque/insumos/{insumoId}` | `estoque:ler` | [consultar-insumos.md](consultar-insumos.md) |
| 10 | Reservar Peça para OS | `POST /estoque/reservas` | `estoque:movimentar` | [reservar-peca-para-os.md](reservar-peca-para-os.md) |
| 11 | Reservar Insumo para OS | `POST /estoque/reservas-insumos` | `estoque:movimentar` | [reservar-insumo-para-os.md](reservar-insumo-para-os.md) |
| 12 | Processar Peças para Reserva e Compra | `POST /estoque/solicitacoes-compra-reserva` | `estoque:movimentar` | [processar-pecas-para-reserva-e-compra.md](processar-pecas-para-reserva-e-compra.md) |
| 13 | Processar Insumos para Reserva e Compra | `POST /estoque/solicitacoes-compra-reserva-insumos` | `estoque:movimentar` | [processar-insumos-para-reserva-e-compra.md](processar-insumos-para-reserva-e-compra.md) |
| 8 | Solicitar Compra de Peças | `POST /compras/pedidos` e `DELETE /compras/pedidos/{pedidoId}` | `compras:escrever` | [solicitar-compra-de-pecas.md](solicitar-compra-de-pecas.md) |
| 9 | Solicitar Compra de Insumos | `POST /compras/pedidos` e `DELETE /compras/pedidos/{pedidoId}` | `compras:escrever` | [solicitar-compra-de-insumos.md](solicitar-compra-de-insumos.md) |
| 4 | Registrar Entrada de Estoque | `POST /estoque/entradas` | `estoque:movimentar` | [registrar-entrada-de-estoque.md](registrar-entrada-de-estoque.md) |
| 14 | Retornar Peça e Insumo ao Estoque | sem endpoint, serviço de domínio | — | [retornar-peca-e-insumo-ao-estoque.md](retornar-peca-e-insumo-ao-estoque.md) |

> A numeração está duplicada em cinco pares (10, 11, 12, 13 e 14) — ver ponto 1 de
> [`pontos-em-aberto.md`](pontos-em-aberto.md).

## Tipos do contexto

**Item de estoque** — peça e insumo compartilham a mesma entidade, diferenciados por `tipo`

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, usado em rotas e vínculos. |
| `codigo` | string | Identificador funcional (`PC-0142`, `IN-0031`, `PEC-000001`), usado na busca. |
| `tipo` | enum | `PECA` \| `INSUMO`. |
| `descricao` | string | Obrigatória. |
| `nome` | string | Presente apenas nos cadastros — ver ponto 4. |
| `categoria` | string | Texto livre hoje. |
| `fabricante` | string | Apenas para peça. |
| `unidadeMedida` | enum | `UN` \| `L` \| `ML` \| `KG` \| `G` \| `M`. |
| `precoVenda` | decimal | Peça: valor cobrado do cliente. |
| `custoUnitario` | decimal | Insumo: custo diluído no serviço. |
| `estoqueMinimo` | número | Inteiro na peça, decimal no insumo. |
| `saldoFisico` | número | O que está na prateleira. |
| `saldoReservado` | número | Comprometido com OS. |
| `saldoDisponivel` | número | Calculado: físico menos reservado. |
| `ativo` | boolean | `false` após a desativação. |
| `version` | int | Controle otimista, usado com `If-Match`. |

**Reserva de estoque**

`ATIVA` → `LIBERADA` ou `CONSUMIDA`. Vincula item, quantidade, OS e, quando vem de compra, o pedido.

**Movimentação de estoque** — histórico imutável

`ENTRADA`, `SAIDA`, `RESERVA`, `LIBERACAO_RESERVA`, `ENTRADA_RETORNO`.

**Pedido de compra**

`ABERTO` → `PARCIAL` → `CONCLUIDO`, ou `CANCELADO`. Itens com `quantidadeNecessaria`,
`quantidadePedida`, `quantidadeReservada` e `quantidadeRecebida`.

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; recursos sob `/estoque` e `/compras`.
- `id` é UUID e serve para referência; `codigo` é o identificador de negócio e serve para busca.
  Nenhuma referência entre recursos usa `codigo`.
- Autenticação `Bearer <JWT>`; perfis `MECANICO` e `GESTOR`; escopos `estoque:ler`,
  `estoque:escrever`, `estoque:movimentar` e `compras:escrever`.
- Exclusão é **lógica**, por `DELETE`, preservando o histórico das Ordens de Serviço.
- Atualização de cadastro usa **controle otimista** com header `If-Match` e campo `version`.
- Operação que movimenta saldo é **transacional** e usa `SELECT ... FOR UPDATE` ordenado por
  `item_id`, para evitar deadlock.
- Operação que movimenta saldo aceita `Idempotency-Key`, obrigatório nas reservas.
- Insumo aceita **quantidade fracionada**; peça, não.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409`, `412` e `422`.

## O que este contexto não faz

- Não decide o preço cobrado do cliente: apenas fornece o valor vigente do item.
- Não altera Ordem de Serviço por conta própria, exceto na entrada de estoque, que libera as OS
  sem itens pendentes — ponto ainda em discussão.
- Não cobre a consulta consolidada de itens faltantes nem a baixa de consumo na execução: as duas
  tarefas foram removidas para reescrita.
