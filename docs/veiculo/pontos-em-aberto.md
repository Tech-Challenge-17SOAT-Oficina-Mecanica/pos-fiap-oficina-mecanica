---
documento: Pontos em Aberto — Contexto de Veículo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Pontos em Aberto — Veículo

Este documento centraliza as decisões pendentes das tarefas do contexto de Veículo.

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Confirmar o dono do documento do contexto Veículo. | — | — |
| 2 | Confirmar se o escopo definitivo para consulta de veículos será `veiculos:ler` ou outro nome padronizado pelo time. | [`consultar-veiculo.md`](consultar-veiculo.md) | — |
| 3 | Definir se a resposta de consulta de veículo deve usar envelope simples, como documentado aqui, ou algum envelope padronizado para recursos únicos. | [`consultar-veiculo.md`](consultar-veiculo.md) | — |
| 4 | Definir o padrão definitivo de validação de placa, incluindo placa Mercosul e placa antiga. | [`consultar-veiculo.md`](consultar-veiculo.md) | — |
| 5 | Confirmar se o escopo definitivo para cadastro de veículos será `veiculos:escrever` ou outro nome padronizado pelo time. | [`cadastrar-veiculo.md`](cadastrar-veiculo.md) | — |
| 6 | Definir as validações de negócio para o campo `ano`, incluindo ano mínimo e se ano futuro é permitido. | [`cadastrar-veiculo.md`](cadastrar-veiculo.md) | — |
| 7 | O refinamento de Deletar Veículo trazia a persona como Atendente e pedia restringir a operação aos perfis `ATENDENTE` e `GESTOR`, além de dizer que `MECANICO` recebe `403` logo depois de listar `MECANICO` entre os perfis permitidos. Foi padronizado como persona Mecânico e perfis `MECANICO` e `GESTOR`. Confirmar. | [`deletar-veiculo.md`](deletar-veiculo.md) | — |
| 8 | A exclusão lógica depende de índice parcial `UNIQUE (placa) WHERE ativo = true` e da ausência de `ON DELETE CASCADE` nas foreign keys de OS. Confirmar com quem cuidar da migration. | [`deletar-veiculo.md`](deletar-veiculo.md) | — |
| 9 | A reativação de veículo exige cliente proprietário ativo e devolve `422` nesse caso. Confirmar se `422` é o código certo, considerando a padronização de `409` e `422` discutida em [`02-decisoes-arquiteturais.md`](../02-decisoes-arquiteturais.md). | [`deletar-veiculo.md`](deletar-veiculo.md) | — |
| 10 | Confirmar se atualização de veículo deve usar controle otimista com header `If-Match`, como outras operações de escrita do projeto. | [`atualizar-veiculo.md`](atualizar-veiculo.md) | — |
| 11 | Confirmar se alteração de placa é permitida para veículo com Ordens de Serviço existentes ou se deve haver restrição específica. | [`atualizar-veiculo.md`](atualizar-veiculo.md) | — |
| 12 | Confirmar se o escopo definitivo para atualização de veículos será `veiculos:escrever` ou se haverá escopo específico. | [`atualizar-veiculo.md`](atualizar-veiculo.md) | — |
