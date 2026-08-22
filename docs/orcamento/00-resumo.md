---
documento: Resumo do Contexto — Orçamento
dono: A definir
versao: 0.3
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
| 3 | Aprovar Orçamento | `POST /orcamentos/{orcamentoId}/aprovar` | `orcamentos:decidir` | [aprovar-orcamento.md](aprovar-orcamento.md) |
| 4 | Recusar Orçamento | `POST /orcamentos/{orcamentoId}/recusar` | `orcamentos:decidir` | [recusar-orcamento.md](recusar-orcamento.md) |

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
- O ator de aprovar, recusar e consultar é o **cliente**, não o funcionário da oficina. Ele se
  autentica por **token de escopo reduzido**, emitido no envio do orçamento e válido apenas para
  aquela OS.
- Aprovar e recusar usam **um escopo só**: `orcamentos:decidir`.
- O complementar é um **orçamento separado**, com identificador próprio, `tipo` e
  `orcamentoOriginalId`. Aprovar e recusar cobrem os dois tipos: não há tarefa separada para o
  complementar.
- O `status` do orçamento é a fonte da verdade da decisão do cliente; o `status` da OS é a fonte da
  verdade da etapa do atendimento.
- O cálculo é acionado duas vezes: ao fim do diagnóstico, antes do envio, e ao fechar cada
  complementar.
- Listagem com envelope `data`, `pagina`, `tamanho`, `totalElementos` e `totalPaginas`.
- O valor unitário é sempre copiado para o orçamento, para que alteração de preço no catálogo não
  mude orçamento já emitido.
- Aprovar dispara o processamento dos itens — reserva o disponível, compra o faltante — e leva a
  OS para `AGUARDANDO_EXECUCAO` ou `AGUARDANDO_RECURSOS`.
- Recusar o principal leva a OS para `CANCELADA` e devolve os itens ao estoque; recusar um
  complementar marca só aquele orçamento como `RECUSADO` e devolve a OS para a fila.
- Um orçamento só aceita decisão uma vez: `CRIADO` é o único status que aceita aprovação ou recusa.
- Códigos de erro usados: `400`, `401`, `403`, `404` e `409`.

## O que este contexto não faz

- Não cria o orçamento nem adiciona itens: isso acontece no contexto de Ordem de Serviço, que
  também é quem abre o complementar.
- Não faz **recusa parcial**: a decisão é sobre o orçamento inteiro. Aprovar parte é um orçamento
  novo.
- Não tem estado de **renegociação**: principal recusado cancela a OS.
- Não movimenta estoque por conta própria: chama os casos de uso de Peças e de Insumos, que fazem
  a reserva, a compra e a devolução.
