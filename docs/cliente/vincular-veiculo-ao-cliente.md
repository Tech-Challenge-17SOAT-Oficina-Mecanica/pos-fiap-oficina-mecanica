---
documento: Refinamento de Requisitos — Vincular Veículo ao Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Vincular Veículo ao Cliente

Este documento detalha a tarefa Vincular Veículo ao Cliente do contexto de Cliente.

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
POST /clientes/{clienteId}/veiculos/{veiculoId}
```

> **Decisão de projeto.** Foi adotada a rota subordinada ao cliente porque o caso de uso altera
> o vínculo observado a partir do cadastro do cliente. A alternativa seria
> `POST /veiculos/{veiculoId}/clientes/{clienteId}`, mas ela desloca a ação para o
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

- [ ] Implementar `POST /clientes/{clienteId}/veiculos/{veiculoId}`
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


