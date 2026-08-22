---
documento: Resumo do Contexto — Ordem de Serviço
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Ordem de Serviço

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. O que ainda não está resolvido fica em
[`pontos-em-aberto.md`](pontos-em-aberto.md).

## O que este contexto cobre

O ciclo de vida do atendimento, do recebimento do veículo até a entrega: abertura da OS,
diagnóstico, registro dos itens necessários, fila de atendimento, execução, finalização, entrega e
os indicadores de acompanhamento. É o contexto central do sistema — quase todos os outros são
consultados por ele ou reagem a ele.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 3 | Criar Ordem de Serviço | `POST /ordens-servico` | `os:escrever` | [criar-ordem-de-servico.md](criar-ordem-de-servico.md) |
| 1 | Registrar Problema Relatado | `POST /ordens-servico/{osId}/problema-relatado` | `os:escrever` | [registrar-problema-relatado.md](registrar-problema-relatado.md) |
| 11 | Registrar Problema Encontrado | `POST /ordens-servico/{osId}/problemas` | `os:escrever` | [registrar-problema-encontrado.md](registrar-problema-encontrado.md) |
| 2 | Registrar Serviços Necessários | `POST /ordens-servico/{osId}/servicos` | `os:escrever` | [registrar-servicos-necessarios.md](registrar-servicos-necessarios.md) |
| 13 | Registrar Peças e Insumos Necessários | rota a definir | `os:escrever` | [registrar-pecas-e-insumos-necessarios.md](registrar-pecas-e-insumos-necessarios.md) |
| 12 | Incluir OS na Fila de Atendimento | sem endpoint, caso de uso interno | — | [incluir-os-na-fila-de-atendimento.md](incluir-os-na-fila-de-atendimento.md) |
| 4 | Consultar Fila de Atendimento | `GET /fila-atendimento` | `os:ler` | [consultar-fila-de-atendimento.md](consultar-fila-de-atendimento.md) |
| 5 | Iniciar Execução | `POST /ordens-servico/{osId}/execucao/iniciar` | `os:escrever` | [iniciar-execucao.md](iniciar-execucao.md) |
| 7 | Finalizar Serviço | `POST /ordens-servico/{osId}/finalizar` | `os:escrever` | [finalizar-servico.md](finalizar-servico.md) |
| 8 | Registrar Entrega de Veículo | `POST /ordens-servico/{osId}/entrega` | `os:escrever` | [registrar-entrega-de-veiculo.md](registrar-entrega-de-veiculo.md) |
| 9 | Consultar Ordem de Serviço | `GET /ordens-servico/{osId}` | `os:ler` | [consultar-ordem-de-servico.md](consultar-ordem-de-servico.md) |
| 10 | Listar Ordens de Serviço | `GET /ordens-servico` | `os:ler` | [listar-ordens-de-servico.md](listar-ordens-de-servico.md) |
| 6 | Monitorar Tempo Médio de Execução | `GET /ordens-servico/{osId}/tempo-execucao` e `GET /ordens-servico/tempos-execucao` | `os:ler` | [monitorar-tempo-medio-de-execucao.md](monitorar-tempo-medio-de-execucao.md) |

## Tipos do contexto

**Status da Ordem de Serviço**

| Status | Quando ocorre |
|---|---|
| `RECEBIDA` | OS criada, veículo recebido. |
| `EM_DIAGNOSTICO` | Após o registro do problema relatado pelo cliente. |
| `AGUARDANDO_APROVACAO` | Orçamento calculado e enviado ao cliente. |
| `AGUARDANDO_RECURSOS` | Itens comprados e ainda não recebidos. |
| `AGUARDANDO_EXECUCAO` | Orçamento aprovado e recursos disponíveis; OS na fila. |
| `EM_EXECUCAO` | Mecânico iniciou a execução. |
| `FINALIZADA` | Serviços concluídos, veículo aguardando retirada. |
| `ENTREGUE` | Pagamento confirmado e veículo entregue. |
| `CANCELADA` | Orçamento recusado. |

> `AGUARDANDO_RECURSOS` e `AGUARDANDO_EXECUCAO` não constam no enunciado do Tech Challenge — ver
> ponto 2 de [`pontos-em-aberto.md`](pontos-em-aberto.md).

**Ordem de Serviço**

| Campo | Tipo | Observação |
|---|---|---|
| `ordemServicoId` | uuid | Identificador da OS. |
| `clienteId` / `veiculoId` | uuid | Vínculos obrigatórios na criação. |
| `status` | enum | Conforme a tabela acima. |
| `problemaRelatado` | objeto | Descrição e observações informadas pelo cliente. |
| `dataInicioDiagnostico` | datetime | Gravada no registro do problema relatado. |
| `dataEntradaFila` | datetime | Define, com o status, a participação na fila. |
| `mecanicoResponsavelId` | uuid | Preservado quando a OS volta à fila. |
| `dataInicioExecucao` / `dataFinalizacao` / `dataEntrega` | datetime | Marcos do atendimento. |
| `observacoesFinalizacao` / `observacoesEntrega` | string | Opcionais. |

**Problema**

`problemaId`, `descricao`, `observacoes`, `identificadoEm`. O problema **não** tem tipo próprio: o
tipo pertence ao orçamento ao qual ele é vinculado.

**Evento da OS (`event_data`)**

`aggregateType`, `aggregateId`, `ordemServicoId`, `eventType`, `statusAnterior`, `statusNovo`,
`etapa`, `payload`, `metadata`, `occurredAt`, `createdAt`. É o histórico técnico e de negócio; não
substitui o status atual da OS.

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; path param `{osId}`; ações de transição como sub-recurso com verbo
  de negócio (`/execucao/iniciar`, `/finalizar`, `/entrega`).
- Autenticação `Bearer <JWT>`; escopos `os:ler` e `os:escrever`; perfis `MECANICO` e `GESTOR`, com
  o indicador de tempo médio restrito ao `GESTOR`.
- A **regra de transição pertence ao domínio**: os casos de uso chamam métodos da OS
  (`iniciarExecucao()`, `finalizar()`, `entregar()`), e não mudam status no handler.
- A **fila não é persistida**: uma OS pertence à fila quando tem `status = AGUARDANDO_EXECUCAO` e
  `dataEntradaFila` preenchida.
- Envelope de listagem paginada: `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409` e `201` na criação.

## O que este contexto não faz

- Não calcula nem aprova orçamento: consulta o contexto de Orçamento.
- Não movimenta estoque: consulta e é notificado pelo contexto de Peças & Insumos.
- Não trata pagamento, embora a entrega dependa da confirmação dele.
