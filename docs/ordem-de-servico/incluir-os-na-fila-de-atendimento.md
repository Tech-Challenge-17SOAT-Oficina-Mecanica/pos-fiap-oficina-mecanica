---
documento: Refinamento de Requisitos — Incluir OS na Fila de Atendimento
dono: Desconhecido
versao: 0.1
atualizado_em: 2026-08-20
status: rascunho
---

# Refinamento de Requisitos — Incluir OS na Fila de Atendimento

Este documento detalha a tarefa Incluir OS na Fila de Atendimento do contexto de Ordem de Serviço.

## 1 · Incluir OS na Fila de Atendimento

### 1.1 Refinamento de Produto

**Persona**

Sistema.

**Objetivo**

Disponibilizar a OS na fila de atendimento quando ela estiver apta para execução.

**Problema**

A OS aprovada precisa ficar disponível para que um mecânico possa selecioná-la e iniciar a execução.

**Pré-condições**

- Deve existir uma OS.
- A OS deve estar com status `AGUARDANDO_EXECUCAO`.
- A OS não deve possuir mecânico responsável.
- A OS não deve estar cancelada.

**Requisitos Funcionais**

- Identificar OS em `AGUARDANDO_EXECUCAO`.
- Disponibilizar a OS na fila de atendimento.
- Registrar a data/hora de entrada na fila, se esse dado for necessário para ordenação.
- Impedir que OS já assumida por mecânico apareça como disponível.

**Requisitos Não Funcionais**

- A disponibilização deve ocorrer automaticamente.
- A operação deve ser persistida.
- Não deve haver duplicidade de OS disponível na fila.

**Fluxo Principal**

1. O orçamento é aprovado.
2. A OS é atualizada para `AGUARDANDO_EXECUCAO`.
3. O sistema registra a data/hora de entrada na fila, se aplicável.
4. A OS passa a aparecer na fila de atendimento.
5. O mecânico pode consultar e selecionar a OS.

**Fluxos Alternativos / Exceções**

- OS fora de `AGUARDANDO_EXECUCAO`: o sistema não a disponibiliza.
- OS já associada a mecânico: o sistema não a disponibiliza.
- Erro ao atualizar a OS: o sistema mantém o status anterior.

**Saída**

- OS disponível na fila de atendimento.

**Pós-condições**

- OS permanece em `AGUARDANDO_EXECUCAO`.
- OS fica disponível para seleção por um mecânico.
- Data/hora de entrada na fila registrada, se aplicável.

### 1.2 Refinamento Técnico

**Gatilho**

- Após a aprovação do orçamento, quando a OS é atualizada para `AGUARDANDO_EXECUCAO`.
- Processamento interno do Sistema.

**Entrada**

- `ordemServicoId`: identificador da OS.

**Validações**

- Validar se a OS existe.
- Validar se a OS está em `AGUARDANDO_EXECUCAO`.
- Validar se a OS não possui mecânico responsável.
- Validar se a OS não está cancelada.

**Processamento**

1. Receber o identificador da OS.
2. Consultar a OS.
3. Validar se ela está apta para execução.
4. Registrar `dataEntradaFila`, caso esse campo seja necessário.
5. Persistir a alteração.
6. Disponibilizar a OS na consulta da fila de atendimento.

**Persistência**

- Atualizar Ordem de Serviço:
  - `status`: permanece `AGUARDANDO_EXECUCAO`;
  - `dataEntradaFila`, se utilizada para ordenação;
  - `mecanicoResponsavelId`: permanece vazio.

**Dependências**

- `OrdemDeServicoRepository`.
- Banco de dados.

**Testes**

- Deve disponibilizar OS em `AGUARDANDO_EXECUCAO`.
- Não deve disponibilizar OS cancelada.
- Não deve disponibilizar OS já associada a mecânico.
- Deve registrar a data de entrada na fila, se o campo existir.
- Deve garantir que a OS não apareça duplicada na consulta da fila.
- Deve garantir que nenhum status seja alterado nesta operação.

### 1.3 Check-list de Implementação

- [ ] Definir se a OS terá o campo `dataEntradaFila`.
- [ ] Criar/ajustar o campo `dataEntradaFila`, se necessário.
- [ ] Criar/ajustar `OrdemDeServicoRepository`.
- [ ] Implementar a regra de disponibilização da OS na fila.
- [ ] Garantir que somente OS em `AGUARDANDO_EXECUCAO` apareçam na fila.
- [ ] Garantir que OS com mecânico responsável não apareçam na fila.
- [ ] Registrar a data/hora de entrada na fila, se aplicável.
- [ ] Integrar a regra ao fluxo de Aprovar Orçamento.
- [ ] Garantir que esta operação não altera o status da OS.
- [ ] Criar testes para OS apta, cancelada e já assumida.
- [ ] Criar teste para não duplicidade na consulta da fila.
- [ ] Executar testes automatizados, code review e validar critérios de aceite.
