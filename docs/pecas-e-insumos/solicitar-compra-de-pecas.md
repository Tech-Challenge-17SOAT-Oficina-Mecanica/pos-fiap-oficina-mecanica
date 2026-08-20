---
documento: Refinamento de Requisitos — Solicitar Compra de Peças
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Solicitar Compra de Peças

Este documento detalha a tarefa Solicitar Compra de Peças do contexto de Peças & Insumos.

## 8 · Solicitar Compra de Peças

### 8.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Registrar um pedido de compra de peças junto ao fornecedor, formalizando a reposição do estoque.

**Problema**
Sem pedido formal, a compra vive na conversa de WhatsApp com o fornecedor: ninguém sabe o que
já foi pedido, o que está a caminho e qual o prazo. Isso gera compra duplicada e impede
informar ao cliente uma data confiável de entrega do veículo.

**Pré-condições**

- As peças devem estar cadastradas e ativas.
- Deve existir fornecedor cadastrado.
- O usuário deve estar autorizado a solicitar compra.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-45 | Permitir criar pedido de compra informando fornecedor, peças, quantidades e prazo previsto de entrega. |
| RF-EST-46 | Permitir gerar o pedido a partir da consulta de peças faltantes, com as quantidades sugeridas já preenchidas. |
| RF-EST-47 | Validar que as quantidades informadas são maiores que zero. |
| RF-EST-48 | Vincular o pedido às OS que dependem das peças solicitadas. |
| RF-EST-49 | Permitir cancelar pedido ainda não recebido. |
| RF-EST-50 | Registrar a situação do pedido: aberto, parcialmente recebido, concluído ou cancelado. |
| RF-EST-51 | Informar a data prevista de entrega para as OS vinculadas. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-38 | A operação deve ser feita por API RESTful. |
| RNF-EST-39 | A operação deve ser acessível somente por usuário autorizado com permissão de compras. |
| RNF-EST-40 | O pedido de compra não altera o saldo de estoque — o saldo só muda no registro da entrada. |
| RNF-EST-41 | O pedido deve ser auditável, com registro de quem solicitou e quando. |
| RNF-EST-42 | O sistema deve alertar quando já existir pedido em aberto para a mesma peça, evitando compra duplicada. |

**Fluxo Principal**

1. O mecânico seleciona as peças a serem compradas.
2. O mecânico informa o fornecedor, as quantidades e o prazo previsto de entrega.
3. O sistema valida as peças e as quantidades informadas.
4. O sistema verifica pedidos em aberto para as mesmas peças.
5. O sistema cria o pedido de compra com situação "aberto".
6. O sistema vincula o pedido às OS que dependem dessas peças.
7. O sistema informa a data prevista de entrega para as OS vinculadas.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Peça não encontrada ou inativa | Impede a inclusão do item no pedido. |
| A2 | Quantidade menor ou igual a zero | Impede a criação do pedido. |
| A3 | Já existe pedido em aberto para a mesma peça | Alerta e exige confirmação antes de criar um novo pedido. |
| A4 | Fornecedor não cadastrado | Impede a criação e permite seguir para o cadastro do fornecedor. |
| A5 | Pedido cancelado após criação | Atualiza a situação, desvincula as OS e sinaliza que a falta permanece. |
| A6 | Usuário sem autorização | Impede a operação. |

**Saída**

- Pedido de compra criado, com número, fornecedor, itens, quantidades, prazo previsto e OS vinculadas; **ou**
- Indicação do motivo pelo qual o pedido foi recusado.

**Pós-condições**

- O pedido de compra está registrado com situação "aberto".
- O saldo de estoque permanece inalterado.
- As OS que dependem das peças têm previsão de entrega associada.
- O pedido fica disponível para vinculação no registro de entrada de estoque.

---

### 8.2 Refinamento Técnico

**Endpoint**

```http
POST   /compras/pedidos
DELETE /compras/pedidos/{pedidoId}
```

O `DELETE` atende ao cancelamento de pedido ainda não recebido.

> **Decisão de projeto.** Compras **não** é um contexto delimitado separado: `pedido_compra` e
> `pedido_compra_item` pertencem ao contexto de Peças & Insumos, junto com o estoque que eles
> repõem. A alternativa era isolar Compras em contexto próprio e integrar por evento, mas isso
> exigiria sincronizar o recebimento (requisito 4) entre dois contextos para uma oficina de médio
> porte — complexidade sem retorno no MVP. Por isso o recebimento atualiza `pedido_compra` na mesma
> transação da entrada, e o prefixo de rota `/compras/` é apenas organização da API.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `compras:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Body | `fornecedorId` | uuid   | Obrigatório e existente |
| Body | `prazoPrevistoEntrega` | date | Obrigatório, no futuro |
| Body | `confirmarDuplicidade` | boolean | Obrigatório quando já houver pedido em aberto para a mesma peça |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido |
| Body | `itens[].itemId` | uuid   | Peça a comprar; item do tipo `INSUMO` é rejeitado |
| Body | `itens[].quantidade` | int | Inteiro maior que zero |
| Path (DELETE) | `pedidoId` | uuid   | Pedido a cancelar |

```json
{
  "fornecedorId": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
  "prazoPrevistoEntrega": "2026-08-20",
  "confirmarDuplicidade": false,
  "itens": [
    { "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90", "quantidade": 5 },
    { "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4", "quantidade": 10 }
  ]
}
```

**Validações**

*Técnicas*

- `fornecedorId` obrigatório e existente.
- `prazoPrevistoEntrega` no futuro.
- `itens` não vazio, sem repetição.
- `quantidade` inteira maior que zero.

*Negócio*

- Todos os itens existem, estão ativos e são do tipo `PECA`.
- Existindo pedido `ABERTO` ou `PARCIAL` para a mesma peça, exigir `confirmarDuplicidade: true`.
- Cancelamento permitido apenas em pedido sem recebimento (status `ABERTO`).

**Processamento**

1. Validar o payload e carregar o fornecedor.
2. Carregar e validar os itens.
3. Buscar pedidos em aberto para as mesmas peças.
4. Se houver e `confirmarDuplicidade = false`, retornar `409` com a lista de pedidos conflitantes.
5. Gerar o número sequencial do pedido.
6. Persistir `pedido_compra` com status `ABERTO` e seus `pedido_compra_item`.
7. Levantar as OS que dependem dessas peças e vincular ao pedido.
8. Publicar `PedidoCompraCriado` — a política notifica as OS impactadas com a previsão de entrega.

**Persistência**

- Consulta: `item_estoque`, `fornecedor`, `pedido_compra`, módulo de OS
- Altera: `pedido_compra` (insert), `pedido_compra_item` (insert)
- Não altera: nenhum saldo de estoque — o saldo só muda no registro da entrada

**Saída da API**

```json
{
  "pedidoId": "f05a1d63-8b47-49e2-a731-6c94d2e08b57",
  "numero": "2026/0118",
  "fornecedor": { "id": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7", "nome": "Auto Peças Recife" },
  "status": "ABERTO",
  "prazoPrevistoEntrega": "2026-08-20",
  "criadoEm": "2026-08-12T17:05:00-03:00",
  "criadoPor": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29",
  "itens": [
    {
      "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90",
      "codigo": "PC-0311",
      "descricao": "Disco de freio ventilado",
      "quantidadePedida": 5,
      "quantidadeRecebida": 0
    },
    {
      "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "codigo": "PC-0142",
      "descricao": "Pastilha de freio dianteira",
      "quantidadePedida": 10,
      "quantidadeRecebida": 0
    }
  ],
  "ordensServicoVinculadas": ["e21b7c46-0d95-4f83-a6b1-3c5d92e74801"]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Pedido criado |
| `204` | Pedido cancelado (`DELETE`) |
| `400` | Body inválido; prazo no passado; quantidade menor ou igual a zero |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `compras:escrever` |
| `404` | Fornecedor, item ou pedido não encontrado |
| `409` | Pedido em aberto para a mesma peça sem `confirmarDuplicidade`; tentativa de cancelar pedido com recebimento parcial |
| `422` | Item inativo ou do tipo `INSUMO` |

**Dependências**

- `PedidoCompraRepository`
- `ItemEstoqueRepository`
- `FornecedorRepository`
- Módulo Ordem de Serviço — vinculação e previsão de entrega
- Publicador de eventos de domínio
- Casos de uso Consultar Peças Faltantes (origem) e Registrar Entrada de Estoque (destino)

**Testes**

*Unitários*

- Rejeita prazo no passado.
- Rejeita quantidade zero.
- Detecta pedido em aberto para a mesma peça.
- Cálculo do número sequencial.

*Integração*

- Pedido válido retorna `201` com status `ABERTO`.
- Duplicidade sem confirmação retorna `409`.
- Duplicidade com `confirmarDuplicidade: true` retorna `201`.
- Item do tipo `INSUMO` retorna `422`.
- Criar pedido não altera nenhum saldo de estoque.
- `DELETE` em pedido `ABERTO` retorna `204`.
- `DELETE` em pedido `PARCIAL` retorna `409`.

---

### 8.3 Checklist de Implementação

**Domínio**

- [ ] Implementar a entidade `PedidoCompra` com status `ABERTO`, `PARCIAL`, `CONCLUIDO` e `CANCELADO`
- [ ] Implementar a entidade `PedidoCompraItem` com `quantidadePedida` e `quantidadeRecebida`
- [ ] Implementar a regra de cancelamento permitido apenas em pedido sem recebimento
- [ ] Implementar a geração de número sequencial do pedido
- [ ] Garantir que criar pedido não altera nenhum saldo de estoque

**Caso de uso**

- [ ] Implementar `SolicitarCompraDePecas`
- [ ] Implementar `CancelarPedidoCompra`
- [ ] Implementar a vinculação das OS que dependem das peças solicitadas

**Repositório**

- [ ] Implementar `PedidoCompraRepository`
- [ ] Implementar a consulta de pedidos em aberto para as mesmas peças
- [ ] Integrar com `FornecedorRepository`

**Handler HTTP**

- [ ] Implementar `POST /compras/pedidos`
- [ ] Implementar `DELETE /compras/pedidos/{pedidoId}`

**Validações**

- [ ] Validar fornecedor existente
- [ ] Validar `prazoPrevistoEntrega` no futuro
- [ ] Validar `quantidade` inteira maior que zero e sem repetição de item
- [ ] Rejeitar item do tipo `INSUMO` neste caso de uso
- [ ] Exigir `confirmarDuplicidade` quando já houver pedido em aberto para a mesma peça

**Eventos**

- [ ] Publicar `PedidoCompraCriado`
- [ ] Implementar a política de notificação de previsão de entrega para as OS vinculadas

**Testes unitários**

- [ ] Rejeição de prazo no passado
- [ ] Rejeição de quantidade zero
- [ ] Detecção de pedido em aberto para a mesma peça
- [ ] Geração do número sequencial

**Testes de integração**

- [ ] Pedido válido retornando `201` com status `ABERTO`
- [ ] Duplicidade sem confirmação retornando `409`
- [ ] Duplicidade com `confirmarDuplicidade: true` retornando `201`
- [ ] Item do tipo `INSUMO` retornando `422`
- [ ] Nenhum saldo de estoque alterado após a criação do pedido
- [ ] `DELETE` em pedido `ABERTO` retornando `204`
- [ ] `DELETE` em pedido `PARCIAL` retornando `409`

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI

**Review**

- [ ] Code Review aprovado

---
