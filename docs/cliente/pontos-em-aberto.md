---
documento: Pontos em Aberto — Contexto de Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Pontos em Aberto — Cliente

Este documento centraliza as decisões pendentes das tarefas do contexto de Cliente.

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Confirmar o dono do documento do contexto Cliente. | — | — |
| 2 | Confirmar se o escopo definitivo para consulta de clientes será `clientes:ler` ou outro nome padronizado pelo time. | [`consultar-cliente.md`](consultar-cliente.md) | — |
| 3 | Definir se a resposta de consulta de cliente deve usar envelope simples, como documentado aqui, ou algum envelope padronizado para recursos únicos. | [`consultar-cliente.md`](consultar-cliente.md) | — |
| 4 | Confirmar se o escopo definitivo para cadastro de clientes será `clientes:escrever` ou outro nome padronizado pelo time. | [`cadastrar-cliente.md`](cadastrar-cliente.md) | — |
| 5 | Confirmar se o escopo definitivo para atualização de clientes também será `clientes:escrever` ou se haverá escopo específico. | [`atualizar-cliente.md`](atualizar-cliente.md) | — |
| 6 | Confirmar se atualização de cliente deve usar controle otimista com header `If-Match`, como outras operações de escrita do projeto. | [`atualizar-cliente.md`](atualizar-cliente.md) | — |
| 7 | Confirmar se o vínculo entre cliente e veículo pertence definitivamente ao contexto Cliente ou se deve ficar no contexto Veículo. | [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) | — |
| 8 | Confirmar se vincular veículo ao cliente deve retornar `200` com confirmação, como documentado aqui, ou `201` por criar um vínculo novo. | [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) | — |
| 9 | Confirmar se o vínculo entre cliente e veículo deve usar controle otimista com header `If-Match`. | [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) | — |


