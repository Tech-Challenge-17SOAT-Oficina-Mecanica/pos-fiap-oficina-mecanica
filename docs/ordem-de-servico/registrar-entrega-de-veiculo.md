---
documento: Refinamento de Requisitos — Registrar Entrega de Veículo
dono: A definir
versao: 0.2
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Registrar Entrega de Veículo

Este documento detalha a tarefa Registrar Entrega de Veículo do contexto de Ordem de Serviço.

## 10 · Registrar Entrega de Veículo

### 10.1 Refinamento de Produto

**Persona**

Mecânico, ou responsável pela entrega.

**Objetivo**

Registrar a retirada do veículo pelo cliente, apresentar o valor final da Ordem de Serviço e concluir
formalmente o atendimento.

**Problema**

Depois que o serviço é finalizado, a oficina precisa garantir que a retirada do veículo fique
registrada — quem levou, quando, e por qual valor final —, e que a Ordem de Serviço seja encerrada
corretamente com o status `ENTREGUE`.

**Pré-condições**

- Deve existir uma Ordem de Serviço com status `FINALIZADA` e serviços concluídos.
- O cliente deve ter sido notificado de que o veículo está disponível para retirada.
- O cliente deve comparecer à oficina para retirar o veículo.
- O valor final da OS deve estar disponível para cobrança.
- O usuário deve estar autenticado e autorizado.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-90 | Permitir consultar a Ordem de Serviço finalizada. |
| RF-OS-91 | Apresentar o valor final devido pelo cliente. |
| RF-OS-92 | Apresentar o valor final da OS na entrega. |
| RF-OS-93 | Registrar o valor final acordado junto com a entrega. |
| RF-OS-94 | Não bloquear a entrega por pagamento: o recebimento é controlado fora do sistema no MVP. |
| RF-OS-95 | Registrar a data e hora da entrega. |
| RF-OS-96 | Associar a entrega ao responsável e registrar quem realizou a retirada, quando aplicável. |
| RF-OS-97 | Registrar observações da entrega, quando informadas. |
| RF-OS-98 | Alterar o status da OS para `ENTREGUE`. |
| RF-OS-99 | Impedir novo registro de entrega para a mesma OS. |
| RF-OS-100 | Registrar a conclusão do atendimento. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-46 | A operação deve ser persistida de forma consistente. |
| RNF-OS-47 | Somente usuários autorizados devem poder registrar a entrega do veículo. |
| RNF-OS-48 | O valor final e a entrega devem manter rastreabilidade. |
| RNF-OS-49 | A alteração da OS para `ENTREGUE` depende apenas de a OS estar `FINALIZADA`. |
| RNF-OS-50 | O histórico da OS deve ser preservado após a entrega. |
| RNF-OS-51 | Uma OS entregue não deve poder ser alterada livremente após o encerramento do atendimento. |
| RNF-OS-52 | A atualização da OS deve ser transacional. |

**Fluxo Principal**

1. O cliente comparece à oficina para retirar o veículo.
2. O responsável localiza a Ordem de Serviço.
3. O sistema verifica se a OS está com status `FINALIZADA`.
4. O sistema apresenta os dados da OS, os serviços executados e o valor final.
5. O responsável registra o valor final acordado com o cliente.
6. O responsável entrega o veículo ao cliente.
7. O sistema registra a data e hora da entrega, o responsável e as observações.
8. O sistema altera o status da OS para `ENTREGUE`.
9. O sistema registra a conclusão do atendimento.
10. O sistema confirma que a entrega foi registrada com sucesso.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | OS não encontrada | Informa que a Ordem de Serviço não existe. |
| A2 | OS ainda não finalizada | Impede a entrega, incluindo OS em `RECEBIDA`, `EM_DIAGNOSTICO`, `AGUARDANDO_APROVACAO` e `EM_EXECUCAO`. |
| A3 | Valor final não informado | Impede a entrega: o valor apresentado ao cliente precisa ficar registrado. |
| A5 | OS já entregue | Impede uma nova entrega. |
| A6 | Cliente diferente do esperado | Quando o responsável pela retirada é registrado, o sistema exige confirmação ou autorização. |
| A7 | Usuário sem autorização | Impede a operação. |
| A8 | Erro ao registrar | Mantém a OS no status anterior. |

**Saída**

- Valor final registrado, veículo entregue e Ordem de Serviço atualizada para o status `ENTREGUE`.

**Pós-condições**

- O valor final da OS fica registrado.
- A data e hora da entrega ficam registradas, com o responsável e as observações.
- O veículo é considerado devolvido ao cliente.
- A OS passa para o status `ENTREGUE` e o ciclo de atendimento é concluído.
- A OS permanece disponível para consulta de histórico, mas não para novas alterações operacionais.

---

### 10.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/entrega
```

> **Decisão de projeto.** Esta é a **última transição** do ciclo de vida da OS, e a regra pertence
> ao domínio: o caso de uso chama `ordemServico.entregar()`, e é a própria Ordem de Serviço que
> protege a passagem de `FINALIZADA` para `ENTREGUE`. O cliente é apenas consultado;
> o agregado alterado continua sendo a Ordem de Serviço.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `os:escrever`.
- O responsável pela entrega é identificado pelo usuário autenticado.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |
| Body | `clienteId` | uuid | Opcional; quem retirou o veículo, quando essa informação é registrada. |
| Body | `observacoes` | string | Opcional; observações da entrega. |

```json
{
  "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "observacoes": "Veículo entregue ao cliente sem ressalvas."
}
```

**Validações**

*Técnicas*

- `osId` em formato UUID válido.
- Payload válido, quando informado.

*Negócio*

- A OS deve existir e estar com status `FINALIZADA`.
- O veículo ainda não pode ter sido entregue.
- A entrega **não** depende de confirmação de pagamento.

> **Decisão de projeto.** **Pagamento fica fora do MVP.** A entrega registra o valor final e não
> bloqueia: não existe contexto, entidade nem rota de pagamento no projeto, e a oficina controla o
> recebimento fora do sistema. Sem essa decisão, a tarefa não seria implementável, porque exigia
> validar algo que não existe (D-25).
- O cliente ou representante responsável pela retirada deve ser válido, quando registrado.

**Regra de domínio**

```
FINALIZADA → ENTREGUE
```

Fluxo esperado: serviço finalizado, cliente notificado, cliente compareceu, valor final
registrado, veículo entregue, OS `ENTREGUE`. Antes disso, o veículo não
pode ser entregue.

**Processamento**

1. Receber o identificador da OS e identificar o usuário autenticado.
2. Buscar a Ordem de Serviço e validar sua existência.
3. Validar se a OS está `FINALIZADA` e se o veículo ainda não foi entregue.
4. Registrar o valor final acordado.
5. Registrar quem retirou o veículo, quando aplicável.
6. Registrar a data e hora da entrega e as observações informadas.
7. Alterar o status da OS para `ENTREGUE`.
8. Persistir a Ordem de Serviço em uma única transação.
9. Registrar a operação em log e retornar os dados da entrega.

**Persistência**

- Consulta: `ordem_servico`, `cliente` (quando a retirada é registrada).
- Altera: `ordem_servico` (`status = ENTREGUE`, `data_entrega`, `responsavel_entrega_id`,
  `cliente_retirada_id`, `observacoes_entrega`).

**Saída da API**

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "status": "ENTREGUE",
  "responsavelEntregaId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29",
  "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "dataEntrega": "2026-08-14T21:40:00-03:00",
  "observacoes": "Veículo entregue ao cliente sem ressalvas."
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Veículo entregue e OS encerrada com sucesso. |
| `400` | Dados da requisição inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:escrever`. |
| `404` | Ordem de Serviço não encontrada. |
| `409` | OS não está `FINALIZADA`; veículo já foi entregue. |

**Dependências**

- `OrdemDeServicoRepository`.
- `ClienteRepository`, quando a retirada for registrada.
- Middleware de autenticação/autorização.

**Fora do escopo desta tarefa**

Finalizar os serviços; registrar diagnóstico; gerar ou aprovar orçamento; executar reparo;
registrar problema adicional; alterar novamente uma OS já `ENTREGUE`.

**Testes**

*Unitários*

- Entrega válida altera o status para `ENTREGUE` e grava a data.
- Registra data, hora e responsável pela entrega.
- Rejeita OS fora de `FINALIZADA`.
- Rejeita OS já `ENTREGUE`.
- Registra o valor final junto com a entrega.

*Integração*

- `POST` válido retorna `200` e persiste status, data de entrega e responsável.
- OS inexistente retorna `404`.
- OS fora de `FINALIZADA` retorna `409`.
- OS já entregue retorna `409`.
- Sem token retorna `401` e perfil sem escopo retorna `403`.
- Status e data de entrega são atualizados na mesma transação.
- A operação encerra corretamente o ciclo da OS.

---

### 10.3 Checklist de Implementação

**Domínio**

- [ ] Criar o método de domínio `entregar()` na Ordem de Serviço
- [ ] Validar que a OS está com status `FINALIZADA`
- [ ] Impedir nova entrega de OS já `ENTREGUE`
- [ ] Registrar o valor final da OS na entrega, sem bloquear por pagamento
- [ ] Definir se será registrado quem retirou o veículo
- [ ] Registrar data e hora da entrega, responsável e observações
- [ ] Alterar o status da OS para `ENTREGUE`
- [ ] Criar ou ajustar os campos `dataEntrega`, `responsavelEntregaId`, `clienteRetiradaId` e `observacoesEntrega` na OS

**Caso de uso**

- [ ] Implementar `RegistrarEntregaVeiculo`
- [ ] Validar que o valor final foi informado
- [ ] Validar o cliente ou responsável pela retirada, quando aplicável
- [ ] Persistir as alterações da Ordem de Serviço

**Repositório**

- [ ] Criar ou ajustar `OrdemDeServicoRepository`
- [ ] Criar a consulta ao cliente, quando necessária

**Integrações**


**Handler HTTP**

- [ ] Implementar `POST /ordens-servico/{osId}/entrega`
- [ ] Implementar a validação do path param e do payload
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Transação**

- [ ] Garantir que status e data de entrega sejam atualizados na mesma transação

**Validações**

- [ ] Retornar `404` para OS inexistente
- [ ] Retornar `409` para OS fora de `FINALIZADA`
- [ ] Retornar `409` para veículo já entregue

**Testes unitários**

- [ ] Entrega válida
- [ ] OS inexistente
- [ ] OS fora de `FINALIZADA`
- [ ] OS já `ENTREGUE`
- [ ] Valor final ausente
- [ ] Usuário sem autorização

**Testes de integração**

- [ ] Persistência do status `ENTREGUE` e da data de entrega
- [ ] Encerramento correto do ciclo da OS

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
