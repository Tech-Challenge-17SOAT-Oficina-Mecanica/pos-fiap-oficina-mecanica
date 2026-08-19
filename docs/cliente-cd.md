---
documento: Refinamento de Requisitos — Contexto de Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos

Este documento reúne, para cada requisito levantado da aplicação, três blocos:

1. **Refinamento de Produto** — o que o usuário precisa e por quê (visão de negócio).
2. **Refinamento Técnico** — como o sistema entrega isso (contrato, processamento, testes).
3. **Checklist de Implementação** — o passo a passo verificável até o merge.

---

## 1 · Consultar Cliente

### 1.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Consultar o cadastro do cliente a partir do CPF/CNPJ informado, para identificar o cliente antes
de seguir com o atendimento.

**Problema**
A oficina precisa identificar o cliente antes de seguir com o atendimento. Sem essa consulta, há
risco de perda de histórico de clientes, vínculo incorreto com veículo e abertura de Ordem de
Serviço sem rastreabilidade adequada.

**Pré-condições**

- O cliente deve informar CPF/CNPJ.
- O usuário deve estar autorizado a consultar clientes.
- Deve existir cadastro de clientes disponível para consulta.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-CLI-01 | Permitir ao mecânico consultar cliente por CPF/CNPJ. |
| RF-CLI-02 | Validar o CPF/CNPJ informado. |
| RF-CLI-03 | Consultar o cadastro do cliente. |
| RF-CLI-04 | Identificar o cliente quando houver cadastro correspondente. |
| RF-CLI-05 | Informar quando o cliente não for encontrado. |
| RF-CLI-06 | Permitir seguir para cadastro do cliente quando ele não for identificado. |
| RF-CLI-07 | Retornar os veículos vinculados ao cliente encontrado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-CLI-01 | A consulta deve ser feita por API RESTful. |
| RNF-CLI-02 | O CPF/CNPJ deve ser validado por se tratar de dado sensível. |
| RNF-CLI-03 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-CLI-04 | A consulta não deve alterar os dados cadastrais do cliente. |
| RNF-CLI-05 | A consulta não deve alterar os dados dos veículos vinculados ao cliente. |

**Fluxo Principal**

1. O cliente informa CPF/CNPJ.
2. O mecânico solicita a consulta do cliente.
3. O sistema valida o CPF/CNPJ informado.
4. O sistema consulta o cadastro de clientes.
5. O sistema encontra o cliente correspondente.
6. O sistema consulta os veículos vinculados ao cliente.
7. O sistema identifica o cliente e apresenta seus dados cadastrais.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | CPF/CNPJ inválido | O sistema informa que o CPF/CNPJ informado não é válido e não consulta o cadastro. |
| A2 | Cliente não encontrado | O sistema informa que o cliente não foi encontrado. |
| A3 | Cliente não identificado | O sistema permite seguir para o cadastro do cliente. |
| A4 | Cliente encontrado sem veículo vinculado | O sistema identifica o cliente e retorna a lista de veículos vazia. |
| A5 | Usuário sem autorização | O sistema impede a consulta. |

**Saída**

- Cliente identificado a partir do CPF/CNPJ informado, com seus veículos vinculados; **ou**
- Indicação de que o cliente não foi encontrado.

**Pós-condições**

- Os dados do cliente permanecem inalterados.
- Os dados dos veículos vinculados permanecem inalterados.
- O cliente identificado fica disponível para as próximas etapas do atendimento.
- Caso o cliente não seja encontrado, o fluxo pode seguir para Cadastrar Cliente.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
GET /api/v1/clientes
```

Consulta por CPF/CNPJ via query param.

> **Decisão de projeto.** Foi adotada a rota plural com prefixo versionado
> `GET /api/v1/clientes?documento=...`, alinhada ao padrão compartilhado do projeto. A
> alternativa `GET /clientes` foi descartada por não usar o prefixo `/api/v1/`.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfil: `MECANICO`
- Escopo: `clientes:ler`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Query | `documento` | string | CPF ou CNPJ do cliente, obrigatório. |

**Validações**

- `documento` deve ser informado.
- `documento` deve possuir formato válido de CPF ou CNPJ.
- O cliente deve existir para que a consulta retorne sucesso.
- A operação não altera dados do cliente.
- A operação não altera dados dos veículos vinculados ao cliente.

**Processamento**

1. Receber o CPF/CNPJ informado no query param `documento`.
2. Validar presença e formato do documento.
3. Consultar o cadastro de clientes pelo CPF/CNPJ.
4. Caso encontre o cliente, consultar os veículos vinculados.
5. Montar a resposta com os dados cadastrais do cliente e seus veículos.
6. Caso não encontre, retornar erro informando que o cliente não foi encontrado.

**Persistência**

- Consulta: agregado/dados de `Cliente`.
- Consulta: vínculo entre `Cliente` e `Veículo`.
- Consulta: dados dos `Veículos` vinculados.
- Altera: nada.

**Saída da API**

```json
{
  "id": "uuid-do-cliente",
  "nome": "Nome do Cliente",
  "documento": "00000000000",
  "veiculos": [
    {
      "id": "uuid-do-veiculo",
      "placa": "ABC1D23",
      "marca": "Marca do Veículo",
      "modelo": "Modelo do Veículo",
      "ano": 2020
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Cliente encontrado. |
| `400` | CPF/CNPJ ausente ou inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `clientes:ler`. |
| `404` | Cliente não encontrado. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de clientes.
- Módulo de veículos.
- `ClienteRepository`.
- `VeiculoRepository`.
- Repositório de vínculo entre cliente e veículo.
- Validador de CPF/CNPJ.
- Caso de uso Cadastrar Cliente, quando o cliente não for encontrado.

**Testes**

*Unitários*

- Retorna cliente quando CPF/CNPJ válido existe no cadastro.
- Retorna cliente com lista de veículos vazia quando não houver veículo vinculado.
- Retorna erro quando CPF/CNPJ não for informado.
- Retorna erro quando CPF/CNPJ for inválido.
- Retorna erro quando cliente não for encontrado.
- Garante que a consulta não altera os dados do cliente.
- Garante que a consulta não altera os dados dos veículos vinculados.

*Integração*

- `GET` válido retorna `200` com os dados esperados do cliente.
- Cliente com veículos vinculados retorna a lista de veículos.
- Cliente sem veículos vinculados retorna `veiculos: []`.
- Cliente inexistente retorna `404`.
- CPF/CNPJ ausente retorna `400`.
- CPF/CNPJ inválido retorna `400`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `Cliente`
- [ ] Definir os campos necessários para identificação do cliente por CPF/CNPJ
- [ ] Garantir que a consulta não altera dados do cliente
- [ ] Garantir que a consulta não altera dados dos veículos vinculados

**Caso de uso**

- [ ] Implementar `ConsultarCliente`
- [ ] Receber o CPF/CNPJ como critério de consulta
- [ ] Consultar o cliente cadastrado pelo CPF/CNPJ
- [ ] Consultar os veículos vinculados ao cliente encontrado
- [ ] Retornar cliente com `veiculos: []` quando não houver veículo vinculado

**Repositório**

- [ ] Criar ou ajustar `ClienteRepository` para consulta por CPF/CNPJ
- [ ] Implementar método de busca de cliente por CPF/CNPJ
- [ ] Criar ou ajustar `VeiculoRepository` para consulta por cliente
- [ ] Implementar consulta do vínculo entre cliente e veículo

**Handler HTTP**

- [ ] Implementar `GET /api/v1/clientes`
- [ ] Implementar leitura do parâmetro `documento` via query param
- [ ] Criar DTO/response de saída com os dados do cliente e veículos
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `clientes:ler`
- [ ] Mapear erros de validação para os códigos HTTP documentados

**Validações**

- [ ] Validar que `documento` foi informado
- [ ] Validar formato de CPF/CNPJ
- [ ] Retornar `400` para CPF/CNPJ ausente ou inválido
- [ ] Retornar `404` quando o cliente não for encontrado
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Testes unitários**

- [ ] Consulta de cliente existente
- [ ] CPF/CNPJ ausente
- [ ] CPF/CNPJ inválido
- [ ] Cliente não encontrado
- [ ] Cliente existente sem veículo vinculado
- [ ] Consulta não altera dados do cliente nem dos veículos vinculados

**Testes de integração**

- [ ] Endpoint retorna os dados esperados do cliente
- [ ] Endpoint retorna os veículos vinculados ao cliente
- [ ] Endpoint retorna `veiculos: []` quando não houver veículo vinculado
- [ ] Endpoint retorna `400` para CPF/CNPJ ausente ou inválido
- [ ] Endpoint retorna `404` quando o cliente não existe
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

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-CLI-06 | A operação deve ser feita por API RESTful. |
| RNF-CLI-07 | Os dados sensíveis do cliente devem ser validados. |
| RNF-CLI-08 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-CLI-09 | O cadastro deve ser persistido de forma consistente. |
| RNF-CLI-10 | O sistema deve evitar duplicidade de clientes. |

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
POST /api/v1/clientes
```

> **Decisão de projeto.** Foi adotada a rota plural com prefixo versionado
> `POST /api/v1/clientes`, alinhada ao padrão compartilhado do projeto. A alternativa
> `POST /clientes` foi descartada por não usar o prefixo `/api/v1/`.

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

```json
{
  "nome": "Nome do Cliente",
  "documento": "00000000000",
  "tipoDocumento": "CPF"
}
```

**Validações**

- `nome` deve ser informado.
- `documento` deve ser informado.
- `tipoDocumento` deve ser informado.
- `tipoDocumento` deve ser `CPF` ou `CNPJ`.
- `documento` deve possuir formato válido conforme o `tipoDocumento`.
- Não pode existir cliente cadastrado com o mesmo CPF/CNPJ.

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
- Persiste: identificador do cliente, nome, CPF/CNPJ e tipo do documento.

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
| `201` | Cliente cadastrado com sucesso. |
| `400` | Dados obrigatórios ausentes ou CPF/CNPJ inválido. |
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

*Integração*

- `POST` válido retorna `201` e persiste o cliente.
- Cliente cadastrado pode ser consultado por CPF/CNPJ.
- Nome ausente retorna `400`.
- Documento ausente retorna `400`.
- Tipo de documento ausente retorna `400`.
- CPF/CNPJ inválido retorna `400`.
- Cliente duplicado retorna `409`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `Cliente`
- [ ] Definir os campos necessários para cadastro do cliente
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

- [ ] Implementar `POST /api/v1/clientes`
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

## 4 · Vincular Veículo ao Cliente

### 4.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Vincular um veículo cadastrado a um cliente identificado no sistema.

**Problema**
A oficina precisa associar corretamente cliente e veículo para permitir a criação da Ordem de
Serviço e manter o histórico de atendimentos por cliente e por veículo. Sem esse vínculo, a
Ordem de Serviço pode ser aberta para a combinação errada de cliente e veículo.

**Pré-condições**

- O cliente deve estar cadastrado.
- O cliente deve ter sido identificado.
- O veículo deve estar cadastrado.
- O veículo deve ter sido identificado.
- O usuário deve estar autorizado a atualizar o vínculo entre cliente e veículo.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-CLI-21 | Permitir ao mecânico vincular um veículo a um cliente. |
| RF-CLI-22 | Validar se o cliente existe. |
| RF-CLI-23 | Validar se o veículo existe. |
| RF-CLI-24 | Associar o veículo ao cadastro do cliente. |
| RF-CLI-25 | Atualizar o cadastro do cliente com o vínculo do veículo. |
| RF-CLI-26 | Confirmar que o vínculo foi realizado. |
| RF-CLI-27 | Disponibilizar o vínculo para criação da Ordem de Serviço. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-CLI-16 | A operação deve ser feita por API RESTful. |
| RNF-CLI-17 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-CLI-18 | A atualização deve ser persistida de forma consistente. |
| RNF-CLI-19 | O vínculo entre cliente e veículo deve preservar o histórico de atendimentos. |
| RNF-CLI-20 | A operação não deve alterar indevidamente os dados cadastrais do cliente ou do veículo. |

**Fluxo Principal**

1. O mecânico consulta o cliente.
2. O sistema identifica o cliente.
3. O mecânico consulta o veículo.
4. O sistema identifica o veículo.
5. O mecânico solicita o vínculo entre cliente e veículo.
6. O sistema valida se o cliente existe.
7. O sistema valida se o veículo existe.
8. O sistema verifica se o vínculo ainda não existe.
9. O sistema vincula o veículo ao cliente.
10. O sistema atualiza o cliente.
11. O sistema confirma que o vínculo foi realizado.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente não encontrado | O sistema informa que o cliente não existe. |
| A2 | Cliente não identificado | O sistema impede o vínculo até que o cliente seja identificado. |
| A3 | Veículo não identificado | O sistema impede o vínculo até que o veículo seja identificado. |
| A4 | Veículo não encontrado | O sistema informa que o veículo não existe. |
| A5 | Vínculo já existente | O sistema informa que o veículo já está vinculado ao cliente. |
| A6 | Usuário sem autorização | O sistema impede a operação. |
| A7 | Erro ao vincular veículo | O sistema informa que não foi possível concluir o vínculo. |

**Saída**

- Veículo vinculado ao cliente.
- Cliente atualizado com o vínculo do veículo.

**Pós-condições**

- O cliente permanece cadastrado no sistema.
- O veículo permanece cadastrado no sistema.
- O veículo fica associado ao cliente.
- O vínculo fica disponível para criação da Ordem de Serviço.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
POST /api/v1/clientes/{clienteId}/veiculos/{veiculoId}
```

> **Decisão de projeto.** Foi adotada a rota subordinada ao cliente porque o caso de uso altera
> o vínculo observado a partir do cadastro do cliente. A alternativa seria
> `POST /api/v1/veiculos/{veiculoId}/clientes/{clienteId}`, mas ela desloca a ação para o
> contexto de Veículo sem mudar a regra de negócio.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfil: `MECANICO`
- Escopo: `clientes:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Path | `clienteId` | string | Identificador do cliente, obrigatório. |
| Path | `veiculoId` | string | Identificador do veículo, obrigatório. |

A operação não recebe corpo.

**Validações**

- `clienteId` deve ser informado e possuir formato válido.
- `veiculoId` deve ser informado e possuir formato válido.
- O cliente deve existir.
- O veículo deve existir.
- O veículo não pode já estar vinculado ao cliente.

**Processamento**

1. Receber o identificador do cliente.
2. Receber o identificador do veículo.
3. Validar os parâmetros de rota.
4. Consultar o cliente pelo identificador.
5. Consultar o veículo pelo identificador.
6. Verificar se já existe vínculo entre cliente e veículo.
7. Criar o vínculo entre cliente e veículo.
8. Atualizar o cliente com o veículo vinculado.
9. Retornar a confirmação do vínculo.

**Persistência**

- Consulta: agregado/dados de `Cliente`.
- Consulta: agregado/dados de `Veículo`.
- Consulta: vínculo entre `Cliente` e `Veículo`.
- Altera: vínculo entre `Cliente` e `Veículo`.
- Atualiza: cadastro do cliente com o veículo associado.
- Não altera: dados cadastrais do veículo.

**Saída da API**

```json
{
  "clienteId": "uuid-do-cliente",
  "veiculoId": "uuid-do-veiculo",
  "vinculado": true
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Vínculo realizado com sucesso. |
| `400` | `clienteId` ou `veiculoId` ausente ou inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `clientes:escrever`. |
| `404` | Cliente ou veículo não encontrado. |
| `409` | Veículo já vinculado ao cliente. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de clientes.
- Módulo de veículos.
- `ClienteRepository`.
- `VeiculoRepository`.
- Caso de uso Consultar Cliente.
- Caso de uso Consultar Veículo.

**Testes**

*Unitários*

- Vincula cliente e veículo quando ambos existem.
- Rejeita vínculo quando o cliente não existir.
- Rejeita vínculo quando o veículo não existir.
- Rejeita vínculo quando `clienteId` não for informado ou for inválido.
- Rejeita vínculo quando `veiculoId` não for informado ou for inválido.
- Rejeita vínculo duplicado entre o mesmo cliente e veículo.

*Integração*

- `POST` válido retorna `200` e persiste o vínculo.
- Cliente inexistente retorna `404`.
- Veículo inexistente retorna `404`.
- `clienteId` ausente ou inválido retorna `400`.
- `veiculoId` ausente ou inválido retorna `400`.
- Vínculo duplicado retorna `409`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.
- Veículo vinculado aparece no cadastro do cliente.
- Vínculo pode ser usado na criação da Ordem de Serviço.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar relacionamento entre `Cliente` e `Veículo`
- [ ] Definir como o vínculo entre `Cliente` e `Veículo` será representado no domínio
- [ ] Garantir que `Cliente` e `Veículo` possam ser associados sem perder histórico
- [ ] Criar método de domínio para vincular veículo ao cliente
- [ ] Impedir duplicidade do vínculo entre o mesmo cliente e veículo

**Caso de uso**

- [ ] Implementar `VincularVeiculoAoCliente`
- [ ] Receber `clienteId` e `veiculoId`
- [ ] Validar que o cliente existe
- [ ] Validar que o veículo existe
- [ ] Validar que o vínculo ainda não existe
- [ ] Atualizar o cliente com o veículo vinculado
- [ ] Persistir o vínculo entre cliente e veículo

**Repositório**

- [ ] Criar ou ajustar `ClienteRepository` para busca e persistência do cliente
- [ ] Criar ou ajustar `VeiculoRepository` para busca do veículo
- [ ] Criar método para consultar cliente por identificador
- [ ] Criar método para consultar veículo por identificador
- [ ] Criar método para verificar se o veículo já está vinculado ao cliente
- [ ] Criar método para persistir vínculo entre cliente e veículo

**Handler HTTP**

- [ ] Implementar `POST /api/v1/clientes/{clienteId}/veiculos/{veiculoId}`
- [ ] Criar DTO/response de saída com confirmação do vínculo
- [ ] Implementar validação do parâmetro `clienteId`
- [ ] Implementar validação do parâmetro `veiculoId`
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `clientes:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar que `clienteId` foi informado
- [ ] Validar que `veiculoId` foi informado
- [ ] Validar formato de `clienteId`
- [ ] Validar formato de `veiculoId`
- [ ] Retornar `200` quando o vínculo for realizado com sucesso
- [ ] Retornar `400` para `clienteId` ou `veiculoId` ausente ou inválido
- [ ] Retornar `404` quando o cliente não existir
- [ ] Retornar `404` quando o veículo não existir
- [ ] Retornar `409` quando o vínculo já existir
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Transação e idempotência**

- [ ] Persistir o vínculo em uma única transação
- [ ] Garantir rollback integral quando a persistência falhar

**Testes unitários**

- [ ] Vínculo válido entre cliente e veículo
- [ ] Cliente inexistente
- [ ] Veículo inexistente
- [ ] Vínculo duplicado
- [ ] `clienteId` ausente ou inválido
- [ ] `veiculoId` ausente ou inválido

**Testes de integração**

- [ ] Endpoint vincula cliente e veículo e retorna `200`
- [ ] Vínculo é persistido no banco
- [ ] Veículo vinculado aparece no cadastro do cliente
- [ ] Vínculo pode ser usado na criação da Ordem de Serviço
- [ ] Endpoint retorna `400` para `clienteId` ou `veiculoId` ausente ou inválido
- [ ] Endpoint retorna `404` quando o cliente não existe
- [ ] Endpoint retorna `404` quando o veículo não existe
- [ ] Endpoint retorna `409` quando o vínculo já existe
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

## Pontos em aberto

| # | Ponto | Responsável |
|---|---|---|
| 1 | Confirmar o dono do documento do contexto Cliente. | — |
| 2 | Confirmar se o escopo definitivo para consulta de clientes será `clientes:ler` ou outro nome padronizado pelo time. | — |
| 3 | Definir se a resposta de consulta de cliente deve usar envelope simples, como documentado aqui, ou algum envelope padronizado para recursos únicos. | — |
| 4 | Confirmar se o escopo definitivo para cadastro de clientes será `clientes:escrever` ou outro nome padronizado pelo time. | — |
| 5 | Confirmar se o escopo definitivo para atualização de clientes também será `clientes:escrever` ou se haverá escopo específico. | — |
| 6 | Confirmar se atualização de cliente deve usar controle otimista com header `If-Match`, como outras operações de escrita do projeto. | — |
| 7 | Confirmar se o vínculo entre cliente e veículo pertence definitivamente ao contexto Cliente ou se deve ficar no contexto Veículo. | — |
| 8 | Confirmar se vincular veículo ao cliente deve retornar `200` com confirmação, como documentado aqui, ou `201` por criar um vínculo novo. | — |
| 9 | Confirmar se o vínculo entre cliente e veículo deve usar controle otimista com header `If-Match`. | — |
