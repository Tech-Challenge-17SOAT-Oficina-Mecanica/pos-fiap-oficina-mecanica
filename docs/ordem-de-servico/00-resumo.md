---
documento: Resumo do Contexto — Ordem de Serviço
dono: A definir
versao: 0.3
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
| 1 | Criar Ordem de Serviço | `POST /ordens-servico` | `os:escrever` | [criar-ordem-de-servico.md](criar-ordem-de-servico.md) |
| 2 | Registrar Problema Relatado | `POST /ordens-servico/{osId}/problema-relatado` | `os:escrever` | [registrar-problema-relatado.md](registrar-problema-relatado.md) |
| 3 | Registrar Problema Encontrado | `POST /ordens-servico/{osId}/problemas` | `os:escrever` | [registrar-problema-encontrado.md](registrar-problema-encontrado.md) |
| 4 | Registrar Serviços Necessários | `POST /ordens-servico/{osId}/servicos` | `os:escrever` | [registrar-servicos-necessarios.md](registrar-servicos-necessarios.md) |
| 5 | Registrar Peças e Insumos Necessários | `POST /ordens-servico/{osId}/pecas`, `POST /ordens-servico/{osId}/insumos` e `GET /ordens-servico/{osId}/orcamento` | `os:escrever` e `os:ler` | [registrar-pecas-e-insumos-necessarios.md](registrar-pecas-e-insumos-necessarios.md) |
| 6 | Incluir OS na Fila de Atendimento | sem endpoint, caso de uso interno | — | [incluir-os-na-fila-de-atendimento.md](incluir-os-na-fila-de-atendimento.md) |
| 7 | Consultar Fila de Atendimento | `GET /fila-atendimento` | `os:ler` | [consultar-fila-de-atendimento.md](consultar-fila-de-atendimento.md) |
| 8 | Iniciar Execução | `POST /ordens-servico/{osId}/execucao/iniciar` | `os:escrever` | [iniciar-execucao.md](iniciar-execucao.md) |
| 9 | Finalizar Serviço | `POST /ordens-servico/{osId}/finalizar` | `os:escrever` | [finalizar-servico.md](finalizar-servico.md) |
| 10 | Registrar Entrega de Veículo | `POST /ordens-servico/{osId}/entrega` | `os:escrever` | [registrar-entrega-de-veiculo.md](registrar-entrega-de-veiculo.md) |
| 11 | Consultar Ordem de Serviço | `GET /ordens-servico/{osId}` | `os:ler` | [consultar-ordem-de-servico.md](consultar-ordem-de-servico.md) |
| 12 | Listar Ordens de Serviço | `GET /ordens-servico` | `os:ler` | [listar-ordens-de-servico.md](listar-ordens-de-servico.md) |
| 13 | Monitorar Tempo Médio de Execução | `GET /ordens-servico/{osId}/tempo-execucao` e `GET /ordens-servico/tempos-execucao` | `os:ler` | [monitorar-tempo-medio-de-execucao.md](monitorar-tempo-medio-de-execucao.md) |

As tarefas estão numeradas **na ordem do fluxo de atendimento**, e os IDs de requisito seguem a
mesma ordem: `RF-OS-01` a `RF-OS-124` e `RNF-OS-01` a `RNF-OS-65`, sem repetição.

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
| `ENTREGUE` | Veículo entregue e valor final registrado. |
| `CANCELADA` | Orçamento recusado. |

**Máquina de estados**

```
RECEBIDA
  └─ registrar problema relatado ─────────────→ EM_DIAGNOSTICO
       └─ calcular e enviar orçamento ────────→ AGUARDANDO_APROVACAO
            ├─ cliente recusa o principal ────→ CANCELADA
            └─ cliente aprova ───────────────→ processamento dos itens
                 ├─ algum item comprado ─────→ AGUARDANDO_RECURSOS
                 │     └─ entrada de estoque, sem pendência → AGUARDANDO_EXECUCAO
                 └─ tudo reservado ──────────→ AGUARDANDO_EXECUCAO
                       └─ iniciar execução ──→ EM_EXECUCAO
                            ├─ problema encontrado → orçamento COMPLEMENTAR em CRIADO,
                            │    a OS volta a AGUARDANDO_APROVACAO até a decisão do cliente
                            └─ finalizar ────→ FINALIZADA
                                 └─ entrega ─→ ENTREGUE
```

Quem manda a transição:

- **RECEBIDA → EM_DIAGNOSTICO**: registrar problema relatado.
- **EM_DIAGNOSTICO → AGUARDANDO_APROVACAO**: envio do orçamento ao cliente.
- **AGUARDANDO_APROVACAO → AGUARDANDO_RECURSOS ou AGUARDANDO_EXECUCAO**: a aprovação do orçamento
  chama o processamento de itens, e o resultado define o estado — pendente de compra ou pronto.
- **AGUARDANDO_RECURSOS → AGUARDANDO_EXECUCAO**: entrada de estoque, quando não resta item pendente.
- **AGUARDANDO_EXECUCAO → EM_EXECUCAO**: início da execução.
- **EM_EXECUCAO → FINALIZADA**: finalização, bloqueada enquanto houver reserva sem baixa.
- **FINALIZADA → ENTREGUE**: entrega, que não depende de pagamento.
- **qualquer estado → CANCELADA**: recusa do orçamento principal, com devolução dos itens ao
  estoque. Recusa de complementar **não** cancela a OS.

> `AGUARDANDO_RECURSOS` e `AGUARDANDO_EXECUCAO` não constam no enunciado do Tech Challenge: são
> estados que o fluxo de estoque exigiu. `AGUARDANDO_RECURSOS` existe porque a OS pode ficar parada
> esperando compra chegar, e `AGUARDANDO_EXECUCAO` distingue "pronta para começar" de "em
> andamento" — sem ele, a fila de atendimento não teria como ser montada.

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
- Autenticação `Bearer <JWT>`; escopos `os:ler` e `os:escrever`; perfil `MECANICO`, com o
  indicador de tempo médio restrito a quem tiver o escopo `os:ler` de gestão. O `CLIENTE`
  acompanha apenas a própria OS.
- A **regra de transição pertence ao domínio**: os casos de uso chamam métodos da OS
  (`iniciarExecucao()`, `finalizar()`, `entregar()`), e não mudam status no handler.
- A **fila não é persistida**: uma OS pertence à fila quando tem `status = AGUARDANDO_EXECUCAO` e
  `dataEntradaFila` preenchida.
- Envelope de listagem paginada: `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409` e `201` na criação.

## O que este contexto não faz

- Não calcula nem aprova orçamento: consulta o contexto de Orçamento.
- Não movimenta estoque: consulta e é chamado pelos contextos de Peças e de Insumos.
- Não trata pagamento, embora a entrega dependa da confirmação dele.
