---
documento: Resumo do Contexto — Peças
dono: A definir
versao: 0.4
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Peças

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. O que ainda não está resolvido fica em
[pontos-em-aberto.md](pontos-em-aberto.md).

## O que este contexto cobre

O catálogo de **peças**, seus saldos de estoque, a reserva para Ordens de Serviço, o ciclo de
compra e o retorno ao estoque quando o serviço não acontece. É o contexto que responde "tem essa
peça na prateleira?" e "está garantida para esta OS?".

Peça é o item que o cliente paga na nota: tem `precoVenda`, quantidade **sempre inteira** e
fabricante. Insumo é outro contexto, em [insumos/](../insumos/) — o que se consome no serviço, em
quantidade fracionária e sem preço de venda.

Compras faz parte deste contexto: `pedido_compra` é compartilhado com Insumos, mas a solicitação
de compra de peça é refinada aqui.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 1 | Cadastrar Peça | `POST /estoque/pecas` | `estoque:escrever` | [cadastrar-peca.md](cadastrar-peca.md) |
| 2 | Consultar Peças | `GET /estoque/pecas` | `estoque:ler` | [consultar-pecas.md](consultar-pecas.md) |
| 3 | Atualizar Peça | `PUT /estoque/pecas/{pecaId}` | `estoque:escrever` | [atualizar-peca.md](atualizar-peca.md) |
| 4 | Deletar Peça | `DELETE /estoque/pecas/{pecaId}` | `estoque:escrever` | [deletar-peca.md](deletar-peca.md) |
| 5 | Reservar Peça para OS | sem endpoint, serviço de domínio chamado pelo processamento | — | [reservar-peca-para-os.md](reservar-peca-para-os.md) |
| 6 | Processar Peças para Reserva e Compra | `POST /estoque/solicitacoes-compra-reserva` | `estoque:movimentar` | [processar-pecas-para-reserva-e-compra.md](processar-pecas-para-reserva-e-compra.md) |
| 7 | Solicitar Compra de Peças | `POST /compras/pedidos` e `DELETE /compras/pedidos/{pedidoId}` | `compras:escrever` | [solicitar-compra-de-pecas.md](solicitar-compra-de-pecas.md) |
| 8 | Registrar Entrada de Peças | `POST /estoque/entradas` (compartilhada com Insumos) | `estoque:movimentar` | [registrar-entrada-de-pecas.md](registrar-entrada-de-pecas.md) |
| 9 | Retornar Peça ao Estoque | sem endpoint, serviço de domínio | — | [retornar-peca-ao-estoque.md](retornar-peca-ao-estoque.md) |
| 10 | Cadastrar Fornecedor | `POST /fornecedores` | `compras:escrever` | [cadastrar-fornecedor.md](cadastrar-fornecedor.md) |
| 11 | Consultar Fornecedores | `GET /fornecedores` e `GET /fornecedores/{fornecedorId}` | `compras:ler` | [consultar-fornecedores.md](consultar-fornecedores.md) |
| 12 | Atualizar Fornecedor | `PUT /fornecedores/{fornecedorId}` | `compras:escrever` | [atualizar-fornecedor.md](atualizar-fornecedor.md) |
| 13 | Desativar e Reativar Fornecedor | `DELETE /fornecedores/{fornecedorId}` e `POST /fornecedores/{fornecedorId}/reativacao` | `compras:escrever` | [desativar-fornecedor.md](desativar-fornecedor.md) |
| 14 | Registrar Consumo e Saída de Peças | `POST /estoque/saidas` (compartilhada com Insumos) | `estoque:movimentar` | [registrar-consumo-e-saida-de-pecas.md](registrar-consumo-e-saida-de-pecas.md) |

Os IDs de requisito deste contexto usam o prefixo **`RF-PEC`** e **`RNF-PEC`**, sem repetição. As
tarefas 1 a 9 usam a faixa `01` a `87`; as tarefas de fornecedor e de baixa usam de `88`
em diante.

## Tipos do contexto

**Peça** — `item_estoque` com `tipo = PECA`

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, usado em rotas e vínculos. |
| `codigo` | string | Identificador funcional gerado pelo sistema, formato `PEC-000001`, usado na busca. |
| `tipo` | enum | Sempre `PECA` neste contexto. |
| `nome` | string | Obrigatório; termo curto usado na busca. |
| `descricao` | string | Obrigatória; detalhamento que sai no orçamento. |
| `descricaoNormalizada` | string | Derivada da descrição; única na categoria entre peças ativas. |
| `categoria` | string | Obrigatória; entra na regra de duplicidade. Texto livre hoje. |
| `fabricante` | string | Específico da peça. |
| `precoVenda` | decimal | Valor cobrado do cliente. Insumo não tem. |
| `estoqueMinimo` | int | Inteiro. |
| `saldoFisico` | int | O que está na prateleira. |
| `saldoReservado` | int | Comprometido com OS. |
| `saldoDisponivel` | int | Calculado: físico menos reservado. |
| `ativo` | boolean | `false` após a desativação. |
| `version` | int | Controle otimista, usado com `If-Match`. |

**Reserva de estoque**

`ATIVA` → `LIBERADA` ou `CONSUMIDA`. Vincula peça, quantidade, OS e, quando vem de compra, o pedido.

**Movimentação de estoque** — histórico imutável

`ENTRADA`, `SAIDA`, `RESERVA`, `LIBERACAO_RESERVA`, `ENTRADA_RETORNO`.

**Fornecedor** — agregado de Compras, cujo dono é este contexto

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico. |
| `razaoSocial` | string | Obrigatória. |
| `nomeFantasia` | string | Opcional. |
| `documento` | string | CNPJ ou CPF, somente dígitos. **Imutável** após o cadastro; único entre ativos. |
| `tipoDocumento` | enum | `CNPJ` \| `CPF`. |
| `telefone` / `email` | string | Ao menos um obrigatório. |
| `ativo` | boolean | `false` após a inativação. |
| `version` | int | Controle otimista. |

**Pedido de compra** — compartilhado com Insumos

`ABERTO` → `PARCIAL` → `CONCLUIDO`, ou `CANCELADO`. Itens com `quantidadeNecessaria`,
`quantidadePedida`, `quantidadeReservada` e `quantidadeRecebida`.

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; recursos sob `/estoque/pecas` e `/compras`.
- `id` é UUID e serve para referência; `codigo` é o identificador de negócio e serve para busca.
  Nenhuma referência entre recursos usa `codigo`.
- Autenticação `Bearer <JWT>`; perfil `MECANICO`; escopos `estoque:ler`, `estoque:escrever`,
  `estoque:movimentar` e `compras:escrever`.
- Exclusão é **lógica**, por `DELETE`, preservando o histórico das Ordens de Serviço.
- Atualização de cadastro usa **controle otimista** com header `If-Match` e campo `version`.
- Operação que movimenta saldo é **transacional** e usa `SELECT ... FOR UPDATE` ordenado por
  `item_id`, para evitar deadlock.
- Operação que movimenta saldo aceita `Idempotency-Key`, obrigatório nas reservas.
- Quantidade de peça é **sempre inteira**. Fração é característica de insumo.
- O `codigo` é gerado pelo sistema, em sequência global sem reset: `PEC-000001`.
- Duplicidade por **descrição normalizada dentro da categoria**, entre peças ativas, por índice
  parcial. O cadastro não aceita estoque inicial.
- `ativo` não é aceito no `PUT`: a inativação acontece só pelo `DELETE`, que bloqueia com saldo
  reservado ou com a peça em orçamento `CRIADO`.
- `Idempotency-Key` é **obrigatório** em toda operação que movimenta saldo, inclusive na entrada.
- O projeto **não usa eventos nem mensageria**: integrações entre contextos são chamadas diretas
  na mesma transação.
- Compras e recebimento são **compartilhados com Insumos**: `POST /compras/pedidos` e
  `POST /estoque/entradas` atendem os dois tipos, e este contexto é o **dono** do agregado de
  Compras.
- O fornecedor é **obrigatório** no pedido de compra, e a quantidade comprada pode ser maior que a
  necessidade apurada.
- A **reserva não tem rota pública**: ela é serviço de domínio, chamado pelo processamento, que por
  sua vez é chamado pela aprovação do orçamento.
- A **baixa consome a reserva**, nunca o saldo livre, e devolve ao saldo livre o que foi reservado
  e não usado. A rota `POST /estoque/saidas` é compartilhada com Insumos.
- O cadastro de fornecedor vive aqui: unicidade de documento entre ativos, documento imutável, e
  desativação bloqueada com pedido de compra em aberto.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409` e `412`.

## O que este contexto não faz

- Não trata insumo: cadastro, consulta, reserva e devolução de insumo estão em
  [insumos/](../insumos/). Compra e recebimento são compartilhados.
- Não decide o preço cobrado do cliente: apenas fornece o `precoVenda` vigente da peça.
- Não altera Ordem de Serviço por conta própria, exceto na entrada de estoque, que libera as OS
  sem itens pendentes — ponto ainda em discussão.
- Não expõe rota de reserva direta: reservar é consequência da aprovação do orçamento.
- Não tem consulta de peças faltantes: a apuração da falta acontece dentro do processamento da
  aprovação do orçamento, que reserva o disponível e abre pedido do restante.
