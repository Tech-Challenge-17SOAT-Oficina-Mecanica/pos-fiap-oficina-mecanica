---
documento: Refinamento de Requisitos — Solicitar Compra de Peças
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Solicitar Compra de Peças

Este documento detalha a tarefa Solicitar Compra de Peças do contexto de Peças & Insumos.

## 8 · Solicitar Compra de Peças

### 8.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar um pedido de compra de peças junto ao fornecedor, reservando as peças para as Ordens de
Serviço que dependem delas e colocando essas OS em espera formal por recursos.

**Problema**

Sem pedido formal, a compra vive na conversa de WhatsApp com o fornecedor: ninguém sabe o que já
foi pedido, o que está a caminho e para qual OS aquilo foi comprado. Isso gera compra duplicada,
peça comprada sendo consumida por outra OS, e impede saber por que um veículo está parado.

**Pré-condições**

- As peças devem estar cadastradas e ativas.
- Deve existir fornecedor cadastrado.
- Deve existir necessidade de peças registrada nas Ordens de Serviço.
- O usuário deve estar autorizado a solicitar compra.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-100 | Permitir criar pedido de compra informando fornecedor, peças e quantidades. |
| RF-EST-101 | Permitir gerar o pedido a partir da consulta de peças faltantes, com as quantidades sugeridas já preenchidas. |
| RF-EST-102 | Validar que as quantidades informadas são maiores que zero. |
| RF-EST-103 | Validar que a quantidade comprada de cada peça é igual à quantidade necessária apurada nas Ordens de Serviço. |
| RF-EST-104 | Vincular o pedido às OS que dependem das peças solicitadas. |
| RF-EST-105 | Reservar integralmente as peças compradas para as OS vinculadas. |
| RF-EST-106 | Atualizar cada OS vinculada com as peças necessárias e suas quantidades reservadas. |
| RF-EST-107 | Alterar o status das OS vinculadas para `AGUARDANDO_RECURSOS`. |
| RF-EST-108 | Permitir cancelar pedido ainda não recebido. |
| RF-EST-109 | Registrar a situação do pedido: aberto, parcialmente recebido, concluído ou cancelado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-74 | A operação deve ser feita por API RESTful. |
| RNF-EST-75 | A operação deve ser acessível somente por usuário autorizado com permissão de compras. |
| RNF-EST-76 | O pedido de compra não altera o saldo físico de estoque — o saldo só muda no registro da entrada. |
| RNF-EST-77 | A reserva criada pelo pedido não consome saldo físico: ela compromete a quantidade comprada para a OS de destino. |
| RNF-EST-78 | O pedido deve ser auditável, com registro de quem solicitou e quando. |
| RNF-EST-79 | O sistema deve alertar quando já existir pedido em aberto para a mesma peça, evitando compra duplicada. |
| RNF-EST-80 | A criação do pedido, as reservas e a mudança de status das OS devem ocorrer na mesma operação. |

**Fluxo Principal**

1. O mecânico seleciona as peças a serem compradas.
2. O mecânico informa o fornecedor e as quantidades.
3. O sistema valida as peças e as quantidades informadas.
4. O sistema apura, nas Ordens de Serviço, a quantidade necessária de cada peça e confere se ela é
   igual à quantidade comprada.
5. O sistema verifica pedidos em aberto para as mesmas peças.
6. O sistema cria o pedido de compra com situação "aberto".
7. O sistema vincula o pedido às OS que dependem dessas peças.
8. O sistema reserva integralmente as peças compradas para as OS vinculadas.
9. O sistema atualiza cada OS vinculada com as peças necessárias e as quantidades reservadas.
10. O sistema altera o status das OS vinculadas para `AGUARDANDO_RECURSOS`.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Peça não encontrada ou inativa | Impede a inclusão do item no pedido. |
| A2 | Quantidade menor ou igual a zero | Impede a criação do pedido. |
| A3 | Quantidade comprada diferente da quantidade necessária apurada nas OS | Impede a criação do pedido. |
| A4 | Peça sem necessidade registrada em nenhuma OS | Impede a inclusão do item no pedido. |
| A5 | Já existe pedido em aberto para a mesma peça | Alerta e exige confirmação antes de criar um novo pedido. |
| A6 | Fornecedor não cadastrado | Impede a criação e permite seguir para o cadastro do fornecedor. |
| A7 | Pedido cancelado após criação | Atualiza a situação, libera as reservas, desvincula as OS e as retorna ao status anterior, sinalizando que a falta permanece. |
| A8 | Usuário sem autorização | Impede a operação. |

**Saída**

- Pedido de compra criado, com número, fornecedor, itens, quantidades, reservas geradas e OS
  vinculadas; **ou** indicação do motivo pelo qual o pedido foi recusado.

**Pós-condições**

- O pedido de compra está registrado com situação "aberto".
- O saldo físico de estoque permanece inalterado.
- Todas as peças compradas estão reservadas para as OS vinculadas.
- As OS vinculadas registram as peças necessárias com as respectivas quantidades reservadas.
- As OS vinculadas estão com status `AGUARDANDO_RECURSOS`.
- O pedido fica disponível para vinculação no registro de entrada de estoque.

---

### 8.2 Refinamento Técnico

**Endpoint**

```http
POST   /compras/pedidos
DELETE /compras/pedidos/{pedidoId}
```

> **Decisão de projeto.** Peça e insumo usam o mesmo recurso `pedido_compra`, porque o ciclo de
> vida é idêntico (`ABERTO` → `PARCIAL` → `CONCLUIDO`), a reserva é a mesma e o recebimento é o
> mesmo; o tipo do item diferencia apenas a validação de unidade de medida. A alternativa
> `POST /compras/pedidos/pecas` foi descartada por duplicar a lógica de reserva e de recebimento.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `compras:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `fornecedorId` | uuid | Obrigatório e existente. |
| Body | `confirmarDuplicidade` | boolean | Obrigatório quando já houver pedido em aberto para a mesma peça. |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido. |
| Body | `itens[].itemId` | uuid | Peça a comprar; item do tipo `INSUMO` é rejeitado. |
| Body | `itens[].quantidade` | int | Inteiro maior que zero, igual à necessidade apurada nas OS. |
| Path (DELETE) | `pedidoId` | uuid | Pedido a cancelar. |

```json
{
  "fornecedorId": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
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
- `itens` não vazio, sem repetição de `itemId`.
- `quantidade` inteira maior que zero.

*Negócio*

- Todos os itens existem, estão ativos e são do tipo `PECA`.
- Cada item deve possuir necessidade registrada em pelo menos uma OS.
- A `quantidade` deve ser igual à quantidade necessária apurada nas OS para aquela peça.
- Existindo pedido `ABERTO` ou `PARCIAL` para a mesma peça, exigir `confirmarDuplicidade: true`.
- A reserva gerada é integral: quantidade reservada igual à quantidade comprada.
- Cancelamento permitido apenas em pedido sem recebimento, com status `ABERTO`.

**Processamento**

1. Validar o payload e carregar o fornecedor.
2. Carregar e validar os itens.
3. Apurar, no módulo de OS, a quantidade necessária de cada peça e as OS demandantes.
4. Validar que a quantidade comprada é igual à quantidade necessária apurada.
5. Buscar pedidos em aberto para as mesmas peças; havendo, e com `confirmarDuplicidade = false`,
   retornar `409` com a lista de pedidos conflitantes.
6. Gerar o número sequencial do pedido.
7. Persistir `pedido_compra` com status `ABERTO` e seus `pedido_compra_item`.
8. Criar as reservas das peças compradas para as OS demandantes.
9. Vincular o pedido às OS e atualizar cada OS com as peças necessárias e as quantidades reservadas.
10. Alterar o status das OS vinculadas para `AGUARDANDO_RECURSOS`.
11. Publicar `PedidoCompraCriado`.

**Persistência**

- Consulta: `item_estoque`, `fornecedor`, `pedido_compra`, módulo de OS.
- Altera: `pedido_compra` (insert), `pedido_compra_item` (insert), `reserva_estoque` (insert),
  OS vinculadas (peças necessárias e status).
- Não altera: nenhum saldo físico de estoque — o saldo só muda no registro da entrada.

**Saída da API**

```json
{
  "pedidoId": "f05a1d63-8b47-49e2-a731-6c94d2e08b57",
  "numero": "2026/0118",
  "fornecedor": {
    "id": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
    "nome": "Auto Peças Recife"
  },
  "status": "ABERTO",
  "criadoEm": "2026-08-12T17:05:00-03:00",
  "criadoPor": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29",
  "itens": [
    {
      "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90",
      "codigo": "PC-0311",
      "descricao": "Disco de freio ventilado",
      "quantidadeNecessaria": 5,
      "quantidadePedida": 5,
      "quantidadeReservada": 5,
      "quantidadeRecebida": 0
    },
    {
      "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "codigo": "PC-0142",
      "descricao": "Pastilha de freio dianteira",
      "quantidadeNecessaria": 10,
      "quantidadePedida": 10,
      "quantidadeReservada": 10,
      "quantidadeRecebida": 0
    }
  ],
  "ordensServicoVinculadas": [
    {
      "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
      "status": "AGUARDANDO_RECURSOS",
      "pecasNecessarias": [
        {
          "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90",
          "quantidade": 5,
          "quantidadeReservada": 5
        },
        {
          "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
          "quantidade": 10,
          "quantidadeReservada": 10
        }
      ]
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Pedido criado, peças reservadas e OS atualizadas para `AGUARDANDO_RECURSOS`. |
| `204` | Pedido cancelado (`DELETE`). |
| `400` | Body inválido, item repetido ou quantidade menor ou igual a zero. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `compras:escrever`. |
| `404` | Fornecedor, item ou pedido não encontrado. |
| `409` | Pedido em aberto para a mesma peça sem `confirmarDuplicidade`; tentativa de cancelar pedido com recebimento parcial. |
| `422` | Item inativo ou do tipo `INSUMO`; quantidade comprada diferente da necessária apurada; item sem necessidade registrada em nenhuma OS. |

**Dependências**

- `PedidoCompraRepository`.
- `ItemEstoqueRepository`.
- `FornecedorRepository`.
- `ReservaEstoqueRepository`.
- Módulo Ordem de Serviço — apuração da necessidade, vinculação, atualização das peças necessárias
  e mudança de status.
- Publicador de eventos de domínio.
- Casos de uso Consultar Peças Faltantes (origem) e Registrar Entrada de Estoque (destino).

**Testes**

*Unitários*

- Rejeita quantidade zero.
- Rejeita quantidade divergente da quantidade necessária apurada.
- Detecta pedido em aberto para a mesma peça.
- Reserva gerada igual à quantidade comprada.
- Cálculo do número sequencial.

*Integração*

- Pedido válido retorna `201` com status `ABERTO`.
- Pedido válido reserva integralmente as peças compradas.
- Pedido válido atualiza as OS vinculadas com as peças necessárias.
- Pedido válido altera as OS vinculadas para `AGUARDANDO_RECURSOS`.
- Duplicidade sem confirmação retorna `409`; com `confirmarDuplicidade: true`, retorna `201`.
- Item do tipo `INSUMO` retorna `422`; quantidade divergente da necessidade retorna `422`.
- Criar pedido não altera nenhum saldo físico de estoque.
- `DELETE` em pedido `ABERTO` retorna `204` e libera as reservas.
- `DELETE` em pedido `PARCIAL` retorna `409`.

---

### 8.3 Checklist de Implementação

**Domínio**

- [ ] Implementar a entidade `PedidoCompra` com status `ABERTO`, `PARCIAL`, `CONCLUIDO` e `CANCELADO`
- [ ] Implementar `PedidoCompraItem` com `quantidadeNecessaria`, `quantidadePedida`, `quantidadeReservada` e `quantidadeRecebida`
- [ ] Implementar a reserva vinculando item, quantidade, pedido e Ordem de Serviço
- [ ] Implementar a regra de quantidade comprada igual à quantidade necessária apurada nas OS
- [ ] Implementar a regra de reserva integral das peças compradas
- [ ] Implementar a regra de cancelamento permitido apenas em pedido sem recebimento
- [ ] Implementar a liberação das reservas no cancelamento
- [ ] Implementar a geração de número sequencial do pedido
- [ ] Garantir que criar pedido não altera nenhum saldo físico de estoque

**Caso de uso**

- [ ] Implementar `SolicitarCompraDePecas`
- [ ] Implementar `CancelarPedidoCompra`
- [ ] Implementar a apuração da quantidade necessária de cada peça nas OS
- [ ] Implementar a vinculação das OS que dependem das peças solicitadas
- [ ] Implementar a reserva das peças compradas para as OS vinculadas
- [ ] Implementar a atualização das OS com as peças necessárias e quantidades reservadas
- [ ] Implementar a transição das OS vinculadas para `AGUARDANDO_RECURSOS`
- [ ] Garantir que pedido, reservas e mudança de status das OS sejam persistidos na mesma operação

**Repositório**

- [ ] Implementar `PedidoCompraRepository`
- [ ] Implementar a persistência das reservas
- [ ] Implementar a consulta de pedidos em aberto para as mesmas peças
- [ ] Integrar com `FornecedorRepository` e `ItemEstoqueRepository`

**Handler HTTP**

- [ ] Implementar `POST /compras/pedidos`
- [ ] Implementar `DELETE /compras/pedidos/{pedidoId}`
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Validar o parâmetro `pedidoId`
- [ ] Aplicar autenticação e autorização na rota
- [ ] Mapear os erros para os códigos HTTP definidos

**Validações**

- [ ] Validar fornecedor existente
- [ ] Validar quantidade inteira maior que zero e sem repetição de item
- [ ] Rejeitar item do tipo `INSUMO` neste caso de uso
- [ ] Rejeitar item sem necessidade registrada em nenhuma OS
- [ ] Rejeitar quantidade divergente da quantidade necessária apurada
- [ ] Exigir `confirmarDuplicidade` quando já houver pedido em aberto para a mesma peça

**Eventos**

- [ ] Publicar `PedidoCompraCriado`
- [ ] Implementar a política de atualização das OS vinculadas para `AGUARDANDO_RECURSOS`

**Testes unitários**

- [ ] Rejeição de quantidade zero
- [ ] Rejeição de quantidade divergente da necessidade apurada
- [ ] Detecção de pedido em aberto para a mesma peça
- [ ] Reserva integral das peças compradas
- [ ] Geração do número sequencial

**Testes de integração**

- [ ] Pedido válido retornando `201` com status `ABERTO`
- [ ] Peças compradas totalmente reservadas após a criação do pedido
- [ ] OS vinculadas atualizadas com as peças necessárias e quantidades reservadas
- [ ] OS vinculadas com status `AGUARDANDO_RECURSOS`
- [ ] Duplicidade sem confirmação retornando `409` e com confirmação retornando `201`
- [ ] Item do tipo `INSUMO` retornando `422`
- [ ] Quantidade divergente da necessidade retornando `422`
- [ ] Nenhum saldo físico de estoque alterado após a criação do pedido
- [ ] `DELETE` em pedido `ABERTO` retornando `204` e liberando as reservas
- [ ] `DELETE` em pedido `PARCIAL` retornando `409`

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite da task

---
