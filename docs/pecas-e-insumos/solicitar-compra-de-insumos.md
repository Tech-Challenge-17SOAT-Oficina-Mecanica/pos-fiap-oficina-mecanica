---
documento: Refinamento de Requisitos — Solicitar Compra de Insumos
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Solicitar Compra de Insumos

Este documento detalha a tarefa Solicitar Compra de Insumos do contexto de Peças & Insumos.

## 9 · Solicitar Compra de Insumos

### 9.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Registrar um pedido de compra de insumos junto ao fornecedor, mantendo o abastecimento
contínuo dos materiais de consumo da oficina.

**Problema**
Insumo não trava uma OS específica, então a falta passa despercebida até o dia em que nenhum
serviço pode ser executado. Como o consumo é contínuo e não vinculado a uma OS, a reposição
precisa ser guiada por estoque mínimo e média de consumo, não por demanda pontual.

**Pré-condições**

- Os insumos devem estar cadastrados e ativos.
- Deve existir fornecedor cadastrado.
- O usuário deve estar autorizado a solicitar compra.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-52 | Permitir criar pedido de compra informando fornecedor, insumos, quantidades e prazo previsto de entrega. |
| RF-EST-53 | Permitir gerar o pedido a partir da consulta de itens faltantes. |
| RF-EST-54 | Validar que a quantidade informada respeita a unidade de medida do insumo. |
| RF-EST-55 | Permitir cancelar pedido ainda não recebido. |
| RF-EST-56 | Registrar a situação do pedido: aberto, parcialmente recebido, concluído ou cancelado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-43 | A operação deve ser acessível somente por usuário autorizado com permissão de compras. |
| RNF-EST-44 | O pedido de compra não altera o saldo de estoque — o saldo só muda no registro da entrada. |
| RNF-EST-45 | O pedido deve ser auditável, com registro de quem solicitou e quando. |
| RNF-EST-46 | O sistema deve alertar quando já existir pedido em aberto para o mesmo insumo. |

**Fluxo Principal**

1. O mecânico seleciona os insumos a serem comprados.
2. O mecânico informa o fornecedor, confirma as quantidades e o prazo previsto de entrega.
3. O sistema valida os insumos, as quantidades e as unidades de medida.
4. O sistema verifica pedidos em aberto para os mesmos insumos.
5. O sistema cria o pedido de compra com situação "aberto".

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Insumo não encontrado ou inativo | Impede a inclusão do item no pedido. |
| A2 | Quantidade incompatível com a unidade de medida | Impede a criação e informa a unidade esperada. |
| A3 | Já existe pedido em aberto para o mesmo insumo | Alerta e exige confirmação. |
| A4 | Fornecedor não cadastrado | Impede a criação e permite seguir para o cadastro do fornecedor. |
| A5 | Usuário sem autorização | Impede a operação. |

**Saída**

- Pedido de compra criado, com número, fornecedor, itens, quantidades, unidades de medida e prazo previsto; **ou**
- Indicação do motivo pelo qual o pedido foi recusado.

**Pós-condições**

- O pedido de compra está registrado com situação "aberto".
- O saldo de estoque permanece inalterado.
- O pedido fica disponível para vinculação no registro de entrada de estoque.

---

### 9.2 Refinamento Técnico

**Endpoint**

```http
POST /compras/pedidos
GET  /estoque/insumos/{insumoId}/sugestao-compra
```

O `GET` auxiliar calcula a quantidade sugerida antes de montar o pedido.

> **Decisão de projeto.** Peça e insumo usam o mesmo recurso `pedido_compra` porque o ciclo de
> vida é idêntico (`ABERTO` → `PARCIAL` → `CONCLUIDO`) e o recebimento é o mesmo; o tipo do item
> diferencia as regras. Se o time preferir rotas separadas para espelhar os dois requisitos,
> `POST /compras/pedidos/insumos` é aceitável — mas duplica a lógica de recebimento.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `compras:escrever` (leitura da sugestão: `estoque:ler`)

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Path (GET) | `insumoId` | uuid   | Insumo para o qual a sugestão é calculada |
| Body | `fornecedorId` | uuid   | Obrigatório e existente |
| Body | `prazoPrevistoEntrega` | date | Obrigatório, no futuro |
| Body | `confirmarDuplicidade` | boolean | Obrigatório quando já houver pedido em aberto para o mesmo insumo |
| Body | `itens[]` | array | Obrigatório, não vazio |
| Body | `itens[].itemId` | uuid   | Insumo a comprar; item do tipo `PECA` é rejeitado |
| Body | `itens[].quantidade` | decimal | Maior que zero, com casas decimais compatíveis com a `unidadeMedida` |

```json
{
  "fornecedorId": "d94c6b13-7f20-4a58-8e93-15b7c2a60df4",
  "prazoPrevistoEntrega": "2026-08-19",
  "confirmarDuplicidade": false,
  "itens": [{ "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48", "quantidade": 60.0 }]
}
```

**Validações**

*Técnicas*

- `quantidade` maior que zero, com casas decimais compatíveis com a `unidadeMedida` do insumo (por exemplo, `UN` não aceita fração).
- `prazoPrevistoEntrega` no futuro.

*Negócio*

- Todos os itens são do tipo `INSUMO` e estão ativos.
- Pedido em aberto para o mesmo insumo exige `confirmarDuplicidade: true`.
- Sem histórico de consumo suficiente, o endpoint de sugestão retorna vazio e a quantidade passa a ser obrigatória no body.

**Processamento**

*Sugestão (GET)*

1. Calcular o consumo médio a partir das `movimentacao_estoque` do tipo `SAIDA` nos últimos 90 dias.
2. `quantidadeSugerida = (consumoMedioDiario × leadTimeFornecedor) + estoqueMinimo − saldoDisponivel`.
3. Arredondar para cima conforme a `unidadeMedida`.
4. Menos de 30 dias de histórico: retornar `sugestaoDisponivel: false`.

*Criação do pedido (POST)*

Mesmo fluxo do requisito 8, sem a etapa de vinculação de OS — insumo não trava OS específica.

**Persistência**

- Consulta: `item_estoque`, `movimentacao_estoque` (cálculo de consumo médio), `fornecedor`, `pedido_compra`
- Altera: `pedido_compra` (insert), `pedido_compra_item` (insert)
- Não altera: nenhum saldo de estoque

**Saída da API**

Sugestão (`GET`):

```json
{
  "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
  "codigo": "IN-0031",
  "descricao": "Óleo lubrificante 15W40",
  "unidadeMedida": "L",
  "saldoDisponivel": 44.0,
  "estoqueMinimo": 20,
  "consumoMedioDiario": 1.8,
  "leadTimeDias": 7,
  "quantidadeSugerida": 60.0,
  "sugestaoDisponivel": true,
  "baseHistorico": { "diasAnalisados": 90, "totalConsumido": 162.0 }
}
```

Pedido criado (`POST`): mesma estrutura do requisito 8, com `ordensServicoVinculadas: []`.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Pedido criado |
| `200` | Sugestão calculada — pode vir com `sugestaoDisponivel: false` |
| `400` | Quantidade com decimal incompatível com a unidade de medida |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo necessário |
| `404` | Fornecedor ou insumo não encontrado |
| `409` | Pedido em aberto para o mesmo insumo sem `confirmarDuplicidade` |
| `422` | Item inativo ou do tipo `PECA` |

**Dependências**

- `PedidoCompraRepository`
- `ItemEstoqueRepository`
- `MovimentacaoEstoqueRepository` (histórico de consumo)
- `FornecedorRepository`
- Publicador de eventos de domínio

**Testes**

*Unitários*

- Cálculo do consumo médio com 90 dias de movimentação.
- `sugestaoDisponivel: false` com menos de 30 dias de histórico.
- Arredondamento por unidade de medida: `UN` sobe para inteiro, `L` mantém decimal.
- Rejeita fração em insumo com unidade `UN`.

*Integração*

- Sugestão retorna `200` com `quantidadeSugerida` calculada.
- Insumo sem histórico retorna `sugestaoDisponivel: false`.
- Pedido com item do tipo `PECA` retorna `422`.
- Criar pedido não altera nenhum saldo.
- Duplicidade sem confirmação retorna `409`.

---

### 9.3 Checklist de Implementação

**Domínio**

- [ ] Reaproveitar a entidade `PedidoCompra` do requisito 8 sem duplicar o ciclo de vida
- [ ] Implementar o cálculo de consumo médio diário a partir das movimentações de `SAIDA` dos últimos 90 dias
- [ ] Implementar a fórmula de `quantidadeSugerida` (consumo médio vezes lead time, mais estoque mínimo, menos saldo disponível)
- [ ] Implementar o arredondamento da sugestão conforme a unidade de medida
- [ ] Implementar a regra de `sugestaoDisponivel: false` com menos de 30 dias de histórico
- [ ] Garantir que o pedido de insumo não vincula OS
- [ ] Garantir que criar pedido não altera nenhum saldo de estoque

**Caso de uso**

- [ ] Implementar `CalcularSugestaoDeCompraDeInsumo`
- [ ] Implementar `SolicitarCompraDeInsumos`

**Repositório**

- [ ] Consultar `MovimentacaoEstoqueRepository` para o histórico de consumo

**Handler HTTP**

- [ ] Implementar `GET /estoque/insumos/{insumoId}/sugestao-compra`
- [ ] Reaproveitar `POST /compras/pedidos` com validação por tipo de item

**Validações**

- [ ] Validar decimais compatíveis com a unidade de medida do insumo
- [ ] Rejeitar item do tipo `PECA` neste caso de uso
- [ ] Validar `prazoPrevistoEntrega` no futuro
- [ ] Exigir `confirmarDuplicidade` quando já houver pedido em aberto para o mesmo insumo

**Eventos**

- [ ] Publicar `PedidoCompraCriado`

**Testes unitários**

- [ ] Cálculo de consumo médio com 90 dias de movimentação
- [ ] `sugestaoDisponivel: false` com menos de 30 dias de histórico
- [ ] Arredondamento por unidade: `UN` sobe para inteiro e `L` mantém decimal
- [ ] Rejeição de fração em insumo com unidade `UN`

**Testes de integração**

- [ ] Sugestão retornando `200` com `quantidadeSugerida` calculada
- [ ] Insumo sem histórico retornando `sugestaoDisponivel: false`
- [ ] Item do tipo `PECA` retornando `422`
- [ ] Duplicidade sem confirmação retornando `409`
- [ ] Nenhum saldo de estoque alterado após a criação do pedido

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, incluindo a decisão de rota compartilhada com o requisito 8

**Review**

- [ ] Code Review aprovado

---
