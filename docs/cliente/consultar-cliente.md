---
documento: Refinamento de Requisitos — Consultar Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Cliente

Este documento detalha a tarefa Consultar Cliente do contexto de Cliente.

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


