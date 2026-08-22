---
documento: Pontos em Aberto — Contexto de Orçamento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Pontos em Aberto — Orçamento

Este documento centraliza as decisões pendentes das tarefas do contexto de Orçamento.

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Status do orçamento: a decisão atual adota `CRIADO`, `APROVADO` e `RECUSADO`. Os demais refinamentos do contexto devem ser revisados para refletir esse padrão. | [`calcular-orcamento.md`](calcular-orcamento.md), [`aprovar-orcamento.md`](aprovar-orcamento.md), [`recusar-orcamento.md`](recusar-orcamento.md) e [`gerar-orcamento-complementar.md`](gerar-orcamento-complementar.md) | — |
| 2 | Efeito da aprovação do orçamento complementar: aprovar e recusar hoje só cobrem o orçamento principal, com transições para `AGUARDANDO_EXECUCAO` e `CANCELADA`. Definir o que a decisão sobre um complementar faz com a OS, que já está em execução. | [`gerar-orcamento-complementar.md`](gerar-orcamento-complementar.md), [`aprovar-orcamento.md`](aprovar-orcamento.md) e [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 3 | Como o cliente se autentica: aprovar, recusar e consultar têm o cliente como ator, mas o enunciado só prevê JWT para as APIs administrativas. Definir se o cliente recebe token próprio, link assinado ou outro mecanismo, e qual escopo ele carrega. | [`aprovar-orcamento.md`](aprovar-orcamento.md), [`recusar-orcamento.md`](recusar-orcamento.md) e [`consultar-orcamento.md`](consultar-orcamento.md) | — |
| 4 | Nomes dos escopos (`orcamentos:ler`, `orcamentos:escrever`, `orcamentos:aprovar`) foram derivados do padrão `recurso:acao` usado nos demais contextos. Confirmar a lista oficial de escopos do projeto. | Todas as tarefas do contexto | — |
| 5 | Definir o canal de comunicação do envio de orçamento e as regras de reenvio. | [`enviar-orcamento.md`](enviar-orcamento.md) | — |
| 6 | Reserva de peças na aprovação: o documento que descrevia a reserva disparada pelo evento `OrcamentoAprovado` foi removido, e hoje quem cria a reserva é o pedido de compra. A aprovação só muda o status da OS. Definir quem dispara a reserva quando as peças já estão em estoque. | [`aprovar-orcamento.md`](aprovar-orcamento.md) e [`../pecas-e-insumos/solicitar-compra-de-pecas.md`](../pecas-e-insumos/solicitar-compra-de-pecas.md) | — |
| 7 | Recusa parcial: o cliente aprova ou recusa o orçamento inteiro. Definir se recusar item a item entra no MVP. | [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 8 | Um orçamento recusado cancela a OS. Confirmar se é isso mesmo para o orçamento principal, e se o veículo pode ser devolvido sem serviço nenhum, ou se existe um estado intermediário de renegociação. | [`recusar-orcamento.md`](recusar-orcamento.md) | — |
| 9 | Notificação ao cliente não tem canal definido, e o cliente não tem contato cadastrado (ver [`pontos-em-aberto.md`](../cliente/pontos-em-aberto.md) do contexto de Cliente). | [`gerar-orcamento-complementar.md`](gerar-orcamento-complementar.md) | — |
| 10 | Definir quando o endpoint de cálculo será acionado após a criação ou alteração dos itens do orçamento. | [`calcular-orcamento.md`](calcular-orcamento.md) | — |
