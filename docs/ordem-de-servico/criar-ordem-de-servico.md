---
documento: Refinamento de Requisitos — Criar Ordem de Serviço
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Criar Ordem de Serviço

Este documento detalha a tarefa Criar Ordem de Serviço do contexto de Ordem de Serviço.

## 3 · Criar Ordem de Serviço

### 3.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Criar uma Ordem de Serviço para registrar o atendimento de um veículo vinculado a um cliente.

**Problema**

A oficina precisa organizar o atendimento, diagnóstico, execução de serviços e entrega dos
veículos, evitando anotações manuais, perda de histórico e dificuldade de acompanhamento do
status dos serviços.

**Pré-condições**

- O cliente deve estar cadastrado e identificado.
- O veículo deve estar cadastrado e identificado.
- O veículo deve estar vinculado ao cliente.
- O problema do veículo deve ter sido relatado.
- O usuário deve estar autorizado a criar Ordem de Serviço.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-12 | Permitir ao mecânico criar uma Ordem de Serviço. |
| RF-OS-13 | Associar a Ordem de Serviço ao cliente. |
| RF-OS-14 | Associar a Ordem de Serviço ao veículo. |
| RF-OS-15 | Registrar o problema relatado. |
| RF-OS-16 | Criar a Ordem de Serviço com status inicial `RECEBIDA`. |
| RF-OS-17 | Atualizar o status da Ordem de Serviço conforme a etapa do atendimento. |
| RF-OS-18 | Permitir que a Ordem de Serviço seja acompanhada posteriormente. |
| RF-OS-19 | Associar serviços solicitados, quando informados. |
| RF-OS-20 | Associar peças e insumos necessários, quando informados. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-10 | A operação deve ser feita por API RESTful. |
| RNF-OS-11 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-OS-12 | A criação da Ordem de Serviço deve ser persistida de forma consistente. |
| RNF-OS-13 | O vínculo entre cliente, veículo e Ordem de Serviço deve preservar o histórico de atendimento. |
| RNF-OS-14 | A alteração de status deve ocorrer de forma consistente. |

**Fluxo Principal**

1. O cliente informa CPF/CNPJ.
2. O mecânico consulta o cliente.
3. O sistema identifica o cliente.
4. O cliente informa a placa do veículo.
5. O mecânico consulta o veículo.
6. O sistema identifica o veículo.
7. O sistema valida o vínculo entre cliente e veículo.
8. O cliente relata o problema do veículo.
9. O mecânico solicita a criação da Ordem de Serviço.
10. O sistema valida serviços solicitados, quando informados.
11. O sistema valida peças e insumos, quando informados.
12. O sistema cria a Ordem de Serviço.
13. O sistema registra o problema relatado.
14. O sistema associa cliente e veículo à Ordem de Serviço.
15. O sistema associa serviços solicitados, quando informados.
16. O sistema associa peças e insumos necessários, quando informados.
17. O sistema define o status inicial da Ordem de Serviço.
18. O sistema confirma a criação da Ordem de Serviço.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente não encontrado | O sistema informa que o cliente não foi encontrado. |
| A2 | Cliente não identificado | O sistema impede a criação da Ordem de Serviço até que o cliente seja identificado. |
| A3 | Veículo não identificado | O sistema impede a criação da Ordem de Serviço até que o veículo seja identificado. |
| A4 | Veículo não encontrado | O sistema informa que o veículo não existe. |
| A5 | Veículo não vinculado ao cliente | O sistema impede a criação da Ordem de Serviço até que o vínculo seja realizado. |
| A6 | Problema não informado | O sistema impede a criação da Ordem de Serviço. |
| A7 | Serviço solicitado não encontrado | O sistema impede a criação da Ordem de Serviço e informa o serviço inválido. |
| A8 | Peça ou insumo não encontrado | O sistema impede a criação da Ordem de Serviço e informa o item inválido. |
| A9 | Quantidade de peça ou insumo inválida | O sistema impede a criação da Ordem de Serviço e informa que a quantidade deve ser maior que zero. |
| A10 | Usuário sem autorização | O sistema impede a operação. |
| A11 | Erro ao criar Ordem de Serviço | O sistema informa que não foi possível concluir a criação. |

**Saída**

- Ordem de Serviço criada.
- Ordem de Serviço vinculada ao cliente e ao veículo.
- Status inicial da Ordem de Serviço registrado.

**Pós-condições**

- A Ordem de Serviço passa a existir no sistema.
- A Ordem de Serviço fica vinculada ao cliente e ao veículo.
- O problema relatado fica registrado.
- Serviços solicitados ficam associados à Ordem de Serviço, quando informados.
- Peças e insumos necessários ficam associados à Ordem de Serviço, quando informados.
- A Ordem de Serviço fica disponível para consulta e acompanhamento.
- O fluxo pode seguir para Registrar Problema Relatado, que inicia o diagnóstico.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico
```

> **Decisão de projeto.** Foi adotada a rota plural com prefixo versionado
> `POST /ordens-servico`, alinhada ao padrão compartilhado do projeto. A alternativa
> `POST /ordens-servico` foi descartada por não usar o prefixo `/`.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `MECANICO`.
- Escopo: `os:escrever`.
- O identificador do usuário é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `clienteId` | string | Identificador obrigatório do cliente. |
| Body | `veiculoId` | string | Identificador obrigatório do veículo. |
| Body | `problemaRelatado` | string | Descrição obrigatória do problema informado pelo cliente. |
| Body | `servicosSolicitados` | array | Lista opcional de serviços solicitados. |
| Body | `servicosSolicitados[].servicoId` | string | Identificador de serviço do catálogo. |
| Body | `pecasInsumos` | array | Lista opcional de peças e insumos necessários. |
| Body | `pecasInsumos[].pecaInsumoId` | string | Identificador de peça ou insumo do catálogo/estoque. |
| Body | `pecasInsumos[].quantidade` | decimal | Quantidade obrigatória quando o item for informado; deve ser maior que zero. |

```json
{
  "clienteId": "uuid-do-cliente",
  "veiculoId": "uuid-do-veiculo",
  "problemaRelatado": "Descrição do problema informado pelo cliente",
  "servicosSolicitados": [
    {
      "servicoId": "uuid-do-servico"
    }
  ],
  "pecasInsumos": [
    {
      "pecaInsumoId": "uuid-da-peca-ou-insumo",
      "quantidade": 1
    }
  ]
}
```

**Validações**

*Técnicas*

- `clienteId` deve ser informado.
- `veiculoId` deve ser informado.
- `problemaRelatado` deve ser informado.
- `servicosSolicitados`, quando informado, deve conter identificadores de serviço válidos.
- `pecasInsumos`, quando informado, deve conter identificadores de peças ou insumos válidos.
- `pecasInsumos[].quantidade` deve ser maior que zero.

*Negócio*

- O cliente deve existir.
- O veículo deve existir.
- O veículo deve estar vinculado ao cliente.
- Serviços solicitados, quando informados, devem existir no catálogo.
- Peças e insumos, quando informados, devem existir no catálogo/estoque.

**Processamento**

1. Receber os dados da Ordem de Serviço.
2. Validar os campos obrigatórios.
3. Consultar e validar cliente, veículo e vínculo entre eles.
4. Validar o problema relatado.
5. Validar serviços solicitados, se houver.
6. Validar peças e insumos, se houver.
7. Criar a Ordem de Serviço.
8. Definir status inicial como `RECEBIDA`.
9. Registrar o problema relatado.
10. Associar cliente e veículo à Ordem de Serviço.
11. Associar serviços solicitados, quando informados.
12. Associar peças e insumos necessários, quando informados.
13. Persistir a Ordem de Serviço.
14. Retornar os dados da Ordem de Serviço criada.

**Persistência**

- Consulta: agregado/dados de `Cliente`.
- Consulta: agregado/dados de `Veículo`.
- Consulta: vínculo entre `Cliente` e `Veículo`.
- Consulta: catálogo de `Serviço`, quando houver serviços solicitados.
- Consulta: catálogo/estoque de Peças e Insumos, quando houver peças ou insumos.
- Altera: `Ordem de Serviço` com novo registro.
- Persiste: identificador da OS, cliente vinculado, veículo vinculado, problema relatado,
  serviços solicitados, peças e insumos necessários e status `RECEBIDA`.

**Saída da API**

```json
{
  "id": "uuid-da-os",
  "clienteId": "uuid-do-cliente",
  "veiculoId": "uuid-do-veiculo",
  "problemaRelatado": "Descrição do problema informado pelo cliente",
  "servicosSolicitados": [
    {
      "servicoId": "uuid-do-servico"
    }
  ],
  "pecasInsumos": [
    {
      "pecaInsumoId": "uuid-da-peca-ou-insumo",
      "quantidade": 1
    }
  ],
  "status": "RECEBIDA"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Ordem de Serviço criada com sucesso. |
| `400` | Dados obrigatórios ausentes ou inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `os:escrever`. |
| `404` | Cliente, veículo, serviço, peça ou insumo não encontrado. |
| `409` | Veículo não vinculado ao cliente. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de clientes.
- Módulo de veículos.
- Módulo de Ordens de Serviço.
- Módulo de serviços.
- Módulo de peças e insumos.
- `ClienteRepository`.
- `VeiculoRepository`.
- `OrdemDeServicoRepository`.
- `ServicoRepository`.
- Repositório de peças e insumos.
- Caso de uso Consultar Cliente.
- Caso de uso Consultar Veículo.
- Caso de uso Vincular Cliente ao Veículo.

**Testes**

*Unitários*

- Cria Ordem de Serviço com status `RECEBIDA` quando os dados são válidos.
- Associa cliente e veículo à Ordem de Serviço.
- Registra o problema relatado.
- Associa serviços solicitados quando informados.
- Associa peças e insumos quando informados.
- Rejeita criação quando `clienteId` não for informado.
- Rejeita criação quando `veiculoId` não for informado.
- Rejeita criação quando `problemaRelatado` não for informado.
- Rejeita criação quando o cliente não existir.
- Rejeita criação quando o veículo não existir.
- Rejeita criação quando o veículo não estiver vinculado ao cliente.
- Rejeita criação quando serviço informado não existir.
- Rejeita criação quando peça ou insumo informado não existir.
- Rejeita criação quando quantidade de peça ou insumo for menor ou igual a zero.

*Integração*

- `POST` válido retorna `201` e persiste a Ordem de Serviço.
- Ordem de Serviço criada pode ser consultada.
- Cliente inexistente retorna `404`.
- Veículo inexistente retorna `404`.
- Veículo não vinculado ao cliente retorna `409`.
- Serviço informado inexistente retorna `404`.
- Peça ou insumo informado inexistente retorna `404`.
- Quantidade de peça ou insumo menor ou igual a zero retorna `400`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `OrdemDeServico`
- [ ] Definir os campos necessários da Ordem de Serviço
- [ ] Definir status inicial da Ordem de Serviço como `RECEBIDA`
- [ ] Definir relacionamento da Ordem de Serviço com `Cliente`
- [ ] Definir relacionamento da Ordem de Serviço com `Veículo`
- [ ] Definir campo para registrar o problema relatado
- [ ] Criar método de domínio para criar Ordem de Serviço
- [ ] Associar serviços solicitados à Ordem de Serviço, quando informados
- [ ] Associar peças e insumos necessários à Ordem de Serviço, quando informados

**Caso de uso**

- [ ] Implementar `CriarOrdemDeServico`
- [ ] Receber `clienteId`, `veiculoId` e problema relatado
- [ ] Receber serviços solicitados, quando informados
- [ ] Receber peças e insumos necessários, quando informados
- [ ] Validar que o cliente existe
- [ ] Validar que o veículo existe
- [ ] Validar que o veículo está vinculado ao cliente
- [ ] Criar a Ordem de Serviço com status `RECEBIDA`
- [ ] Associar cliente e veículo à Ordem de Serviço
- [ ] Registrar o problema relatado na Ordem de Serviço
- [ ] Persistir a Ordem de Serviço no banco

**Repositório**

- [ ] Criar ou ajustar `OrdemDeServicoRepository` para persistência da OS
- [ ] Criar ou ajustar `ClienteRepository` para consulta do cliente
- [ ] Criar ou ajustar `VeiculoRepository` para consulta do veículo
- [ ] Consultar serviços solicitados no `ServicoRepository`
- [ ] Consultar peças e insumos no repositório de peças e insumos

**Integrações**

- [ ] Integrar consulta ao contexto Cliente
- [ ] Integrar consulta ao contexto Veículo
- [ ] Integrar consulta ao contexto Serviços
- [ ] Integrar consulta ao contexto Peças & Insumos

**Handler HTTP**

- [ ] Implementar `POST /ordens-servico`
- [ ] Criar DTO/request de entrada
- [ ] Criar DTO/response de saída
- [ ] Implementar validação do payload
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `os:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar que `clienteId` foi informado
- [ ] Validar que `veiculoId` foi informado
- [ ] Validar que `problemaRelatado` foi informado
- [ ] Validar serviços solicitados, quando informados
- [ ] Validar peças e insumos, quando informados
- [ ] Validar que quantidade de peça ou insumo é maior que zero
- [ ] Retornar `201` quando a Ordem de Serviço for criada com sucesso
- [ ] Retornar `400` para dados obrigatórios ausentes ou inválidos
- [ ] Retornar `404` quando cliente ou veículo não existir
- [ ] Retornar `404` quando serviço, peça ou insumo não existir
- [ ] Retornar `409` quando o veículo não estiver vinculado ao cliente
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Transação e idempotência**

- [ ] Persistir a Ordem de Serviço e suas associações em uma única transação
- [ ] Garantir rollback integral quando alguma associação for inválida ou a persistência falhar

**Testes unitários**

- [ ] Criação válida de Ordem de Serviço
- [ ] `clienteId` ausente
- [ ] `veiculoId` ausente
- [ ] Problema relatado ausente
- [ ] Cliente inexistente
- [ ] Veículo inexistente
- [ ] Veículo não vinculado ao cliente
- [ ] Status inicial `RECEBIDA`
- [ ] Associação com cliente e veículo
- [ ] Associação de serviços solicitados quando informados
- [ ] Associação de peças e insumos quando informados
- [ ] Quantidade de peça ou insumo menor ou igual a zero

**Testes de integração**

- [ ] Endpoint cria Ordem de Serviço válida e retorna `201`
- [ ] Ordem de Serviço é persistida corretamente no banco
- [ ] Ordem de Serviço criada pode ser consultada
- [ ] Endpoint retorna `400` para dados obrigatórios ausentes ou inválidos
- [ ] Endpoint retorna `404` quando cliente, veículo, serviço, peça ou insumo não existe
- [ ] Endpoint retorna `409` quando veículo não está vinculado ao cliente
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
