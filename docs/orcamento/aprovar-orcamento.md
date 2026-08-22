---
documento: Refinamento de Requisitos — Aprovar Orçamento
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Aprovar Orçamento

Este documento detalha a tarefa Aprovar Orçamento do contexto de Orçamento.

## 3 · Aprovar Orçamento

### 3.1 Refinamento de Produto

**Persona**

Cliente.

**Objetivo**

Autorizar a execução dos serviços apresentados em um orçamento principal ou complementar.

**Problema**

A oficina precisa registrar a aprovação do cliente antes de liberar a OS para execução.

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
| RF-ORC-18 | Permitir ao cliente aprovar orçamento principal. |
| RF-ORC-19 | Permitir ao cliente aprovar orçamento complementar. |
| RF-ORC-20 | Identificar o tipo do orçamento a ser aprovado. |
| RF-ORC-21 | Registrar o cliente responsável pela aprovação. |
| RF-ORC-22 | Registrar a data e hora da aprovação. |
| RF-ORC-44 | Atualizar o `statusOrcamento` para `APROVADO`. |
| RF-ORC-45 | Atualizar a OS para `AGUARDANDO_EXECUCAO`. |
| RF-ORC-46 | Impedir nova aprovação para orçamento já aprovado ou recusado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-08 | A aprovação deve ser persistida de forma transacional. |
| RNF-ORC-09 | O cliente só pode aprovar orçamento vinculado à sua OS. |
| RNF-ORC-10 | A operação deve ser registrada para rastreabilidade. |

**Fluxo Principal**

1. O cliente consulta o orçamento.
2. O sistema apresenta os serviços, peças, insumos, valores e tipo do orçamento.
3. O cliente aprova o orçamento.
4. O sistema valida o orçamento, a OS e a autorização do cliente.
5. O sistema identifica se o orçamento é `PRINCIPAL` ou `COMPLEMENTAR`.
6. O sistema atualiza o `statusOrcamento` para `APROVADO`.
7. O sistema registra o cliente e a data/hora da aprovação.
8. O sistema atualiza a OS para `AGUARDANDO_EXECUCAO`.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente recusa o orçamento | O fluxo segue para Recusar Orçamento. |
| A2 | Orçamento já aprovado | Impede nova aprovação. |
| A3 | Orçamento já recusado | Impede a aprovação. |
| A4 | Orçamento complementar sem principal vinculado | Impede a aprovação. |
| A5 | Orçamento não encontrado | Informa que o orçamento não existe. |
| A6 | Cliente não autorizado | Impede a aprovação. |

**Saída**

- Aprovação do orçamento registrada e OS liberada para aguardar execução.

**Pós-condições**

- `statusOrcamento` atualizado para `APROVADO`.
- Cliente e data e hora da aprovação registrados.
- A OS fica com status `AGUARDANDO_EXECUCAO`.
- O orçamento principal permanece como referência para eventuais orçamentos complementares.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
POST /orcamentos/{orcamentoId}/aprovar
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `CLIENTE`, apenas o cliente vinculado à OS do orçamento.
- Escopo: `orcamentos:decidir`.
- O cliente aprovador é identificado pelo usuário autenticado; não se envia `clienteId` no body.

> **Decisão de projeto.** Aprovar e recusar usam **um escopo só**, `orcamentos:decidir`. Os dois
> escopos anteriores — `orcamentos:aprovar` e `orcamentos:recusar` — separavam duas metades da
> mesma decisão, e ninguém no projeto tem uma sem a outra (D-23).

> **Decisão de projeto.** O cliente se autentica por **token de escopo reduzido**, emitido no envio
> do orçamento e válido apenas para aquela OS. Evita criar cadastro com senha para cliente no MVP,
> e o mesmo token serve para consultar, aprovar e recusar.

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
- `tipoOrcamento` deve ser `PRINCIPAL` ou `COMPLEMENTAR`.
- `statusOrcamento` deve estar em `CRIADO`.
- Quando o tipo for `COMPLEMENTAR`, deve existir orçamento principal vinculado à mesma OS.
- Quando o tipo for `COMPLEMENTAR`, `orcamentoOriginalId` deve referenciar o orçamento principal da mesma OS.

**Regra de domínio**

```
OS em AGUARDANDO_APROVACAO → aprovar orçamento → OS em AGUARDANDO_RECURSOS ou AGUARDANDO_EXECUCAO
```

Orçamento em CRIADO → aprovar orçamento → orçamento em APROVADO.

> **Decisão de projeto.** O `status` do **orçamento** é a fonte da verdade da decisão do cliente;
> o `status` da **OS** é a fonte da verdade da etapa do atendimento. A transição da OS é
> consequência da decisão, nunca o contrário — nenhuma regra deve ler o status da OS para saber se
> o cliente aprovou.

> **Decisão de projeto.** Aprovar orçamento complementar **devolve a OS para a fila**, igual à
> aprovação do principal. A diferença está na recusa: recusar o principal cancela a OS; recusar um
> complementar apenas marca aquele orçamento como `RECUSADO` e devolve a OS para a fila, com o
> serviço original mantido.

**Processamento**

1. Receber o identificador do orçamento e identificar o cliente autenticado.
2. Consultar o orçamento e a OS vinculada.
3. Validar a autorização do cliente.
4. Validar se a OS está em `AGUARDANDO_APROVACAO` e o orçamento está em `CRIADO`.
5. Identificar o tipo do orçamento.
6. Caso seja complementar, validar o vínculo com o orçamento principal.
7. Atualizar `statusOrcamento` para `APROVADO`.
8. Registrar o cliente e a data e hora da aprovação.
9. Chamar o processamento de itens do orçamento aprovado — reserva o disponível e abre pedido de
   compra do faltante, por tipo de item.
10. Definir o status da OS a partir do resultado: `AGUARDANDO_EXECUCAO` quando tudo foi reservado,
    `AGUARDANDO_RECURSOS` quando algum item ficou pendente de compra.
11. Persistir as alterações em uma única transação.
12. Registrar a operação em log, sem expor dados sensíveis.

**Fluxo de liberação de itens — proposta**

A aprovação é o único gatilho que compromete estoque. O fluxo proposto, dentro da mesma transação:

```
AprovarOrcamento
├── valida cliente, OS e orçamento
├── marca o orçamento como APROVADO
├── ProcessarPecas(os, itens do tipo PECA)      → reserva o disponível, abre pedido do faltante
├── ProcessarInsumos(os, itens do tipo INSUMO)  → reserva o disponível, abre pedido do faltante
├── define o status da OS pelo resultado:
│     nada pendente        → AGUARDANDO_EXECUCAO
│     algum item comprado  → AGUARDANDO_RECURSOS
└── confirma a transação
```

O que a aprovação **não** faz: não chama a reserva direta nem o pedido de compra por conta
própria. Ela chama apenas os dois processamentos, que já resolvem reserva e compra juntos. É essa
regra que elimina os caminhos concorrentes de reserva apontados na D-16 — a reserva direta deixa de
ter chamador público.

Quando os itens comprados chegam, a entrada de estoque devolve a OS para `AGUARDANDO_EXECUCAO`.
Aprovar um complementar repete o mesmo fluxo, apenas sobre os itens daquele orçamento.

**Persistência**

- Consulta: `orcamento`, `ordem_servico`, `cliente`.
- Altera: `orcamento` (`status_orcamento = APROVADO`, `cliente_aprovador_id`, `data_aprovacao`), `ordem_servico.status`.
- Altera, por meio dos processamentos chamados: `reserva_estoque`, `item_estoque.saldo_reservado`, `movimentacao_estoque` e `pedido_compra`.

**Saída da API**

```json
{
  "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "tipoOrcamento": "PRINCIPAL",
  "statusOrcamento": "APROVADO",
  "statusOrdemServico": "AGUARDANDO_EXECUCAO",
  "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "dataAprovacao": "2026-08-18T10:30:00-03:00"
}
```

Exemplo para orçamento complementar:

```json
{
  "orcamentoId": "uuid-do-orcamento-complementar",
  "ordemServicoId": "uuid-da-os",
  "tipoOrcamento": "COMPLEMENTAR",
  "orcamentoOriginalId": "uuid-do-orcamento-principal",
  "statusOrcamento": "APROVADO",
  "statusOrdemServico": "AGUARDANDO_EXECUCAO",
  "clienteId": "uuid-do-cliente",
  "dataAprovacao": "2026-08-22T10:30:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Aprovação registrada com sucesso. |
| `401` | Token ausente ou expirado. |
| `403` | Cliente sem o escopo `orcamentos:decidir`, ou orçamento de outra OS. |
| `404` | Orçamento não encontrado. |
| `409` | Orçamento já aprovado, recusado ou OS fora de `AGUARDANDO_APROVACAO`. |
| `409` | Orçamento complementar sem vínculo válido com orçamento principal. |
| `500` | Erro inesperado. |

**Dependências**

- `OrcamentoRepository`.
- `OrdemDeServicoRepository`.
- Caso de uso Processar Peças para Reserva e Compra.
- Caso de uso Processar Insumos para Reserva e Compra.
- Contexto do cliente autenticado.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Registra aprovação para orçamento válido vinculado ao cliente autenticado.
- Aprova orçamento complementar válido vinculado ao orçamento principal da mesma OS.
- Registra cliente e data e hora da aprovação.
- Atualiza `statusOrcamento` para `APROVADO`.
- Rejeita orçamento já aprovado ou recusado.
- Rejeita orçamento complementar sem principal vinculado.

*Integração*

- Aprovação válida com todos os itens em estoque retorna `200` e atualiza a OS para `AGUARDANDO_EXECUCAO`.
- Aprovação válida com item faltante retorna `200`, abre pedido de compra e atualiza a OS para `AGUARDANDO_RECURSOS`.
- Aprovação de complementar devolve a OS para a fila, sem alterar o orçamento principal.
- Orçamento inexistente retorna `404`.
- Orçamento de outro cliente retorna `403`.
- Orçamento já decidido retorna `409`.
- Orçamento complementar sem vínculo válido retorna `409`.
- Sem autenticação retorna `401`.
- Aprovação, reserva, pedido de compra e atualização da OS ocorrem na mesma transação.
- Falha no processamento de itens desfaz a aprovação por inteiro.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar `TipoOrcamento` com `PRINCIPAL` e `COMPLEMENTAR`
- [ ] Criar ou ajustar `StatusOrcamento` com `CRIADO`, `APROVADO` e `RECUSADO`
- [ ] Criar ou ajustar os campos `clienteAprovadorId` e `dataAprovacao` no orçamento
- [ ] Garantir que orçamento complementar tenha orçamento principal vinculado
- [ ] Garantir a transição do orçamento de `CRIADO` para `APROVADO`
- [ ] Garantir a transição da OS de `AGUARDANDO_APROVACAO` para `AGUARDANDO_EXECUCAO` ou `AGUARDANDO_RECURSOS`, conforme o resultado do processamento dos itens

**Caso de uso**

- [ ] Implementar `AprovarOrcamento`
- [ ] Validar o vínculo entre cliente, OS e orçamento
- [ ] Validar que a OS está em `AGUARDANDO_APROVACAO`
- [ ] Validar o tipo e o status do orçamento
- [ ] Validar orçamento complementar e seu vínculo com o principal
- [ ] Atualizar `statusOrcamento` para `APROVADO`
- [ ] Registrar o cliente e a data e hora da aprovação
- [ ] Chamar `ProcessarPecas` para os itens do tipo `PECA`
- [ ] Chamar `ProcessarInsumos` para os itens do tipo `INSUMO`
- [ ] Definir o status da OS pelo resultado do processamento

**Repositório**

- [ ] Criar ou ajustar `OrcamentoRepository`
- [ ] Criar ou ajustar `OrdemDeServicoRepository`

**Integrações**

- [ ] Integrar com o caso de uso Processar Peças para Reserva e Compra
- [ ] Integrar com o caso de uso Processar Insumos para Reserva e Compra

**Transação**

- [ ] Garantir persistência transacional da aprovação e da atualização da OS

**Handler HTTP**

- [ ] Implementar `POST /orcamentos/{orcamentoId}/aprovar`
- [ ] Obter o cliente pelo JWT
- [ ] Criar DTO/response de saída
- [ ] Aplicar autenticação e autorização na rota, com o escopo `orcamentos:decidir`
- [ ] Aceitar o token de escopo reduzido emitido para a OS
- [ ] Retornar os erros `400`, `401`, `403`, `404` e `409`

**Testes unitários**

- [ ] Aprovação válida
- [ ] Aprovação com item faltante abrindo pedido de compra
- [ ] Orçamento inexistente
- [ ] Orçamento de outro cliente
- [ ] Orçamento já aprovado ou recusado
- [ ] Orçamento complementar sem principal vinculado

**Testes de integração**

- [ ] `200` com a OS em `AGUARDANDO_EXECUCAO` quando todos os itens estão em estoque
- [ ] `200` com a OS em `AGUARDANDO_RECURSOS` quando algum item precisa ser comprado
- [ ] Aprovação de complementar devolvendo a OS para a fila
- [ ] Persistência transacional da aprovação, da reserva e do pedido de compra
- [ ] `400`, `401`, `403`, `404` e `409`

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
