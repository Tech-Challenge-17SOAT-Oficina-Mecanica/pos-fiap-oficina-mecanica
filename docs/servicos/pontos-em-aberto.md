---
documento: Pontos em Aberto — Contexto de Serviços
dono: João Victor Silva de Oliveira
versao: 0.1
atualizado_em: 2026-08-20
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
