---
documento: Pontos em Aberto — Contexto de Ordem de Serviço
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Pontos em Aberto — Ordem de Serviço

Este documento centraliza as decisões pendentes das tarefas do contexto de Ordem de Serviço.

> O arquivo anterior foi retirado do repositório junto com os documentos que estão sendo
> reenviados. Os pontos abaixo vêm das tarefas atualmente presentes; quando os documentos
> reenviados chegarem, os pontos deles voltam para esta tabela.

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Máquina de estados completa da OS: `AGUARDANDO_EXECUCAO` e `CANCELADA` aparecem nos fluxos, mas não estão na lista de status do enunciado (`Recebida`, `Em diagnóstico`, `Aguardando aprovação`, `Em execução`, `Finalizada`, `Entregue`). Fechar a lista oficial e as transições permitidas. | [`listar-ordens-de-servico.md`](listar-ordens-de-servico.md), [`finalizar-servico.md`](finalizar-servico.md) e [`registrar-entrega-de-veiculo.md`](registrar-entrega-de-veiculo.md) | — |
| 2 | Pagamento não é um contexto documentado do projeto, mas a entrega do veículo depende da confirmação do pagamento e apresenta o valor final ao cliente. Definir se pagamento entra no MVP, e como, ou se a regra de bloqueio por pagamento sai do fluxo de entrega. | [`registrar-entrega-de-veiculo.md`](registrar-entrega-de-veiculo.md) | — |
| 3 | Nome do path param da OS: os documentos usavam `{id}`, `{ordemServicoId}` e `{osId}`. Foi padronizado `{osId}`. Confirmar e alinhar os documentos reenviados. | Todas as tarefas do contexto | — |
| 4 | Envelope de paginação: os refinamentos usavam `content`/`page`/`size`/`totalElements` e `pagina`/`tamanho`/`total`/`itens`. Foi padronizado `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`. Confirmar. | [`listar-ordens-de-servico.md`](listar-ordens-de-servico.md) e [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 5 | Valores de status na API: a listagem trazia rótulos com acento e espaço (`Em diagnóstico`), enquanto os demais documentos usam `EM_DIAGNOSTICO`. Foi padronizado o formato em maiúsculas com sublinhado. Confirmar se a apresentação amigável fica com o cliente da API. | [`listar-ordens-de-servico.md`](listar-ordens-de-servico.md) | — |
| 6 | `event_data` usa `snake_case` no meio de uma resposta em `camelCase`, e seus campos internos estão em inglês (`aggregateType`, `eventType`, `occurredAt`). Definir o padrão de nomenclatura da API para esse bloco. | [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 7 | A consulta por CPF/CNPJ foi resolvida como filtro da listagem (`GET /ordens-servico?documento=`), porque as duas tarefas propunham a mesma rota. Confirmar essa divisão entre listar e consultar. | [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) e [`listar-ordens-de-servico.md`](listar-ordens-de-servico.md) | — |
| 8 | Autenticação do cliente: o cliente consulta a própria OS, mas o enunciado só prevê JWT para as APIs administrativas. Mesma pendência registrada no contexto de Orçamento. | [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 9 | Problemas da OS são retornados sem o campo de tipo, por decisão explícita do refinamento. Confirmar o motivo e se o tipo é usado internamente. | [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 10 | A finalização exige que as movimentações de estoque estejam concluídas, mas não define se isso bloqueia a finalização. Fechar a regra com o contexto de Peças & Insumos. | [`finalizar-servico.md`](finalizar-servico.md) | — |
| 11 | Notificação ao cliente na finalização não tem canal definido, e o cliente não tem contato cadastrado. Mesma pendência do contexto de Cliente. | [`finalizar-servico.md`](finalizar-servico.md) | — |
| 12 | Numeração das tarefas do contexto: a ordem atual segue a ordem em que foram refinadas, não a ordem do fluxo. Decidir se renumera quando os documentos reenviados chegarem. | Todas as tarefas do contexto | — |
