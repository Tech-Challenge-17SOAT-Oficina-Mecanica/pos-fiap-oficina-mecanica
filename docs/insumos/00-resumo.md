---
documento: Resumo do Contexto — Insumos
dono: A definir
versao: 0.4
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Insumos

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. O que ainda não está resolvido fica em
[pontos-em-aberto.md](pontos-em-aberto.md).

## O que este contexto cobre

O catálogo de **insumos**, seus saldos de estoque, a reserva para Ordens de Serviço, o ciclo de
compra e o retorno ao estoque quando o serviço não acontece. Insumo é o que se consome durante o
serviço — óleo, fluido, estopa, produto de limpeza.

Insumo tem `custoUnitario` e **quantidade fracionária**, medida por `unidadeMedida`. Peça é outro
contexto, em [pecas/](../pecas/) — o item que o cliente paga na nota, sempre em quantidade
inteira e com `precoVenda`.

Compras faz parte deste contexto: `pedido_compra` é compartilhado com Peças, mas a solicitação de
compra de insumo é refinada aqui.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 1 | Cadastrar Insumo | `POST /estoque/insumos` | `estoque:escrever` | [cadastrar-insumo.md](cadastrar-insumo.md) |
| 2 | Consultar Insumos | `GET /estoque/insumos` e `GET /estoque/insumos/{insumoId}` | `estoque:ler` | [consultar-insumos.md](consultar-insumos.md) |
| 3 | Atualizar Insumo | `PUT /estoque/insumos/{insumoId}` | `estoque:escrever` | [atualizar-insumo.md](atualizar-insumo.md) |
| 4 | Deletar Insumo | `DELETE /estoque/insumos/{insumoId}` | `estoque:escrever` | [deletar-insumo.md](deletar-insumo.md) |
| 5 | Reservar Insumo para OS | sem endpoint, serviço de domínio chamado pelo processamento | — | [reservar-insumo-para-os.md](reservar-insumo-para-os.md) |
| 6 | Processar Insumos para Reserva e Compra | `POST /estoque/solicitacoes-compra-reserva-insumos` | `estoque:movimentar` | [processar-insumos-para-reserva-e-compra.md](processar-insumos-para-reserva-e-compra.md) |
| 7 | Solicitar Compra de Insumos | `POST /compras/pedidos` e `DELETE /compras/pedidos/{pedidoId}` | `compras:escrever` | [solicitar-compra-de-insumos.md](solicitar-compra-de-insumos.md) |
| 8 | Registrar Entrada de Insumos | `POST /estoque/entradas` (compartilhada com Peças) | `estoque:movimentar` | [registrar-entrada-de-insumos.md](registrar-entrada-de-insumos.md) |
| 9 | Retornar Insumo ao Estoque | sem endpoint, serviço de domínio | — | [retornar-insumo-ao-estoque.md](retornar-insumo-ao-estoque.md) |
| 10 | Registrar Consumo e Saída de Insumos | `POST /estoque/saidas` (compartilhada com Peças) | `estoque:movimentar` | [registrar-consumo-e-saida-de-insumos.md](registrar-consumo-e-saida-de-insumos.md) |

Os IDs de requisito deste contexto usam o prefixo **`RF-INS`** e **`RNF-INS`**, sem repetição. As
tarefas 1 a 9 usam a faixa `01` a `86`; a baixa de consumo usa de `87` em diante.

## Tipos do contexto

**Insumo** — `item_estoque` com `tipo = INSUMO`

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, usado em rotas e vínculos. |
| `codigo` | string | Identificador funcional gerado pelo sistema, formato `INS-000001`, usado na busca. |
| `tipo` | enum | Sempre `INSUMO` neste contexto. |
| `nome` | string | Obrigatório; termo curto usado na busca. |
| `descricao` | string | Obrigatória; detalhamento que sai no orçamento. |
| `descricaoNormalizada` | string | Derivada da descrição; única na categoria e unidade entre insumos ativos. |
| `categoria` | string | Obrigatória; entra na regra de duplicidade. Texto livre hoje. |
| `unidadeMedida` | enum | `UN` \| `L` \| `ML` \| `KG` \| `G` \| `M`. Define as casas decimais aceitas. Cada unidade é um item independente, sem conversão. |
| `custoUnitario` | decimal | Custo diluído no serviço. Insumo não tem preço de venda. |
| `estoqueMinimo` | decimal | Aceita fração. |
| `saldoFisico` | decimal | O que está na prateleira. |
| `saldoReservado` | decimal | Comprometido com OS. |
| `saldoDisponivel` | decimal | Calculado: físico menos reservado. |
| `ativo` | boolean | `false` após a desativação. |
| `version` | int | Controle otimista, usado com `If-Match`. |

**Reserva de estoque**

`ATIVA` → `LIBERADA` ou `CONSUMIDA`. Vincula insumo, quantidade, OS e, quando vem de compra, o
pedido.

**Movimentação de estoque** — histórico imutável

`ENTRADA`, `SAIDA`, `RESERVA`, `LIBERACAO_RESERVA`, `ENTRADA_RETORNO`.

**Pedido de compra** — compartilhado com Peças

`ABERTO` → `PARCIAL` → `CONCLUIDO`, ou `CANCELADO`. Itens com `quantidadeNecessaria`,
`quantidadePedida`, `quantidadeReservada` e `quantidadeRecebida`.

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; recursos sob `/estoque/insumos` e `/compras`.
- `id` é UUID e serve para referência; `codigo` é o identificador de negócio e serve para busca.
  Nenhuma referência entre recursos usa `codigo`.
- Autenticação `Bearer <JWT>`; perfil `MECANICO`; escopos `estoque:ler`, `estoque:escrever`,
  `estoque:movimentar` e `compras:escrever`.
- Exclusão é **lógica**, por `DELETE`, preservando o histórico das Ordens de Serviço.
- Atualização de cadastro usa **controle otimista** com header `If-Match` e campo `version`.
- Operação que movimenta saldo é **transacional** e usa `SELECT ... FOR UPDATE` ordenado por
  `item_id`, para evitar deadlock.
- Operação que movimenta saldo aceita `Idempotency-Key`, obrigatório nas reservas.
- Quantidade de insumo é **fracionária**, com as casas decimais definidas pela `unidadeMedida`. O
  arredondamento segue a precisão da unidade, sempre na mesma direção, em um único ponto do
  domínio.
- Insumo **é reservado** como a peça, e a baixa acontece na execução, sobre a reserva. A regra
  antiga, de baixa direta sem reserva, foi revogada.
- O `codigo` é gerado pelo sistema, em sequência global sem reset: `INS-000001`.
- Duplicidade por **descrição normalizada dentro da categoria e da unidade de medida**, entre
  insumos ativos, por índice parcial. O cadastro não aceita estoque inicial.
- **Não há conversão entre unidades de medida**: o mesmo produto em `L` e em `ML` são dois itens
  de estoque, com saldos próprios.
- `ativo` não é aceito no `PUT`: a inativação acontece só pelo `DELETE`, que bloqueia com saldo
  reservado ou com o insumo em orçamento `CRIADO`.
- O custo efetivo é atualizado pela **entrada de estoque**, com o último custo recebido; o campo do
  cadastro é o custo de referência.
- `Idempotency-Key` é **obrigatório** em toda operação que movimenta saldo, inclusive na entrada.
- O projeto **não usa eventos nem mensageria**: integrações entre contextos são chamadas diretas
  na mesma transação.
- Compras e recebimento são **compartilhados com Peças**: `POST /compras/pedidos` e
  `POST /estoque/entradas` atendem os dois tipos, e o **dono** do agregado de Compras é o contexto
  de Peças.
- O fornecedor é **obrigatório** no pedido de compra, e a quantidade comprada pode ser maior que a
  necessidade apurada — é o que permite comprar embalagem fechada. O cadastro de fornecedor é
  documentado em [pecas/](../pecas/).
- A **reserva não tem rota pública**: ela é serviço de domínio, chamado pelo processamento, que por
  sua vez é chamado pela aprovação do orçamento.
- A **baixa consome a reserva**, nunca o saldo livre, e devolve ao saldo livre o que foi reservado
  e não usado. A rota `POST /estoque/saidas` é compartilhada com Peças.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409` e `412`.

## O que este contexto não faz

- Não trata peça: cadastro, consulta, reserva e devolução de peça estão em [pecas/](../pecas/).
  Compra e recebimento são compartilhados, e o refinamento do fornecedor mora lá.
- Não define preço de venda: insumo entra no orçamento pelo custo, diluído no valor do serviço.
- Não altera Ordem de Serviço por conta própria, exceto na entrada de estoque, que libera as OS
  sem itens pendentes — ponto ainda em discussão.
- Não expõe rota de reserva direta: reservar é consequência da aprovação do orçamento.
- Não tem consulta de insumos faltantes: a apuração da falta acontece dentro do processamento da
  aprovação do orçamento, que reserva o disponível e abre pedido do restante.
