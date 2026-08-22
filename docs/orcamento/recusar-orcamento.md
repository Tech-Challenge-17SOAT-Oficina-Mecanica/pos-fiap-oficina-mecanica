---
documento: Refinamento de Requisitos — Recusar Orçamento
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Recusar Orçamento

Este documento detalha a tarefa Recusar Orçamento do contexto de Orçamento.

## 4 · Recusar Orçamento

### 4.1 Refinamento de Produto

**Persona**

Cliente.

**Objetivo**

Recusar os serviços apresentados em um orçamento principal ou complementar.

**Problema**

A oficina precisa registrar a decisão do cliente quando ele não autoriza um orçamento. A recusa
do orçamento principal encerra o atendimento; a recusa de um orçamento complementar não deve
cancelar os serviços já aprovados no orçamento principal.

**Pré-condições**

- Deve existir uma OS associada ao cliente.
- Deve existir um orçamento associado à OS.
- O orçamento deve possuir `tipoOrcamento` igual a `PRINCIPAL` ou `COMPLEMENTAR`.
- O orçamento deve estar com status `CRIADO`.
- Quando o orçamento for `COMPLEMENTAR`, ele deve estar vinculado ao orçamento principal da mesma OS.
- A OS deve estar com status `AGUARDANDO_APROVACAO`.
- O cliente deve estar autenticado e autorizado.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-ORC-23 | Permitir ao cliente recusar orçamento principal. |
| RF-ORC-24 | Permitir ao cliente recusar orçamento complementar. |
| RF-ORC-25 | Identificar o tipo do orçamento a ser recusado. |
| RF-ORC-26 | Registrar o cliente responsável pela recusa. |
| RF-ORC-27 | Registrar a data e hora da recusa. |
| RF-ORC-28 | Registrar o motivo da recusa, quando informado. |
| RF-ORC-29 | Atualizar o `statusOrcamento` para `RECUSADO`. |
| RF-ORC-30 | Atualizar a OS para `CANCELADA` quando o orçamento principal for recusado. |
| RF-ORC-31 | Não cancelar a OS quando o orçamento complementar for recusado. |
| RF-ORC-32 | Impedir nova decisão para orçamento já aprovado ou recusado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-11 | A recusa deve exigir autenticação e autorização do cliente. |
| RNF-ORC-12 | A recusa deve ser registrada de forma persistente. |
| RNF-ORC-13 | A operação deve ser rastreável. |
| RNF-ORC-14 | A recusa e as atualizações necessárias devem ocorrer de forma transacional. |

**Fluxo Principal**

1. O cliente consulta o orçamento.
2. O sistema apresenta os serviços, peças, insumos, valores e tipo do orçamento.
3. O cliente seleciona a recusa.
4. O cliente confirma a recusa.
5. O sistema valida o orçamento, a OS e a autorização do cliente.
6. O sistema identifica se o orçamento é `PRINCIPAL` ou `COMPLEMENTAR`.
7. O sistema atualiza o `statusOrcamento` para `RECUSADO`.
8. O sistema registra o cliente, a data/hora e o motivo da recusa.
9. Quando o orçamento for `PRINCIPAL`, o sistema atualiza a OS para `CANCELADA`.
10. Quando o orçamento for `COMPLEMENTAR`, o sistema mantém a OS ativa no fluxo dos serviços do orçamento principal.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente cancela a recusa | Nenhuma alteração é realizada. |
| A2 | Orçamento já aprovado | Impede a recusa. |
| A3 | Orçamento já recusado | Impede nova recusa. |
| A4 | Orçamento complementar sem principal vinculado | Impede a recusa. |
| A5 | Orçamento não encontrado | Informa que o orçamento não existe. |
| A6 | Cliente não possui autorização | Impede a recusa. |

**Saída**

- Orçamento recusado.
- Quando for orçamento principal, OS cancelada.
- Quando for orçamento complementar, OS permanece ativa.

**Pós-condições**

- `statusOrcamento` atualizado para `RECUSADO`.
- Cliente, data/hora e motivo da recusa registrados.
- A recusa do orçamento principal impede a execução da OS.
- A recusa do orçamento complementar impede somente os serviços complementares.
- A OS não é cancelada pela recusa de um orçamento complementar.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
POST /orcamentos/{orcamentoId}/recusar
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Permitido apenas para o cliente vinculado à OS do orçamento.
- Escopo: `orcamentos:recusar`.
- O cliente que recusa é identificado pelo usuário autenticado.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `orcamentoId` | uuid | Identificador do orçamento. |
| Body | `motivo` | string | Motivo da recusa, opcional. |

Exemplo:

```json
{
  "motivo": "Valor acima do esperado."
}
```

**Validações**

*Técnicas*

- `orcamentoId` deve possuir formato UUID válido.
- O motivo da recusa, quando informado, deve obedecer ao tamanho máximo definido pelo sistema.

*Negócio*

- O orçamento deve existir.
- A OS vinculada deve pertencer ao cliente autenticado.
- A OS deve estar em `AGUARDANDO_APROVACAO`.
- `tipoOrcamento` deve ser `PRINCIPAL` ou `COMPLEMENTAR`.
- `statusOrcamento` deve estar em `CRIADO`.
- Quando o tipo for `COMPLEMENTAR`, deve existir orçamento principal vinculado à mesma OS.
- Quando o tipo for `COMPLEMENTAR`, `orcamentoOriginalId` deve referenciar o orçamento principal da mesma OS.

**Processamento**

1. Receber o identificador do orçamento e o motivo da recusa.
2. Identificar o cliente autenticado.
3. Consultar o orçamento e a OS vinculada.
4. Validar a autorização do cliente.
5. Validar se a OS está em `AGUARDANDO_APROVACAO`.
6. Validar se o orçamento está em `CRIADO`.
7. Identificar o tipo do orçamento.
8. Caso seja complementar, validar o vínculo com o orçamento principal.
9. Atualizar `statusOrcamento` para `RECUSADO`.
10. Registrar o cliente, a data/hora e o motivo da recusa.
11. Quando o orçamento for `PRINCIPAL`, atualizar a OS para `CANCELADA`.
12. Quando o orçamento for `COMPLEMENTAR`, consultar o orçamento principal:
    - se o principal estiver `APROVADO`, atualizar a OS para `AGUARDANDO_EXECUCAO`;
    - caso contrário, manter a OS em `AGUARDANDO_APROVACAO`.
13. Persistir as alterações em uma única transação.
14. Registrar a operação em log, sem expor dados sensíveis.

**Persistência**

*Orçamento*

- Atualizar `statusOrcamento` para `RECUSADO`.
- Registrar `clienteRecusadorId`.
- Registrar `dataRecusa`.
- Registrar `motivoRecusa`, quando informado.

*Ordem de Serviço*

- Quando o orçamento for principal: atualizar `status` para `CANCELADA`.
- Quando o orçamento for complementar e o principal estiver aprovado: atualizar `status` para `AGUARDANDO_EXECUCAO`.
- Quando o orçamento for complementar e o principal não estiver aprovado: manter `status` em `AGUARDANDO_APROVACAO`.

As alterações no Orçamento e na Ordem de Serviço devem ocorrer na mesma transação.

**Saída da API**

Exemplo para orçamento principal:

```json
{
  "orcamentoId": "uuid-do-orcamento-principal",
  "ordemServicoId": "uuid-da-os",
  "tipoOrcamento": "PRINCIPAL",
  "statusOrcamento": "RECUSADO",
  "statusOrdemServico": "CANCELADA",
  "clienteId": "uuid-do-cliente",
  "dataRecusa": "2026-08-22T10:30:00-03:00",
  "motivo": "Valor acima do esperado."
}
```

Exemplo para orçamento complementar:

```json
{
  "orcamentoId": "uuid-do-orcamento-complementar",
  "ordemServicoId": "uuid-da-os",
  "tipoOrcamento": "COMPLEMENTAR",
  "orcamentoOriginalId": "uuid-do-orcamento-principal",
  "statusOrcamento": "RECUSADO",
  "statusOrdemServico": "AGUARDANDO_EXECUCAO",
  "clienteId": "uuid-do-cliente",
  "dataRecusa": "2026-08-22T10:30:00-03:00",
  "motivo": "Não autorizo a substituição da correia."
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| 200 OK | Recusa registrada com sucesso. |
| 400 Bad Request | Motivo da recusa inválido ou identificador inválido. |
| 401 Unauthorized | Cliente não autenticado. |
| 403 Forbidden | Cliente sem permissão para recusar o orçamento. |
| 404 Not Found | Orçamento não encontrado. |
| 409 Conflict | Orçamento já aprovado, recusado ou OS fora de `AGUARDANDO_APROVACAO`. |
| 422 Unprocessable Entity | Orçamento complementar sem vínculo válido com orçamento principal. |
| 500 Internal Server Error | Erro inesperado. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de autorização.
- `OrcamentoRepository`.
- `OrdemDeServicoRepository`.
- Contexto do cliente autenticado.
- Banco de dados.

**Testes**

*Unitários*

- Deve recusar orçamento principal válido vinculado ao cliente autenticado.
- Deve recusar orçamento complementar válido vinculado ao principal da mesma OS.
- Deve registrar cliente, data/hora e motivo da recusa.
- Deve atualizar `statusOrcamento` para `RECUSADO`.
- Deve atualizar a OS para `CANCELADA` quando o orçamento principal for recusado.
- Não deve cancelar a OS quando o orçamento complementar for recusado.
- Deve atualizar a OS para `AGUARDANDO_EXECUCAO` quando o complementar for recusado e o principal estiver aprovado.
- Deve impedir recusa de orçamento já aprovado.
- Deve impedir recusa de orçamento já recusado.
- Deve impedir recusa de orçamento complementar sem principal vinculado.

*Integração*

- Deve retornar `400` para motivo ou identificador inválido.
- Deve retornar `401` sem autenticação.
- Deve retornar `403` para orçamento de outro cliente.
- Deve retornar `404` para orçamento inexistente.
- Deve retornar `409` quando a OS não estiver em estado compatível.
- Deve retornar `422` para orçamento complementar sem vínculo válido com principal.
- Deve garantir que recusa e atualização da OS ocorram na mesma transação.

---

### 4.3 Check-list de Implementação

**Domínio**

- [ ] Criar/ajustar `TipoOrcamento` com `PRINCIPAL` e `COMPLEMENTAR`.
- [ ] Criar/ajustar `StatusOrcamento` com `CRIADO`, `APROVADO` e `RECUSADO`.
- [ ] Criar/ajustar os campos `clienteRecusadorId`, `dataRecusa` e `motivoRecusa` no Orçamento.
- [ ] Garantir a transição `CRIADO` → `RECUSADO` no orçamento.
- [ ] Garantir que orçamento complementar tenha orçamento principal vinculado.
- [ ] Garantir que a recusa do complementar não cancele a OS.
- [ ] Garantir a transição da OS `AGUARDANDO_APROVACAO` → `CANCELADA` para recusa do orçamento principal.

**Caso de uso**

- [ ] Implementar o caso de uso `RecusarOrcamento`.
- [ ] Validar o vínculo entre cliente, OS e orçamento.
- [ ] Validar o tipo e o status do orçamento.
- [ ] Validar orçamento complementar e seu vínculo com o principal.
- [ ] Atualizar `statusOrcamento` para `RECUSADO`.
- [ ] Registrar o cliente, a data/hora e o motivo da recusa.
- [ ] Atualizar a OS conforme o tipo do orçamento recusado.

**Repositório**

- [ ] Criar/ajustar `OrcamentoRepository`.
- [ ] Criar/ajustar `OrdemDeServicoRepository`.

**Handler HTTP**

- [ ] Criar handler para `POST /orcamentos/{orcamentoId}/recusar`.
- [ ] Obter o cliente pelo JWT.
- [ ] Aplicar autenticação e autorização na rota.
- [ ] Retornar erros `400`, `401`, `403`, `404`, `409` e `422`.

**Transação**

- [ ] Garantir persistência transacional entre orçamento e OS.

**Testes unitários**

- [ ] Criar testes para recusa principal, complementar, status inválido e vínculo inválido.

**Testes de integração**

- [ ] Criar testes de autenticação, autorização e contrato do endpoint.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Executar testes automatizados.
- [ ] Realizar code review.
- [ ] Validar critérios de aceite da task.
