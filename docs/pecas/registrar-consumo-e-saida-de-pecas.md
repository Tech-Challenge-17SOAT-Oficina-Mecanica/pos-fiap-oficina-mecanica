---
documento: Refinamento de Requisitos — Registrar Consumo e Saída de Peças
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Registrar Consumo e Saída de Peças

Este documento detalha a tarefa Registrar Consumo e Saída de Peças do contexto de Peças.

> **Escopo deste documento.** A baixa é **uma rota só**, `POST /estoque/saidas`, compartilhada com
> Insumos — mesmo padrão da entrada e da compra, porque o mecânico dá baixa de tudo o que usou no
> serviço de uma vez. Este documento descreve a parte de **peças**; a de insumos está em
> [registrar-consumo-e-saida-de-insumos.md](../insumos/registrar-consumo-e-saida-de-insumos.md).

## 14 · Registrar Consumo e Saída de Peças

### 14.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Dar baixa no estoque das peças efetivamente usadas na execução do serviço, consumindo a reserva
feita na aprovação do orçamento e reduzindo o saldo físico.

**Problema**

Hoje a peça é reservada na aprovação e nunca sai do estoque: o saldo reservado cresce, o saldo
físico não diminui, e o sistema continua dizendo que a peça está na prateleira depois de ela ter
sido montada no carro. Sem essa baixa, **o controle de estoque inteiro perde o sentido** — o
inventário nunca bate, a compra é disparada tarde, e o custo do material não chega à OS.

**Pré-condições**

- A OS deve existir e estar em execução.
- As peças devem possuir reserva `ATIVA` vinculada à OS.
- O usuário deve estar autorizado a movimentar estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-113 | Permitir registrar a baixa das peças usadas na execução da OS. |
| RF-PEC-114 | Permitir registrar a baixa de várias peças em uma única operação. |
| RF-PEC-115 | Consumir a reserva ativa da peça, na quantidade baixada. |
| RF-PEC-116 | Reduzir o saldo físico na quantidade baixada. |
| RF-PEC-117 | Reduzir o saldo reservado na quantidade consumida. |
| RF-PEC-118 | Impedir baixa maior que a quantidade reservada para a OS. |
| RF-PEC-119 | Registrar a movimentação de saída no histórico de estoque. |
| RF-PEC-120 | Devolver ao saldo livre a quantidade reservada e não consumida, quando a baixa for menor que a reserva. |
| RF-PEC-121 | Informar à OS o custo das peças baixadas. |
| RF-PEC-122 | Permitir mais de uma baixa para a mesma OS ao longo da execução. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-68 | A operação deve ser feita por API RESTful. |
| RNF-PEC-69 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-PEC-70 | A baixa deve ser transacional: ou todas as linhas saem, ou nenhuma sai. |
| RNF-PEC-71 | A operação deve ser idempotente por `Idempotency-Key`. |
| RNF-PEC-72 | O saldo físico e o saldo reservado nunca podem ficar negativos. |
| RNF-PEC-73 | A operação deve ser protegida contra concorrência, com lock de linha ordenado por `item_id`. |
| RNF-PEC-74 | A movimentação registrada deve ser auditável e imutável. |

**Fluxo Principal**

1. O mecânico informa a OS e as peças efetivamente usadas, com as quantidades.
2. O sistema valida a OS, o status e as reservas.
3. O sistema confere que a quantidade baixada não excede a reservada.
4. O sistema reduz o saldo físico e o saldo reservado de cada peça.
5. O sistema marca a reserva como `CONSUMIDA`, total ou parcialmente.
6. O sistema registra a movimentação de saída no histórico.
7. O sistema devolve ao saldo livre o que foi reservado e não usado, quando houver.
8. O sistema confirma a baixa e devolve os saldos atualizados e o custo apurado.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | OS não encontrada | Impede a operação. |
| A2 | OS fora de execução | Impede a operação: a baixa acontece durante a execução do serviço. |
| A3 | Peça sem reserva ativa para a OS | Impede a baixa e sinaliza que a peça não foi reservada para aquela OS. |
| A4 | Quantidade baixada maior que a reservada | Impede a baixa e informa a quantidade disponível na reserva. |
| A5 | Quantidade menor ou igual a zero | Impede a baixa. |
| A6 | Item repetido na mesma requisição | Impede a baixa. |
| A7 | Item de tipo divergente da linha | Trata a linha pelas regras do seu tipo: peça é inteira, insumo é fracionário. |
| A8 | Baixa parcial | Consome a parte usada e devolve o restante ao saldo livre. |
| A9 | Requisição repetida com a mesma `Idempotency-Key` | Devolve a resposta original, sem repetir a baixa. |
| A10 | Usuário sem autorização | Impede a operação. |

**Saída**

- Relação das peças baixadas, com quantidade, saldo físico e saldo reservado atualizados.
- Custo total das peças baixadas para a OS.
- Relação do que foi devolvido ao saldo livre, quando houver.
- Ou indicação do motivo pelo qual a baixa foi recusada.

**Pós-condições**

- O saldo físico das peças baixadas está reduzido.
- As reservas correspondentes estão `CONSUMIDA`, e o saldo reservado foi reduzido.
- As movimentações de saída estão registradas no histórico.
- A OS conhece o custo das peças usadas.

---

### 14.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/saidas
```

> **Decisão de projeto.** A baixa é **uma rota só** para peça e insumo, como a entrada e a compra.
> O mecânico termina o serviço e dá baixa de tudo o que usou de uma vez; dividir por tipo obrigaria
> duas chamadas para o mesmo consumo e quebraria a transação e a idempotência.

> **Decisão de projeto.** A baixa **consome a reserva**, não o saldo livre. Só sai do estoque o que
> foi reservado na aprovação do orçamento — é isso que impede o serviço de consumir peça
> comprometida com outra OS. Peça não reservada para aquela OS é recusada, e não reservada na hora.

> **Decisão de projeto.** A baixa **menor** que a reserva libera a diferença de volta ao saldo
> livre na mesma operação. Sem isso, o excedente ficaria reservado para sempre em uma OS que já
> terminou.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `estoque:movimentar`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | **Obrigatório**; impede dupla baixa em caso de reenvio. |
| Body | `ordemServicoId` | uuid | Obrigatório; OS em execução. |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido. |
| Body | `itens[].itemId` | uuid | Obrigatório; item com reserva ativa para a OS. |
| Body | `itens[].quantidade` | decimal | Maior que zero; inteiro quando o item for peça. |
| Body | `liberarSaldoNaoUsado` | boolean | Default `true`; devolve ao saldo livre o reservado e não consumido. |

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "itens": [
    { "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", "quantidade": 4 }
  ],
  "liberarSaldoNaoUsado": true
}
```

**Validações**

*Técnicas*

- `Idempotency-Key` obrigatório, em formato UUID.
- `ordemServicoId` obrigatório e em formato UUID.
- `itens` não vazio, sem `itemId` repetido.
- `quantidade` maior que zero; inteira nas linhas do tipo `PECA`.

*Negócio*

- A OS deve existir e estar em execução.
- Cada item deve possuir reserva `ATIVA` vinculada à OS.
- A quantidade baixada não pode exceder a quantidade ainda reservada para a OS.
- O saldo físico e o saldo reservado não podem ficar negativos.
- A baixa não altera o orçamento nem o status da OS.

**Processamento**

1. Verificar o `Idempotency-Key`; se já processada, retornar a resposta original.
2. Abrir transação.
3. Carregar a OS e validar o status.
4. Carregar os itens com `SELECT ... FOR UPDATE`, ordenados por `item_id`.
5. Carregar as reservas `ATIVA` da OS para cada item.
6. Validar que a quantidade baixada não excede a reservada.
7. Para cada linha: `saldo_fisico -= quantidade` e `saldo_reservado -= quantidade`.
8. Marcar a reserva como `CONSUMIDA` quando a quantidade reservada se esgotar, ou reduzir a
   quantidade remanescente.
9. Inserir uma `movimentacao_estoque` do tipo `SAIDA` por linha.
10. Havendo `liberarSaldoNaoUsado`, liberar o remanescente da reserva e inserir movimentação
    `LIBERACAO_RESERVA`.
11. Calcular o custo das peças baixadas e informá-lo à OS.
12. Commit e registro na trilha de auditoria.

**Persistência**

- Consulta: `ordem_servico`, `item_estoque`, `reserva_estoque`, `chave_idempotencia`.
- Altera: `item_estoque.saldo_fisico`, `item_estoque.saldo_reservado`, `reserva_estoque.status`,
  `movimentacao_estoque` (insert), `chave_idempotencia` (insert).
- Não altera: `orcamento`, `pedido_compra` e o status da OS.
- Tudo em uma transação: ou todas as linhas saem, ou nada é gravado.

**Saída da API**

```json
{
  "saidaId": "7e1c4a92-05bd-4f37-9a68-2c31de84b0f5",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "registradoEm": "2026-08-23T14:20:00-03:00",
  "registradoPor": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29",
  "itens": [
    {
      "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "codigo": "PEC-000142",
      "tipo": "PECA",
      "quantidadeBaixada": 4,
      "quantidadeReservadaAntes": 5,
      "quantidadeLiberada": 1,
      "saldoFisicoAtual": 12,
      "saldoReservadoAtual": 0,
      "custoUnitario": 118.4,
      "custoTotal": 473.6
    }
  ],
  "custoTotalSaida": 473.6
}
```

No exemplo, cinco peças estavam reservadas, quatro foram usadas e uma voltou ao saldo livre.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Baixa registrada. |
| `200` | Requisição repetida com a mesma `Idempotency-Key` — retorna a resposta original. |
| `400` | Body inválido; quantidade menor ou igual a zero; item repetido; `Idempotency-Key` ausente. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `estoque:movimentar`. |
| `404` | Ordem de Serviço ou item não encontrado. |
| `409` | OS fora de execução. |
| `409` | Item sem reserva ativa para a OS; quantidade baixada maior que a reservada. |

**Dependências**

- `ItemEstoqueRepository`.
- `ReservaEstoqueRepository`.
- `MovimentacaoEstoqueRepository`.
- `OrdemDeServicoRepository`, para validar o status e informar o custo.
- Serviço de idempotência.
- Trilha de auditoria.

**Testes**

*Unitários*

- Baixa total consome a reserva e reduz saldo físico e reservado.
- Baixa parcial consome parte e libera o restante ao saldo livre.
- Rejeita baixa maior que a quantidade reservada.
- Rejeita item sem reserva ativa para a OS.
- Rejeita quantidade zero e item repetido.
- Saldo físico e reservado nunca ficam negativos.
- Cálculo do custo total da saída.

*Integração*

- `POST /estoque/saidas` válido retorna `201` e reduz o saldo físico.
- Baixa parcial devolve o remanescente ao saldo disponível.
- Repetição com a mesma `Idempotency-Key` retorna `200` sem baixar de novo.
- Item sem reserva retorna `409`.
- Quantidade acima da reservada retorna `409`.
- OS fora de execução retorna `409`.
- A baixa não altera o orçamento nem o status da OS.

*Concorrência*

- Itens carregados com lock de linha, ordenados por `item_id`, para reduzir risco de deadlock.
- Duas baixas simultâneas na mesma OS não deixam saldo negativo.

---

### 14.3 Checklist de Implementação

**Domínio**

- [ ] Implementar a movimentação do tipo `SAIDA`
- [ ] Implementar o consumo da reserva, total e parcial, com o status `CONSUMIDA`
- [ ] Implementar a liberação do saldo reservado e não consumido
- [ ] Garantir que saldo físico e reservado nunca fiquem negativos
- [ ] Garantir que a baixa não altera orçamento nem status da OS

**Caso de uso**

- [ ] Implementar `RegistrarSaidaDeEstoque`
- [ ] Validar a OS, o status e as reservas
- [ ] Calcular o custo total da saída e informá-lo à OS

**Repositório**

- [ ] Criar consulta das reservas ativas por OS e item
- [ ] Persistir movimentação, saldos e status da reserva

**Integrações**

- [ ] Integrar com `OrdemDeServicoRepository` para validar o status e registrar o custo

**Handler HTTP**

- [ ] Implementar `POST /estoque/saidas`, compartilhado com Insumos
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização pelo escopo `estoque:movimentar`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar `Idempotency-Key` obrigatório
- [ ] Validar quantidade maior que zero e inteira para peça
- [ ] Validar reserva ativa para a OS
- [ ] Validar que a baixa não excede a quantidade reservada

**Concorrência**

- [ ] Carregar os itens com lock de linha, ordenados por `item_id`

**Transação e idempotência**

- [ ] Executar toda a baixa em uma única transação
- [ ] Devolver a resposta original na repetição da mesma `Idempotency-Key`

**Auditoria**

- [ ] Registrar a saída na trilha de auditoria

**Testes unitários**

- [ ] Baixa total
- [ ] Baixa parcial com liberação do remanescente
- [ ] Baixa acima da reserva rejeitada
- [ ] Item sem reserva rejeitado
- [ ] Cálculo do custo

**Testes de integração**

- [ ] `201` com saldo físico reduzido
- [ ] `200` na repetição da mesma `Idempotency-Key`
- [ ] `409` para item sem reserva e para quantidade acima da reservada
- [ ] `409` para OS fora de execução

**Testes de concorrência**

- [ ] Duas baixas simultâneas na mesma OS não geram saldo negativo

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
