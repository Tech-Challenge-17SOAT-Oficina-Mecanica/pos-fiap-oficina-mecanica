---
documento: Refinamento de Requisitos — Criar Ordem de Serviço
dono: Helena Miranda
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Criar Ordem de Serviço

Este documento detalha a tarefa Criar Ordem de Serviço do contexto de Ordem de Serviço.

## 3 · Criar Ordem de Serviço

### 3.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Iniciar uma Ordem de Serviço para registrar a entrada de um veículo vinculado a um cliente.

**Problema**

A oficina precisa iniciar formalmente o atendimento do veículo para permitir o acompanhamento da OS desde sua abertura.

**Pré-condições**

- O cliente deve estar cadastrado e identificado.
- O veículo deve estar cadastrado e identificado.
- O veículo deve estar vinculado ao cliente.
- O usuário deve estar autorizado a criar Ordem de Serviço.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-12 | Permitir ao mecânico criar uma Ordem de Serviço. |
| RF-OS-13 | Associar a Ordem de Serviço ao cliente. |
| RF-OS-14 | Associar a Ordem de Serviço ao veículo. |
| RF-OS-15 | Validar se o veículo pertence ao cliente informado. |
| RF-OS-16 | Criar a Ordem de Serviço com status inicial `RECEBIDA`. |
| RF-OS-17 | Registrar a data e hora de criação da Ordem de Serviço. |
| RF-OS-18 | Permitir que a Ordem de Serviço seja acompanhada posteriormente. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-10 | A operação deve ser feita por API RESTful. |
| RNF-OS-11 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-OS-12 | A criação da Ordem de Serviço deve ser persistida de forma consistente. |
| RNF-OS-13 | O vínculo entre cliente, veículo e Ordem de Serviço deve preservar o histórico de atendimento. |
| RNF-OS-14 | A criação da Ordem de Serviço deve ocorrer de forma atômica. |

**Fluxo Principal**

1. O mecânico seleciona o cliente.
2. O mecânico seleciona o veículo.
3. O mecânico solicita a criação da Ordem de Serviço.
4. O sistema verifica se o cliente existe.
5. O sistema verifica se o veículo existe.
6. O sistema verifica se o veículo está vinculado ao cliente.
7. O sistema cria a Ordem de Serviço.
8. O sistema define o status inicial como `RECEBIDA`.
9. O sistema registra a data e hora de criação.
10. O sistema confirma a criação da Ordem de Serviço.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente não encontrado | O sistema informa que o cliente não existe. |
| A2 | Veículo não encontrado | O sistema informa que o veículo não existe. |
| A3 | Veículo não vinculado ao cliente | O sistema impede a criação da Ordem de Serviço. |
| A4 | Usuário sem autorização | O sistema impede a operação. |
| A5 | Erro ao criar Ordem de Serviço | O sistema informa que não foi possível concluir a criação. |

**Saída**

- Ordem de Serviço criada.
- Ordem de Serviço vinculada ao cliente e ao veículo.
- Status inicial `RECEBIDA` registrado.

**Pós-condições**

- A Ordem de Serviço passa a existir no sistema.
- A Ordem de Serviço fica vinculada ao cliente e ao veículo.
- A Ordem de Serviço fica disponível para consulta e acompanhamento.
- O fluxo pode seguir para Iniciar Diagnóstico ou Registrar problema identificado, conforme o processo definido.

### 3.2 Refinamento Técnico

**Endpoint**

- `POST /ordens-servico`

**Autenticação / Autorização**

- Requer autenticação JWT.
- Permitido para usuário com perfil/permissão de Mecânico.
- Requer escopo `os:escrever`.

**Entrada**

Body:

```json
{
  "clienteId": "uuid-do-cliente",
  "veiculoId": "uuid-do-veiculo"
}
```

**Validações**

- Validar autenticação e permissão do mecânico.
- Validar se `clienteId` e `veiculoId` foram informados.
- Validar o formato dos identificadores.
- Validar se o cliente existe.
- Validar se o veículo existe.
- Validar se o veículo está vinculado ao cliente informado.
- Não exigir problema relatado, serviços, peças ou insumos na criação da OS.
- Não gerar orçamento durante a criação da OS.

**Processamento**

- Receber `clienteId` e `veiculoId`.
- Validar o usuário autenticado.
- Consultar o cliente e o veículo.
- Validar o vínculo entre cliente e veículo.
- Criar a Ordem de Serviço com status `RECEBIDA`.
- Associar o cliente e o veículo à Ordem de Serviço.
- Registrar a data e hora de criação.
- Persistir a Ordem de Serviço em uma transação.
- Retornar os dados da OS criada.

**Persistência**

- Consultar Cliente.
- Consultar Veículo e seu vínculo com o Cliente.
- Criar Ordem de Serviço com:
  - `id`;
  - `clienteId`;
  - `veiculoId`;
  - `status = RECEBIDA`;
  - `criadaEm`.
- Nenhum orçamento deve ser criado nesta operação.

**Saída da API**

```json
{
  "ordemServicoId": "uuid-da-os",
  "clienteId": "uuid-do-cliente",
  "veiculoId": "uuid-do-veiculo",
  "status": "RECEBIDA",
  "criadaEm": "2026-08-22T10:30:00-03:00"
}
```

**Códigos HTTP / Erros**

- `201 Created`: Ordem de Serviço criada com sucesso.
- `400 Bad Request`: identificadores ausentes ou inválidos.
- `401 Unauthorized`: usuário não autenticado.
- `403 Forbidden`: usuário sem permissão.
- `404 Not Found`: cliente ou veículo não encontrado.
- `409 Conflict`: veículo não vinculado ao cliente.
- `500 Internal Server Error`: erro inesperado.

**Dependências**

- Módulo de autenticação JWT.
- Módulo de autorização.
- ClienteRepository.
- VeiculoRepository.
- OrdemDeServicoRepository.
- Contexto do usuário autenticado.
- Banco de dados.

**Testes**

- Deve criar Ordem de Serviço com cliente e veículo válidos e vinculados.
- Deve criar a OS com status inicial `RECEBIDA`.
- Deve registrar a data e hora de criação.
- Deve associar a OS ao cliente e ao veículo.
- Deve permitir criar a OS sem problema relatado, serviços, peças ou insumos.
- Deve garantir que nenhum orçamento seja gerado na criação da OS.
- Deve retornar `400 Bad Request` para `clienteId` ou `veiculoId` ausente ou inválido.
- Deve retornar `404 Not Found` para cliente ou veículo inexistente.
- Deve retornar `409 Conflict` quando o veículo não estiver vinculado ao cliente.
- Deve retornar `401 Unauthorized` sem autenticação.
- Deve retornar `403 Forbidden` sem permissão.
- Deve garantir a persistência atômica da Ordem de Serviço.
- Deve possuir teste de integração do endpoint.

### 3.3 Check-list de Implementação

**Domínio e Persistência**

- [ ] Criar/ajustar o modelo `OrdemDeServico`.
- [ ] Definir os campos necessários para abertura da Ordem de Serviço.
- [ ] Definir o status inicial como `RECEBIDA`.
- [ ] Definir os relacionamentos da Ordem de Serviço com Cliente e Veículo.
- [ ] Criar/ajustar `OrdemDeServicoRepository` para persistência da OS.
- [ ] Criar/ajustar `ClienteRepository` para consulta do cliente.
- [ ] Criar/ajustar `VeiculoRepository` para consulta do veículo e vínculo com o cliente.

**Caso de Uso**

- [ ] Implementar o caso de uso `CriarOrdemDeServico`.
- [ ] Garantir que o caso de uso receba somente `clienteId` e `veiculoId`.
- [ ] Validar os identificadores informados.
- [ ] Validar que cliente e veículo existem.
- [ ] Validar que o veículo está vinculado ao cliente.
- [ ] Criar a OS com status `RECEBIDA`.
- [ ] Associar cliente e veículo à Ordem de Serviço.
- [ ] Registrar a data e hora de criação.
- [ ] Persistir a criação de forma atômica.
- [ ] Garantir que problema relatado, serviços, peças e insumos não sejam obrigatórios.
- [ ] Garantir que orçamento não seja gerado nesta operação.

**API e Segurança**

- [ ] Criar DTO de entrada somente com `clienteId` e `veiculoId`.
- [ ] Criar DTO de resposta da Ordem de Serviço criada.
- [ ] Criar handler para `POST /ordens-servico`.
- [ ] Implementar validação do payload.
- [ ] Aplicar autenticação JWT na rota.
- [ ] Aplicar autorização para o perfil/permissão de Mecânico.
- [ ] Retornar `201` para criação bem-sucedida.
- [ ] Mapear erros para `400`, `401`, `403`, `404` e `409`.

**Testes e Documentação**

- [ ] Criar testes unitários para criação válida, identificadores ausentes ou inválidos, cliente inexistente, veículo inexistente e vínculo inválido.
- [ ] Criar testes para status inicial, data de criação e associação com cliente e veículo.
- [ ] Criar testes para garantir que problema, serviços, peças e insumos não são obrigatórios.
- [ ] Criar teste para garantir que orçamento não é gerado na criação.
- [ ] Criar testes de autenticação e autorização.
- [ ] Criar teste de integração do endpoint.
- [ ] Documentar o endpoint no Swagger/OpenAPI.
- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto.
- [ ] Executar testes automatizados, realizar code review e validar critérios de aceite da task.
