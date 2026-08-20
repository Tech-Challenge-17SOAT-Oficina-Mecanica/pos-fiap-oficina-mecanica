---
documento: Refinamento de Requisitos — Gerar Orçamento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Gerar Orçamento

Este documento detalha a tarefa Gerar Orçamento do contexto de Orçamento.

## 1 · Gerar Orçamento

### 1.1 Refinamento de Produto

**Persona**

Sistema, acionado pela finalização do diagnóstico.

**Objetivo**

Calcular e criar o orçamento da Ordem de Serviço após a finalização do diagnóstico.

**Problema**

A oficina precisa consolidar serviços, peças e insumos identificados no diagnóstico para que o
cliente possa analisar o custo do atendimento.

**Pré-condições**

- A Ordem de Serviço existe e está com status `EM_DIAGNOSTICO`.
- O diagnóstico está finalizado.
- Existem serviços, peças ou insumos vinculados à OS.
- Todos os itens possuem quantidade e valor válidos.
- Não existe outro orçamento principal pendente de decisão para a mesma OS.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-ORC-01 | Gerar orçamento a partir dos itens identificados no diagnóstico. |
| RF-ORC-02 | Calcular o valor individual e o valor total dos itens. |
| RF-ORC-03 | Associar o orçamento à OS. |
| RF-ORC-04 | Registrar o orçamento criado. |
| RF-ORC-05 | Definir o tipo do orçamento como `PRINCIPAL`. |
| RF-ORC-06 | Atualizar a OS para o status `AGUARDANDO_APROVACAO`. |
| RF-ORC-07 | Impedir a geração duplicada de orçamento principal para a mesma OS. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-01 | A criação do orçamento e de seus itens deve ser transacional. |
| RNF-ORC-02 | Valores monetários devem usar precisão decimal. |
| RNF-ORC-03 | A operação deve ser registrada para auditoria e rastreabilidade. |

**Fluxo Principal**

1. O diagnóstico da OS é finalizado.
2. O sistema identifica serviços, peças e insumos necessários.
3. O sistema valida os itens e seus valores.
4. O sistema calcula os valores e o total do orçamento.
5. O sistema cria o orçamento principal e seus itens.
6. O sistema associa o orçamento à OS.
7. O sistema atualiza a OS para `AGUARDANDO_APROVACAO`.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Não há itens para composição | Não cria o orçamento. |
| A2 | Item sem valor ou com quantidade inválida | Impede a geração e informa a pendência. |
| A3 | Já existe orçamento principal pendente de decisão | Não cria outro orçamento principal. |
| A4 | Falha na persistência | Nenhuma informação parcial permanece gravada. |

**Saída**

- Orçamento principal criado e associado à OS, com a OS atualizada para `AGUARDANDO_APROVACAO`.

**Pós-condições**

- Orçamento e itens persistidos, com o total calculado.
- A OS fica disponível para o próximo fluxo: envio do orçamento ao cliente.
- O orçamento não recebe status próprio.

---

### 1.2 Refinamento Técnico

**Gatilho**

Não há endpoint público: depois da finalização do diagnóstico, o sistema executa internamente a
geração do orçamento.

> **Decisão de projeto.** O orçamento **não tem status próprio**: a etapa do fluxo é controlada
> pelo status da Ordem de Serviço. A decisão do cliente é registrada nos campos de aprovação e
> recusa do próprio orçamento. A alternativa — um `status` no orçamento — foi descartada por
> duplicar a máquina de estados da OS. No MVP a geração é síncrona, sem fila, mensageria, retry
> nem DLQ.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Interno | `ordemServicoId` | uuid | Identificador da Ordem de Serviço cujo diagnóstico foi finalizado. |

**Validações**

*Técnicas*

- `ordemServicoId` informado e em formato UUID válido.

*Negócio*

- A OS deve existir e estar em `EM_DIAGNOSTICO`, com o diagnóstico finalizado.
- Deve existir ao menos um serviço, peça ou insumo vinculado à OS.
- Todos os itens devem ter quantidade e valor unitário válidos.
- Não pode existir outro orçamento principal pendente de decisão para a OS.
- A OS deve permitir a transição para `AGUARDANDO_APROVACAO`.

**Processamento**

1. Receber o identificador da OS.
2. Consultar a OS e os itens identificados no diagnóstico.
3. Validar os itens e seus valores.
4. Calcular `valorTotal = quantidade × valorUnitario` para cada item.
5. Calcular o total do orçamento.
6. Criar o orçamento com `tipo = PRINCIPAL`.
7. Criar os itens do orçamento.
8. Associar o orçamento à OS.
9. Atualizar o status da OS para `AGUARDANDO_APROVACAO`.
10. Persistir tudo em uma única transação.
11. Registrar a operação em log.

**Persistência**

- Consulta: `ordem_servico` e os itens do diagnóstico.
- Altera: `orcamento` (insert), `orcamento_item` (insert), `ordem_servico.status`.

Campos de `orcamento`: `id`, `ordem_servico_id`, `tipo` (`PRINCIPAL` ou `COMPLEMENTAR`),
`orcamento_original_id` (preenchido apenas quando o tipo for `COMPLEMENTAR`), `valor_total`,
`data_geracao`.

Campos de `orcamento_item`: `id`, `orcamento_id`, `tipo`, `item_id`, `descricao`, `quantidade`,
`valor_unitario`, `valor_total`.

**Resultado do processamento**

```json
{
  "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "tipo": "PRINCIPAL",
  "orcamentoOriginalId": null,
  "itens": [
    {
      "tipo": "SERVICO",
      "itemId": "7b4e08d5-3c61-4f92-a0d7-51e83b62c40f",
      "descricao": "Troca de óleo",
      "quantidade": 1,
      "valorUnitario": 150.0,
      "valorTotal": 150.0
    }
  ],
  "valorTotal": 150.0,
  "statusOrdemServico": "AGUARDANDO_APROVACAO"
}
```

**Erros**

| Situação | Comportamento |
|---|---|
| OS não encontrada | Não gera orçamento. |
| Diagnóstico não finalizado | Não gera orçamento. |
| Ausência de itens ou item sem valor válido | Não gera orçamento. |
| Orçamento principal pendente já existente | Não gera novo orçamento principal. |
| Falha inesperada | Reverte toda a operação. |

**Dependências**

- `OrdemDeServicoRepository`.
- `OrcamentoRepository`.
- `OrcamentoItemRepository`.
- Consulta de serviços, peças, insumos e estoque.

**Testes**

*Unitários*

- Cria orçamento principal quando o diagnóstico finalizado possui itens válidos.
- Registra o tipo `PRINCIPAL` e mantém `orcamentoOriginalId` vazio.
- Calcula corretamente os valores por item e o total.
- Impede orçamento principal duplicado.
- Rejeita OS inexistente, diagnóstico não finalizado, ausência de itens e item sem valor.
- Garante que o orçamento não recebe status próprio.

*Integração*

- A OS passa para `AGUARDANDO_APROVACAO` após a geração.
- Orçamento e itens são persistidos na mesma transação.
- Falha durante a persistência provoca rollback completo.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o agregado `Orcamento`
- [ ] Criar ou ajustar a entidade `OrcamentoItem`
- [ ] Definir o campo `tipo` com os valores `PRINCIPAL` e `COMPLEMENTAR`
- [ ] Criar ou ajustar o campo `orcamentoOriginalId`
- [ ] Garantir que orçamento complementar seja vinculado a orçamento principal da mesma OS
- [ ] Definir o tipo de item: serviço, peça e insumo
- [ ] Definir valores monetários com precisão decimal
- [ ] Não criar campo de status no orçamento

**Caso de uso**

- [ ] Implementar `GerarOrcamento`
- [ ] Validar OS existente, em `EM_DIAGNOSTICO` e com diagnóstico finalizado
- [ ] Validar itens, quantidades e valores
- [ ] Implementar o cálculo dos subtotais e do valor total
- [ ] Impedir orçamento principal duplicado
- [ ] Atualizar a OS para `AGUARDANDO_APROVACAO`

**Repositório**

- [ ] Criar `OrcamentoRepository` e `OrcamentoItemRepository`
- [ ] Implementar a consulta da OS e dos itens do diagnóstico

**Transação**

- [ ] Criar orçamento principal e itens em uma única transação
- [ ] Garantir rollback integral em caso de falha

**Testes unitários**

- [ ] Cálculo dos subtotais e do valor total
- [ ] Tipo `PRINCIPAL` com `orcamentoOriginalId` vazio
- [ ] Orçamento principal duplicado
- [ ] OS inexistente, diagnóstico não finalizado, ausência de itens e item sem valor

**Testes de integração**

- [ ] Criação transacional do orçamento e dos itens
- [ ] Transição da OS para `AGUARDANDO_APROVACAO`
- [ ] Rollback em caso de falha na persistência

**Documentação**

- [ ] Documentar o fluxo e a estrutura no Swagger/OpenAPI, quando houver endpoint
- [ ] Criar DTO de resultado com `tipo` e `orcamentoOriginalId`

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
