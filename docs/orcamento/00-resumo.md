---
documento: Resumo do Contexto — Orçamento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Orçamento

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. O que ainda não está resolvido fica em
[`pontos-em-aberto.md`](pontos-em-aberto.md).

## O que este contexto cobre

O orçamento da Ordem de Serviço: o cálculo dos valores, a apresentação ao cliente e a decisão dele
— aprovar ou recusar. É o ponto do fluxo em que o cliente autoriza, ou não, a execução.

O orçamento **não** é criado aqui: ele nasce no contexto de Ordem de Serviço, no registro do
problema encontrado. Este contexto cuida do que acontece depois.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 1 | Calcular Orçamento | `POST /orcamentos/{orcamentoId}/calcular` | `orcamentos:escrever` | [calcular-orcamento.md](calcular-orcamento.md) |
| 2 | Consultar Orçamento | `GET /orcamentos` | `orcamentos:ler` | [consultar-orcamento.md](consultar-orcamento.md) |
| 3 | Aprovar Orçamento | `POST /orcamentos/{orcamentoId}/aprovar` | `orcamentos:aprovar` | [aprovar-orcamento.md](aprovar-orcamento.md) |
| 4 | Recusar Orçamento | `POST /orcamentos/{orcamentoId}/recusar` | `orcamentos:recusar` | [recusar-orcamento.md](recusar-orcamento.md) |

## Tipos do contexto

**Orçamento**

| Campo | Tipo | Observação |
|---|---|---|
| `orcamentoId` | uuid | Identificador do orçamento. |
| `ordemServicoId` | uuid | OS à qual o orçamento pertence. |
| `tipo` | enum | `PRINCIPAL` \| `COMPLEMENTAR`. |
| `orcamentoOriginalId` | uuid | Preenchido apenas quando o tipo é `COMPLEMENTAR`. |
| `status` | enum | `CRIADO` \| `APROVADO` \| `RECUSADO`. |
| `valorTotal` | decimal | Soma dos itens do orçamento. |
| `valorTotalGeral` | decimal | Soma do principal com os complementares da mesma OS. |
| `estimativaEntregaDias` | int | Calculada no cálculo do orçamento. |
| `dataAprovacao` / `dataRecusa` | datetime | Registradas na decisão do cliente. |
| `motivoRecusa` | string | Opcional, informado pelo cliente. |

**Item de Orçamento**

| Campo | Tipo | Observação |
|---|---|---|
| `tipo` | enum | `SERVICO` \| `PECA` \| `INSUMO`. |
| `itemId` | uuid | Referência ao serviço, peça ou insumo. |
| `descricao` | string | Cópia da descrição no momento do registro. |
| `quantidade` | número | Inteiro para peça e serviço; decimal para insumo. |
| `valorUnitario` | decimal | **Congelado** no momento do registro. |
| `valorTotal` | decimal | `quantidade × valorUnitario`. |

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; ações de decisão como sub-recurso com verbo: `/aprovar`, `/recusar`,
  `/calcular`.
- O ator de aprovar, recusar e consultar é o **cliente**, não o funcionário da oficina.
- O valor unitário é sempre copiado para o orçamento, para que alteração de preço no catálogo não
  mude orçamento já emitido.
- Aprovar leva a OS para `AGUARDANDO_EXECUCAO`; recusar leva a OS para `CANCELADA`.
- Um orçamento só aceita decisão uma vez: `CRIADO` é o único status que aceita aprovação ou recusa.
- Códigos de erro usados: `400`, `401`, `403`, `404` e `409`.

## O que este contexto não faz

- Não cria o orçamento nem adiciona itens: isso acontece no contexto de Ordem de Serviço.
- Não envia o orçamento ao cliente — a tarefa existia e foi removida para reescrita.
- Não trata aprovação e recusa do orçamento **complementar** como tarefas próprias.
- Não movimenta estoque: a reserva das peças é do contexto de Peças & Insumos.
