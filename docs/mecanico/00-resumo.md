---
documento: Resumo do Contexto — Mecânico
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-25
status: rascunho
---

# Resumo do Contexto — Mecânico

## O que este contexto cobre

Cadastro e identificação do profissional da oficina responsável por executar Ordens de Serviço.

## Tarefas documentadas

| # | Tarefa | Rota | Arquivo |
|---|---|---|---|
| 1 | Cadastrar Mecânico | `POST /mecanicos` | [cadastrar-mecanico.md](cadastrar-mecanico.md) |
| 2 | Atualizar Mecânico | `PUT /mecanicos/{mecanicoId}` | [atualizar-mecanico.md](atualizar-mecanico.md) |

## Tipos do contexto

**Mecânico**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador do profissional no domínio. |
| `usuarioId` | uuid | Vínculo único com a conta de Segurança. |
| `nome` | string | Nome do profissional. |
| `version` | integer | Controle otimista para atualizações cadastrais. |

## Convenções em vigor

- `Mecanico` é uma entidade de negócio e pode ser responsável por uma Ordem de Serviço.
- `Usuario`, senha, JWT e escopos pertencem ao contexto de Segurança.
- O cadastro e a atualização de mecânico alteram profissional, conta e escopos em uma única transação.
- Atualizações exigem `If-Match` com a versão atual do mecânico.

## O que este contexto não faz

- Não autentica usuários nem emite JWT.
- Não altera senha, recupera acesso ou inativa contas nesta entrega.
