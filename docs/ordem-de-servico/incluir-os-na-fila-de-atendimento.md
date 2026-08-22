---
documento: Refinamento de Requisitos — Incluir OS na Fila de Atendimento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Incluir OS na Fila de Atendimento

Este documento detalha a tarefa Incluir OS na Fila de Atendimento do contexto de Ordem de Serviço.

## 12 · Incluir OS na Fila de Atendimento

### 12.1 Refinamento de Produto

**Persona**

Sistema, acionado pela aprovação de um orçamento.

**Objetivo**

Disponibilizar a Ordem de Serviço na fila de atendimento quando ela estiver apta para execução ou
para retomada da execução.

**Problema**

Depois da aprovação de um orçamento principal ou complementar, a OS precisa voltar ao fluxo de
execução e ficar disponível para consulta pelos mecânicos.

**Pré-condições**

- Deve existir uma Ordem de Serviço.
- O orçamento correspondente deve estar aprovado.
- As peças e insumos necessários devem estar disponíveis ou reservados.
- A OS deve estar apta para execução.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-113 | Alterar a OS para `AGUARDANDO_EXECUCAO`. |
| RF-OS-114 | Registrar a data e hora de entrada na fila. |
| RF-OS-115 | Permitir a entrada da OS na fila após a aprovação de orçamento `PRINCIPAL`. |
| RF-OS-116 | Permitir o retorno da OS à fila após a aprovação de orçamento `COMPLEMENTAR`. |
| RF-OS-117 | Manter o mecânico responsável, caso já exista. |
| RF-OS-118 | Permitir que a consulta da fila identifique a OS como disponível. |
| RF-OS-119 | Atualizar `dataEntradaFila` quando a OS retornar à fila. |
| RF-OS-120 | Não criar registros duplicados de fila. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-60 | A inclusão deve ocorrer automaticamente. |
| RNF-OS-61 | A operação deve manter a consistência dos dados da OS. |
| RNF-OS-62 | A fila não deve possuir persistência própria. |
| RNF-OS-63 | A operação não deve calcular prioridade. |
| RNF-OS-64 | O vínculo com o mecânico não deve ser removido durante a inclusão. |

**Fluxo Principal — orçamento principal**

1. O orçamento principal é aprovado.
2. O sistema valida se a OS está apta para execução.
3. O sistema valida a disponibilidade ou reserva dos recursos necessários.
4. O sistema altera a OS para `AGUARDANDO_EXECUCAO`.
5. O sistema registra a `dataEntradaFila`.
6. A OS passa a ficar disponível na consulta da fila.

**Fluxo Principal — orçamento complementar**

1. O orçamento complementar é aprovado.
2. O sistema valida se a OS está apta para retomar a execução.
3. O sistema valida a disponibilidade ou reserva dos recursos necessários.
4. O sistema altera a OS para `AGUARDANDO_EXECUCAO`.
5. O sistema atualiza a `dataEntradaFila`.
6. O sistema mantém o mecânico responsável, caso exista.
7. A OS volta a ficar disponível na consulta da fila.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | O fluxo que disparou a operação trata o erro. |
| A2 | Orçamento não aprovado | A OS não entra na fila. |
| A3 | Peça ou insumo necessário indisponível | A OS não entra na fila. |
| A4 | OS não está apta para execução | A OS não entra na fila. |
| A5 | Falha ao persistir a alteração | Nenhuma alteração parcial permanece salva. |

**Saída**

- Ordem de Serviço disponível para consulta na fila de atendimento.

**Pós-condições**

- A OS está em `AGUARDANDO_EXECUCAO`.
- A `dataEntradaFila` está preenchida com a entrada mais recente.
- O `mecanicoResponsavelId` é preservado quando já existir.
- Nenhum registro separado de fila é criado.

---

### 12.2 Refinamento Técnico

**Gatilho**

Não há endpoint público: é um processamento interno disparado após a aprovação de um orçamento,
seja `PRINCIPAL` ou `COMPLEMENTAR`.

> **Decisão de projeto.** A fila **não é persistida** em tabela própria. Uma OS pertence à fila
> quando tem `status = AGUARDANDO_EXECUCAO` e `data_entrada_fila` preenchida — a própria Ordem de
> Serviço representa sua participação. Isso elimina duplicidade de registro e dispensa
> sincronização entre duas fontes. A prioridade não é calculada aqui: ela é aplicada apenas na
> consulta da fila.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Interno | `ordemServicoId` | uuid | Identificador da Ordem de Serviço. |

**Validações**

*Técnicas*

- `ordemServicoId` informado e em formato UUID válido.

*Negócio*

- A Ordem de Serviço deve existir.
- O orçamento correspondente deve estar `APROVADO`.
- A OS deve estar apta para execução.
- As peças e insumos necessários devem estar disponíveis ou reservados.
- Caso a OS já possua `mecanicoResponsavelId`, esse vínculo é mantido.
- Ao retornar à fila, a `dataEntradaFila` é atualizada para a nova entrada.

**Processamento**

1. Receber o `ordemServicoId`.
2. Buscar a Ordem de Serviço e validar sua existência.
3. Validar se o orçamento correspondente está aprovado.
4. Validar se as peças e insumos necessários estão disponíveis ou reservados.
5. Alterar o status da OS para `AGUARDANDO_EXECUCAO`.
6. Registrar a data e hora atual em `dataEntradaFila`.
7. Manter o `mecanicoResponsavelId`, caso esteja preenchido.
8. Persistir as alterações.

**Persistência**

- Altera: `ordem_servico` (`status = AGUARDANDO_EXECUCAO`, `data_entrada_fila`).
- Preserva: `mecanico_responsavel_id`.
- Não cria tabela `fila_atendimento` nem registro separado de fila.
- A atualização ocorre na mesma transação da aprovação do orçamento, ou de forma consistente com
  ela; em caso de erro, a OS não fica parcialmente atualizada.

**Resultado esperado**

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "status": "AGUARDANDO_EXECUCAO",
  "dataEntradaFila": "2026-08-21T21:10:00-03:00",
  "mecanicoResponsavelId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
}
```

O `mecanicoResponsavelId` pode ser nulo quando ainda não houver mecânico vinculado.

**Erros**

Por ser uma operação interna, os erros são tratados pelo fluxo que a disparou: OS não encontrada,
orçamento não aprovado, peça ou insumo necessário indisponível, falha ao persistir a alteração.

**Dependências**

- `OrdemDeServicoRepository`.
- Consulta ao orçamento aprovado.
- Consulta à disponibilidade ou reserva de estoque.

**Testes**

*Unitários*

- Coloca a OS em `AGUARDANDO_EXECUCAO` e registra a `dataEntradaFila`.
- Mantém o mecânico responsável, caso já exista.
- Atualiza a `dataEntradaFila` no retorno à fila.
- Não calcula prioridade.

*Integração*

- Inclusão após a aprovação do orçamento `PRINCIPAL`.
- Retorno à fila após a aprovação do orçamento `COMPLEMENTAR`.
- Orçamento não aprovado não coloca a OS na fila.
- Recursos necessários indisponíveis não colocam a OS na fila.
- Rollback em caso de falha, sem deixar a OS parcialmente atualizada.

---

### 12.3 Checklist de Implementação

**Domínio**

- [ ] Alterar o status da OS para `AGUARDANDO_EXECUCAO`
- [ ] Registrar `dataEntradaFila` com a data e hora atual
- [ ] Atualizar `dataEntradaFila` quando a OS retornar à fila
- [ ] Manter `mecanicoResponsavelId` quando já estiver preenchido
- [ ] Não exigir `mecanicoResponsavelId` vazio e não remover o mecânico existente
- [ ] Não criar tabela `fila_atendimento` nem registro separado de fila
- [ ] Não calcular prioridade nesta tarefa
- [ ] Garantir que a fila seja derivada da própria `ordem_servico`

**Caso de uso**

- [ ] Implementar `IncluirOSNaFilaAtendimento`
- [ ] Validar que a OS exista
- [ ] Validar que o orçamento correspondente esteja `APROVADO`
- [ ] Validar que a OS esteja apta para execução
- [ ] Validar disponibilidade ou reserva das peças e dos insumos necessários
- [ ] Garantir que a operação funcione para orçamento `PRINCIPAL` e `COMPLEMENTAR`

**Repositório**

- [ ] Criar ou ajustar `OrdemDeServicoRepository`
- [ ] Persistir status e `data_entrada_fila` na OS

**Integrações**

- [ ] Integrar o caso de uso ao fluxo de aprovação do orçamento
- [ ] Garantir consistência com a transação da aprovação
- [ ] Garantir rollback em caso de erro

**Testes unitários**

- [ ] Inclusão após aprovação do orçamento `PRINCIPAL`
- [ ] Retorno após aprovação do orçamento `COMPLEMENTAR`
- [ ] Manutenção do mecânico responsável
- [ ] Atualização de `dataEntradaFila`
- [ ] Orçamento não aprovado
- [ ] Recursos necessários indisponíveis
- [ ] OS inexistente

**Testes de integração**

- [ ] Integração com o fluxo de aprovação do orçamento
- [ ] Rollback em caso de falha

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite da task

---
