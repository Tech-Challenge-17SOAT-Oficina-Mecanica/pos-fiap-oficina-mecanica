---
documento: Refinamento de Requisitos — Consultar Peças Faltantes
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Peças Faltantes

Este documento detalha a tarefa Consultar Peças Faltantes do contexto de Peças & Insumos.

## 7 · Consultar Peças Faltantes

### 7.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Identificar quais peças e insumos precisam de reposição, seja por estarem abaixo do estoque
mínimo, seja por terem sido demandados por uma OS sem saldo disponível.

**Problema**
Sem uma visão consolidada da falta, a reposição é reativa: só se compra quando o serviço já
parou. O resultado é veículo ocupando box da oficina esperando peça e prazo de entrega
estourado com o cliente.

**Pré-condições**

- Deve existir cadastro de peças e insumos com estoque mínimo definido.
- O usuário deve estar autorizado a consultar o estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-38 | Listar os itens cujo saldo disponível está abaixo do estoque mínimo. |
| RF-EST-39 | Listar os itens demandados por OS que não puderam ser reservados por falta de saldo. |
| RF-EST-40 | Exibir, para cada item, saldo físico, saldo reservado, saldo disponível, estoque mínimo e quantidade sugerida de compra. |
| RF-EST-41 | Identificar as OS impactadas por cada item em falta. |
| RF-EST-42 | Permitir filtrar por tipo (peça ou insumo) e por categoria. |
| RF-EST-43 | Permitir ordenar por criticidade, priorizando itens que travam OS em andamento. |
| RF-EST-44 | Permitir seguir para a solicitação de compra a partir do resultado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-33 | A consulta deve ser feita por API RESTful. |
| RNF-EST-34 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-EST-35 | A consulta não deve alterar saldo nem gerar solicitação de compra automaticamente. |
| RNF-EST-36 | A listagem deve ser paginada. |
| RNF-EST-37 | O cálculo de falta deve considerar o saldo disponível, nunca o saldo físico isolado. |

**Fluxo Principal**

1. O mecânico solicita a relação de itens em falta.
2. O sistema calcula o saldo disponível de cada item ativo.
3. O sistema identifica os itens abaixo do estoque mínimo.
4. O sistema identifica os itens demandados por OS sem saldo suficiente.
5. O sistema calcula a quantidade sugerida de compra de cada item.
6. O sistema retorna a relação consolidada, com as OS impactadas.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhum item em falta | Informa que o estoque está regular. |
| A2 | Item sem estoque mínimo definido | Considera apenas a demanda de OS não atendida e sinaliza a ausência do parâmetro. |
| A3 | Item inativo com demanda de OS | Sinaliza o item como descontinuado e orienta a substituição na OS. |
| A4 | Usuário sem autorização | Impede a consulta. |

**Saída**

- Relação de peças e insumos em falta, com saldos, estoque mínimo, quantidade sugerida de compra e OS impactadas; **ou**
- Indicação de que não há itens em falta.

**Pós-condições**

- Os saldos permanecem inalterados — a consulta não movimenta estoque.
- A relação fica disponível como base para a solicitação de compra.

---

### 7.2 Refinamento Técnico

**Endpoint**

```http
GET /estoque/itens/faltantes
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:ler`

**Entrada** — query params, todos opcionais:

| Param | Tipo | Descrição |
|---|---|---|
| `tipo` | enum | `PECA` \| `INSUMO` |
| `categoria` | string | Filtro por categoria |
| `origem` | enum | `MINIMO` \| `DEMANDA_OS` \| `TODAS` (default `TODAS`) |
| `ordenarPor` | enum | `CRITICIDADE` (default) \| `DESCRICAO` |
| `page` / `size` | int | Paginação |

**Validações**

- `size` não pode exceder 100.
- Enums de `tipo`, `origem` e `ordenarPor` devem ser válidos.
- Nenhuma validação de negócio — operação puramente de leitura.

**Processamento**

1. Buscar itens ativos com `saldo_fisico - saldo_reservado < estoque_minimo` — origem `MINIMO`.
2. Buscar itens com demanda de OS não atendida (itens de OS com orçamento aprovado sem reserva ativa correspondente) — origem `DEMANDA_OS`.
3. Unir os dois conjuntos, deduplicando por `item_id` e acumulando as origens.
4. Calcular `quantidadeSugerida = max(estoqueMinimo - saldoDisponivel, demandaNaoAtendida)`.
5. Levantar as OS impactadas por item.
6. Calcular a criticidade: item que trava OS em andamento tem prioridade sobre item apenas abaixo do mínimo.
7. Ordenar e paginar.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `pedido_compra_item` (para sinalizar pedido já em aberto), módulo de OS (itens demandados)
- Altera: nada
- Consulta pesada — candidata natural a read model materializado, atualizado pelos eventos de estoque e de OS, se a performance apertar.

**Saída da API**

```json
{
  "data": [
    {
      "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90",
      "codigo": "PC-0311",
      "tipo": "PECA",
      "descricao": "Disco de freio ventilado",
      "saldoFisico": 1,
      "saldoReservado": 1,
      "saldoDisponivel": 0,
      "estoqueMinimo": 3,
      "quantidadeSugerida": 5,
      "origens": ["MINIMO", "DEMANDA_OS"],
      "criticidade": "ALTA",
      "possuiPedidoEmAberto": false,
      "ordensServicoImpactadas": [
        { "id": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801", "status": "AGUARDANDO_APROVACAO", "quantidade": 2 }
      ]
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 1,
  "totalPaginas": 1
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Consulta realizada — a lista pode vir vazia (estoque regular) |
| `400` | Query param inválido |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `estoque:ler` |

> Estoque regular é `200` com `"data": []`, nunca `404`.

**Dependências**

- `ItemEstoqueRepository`
- `ReservaEstoqueRepository`
- `PedidoCompraRepository`
- Módulo Ordem de Serviço — demanda não atendida
- Caso de uso Solicitar Compra (destino da ação)

**Testes**

*Unitários*

- Item abaixo do mínimo entra com origem `MINIMO`.
- Item com demanda de OS sem saldo entra com origem `DEMANDA_OS`.
- Item nas duas condições aparece uma vez, com as duas origens.
- `quantidadeSugerida` é o maior entre reposição de mínimo e demanda.
- Criticidade `ALTA` quando trava OS em andamento.
- Item sem `estoqueMinimo` definido considera só a demanda.

*Integração*

- Estoque regular retorna `200` com lista vazia.
- Filtro `origem=DEMANDA_OS` exclui itens apenas abaixo do mínimo.
- Item inativo com demanda aparece sinalizado como descontinuado.
- `possuiPedidoEmAberto` fica `true` quando há pedido não recebido.

---

### 7.3 Checklist de Implementação

**Domínio**

- [ ] Implementar a identificação de item abaixo do estoque mínimo usando saldo disponível
- [ ] Implementar a identificação de demanda de OS não atendida (OS aprovada sem reserva correspondente)
- [ ] Implementar a deduplicação por `item_id` acumulando as origens `MINIMO` e `DEMANDA_OS`
- [ ] Implementar o cálculo de `quantidadeSugerida` como o maior entre reposição de mínimo e demanda
- [ ] Implementar o cálculo de criticidade priorizando item que trava OS em andamento

**Caso de uso**

- [ ] Implementar `ConsultarItensFaltantes` com filtros e ordenação

**Repositório**

- [ ] Consultar pedidos em aberto para preencher `possuiPedidoEmAberto`
- [ ] Avaliar read model materializado caso a consulta passe do tempo aceitável

**Integrações**

- [ ] Consultar o módulo de OS para levantar as OS impactadas

**Handler HTTP**

- [ ] Implementar `GET /estoque/itens/faltantes`

**Validações**

- [ ] Validar os enums de `origem` e `ordenarPor`
- [ ] Validar `size` com máximo de 100

**Testes unitários**

- [ ] Item abaixo do mínimo entrando com origem `MINIMO`
- [ ] Item com demanda sem saldo entrando com origem `DEMANDA_OS`
- [ ] Item nas duas condições aparecendo uma única vez com as duas origens
- [ ] Cálculo de `quantidadeSugerida`
- [ ] Criticidade `ALTA` quando trava OS em andamento
- [ ] Item sem `estoqueMinimo` definido considerando apenas a demanda

**Testes de integração**

- [ ] Estoque regular retornando `200` com lista vazia
- [ ] Filtro `origem=DEMANDA_OS` excluindo itens só abaixo do mínimo
- [ ] Item inativo com demanda sinalizado como descontinuado
- [ ] `possuiPedidoEmAberto` igual a `true` quando há pedido não recebido

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, com exemplo de item nas duas origens

**Review**

- [ ] Code Review aprovado

---
