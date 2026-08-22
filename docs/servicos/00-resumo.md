---
documento: Resumo do Contexto — Serviços
dono: A definir
versao: 0.2
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
| 1 | Cadastrar Serviço | `POST /servicos` | `servicos:escrever` | [cadastrar-servico.md](cadastrar-servico.md) |
| 2 | Consultar Serviços | `GET /servicos` e `GET /servicos/{servicoId}` | `servicos:ler` | [consultar-servicos.md](consultar-servicos.md) |
| 3 | Atualizar Serviço | `PATCH /servicos/{servicoId}` | `servicos:escrever` | [atualizar-servico.md](atualizar-servico.md) |
| 4 | Desativar e Reativar Serviço | `DELETE /servicos/{servicoId}` e `POST /servicos/{servicoId}/reativacao` | `servicos:escrever` | [desativar-servico.md](desativar-servico.md) |

## Tipos do contexto

**Serviço**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, gerado pelo sistema. |
| `codigo` | string | Código funcional do catálogo, no formato `SER-000001`. |
| `nome` | string | Obrigatório. |
| `nomeNormalizado` | string | Derivado do nome, sem acento, sem espaço duplo e em minúsculas. Único entre serviços ativos. |
| `descricao` | string | Opcional. |
| `valor` | decimal | Maior ou igual a zero. |
| `tempoEstimadoMinutos` | int | Obrigatório, mínimo de 1 minuto. Insumo do indicador de tempo médio de execução. |
| `ativo` | boolean | `false` após a desativação. |
| `dataDesativacao` | datetime | Preenchido na desativação; nulo após a reativação. |
| `usuarioDesativacao` | uuid | Usuário responsável pela desativação; nulo após a reativação. |
| `dataCriacao` | datetime | Imutável. |
| `version` | int | Controle otimista; enviada no `If-Match` da atualização. |

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; recurso no plural; path param `{servicoId}`.
- Autenticação `Bearer <JWT>`; perfil `MECANICO`; escopos `servicos:ler` e `servicos:escrever`. O
  catálogo é restrito à oficina: o cliente vê os serviços pelo orçamento.
- A retirada do catálogo é **lógica**, feita por `DELETE /servicos/{servicoId}`, com reativação
  por `POST /servicos/{servicoId}/reativacao`.
- Unicidade por **nome normalizado**, apenas entre serviços ativos, por índice parcial
  `UNIQUE (nome_normalizado) WHERE ativo = true`.
- `codigo` no formato `SER-000001`, gerado pelo sistema em sequência global, sem reset.
- Campos imutáveis na atualização: `id`, `codigo` e `dataCriacao`. `ativo` não é alterado pelo
  `PATCH`, e sim pelas rotas de desativação e reativação.
- `PATCH` é atualização parcial de verdade: campo ausente não é alterado.
- A atualização usa controle otimista: `If-Match` com a `version` atual, `412` quando diverge e
  `428` quando o header não vem.
- Listagem com envelope `data`, `pagina`, `tamanho`, `totalElementos` e `totalPaginas`; recurso
  único vai direto, sem envelope. `tamanho` tem teto de **50**.
- Serviços inativos ficam fora da listagem por padrão; aparecem com `incluirInativos=true`.
- O valor do serviço é copiado para a OS e para o orçamento no momento do uso, para que alteração
  de preço não mude documento já emitido.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409`, `412` e `428`.

## O que este contexto não faz

- Não executa serviço nem controla andamento: isso é da Ordem de Serviço.
- Não calcula o tempo médio de execução; apenas fornece o tempo estimado de cada serviço.
- Não remove serviço fisicamente: a exclusão é sempre lógica, e a remoção física ficou fora do MVP.
- Não expõe o catálogo ao cliente.
