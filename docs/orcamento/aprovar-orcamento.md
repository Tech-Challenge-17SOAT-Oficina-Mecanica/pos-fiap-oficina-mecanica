---
documento: Refinamento de Requisitos — Aprovar Orçamento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Aprovar Orçamento

Este documento detalha a tarefa Aprovar Orçamento do contexto de Orçamento.

## 3 · Aprovar Orçamento

### 3.1 Refinamento de Produto

**Persona**

Cliente.

**Objetivo**

Autorizar a execução dos serviços apresentados pela oficina.

**Problema**

A oficina precisa registrar a aprovação do cliente antes de liberar a OS para execução.

**Pré-condições**

- Deve existir uma OS associada ao cliente.
- Deve existir um orçamento associado à OS.
- A OS deve estar com status `AGUARDANDO_APROVACAO`.
- O cliente deve estar autenticado e autorizado.
- O orçamento não deve possuir aprovação ou recusa registrada.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-ORC-18 | Permitir ao cliente aprovar o orçamento. |
| RF-ORC-19 | Registrar o cliente responsável pela aprovação. |
| RF-ORC-20 | Registrar a data e hora da aprovação. |
| RF-ORC-21 | Atualizar a OS para `AGUARDANDO_EXECUCAO`. |
| RF-ORC-22 | Impedir nova aprovação para orçamento que já possua aprovação ou recusa registrada. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-08 | A aprovação deve ser persistida de forma transacional. |
| RNF-ORC-09 | O cliente só pode aprovar orçamento vinculado à sua OS. |
| RNF-ORC-10 | A operação deve ser registrada para rastreabilidade. |

**Fluxo Principal**

1. O cliente consulta o orçamento.
2. O sistema apresenta os serviços, peças, insumos e valores.
3. O cliente aprova o orçamento.
4. O sistema valida o orçamento, a OS e a autorização do cliente.
5. O sistema registra a aprovação.
6. O sistema atualiza a OS para `AGUARDANDO_EXECUCAO`.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente recusa o orçamento | O fluxo segue para Recusar Orçamento. |
| A2 | Orçamento já possui aprovação ou recusa registrada | Impede nova decisão. |
| A3 | Orçamento não encontrado | Informa que o orçamento não existe. |
| A4 | Cliente não autorizado | Impede a aprovação. |

**Saída**

- Aprovação do orçamento registrada e OS liberada para aguardar execução.

**Pós-condições**

- Cliente e data e hora da aprovação registrados.
- A OS fica com status `AGUARDANDO_EXECUCAO`.
- O orçamento não recebe status próprio.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
POST /orcamentos/{orcamentoId}/aprovar
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Permitido apenas para o cliente vinculado à OS do orçamento.
- Escopo: `orcamentos:aprovar`.
- O cliente aprovador é identificado pelo usuário autenticado; não se envia `clienteId` no body.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `orcamentoId` | uuid | Identificador do orçamento. |

Não há corpo na requisição.

**Validações**

*Técnicas*

- `orcamentoId` em formato UUID válido.

*Negócio*

- O orçamento deve existir.
- A OS vinculada deve pertencer ao cliente autenticado.
- A OS deve estar em `AGUARDANDO_APROVACAO`.
- Não pode existir aprovação nem recusa registrada para o orçamento.
- A aprovação não pode ser registrada duas vezes.

**Regra de domínio**

```
OS em AGUARDANDO_APROVACAO → aprovar orçamento → OS em AGUARDANDO_EXECUCAO
```

O orçamento não possui campo de status: a etapa do fluxo é controlada pelo status da OS, e a
decisão do cliente fica registrada nos campos de aprovação do orçamento.

**Processamento**

1. Receber o identificador do orçamento e identificar o cliente autenticado.
2. Consultar o orçamento e a OS vinculada.
3. Validar a autorização do cliente.
4. Validar se o orçamento ainda não possui decisão registrada.
5. Registrar o cliente e a data e hora da aprovação.
6. Atualizar a OS para `AGUARDANDO_EXECUCAO`.
7. Persistir as alterações em uma única transação.
8. Registrar a operação em log.

**Persistência**

- Consulta: `orcamento`, `ordem_servico`, `cliente`.
- Altera: `orcamento` (`cliente_aprovador_id`, `data_aprovacao`), `ordem_servico.status`.

**Saída da API**

```json
{
  "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "statusOrdemServico": "AGUARDANDO_EXECUCAO",
  "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "dataAprovacao": "2026-08-18T10:30:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Aprovação registrada com sucesso. |
| `401` | Token ausente ou expirado. |
| `403` | Cliente sem permissão para aprovar o orçamento. |
| `404` | Orçamento não encontrado. |
| `409` | Orçamento já possui aprovação ou recusa registrada; OS fora de `AGUARDANDO_APROVACAO`. |

**Dependências**

- `OrcamentoRepository`.
- `OrdemDeServicoRepository`.
- Contexto do cliente autenticado.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Registra aprovação para orçamento válido vinculado ao cliente autenticado.
- Registra cliente e data e hora da aprovação.
- Não cria nem atualiza status próprio do orçamento.
- Rejeita orçamento com aprovação ou recusa já registrada.

*Integração*

- Aprovação válida retorna `200` e atualiza a OS para `AGUARDANDO_EXECUCAO`.
- Orçamento inexistente retorna `404`.
- Orçamento de outro cliente retorna `403`.
- Orçamento já decidido retorna `409`.
- Sem autenticação retorna `401`.
- Aprovação e atualização da OS ocorrem na mesma transação.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar os campos `clienteAprovadorId` e `dataAprovacao` no orçamento
- [ ] Não criar campo de status no orçamento
- [ ] Garantir a transição da OS de `AGUARDANDO_APROVACAO` para `AGUARDANDO_EXECUCAO`

**Caso de uso**

- [ ] Implementar `AprovarOrcamento`
- [ ] Validar o vínculo entre cliente, OS e orçamento
- [ ] Validar que a OS está em `AGUARDANDO_APROVACAO`
- [ ] Validar que o orçamento não possui aprovação nem recusa anterior
- [ ] Registrar o cliente e a data e hora da aprovação
- [ ] Atualizar a OS para `AGUARDANDO_EXECUCAO`

**Repositório**

- [ ] Criar ou ajustar `OrcamentoRepository`
- [ ] Criar ou ajustar `OrdemDeServicoRepository`

**Transação**

- [ ] Garantir persistência transacional da aprovação e da atualização da OS

**Handler HTTP**

- [ ] Implementar `POST /orcamentos/{orcamentoId}/aprovar`
- [ ] Obter o cliente pelo JWT
- [ ] Criar DTO/response de saída
- [ ] Aplicar autenticação e autorização na rota
- [ ] Retornar os erros `401`, `403`, `404` e `409`

**Testes unitários**

- [ ] Aprovação válida
- [ ] Orçamento inexistente
- [ ] Orçamento de outro cliente
- [ ] Orçamento já aprovado ou recusado

**Testes de integração**

- [ ] `200` com a OS em `AGUARDANDO_EXECUCAO`
- [ ] Persistência transacional
- [ ] `401`, `403`, `404` e `409`

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
