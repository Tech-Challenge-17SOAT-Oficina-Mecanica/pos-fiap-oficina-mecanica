---
documento: Refinamento de Requisitos — Registrar Consumo e Saída
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Registrar Consumo e Saída

Este documento detalha a tarefa Registrar Consumo e Saída do contexto de Peças & Insumos.

## 6 · Registrar Consumo e Saída

### 6.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Dar baixa nas peças e insumos efetivamente utilizados na execução do serviço, reduzindo o
saldo físico do estoque.

**Problema**
A reserva apenas separa a peça; ela continua contando como saldo físico. Se a baixa não for
registrada quando a peça é montada no veículo, o sistema segue afirmando que ela existe na
prateleira — e o estoque acumula divergência até o inventário.

**Pré-condições**

- A OS deve estar em execução.
- As peças devem estar reservadas para a OS.
- O usuário deve estar autorizado a movimentar estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-31 | Permitir registrar a saída de peças reservadas para a OS. |
| RF-EST-32 | Permitir registrar o consumo de insumos vinculado à OS. |
| RF-EST-33 | Reduzir o saldo físico e o saldo reservado da peça na mesma operação. |
| RF-EST-34 | Reduzir o saldo físico do insumo consumido. |
| RF-EST-35 | Permitir registrar quantidade consumida menor que a reservada e devolver a diferença ao saldo disponível. |
| RF-EST-36 | Registrar a movimentação de saída no histórico, vinculada à OS. |
| RF-EST-37 | Impedir saída de peça não reservada para a OS. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-27 | A operação deve ser feita por API RESTful. |
| RNF-EST-28 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-EST-29 | A operação deve ser transacional — a baixa do saldo físico e a baixa da reserva ocorrem juntas ou não ocorrem. |
| RNF-EST-30 | O saldo físico não pode ficar negativo em nenhuma hipótese. |
| RNF-EST-31 | A movimentação deve ser auditável e o histórico imutável. |
| RNF-EST-32 | A operação deve ser idempotente por item da OS, para impedir baixa em duplicidade. |

**Fluxo Principal**

1. O mecânico informa a OS e os itens efetivamente utilizados.
2. O sistema valida que a OS está em execução.
3. O sistema valida que as peças informadas estão reservadas para essa OS.
4. O sistema reduz o saldo físico e o saldo reservado das peças.
5. O sistema reduz o saldo físico dos insumos consumidos.
6. O sistema registra a movimentação de saída no histórico, vinculada à OS.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Peça não reservada para a OS | Impede a saída e orienta a registrar a peça na OS e reservá-la antes. |
| A2 | Consumo menor que o reservado | Baixa a quantidade utilizada e libera a diferença de volta ao saldo disponível. |
| A3 | Consumo maior que o reservado | Verifica o saldo disponível e, havendo saldo, registra a diferença como consumo adicional; não havendo, impede a operação e sinaliza a indisponibilidade. |
| A4 | Insumo com saldo físico insuficiente | Impede a baixa e sinaliza a necessidade de reposição. |
| A5 | Baixa já registrada para o item | Informa que a saída já foi lançada e não altera o saldo. |
| A6 | OS fora do status de execução | Impede a operação. |
| A7 | Usuário sem autorização | Impede a operação. |

**Saída**

- Confirmação da saída com os saldos atualizados dos itens movimentados; **ou**
- Indicação do motivo pelo qual a baixa foi recusada.

**Pós-condições**

- O saldo físico das peças e insumos utilizados está reduzido.
- O saldo reservado das peças baixadas está zerado para essa OS.
- Eventual diferença entre reservado e consumido voltou ao saldo disponível.
- A movimentação está registrada no histórico e vinculada à OS, permitindo apurar o custo real do serviço.

---

### 6.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/saidas
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:movimentar`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | Recomendado; impede baixa em duplicidade |
| Body | `ordemServicoId` | uuid   | Obrigatório; OS em execução |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido |
| Body | `itens[].itemId` | uuid   | Peça reservada para a OS ou insumo consumido |
| Body | `itens[].quantidade` | decimal | Maior ou igual a zero, com casas decimais compatíveis com a `unidadeMedida` |

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "itens": [
    { "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", "quantidade": 4 },
    { "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90", "quantidade": 0 },
    { "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48", "quantidade": 3.5 }
  ]
}
```

`quantidade: 0` em peça reservada significa peça não utilizada: a reserva é devolvida ao saldo disponível.

**Validações**

*Técnicas*

- `ordemServicoId` obrigatório.
- `itens` não vazio, sem `itemId` repetido.
- `quantidade` maior ou igual a zero, com casas decimais compatíveis com a `unidadeMedida`.

*Negócio*

- A OS está com status `EM_EXECUCAO`.
- Toda peça informada possui reserva `ATIVA` para essa OS.
- Consumo de peça maior que o reservado exige saldo disponível para a diferença.
- Insumo não exige reserva, mas exige `saldoFisico` suficiente.
- `saldoFisico` não pode ficar negativo em nenhuma hipótese.

**Processamento**

1. Verificar o `Idempotency-Key`.
2. Consultar o módulo de OS: o status é `EM_EXECUCAO`?
3. Abrir transação.
4. Carregar os itens com `SELECT ... FOR UPDATE`, ordenados por `item_id`.
5. Para cada peça:
   - Carregar a reserva ativa da OS.
   - Consumido menor ou igual ao reservado: `saldo_fisico -= consumido`, `saldo_reservado -= reservado`, devolvendo a diferença ao disponível.
   - Consumido maior que o reservado: conferir o disponível para a diferença; havendo, baixar; não havendo, abortar com `409`.
   - Marcar a reserva como `CONSUMIDA`.
6. Para cada insumo: conferir `saldo_fisico >= consumido` e baixar.
7. Inserir `movimentacao_estoque` do tipo `SAIDA` por linha, com o `os_id`.
8. Commit.
9. Publicar os eventos `PecaBaixada` e `InsumoConsumido`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `chave_idempotencia`, módulo de OS
- Altera: `item_estoque.saldo_fisico` e `item_estoque.saldo_reservado`, `reserva_estoque.status`, `movimentacao_estoque` (insert)
- A baixa do físico e a baixa da reserva ocorrem na mesma transação — é o ponto onde o estoque mais diverge se houver falha parcial.

**Saída da API**

```json
{
  "saidaId": "41d9c3a8-7e56-4b02-8f37-9a2065bd1ce4",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "registradoEm": "2026-08-12T16:40:00-03:00",
  "itens": [
    { "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", "codigo": "PC-0142", "consumido": 4, "reservado": 4, "devolvido": 0, "saldoFisicoAtual": 12 },
    { "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90", "codigo": "PC-0311", "consumido": 0, "reservado": 1, "devolvido": 1, "saldoFisicoAtual": 3 },
    { "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48", "codigo": "IN-0031", "consumido": 3.5, "reservado": 0, "devolvido": 0, "saldoFisicoAtual": 44.0 }
  ],
  "custoTotalMateriais": 512.10
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Saída registrada |
| `200` | Repetição de `Idempotency-Key` — retorna a resposta original |
| `400` | Body inválido; quantidade negativa; decimal incompatível com a unidade |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `estoque:movimentar` |
| `404` | OS ou item não encontrado |
| `409` | Peça sem reserva para a OS; saldo insuficiente para consumo acima do reservado |
| `422` | OS fora do status `EM_EXECUCAO` |

**Dependências**

- `ItemEstoqueRepository`
- `ReservaEstoqueRepository`
- `MovimentacaoEstoqueRepository`
- Módulo Ordem de Serviço — status e itens
- Serviço de idempotência
- Caso de uso Reservar Peça para OS (pré-requisito)

**Testes**

*Unitários*

- Consumo igual ao reservado: reserva zerada, nada devolvido.
- Consumo menor que o reservado: a diferença volta ao disponível.
- Consumo maior que o reservado com saldo: baixa a diferença.
- Consumo maior que o reservado sem saldo: aborta.
- Consumo zero: devolve toda a reserva.
- Insumo com decimal na unidade correta.

*Integração*

- Saída completa reduz físico e reservado juntos.
- Peça sem reserva retorna `409` e nada é alterado.
- OS em `AGUARDANDO_APROVACAO` retorna `422`.
- Repetição de `Idempotency-Key` não baixa duas vezes.
- `saldo_fisico` nunca fica negativo, mesmo com payload malicioso.

*Regressão*

- Após a saída, `saldo_fisico - saldo_reservado` continua igual ao disponível esperado.

---

### 6.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `baixar()` na entidade `ItemEstoque` reduzindo saldo físico e reservado juntos
- [ ] Implementar a invariante de saldo físico nunca negativo
- [ ] Implementar a regra de consumo menor que o reservado devolvendo a diferença ao disponível
- [ ] Implementar a regra de consumo maior que o reservado consumindo do disponível quando houver
- [ ] Implementar a transição da `ReservaEstoque` para `CONSUMIDA`

**Caso de uso**

- [ ] Implementar `RegistrarSaidaEstoque` cobrindo peças e insumos
- [ ] Implementar o cálculo de `custoTotalMateriais` da OS

**Repositório**

- [ ] Registrar `MovimentacaoEstoque` do tipo `SAIDA` vinculada ao `osId`

**Integrações**

- [ ] Consultar o módulo de OS para validar o status `EM_EXECUCAO`

**Handler HTTP**

- [ ] Implementar `POST /estoque/saidas`

**Validações**

- [ ] Validar `quantidade` maior ou igual a zero
- [ ] Validar decimais compatíveis com a unidade de medida do item
- [ ] Validar que toda peça informada possui reserva ativa para a OS, retornando `409` caso contrário

**Transação e idempotência**

- [ ] Executar a baixa do saldo físico e da reserva na mesma transação
- [ ] Implementar `SELECT ... FOR UPDATE` ordenado por `item_id`
- [ ] Implementar a idempotência por item da OS

**Eventos**

- [ ] Publicar `PecaBaixada`
- [ ] Publicar `InsumoConsumido`

**Testes unitários**

- [ ] Consumo igual ao reservado: reserva zerada e nada devolvido
- [ ] Consumo menor que o reservado: diferença volta ao disponível
- [ ] Consumo maior que o reservado com saldo suficiente
- [ ] Consumo maior que o reservado sem saldo: aborta
- [ ] Consumo zero devolvendo toda a reserva

**Testes de integração**

- [ ] Saída reduzindo físico e reservado juntos
- [ ] Peça sem reserva retornando `409` sem alterar nada
- [ ] OS fora do status `EM_EXECUCAO` retornando `422`
- [ ] `Idempotency-Key` repetida não baixando duas vezes
- [ ] Saldo físico nunca ficando negativo, mesmo com payload malicioso
- [ ] Regressão: saldo físico menos saldo reservado continua igual ao disponível esperado

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, explicando o comportamento de `quantidade: 0`

**Review**

- [ ] Code Review aprovado

---
