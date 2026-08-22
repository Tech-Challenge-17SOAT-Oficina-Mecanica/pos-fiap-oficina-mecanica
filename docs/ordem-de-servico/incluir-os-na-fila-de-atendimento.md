---
documento: Refinamento de Requisitos — Incluir OS na Fila de Atendimento
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Incluir OS na Fila de Atendimento

Este documento detalha a tarefa interna Incluir OS na Fila de Atendimento, do contexto de Ordem
de Serviço.

## 12 · Incluir OS na Fila de Atendimento

### 12.1 Refinamento de Produto

**Persona**

Sistema.

**Objetivo**

Disponibilizar a Ordem de Serviço na fila de atendimento quando ela estiver apta para iniciar ou
retomar a execução.

**Problema**

Após a aprovação de um orçamento principal ou complementar, a Ordem de Serviço precisa voltar ao
fluxo de execução e ficar disponível para consulta pelos mecânicos. Sem essa atualização, uma OS
aprovada pode não aparecer na fila ou pode perder o vínculo com o profissional que já a executava.

**Pré-condições**

- A Ordem de Serviço deve existir.
- O orçamento correspondente deve estar `APROVADO`.
- A Ordem de Serviço deve estar apta para iniciar ou retomar a execução.
- As peças e os insumos necessários devem estar disponíveis ou reservados.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-129 | Alterar a Ordem de Serviço para `AGUARDANDO_EXECUCAO`. |
| RF-OS-130 | Registrar a data e a hora da entrada mais recente na fila. |
| RF-OS-131 | Incluir a Ordem de Serviço na fila após a aprovação do orçamento `PRINCIPAL`. |
| RF-OS-132 | Permitir o retorno da Ordem de Serviço à fila após a aprovação de orçamento `COMPLEMENTAR`. |
| RF-OS-133 | Preservar o mecânico responsável, quando já existir. |
| RF-OS-134 | Permitir que a consulta da fila identifique a Ordem de Serviço como disponível. |
| RF-OS-135 | Atualizar `dataEntradaFila` sempre que a Ordem de Serviço retornar à fila. |
| RF-OS-136 | Representar a participação na fila pelos dados da própria Ordem de Serviço, sem criar registro duplicado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-70 | Executar a inclusão automaticamente após a aprovação do orçamento. |
| RNF-OS-71 | Persistir a mudança de situação e a data de entrada de forma consistente. |
| RNF-OS-72 | Não utilizar tabela ou agregado próprio para a fila de atendimento. |
| RNF-OS-73 | Não calcular prioridade durante a inclusão; a ordenação pertence à consulta da fila. |
| RNF-OS-74 | Preservar o vínculo com o mecânico durante a inclusão ou o retorno à fila. |

**Fluxo Principal — Orçamento Principal**

1. O orçamento principal é aprovado.
2. O sistema valida se a Ordem de Serviço está apta para execução.
3. O sistema valida a disponibilidade ou a reserva dos recursos necessários.
4. O sistema altera a Ordem de Serviço para `AGUARDANDO_EXECUCAO`.
5. O sistema registra `dataEntradaFila`.
6. A Ordem de Serviço passa a ficar disponível na consulta da fila.

**Fluxo Principal — Orçamento Complementar**

1. O orçamento complementar é aprovado.
2. O sistema valida se a Ordem de Serviço está apta para retomar a execução.
3. O sistema valida a disponibilidade ou a reserva dos recursos necessários.
4. O sistema altera a Ordem de Serviço para `AGUARDANDO_EXECUCAO`.
5. O sistema atualiza `dataEntradaFila` com a nova entrada.
6. O sistema preserva o mecânico responsável, quando houver.
7. A Ordem de Serviço volta a ficar disponível na consulta da fila.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | Interrompe a operação e informa a falha ao fluxo de aprovação. |
| A2 | Orçamento não aprovado | Impede a inclusão e mantém a Ordem de Serviço inalterada. |
| A3 | Peça ou insumo necessário indisponível e sem reserva | Impede a inclusão e informa quais recursos bloqueiam a execução. |
| A4 | Ordem de Serviço não apta para execução ou retomada | Impede a inclusão e mantém seus dados inalterados. |
| A5 | Falha ao persistir a alteração | Reverte a operação sem deixar situação ou data parcialmente atualizadas. |

**Saída**

- Ordem de Serviço disponível para consulta na fila de atendimento; ou
- Falha devolvida ao fluxo interno que solicitou a inclusão.

**Pós-condições**

- A Ordem de Serviço está em `AGUARDANDO_EXECUCAO`.
- `dataEntradaFila` contém a data e a hora da entrada mais recente.
- `mecanicoResponsavelId` permanece inalterado quando já estiver preenchido.
- Nenhum registro separado de fila é criado.

---

### 12.2 Refinamento Técnico

**Endpoint**

Esta tarefa não expõe endpoint. É um caso de uso interno disparado pelo fluxo de aprovação do
orçamento principal ou complementar.

> **Decisão de projeto.** A fila é uma visão derivada de `ordem_servico`: uma OS pertence à fila
> quando possui `status = AGUARDANDO_EXECUCAO` e `data_entrada_fila IS NOT NULL`. A alternativa
> seria persistir uma entidade `fila_atendimento`, mas isso duplicaria estado e exigiria
> sincronização entre a fila e a Ordem de Serviço.

**Autenticação / Autorização**

- Não há chamada direta por usuário.
- A autorização é herdada do fluxo que aprova o orçamento.
- O processo interno deve possuir permissão para consultar orçamento e estoque e alterar a Ordem
  de Serviço.

**Entrada**

| Origem | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Caso de uso | `ordemServicoId` | UUID | Identificador obrigatório da Ordem de Serviço. |
| Caso de uso | `orcamentoId` | UUID | Identificador do orçamento cuja aprovação disparou a inclusão. |

**Validações**

*Técnicas*

- `ordemServicoId` e `orcamentoId` devem ser UUIDs válidos.

*Negócio*

- A Ordem de Serviço e o orçamento devem existir e estar vinculados.
- O orçamento deve estar `APROVADO` e ser do tipo `PRINCIPAL` ou `COMPLEMENTAR`.
- A Ordem de Serviço deve permitir a transição para `AGUARDANDO_EXECUCAO`.
- As peças e os insumos necessários devem estar disponíveis ou reservados.
- A existência de `mecanicoResponsavelId` não impede a inclusão e o vínculo não pode ser removido.

**Processamento**

1. Receber `ordemServicoId` e `orcamentoId` do fluxo de aprovação.
2. Buscar a Ordem de Serviço e o orçamento correspondente.
3. Validar existência, vínculo, aprovação e tipo do orçamento.
4. Validar se a Ordem de Serviço pode iniciar ou retomar a execução.
5. Consultar a disponibilidade ou a reserva das peças e dos insumos necessários.
6. Alterar a situação da Ordem de Serviço para `AGUARDANDO_EXECUCAO`.
7. Registrar a data e a hora atuais em `dataEntradaFila`, substituindo a entrada anterior quando
   for um retorno à fila.
8. Preservar `mecanicoResponsavelId`, quando estiver preenchido.
9. Persistir a situação, a data e a nova versão da Ordem de Serviço.
10. Finalizar a operação para que a OS passe a aparecer na consulta da fila.

A operação não calcula prioridade. A regra de ordenação é aplicada por Consultar Fila de
Atendimento.

**Persistência**

- Consulta: `ordem_servico`, orçamento aprovado e disponibilidade ou reserva dos recursos.
- Altera: `ordem_servico.status`, `ordem_servico.data_entrada_fila` e
  `ordem_servico.version`.
- Preserva: `ordem_servico.mecanico_responsavel_id`.
- Não cria: tabela `fila_atendimento` ou qualquer registro separado de fila.
- A situação, a data de entrada e a versão devem ser persistidas atomicamente.

Uma Ordem de Serviço pertence à fila quando:

```sql
status = 'AGUARDANDO_EXECUCAO'
AND data_entrada_fila IS NOT NULL
```

**Saída da Operação**

```json
{
  "ordemServicoId": "550e8400-e29b-41d4-a716-446655440000",
  "status": "AGUARDANDO_EXECUCAO",
  "dataEntradaFila": "2026-08-21T21:10:00-03:00",
  "mecanicoResponsavelId": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "version": 8
}
```

`mecanicoResponsavelId` pode ser `null` quando a Ordem de Serviço ainda não possuir mecânico
vinculado.

**Códigos HTTP / Erros**

Não se aplicam diretamente, pois não há endpoint. Os erros são propagados e tratados pelo fluxo
de aprovação que disparou a operação:

- Ordem de Serviço ou orçamento não encontrado.
- Orçamento não aprovado ou não vinculado à Ordem de Serviço.
- Ordem de Serviço sem transição válida para `AGUARDANDO_EXECUCAO`.
- Recurso necessário indisponível e sem reserva.
- Falha ao persistir a alteração.

**Dependências**

- `OrdemDeServicoRepository`.
- Consulta ao orçamento aprovado.
- Consulta à disponibilidade e às reservas de estoque.
- Relógio da aplicação.
- Gerenciador de transações.
- Caso de uso Aprovar Orçamento.
- Caso de uso Consultar Fila de Atendimento.

**Testes**

*Unitários*

- Inclui a Ordem de Serviço após aprovação do orçamento principal.
- Reinclui a Ordem de Serviço após aprovação do orçamento complementar.
- Atualiza `dataEntradaFila` ao retornar à fila.
- Preserva `mecanicoResponsavelId` quando preenchido e aceita valor nulo.
- Rejeita orçamento não aprovado, recurso indisponível e Ordem de Serviço inexistente.
- Não cria estrutura própria de fila nem calcula prioridade.

*Integração*

- Aprovação principal deixa a Ordem de Serviço consultável na fila.
- Aprovação complementar devolve a Ordem de Serviço à fila com a nova data.
- Mecânico responsável permanece vinculado após o retorno.
- Falha durante a persistência não deixa situação e data parcialmente atualizadas.
- Consulta da fila encontra a OS por `AGUARDANDO_EXECUCAO` e `dataEntradaFila` preenchida.

---

### 12.3 Checklist de Implementação

**Domínio**

- [ ] Implementar a transição da Ordem de Serviço para `AGUARDANDO_EXECUCAO`
- [ ] Registrar ou atualizar `dataEntradaFila`
- [ ] Preservar `mecanicoResponsavelId`
- [ ] Representar a participação na fila pelos dados da própria Ordem de Serviço

**Caso de uso**

- [ ] Implementar `IncluirOSNaFilaAtendimento`
- [ ] Validar orçamento principal ou complementar aprovado
- [ ] Validar aptidão para execução ou retomada
- [ ] Obter a data e a hora por abstração de relógio
- [ ] Integrar o caso de uso ao fluxo de aprovação do orçamento

**Repositório**

- [ ] Buscar e persistir a Ordem de Serviço por `ordemServicoId`
- [ ] Persistir situação, `dataEntradaFila` e versão atomicamente
- [ ] Não criar repositório ou tabela de fila de atendimento

**Integrações**

- [ ] Consultar o orçamento aprovado e seu vínculo com a Ordem de Serviço
- [ ] Consultar a disponibilidade ou reserva das peças necessárias
- [ ] Consultar a disponibilidade ou reserva dos insumos necessários
- [ ] Integrar a operação com Aprovar Orçamento e Consultar Fila de Atendimento

**Validações**

- [ ] Validar existência e aptidão da Ordem de Serviço
- [ ] Validar orçamento `PRINCIPAL` ou `COMPLEMENTAR` aprovado
- [ ] Validar disponibilidade ou reserva dos recursos necessários
- [ ] Não exigir `mecanicoResponsavelId` vazio

**Concorrência**

- [ ] Proteger a atualização da versão da Ordem de Serviço contra concorrência
- [ ] Evitar que uma atualização simultânea remova o mecânico responsável

**Transação e idempotência**

- [ ] Persistir situação, data de entrada e versão na mesma transação
- [ ] Garantir rollback integral em caso de falha
- [ ] Garantir que repetição do gatilho não crie registro duplicado de fila

**Testes unitários**

- [ ] Inclusão após aprovação do orçamento principal
- [ ] Retorno após aprovação do orçamento complementar
- [ ] Atualização de `dataEntradaFila`
- [ ] Preservação do mecânico responsável
- [ ] Rejeição de orçamento não aprovado, recurso indisponível e OS inexistente

**Testes de integração**

- [ ] Integração com aprovação principal e complementar
- [ ] Ordem de Serviço disponível na consulta da fila
- [ ] Ausência de tabela ou registro próprio de fila
- [ ] Rollback integral diante de falha de persistência

**Documentação**

- [ ] Documentar o gatilho interno e a regra de pertencimento à fila no OpenAPI ou documentação técnica aplicável

**Review**

- [ ] Code Review aprovado

---
