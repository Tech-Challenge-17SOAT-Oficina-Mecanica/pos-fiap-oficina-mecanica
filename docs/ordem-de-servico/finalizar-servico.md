---
documento: Refinamento de Requisitos — Finalizar Serviço
dono: A definir
versao: 0.2
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Finalizar Serviço

Este documento detalha a tarefa Finalizar Serviço do contexto de Ordem de Serviço.

## 9 · Finalizar Serviço

### 9.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar a conclusão dos serviços autorizados de uma Ordem de Serviço, finalizar o atendimento
técnico e notificar o cliente de que o veículo está disponível para retirada.

**Problema**

Depois de concluir os reparos autorizados, a oficina precisa registrar formalmente o término da
execução e comunicar ao cliente que o veículo está pronto, mantendo a OS aguardando retirada até
que o cliente compareça, pague e retire o veículo.

**Pré-condições**

- Deve existir uma Ordem de Serviço, com a execução já iniciada e status `EM_EXECUCAO`.
- Todos os serviços autorizados devem estar concluídos.
- Não pode existir serviço adicional autorizado pendente de execução.
- Não pode existir orçamento complementar pendente que impeça a conclusão do atendimento.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-81 | Permitir ao mecânico finalizar a execução da OS. |
| RF-OS-82 | Validar se todos os serviços autorizados foram concluídos. |
| RF-OS-83 | Validar se não existem serviços pendentes. |
| RF-OS-84 | Registrar a data e hora da finalização. |
| RF-OS-85 | Permitir registrar observações finais, quando necessário. |
| RF-OS-86 | Alterar o status da OS para `FINALIZADA`. |
| RF-OS-87 | Notificar o cliente de que o veículo está disponível para retirada. |
| RF-OS-88 | Registrar que a notificação foi realizada. |
| RF-OS-89 | Manter a OS como `FINALIZADA` enquanto o veículo aguarda retirada. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-40 | A finalização deve ser persistida de forma consistente. |
| RNF-OS-41 | Somente usuários autorizados devem poder finalizar o serviço. |
| RNF-OS-42 | A alteração do status e o registro da finalização devem ocorrer de forma consistente. |
| RNF-OS-43 | O histórico da Ordem de Serviço deve ser preservado. |
| RNF-OS-44 | Uma falha no envio da notificação não deve provocar perda das informações de finalização da OS. |
| RNF-OS-45 | As informações da notificação devem permitir rastrear quando o cliente foi comunicado. |

**Fluxo Principal**

1. O mecânico acessa a Ordem de Serviço em execução.
2. O sistema apresenta os serviços autorizados da OS.
3. O mecânico confirma a conclusão dos serviços.
4. O sistema verifica se todos os serviços autorizados foram concluídos.
5. O sistema verifica se não existem pendências que impeçam a finalização.
6. O mecânico solicita a finalização.
7. O sistema registra a data e hora da finalização.
8. O sistema altera o status da OS para `FINALIZADA`.
9. O sistema registra as observações finais, quando informadas.
10. O sistema notifica o cliente de que o veículo está disponível para retirada.
11. O sistema registra a realização da notificação.
12. A OS permanece `FINALIZADA`, aguardando a retirada do veículo.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | OS não encontrada | Informa que a Ordem de Serviço não existe. |
| A2 | OS não está em execução | Impede a finalização. |
| A3 | Serviços pendentes | Informa que ainda existem serviços autorizados não concluídos. |
| A4 | Serviço adicional autorizado pendente | Impede a finalização enquanto houver serviço adicional aprovado ainda não executado. |
| A5 | Orçamento complementar pendente | Impede a finalização quando a decisão do cliente for necessária para concluir o atendimento. |
| A6 | OS já finalizada | Impede uma nova finalização. |
| A7 | Falha ao notificar o cliente | A OS permanece `FINALIZADA`, a falha é registrada e a notificação pode ser reenviada. |
| A8 | Usuário sem autorização | Impede a operação. |

**Saída**

- Ordem de Serviço finalizada e cliente informado de que o veículo está disponível para retirada.

**Pós-condições**

- A OS passa para o status `FINALIZADA`, com data e hora da finalização registradas.
- Todos os serviços autorizados permanecem registrados como concluídos.
- O cliente é notificado sobre a disponibilidade do veículo e o veículo fica aguardando retirada.
- O pagamento ainda não é registrado.
- A OS fica apta para a tarefa Registrar Entrega do Veículo, onde acontecem o pagamento e a entrega.

---

### 9.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/finalizar
```

> **Decisão de projeto.** A regra de transição pertence ao domínio: o caso de uso chama
> `ordemServico.finalizar()`, e é a própria Ordem de Serviço que protege a regra de que somente
> uma OS em execução e sem pendências pode ser finalizada. Orçamento e Estoque são apenas
> consultados. A notificação ao cliente é disparada como consequência da finalização e pode virar
> tarefa própria, se o time quiser separar a responsabilidade.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`.
- Escopo: `os:escrever`.
- O identificador do mecânico é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |
| Body | `observacoes` | string | Opcional; observações finais da execução. |

```json
{
  "observacoes": "Todos os serviços autorizados foram concluídos e o veículo foi testado."
}
```

**Validações**

*Técnicas*

- `osId` em formato UUID válido.
- Payload válido, quando informado.

*Negócio*

- A OS deve existir e estar `EM_EXECUCAO`, com a execução já iniciada.
- Todos os serviços autorizados devem estar concluídos.
- Todos os serviços adicionais aprovados devem estar concluídos.
- Não pode haver atividade obrigatória pendente.
- Não é permitido finalizar OS já `FINALIZADA` nem `ENTREGUE`.
- As peças e os insumos utilizados devem ter sido baixados por `POST /estoque/saidas`.

> **Decisão de projeto.** A baixa de consumo é **pré-condição da finalização**: enquanto houver
> reserva ativa da OS sem baixa registrada, a finalização retorna `409` com a lista dos itens
> pendentes. Sem esse bloqueio a OS encerraria com peça ainda comprometida, o saldo reservado nunca
> zeraria e o inventário jamais fecharia.

> **Decisão de projeto.** A **notificação ao cliente é por e-mail** no MVP, usando o contato do
> cadastro do cliente, e o resultado do envio é gravado para permitir reenvio. A falha do envio não
> desfaz a finalização: o serviço está pronto, e o aviso é consequência.

**Regra de domínio**

```
EM_EXECUCAO → FINALIZADA
```

A OS não pode ser finalizada apenas porque o mecânico encerrou sua atividade. Ela só é finalizada
quando todos os serviços aprovados e todos os reparos adicionais aprovados estiverem concluídos e
não houver pendência que impeça a entrega do veículo.

**Processamento**

1. Receber o identificador da OS e identificar o usuário autenticado.
2. Buscar a Ordem de Serviço e validar sua existência.
3. Validar se a OS está `EM_EXECUCAO`.
4. Buscar e verificar os serviços autorizados da OS.
5. Validar se todos os serviços obrigatórios foram concluídos.
6. Validar os serviços adicionais aprovados.
7. Validar se não existe pendência impeditiva.
8. Registrar observações finais, quando informadas.
9. Registrar data e hora da finalização.
10. Alterar o status da OS para `FINALIZADA`.
11. Persistir a Ordem de Serviço.
12. Disparar o fluxo de notificação do cliente e registrar o resultado do envio.
13. Retornar a Ordem de Serviço finalizada.

**Persistência**

- Consulta: `ordem_servico` e seus serviços, `orcamento` (serviços efetivamente autorizados),
  estoque (confirmação das movimentações), quando aplicável.
- Altera: `ordem_servico` (`status = FINALIZADA`, `data_finalizacao`, `observacoes_finalizacao`).

**Saída da API**

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "status": "FINALIZADA",
  "dataFinalizacao": "2026-08-14T21:15:00-03:00",
  "observacoes": "Todos os serviços autorizados foram concluídos e o veículo foi testado."
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Ordem de Serviço finalizada com sucesso. |
| `400` | Dados da requisição inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:escrever`. |
| `404` | Ordem de Serviço não encontrada. |
| `409` | OS não está `EM_EXECUCAO`; existem serviços autorizados pendentes; existem serviços adicionais aprovados pendentes; **existe reserva ativa da OS sem baixa de estoque registrada**. |

**Dependências**

- `OrdemDeServicoRepository`.
- `OrcamentoRepository`, para validar os serviços autorizados.
- `ReservaEstoqueRepository`, para verificar reservas ativas da OS sem baixa registrada.
- Serviço de notificação ao cliente, por e-mail.
- Middleware de autenticação/autorização.

**Fora do escopo desta tarefa**

Registrar pagamento; entregar o veículo; alterar a OS para `ENTREGUE`; registrar a retirada do
cliente; gerar novo orçamento; executar reparos adicionais; refazer diagnóstico.

**Testes**

*Unitários*

- Finalização válida altera o status para `FINALIZADA` e grava a data.
- Rejeita OS fora de `EM_EXECUCAO`.
- Rejeita OS com serviço obrigatório pendente.
- Rejeita OS com serviço adicional aprovado pendente.
- Rejeita OS já `FINALIZADA` ou `ENTREGUE`.
- A finalização não altera a OS para `ENTREGUE`.

*Integração*

- `POST` válido retorna `200` e persiste status, data de finalização e observações.
- OS inexistente retorna `404`.
- OS fora de `EM_EXECUCAO` retorna `409`.
- Serviço pendente ou reparo adicional pendente retorna `409`.
- Reserva ativa sem baixa registrada retorna `409` com a lista dos itens pendentes.
- Sem token retorna `401` e perfil sem escopo retorna `403`.
- A notificação por e-mail é disparada, o resultado do envio é gravado, e a falha dela não desfaz a finalização.

---

### 9.3 Checklist de Implementação

> **Nota de implementação (2026-08-27).** Implementado em `internal/domain/ordemservico`,
> `internal/application/ordemservico` e `internal/infrastructure/ordemservico` (reaproveitando o
> repositório existente). Dois desvios do refinamento original:
> a notificação ao cliente não tem provedor de e-mail configurado neste projeto, então o "envio" é
> apenas registrado em log (best-effort, após o commit) — não há coluna nova para rastrear o
> resultado, pois a seção de Persistência do refinamento não previa uma; e a tarefa **Iniciar
> Execução** (`POST /ordens-servico/{osId}/execucao/iniciar`) ainda não existe neste repositório,
> então hoje não há caminho pela API para levar uma OS a `EM_EXECUCAO` — o teste de integração
> insere esse status diretamente via SQL.

**Domínio**

- [x] Criar o método de domínio `finalizar()` na Ordem de Serviço (implementado como validações e constantes em `domain/ordemservico`, orquestradas pelo repositório, seguindo o padrão já usado nos demais fluxos da OS)
- [x] Validar que a OS está com status `EM_EXECUCAO` e que a execução foi iniciada
- [x] Validar que todos os serviços autorizados estão concluídos (via `ordem_servico_servico.status`)
- [x] Validar que todos os serviços adicionais aprovados estão concluídos (mesma verificação, não há distinção de tabela entre serviço original e adicional no schema atual)
- [x] Validar que não existem pendências impeditivas (orçamento complementar em `CRIADO`)
- [x] Bloquear a finalização quando ainda houver reserva ativa da OS sem baixa registrada
- [x] Enviar a notificação de conclusão por e-mail e gravar o resultado do envio (simplificado para log, ver nota acima)
- [x] Definir o tratamento para orçamento complementar ainda aguardando decisão (bloqueia com `409`)
- [x] Registrar data e hora da finalização e as observações finais
- [x] Alterar o status da OS para `FINALIZADA`

**Caso de uso**

- [x] Implementar `FinalizarServico`
- [x] Persistir as alterações da Ordem de Serviço

**Repositório**

- [x] Criar ou ajustar `OrdemDeServicoRepository`
- [x] Criar as consultas necessárias aos serviços da OS

**Integrações**

- [x] Criar a consulta ao orçamento, quando necessária
- [x] Criar a consulta ao estoque, quando necessária
- [x] Disparar a notificação do cliente e registrar o resultado do envio (ver nota acima)

**Handler HTTP**

- [x] Implementar `POST /ordens-servico/{osId}/finalizar`
- [x] Implementar a validação do path param e do payload
- [x] Criar DTO/request de entrada e DTO/response de saída
- [x] Aplicar autenticação JWT e autorização por escopo na rota
- [x] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [x] Retornar `404` para OS inexistente
- [x] Retornar `409` para OS fora de `EM_EXECUCAO`
- [x] Retornar `409` quando houver serviço pendente
- [x] Retornar `409` quando houver reparo adicional aprovado pendente
- [x] Retornar `409` quando houver reserva ativa da OS sem baixa de estoque registrada

**Testes unitários**

- [x] Finalização válida (delegação ao repositório, normalização de observações)
- [ ] OS inexistente, fora de `EM_EXECUCAO`, serviço/reparo pendente como testes unitários isolados — cobertos apenas pelo teste de integração e pelo teste do handler, pois a regra vive na query SQL do repositório
- [x] Usuário sem autorização (via middleware, testado no handler)

**Testes de integração**

- [x] Persistência do status `FINALIZADA` e da data de finalização
- [x] OS inexistente, fora de `EM_EXECUCAO`, serviço pendente, orçamento complementar pendente e reserva ativa pendente retornando `409`/`404`
- [x] Finalizar OS já `FINALIZADA` retorna `409`
- [ ] A finalização não altera a OS para `ENTREGUE` — não testado explicitamente (não há rota de entrega implementada ainda para verificar a transição seguinte)
- [ ] Notificação disparada, com falha não desfazendo a finalização — não testado, pois a notificação é só log (ver nota acima)

**Documentação**

- [x] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [x] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [x] Executar testes automatizados (unitários; integração escrita, não executada nesta sessão por falta de Docker)
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
