---
documento: Resumo do Contexto — Veículo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Veículo

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. O que ainda não está resolvido fica em
[`pontos-em-aberto.md`](pontos-em-aberto.md).

## O que este contexto cobre

O cadastro dos veículos atendidos pela oficina, identificados pela placa, e a manutenção desse
cadastro ao longo do tempo. O veículo é sempre atendido no contexto de um cliente e de uma Ordem
de Serviço.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 1 | Consultar Veículo | `GET /veiculos` | `veiculos:ler` | [consultar-veiculo.md](consultar-veiculo.md) |
| 2 | Cadastrar Veículo | `POST /veiculos` | `veiculos:escrever` | [cadastrar-veiculo.md](cadastrar-veiculo.md) |
| 3 | Deletar Veículo | `DELETE /veiculos/{veiculoId}` e `POST /veiculos/{veiculoId}/reativacao` | `veiculos:escrever` | [deletar-veiculo.md](deletar-veiculo.md) |
| 4 | Atualizar Veículo | `PUT /veiculos/{veiculoId}` | `veiculos:escrever` | [atualizar-veiculo.md](atualizar-veiculo.md) |
| 5 | Cadastrar Veículo e Vincular ao Cliente | `POST /clientes/{clienteId}/veiculos` | `veiculos:escrever` mais `clientes:ler` na consulta prévia | [cadastrar-veiculo-e-vincular-ao-cliente.md](cadastrar-veiculo-e-vincular-ao-cliente.md) |

## Tipos do contexto

**Veículo**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, gerado pelo sistema. |
| `placa` | string | Identificador de negócio; único entre veículos ativos. |
| `marca` | string | Obrigatória. |
| `modelo` | string | Obrigatório. |
| `ano` | int | Obrigatório; faixa de validação ainda não definida. |
| `ativo` | boolean | `false` após a exclusão lógica. |
| `inativadoEm` | datetime | Preenchido na inativação. |
| `inativadoPor` | uuid | Usuário responsável pela inativação. |
| `motivoInativacao` | string | Até 200 caracteres, opcional. |

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; identificadores em rota sempre UUID; `{veiculoId}` como path param.
- Autenticação `Bearer <JWT>`; perfis `MECANICO` e `GESTOR`; escopos `veiculos:ler` e `veiculos:escrever`.
- Exclusão é **lógica**: o `DELETE` inativa e preserva o histórico das Ordens de Serviço.
- Unicidade de placa **apenas entre veículos ativos**, por índice parcial
  `UNIQUE (placa) WHERE ativo = true`.
- A reativação exige que não exista outro veículo ativo com a mesma placa e que o cliente
  proprietário esteja ativo — caso contrário, `422`.
- O veículo é inativado em cascata quando o cliente é inativado, consumindo o evento `ClienteInativado`.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409`, `422` e `204` para operação idempotente.

## O que este contexto não faz

- Não decide o dono do vínculo cliente-veículo, que hoje também é descrito no contexto de Cliente.
- Não mantém histórico de proprietários anteriores.
