---
documento: Refinamento de Requisitos — Reservar Peça para Ordem de Serviço
dono: A definir
versao: 0.4
atualizado_em: 2026-08-22
status: em revisao
---

# Refinamento de Requisitos — Reservar Peça para Ordem de Serviço

Este documento detalha a tarefa Reservar Peça para Ordem de Serviço do contexto de Peças.

## 5 · Reservar Peça para Ordem de Serviço

### 5.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Separar as peças necessárias de uma Ordem de Serviço aprovada para garantir sua disponibilidade durante a execução do serviço.

**Problema**

Duas Ordens de Serviço aprovadas podem depender da última unidade de uma mesma peça. Sem reserva, uma OS pode consumir a peça prometida a outra e o cliente só descobre a indisponibilidade com o veículo já desmontado.

**Pré-condições**

- A Ordem de Serviço existe e possui orçamento aprovado vigente.
- A Ordem de Serviço possui peças necessárias registradas no orçamento aprovado.
- As peças solicitadas estão cadastradas, ativas e pertencem à OS ou ao orçamento aprovado.
- O usuário ou serviço possui permissão para reservar peças.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-41 | Permitir reservar uma ou mais peças para uma Ordem de Serviço. |
| RF-PEC-42 | Validar que cada peça e quantidade solicitada pertence à OS ou ao seu orçamento aprovado. |
| RF-PEC-43 | Validar o saldo disponível de todas as peças antes de confirmar a reserva. |
| RF-PEC-44 | Aumentar o saldo reservado e reduzir o saldo disponível lógico sem alterar o saldo físico. |
| RF-PEC-45 | Vincular cada reserva à OS de origem e atualizar a OS com as peças reservadas. |
| RF-PEC-46 | Informar as peças indisponíveis quando não houver saldo para concluir a reserva. |
| RF-PEC-47 | Registrar a movimentação de reserva no histórico de estoque. |
| RF-PEC-48 | Retornar a reserva vigente quando a mesma solicitação já tiver sido processada. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-26 | A operação deve ser RESTful, autenticada e autorizada. |
| RNF-PEC-27 | A reserva deve ser atômica: todas as peças são reservadas ou nenhuma é. |
| RNF-PEC-28 | A operação deve impedir que duas OS reservem simultaneamente a mesma unidade. |
| RNF-PEC-29 | A operação deve ser idempotente por `Idempotency-Key`, devolvendo a resposta original em repetição válida. |
| RNF-PEC-30 | A reserva não deve alterar saldo físico nem outras movimentações já registradas. |

**Fluxo Principal**

1. O mecânico solicita a reserva das peças para uma OS aprovada.
2. O sistema valida a chave de idempotência, a OS e os itens informados.
3. O sistema confirma que as peças pertencem à OS ou ao orçamento aprovado e estão ativas.
4. O sistema verifica o saldo disponível de todas as peças.
5. O sistema reserva as quantidades, vincula as reservas à OS e registra as movimentações.
6. O sistema confirma a reserva com os saldos disponíveis após a operação.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Uma ou mais peças não têm saldo disponível suficiente | Não reserva nenhuma peça, informa as indisponíveis e registra a indisponibilidade. |
| A2 | A mesma `Idempotency-Key` já foi processada | Retorna a resposta original, sem repetir a reserva. |
| A3 | Já existe reserva ativa equivalente para a OS | Retorna a reserva vigente, sem duplicar saldos ou vínculos. |
| A4 | A OS não possui orçamento aprovado vigente | Impede a reserva. |
| A5 | Item inativo, insumo ou não pertencente à OS/orçamento | Impede a reserva. |
| A6 | Usuário ou serviço sem autorização | Impede a operação. |

**Saída**

- Confirmação da reserva com as peças, quantidades e saldos disponíveis após a reserva; ou a indicação das peças indisponíveis que impediram a operação.

**Pós-condições**

- O saldo reservado das peças é acrescido das quantidades reservadas.
- O saldo disponível lógico é reduzido na mesma proporção.
- O saldo físico permanece inalterado.
- As peças reservadas ficam vinculadas à OS e a movimentação de reserva fica registrada.

---

### 5.2 Refinamento Técnico

**Gatilho**

Não há endpoint: é uma chamada em processo, dentro de `ProcessarPecas`.

```
ProcessarPecas
├── separa, por item, o disponível do faltante
├── ReservarPecas(os, itens disponíveis)   ← esta tarefa
├── SolicitarCompra(os, itens faltantes)
└── confirma a transação
```

> **Decisão de projeto — rota aposentada.** `POST /estoque/reservas` **saiu da API**. Com a D-16 fechada, a
> aprovação do orçamento passou a ser o único gatilho que compromete estoque, e ela chama o
> processamento, não a reserva direta. A rota ficou sem chamador público, e manter uma porta
> aberta para comprometer saldo por fora do fluxo de aprovação é justamente o que a D-16
> resolveu. O refinamento abaixo continua valendo: ele descreve as **regras da reserva**,
> agora executadas como serviço de domínio chamado por
> [processar-pecas-para-reserva-e-compra.md](processar-pecas-para-reserva-e-compra.md).

> As regras de concorrência, idempotência e histórico continuam necessárias — elas passam a
> valer para a transação aberta pelo processamento, que é quem recebe a `Idempotency-Key`.

**Autenticação / Autorização**

Não se aplica: a autorização já foi verificada pelo caso de uso que expõe o endpoint — o
processamento de reserva e compra, com escopo `estoque:movimentar`.

**Entrada**

Os parâmetros abaixo chegam do caso de uso chamador, não de um corpo HTTP.

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Interno | `ordemServicoId` | uuid | Obrigatório. Identificador da OS com orçamento aprovado. |
| Interno | `itens` | array | Obrigatório, não vazio e sem `itemId` repetido. |
| Interno | `itens[].itemId` | uuid | Obrigatório. Deve identificar uma peça ativa vinculada à OS ou ao orçamento. |
| Interno | `itens[].quantidade` | integer | Obrigatório. Inteiro maior que zero. |

```json
{
  "ordemServicoId": "3c4f321d-9e62-4cc4-8d3c-412c9c2035c7",
  "itens": [
    { "itemId": "f0b13c55-39b7-4e31-a258-619b6c77c18b", "quantidade": 4 },
    { "itemId": "59ca8fd5-7371-4df2-9b87-ddf657818da4", "quantidade": 1 }
  ]
}
```

**Validações**

*Técnicas*

- `Idempotency-Key` é obrigatório e deve ter formato UUID.
- `ordemServicoId`, `itens`, `itemId` e `quantidade` são obrigatórios.
- `itens` não pode ser vazio, conter `itemId` repetido nem quantidade não inteira ou menor que um.

*Negócio*

- A OS existe e possui orçamento aprovado vigente.
- Cada item existe, está ativo, é do tipo `PECA` e pertence à OS ou ao orçamento aprovado.
- Todas as peças possuem saldo disponível suficiente.
- Não existe reserva `ATIVA` equivalente para a mesma OS e os mesmos itens.

**Processamento**

1. Validar o header e o payload; se a chave já foi processada, devolver a resposta original.
2. Consultar a OS, o orçamento aprovado vigente e os itens necessários.
3. Verificar se já há reserva `ATIVA` equivalente; se houver, retornar a reserva vigente.
4. Abrir transação e carregar as peças por `item_id` em ordem ascendente, com lock de linha.
5. Calcular `saldoDisponivel = saldoFisico - saldoReservado` para cada peça e validar todas as quantidades.
6. Se qualquer peça não tiver saldo, desfazer a transação, registrar a indisponibilidade e retornar os itens insuficientes.
7. Aumentar `saldo_reservado`, criar as reservas `ATIVA`, vincular as peças à OS e registrar movimentações `RESERVA`.
8. Confirmar a transação e armazenar a resposta para a chave de idempotência.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `chave_idempotencia`, Ordem de Serviço, orçamento aprovado e itens vinculados.
- Altera: `item_estoque.saldo_reservado`, `reserva_estoque`, `movimentacao_estoque`, registro da OS e `chave_idempotencia`.
- Não altera: `item_estoque.saldo_fisico`.
- Transação: isolamento mínimo `READ COMMITTED` com lock explícito de linha e ordenação por `item_id`.

**Saída da API**

```json
{
  "reservaId": "e7a15fb3-9f73-4a0a-a64c-3b61d0126c37",
  "ordemServicoId": "3c4f321d-9e62-4cc4-8d3c-412c9c2035c7",
  "status": "ATIVA",
  "reservadoEm": "2026-08-22T15:20:00-03:00",
  "itens": [
    {
      "itemId": "f0b13c55-39b7-4e31-a258-619b6c77c18b",
      "quantidade": 4,
      "saldoDisponivelApos": 8
    },
    {
      "itemId": "59ca8fd5-7371-4df2-9b87-ddf657818da4",
      "quantidade": 1,
      "saldoDisponivelApos": 2
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Reserva criada. |
| `200` | Reserva ativa equivalente encontrada ou repetição da mesma `Idempotency-Key`. |
| `400` | Header ou corpo inválido, lista vazia, quantidade inválida ou item repetido. |
| `401` | Usuário ou serviço não autenticado. |
| `403` | Perfil sem o escopo `estoque:movimentar`. |
| `404` | OS ou peça não encontrada. |
| `409` | Saldo insuficiente; nenhuma peça foi reservada. |
| `409` | OS sem orçamento aprovado, peça inativa, insumo ou peça fora da OS/orçamento. |

**Dependências**

- `ItemEstoqueRepository`, `ReservaEstoqueRepository`, `MovimentacaoEstoqueRepository` e `ChaveIdempotenciaRepository`.
- Módulos de Ordem de Serviço e Orçamento.
- Serviço de idempotência e trilha de auditoria.

**Testes**

*Unitários*

- Reserva válida atualiza somente `saldo_reservado` e retorna os saldos disponíveis após a operação.
- Entrada ausente ou inválida, item repetido, insumo, peça inexistente/inativa e peça fora da OS são rejeitados.
- OS sem orçamento aprovado, saldo insuficiente e reserva ativa equivalente respeitam os códigos e efeitos previstos.
- A repetição da mesma chave não cria nova reserva; falha de saldo mantém todos os saldos inalterados.

*Integração*

- `POST /estoque/reservas` cria reservas, movimentações e vínculo com a OS quando autenticado e autorizado.
- Ausência de autenticação, falta de escopo, OS/peça inexistente e regra de negócio inválida retornam `401`, `403`, `404` e `409` adequadamente.
- Duas OS concorrentes disputando a última peça não deixam `saldo_reservado` maior que `saldo_fisico` e não geram deadlock.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Modelar `ReservaEstoque` com status `ATIVA` e o cálculo de saldo disponível.
- [ ] Garantir que somente item do tipo `PECA` seja reservável.
- [ ] Modelar as movimentações `RESERVA`, `PecaReservada` e `PecaIndisponivel`.

**Caso de uso**

- [ ] Implementar `ReservarPecaParaOS` com validação de OS, orçamento e vínculo dos itens.
- [ ] Garantir a regra tudo ou nada e o retorno da reserva ativa equivalente.

**Repositório**

- [ ] Implementar consulta de peças com lock ordenado por `item_id`.
- [ ] Implementar persistência de reservas, movimentações e chave de idempotência.

**Integrações**

- [ ] Consultar a OS, seu orçamento aprovado e os itens necessários.
- [ ] Atualizar a OS com as peças reservadas, por chamada direta na mesma transação.

**Handler HTTP**

- [ ] Implementar `POST /estoque/reservas` com `Idempotency-Key` e resposta da reserva.
- [ ] Aplicar JWT, perfis permitidos e escopo `estoque:movimentar`.

**Validações**

- [ ] Validar UUID da chave de idempotência, OS, itens e quantidades.
- [ ] Validar itens repetidos, tipo `PECA`, item ativo, vínculo com OS/orçamento e saldo disponível.

**Transação e idempotência**

- [ ] Executar reserva, atualização de saldos, vínculo com OS e movimentações na mesma transação.
- [ ] Retornar a resposta original para repetição da mesma `Idempotency-Key`.

**Auditoria**

- [ ] Registrar na auditoria a reserva confirmada e a falha por saldo insuficiente.

**Testes unitários**

- [ ] Cobrir reserva válida, regras de entrada, OS/orçamento, itens e saldo insuficiente.
- [ ] Cobrir tudo ou nada, saldo físico inalterado, reserva equivalente e idempotência.

**Testes de integração**

- [ ] Cobrir contrato de sucesso, autenticação, autorização e todos os erros mapeados.
- [ ] Validar persistência de reservas, movimentações e vínculo com a OS.

**Testes de concorrência**

- [ ] Validar disputa pela última peça, ausência de saldo reservado acima do físico e prevenção de deadlock.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Code Review aprovado.
