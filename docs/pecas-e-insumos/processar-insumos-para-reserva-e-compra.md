---
documento: Refinamento de Requisitos — Processar Insumos para Reserva e Compra
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em revisao
---

# Refinamento de Requisitos — Processar Insumos para Reserva e Compra

Este documento detalha a tarefa Processar Insumos para Reserva e Compra do contexto de Peças & Insumos.

## 13 · Processar Insumos para Reserva e Compra

### 13.1 Refinamento de Produto

**Persona**

Sistema, acionado pela aprovação do orçamento. O mecânico acompanha os insumos garantidos e os que dependem de compra.

**Objetivo**

Reservar os insumos disponíveis e solicitar compra somente do saldo faltante para uma Ordem de Serviço aprovada.

**Problema**

Uma OS pode ter somente parte dos insumos disponíveis. A oficina precisa garantir o que existe, registrar o que falta comprar e deixar claro por que a execução aguarda recursos.

**Pré-condições**

- A Ordem de Serviço existe e possui orçamento aprovado vigente.
- Os insumos e quantidades informados estão cadastrados, ativos e pertencem à OS ou ao orçamento aprovado.
- O fornecedor informado existe e está ativo.
- O usuário ou serviço possui permissão para processar reserva e compra.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-144 | Receber uma OS e os insumos que devem ser processados após a aprovação do orçamento. |
| RF-EST-145 | Identificar, para cada insumo, a quantidade disponível para reserva e a quantidade pendente de compra. |
| RF-EST-146 | Reservar somente a quantidade disponível em estoque e vinculá-la à OS. |
| RF-EST-147 | Criar solicitação de compra para a quantidade pendente, vinculada à OS e ao fornecedor. |
| RF-EST-148 | Atualizar a OS com insumos reservados, insumos pendentes e o status `AGUARDANDO_RECURSOS` quando houver pendência. |
| RF-EST-149 | Registrar as movimentações e solicitações no histórico. |
| RF-EST-150 | Publicar os eventos de reserva parcial e de compra solicitada quando aplicáveis. |
| RF-EST-151 | Retornar o estado vigente quando a mesma solicitação já tiver sido processada. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-103 | A operação deve ser RESTful, autenticada e autorizada. |
| RNF-EST-104 | A reserva de saldo existente deve ser protegida contra concorrência. |
| RNF-EST-105 | As reservas, a solicitação de compra e a atualização da OS devem ser consistentes e transacionais. |
| RNF-EST-106 | A operação deve ser idempotente por `Idempotency-Key`. |
| RNF-EST-107 | O saldo físico não deve ser alterado durante o processamento. |

**Fluxo Principal**

1. A aprovação do orçamento aciona o processamento dos insumos da OS.
2. O sistema valida a OS, o fornecedor, a chave de idempotência e os itens informados.
3. O sistema calcula o saldo disponível de cada insumo e separa as quantidades reserváveis das pendentes de compra.
4. O sistema reserva as quantidades disponíveis e registra as respectivas movimentações.
5. O sistema cria a solicitação de compra para as quantidades pendentes.
6. O sistema atualiza a OS; havendo pendência, define o status como `AGUARDANDO_RECURSOS`.
7. O sistema confirma a operação e publica os eventos aplicáveis.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhum insumo tem saldo disponível | Não cria reservas e solicita compra de todos os itens. |
| A2 | Todos os insumos têm saldo disponível | Reserva todos os itens e não cria solicitação de compra. |
| A3 | Uma parte dos insumos está disponível | Reserva o disponível e solicita compra somente do saldo faltante. |
| A4 | OS sem orçamento aprovado, fornecedor inativo ou insumo inválido | Impede o processamento. |
| A5 | Reserva ou solicitação equivalente já existe | Retorna o estado vigente, sem duplicar a operação. |
| A6 | Falha durante reserva, compra ou atualização da OS | Desfaz a operação e não deixa a OS em estado inconsistente. |
| A7 | Usuário ou serviço sem autorização | Impede a operação. |

**Saída**

- Relação de insumos reservados, insumos com compra solicitada, fornecedor e status atualizado da OS.

**Pós-condições**

- Insumos disponíveis ficam reservados para a OS.
- Insumos pendentes ficam vinculados a uma solicitação de compra.
- A OS registra as duas situações e fica em `AGUARDANDO_RECURSOS` quando houver pendência.
- O saldo físico permanece inalterado.

---

### 13.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/solicitacoes-compra-reserva-insumos
```

> **Decisão de projeto.** Esta operação orquestra reserva e compra para insumos com base no saldo disponível. A alternativa de chamadas independentes a reserva e compra foi descartada porque não preserva a consistência da OS quando o saldo muda entre as chamadas.

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
| Body | `itens[].itemId` | uuid | Obrigatório. Deve identificar um insumo ativo vinculado à OS ou ao orçamento. |
| Body | `itens[].quantidade` | decimal | Obrigatório. Quantidade maior que zero e compatível com a unidade de medida. |

```json
{
  "ordemServicoId": "3c4f321d-9e62-4cc4-8d3c-412c9c2035c7",
  "fornecedorId": "dfe13e8c-e5a7-44ae-9712-8fdd8db4e932",
  "itens": [
    { "itemId": "ff5a1b6e-fdfa-4e2a-a044-aaab498a41d2", "quantidade": 2.0 },
    { "itemId": "b2994057-a7d9-476c-86f4-01472ba7be45", "quantidade": 1.0 }
  ]
}
```

**Validações**

*Técnicas*

- `Idempotency-Key`, `ordemServicoId`, `fornecedorId`, `itens`, `itemId` e `quantidade` são obrigatórios.
- Identificadores devem ter formato UUID; `itens` não pode ser vazio, conter repetição nem quantidade menor ou igual a zero.
- A quantidade deve ser compatível com a unidade de medida do insumo.

*Negócio*

- A OS existe e possui orçamento aprovado vigente; o fornecedor existe e está ativo.
- Cada item existe, está ativo, é do tipo `INSUMO` e pertence à OS ou ao orçamento aprovado.
- Não existe reserva `ATIVA` ou solicitação de compra aberta equivalente para a mesma OS e os mesmos itens.

**Processamento**

1. Validar o header e o payload; se a chave já foi processada, devolver a resposta original.
2. Consultar e validar OS, orçamento aprovado, fornecedor e insumos vinculados.
3. Abrir transação e carregar os insumos por `item_id` em ordem ascendente, com lock de linha.
4. Calcular `saldoDisponivel = saldoFisico - saldoReservado`; para cada item, separar `quantidadeReservada` e `quantidadePendenteCompra`.
5. Para quantidades reserváveis, aumentar `saldo_reservado`, criar reservas `ATIVA` e movimentações `RESERVA` vinculadas à OS.
6. Para quantidades pendentes, criar a solicitação de compra, seus itens e o vínculo com a OS; registrar valor parcial quando houver valor disponível.
7. Atualizar a OS com as duas listas e definir `AGUARDANDO_RECURSOS` se houver pendência de compra.
8. Confirmar a transação, registrar a resposta da chave de idempotência e publicar `InsumoReservadoParcialmente` e/ou `CompraDeInsumoSolicitada`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `solicitacao_compra`, `fornecedor`, `chave_idempotencia`, Ordem de Serviço, orçamento aprovado e itens vinculados.
- Altera: `item_estoque.saldo_reservado`, `reserva_estoque`, `solicitacao_compra`, `movimentacao_estoque`, OS e `chave_idempotencia`.
- Não altera: `item_estoque.saldo_fisico`.
- Transação: isolamento mínimo `READ COMMITTED` com lock explícito de linha e ordenação por `item_id`.

**Saída da API**

```json
{
  "pedidoId": "8eb98188-2fc7-4b22-82d5-c4c10e06c1a3",
  "numero": "2026/0121",
  "fornecedor": {
    "id": "dfe13e8c-e5a7-44ae-9712-8fdd8db4e932",
    "nome": "Fornecedor de Insumos"
  },
  "status": "ABERTO",
  "criadoEm": "2026-08-22T17:40:00-03:00",
  "criadoPor": "a4e50101-9ae7-4e63-ac7c-7c584ce052a0",
  "itens": [
    {
      "itemId": "ff5a1b6e-fdfa-4e2a-a044-aaab498a41d2",
      "descricao": "Oleo lubrificante",
      "unidadeMedida": "L",
      "quantidadeNecessaria": 60.0,
      "quantidadePedida": 40.0,
      "quantidadeReservada": 20.0,
      "quantidadeRecebida": 0.0
    }
  ],
  "ordensServicoVinculadas": [
    {
      "ordemServicoId": "3c4f321d-9e62-4cc4-8d3c-412c9c2035c7",
      "status": "AGUARDANDO_RECURSOS",
      "insumosNecessarios": [
        { "itemId": "ff5a1b6e-fdfa-4e2a-a044-aaab498a41d2", "quantidade": 60.0 }
      ]
    }
  ]
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
| `404` | OS, fornecedor ou insumo não encontrado. |
| `409` | Reserva ou solicitação equivalente já existente. |
| `422` | OS sem orçamento aprovado, fornecedor inativo, insumo inativo, peça ou insumo fora da OS/orçamento. |

**Dependências**

- `FornecedorRepository`, `ItemEstoqueRepository`, `ReservaEstoqueRepository`, `SolicitacaoCompraRepository`, `MovimentacaoEstoqueRepository` e `ChaveIdempotenciaRepository`.
- Módulos de Ordem de Serviço e Orçamento, serviço de idempotência e publicador de eventos.

**Testes**

*Unitários*

- Processa todos disponíveis, nenhum disponível e disponibilidade parcial com as quantidades corretas.
- Rejeita payload inválido, item repetido, peça, insumo inexistente/inativo, unidade incompatível, OS sem orçamento e fornecedor inexistente/inativo.
- Não altera saldo físico, evita duplicidade e mantém a OS consistente em falha.

*Integração*

- Cria reservas, solicitação de compra e atualização de OS conforme o saldo, autenticado e autorizado.
- Retorna os códigos previstos para autenticação, autorização, recursos inexistentes e regras de negócio.
- Execuções concorrentes não permitem `saldo_reservado` maior que `saldo_fisico` nem geram deadlock.

---

### 13.3 Checklist de Implementação

**Domínio**

- [ ] Modelar `SolicitacaoCompra` e a separação entre quantidade reservada e pendente de compra para insumos.
- [ ] Garantir que somente item do tipo `INSUMO` seja processado e respeite sua unidade de medida.

**Caso de uso**

- [ ] Implementar `SolicitarCompraEReservarInsumos` com cálculo de saldo e atualização consistente da OS.
- [ ] Garantir processamento parcial, sem alterar o saldo físico.

**Repositório**

- [ ] Implementar consulta de insumos com lock ordenado por `item_id`.
- [ ] Persistir reservas, solicitações de compra, movimentações e chave de idempotência.

**Integrações**

- [ ] Consultar OS, orçamento aprovado, fornecedor e itens vinculados.
- [ ] Atualizar a OS e publicar `InsumoReservadoParcialmente` e `CompraDeInsumoSolicitada`.

**Handler HTTP**

- [ ] Implementar `POST /estoque/solicitacoes-compra-reserva-insumos` com `Idempotency-Key`.
- [ ] Aplicar JWT, perfis permitidos e escopo `estoque:movimentar`.

**Validações**

- [ ] Validar chave, identificadores, itens, fornecedor, tipo `INSUMO`, unidade de medida, vínculo e situação da OS.

**Transação e idempotência**

- [ ] Executar reserva, solicitação, atualização da OS e movimentações na mesma transação.
- [ ] Retornar a resposta original para repetição da mesma `Idempotency-Key`.

**Eventos**

- [ ] Publicar os eventos de reserva parcial e de compra solicitada conforme o resultado.

**Testes unitários**

- [ ] Cobrir disponibilidade total, indisponibilidade total, parcial, unidade de medida, erros de validação e idempotência.

**Testes de integração**

- [ ] Cobrir contrato, persistência, autenticação, autorização e atualização da OS.

**Testes de concorrência**

- [ ] Validar disputa de saldo, ausência de saldo reservado acima do físico e prevenção de deadlock.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Code Review aprovado.
