---
documento: Refinamento de Requisitos — Iniciar Execução
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Iniciar Execução

Este documento detalha a tarefa Iniciar Execução do contexto de Ordem de Serviço.

## 5 · Iniciar Execução

### 5.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Iniciar formalmente a execução dos serviços autorizados de uma Ordem de Serviço, registrando que
o veículo entrou na etapa de reparo.

**Problema**

A oficina precisa controlar quando os serviços de uma OS efetivamente começaram, garantindo que
apenas serviços autorizados sejam executados e permitindo acompanhar corretamente o andamento do
atendimento.

**Pré-condições**

- Deve existir uma Ordem de Serviço.
- O diagnóstico deve ter sido concluído.
- Deve existir um orçamento aprovado pelo cliente.
- Os serviços que serão executados devem estar autorizados.
- As peças e insumos necessários devem estar disponíveis ou reservados, quando aplicável.
- A OS deve estar apta para execução e disponível na fila de atendimento.
- A OS não pode estar finalizada nem entregue.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-28 | Permitir ao mecânico selecionar uma OS disponível na fila de atendimento. |
| RF-OS-29 | Validar se a OS está apta para execução. |
| RF-OS-30 | Validar se o orçamento foi aprovado. |
| RF-OS-31 | Validar se os serviços estão autorizados. |
| RF-OS-32 | Validar a disponibilidade das peças e insumos necessários, quando aplicável. |
| RF-OS-33 | Registrar o início da execução. |
| RF-OS-34 | Registrar a data e hora de início da execução. |
| RF-OS-35 | Alterar o status da OS para `EM_EXECUCAO`. |
| RF-OS-36 | Retirar ou marcar a OS como atendida na fila de atendimento. |
| RF-OS-37 | Disponibilizar ao mecânico os serviços autorizados que deverão ser executados. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-19 | A operação deve ser persistida de forma consistente. |
| RNF-OS-20 | Somente usuários autorizados devem poder iniciar a execução. |
| RNF-OS-21 | A mudança de status e o registro do início da execução devem ocorrer de forma consistente. |
| RNF-OS-22 | Deve ser mantida a rastreabilidade do início da execução da OS. |

**Fluxo Principal**

1. O mecânico consulta a fila de atendimento.
2. O mecânico seleciona uma Ordem de Serviço apta para execução.
3. O sistema consulta os dados da OS.
4. O sistema valida se existe um orçamento aprovado.
5. O sistema valida se os serviços estão autorizados.
6. O sistema verifica se as peças e insumos necessários estão disponíveis ou reservados, quando aplicável.
7. O mecânico solicita o início da execução.
8. O sistema registra a data e hora de início.
9. O sistema altera o status da OS para `EM_EXECUCAO`.
10. O sistema atualiza a situação da OS na fila de atendimento.
11. O sistema confirma o início da execução.
12. O sistema disponibiliza os serviços autorizados para execução pelo mecânico.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | OS não encontrada | Informa que a Ordem de Serviço não existe. |
| A2 | OS não está na fila | Impede o início caso ela ainda não esteja apta para execução. |
| A3 | Orçamento não aprovado | Impede o início dos serviços. |
| A4 | Orçamento recusado | A OS não pode seguir para execução dos serviços recusados. |
| A5 | Peça ou insumo necessário indisponível | Impede o início quando a ausência do item inviabilizar o serviço. |
| A6 | Serviço não autorizado | Impede que o serviço seja executado. |
| A7 | Execução já iniciada | Impede um novo início da mesma execução. |
| A8 | OS finalizada ou entregue | Impede o início da execução. |
| A9 | Usuário sem autorização | Impede a operação. |

**Saída**

- Execução da Ordem de Serviço iniciada, com a OS atualizada para `EM_EXECUCAO`.

**Pós-condições**

- A OS fica com status `EM_EXECUCAO` e a data e hora de início ficam registradas.
- A OS deixa de estar aguardando atendimento na fila.
- Os serviços autorizados ficam disponíveis para execução e o mecânico pode registrar o andamento.
- Caso seja encontrado um novo problema durante o reparo, o fluxo pode seguir para Registrar
  Problema Adicional.

---

### 5.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/execucao/iniciar
```

> **Decisão de projeto.** A regra de transição pertence ao domínio da Ordem de Serviço, e não ao
> handler HTTP: o caso de uso chama `ordemServico.iniciarExecucao()`. Orçamento e Estoque são
> apenas **consultados** para validar a execução; o único agregado alterado aqui é a Ordem de
> Serviço.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`.
- Escopo: `os:escrever`.
- O identificador do mecânico é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |

Não há corpo na requisição.

**Validações**

*Técnicas*

- `osId` em formato UUID válido.

*Negócio*

- A OS deve existir.
- Deve existir orçamento aprovado.
- A OS deve estar na fila de atendimento e apta para execução.
- Devem existir serviços autorizados para execução.
- As peças e insumos necessários devem estar liberados para utilização, quando aplicável.
- A OS ainda não pode estar `EM_EXECUCAO`.
- A OS não pode estar `FINALIZADA` nem `ENTREGUE`.

**Regra de domínio**

```
AGUARDANDO_EXECUCAO → EM_EXECUCAO
```

Fluxo conceitual: orçamento aprovado, recursos necessários disponíveis, OS incluída na fila,
mecânico seleciona a OS, iniciar execução.

**Processamento**

1. Receber o identificador da OS e identificar o usuário autenticado.
2. Buscar a Ordem de Serviço e validar sua existência.
3. Validar o estado atual da OS.
4. Validar se existe orçamento aprovado.
5. Validar se a OS está apta para execução.
6. Validar disponibilidade ou reserva dos itens necessários, quando aplicável.
7. Registrar o início da execução, com data e hora.
8. Associar o mecânico responsável.
9. Alterar o status da OS para `EM_EXECUCAO`.
10. Retirar ou marcar a OS como atendida na fila de atendimento.
11. Persistir as alterações.
12. Retornar a Ordem de Serviço atualizada.

**Persistência**

- Consulta: `orcamento` (autorização dos serviços), estoque (disponibilidade ou reserva dos itens).
- Altera: `ordem_servico` (`status = EM_EXECUCAO`, `data_inicio_execucao`, `mecanico_id`, situação na fila).

**Saída da API**

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "status": "EM_EXECUCAO",
  "mecanicoId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29",
  "dataInicioExecucao": "2026-08-14T20:45:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Execução iniciada com sucesso. |
| `400` | Requisição inválida. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:escrever`. |
| `404` | Ordem de Serviço não encontrada. |
| `409` | Estado atual da OS não permite iniciar a execução; orçamento ainda não aprovado; peças ou insumos indisponíveis; execução já iniciada. |

**Dependências**

- `OrdemDeServicoRepository`.
- `OrcamentoRepository`.
- Repositório ou serviço de Estoque.
- Middleware de autenticação/autorização.

**Fora do escopo desta tarefa**

Executar ou concluir individualmente os serviços; registrar problema adicional; gerar e aprovar
orçamento complementar; registrar consumo e baixa das peças e insumos; finalizar a OS; notificar
o cliente sobre a conclusão. Cada uma dessas operações tem o seu próprio caso de uso.

**Testes**

*Unitários*

- Início válido altera o status para `EM_EXECUCAO` e grava a data de início.
- Rejeita OS sem orçamento aprovado.
- Rejeita execução já iniciada.
- Rejeita OS `FINALIZADA` ou `ENTREGUE`.
- Rejeita quando peças ou insumos necessários estão indisponíveis.
- Aceita OS que não utiliza peças nem insumos.

*Integração*

- `POST` válido retorna `200` e persiste status, data de início e mecânico responsável.
- OS inexistente retorna `404`.
- Orçamento não aprovado retorna `409`.
- Execução já iniciada retorna `409`.
- Peças ou insumos indisponíveis retornam `409`.
- Sem token retorna `401` e perfil sem escopo retorna `403`.
- A fila de atendimento é atualizada após o início.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Criar o método de domínio `iniciarExecucao()` na Ordem de Serviço
- [ ] Validar o status atual da OS e que a execução ainda não foi iniciada
- [ ] Validar existência de orçamento aprovado e de serviços autorizados
- [ ] Validar que a OS está apta para execução e presente na fila de atendimento
- [ ] Validar disponibilidade ou reserva das peças e dos insumos necessários
- [ ] Definir a regra para OS que não utiliza peças nem insumos
- [ ] Alterar o status da OS para `EM_EXECUCAO` e registrar a data e hora de início
- [ ] Associar o mecânico responsável à execução

**Caso de uso**

- [ ] Implementar `IniciarExecucao`
- [ ] Atualizar ou remover a OS da fila de atendimento
- [ ] Persistir as alterações da Ordem de Serviço

**Repositório**

- [ ] Criar ou ajustar `OrdemDeServicoRepository`
- [ ] Criar ou ajustar a consulta ao orçamento aprovado

**Integrações**

- [ ] Criar ou ajustar a consulta ao estoque, quando necessária

**Handler HTTP**

- [ ] Implementar `POST /ordens-servico/{osId}/execucao/iniciar`
- [ ] Implementar a validação do path param `osId`
- [ ] Criar DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Retornar `404` para OS inexistente
- [ ] Retornar `409` para orçamento não aprovado
- [ ] Retornar `409` para OS em estado incompatível
- [ ] Retornar `409` quando recursos obrigatórios não estiverem disponíveis

**Testes unitários**

- [ ] Início de execução válido
- [ ] OS inexistente
- [ ] Orçamento não aprovado
- [ ] Execução já iniciada
- [ ] OS `FINALIZADA` e OS `ENTREGUE`
- [ ] Peças necessárias indisponíveis
- [ ] Insumos necessários indisponíveis
- [ ] OS que não necessita de peças nem insumos
- [ ] Usuário sem autorização

**Testes de integração**

- [ ] Persistência do status `EM_EXECUCAO`
- [ ] Persistência da data de início
- [ ] Atualização da fila de atendimento

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
