---
documento: Pontos em Aberto — Contexto de Orçamento
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Orçamento

## O que é este documento

Um ponto em aberto é uma **decisão que ainda não foi tomada** ou uma **inconsistência encontrada**
entre os documentos deste contexto. Enquanto o ponto estiver aberto, quem for implementar não deve
resolver sozinho: a escolha muda contrato de API, modelo de dados ou regra de negócio.

Para fechar um ponto:

1. aplique a decisão nos documentos afetados;
2. registre o porquê — como *Decisão de projeto* no documento da tarefa, ou em
   [`02-decisoes-arquiteturais.md`](../02-decisoes-arquiteturais.md) quando valer para todos os contextos;
3. remova a linha desta tabela.

O retrato do que já está decidido está em [`00-resumo.md`](00-resumo.md).

## Inconsistências a corrigir

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Duas modelagens concorrentes para o complementar: o contexto de Ordem de Serviço fala em **adições** dentro de um mesmo orçamento (`orcamento_adicao`), enquanto aqui o complementar é um **orçamento separado** com `tipo` e `orcamentoOriginalId`. Escolher uma. | [`calcular-orcamento.md`](calcular-orcamento.md) e [`../ordem-de-servico/registrar-pecas-e-insumos-necessarios.md`](../ordem-de-servico/registrar-pecas-e-insumos-necessarios.md) | — |
| 2 | Aprovar e recusar cobrem apenas o orçamento principal. Falta definir o efeito da decisão sobre um complementar, com a OS já em execução. | [`aprovar-orcamento.md`](aprovar-orcamento.md) e [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 3 | Escopos diferentes para a mesma decisão do cliente: aprovar usa `orcamentos:aprovar` e recusar usa `orcamentos:recusar`. Confirmar se são dois escopos ou um só. | [`aprovar-orcamento.md`](aprovar-orcamento.md) e [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 4 | Envelope de paginação divergente: a consulta usa `content`, `page`, `size`, `totalElements` e `totalPages`, enquanto o resto do projeto usa `data`, `pagina`, `tamanho`, `totalElementos` e `totalPaginas`. | [`consultar-orcamento.md`](consultar-orcamento.md) | — |
| 5 | A recusa dispara a devolução dos itens ao estoque, descrita em Peças & Insumos, mas o documento de recusa não cita essa chamada. Alinhar os dois lados. | [`recusar-orcamento.md`](recusar-orcamento.md) e [`../pecas-e-insumos/retornar-peca-e-insumo-ao-estoque.md`](../pecas-e-insumos/retornar-peca-e-insumo-ao-estoque.md) | — |
| 6 | A aprovação deveria colocar a OS na fila e disparar reserva ou compra dos itens. Hoje a aprovação só muda o status da OS, e quem reserva é o fluxo de Peças & Insumos. Definir o encadeamento. | [`aprovar-orcamento.md`](aprovar-orcamento.md) | — |
| 7 | Status do orçamento: a decisão atual adota `CRIADO`, `APROVADO` e `RECUSADO`. Os demais refinamentos do contexto devem ser revisados para refletir esse padrão. | [`calcular-orcamento.md`](calcular-orcamento.md), [`aprovar-orcamento.md`](aprovar-orcamento.md), [`recusar-orcamento.md`](recusar-orcamento.md) e `gerar-orcamento-complementar.md`, documento retirado para reescrita | — |
| 8 | Efeito da aprovação do orçamento complementar: aprovar e recusar hoje só cobrem o orçamento principal, com transições para `AGUARDANDO_EXECUCAO` e `CANCELADA`. Definir o que a decisão sobre um complementar faz com a OS, que já está em execução. | `gerar-orcamento-complementar.md`, documento retirado para reescrita, [`aprovar-orcamento.md`](aprovar-orcamento.md) e [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 9 | Como o cliente se autentica: aprovar, recusar e consultar têm o cliente como ator, mas o enunciado só prevê JWT para as APIs administrativas. Definir se o cliente recebe token próprio, link assinado ou outro mecanismo, e qual escopo ele carrega. | [`aprovar-orcamento.md`](aprovar-orcamento.md), [`recusar-orcamento.md`](recusar-orcamento.md) e [`consultar-orcamento.md`](consultar-orcamento.md) | — |
| 10 | Nomes dos escopos (`orcamentos:ler`, `orcamentos:escrever`, `orcamentos:aprovar`) foram derivados do padrão `recurso:acao` usado nos demais contextos. Confirmar a lista oficial de escopos do projeto. | Todas as tarefas do contexto | — |
| 11 | Definir o canal de comunicação do envio de orçamento e as regras de reenvio. | `enviar-orcamento.md`, documento retirado para reescrita | — |
| 12 | Reserva de peças na aprovação: o documento que descrevia a reserva disparada pelo evento `OrcamentoAprovado` foi removido, e hoje quem cria a reserva é o pedido de compra. A aprovação só muda o status da OS. Definir quem dispara a reserva quando as peças já estão em estoque. | [`aprovar-orcamento.md`](aprovar-orcamento.md) e [`../pecas-e-insumos/solicitar-compra-de-pecas.md`](../pecas-e-insumos/solicitar-compra-de-pecas.md) | — |
| 13 | Recusa parcial: o cliente aprova ou recusa o orçamento inteiro. Definir se recusar item a item entra no MVP. | [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 14 | Um orçamento recusado cancela a OS. Confirmar se é isso mesmo para o orçamento principal, e se o veículo pode ser devolvido sem serviço nenhum, ou se existe um estado intermediário de renegociação. | [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 15 | Notificação ao cliente não tem canal definido, e o cliente não tem contato cadastrado (ver [`pontos-em-aberto.md`](../cliente/pontos-em-aberto.md) do contexto de Cliente). | `gerar-orcamento-complementar.md`, documento retirado para reescrita | — |
| 16 | Definir quando o endpoint de cálculo será acionado após a criação ou alteração dos itens do orçamento. | [`calcular-orcamento.md`](calcular-orcamento.md) | — |
