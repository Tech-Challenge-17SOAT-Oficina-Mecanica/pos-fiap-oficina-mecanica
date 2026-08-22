---
documento: Resumo do Contexto — Serviços
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Serviços

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. O que ainda não está resolvido fica em
[`pontos-em-aberto.md`](pontos-em-aberto.md).

## O que este contexto cobre

O **catálogo de serviços** oferecidos pela oficina: o que pode ser vendido, por quanto e em quanto
tempo. A execução do serviço em si pertence ao contexto de Ordem de Serviço; aqui fica apenas o
cadastro que alimenta a OS e o orçamento.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 1 | Consultar Serviços | `GET /servicos` e `GET /servicos/{servicoId}` | `servicos:ler` | [consultar-servicos.md](consultar-servicos.md) |
| 2 | Cadastrar Serviço | `POST /servicos` | `servicos:escrever` | [cadastrar-servico.md](cadastrar-servico.md) |
| 3 | Atualizar Serviço | `PATCH /servicos/{servicoId}` | `servicos:escrever` | [atualizar-servico.md](atualizar-servico.md) |
| 1 | Desativar Serviço | `PATCH /servicos/{servicoId}/desativar` | `servicos:escrever` | [desativar-servico.md](desativar-servico.md) |

> A numeração das tarefas está duplicada: Consultar e Desativar são ambas `1 ·`. Ver o ponto 1 de
> [`pontos-em-aberto.md`](pontos-em-aberto.md).

## Tipos do contexto

**Serviço**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, gerado pelo sistema. |
| `codigo` | string | Código funcional do catálogo, no formato `SER-000001`. |
| `nome` | string | Obrigatório. |
| `descricao` | string | Opcional. |
| `valor` | decimal | Maior que zero. |
| `tempoEstimadoMinutos` | int | Insumo do indicador de tempo médio de execução. |
| `status` ou `ativo` | enum ou boolean | **Divergente entre os documentos** — ver ponto 2. |
| `dataDesativacao` | datetime | Preenchido na desativação. |
| `usuarioDesativacao` | uuid | Usuário responsável pela desativação. |

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; recurso no plural.
- Autenticação `Bearer <JWT>`; escopos `servicos:ler` e `servicos:escrever`; a desativação está
  restrita ao perfil `GESTOR`.
- A retirada do catálogo é **lógica**, feita por `PATCH .../desativar` — e não por `DELETE`, como
  nos demais contextos.
- O valor do serviço é copiado para a OS e para o orçamento no momento do uso, para que alteração
  de preço não mude documento já emitido.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409` e `412` no controle otimista.

## O que este contexto não faz

- Não executa serviço nem controla andamento: isso é da Ordem de Serviço.
- Não calcula o tempo médio de execução; apenas fornece o tempo estimado de cada serviço.
- Não prevê reativação de serviço inativo.
