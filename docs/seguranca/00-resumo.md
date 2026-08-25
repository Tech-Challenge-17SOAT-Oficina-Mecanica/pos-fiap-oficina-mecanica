---
documento: Resumo do Contexto — Segurança
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-23
status: rascunho
---

# Resumo do Contexto — Segurança

## O que este contexto cobre

Autenticação de usuários internos da oficina e autorização por escopos nas APIs administrativas.

## Tarefas documentadas

| # | Tarefa | Rota | Arquivo |
|---|---|---|---|
| 1 | Autenticar Mecânico | `POST /autenticacao/login` | [autenticar-mecanico.md](autenticar-mecanico.md) |
| 2 | Cadastrar Mecânico | `POST /mecanicos` | [cadastrar-mecanico.md](cadastrar-mecanico.md) |

## Tipos do contexto

**Usuário**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador da conta autenticada. |
| `email` | string | Identificador de login, único. |
| `senhaHash` | string | Hash da senha; nunca é exposto pela API. |
| `ativo` | boolean | Conta inativa não pode autenticar. |
| `criadoEm` | datetime | Data de criação da conta. |

**Mecânico**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador do profissional no domínio. |
| `usuarioId` | uuid | Vínculo único com `usuario`. |
| `nome` | string | Nome do profissional. |

## Convenções em vigor

- JWT é usado nas APIs administrativas, enviado como `Authorization: Bearer <token>`.
- O token identifica o usuário (`sub`), os escopos e sua expiração.
- A autorização é feita por escopo.
- A senha é persistida somente como hash BCrypt.
- O cliente não possui login nesta entrega. O acesso dele, quando necessário, será por token temporário e restrito à sua Ordem de Serviço.

## O que este contexto não faz

- Não atualiza ou inativa mecânicos nesta entrega.
- Não implementa recuperação ou troca de senha.
