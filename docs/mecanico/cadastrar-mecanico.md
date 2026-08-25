---
documento: Refinamento de Requisitos — Contexto de Mecânico
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-25
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Mecânico

Este documento define o cadastro do profissional da oficina. A conta de acesso, a senha e os
escopos são criados em integração com o contexto de Segurança.

## 1 · Cadastrar Mecânico

### 1.1 Refinamento de Produto

**Persona**

Mecânico autorizado.

**Objetivo**

Cadastrar outro mecânico para que ele possa acessar o sistema e ser associado às Ordens de Serviço que executar.

**Problema**

Sem um cadastro de profissional e conta de acesso, novos mecânicos não conseguem autenticar nem podem ser identificados como responsáveis por uma Ordem de Serviço.

**Pré-condições**

- O solicitante está autenticado.
- O solicitante possui o escopo `mecanicos:escrever`.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-MEC-01 | Permitir que um mecânico autorizado cadastre outro mecânico. |
| RF-MEC-02 | Criar a conta de acesso e o cadastro profissional na mesma operação. |
| RF-MEC-03 | Permitir que o novo mecânico faça login com o e-mail e a senha cadastrados. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-MEC-01 | A senha inicial deve possuir no mínimo 15 caracteres e ser persistida somente como hash BCrypt. |
| RNF-MEC-02 | A criação deve ser atômica: não pode existir conta sem mecânico, nem mecânico sem conta. |

**Fluxo Principal**

1. O mecânico autorizado informa nome, e-mail e senha do novo profissional.
2. O sistema valida os dados e a unicidade do e-mail.
3. O sistema cria a conta ativa, o mecânico e seus escopos na mesma transação.
4. O sistema retorna os dados públicos do mecânico cadastrado.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | E-mail já cadastrado | Impede o cadastro. |
| A2 | Usuário sem escopo | Impede o cadastro. |

**Saída**

- Mecânico ativo, com conta criada e apto a autenticar.

**Pós-condições**

- A senha não é retornada nem armazenada em texto puro.
- O novo mecânico pode ser associado a uma Ordem de Serviço em tarefa futura.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /mecanicos
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Escopo: `mecanicos:escrever`.

**Entrada**

```json
{
  "nome": "Maria Souza",
  "email": "maria.souza@oficina.local",
  "senha": "senha-da-maria",
  "escopos": ["veiculos:ler", "veiculos:escrever"]
}
```

| Campo | Tipo | Descrição |
|---|---|---|
| `nome` | string | Obrigatório; nome do mecânico. |
| `email` | string | Obrigatório; identificador de login e único. |
| `senha` | string | Obrigatória; persistida somente como hash BCrypt. |
| `escopos` | array de string | Obrigatório; permissões atribuídas ao novo mecânico. |

**Validações**

- `nome`, `email`, `senha` e `escopos` são obrigatórios.
- A senha deve possuir no mínimo 15 caracteres.
- O e-mail não pode estar cadastrado.
- Cada escopo deve pertencer à lista oficial do projeto.

**Processamento**

1. Validar o JWT e o escopo `mecanicos:escrever`.
2. Validar o corpo e verificar a unicidade do e-mail.
3. Gerar o hash da senha.
4. Abrir transação.
5. Criar `usuario`, `mecanico` e os registros de `usuario_escopo`.
6. Confirmar a transação e retornar o mecânico criado.

**Persistência**

- Consulta: `usuario` e a lista de escopos válidos.
- Altera: `usuario`, `mecanico` e `usuario_escopo`.

**Saída da API**

```json
{
  "id": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
  "nome": "Maria Souza",
  "email": "maria.souza@oficina.local",
  "ativo": true,
  "escopos": ["veiculos:ler", "veiculos:escrever"]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Mecânico cadastrado. |
| `400` | Corpo inválido, campo obrigatório ausente, senha com menos de 15 caracteres ou escopo desconhecido. |
| `401` | Token ausente, inválido ou expirado. |
| `403` | Token sem `mecanicos:escrever`. |
| `409` | E-mail já cadastrado. |

**Dependências**

- Middleware HTTP de autenticação e autorização.
- `UsuarioRepository`.
- `MecanicoRepository`.

**Testes**

*Unitários*

- Gera hash de senha sem expor o valor original.
- Rejeita senha com menos de 15 caracteres.
- Rejeita e-mail duplicado e escopo desconhecido.

*Integração*

- Mecânico autorizado cadastra outro mecânico e o novo login funciona.
- Requisição sem token retorna `401`.
- Requisição sem escopo retorna `403`.
- E-mail duplicado retorna `409`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar a regra de cadastro de mecânico e conta de acesso.

**Caso de uso**

- [ ] Implementar cadastro transacional de mecânico.

**Repositório**

- [ ] Implementar `MecanicoRepository` para persistir o profissional vinculado à conta.

**Integrações**

- [ ] Integrar com Segurança para criar `usuario`, gerar o hash BCrypt e associar `usuario_escopo`.

**Handler HTTP**

- [ ] Implementar `POST /mecanicos` no contexto de Mecânico, com autenticação por escopo.

**Validações**

- [ ] Validar campos obrigatórios, senha com no mínimo 15 caracteres, e-mail único e escopos permitidos.

**Transação e idempotência**

- [ ] Garantir rollback quando uma das três gravações falhar.

**Testes unitários**

- [ ] Cobrir hash de senha, senha com menos de 15 caracteres, e-mail duplicado e escopo desconhecido.

**Testes de integração**

- [ ] Cobrir sucesso, autenticação, autorização e e-mail duplicado.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Migration versionada e reversível.
- [ ] Code Review aprovado.

