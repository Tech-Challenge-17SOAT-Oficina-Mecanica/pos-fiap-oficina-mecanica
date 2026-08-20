---
documento: Refinamento de Requisitos — Recusar Orçamento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Recusar Orçamento

Este documento detalha a tarefa Recusar Orçamento do contexto de Orçamento.

## 4 · Recusar Orçamento

### 4.1 Refinamento de Produto

**Persona**

Cliente.

**Objetivo**

Recusar o orçamento e não autorizar a execução dos serviços propostos.

**Problema**

O cliente pode optar por não realizar os serviços apresentados, sendo necessário registrar
formalmente sua decisão e encerrar o atendimento.

**Pré-condições**

- Deve existir uma OS associada ao cliente.
- Deve existir um orçamento associado à OS.
- A OS deve estar com status `AGUARDANDO_APROVACAO`.
- O cliente deve estar autenticado e autorizado.
- O orçamento não deve possuir aprovação ou recusa registrada.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-ORC-23 | Permitir ao cliente recusar o orçamento. |
| RF-ORC-24 | Registrar o cliente responsável pela recusa. |
| RF-ORC-25 | Registrar a data e hora da recusa. |
| RF-ORC-26 | Registrar o motivo da recusa, quando informado. |
| RF-ORC-27 | Atualizar a OS para `CANCELADA`. |
| RF-ORC-28 | Impedir nova decisão para orçamento que já possua aprovação ou recusa registrada. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-11 | A recusa deve exigir autenticação e autorização do cliente. |
| RNF-ORC-12 | A recusa deve ser registrada de forma persistente. |
| RNF-ORC-13 | A operação deve ser rastreável. |
| RNF-ORC-14 | A recusa e o cancelamento da OS devem ocorrer de forma transacional. |

**Fluxo Principal**

1. O cliente consulta o orçamento.
2. O sistema apresenta os serviços, peças, insumos e valores.
3. O cliente seleciona a recusa.
4. O cliente confirma a recusa.
5. O sistema valida o orçamento, a OS e a autorização do cliente.
6. O sistema registra a recusa.
7. O sistema atualiza a OS para `CANCELADA`.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente cancela a recusa | Nenhuma alteração é realizada. |
| A2 | Orçamento já possui aprovação ou recusa registrada | Impede nova decisão. |
| A3 | Orçamento não encontrado | Informa que o orçamento não existe. |
| A4 | Cliente não possui autorização | Impede a recusa. |

**Saída**

- Recusa do orçamento registrada e OS cancelada.

**Pós-condições**

- Cliente, data e hora e motivo da recusa registrados.
- A OS fica com status `CANCELADA` e não pode seguir para execução.
- O orçamento não recebe status próprio.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
POST /orcamentos/{orcamentoId}/recusar
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Permitido apenas para o cliente vinculado à OS do orçamento.
- Escopo: `orcamentos:aprovar`.
- O cliente que recusa é identificado pelo usuário autenticado.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `orcamentoId` | uuid | Identificador do orçamento. |
| Body | `motivo` | string | Opcional; motivo da recusa. |

```json
{
  "motivo": "Valor acima do esperado."
}
```

**Validações**

*Técnicas*

- `orcamentoId` em formato UUID válido.
- `motivo` válido, quando informado.

*Negócio*

- O orçamento deve existir.
- A OS vinculada deve pertencer ao cliente autenticado.
- A OS deve estar em `AGUARDANDO_APROVACAO`.
- Não pode existir aprovação nem recusa registrada para o orçamento.

**Regra de domínio**

```
OS em AGUARDANDO_APROVACAO → recusar orçamento → OS em CANCELADA
```

O orçamento não possui campo de status: a etapa do fluxo é controlada pelo status da OS.

**Processamento**

1. Receber o identificador do orçamento e o motivo da recusa.
2. Identificar o cliente autenticado.
3. Consultar o orçamento e a OS vinculada.
4. Validar a autorização do cliente.
5. Validar se o orçamento ainda não possui decisão registrada.
6. Registrar o cliente, a data e hora e o motivo da recusa.
7. Atualizar a OS para `CANCELADA`.
8. Persistir as alterações em uma única transação.
9. Registrar a operação em log.

**Persistência**

- Consulta: `orcamento`, `ordem_servico`, `cliente`.
- Altera: `orcamento` (`cliente_recusador_id`, `data_recusa`, `motivo_recusa`),
  `ordem_servico.status`.

**Saída da API**

```json
{
  "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "statusOrdemServico": "CANCELADA",
  "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "dataRecusa": "2026-08-18T10:30:00-03:00",
  "motivo": "Valor acima do esperado."
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Recusa registrada com sucesso. |
| `400` | Motivo da recusa inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Cliente sem permissão para recusar o orçamento. |
| `404` | Orçamento não encontrado. |
| `409` | Orçamento já possui aprovação ou recusa registrada; OS fora de `AGUARDANDO_APROVACAO`. |

**Dependências**

- `OrcamentoRepository`.
- `OrdemDeServicoRepository`.
- Contexto do cliente autenticado.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Registra recusa para orçamento válido vinculado ao cliente autenticado.
- Registra cliente, data e hora e motivo da recusa.
- Não cria nem atualiza status próprio do orçamento.
- Rejeita orçamento com aprovação ou recusa já registrada.

*Integração*

- Recusa válida retorna `200` e atualiza a OS para `CANCELADA`.
- Orçamento inexistente retorna `404`.
- Orçamento de outro cliente retorna `403`.
- Orçamento já decidido retorna `409`.
- Sem autenticação retorna `401`.
- Recusa e cancelamento da OS ocorrem na mesma transação.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar os campos `clienteRecusadorId`, `dataRecusa` e `motivoRecusa` no orçamento
- [ ] Não criar campo de status no orçamento
- [ ] Garantir a transição da OS de `AGUARDANDO_APROVACAO` para `CANCELADA`

**Caso de uso**

- [ ] Implementar `RecusarOrcamento`
- [ ] Validar o vínculo entre cliente, OS e orçamento
- [ ] Validar que a OS está em `AGUARDANDO_APROVACAO`
- [ ] Validar que o orçamento não possui aprovação nem recusa anterior
- [ ] Registrar o cliente, a data e hora e o motivo da recusa
- [ ] Atualizar a OS para `CANCELADA`

**Repositório**

- [ ] Criar ou ajustar `OrcamentoRepository`
- [ ] Criar ou ajustar `OrdemDeServicoRepository`

**Transação**

- [ ] Garantir persistência transacional da recusa e do cancelamento da OS

**Handler HTTP**

- [ ] Implementar `POST /orcamentos/{orcamentoId}/recusar`
- [ ] Obter o cliente pelo JWT
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação e autorização na rota
- [ ] Retornar os erros `400`, `401`, `403`, `404` e `409`

**Testes unitários**

- [ ] Recusa válida
- [ ] Orçamento inexistente
- [ ] Orçamento de outro cliente
- [ ] Orçamento já aprovado ou recusado

**Testes de integração**

- [ ] `200` com a OS em `CANCELADA`
- [ ] Persistência transacional
- [ ] `400`, `401`, `403`, `404` e `409`

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
