---
documento: Refinamento de Requisitos — Processar Peças para Reserva e Compra
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em revisao
---

# Refinamento de Requisitos — Processar Peças para Reserva e Compra

Este documento detalha a tarefa Processar Peças para Reserva e Compra do contexto de Peças & Insumos.

## 12 · Processar Peças para Reserva e Compra

### 12.1 Refinamento de Produto

**Persona**

Sistema, acionado pela aprovação do orçamento. O mecânico acompanha as peças garantidas e as que dependem de compra.

**Objetivo**

Reservar as peças disponíveis e solicitar compra somente das peças faltantes para uma Ordem de Serviço aprovada.

**Problema**

Uma OS pode ter apenas parte das peças disponível. A oficina precisa garantir imediatamente o que existe, registrar o que falta comprar e deixar claro por que a execução está aguardando recursos.

**Pré-condições**

- A Ordem de Serviço existe e possui orçamento aprovado vigente.
- As peças e quantidades informadas estão cadastradas, ativas e pertencem à OS ou ao orçamento aprovado.
- O fornecedor informado existe e está ativo.
- O usuário ou serviço possui permissão para processar reserva e compra.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-136 | Receber uma OS e as peças que devem ser processadas após a aprovação do orçamento. |
| RF-EST-137 | Identificar, para cada peça, a quantidade disponível para reserva e a quantidade pendente de compra. |
| RF-EST-138 | Reservar somente a quantidade disponível em estoque e vinculá-la à OS. |
| RF-EST-139 | Criar solicitação de compra para a quantidade pendente, vinculada à OS e ao fornecedor. |
| RF-EST-140 | Atualizar a OS com peças reservadas, peças pendentes e o status `AGUARDANDO_RECURSOS` quando houver pendência. |
| RF-EST-141 | Registrar as movimentações e solicitações no histórico. |
| RF-EST-142 | Publicar os eventos de reserva parcial e de compra solicitada quando aplicáveis. |
| RF-EST-143 | Retornar o estado vigente quando a mesma solicitação já tiver sido processada. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-98 | A operação deve ser RESTful, autenticada e autorizada. |
| RNF-EST-99 | A reserva de saldo existente deve ser protegida contra concorrência. |
| RNF-EST-100 | As reservas, a solicitação de compra e a atualização da OS devem ser consistentes e transacionais. |
| RNF-EST-101 | A operação deve ser idempotente por `Idempotency-Key`. |
| RNF-EST-102 | O saldo físico não deve ser alterado durante o processamento. |

**Fluxo Principal**

1. A aprovação do orçamento aciona o processamento das peças da OS.
2. O sistema valida a OS, o fornecedor, a chave de idempotência e os itens informados.
3. O sistema calcula o saldo disponível de cada peça e separa as quantidades reserváveis das pendentes de compra.
4. O sistema reserva as quantidades disponíveis e registra as respectivas movimentações.
5. O sistema cria a solicitação de compra para as quantidades pendentes.
6. O sistema atualiza a OS; havendo pendência, define o status como `AGUARDANDO_RECURSOS`.
7. O sistema confirma a operação e publica os eventos aplicáveis.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhuma peça tem saldo disponível | Não cria reservas e solicita compra de todos os itens. |
| A2 | Todas as peças têm saldo disponível | Reserva todos os itens e não cria solicitação de compra. |
| A3 | Uma parte das peças está disponível | Reserva o disponível e solicita compra somente do saldo faltante. |
| A4 | OS sem orçamento aprovado, fornecedor inativo ou peça inválida | Impede o processamento. |
| A5 | Reserva ou solicitação equivalente já existe | Retorna o estado vigente, sem duplicar a operação. |
| A6 | Falha durante reserva, compra ou atualização da OS | Desfaz a operação e não deixa a OS em estado inconsistente. |
| A7 | Usuário ou serviço sem autorização | Impede a operação. |

**Saída**

- Relação de peças reservadas, peças com compra solicitada, valor parcial conhecido, fornecedor e status atualizado da OS.

**Pós-condições**

- Peças disponíveis ficam reservadas para a OS.
- Peças pendentes ficam vinculadas a uma solicitação de compra.
- A OS registra as duas situações e fica em `AGUARDANDO_RECURSOS` quando houver pendência.
- O saldo físico permanece inalterado.

---

### 12.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/solicitacoes-compra-reserva
```

> **Decisão de projeto.** Esta é uma operação de orquestração, e não uma alternativa genérica às rotas de reserva ou pedido de compra: ela decide as duas saídas com base no saldo disponível. A alternativa de o cliente chamar `POST /estoque/reservas` e `POST /compras/pedidos` separadamente foi descartada por não garantir consistência quando o estoque muda entre as chamadas.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório para chamadas de usuário ou serviço.
- Perfis permitidos: `SERVICO`, `MECANICO`, `GESTOR`.
- Escopo: `estoque:movimentar`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | Obrigatório. Impede reprocessamento da mesma solicitação. |
| Body | `ordemServicoId` | uuid | Obrigatório. OS com orçamento aprovado. |
| Body | `fornecedorId` | uuid | Obrigatório. Fornecedor ativo da compra eventual. |
| Body | `itens` | array | Obrigatório, não vazio e sem `itemId` repetido. |
| Body | `itens[].itemId` | uuid | Obrigatório. Deve identificar uma peça ativa vinculada à OS ou ao orçamento. |
| Body | `itens[].quantidade` | integer | Obrigatório. Inteiro maior que zero. |

```json
{
  "ordemServicoId": "3c4f321d-9e62-4cc4-8d3c-412c9c2035c7",
  "fornecedorId": "dfe13e8c-e5a7-44ae-9712-8fdd8db4e932",
  "itens": [
    { "itemId": "f0b13c55-39b7-4e31-a258-619b6c77c18b", "quantidade": 2 },
    { "itemId": "59ca8fd5-7371-4df2-9b87-ddf657818da4", "quantidade": 1 }
  ]
}
```

**Validações**

*Técnicas*

- `Idempotency-Key`, `ordemServicoId`, `fornecedorId`, `itens`, `itemId` e `quantidade` são obrigatórios.
- Identificadores devem ter formato UUID; `itens` não pode ser vazio, conter repetição nem quantidade não inteira ou menor que um.

*Negócio*

- A OS existe e possui orçamento aprovado vigente; o fornecedor existe e está ativo.
- Cada item existe, está ativo, é do tipo `PECA` e pertence à OS ou ao orçamento aprovado.
- Não existe reserva `ATIVA` ou solicitação de compra aberta equivalente para a mesma OS e os mesmos itens.

**Processamento**

1. Validar o header e o payload; se a chave já foi processada, devolver a resposta original.
2. Consultar e validar OS, orçamento aprovado, fornecedor e peças vinculadas.
3. Abrir transação e carregar as peças por `item_id` em ordem ascendente, com lock de linha.
4. Calcular `saldoDisponivel = saldoFisico - saldoReservado`; para cada item, separar `quantidadeReservada` e `quantidadePendenteCompra`.
5. Para quantidades reserváveis, aumentar `saldo_reservado`, criar reservas `ATIVA` e movimentações `RESERVA` vinculadas à OS.
6. Para quantidades pendentes, criar a solicitação de compra, seus itens e o vínculo com a OS; registrar valor parcial quando houver preço disponível.
7. Atualizar a OS com as duas listas e definir `AGUARDANDO_RECURSOS` se houver pendência de compra.
8. Confirmar a transação, registrar a resposta da chave de idempotência e publicar `PecaReservadaParcialmente` e/ou `CompraDePecaSolicitada`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `solicitacao_compra`, `fornecedor`, `chave_idempotencia`, Ordem de Serviço, orçamento aprovado e itens vinculados.
- Altera: `item_estoque.saldo_reservado`, `reserva_estoque`, `solicitacao_compra`, `movimentacao_estoque`, OS e `chave_idempotencia`.
- Não altera: `item_estoque.saldo_fisico`.
- Transação: isolamento mínimo `READ COMMITTED` com lock explícito de linha e ordenação por `item_id`.

**Saída da API**

```json
{
  "ordemServicoId": "3c4f321d-9e62-4cc4-8d3c-412c9c2035c7",
  "statusOrdemServico": "AGUARDANDO_RECURSOS",
  "pecasReservadas": [
    {
      "itemId": "f0b13c55-39b7-4e31-a258-619b6c77c18b",
      "quantidade": 2,
      "saldoDisponivelApos": 0
    }
  ],
  "pecasCompraSolicitada": [
    {
      "itemId": "59ca8fd5-7371-4df2-9b87-ddf657818da4",
      "quantidade": 1,
      "valorParcial": 150.0
    }
  ],
  "fornecedor": {
    "id": "dfe13e8c-e5a7-44ae-9712-8fdd8db4e932",
    "nome": "Auto Peças Recife"
  },
  "valorTotalCompraParcial": 150.0
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Reservas e/ou solicitação de compra criadas. |
| `200` | Repetição da mesma `Idempotency-Key` ou estado equivalente já existente. |
| `400` | Header ou corpo inválido, lista vazia, quantidade inválida ou item repetido. |
| `401` | Usuário ou serviço não autenticado. |
| `403` | Perfil sem o escopo `estoque:movimentar`. |
| `404` | OS, fornecedor ou peça não encontrada. |
| `409` | Reserva ou solicitação equivalente já existente. |
| `422` | OS sem orçamento aprovado, fornecedor inativo, peça inativa, insumo ou peça fora da OS/orçamento. |

**Dependências**

- `FornecedorRepository`, `ItemEstoqueRepository`, `ReservaEstoqueRepository`, `SolicitacaoCompraRepository`, `MovimentacaoEstoqueRepository` e `ChaveIdempotenciaRepository`.
- Módulos de Ordem de Serviço e Orçamento, serviço de idempotência e publicador de eventos.

**Testes**

*Unitários*

- Processa todas disponíveis, nenhuma disponível e disponibilidade parcial com as quantidades corretas.
- Rejeita payload inválido, item repetido, insumo, peça inexistente/inativa, OS sem orçamento e fornecedor inexistente/inativo.
- Não altera saldo físico, evita duplicidade e mantém a OS consistente em falha.

*Integração*

- Cria reservas, solicitação de compra e atualização de OS conforme o saldo, autenticado e autorizado.
- Retorna os códigos previstos para autenticação, autorização, recursos inexistentes e regras de negócio.
- Execuções concorrentes não permitem `saldo_reservado` maior que `saldo_fisico` nem geram deadlock.

---

### 12.3 Checklist de Implementação

**Domínio**

- [ ] Modelar `SolicitacaoCompra` e a separação entre quantidade reservada e pendente de compra.
- [ ] Garantir que somente item do tipo `PECA` seja processado.

**Caso de uso**

- [ ] Implementar `SolicitarCompraEReservarPecas` com cálculo de saldo e atualização consistente da OS.
- [ ] Garantir processamento parcial, sem alterar o saldo físico.

**Repositório**

- [ ] Implementar consulta de peças com lock ordenado por `item_id`.
- [ ] Persistir reservas, solicitações de compra, movimentações e chave de idempotência.

**Integrações**

- [ ] Consultar OS, orçamento aprovado, fornecedor e itens vinculados.
- [ ] Atualizar a OS e publicar `PecaReservadaParcialmente` e `CompraDePecaSolicitada`.

**Handler HTTP**

- [ ] Implementar `POST /estoque/solicitacoes-compra-reserva` com `Idempotency-Key`.
- [ ] Aplicar JWT, perfis permitidos e escopo `estoque:movimentar`.

**Validações**

- [ ] Validar chave, identificadores, itens, fornecedor, tipo `PECA`, vínculo e situação da OS.

**Transação e idempotência**

- [ ] Executar reserva, solicitação, atualização da OS e movimentações na mesma transação.
- [ ] Retornar a resposta original para repetição da mesma `Idempotency-Key`.

**Eventos**

- [ ] Publicar os eventos de reserva parcial e de compra solicitada conforme o resultado.

**Testes unitários**

- [ ] Cobrir disponibilidade total, indisponibilidade total, parcial, erros de validação e idempotência.

**Testes de integração**

- [ ] Cobrir contrato, persistência, autenticação, autorização e atualização da OS.

**Testes de concorrência**

- [ ] Validar disputa de saldo, ausência de saldo reservado acima do físico e prevenção de deadlock.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Code Review aprovado.
