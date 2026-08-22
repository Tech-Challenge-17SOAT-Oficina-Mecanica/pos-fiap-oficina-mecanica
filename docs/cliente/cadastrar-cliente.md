---
documento: Refinamento de Requisitos — Cadastrar Cliente
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Cliente

Este documento detalha a tarefa Cadastrar Cliente do contexto de Cliente.

## 2 · Cadastrar Cliente

### 2.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Cadastrar um cliente no sistema quando ele não for encontrado na consulta por CPF/CNPJ.

**Problema**
A oficina precisa manter o histórico de clientes e permitir que o cliente seja identificado e
vinculado corretamente ao veículo e à Ordem de Serviço. Sem cadastro, o atendimento perde
rastreabilidade e o histórico do cliente fica fragmentado ou inexistente.

**Pré-condições**

- O cliente deve ter informado CPF/CNPJ.
- O cliente não deve ter sido encontrado na consulta de cadastro.
- O CPF/CNPJ deve ser válido.
- O usuário deve estar autorizado a cadastrar clientes.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-CLI-08 | Permitir ao mecânico cadastrar um novo cliente. |
| RF-CLI-09 | Validar o CPF/CNPJ informado. |
| RF-CLI-10 | Registrar os dados do cliente. |
| RF-CLI-11 | Impedir cadastro duplicado para o mesmo CPF/CNPJ. |
| RF-CLI-12 | Confirmar que o cliente foi cadastrado. |
| RF-CLI-13 | Permitir que o cliente cadastrado seja usado nas próximas etapas do atendimento. |
| RF-CLI-36 | Registrar ao menos um contato do cliente, telefone ou e-mail. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-CLI-06 | A operação deve ser feita por API RESTful. |
| RNF-CLI-07 | Os dados sensíveis do cliente devem ser validados. |
| RNF-CLI-08 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-CLI-09 | O cadastro deve ser persistido de forma consistente. |
| RNF-CLI-10 | O sistema deve evitar duplicidade de clientes. |
| RNF-CLI-28 | O contato registrado deve permitir avisar o cliente sobre orçamento e conclusão do serviço. |

**Fluxo Principal**

1. O mecânico consulta o cliente pelo CPF/CNPJ.
2. O sistema informa que o cliente não foi encontrado.
3. O mecânico solicita o cadastro do cliente.
4. O sistema valida o CPF/CNPJ informado.
5. O mecânico informa os dados necessários do cliente.
6. O sistema registra o novo cliente.
7. O sistema confirma que o cliente foi cadastrado.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | CPF/CNPJ inválido | O sistema informa que o CPF/CNPJ informado não é válido e não cadastra o cliente. |
| A2 | Cliente já cadastrado | O sistema impede novo cadastro para o mesmo CPF/CNPJ. |
| A3 | Dados obrigatórios ausentes | O sistema informa quais dados precisam ser preenchidos e não cadastra o cliente. |
| A6 | Nenhum contato informado | O sistema informa que é preciso informar telefone ou e-mail e não cadastra o cliente. |
| A7 | Contato em formato inválido | O sistema informa qual contato está inválido e não cadastra o cliente. |
| A4 | Usuário sem autorização | O sistema impede o cadastro. |
| A5 | Erro ao cadastrar cliente | O sistema informa que não foi possível concluir o cadastro. |

**Saída**

- Cliente cadastrado no sistema.
- Confirmação de cadastro realizado.

**Pós-condições**

- O cliente passa a existir no cadastro da oficina.
- O cliente fica disponível para consulta.
- O cliente pode ser vinculado a um veículo.
- O fluxo pode seguir para Consultar Veículo ou Cadastrar Veículo.

---

### 2.2 Refinamento Técnico

**Endpoint**

```http
POST /clientes
```

> **Decisão de projeto.** A rota segue o padrão compartilhado do projeto: recurso no plural,
> em minúsculas e **sem prefixo de versão**. A alternativa com prefixo `/api/v1` foi descartada
> para manter todas as rotas do sistema no mesmo formato.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfil: `MECANICO`
- Escopo: `clientes:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Body | `nome` | string | Nome do cliente, obrigatório. |
| Body | `documento` | string | CPF ou CNPJ do cliente, obrigatório. |
| Body | `tipoDocumento` | enum | `CPF` ou `CNPJ`, obrigatório. |
| Body | `telefone` | string | Telefone do cliente, somente dígitos. Obrigatório se `email` não for informado. |
| Body | `email` | string | E-mail do cliente. Obrigatório se `telefone` não for informado. |

```json
{
  "nome": "Nome do Cliente",
  "documento": "00000000000",
  "tipoDocumento": "CPF",
  "telefone": "11988887777",
  "email": "cliente@exemplo.com"
}
```

> **Decisão de projeto.** O cliente passa a ter contato no cadastro, com **pelo menos um** entre
> `telefone` e `email` obrigatório. Sem isso não há como avisar o cliente sobre o orçamento e a
> conclusão do serviço, que o fluxo do negócio prevê.

**Validações**

- `nome` deve ser informado.
- `documento` deve ser informado.
- `tipoDocumento` deve ser informado.
- `tipoDocumento` deve ser `CPF` ou `CNPJ`.
- `documento` deve possuir formato válido conforme o `tipoDocumento`.
- Não pode existir cliente cadastrado com o mesmo CPF/CNPJ.
- Pelo menos um entre `telefone` e `email` deve ser informado.
- `telefone`, quando informado, deve ter 10 ou 11 dígitos.
- `email`, quando informado, deve ter formato válido.

**Processamento**

1. Receber os dados do cliente.
2. Validar os campos obrigatórios.
3. Validar o CPF/CNPJ informado conforme o `tipoDocumento`.
4. Consultar se já existe cliente com o mesmo CPF/CNPJ.
5. Criar o cadastro do cliente.
6. Persistir o novo cliente.
7. Retornar os dados do cliente cadastrado.

**Persistência**

- Consulta: agregado/dados de `Cliente` para verificar duplicidade.
- Altera: `Cliente` com novo registro.
- Persiste: identificador do cliente, nome, CPF/CNPJ, tipo do documento e contato.

**Saída da API**

```json
{
  "id": "uuid-do-cliente",
  "nome": "Nome do Cliente",
  "documento": "00000000000",
  "tipoDocumento": "CPF",
  "telefone": "11988887777",
  "email": "cliente@exemplo.com"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Cliente cadastrado com sucesso. |
| `400` | Dados obrigatórios ausentes, CPF/CNPJ inválido, nenhum contato informado ou contato em formato inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `clientes:escrever`. |
| `409` | Cliente já cadastrado com o CPF/CNPJ informado. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de clientes.
- `ClienteRepository`.
- Validador de CPF/CNPJ.
- Caso de uso Consultar Cliente, para verificar se o cliente já existe.

**Testes**

*Unitários*

- Cadastra cliente quando os dados são válidos.
- Rejeita cadastro quando `nome` não for informado.
- Rejeita cadastro quando `documento` não for informado.
- Rejeita cadastro quando `tipoDocumento` não for informado.
- Rejeita cadastro quando CPF/CNPJ for inválido.
- Rejeita cadastro quando já existir cliente com o mesmo CPF/CNPJ.
- Rejeita cadastro quando nenhum contato for informado.
- Rejeita cadastro quando o telefone tiver formato inválido.
- Rejeita cadastro quando o e-mail tiver formato inválido.
- Cadastra cliente informando apenas telefone.
- Cadastra cliente informando apenas e-mail.

*Integração*

- `POST` válido retorna `201` e persiste o cliente.
- Cliente cadastrado pode ser consultado por CPF/CNPJ.
- Nome ausente retorna `400`.
- Documento ausente retorna `400`.
- Tipo de documento ausente retorna `400`.
- CPF/CNPJ inválido retorna `400`.
- Cliente duplicado retorna `409`.
- Cadastro sem telefone e sem e-mail retorna `400`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `Cliente`
- [ ] Definir os campos necessários para cadastro do cliente, incluindo `telefone` e `email`
- [ ] Garantir que o cliente possua CPF/CNPJ como identificador de negócio
- [ ] Criar validação de CPF/CNPJ
- [ ] Impedir cadastro duplicado de cliente

**Caso de uso**

- [ ] Implementar `CadastrarCliente`
- [ ] Receber os dados necessários do cliente
- [ ] Verificar se já existe cliente cadastrado com o mesmo CPF/CNPJ
- [ ] Criar novo cliente
- [ ] Persistir o cliente no banco de dados

**Repositório**

- [ ] Criar ou ajustar `ClienteRepository` para persistência do cliente
- [ ] Criar método para consultar cliente por CPF/CNPJ
- [ ] Criar método para salvar novo cliente

**Handler HTTP**

- [ ] Implementar `POST /clientes`
- [ ] Criar DTO/request de entrada
- [ ] Criar DTO/response de saída com os dados do cliente cadastrado
- [ ] Implementar validação do payload
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `clientes:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar que o nome do cliente foi informado
- [ ] Validar que o CPF/CNPJ foi informado
- [ ] Validar que `tipoDocumento` foi informado
- [ ] Validar que `tipoDocumento` é `CPF` ou `CNPJ`
- [ ] Validar formato do CPF/CNPJ
- [ ] Validar que ao menos um entre `telefone` e `email` foi informado
- [ ] Validar o formato do telefone e do e-mail quando informados
- [ ] Retornar `201` quando o cliente for cadastrado com sucesso
- [ ] Retornar `400` para dados obrigatórios ausentes ou CPF/CNPJ inválido
- [ ] Retornar `409` quando já existir cliente com o mesmo CPF/CNPJ
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Testes unitários**

- [ ] Cadastro válido de cliente
- [ ] Nome ausente
- [ ] CPF/CNPJ ausente
- [ ] `tipoDocumento` ausente
- [ ] CPF/CNPJ inválido
- [ ] Cliente duplicado
- [ ] Cadastro sem nenhum contato informado
- [ ] Contato em formato inválido

**Testes de integração**

- [ ] Endpoint cadastra cliente válido e retorna `201`
- [ ] Cliente é persistido corretamente no cadastro
- [ ] Cliente cadastrado pode ser consultado por CPF/CNPJ
- [ ] Endpoint retorna `400` para dados obrigatórios ausentes ou CPF/CNPJ inválido
- [ ] Endpoint retorna `409` para cliente duplicado
- [ ] Endpoint retorna `401` sem autenticação
- [ ] Endpoint retorna `403` sem permissão

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI
- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite da task
- [ ] Code Review aprovado

---


