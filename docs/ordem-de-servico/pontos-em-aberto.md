---
documento: Pontos em Aberto — Contexto de Ordem de Serviço
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Ordem de Serviço

## O que é este documento

Um ponto em aberto é uma **decisão que ainda não foi tomada** ou uma **inconsistência encontrada**
entre os documentos deste contexto. Enquanto o ponto estiver aberto, quem for implementar não deve
resolver sozinho: a escolha muda contrato de API, modelo de dados ou regra de negócio.

Para fechar um ponto:

1. aplique a decisão nos documentos afetados;
2. registre o porquê — como *Decisão de projeto* no documento da tarefa, ou em
   [`02-decisoes-arquiteturais.md`](../02-decisoes-arquiteturais.md) quando valer para todos os contextos;
3. remova a linha desta tabela.

Este é o contexto com mais dependências do projeto, então boa parte dos pontos aqui só fecha em
conversa com outro contexto. O retrato do que já está decidido está em [`00-resumo.md`](00-resumo.md).

## Inconsistências a corrigir

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Duplicidade de IDs: `RF-OS-38` e `RF-OS-39` aparecem em dois documentos, e `RF-OS-129` a `RF-OS-136` também. Renumerar o contexto inteiro de uma vez. | Todos os documentos do contexto | — |
| 2 | Máquina de estados da OS: `AGUARDANDO_RECURSOS` e `AGUARDANDO_EXECUCAO` são usados nos fluxos, mas não constam no enunciado. Fechar a lista oficial de status e as transições permitidas, incluindo `CANCELADA`. | Todos os documentos do contexto | — |
| 3 | Duas modelagens concorrentes para o complementar: registrar peças e insumos fala em **adições** do orçamento (`orcamento_adicao`), e registrar problema encontrado fala em **orçamentos separados** por tipo. Escolher uma antes de implementar. | [`registrar-pecas-e-insumos-necessarios.md`](registrar-pecas-e-insumos-necessarios.md) e [`registrar-problema-encontrado.md`](registrar-problema-encontrado.md) | — |
| 4 | Registrar peças e insumos necessários veio só com o refinamento de produto: faltam o refinamento técnico, a rota e o checklist. | [`registrar-pecas-e-insumos-necessarios.md`](registrar-pecas-e-insumos-necessarios.md) | — |
| 5 | Criar Ordem de Serviço declara o endpoint em texto (`- POST /ordens-servico`) em vez de bloco `http`, como manda o guia. Ajustar o formato. | [`criar-ordem-de-servico.md`](criar-ordem-de-servico.md) | — |
| 6 | O registro de peças e insumos exige que o item já tenha reserva ou pedido de compra vinculado à OS, mas quem cria a reserva é o contexto de Peças & Insumos, depois da aprovação. A ordem dos passos entre os dois contextos precisa ser desenhada junto. | [`registrar-pecas-e-insumos-necessarios.md`](registrar-pecas-e-insumos-necessarios.md) | — |
| 7 | A entrada de estoque muda o status das OS vinculadas ao pedido recebido. Confirmar se essa transição pertence a Peças & Insumos ou se deve ser feita aqui, por evento. | [`../pecas-e-insumos/registrar-entrada-de-estoque.md`](../pecas-e-insumos/registrar-entrada-de-estoque.md) | — |
| 8 | Pagamento não é contexto documentado, mas a entrega depende da confirmação dele e apresenta o valor final ao cliente. Definir se entra no MVP. | [`registrar-entrega-de-veiculo.md`](registrar-entrega-de-veiculo.md) | — |
| 9 | Autenticação do cliente: ele consulta a própria OS, mas o enunciado só prevê JWT para as APIs administrativas. Mesma pendência do contexto de Orçamento. | [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 10 | `event_data` usa `snake_case` no meio de uma resposta em `camelCase`, e seus campos internos estão em inglês. Definir o padrão de nomenclatura desse bloco. | [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 11 | A consulta por CPF/CNPJ é filtro da listagem, e o detalhamento é por identificador. Confirmar essa divisão entre listar e consultar. | [`listar-ordens-de-servico.md`](listar-ordens-de-servico.md) e [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 12 | Problemas da OS são retornados sem o campo de tipo, por decisão explícita. Confirmar o motivo e se o tipo é usado internamente. | [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md) | — |
| 13 | A finalização exige que as movimentações de estoque estejam concluídas, mas não define se isso bloqueia a operação. Fechar a regra com Peças & Insumos. | [`finalizar-servico.md`](finalizar-servico.md) | — |
| 14 | Notificação ao cliente na finalização não tem canal definido, e o cliente não tem contato cadastrado. | [`finalizar-servico.md`](finalizar-servico.md) | — |
| 15 | A numeração das tarefas segue a ordem em que foram refinadas, não a do fluxo. Decidir se renumera. | Todos os documentos do contexto | — |
| 16 | Tarefas ainda não refinadas do contexto: selecionar próxima OS para execução e registrar problema adicional. | — | — |
