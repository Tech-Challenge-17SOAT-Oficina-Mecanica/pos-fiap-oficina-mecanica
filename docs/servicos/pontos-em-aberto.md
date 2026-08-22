---
documento: Pontos em Aberto — Contexto de Serviços
dono: João Victor Silva de Oliveira
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Pontos em Aberto — Serviços

Este documento centraliza as decisões pendentes das tarefas do contexto de Serviços.

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Confirmar se clientes também poderão consultar o catálogo de serviços ou se a consulta será restrita ao Gestor. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 2 | Confirmar se o escopo definitivo para consulta de serviços será `servicos:ler` ou outro nome padronizado pelo time. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 3 | Definir o limite máximo de `size` para paginação da listagem de serviços. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 4 | Confirmar se a listagem paginada de Serviços deve usar `content`, `page`, `size`, `totalElements` e `totalPages`, ou o envelope em português já usado em outros contextos. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 5 | Definir se serviços inativos aparecem na consulta padrão ou apenas quando filtrados por `status=INATIVO`. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 6 | Definir o critério de unicidade para impedir serviço duplicado: nome exato, nome normalizado, código ou combinação de campos. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 7 | Definir se `tempoEstimadoMinutos` será obrigatório no MVP ou opcional. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 8 | Definir a regra de geração do código funcional `SER-000001`, incluindo sequência, reset anual ou sequência global. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 9 | Confirmar se o escopo definitivo para cadastro de serviços será `servicos:escrever` ou outro nome padronizado pelo time. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 10 | Confirmar se atualização de serviço deve usar controle otimista com header `If-Match`, como outras operações de escrita do projeto. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 11 | Definir se `PATCH /servicos/{id}` aceitará atualização parcial real ou payload completo dos campos editáveis. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 12 | Confirmar quais campos são imutáveis na atualização, especialmente `codigo`, `status` e identificadores técnicos. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 13 | Definir como os valores históricos de serviços em OS e orçamentos serão preservados: cópia do valor no momento do uso ou histórico versionado do serviço. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 14 | Verbo da desativação: aqui é `PATCH /servicos/{servicoId}/desativar`, enquanto Cliente, Veículo e Peças & Insumos usam `DELETE` para a mesma transição lógica. Padronizar entre os contextos. | [`desativar-servico.md`](desativar-servico.md) | — |
| 15 | Situação do serviço: o refinamento da desativação usava `status` com `ATIVO` e `INATIVO`. Foi padronizado como o booleano `ativo`, mais `dataDesativacao` e `usuarioDesativacao`, como nos demais contextos. Confirmar e alinhar com os documentos de cadastro, consulta e atualização. | [`desativar-servico.md`](desativar-servico.md) | — |
| 16 | Remoção física quando não há vínculo (RF-SRV-06) não tem endpoint nem regra definida. Decidir se o MVP suporta os dois caminhos ou apenas a desativação lógica. | [`desativar-servico.md`](desativar-servico.md) | — |
| 17 | Reativação de serviço inativo não está prevista, embora Cliente e Veículo tenham rota de reativação. Definir se entra no MVP. | [`desativar-servico.md`](desativar-servico.md) | — |
