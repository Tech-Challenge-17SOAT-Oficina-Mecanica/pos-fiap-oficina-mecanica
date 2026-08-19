---
documento: Refinamento de Requisitos — Atualizar Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Atualizar Cliente

Este documento detalha a tarefa Atualizar Cliente do contexto de Cliente.

## 3 · Atualizar Cliente

### 3.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Atualizar os dados cadastrais de um cliente já existente no sistema.

**Problema**
A oficina precisa manter os dados do cliente corretos e atualizados para garantir identificação,
vínculo com veículos, criação de Ordem de Serviço e preservação do histórico. Dados incorretos
podem gerar atendimento duplicado, vínculo errado com veículo e perda de rastreabilidade.

**Pré-condições**

- O cliente deve estar cadastrado no sistema.
- O cliente deve ter sido identificado.
- O usuário deve estar autorizado a atualizar clientes.
- Os novos dados informados devem ser válidos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-CLI-14 | Permitir ao mecânico atualizar dados de um cliente cadastrado. |
| RF-CLI-15 | Consultar o cliente antes da atualização. |
| RF-CLI-16 | Validar os dados alterados. |
| RF-CLI-17 | Validar CPF/CNPJ quando esse dado for informado ou alterado. |
| RF-CLI-18 | Persistir as alterações no cadastro do cliente. |
| RF-CLI-19 | Confirmar que o cliente foi atualizado. |
| RF-CLI-20 | Manter o vínculo do cliente com seus veículos e Ordens de Serviço. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-CLI-11 | A operação deve ser feita por API RESTful. |
| RNF-CLI-12 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-CLI-13 | Dados sensíveis, como CPF/CNPJ, devem ser validados. |
| RNF-CLI-14 | A atualização deve ser persistida de forma consistente. |
| RNF-CLI-15 | A operação não deve remover o histórico do cliente. |

**Fluxo Principal**

1. O mecânico consulta o cliente.
2. O sistema identifica o cliente.
3. O mecânico solicita a atualização dos dados do cliente.
4. O mecânico informa os dados que devem ser alterados.
5. O sistema valida os dados informados.
6. O sistema atualiza o cadastro do cliente.
7. O sistema confirma que o cliente foi atualizado.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente não encontrado | O sistema informa que o cliente não existe. |
| A2 | Cliente não identificado | O sistema impede a atualização até que o cliente seja identificado. |
| A3 | CPF/CNPJ inválido | O sistema informa que o CPF/CNPJ informado não é válido. |
| A4 | Dados obrigatórios ausentes | O sistema informa quais dados precisam ser preenchidos. |
| A5 | CPF/CNPJ já vinculado a outro cliente | O sistema impede a atualização para evitar duplicidade. |
| A6 | Usuário sem autorização | O sistema impede a atualização. |
| A7 | Erro ao atualizar cliente | O sistema informa que não foi possível concluir a atualização. |

**Saída**

- Cliente atualizado no sistema.
- Confirmação de atualização realizada.

**Pós-condições**

- Os dados do cliente ficam atualizados.
- O cliente permanece vinculado aos mesmos veículos e Ordens de Serviço.
- O cliente atualizado fica disponível para consulta e continuidade do atendimento.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
PUT /api/v1/clientes/{clienteId}
```

> **Decisão de projeto.** Foi adotada a rota plural com prefixo versionado
> `PUT /api/v1/clientes/{clienteId}`, alinhada ao padrão compartilhado do projeto. A alternativa
> `PUT /clientes/{clienteId}` foi descartada por não usar o prefixo `/api/v1/`.
>
> **Decisão de projeto.** O campo do documento foi padronizado como `documento`, o mesmo usado
> nos requisitos de consulta e cadastro. A alternativa `cpfCnpj` foi descartada para evitar dois
> nomes técnicos para o mesmo dado de negócio.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfil: `MECANICO`
- Escopo: `clientes:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Path | `clienteId` | string | Identificador do cliente, obrigatório. |
| Body | `nome` | string | Nome do cliente, obrigatório. |
| Body | `documento` | string | CPF ou CNPJ do cliente, obrigatório. |
| Body | `tipoDocumento` | enum | `CPF` ou `CNPJ`, obrigatório. |

```json
{
  "nome": "Nome do Cliente",
  "documento": "00000000000",
  "tipoDocumento": "CPF"
}
```

**Validações**

- `clienteId` deve ser informado.
- Deve existir cliente cadastrado para o `clienteId`.
- `nome` deve ser informado.
- `documento` deve ser informado.
- `tipoDocumento` deve ser informado.
- `tipoDocumento` deve ser `CPF` ou `CNPJ`.
- `documento` deve possuir formato válido conforme o `tipoDocumento`.
- O CPF/CNPJ informado não pode estar vinculado a outro cliente.

**Processamento**

1. Receber o identificador do cliente.
2. Receber os dados informados para atualização.
3. Validar os campos obrigatórios.
4. Consultar o cliente pelo identificador.
5. Validar o CPF/CNPJ informado conforme o `tipoDocumento`.
6. Verificar se o CPF/CNPJ já pertence a outro cliente.
7. Atualizar os dados do cliente.
8. Persistir as alterações.
9. Retornar os dados atualizados.

**Persistência**

- Consulta: agregado/dados de `Cliente`.
- Consulta: `Cliente` por CPF/CNPJ para verificar duplicidade.
- Altera: registro de `Cliente`.
- Persiste: nome, CPF/CNPJ e tipo do documento.
- Não altera: vínculos com `Veículo` e Ordem de Serviço.

**Saída da API**

```json
{
  "id": "uuid-do-cliente",
  "nome": "Nome do Cliente",
  "documento": "00000000000",
  "tipoDocumento": "CPF"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Cliente atualizado com sucesso. |
| `400` | Identificador ausente, dados obrigatórios ausentes ou CPF/CNPJ inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `clientes:escrever`. |
| `404` | Cliente não encontrado. |
| `409` | CPF/CNPJ já vinculado a outro cliente. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de clientes.
- `ClienteRepository`.
- Validador de CPF/CNPJ.
- Caso de uso Consultar Cliente, para localizar o cliente antes da atualização.

**Testes**

*Unitários*

- Atualiza cliente quando os dados são válidos.
- Rejeita atualização quando `clienteId` não for informado.
- Rejeita atualização quando o cliente não existir.
- Rejeita atualização quando `nome` não for informado.
- Rejeita atualização quando `documento` não for informado.
- Rejeita atualização quando `tipoDocumento` não for informado.
- Rejeita atualização quando CPF/CNPJ for inválido.
- Rejeita atualização quando CPF/CNPJ pertencer a outro cliente.
- Preserva os vínculos com veículos e Ordens de Serviço.

*Integração*

- `PUT` válido retorna `200` e persiste os dados atualizados.
- Cliente inexistente retorna `404`.
- `clienteId` ausente retorna `400`.
- Nome ausente retorna `400`.
- Documento ausente retorna `400`.
- Tipo de documento ausente retorna `400`.
- CPF/CNPJ inválido retorna `400`.
- CPF/CNPJ pertencente a outro cliente retorna `409`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.
- Vínculos com veículos e Ordens de Serviço são preservados.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `Cliente`
- [ ] Definir quais campos do cliente podem ser atualizados
- [ ] Garantir que o cliente possua CPF/CNPJ como identificador de negócio
- [ ] Criar ou ajustar validação de CPF/CNPJ
- [ ] Impedir duplicidade de CPF/CNPJ entre clientes
- [ ] Garantir que vínculos existentes com veículos e Ordens de Serviço sejam preservados

**Caso de uso**

- [ ] Implementar `AtualizarCliente`
- [ ] Receber o identificador do cliente
- [ ] Receber os dados atualizados do cliente
- [ ] Consultar o cliente pelo identificador
- [ ] Verificar se o CPF/CNPJ informado já pertence a outro cliente
- [ ] Atualizar os dados do cliente
- [ ] Persistir as alterações no banco de dados

**Repositório**

- [ ] Criar ou ajustar `ClienteRepository` para busca e persistência do cliente
- [ ] Criar método para consultar cliente por identificador
- [ ] Criar método para consultar cliente por CPF/CNPJ
- [ ] Criar método para salvar alterações do cliente

**Handler HTTP**

- [ ] Implementar `PUT /api/v1/clientes/{clienteId}`
- [ ] Criar DTO/request de entrada
- [ ] Criar DTO/response de saída com os dados atualizados do cliente
- [ ] Implementar validação do parâmetro `clienteId`
- [ ] Implementar validação do payload
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `clientes:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar que o identificador do cliente foi informado
- [ ] Validar que o cliente existe
- [ ] Validar que o nome do cliente foi informado
- [ ] Validar que o CPF/CNPJ foi informado
- [ ] Validar que `tipoDocumento` foi informado
- [ ] Validar formato do CPF/CNPJ
- [ ] Retornar `200` quando o cliente for atualizado com sucesso
- [ ] Retornar `400` para identificador ausente, dados obrigatórios ausentes ou CPF/CNPJ inválido
- [ ] Retornar `404` quando o cliente não for encontrado
- [ ] Retornar `409` quando o CPF/CNPJ pertencer a outro cliente
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Testes unitários**

- [ ] Atualização válida de cliente
- [ ] Identificador ausente
- [ ] Cliente inexistente
- [ ] Nome ausente
- [ ] CPF/CNPJ ausente
- [ ] `tipoDocumento` ausente
- [ ] CPF/CNPJ inválido
- [ ] CPF/CNPJ já vinculado a outro cliente
- [ ] Preservação dos vínculos com veículos e Ordens de Serviço

**Testes de integração**

- [ ] Endpoint atualiza cliente válido e retorna `200`
- [ ] Dados atualizados são persistidos no banco
- [ ] Endpoint retorna `400` para identificador ausente, dados obrigatórios ausentes ou CPF/CNPJ inválido
- [ ] Endpoint retorna `404` quando o cliente não existe
- [ ] Endpoint retorna `409` quando CPF/CNPJ pertence a outro cliente
- [ ] Endpoint retorna `401` sem autenticação
- [ ] Endpoint retorna `403` sem permissão
- [ ] Vínculos com veículos e Ordens de Serviço são preservados após a atualização

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI
- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite da task
- [ ] Code Review aprovado

---


