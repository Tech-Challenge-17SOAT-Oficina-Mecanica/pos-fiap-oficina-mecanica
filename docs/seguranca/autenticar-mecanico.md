---
documento: Refinamento de Requisitos — Autenticar Mecânico
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-23
status: rascunho
---

# Refinamento de Requisitos — Autenticar Mecânico

Este documento define o primeiro acesso autenticado da oficina. Ele habilita a proteção das APIs administrativas e a identificação do responsável em ações auditáveis.

## 1 · Autenticar Mecânico

### 1.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Entrar no sistema da oficina para executar ações permitidas e deixar rastreável quem as realizou.

**Problema**

Sem autenticação, qualquer pessoa que conheça a URL da API pode alterar dados e não há como identificar o responsável pela ação.

**Pré-condições**

- O mecânico possui uma conta ativa vinculada ao seu cadastro profissional.
- A conta possui os escopos necessários para suas atividades.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-SEG-01 | Permitir que um mecânico ativo se autentique por e-mail e senha. |
| RF-SEG-02 | Emitir um JWT com a identidade, escopos e expiração da conta autenticada. |
| RF-SEG-03 | Permitir que as próximas rotas administrativas validem o JWT e consultem o usuário autenticado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SEG-01 | A senha deve ser armazenada somente como hash BCrypt. |
| RNF-SEG-02 | Credenciais inválidas não devem revelar se o e-mail existe. |
| RNF-SEG-03 | O token deve ser assinado e ter expiração. |

**Fluxo Principal**

1. O mecânico informa e-mail e senha.
2. O sistema encontra a conta ativa e verifica a senha.
3. O sistema obtém os escopos da conta.
4. O sistema devolve um JWT de acesso.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | E-mail ou senha inválidos | Retorna erro de autenticação sem informar qual dado falhou. |
| A2 | Conta inativa | Retorna erro de autenticação. |

**Saída**

- Token de acesso que permite chamar rotas compatíveis com os escopos recebidos.

**Pós-condições**

- Nenhum dado de negócio é alterado.
- O token pode identificar o usuário em requisições posteriores até expirar.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /autenticacao/login
```

> Um mecânico inicial com o escopo `mecanicos:escrever` será incluído no seed de desenvolvimento e testes para permitir o primeiro acesso e o cadastro dos demais.

**Autenticação / Autorização**

- Não se aplica ao login.

**Entrada**

```json
{
  "email": "mecanico@oficina.local",
  "senha": "senha-do-mecanico"
}
```

| Campo | Tipo | Descrição |
|---|---|---|
| `email` | string | Obrigatório; e-mail da conta. |
| `senha` | string | Obrigatória; enviada somente no login. |

**Validações**

- `email` e `senha` são obrigatórios.
- A conta deve existir e estar ativa.
- A senha deve corresponder ao hash persistido.

**Processamento**

1. Validar o corpo da requisição.
2. Buscar o usuário pelo e-mail.
3. Verificar se a conta está ativa e se a senha confere.
4. Buscar os escopos do usuário.
5. Gerar JWT HS256 assinado pela chave configurada em `JWT_SECRET`, com expiração de uma hora.
6. Retornar o token e sua expiração.

**Persistência**

- Consulta: `usuario`, `mecanico` e `usuario_escopo`.
- Altera: nenhuma tabela.
- Novas tabelas: `usuario`, `mecanico` e `usuario_escopo`.

**Modelo de relacionamento**

```text
usuario 1 ─── 1 mecanico
usuario 1 ─── N usuario_escopo
ordem_servico N ─── 1 mecanico
```

`usuario` representa a identidade e o acesso; `mecanico` representa o profissional que executa serviços e poderá ser vinculado à Ordem de Serviço. O cliente continua sendo uma entidade de negócio separada e não possui login nesta entrega.

**Saída da API**

```json
{
  "accessToken": "<jwt>",
  "tokenType": "Bearer",
  "expiresIn": 3600
}
```

O JWT contém `sub` (id do usuário), `escopos` e `exp`.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Mecânico autenticado. |
| `400` | Corpo ausente ou campos obrigatórios inválidos. |
| `401` | Credenciais inválidas ou conta inativa. |
| `500` | Configuração de assinatura ausente ou falha interna. |

**Dependências**

- `UsuarioRepository`.
- `MecanicoRepository`.
- Configuração `JWT_SECRET`.
- Middleware HTTP de autenticação e autorização por escopo.

**Testes**

*Unitários*

- Gera token com usuário, escopos e expiração.
- Rejeita token ausente, expirado ou com assinatura inválida.
- Rejeita usuário inativo e senha inválida.

*Integração*

- Login válido retorna `200` e JWT válido.
- Credenciais inválidas retornam `401`.
- Corpo inválido retorna `400`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar entidades `Usuario` e `Mecanico`.

**Caso de uso**

- [ ] Implementar autenticação do mecânico.

**Repositório**

- [ ] Implementar consulta de usuário, mecânico e escopos por e-mail.

**Handler HTTP**

- [ ] Implementar `POST /autenticacao/login`.
- [ ] Implementar middleware reutilizável de autenticação e autorização por escopo.

**Validações**

- [ ] Validar corpo, conta ativa e senha.
- [ ] Validar assinatura e expiração do JWT.

**Testes unitários**

- [ ] Cobrir geração e validação de JWT e credenciais inválidas.

**Testes de integração**

- [ ] Cobrir login válido, credenciais inválidas e corpo inválido.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Migration versionada e reversível.
- [ ] Code Review aprovado.

## Pontos em aberto

| # | Ponto | Responsável |
|---|---|---|
| 1 | Inativação, recuperação e troca de senha de mecânicos serão refinadas em tarefas próprias. | — |
| 2 | O token temporário do cliente será refinado junto ao envio de orçamento. | — |
