---
documento: Refinamento de Requisitos — Calcular Orçamento
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Calcular Orçamento

Este documento detalha a tarefa Calcular Orçamento do contexto de Orçamento.

## 1 · Calcular Orçamento

### 1.1 Refinamento de Produto

**Persona**

Sistema.

**Objetivo**

Calcular e atualizar os valores dos itens do orçamento existente, o valor total geral da OS e sua estimativa de entrega em dias.

**Problema**

A oficina precisa consolidar os valores de serviços, peças e insumos de uma OS e informar uma estimativa de entrega baseada na disponibilidade dos itens, no tempo médio dos serviços e na fila de atendimento.

**Pré-condições**

- Deve existir uma OS com orçamento principal associado.
- Toda OS deve possuir um único orçamento `PRINCIPAL`.
- O orçamento complementar é opcional.
- Um orçamento `COMPLEMENTAR` nunca pode existir sem um orçamento `PRINCIPAL` da mesma OS.
- Todo orçamento existente deve possuir `tipoOrcamento` preenchido como `PRINCIPAL` ou `COMPLEMENTAR`.
- O orçamento deve estar com status `CRIADO`.
- Quando existir orçamento complementar, ele deve possuir `orcamentoOriginalId` apontando para o orçamento principal da mesma OS.
- O orçamento deve possuir serviços, peças ou insumos vinculados.
- Todos os itens devem possuir quantidade e valor unitário válidos.
- Devem existir dados de prazo de entrega para peças e insumos indisponíveis.
- Deve existir histórico ou configuração de tempo médio dos serviços.
- Deve estar definida a quantidade de OS atendidas por dia.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-ORC-01 | Permitir calcular um orçamento existente. |
| RF-ORC-02 | Calcular o valor total de cada item do orçamento. |
| RF-ORC-03 | Atualizar os valores calculados dos itens. |
| RF-ORC-04 | Calcular o `valorTotalGeral` da OS. |
| RF-ORC-05 | Considerar no `valorTotalGeral` os itens do orçamento principal e dos complementares com status `CRIADO` ou `APROVADO`. |
| RF-ORC-06 | Não considerar no `valorTotalGeral` itens de orçamentos com status `RECUSADO`. |
| RF-ORC-07 | Retornar somente o `valorTotalGeral` da OS no resultado do cálculo. |
| RF-ORC-36 | Calcular a estimativa do orçamento principal. |
| RF-ORC-37 | Calcular a estimativa de orçamento complementar somente quando ele existir. |
| RF-ORC-38 | Considerar prazo de entrega de peças e insumos, tempo médio do serviço e posição na fila. |
| RF-ORC-39 | Considerar a capacidade diária de atendimento da oficina. |
| RF-ORC-40 | Para orçamento complementar, considerar a estimativa já existente do orçamento principal vinculado. |
| RF-ORC-41 | Impedir orçamento complementar sem orçamento principal vinculado. |
| RF-ORC-42 | Registrar a estimativa em dias, sem informar uma data exata de entrega. |
| RF-ORC-43 | Manter o status atual do orçamento após o cálculo. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-01 | A atualização dos itens, do valor total geral e da estimativa deve ocorrer de forma transacional. |
| RNF-ORC-02 | Valores monetários devem usar precisão decimal. |
| RNF-ORC-03 | O cálculo deve ser rastreável. |
| RNF-ORC-16 | A estimativa deve ser armazenada em dias inteiros. |
| RNF-ORC-17 | Valores e dados utilizados no cálculo devem ser preservados para auditoria. |

**Fluxo Principal**

1. O sistema recebe o identificador de um orçamento existente.
2. O sistema consulta o orçamento, a OS e seus itens.
3. O sistema identifica se o orçamento é `PRINCIPAL` ou `COMPLEMENTAR`.
4. Caso seja complementar, o sistema valida o vínculo com o orçamento principal da mesma OS.
5. O sistema valida quantidade e valor unitário dos itens.
6. O sistema calcula o valor total de cada item.
7. O sistema consulta os demais orçamentos da mesma OS.
8. O sistema calcula o `valorTotalGeral` pela soma dos valores dos itens dos orçamentos não recusados.
9. O sistema consulta o prazo de peças e insumos necessários.
10. O sistema consulta o tempo médio dos serviços.
11. O sistema identifica as OS à frente na fila.
12. O sistema calcula o impacto da fila conforme a capacidade diária da oficina.
13. O sistema calcula a estimativa de entrega em dias.
14. O sistema atualiza os valores dos itens e a estimativa do orçamento.
15. O sistema retorna o `valorTotalGeral` da OS.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Orçamento não encontrado | Informa que o orçamento não existe. |
| A2 | Orçamento fora de `CRIADO` | Impede o cálculo. |
| A3 | Tipo de orçamento inválido | Impede o cálculo. |
| A4 | Ausência de orçamento complementar | Calcula normalmente apenas o orçamento principal. |
| A5 | Orçamento complementar sem orçamento principal vinculado | Impede o cálculo complementar. |
| A6 | Item com quantidade ou valor unitário inválido | Impede o cálculo. |
| A7 | Prazo de peça ou insumo indisponível | Informa que não é possível calcular a estimativa completa. |
| A8 | Tempo médio do serviço indisponível | Impede o cálculo. |
| A9 | Capacidade diária não configurada | Impede o cálculo. |
| A10 | Falha na persistência | Nenhuma informação parcial permanece gravada. |

**Saída**

- Orçamento atualizado com valores dos itens, `valorTotalGeral` da OS e estimativa de entrega em dias.

**Pós-condições**

- Os valores dos itens ficam calculados e registrados.
- O `valorTotalGeral` da OS fica disponível no retorno da operação.
- A estimativa de entrega fica registrada no orçamento.
- O orçamento permanece com o mesmo status.
- Nenhuma data exata de entrega é informada.
- O orçamento principal permanece como referência para eventuais orçamentos complementares.
- A ausência de orçamento complementar não impede o cálculo do orçamento principal.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /orcamentos/{orcamentoId}/calcular
```

O endpoint recalcula e atualiza um orçamento já existente; não cria um novo orçamento.

> **Decisão de projeto — quando o cálculo é acionado.** Duas vezes, e sempre pela oficina, nunca
> pelo cliente: **ao fim do diagnóstico**, depois de registrados os problemas, serviços, peças e
> insumos, imediatamente antes de enviar o orçamento ao cliente; e **ao fechar cada complementar**,
> depois de registrado o que foi encontrado durante a execução. Registrar item não recalcula
> sozinho — o cálculo é um passo explícito, para o valor não mudar debaixo do cliente enquanto o
> mecânico ainda está lançando itens.

> **Decisão de projeto.** O `status` do **orçamento** é a fonte da verdade da decisão do cliente, e
> o `status` da **OS** é a fonte da verdade da etapa do atendimento. Por isso o cálculo valida
> `statusOrcamento = CRIADO` e não olha o status da OS.

> **Decisão de projeto.** O complementar é um **orçamento separado**, com identificador próprio,
> `tipoOrcamento` e `orcamentoOriginalId` apontando para o principal — e não uma adição dentro de
> um orçamento único (D-17).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Permitido para usuário ou processo interno autorizado a calcular orçamentos.
- Escopo: `orcamentos:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `orcamentoId` | uuid | Identificador do orçamento a calcular. |

**Validações**

*Técnicas*

- `orcamentoId` deve possuir formato UUID válido.

*Negócio*

- O orçamento deve existir.
- `statusOrcamento` deve estar em `CRIADO`.
- `tipoOrcamento` deve ser `PRINCIPAL` ou `COMPLEMENTAR`.
- A OS deve estar vinculada ao orçamento.
- Quando o tipo for `PRINCIPAL`, `orcamentoOriginalId` deve permanecer vazio.
- Quando o tipo for `COMPLEMENTAR`, deve existir orçamento principal para a mesma OS.
- Quando o tipo for `COMPLEMENTAR`, `orcamentoOriginalId` deve referenciar o orçamento principal da mesma OS.
- Deve existir ao menos um item no orçamento.
- Quantidade e valor unitário de todos os itens devem ser válidos.
- Os itens devem possuir prazo de entrega, quando necessário.
- Deve existir tempo médio para os serviços.
- A capacidade diária de OS deve estar configurada.

**Processamento**

1. Receber o identificador do orçamento.
2. Consultar o orçamento, a OS e seus itens.
3. Identificar o tipo do orçamento.
4. Validar os itens, as quantidades e os valores unitários.
5. Calcular o valor de cada item:

```text
valorTotalItem = quantidade × valorUnitario
```

6. Consultar os orçamentos principal e complementares da mesma OS com status `CRIADO` ou `APROVADO`.
7. Calcular o valor total geral da OS:

```text
valorTotalGeral = soma(valorTotalItem) dos itens dos orçamentos da OS com status CRIADO ou APROVADO
```

8. Consultar o prazo de entrega das peças e insumos necessários.
9. Consultar o tempo médio dos serviços.
10. Consultar a quantidade de OS à frente na fila.
11. Consultar a capacidade diária de atendimento.
12. Calcular os dias de fila:

```text
diasFila = arredondarParaCima(quantidadeOsNaFrente / capacidadeDiariaOs)
```

Exemplo: se são atendidas 3 OS por dia e existem 3 OS à frente, a quarta OS terá mais 1 dia de fila.

13. Quando o orçamento for `PRINCIPAL`, calcular:

```text
estimativaEntregaDias = prazoPecasInsumosDias + tempoMedioServicosDias + diasFila
```

14. Quando o orçamento for `COMPLEMENTAR`, consultar a estimativa do orçamento principal vinculado e calcular:

```text
estimativaEntregaDias = estimativaOrcamentoPrincipalDias + prazoPecasInsumosAdicionaisDias + tempoMedioServicosAdicionaisDias
```

15. Atualizar os valores dos itens e a estimativa de entrega do orçamento.
16. Manter `statusOrcamento` sem alteração.
17. Persistir tudo em uma única transação.
18. Retornar somente `valorTotalGeral` no corpo da resposta.

**Persistência**

*Orçamento*

- Manter `tipoOrcamento` como `PRINCIPAL` ou `COMPLEMENTAR`.
- Manter `statusOrcamento` como `CRIADO`, `APROVADO` ou `RECUSADO`.
- Manter `orcamentoOriginalId` vazio para principal e obrigatório para complementar.
- Atualizar `estimativaEntregaDias`.
- Atualizar `dataAtualizacao`.

*Item do Orçamento*

- Atualizar `quantidade`.
- Atualizar `valorUnitario`.
- Atualizar `valorTotal`.

O `valorTotalGeral` é calculado pela soma dos valores dos itens dos orçamentos válidos da OS e retornado pela operação; não precisa ser persistido como campo duplicado.

**Saída da API**

Exemplo de cálculo do orçamento principal:

```json
{
  "orcamentoId": "uuid-do-orcamento-principal",
  "ordemServicoId": "uuid-da-os",
  "tipoOrcamento": "PRINCIPAL",
  "statusOrcamento": "CRIADO",
  "itens": [
    {
      "tipo": "SERVICO",
      "descricao": "Troca de óleo",
      "quantidade": 1,
      "valorUnitario": 150.00,
      "valorTotal": 150.00
    },
    {
      "tipo": "PECA",
      "descricao": "Filtro de óleo",
      "quantidade": 1,
      "valorUnitario": 50.00,
      "valorTotal": 50.00
    }
  ],
  "valorTotalGeral": 200.00,
  "estimativaEntregaDias": 4
}
```

Exemplo de cálculo de orçamento complementar:

```json
{
  "orcamentoId": "uuid-do-orcamento-complementar",
  "ordemServicoId": "uuid-da-os",
  "tipoOrcamento": "COMPLEMENTAR",
  "orcamentoOriginalId": "uuid-do-orcamento-principal",
  "statusOrcamento": "CRIADO",
  "itens": [
    {
      "tipo": "PECA",
      "descricao": "Correia dentada",
      "quantidade": 1,
      "valorUnitario": 150.00,
      "valorTotal": 150.00
    }
  ],
  "valorTotalGeral": 350.00,
  "estimativaEntregaDias": 6
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| 200 OK | Orçamento calculado com sucesso. |
| 400 Bad Request | Identificador inválido. |
| 401 Unauthorized | Usuário não autenticado. |
| 403 Forbidden | Usuário sem permissão. |
| 404 Not Found | Orçamento não encontrado. |
| 409 Conflict | Orçamento não está em `CRIADO` ou não permite cálculo. |
| 409 Conflict | Itens ou dados insuficientes para cálculo. |
| 500 Internal Server Error | Erro inesperado. |

**Dependências**

- `OrcamentoRepository`.
- `OrcamentoItemRepository`.
- `OrdemDeServicoRepository`.
- Repositório ou serviço de peças e insumos.
- Repositório ou serviço de tempo médio de execução.
- Consulta da fila de atendimento.
- Configuração de capacidade diária da oficina.
- Banco de dados.

**Testes**

*Unitários*

- Deve calcular valores dos itens para orçamento principal.
- Deve calcular valores dos itens para orçamento complementar.
- Deve retornar `valorTotalGeral` igual à soma dos itens do orçamento principal quando não houver complementar.
- Deve retornar `valorTotalGeral` com a soma dos itens do principal e complementar.
- Não deve incluir itens de orçamento com status `RECUSADO` no `valorTotalGeral`.
- Deve calcular estimativa para orçamento principal sem orçamento complementar.
- Deve calcular estimativa para orçamento complementar quando existir orçamento principal vinculado.
- Deve impedir orçamento complementar sem orçamento principal.
- Deve impedir orçamento complementar com `orcamentoOriginalId` de outra OS.
- Deve considerar prazo de peças e insumos, tempo médio dos serviços e OS à frente na fila.
- Deve calcular corretamente a fila conforme a capacidade diária.
- Deve atualizar valores e estimativa sem alterar o status do orçamento.
- Deve garantir rollback quando houver falha durante a persistência.

*Integração*

- Deve retornar `400` para identificador inválido.
- Deve retornar `401` sem autenticação.
- Deve retornar `403` sem permissão.
- Deve retornar `404` para orçamento inexistente.
- Deve retornar `409` para orçamento aprovado ou recusado.
- Deve retornar `409` para dados insuficientes para cálculo, e `400` para item inválido.

---

### 1.3 Check-list de Implementação

**Domínio**

- [ ] Criar/ajustar `TipoOrcamento` com `PRINCIPAL` e `COMPLEMENTAR`.
- [ ] Criar/ajustar `StatusOrcamento` com `CRIADO`, `APROVADO` e `RECUSADO`.
- [ ] Garantir que toda OS possua um orçamento principal.
- [ ] Garantir que orçamento complementar seja opcional e não exista sem orçamento principal.
- [ ] Garantir que `orcamentoOriginalId` seja vazio para principal e obrigatório para complementar.
- [ ] Criar/ajustar `estimativaEntregaDias` e `dataAtualizacao`.
- [ ] Criar/ajustar `valorTotal` do item do orçamento.

**Caso de uso**

- [ ] Implementar o caso de uso `CalcularOrcamento`.
- [ ] Implementar cálculo de valores dos itens.
- [ ] Implementar cálculo de `valorTotalGeral` pela soma dos valores dos itens.
- [ ] Garantir que orçamentos `RECUSADO` não participem do `valorTotalGeral`.
- [ ] Implementar cálculo de estimativa para orçamento principal e complementar.
- [ ] Implementar cálculo de dias da fila.

**Repositório e integrações**

- [ ] Criar/ajustar `OrcamentoRepository` e `OrcamentoItemRepository`.
- [ ] Criar/ajustar `OrdemDeServicoRepository`.
- [ ] Implementar consulta de todos os orçamentos válidos da mesma OS.
- [ ] Implementar consulta de prazo de peças e insumos.
- [ ] Implementar consulta de tempo médio dos serviços.
- [ ] Definir e persistir a capacidade diária de OS atendidas.

**Handler HTTP**

- [ ] Criar handler para `POST /orcamentos/{orcamentoId}/calcular`.
- [ ] Aplicar autenticação e autorização na rota.
- [ ] Retornar somente `valorTotalGeral` no resultado do cálculo.
- [ ] Retornar erros `400`, `401`, `403`, `404` e `409`.

**Testes unitários**

- [ ] Criar testes para valores, total geral, estimativas e regras de orçamento principal/complementar.

**Testes de integração**

- [ ] Criar teste de integração do endpoint.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Executar testes automatizados.
- [ ] Realizar code review.
- [ ] Validar critérios de aceite da task.
