---
documento: Resumo do Contexto — Veículo
dono: A definir
versao: 0.3
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
| 2 | Cadastrar Veículo | — rota aposentada; ver tarefa 5 | `veiculos:escrever` | [cadastrar-veiculo.md](cadastrar-veiculo.md) |
| 3 | Deletar Veículo | `DELETE /veiculos/{veiculoId}` e `POST /veiculos/{veiculoId}/reativacao` | `veiculos:escrever` | [deletar-veiculo.md](deletar-veiculo.md) |
| 4 | Atualizar Veículo | `PUT /veiculos/{veiculoId}` | `veiculos:escrever` | [atualizar-veiculo.md](atualizar-veiculo.md) |
| 5 | Cadastrar Veículo e Vincular ao Cliente | `POST /clientes/{clienteId}/veiculos` | `veiculos:escrever` mais `clientes:ler` na consulta prévia | [cadastrar-veiculo-e-vincular-ao-cliente.md](cadastrar-veiculo-e-vincular-ao-cliente.md) |

## Tipos do contexto

**Veículo**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, gerado pelo sistema. |
| `placa` | string | Identificador de negócio; único entre veículos ativos. Formato Mercosul `ABC1D23` ou antigo `ABC1234`, normalizada em maiúsculas, sem hífen e sem espaço. |
| `marca` | string | Obrigatória. |
| `modelo` | string | Obrigatório. |
| `ano` | int | Obrigatório; de `1900` até o ano corrente mais um. |
| `ativo` | boolean | `false` após a exclusão lógica. |
| `inativadoEm` | datetime | Preenchido na inativação. |
| `inativadoPor` | uuid | Usuário responsável pela inativação. |
| `motivoInativacao` | string | Até 200 caracteres, opcional. |
| `version` | int | Controle otimista; enviada no `If-Match` das atualizações. |

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; identificadores em rota sempre UUID; `{veiculoId}` como path param.
- Autenticação `Bearer <JWT>`; perfil `MECANICO`; escopos `veiculos:ler` e `veiculos:escrever`.
- Exclusão é **lógica**: o `DELETE` inativa e preserva o histórico das Ordens de Serviço.
- Unicidade de placa **apenas entre veículos ativos**, por índice parcial
  `UNIQUE (placa) WHERE ativo = true`.
- A reativação exige que não exista outro veículo ativo com a mesma placa e que o cliente
  proprietário esteja ativo — caso contrário, `409`.
- O veículo é inativado em cascata quando o cliente é inativado, por chamada direta dentro da
  mesma transação. O projeto não usa eventos nem mensageria.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409`, `412`, `428` e `204` para
  operação idempotente.
- A atualização usa controle otimista: `If-Match` com a `version` atual, `412` quando diverge e
  `428` quando o header não vem. A consulta expõe `version`.
- Recurso único devolve o objeto direto; listagem devolve o envelope paginado.
- O cadastro de veículo acontece sempre dentro do cliente, por
  `POST /clientes/{clienteId}/veiculos`. `POST /veiculos` foi aposentada.
- A placa pode ser corrigida mesmo com Ordens de Serviço existentes: a OS grava a placa vigente
  no momento em que é criada.

## O que este contexto não faz

- Não mantém histórico de proprietários anteriores: ao trocar o dono, o vínculo anterior não é
  preservado. Decisão do time, fora do MVP.
- Não desvincula veículo do cliente — o desvínculo ficou para depois do MVP.
- Não cadastra veículo sem dono: `POST /veiculos` foi aposentada.
