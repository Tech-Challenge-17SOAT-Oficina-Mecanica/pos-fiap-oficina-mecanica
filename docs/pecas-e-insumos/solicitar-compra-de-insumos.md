---
documento: Refinamento de Requisitos — Solicitar Compra de Insumos
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Solicitar Compra de Insumos

Este documento detalha a tarefa Solicitar Compra de Insumos do contexto de Peças & Insumos.

## 9 · Solicitar Compra de Insumos

### 9.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar um pedido de compra dos insumos necessários às Ordens de Serviço, reservando as
quantidades compradas para essas OS e colocando-as em espera formal por recursos.

**Problema**

Quando um insumo necessário a uma OS não tem saldo, a compra vive fora do sistema: ninguém sabe o
que já foi solicitado, para qual OS aquilo foi comprado e por que o veículo está parado. Sem
pedido formal, o mesmo insumo é solicitado duas vezes e o que chega acaba consumido por outro
serviço.

**Pré-condições**

- Os insumos devem estar cadastrados e ativos.
- Deve existir necessidade de insumos registrada nas Ordens de Serviço.
- O usuário deve estar autorizado a solicitar compra.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-110 | Permitir criar pedido de compra informando insumos e quantidades. |
| RF-EST-111 | Permitir informar o fornecedor de forma opcional. |
| RF-EST-112 | Permitir gerar o pedido a partir da consulta de itens faltantes, com as quantidades necessárias já preenchidas. |
| RF-EST-113 | Validar que a quantidade informada respeita a unidade de medida do insumo. |
| RF-EST-114 | Validar que a quantidade comprada de cada insumo é igual à quantidade necessária apurada nas Ordens de Serviço. |
| RF-EST-115 | Vincular o pedido às OS que dependem dos insumos solicitados. |
| RF-EST-116 | Reservar integralmente os insumos comprados para as OS vinculadas. |
| RF-EST-117 | Atualizar cada OS vinculada com os insumos necessários e suas quantidades reservadas. |
| RF-EST-118 | Alterar o status das OS vinculadas para `AGUARDANDO_RECURSOS`. |
| RF-EST-119 | Permitir cancelar pedido ainda não recebido e registrar a situação do pedido: aberto, parcialmente recebido, concluído ou cancelado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-81 | A operação deve ser feita por API RESTful. |
| RNF-EST-82 | A operação deve ser acessível somente por usuário autorizado com permissão de compras. |
| RNF-EST-83 | O pedido de compra não altera o saldo físico de estoque — o saldo só muda no registro da entrada. |
| RNF-EST-84 | A reserva criada pelo pedido não consome saldo físico: ela compromete a quantidade comprada para a OS de destino. |
| RNF-EST-85 | O pedido deve ser auditável, com registro de quem solicitou e quando. |
| RNF-EST-86 | O sistema deve alertar quando já existir pedido em aberto para o mesmo insumo. |
| RNF-EST-87 | A criação do pedido, as reservas e a mudança de status das OS devem ocorrer na mesma operação. |

**Fluxo Principal**

1. O mecânico consulta os insumos faltantes das Ordens de Serviço.
2. O mecânico seleciona os insumos a serem comprados e confirma as quantidades.
3. O mecânico informa o fornecedor, quando houver.
4. O sistema valida os insumos, as quantidades e as unidades de medida.
5. O sistema apura, nas Ordens de Serviço, a quantidade necessária de cada insumo e confere se ela
   é igual à quantidade comprada.
6. O sistema verifica pedidos em aberto para os mesmos insumos.
7. O sistema cria o pedido de compra com situação "aberto".
8. O sistema vincula o pedido às OS que dependem desses insumos.
9. O sistema reserva integralmente os insumos comprados para as OS vinculadas.
10. O sistema atualiza cada OS vinculada com os insumos necessários e as quantidades reservadas.
11. O sistema altera o status das OS vinculadas para `AGUARDANDO_RECURSOS`.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Insumo não encontrado ou inativo | Impede a inclusão do item no pedido. |
| A2 | Quantidade incompatível com a unidade de medida | Impede a criação e informa a unidade esperada. |
| A3 | Quantidade comprada diferente da quantidade necessária apurada nas OS | Impede a criação do pedido. |
| A4 | Insumo sem necessidade registrada em nenhuma OS | Impede a inclusão do item no pedido. |
| A5 | Já existe pedido em aberto para o mesmo insumo | Alerta e exige confirmação. |
| A6 | Fornecedor informado e não cadastrado | Impede a criação e permite seguir para o cadastro do fornecedor. |
| A7 | Pedido cancelado após criação | Atualiza a situação, libera as reservas, desvincula as OS e as retorna ao status anterior, sinalizando que a falta permanece. |
| A8 | Usuário sem autorização | Impede a operação. |

**Saída**

- Pedido de compra criado, com número, fornecedor quando informado, itens, quantidades, unidades
  de medida, reservas geradas e OS vinculadas; **ou** indicação do motivo pelo qual o pedido foi
  recusado.

**Pós-condições**

- O pedido de compra está registrado com situação "aberto".
- O saldo físico de estoque permanece inalterado.
- Todos os insumos comprados estão reservados para as OS vinculadas.
- As OS vinculadas registram os insumos necessários com as respectivas quantidades reservadas.
- As OS vinculadas estão com status `AGUARDANDO_RECURSOS`.
- O pedido fica disponível para vinculação no registro de entrada de estoque.

---

### 9.2 Refinamento Técnico

**Endpoint**

```http
POST   /compras/pedidos
DELETE /compras/pedidos/{pedidoId}
```

> **Decisão de projeto.** Peça e insumo usam o mesmo recurso `pedido_compra`, porque o ciclo de
> vida é idêntico (`ABERTO` → `PARCIAL` → `CONCLUIDO`), a reserva é a mesma e o recebimento é o
> mesmo; o tipo do item diferencia apenas a validação de unidade de medida. A alternativa
> `POST /compras/pedidos/insumos` foi descartada por duplicar a lógica de reserva e recebimento.
> Nesta tarefa o **fornecedor é opcional**, porque a solicitação pode ser feita por canal externo,
> e não há cálculo de sugestão por estoque mínimo ou consumo médio: a quantidade vem da
> necessidade registrada nas OS.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `compras:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `fornecedorId` | uuid | Opcional; quando informado, deve existir e estar ativo. |
| Body | `confirmarDuplicidade` | boolean | Obrigatório quando já houver pedido em aberto para o mesmo insumo. |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido. |
| Body | `itens[].itemId` | uuid | Insumo a comprar; item do tipo `PECA` é rejeitado. |
| Body | `itens[].quantidade` | decimal | Maior que zero, compatível com a `unidadeMedida` e igual à necessidade apurada nas OS. |
| Path (DELETE) | `pedidoId` | uuid | Pedido a cancelar. |

```json
{
  "fornecedorId": "d94c6b13-7f20-4a58-8e93-15b7c2a60df4",
  "confirmarDuplicidade": false,
  "itens": [
    { "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48", "quantidade": 60.0 }
  ]
}
```

O `fornecedorId` pode ser omitido ou enviado como `null` quando a solicitação for feita por canal
externo.

**Validações**

*Técnicas*

- `fornecedorId` opcional; quando informado, deve existir e estar ativo.
- `itens` não vazio, sem repetição de `itemId`.
- `quantidade` maior que zero, com casas decimais compatíveis com a `unidadeMedida` do insumo, por
  exemplo `UN` não aceita fração.

*Negócio*

- Todos os itens existem, estão ativos e são do tipo `INSUMO`.
- Cada item deve possuir necessidade registrada em pelo menos uma OS.
- A `quantidade` deve ser igual à quantidade necessária apurada nas OS para aquele insumo.
- Existindo pedido `ABERTO` ou `PARCIAL` para o mesmo insumo, exigir `confirmarDuplicidade: true`.
- A reserva gerada é integral: quantidade reservada igual à quantidade comprada.
- Cancelamento permitido apenas em pedido sem recebimento, com status `ABERTO`.

**Processamento**

1. Validar o payload.
2. Carregar o fornecedor, quando `fornecedorId` for informado.
3. Carregar e validar os itens, conferindo tipo, situação e unidade de medida.
4. Apurar, no módulo de OS, a quantidade necessária de cada insumo e as OS demandantes.
5. Validar que a quantidade comprada é igual à quantidade necessária apurada.
6. Buscar pedidos em aberto para os mesmos insumos; havendo, e com `confirmarDuplicidade = false`,
   retornar `409` com a lista de pedidos conflitantes.
7. Gerar o número sequencial do pedido.
8. Persistir `pedido_compra` com status `ABERTO` e seus `pedido_compra_item`.
9. Criar as reservas dos insumos comprados para as OS demandantes.
10. Vincular o pedido às OS e atualizar cada OS com os insumos necessários e as quantidades reservadas.
11. Alterar o status das OS vinculadas para `AGUARDANDO_RECURSOS`.
12. Publicar `PedidoCompraCriado`.

**Persistência**

- Consulta: `item_estoque`, `fornecedor` (quando informado), `pedido_compra`, módulo de OS.
- Altera: `pedido_compra` (insert), `pedido_compra_item` (insert), `reserva_estoque` (insert),
  OS vinculadas (insumos necessários e status).
- Não altera: nenhum saldo físico de estoque — o saldo só muda no registro da entrada.

**Saída da API**

```json
{
  "pedidoId": "6b2e9f47-3a15-4c80-9d62-7e10b4f83a95",
  "numero": "2026/0121",
  "fornecedor": {
    "id": "d94c6b13-7f20-4a58-8e93-15b7c2a60df4",
    "nome": "Distribuidora Norte"
  },
  "status": "ABERTO",
  "criadoEm": "2026-08-12T17:40:00-03:00",
  "criadoPor": "4c1d8e62-9b07-4a53-8f16-2d7e5a90c3b1",
  "itens": [
    {
      "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
      "codigo": "IN-0031",
      "descricao": "Óleo lubrificante 15W40",
      "unidadeMedida": "L",
      "quantidadeNecessaria": 60.0,
      "quantidadePedida": 60.0,
      "quantidadeReservada": 60.0,
      "quantidadeRecebida": 0
    }
  ],
  "ordensServicoVinculadas": [
    {
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "status": "AGUARDANDO_RECURSOS",
      "insumosNecessarios": [
        {
          "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
          "quantidade": 60.0,
          "quantidadeReservada": 60.0
        }
      ]
    }
  ]
}
```

Quando o pedido for criado sem fornecedor, o campo `fornecedor` retorna `null`.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Pedido criado, insumos reservados e OS atualizadas para `AGUARDANDO_RECURSOS`. |
| `204` | Pedido cancelado (`DELETE`). |
| `400` | Body inválido, item repetido, quantidade menor ou igual a zero, ou decimal incompatível com a unidade de medida. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `compras:escrever`. |
| `404` | Fornecedor informado, insumo ou pedido não encontrado. |
| `409` | Pedido em aberto para o mesmo insumo sem `confirmarDuplicidade`; tentativa de cancelar pedido com recebimento parcial. |
| `422` | Item inativo ou do tipo `PECA`; quantidade comprada diferente da necessária apurada; item sem necessidade registrada em nenhuma OS. |

**Dependências**

- `PedidoCompraRepository`.
- `ItemEstoqueRepository`.
- `FornecedorRepository`, de uso condicional, apenas quando houver fornecedor.
- `ReservaEstoqueRepository`, a mesma reserva usada na compra de peças.
- Módulo Ordem de Serviço — apuração da necessidade, vinculação, atualização dos insumos
  necessários e mudança de status.
- Publicador de eventos de domínio.
- Casos de uso Consultar Peças Faltantes (origem) e Registrar Entrada de Estoque (destino).

**Testes**

*Unitários*

- Rejeita quantidade zero.
- Rejeita fração em insumo com unidade `UN` e aceita decimal em insumo com unidade `L`.
- Rejeita quantidade divergente da quantidade necessária apurada.
- Detecta pedido em aberto para o mesmo insumo.
- Reserva gerada igual à quantidade comprada.
- Cálculo do número sequencial.

*Integração*

- Pedido válido retorna `201` com status `ABERTO`.
- Pedido sem `fornecedorId` retorna `201` com fornecedor nulo.
- Fornecedor informado e inexistente retorna `404`.
- Pedido válido reserva integralmente os insumos comprados.
- Pedido válido atualiza as OS vinculadas com os insumos necessários.
- Pedido válido altera as OS vinculadas para `AGUARDANDO_RECURSOS`.
- Duplicidade sem confirmação retorna `409`; com `confirmarDuplicidade: true`, retorna `201`.
- Item do tipo `PECA` retorna `422`; quantidade divergente da necessidade retorna `422`.
- Criar pedido não altera nenhum saldo físico de estoque.
- `DELETE` em pedido `ABERTO` retorna `204` e libera as reservas.
- `DELETE` em pedido `PARCIAL` retorna `409`.

---

### 9.3 Checklist de Implementação

**Domínio**

- [ ] Reaproveitar a entidade `PedidoCompra` da compra de peças, sem duplicar o ciclo de vida
- [ ] Reaproveitar `PedidoCompraItem` com `quantidadeNecessaria`, `quantidadePedida`, `quantidadeReservada` e `quantidadeRecebida`
- [ ] Reaproveitar a reserva vinculando item, quantidade, pedido e Ordem de Serviço
- [ ] Tornar o fornecedor opcional no pedido de compra
- [ ] Implementar a validação de casas decimais conforme a unidade de medida do insumo
- [ ] Implementar a regra de quantidade comprada igual à quantidade necessária apurada nas OS
- [ ] Implementar a regra de reserva integral dos insumos comprados
- [ ] Implementar a regra de cancelamento permitido apenas em pedido sem recebimento
- [ ] Implementar a liberação das reservas no cancelamento
- [ ] Reaproveitar a geração de número sequencial do pedido
- [ ] Garantir que criar pedido não altera nenhum saldo físico de estoque

**Caso de uso**

- [ ] Implementar `SolicitarCompraDeInsumos`
- [ ] Reaproveitar `CancelarPedidoCompra`
- [ ] Implementar a apuração da quantidade necessária de cada insumo nas OS
- [ ] Implementar a vinculação das OS que dependem dos insumos solicitados
- [ ] Implementar a reserva dos insumos comprados para as OS vinculadas
- [ ] Implementar a atualização das OS com os insumos necessários e quantidades reservadas
- [ ] Implementar a transição das OS vinculadas para `AGUARDANDO_RECURSOS`
- [ ] Garantir que pedido, reservas e mudança de status das OS sejam persistidos na mesma operação

**Repositório**

- [ ] Reaproveitar `PedidoCompraRepository` e o repositório de reservas
- [ ] Reaproveitar a consulta de pedidos em aberto para os mesmos itens
- [ ] Integrar com `ItemEstoqueRepository`
- [ ] Integrar com `FornecedorRepository` apenas quando houver fornecedor informado

**Handler HTTP**

- [ ] Reaproveitar `POST /compras/pedidos` com validação por tipo de item
- [ ] Reaproveitar `DELETE /compras/pedidos/{pedidoId}`
- [ ] Criar DTO/request de entrada com `fornecedorId` opcional
- [ ] Criar DTO/response de saída com fornecedor anulável
- [ ] Validar o parâmetro `pedidoId`
- [ ] Aplicar autenticação e autorização por escopo nas rotas
- [ ] Mapear os erros para os códigos HTTP definidos

**Validações**

- [ ] Validar fornecedor existente somente quando informado
- [ ] Validar decimais compatíveis com a unidade de medida do insumo
- [ ] Validar quantidade maior que zero e sem repetição de item
- [ ] Rejeitar item do tipo `PECA` neste caso de uso
- [ ] Rejeitar item sem necessidade registrada em nenhuma OS
- [ ] Rejeitar quantidade divergente da quantidade necessária apurada
- [ ] Exigir `confirmarDuplicidade` quando já houver pedido em aberto para o mesmo insumo

**Eventos**

- [ ] Publicar `PedidoCompraCriado`
- [ ] Reaproveitar a política de atualização das OS vinculadas para `AGUARDANDO_RECURSOS`

**Testes unitários**

- [ ] Rejeição de quantidade zero
- [ ] Rejeição de fração em insumo com unidade `UN`
- [ ] Aceite de decimal em insumo com unidade `L`
- [ ] Rejeição de quantidade divergente da necessidade apurada
- [ ] Detecção de pedido em aberto para o mesmo insumo
- [ ] Reserva integral dos insumos comprados
- [ ] Geração do número sequencial

**Testes de integração**

- [ ] Pedido válido retornando `201` com status `ABERTO`
- [ ] Pedido sem `fornecedorId` retornando `201` com fornecedor nulo
- [ ] Fornecedor informado e inexistente retornando `404`
- [ ] Insumos comprados totalmente reservados após a criação do pedido
- [ ] OS vinculadas atualizadas com os insumos necessários e quantidades reservadas
- [ ] OS vinculadas com status `AGUARDANDO_RECURSOS`
- [ ] Duplicidade sem confirmação retornando `409` e com confirmação retornando `201`
- [ ] Item do tipo `PECA` retornando `422`
- [ ] Quantidade divergente da necessidade retornando `422`
- [ ] Nenhum saldo físico de estoque alterado após a criação do pedido
- [ ] `DELETE` em pedido `ABERTO` retornando `204` e liberando as reservas
- [ ] `DELETE` em pedido `PARCIAL` retornando `409`

**Documentação**

- [ ] Documentar os endpoints no Swagger/OpenAPI, incluindo a decisão de rota compartilhada com a compra de peças e o fornecedor opcional

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite da task

---
