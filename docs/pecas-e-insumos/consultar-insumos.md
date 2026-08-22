---
documento: Refinamento de Requisitos — Consultar Insumos
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Consultar Insumos

Este documento detalha a tarefa Consultar Insumos, do contexto de Peças e Insumos.

## 14 · Consultar Insumos

### 14.1 Refinamento de Produto

**Persona**

Mecânico e Gestor.

**Objetivo**

Consultar os insumos cadastrados e verificar se há quantidade suficiente para uma necessidade,
considerando o saldo disponível e a unidade de medida do item.

**Problema**

A existência do insumo no catálogo não garante quantidade suficiente para o serviço. Como os
insumos podem ser controlados em litros, quilogramas, metros ou unidades, a consulta precisa
comparar a necessidade informada com o estoque atual e apresentar claramente o resultado.

**Pré-condições**

- Deve existir cadastro de insumos disponível para consulta.
- Os insumos devem possuir unidade de medida cadastrada.
- Os saldos dos insumos devem estar registrados.
- O usuário deve estar autorizado a consultar o estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-120 | Permitir consultar insumo por código exato. |
| RF-EST-121 | Permitir consultar insumos por descrição parcial e categoria. |
| RF-EST-122 | Permitir consultar um insumo específico por identificador. |
| RF-EST-123 | Permitir informar a quantidade desejada. |
| RF-EST-124 | Apresentar a unidade de medida do insumo. |
| RF-EST-125 | Apresentar saldo físico, reservado e disponível. |
| RF-EST-126 | Comparar a quantidade desejada com o saldo disponível. |
| RF-EST-127 | Indicar explicitamente se a quantidade desejada está disponível. |
| RF-EST-128 | Indicar quando o saldo disponível está abaixo do estoque mínimo. |
| RF-EST-129 | Indicar se existe pedido de compra em aberto. |
| RF-EST-130 | Permitir filtrar somente insumos disponíveis para a quantidade desejada. |
| RF-EST-131 | Ocultar insumos inativos por padrão e permitir sua inclusão explícita. |
| RF-EST-132 | Paginar a listagem de insumos. |
| RF-EST-133 | Informar quando nenhum insumo corresponder aos filtros. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-88 | Disponibilizar a consulta por API REST e restringi-la a usuário autorizado. |
| RNF-EST-89 | Não alterar cadastro ou saldo de estoque. |
| RNF-EST-90 | Não reservar insumo nem gerar pedido de compra. |
| RNF-EST-91 | Retornar listagem paginada no envelope padronizado do projeto. |
| RNF-EST-92 | Refletir o estado atual do estoque no momento da consulta. |
| RNF-EST-93 | Calcular disponibilidade usando saldo disponível e quantidade desejada. |
| RNF-EST-94 | Apresentar a unidade de medida de forma explícita. |

**Fluxo Principal**

1. O mecânico ou gestor acessa a consulta de insumos.
2. O usuário informa um ou mais critérios de busca.
3. O usuário informa a quantidade desejada, quando precisar verificar suficiência.
4. O sistema valida os parâmetros e a autorização.
5. O sistema consulta os insumos correspondentes.
6. O sistema identifica a unidade de medida e os saldos de cada insumo.
7. O sistema calcula o saldo disponível.
8. Quando houver quantidade desejada, o sistema compara a necessidade com o saldo disponível.
9. O sistema indica disponibilidade, estoque mínimo e pedido de compra em aberto.
10. O sistema retorna os resultados paginados.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhum insumo corresponde aos filtros da listagem | Retorna lista vazia sem alterar o estoque. |
| A2 | Insumo específico não encontrado | Informa que o insumo não existe. |
| A3 | Saldo insuficiente | Retorna o insumo com `disponivel: false` e os saldos atuais. |
| A4 | Insumo abaixo do mínimo | Retorna o insumo com `abaixoDoMinimo: true`. |
| A5 | Quantidade inválida ou incompatível com a unidade | Impede a consulta e informa o parâmetro inválido. |
| A6 | Insumo inativo | Não retorna o item na consulta padrão. |
| A7 | Usuário não autenticado ou sem autorização | Impede a consulta. |
| A8 | Falha na consulta | Informa a indisponibilidade sem alterar nenhum dado. |

**Saída**

- Relação paginada de insumos com código, descrição, categoria, unidade de medida, quantidade
  desejada, saldos, disponibilidade, estoque mínimo, pedido em aberto e situação; ou
- Detalhes de um insumo específico; ou
- Indicação de que o insumo específico não foi encontrado.

**Pós-condições**

- Nenhum dado ou saldo é alterado.
- Nenhum insumo é reservado.
- Nenhum pedido de compra é criado.
- O usuário obtém a disponibilidade atual para apoiar a Ordem de Serviço.

---

### 14.2 Refinamento Técnico

**Endpoint**

```http
GET /estoque/insumos
GET /estoque/insumos/{insumoId}
```

A primeira rota pesquisa e lista insumos; a segunda consulta um item específico.

> **Decisão de projeto.** A consulta específica de insumos expõe campos e regras de quantidade
> próprios desse tipo de item. A futura rota geral `GET /estoque/itens` pode oferecer uma visão
> unificada de peça e insumo, mas sua coexistência e responsabilidade ainda precisam ser
> confirmadas.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO` e `GESTOR`.
- Escopo: `estoque:ler`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `insumoId` | UUID | Identificador do insumo na consulta individual. |
| Query | `codigo` | string | Código exato do insumo. |
| Query | `descricao` | string | Trecho da descrição, com pelo menos 2 caracteres. |
| Query | `categoria` | string | Categoria do insumo. |
| Query | `quantidadeDesejada` | decimal | Quantidade positiva cuja disponibilidade será verificada. |
| Query | `somenteDisponiveis` | boolean | Quando `true`, exige `quantidadeDesejada` e filtra saldo suficiente. Padrão `false`. |
| Query | `incluirInativos` | boolean | Inclui insumos inativos. Padrão `false`. |
| Query | `pagina` | inteiro | Página iniciada em zero. Padrão `0`. |
| Query | `tamanho` | inteiro | Itens por página. Padrão `20`, máximo `100`. |

Exemplos:

```http
GET /estoque/insumos?codigo=INS-0012&quantidadeDesejada=2
GET /estoque/insumos?descricao=oleo&quantidadeDesejada=5
GET /estoque/insumos?categoria=Lubrificantes&quantidadeDesejada=3&somenteDisponiveis=true
GET /estoque/insumos/550e8400-e29b-41d4-a716-446655440000?quantidadeDesejada=3
```

**Validações**

*Técnicas*

- Na listagem, deve existir pelo menos um critério entre `codigo`, `descricao` e `categoria`.
- `descricao` deve possuir no mínimo 2 caracteres.
- `quantidadeDesejada` deve ser maior que zero, quando informada.
- `somenteDisponiveis: true` exige `quantidadeDesejada`.
- `pagina` deve ser maior ou igual a zero.
- `tamanho` deve estar entre 1 e 100.
- `insumoId` deve ser um UUID válido na consulta individual.

*Negócio*

- O insumo deve possuir unidade de medida cadastrada.
- A quantidade desejada deve respeitar a precisão permitida pela unidade de medida.
- A disponibilidade deve comparar valores na mesma unidade de medida.
- Insumos inativos são omitidos, salvo quando `incluirInativos: true`.
- A operação é exclusivamente de leitura.

**Processamento**

1. Validar e normalizar os parâmetros.
2. Identificar o usuário e validar o escopo `estoque:ler`.
3. Montar a consulta com os filtros informados.
4. Aplicar o filtro de itens ativos, salvo solicitação explícita em contrário.
5. Buscar os insumos e seus dados atuais de estoque.
6. Calcular `saldoDisponivel = saldoFisico - saldoReservado`.
7. Quando houver quantidade desejada, calcular
   `disponivel = saldoDisponivel >= quantidadeDesejada`.
8. Calcular `abaixoDoMinimo` a partir do saldo disponível.
9. Consultar pedidos de compra em aberto.
10. Aplicar `somenteDisponiveis`, quando solicitado.
11. Montar e retornar a resposta.

**Persistência**

- Consulta: `item_estoque`, dados específicos do insumo e `pedido_compra_item`.
- Altera: nada.
- Não cria: reserva, movimentação ou pedido de compra.

**Saída da API**

```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "codigo": "INS-0012",
      "descricao": "Óleo lubrificante",
      "categoria": "Lubrificantes",
      "unidadeMedida": "L",
      "quantidadeDesejada": 3.0,
      "saldoFisico": 10.0,
      "saldoReservado": 0.0,
      "saldoDisponivel": 10.0,
      "estoqueMinimo": 5.0,
      "disponivel": true,
      "abaixoDoMinimo": false,
      "possuiPedidoEmAberto": false,
      "ativo": true
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 1,
  "totalPaginas": 1
}
```

Quando o insumo existe, mas a quantidade é insuficiente, a API retorna `200` com
`disponivel: false`. Quando a listagem não encontra resultados, retorna `200` com `data: []`.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Consulta concluída, inclusive lista vazia ou quantidade insuficiente. |
| `400` | Parâmetro ou quantidade inválida, ausência de filtro ou paginação fora dos limites. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `estoque:ler`. |
| `404` | Insumo não encontrado na consulta individual. |
| `500` | Falha inesperada sem alteração do estoque. |

**Dependências**

- `InsumoRepository` ou consulta tipada de `ItemEstoqueRepository`.
- `PedidoCompraRepository`.
- Middleware de autenticação e autorização.
- Caso de uso Registrar Insumo Necessário na Ordem de Serviço.
- Caso de uso Solicitar Compra de Insumos.

**Testes**

*Unitários*

- Calcula saldo disponível com reserva zero, parcial e igual ao saldo físico.
- Retorna disponibilidade verdadeira para saldo igual ou superior à quantidade desejada.
- Retorna disponibilidade falsa para saldo inferior.
- Valida quantidade, unidade de medida, descrição, filtros e paginação.
- Calcula corretamente `abaixoDoMinimo`.

*Integração*

- Busca por código, descrição, categoria e identificador retorna `200` quando encontra insumo.
- Quantidade suficiente ou insuficiente é indicada corretamente.
- `somenteDisponiveis: true` exclui itens insuficientes.
- Item inativo não aparece por padrão e aparece quando solicitado.
- Listagem sem resultado retorna `200` com `data: []`.
- Consulta individual inexistente retorna `404`.
- Falhas de autenticação e autorização retornam `401` e `403`.
- A consulta não altera saldo, não reserva e não cria pedido.

---

### 14.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o cálculo de `saldoDisponivel`
- [ ] Implementar disponibilidade considerando `quantidadeDesejada`
- [ ] Validar quantidade conforme a unidade de medida
- [ ] Implementar `abaixoDoMinimo` com base no saldo disponível

**Caso de uso**

- [ ] Implementar `ConsultarInsumos`
- [ ] Implementar filtros, paginação e quantidade desejada
- [ ] Diferenciar listagem vazia de consulta individual inexistente

**Repositório**

- [ ] Consultar por identificador, código, descrição e categoria
- [ ] Filtrar itens ativos e consultar dados de estoque
- [ ] Consultar pedidos de compra em aberto

**Handler HTTP**

- [ ] Implementar `GET /estoque/insumos`
- [ ] Implementar `GET /estoque/insumos/{insumoId}`
- [ ] Aplicar envelope paginado e validar parâmetros
- [ ] Aplicar autenticação, autorização e tratamento de erros

**Validações**

- [ ] Validar presença de critério na listagem
- [ ] Validar descrição, quantidade, unidade, paginação e UUID
- [ ] Exigir quantidade para `somenteDisponiveis: true`

**Testes unitários**

- [ ] Cálculos de saldo, disponibilidade e estoque mínimo
- [ ] Quantidade igual, menor e maior que o saldo disponível
- [ ] Validações de quantidade e unidade de medida

**Testes de integração**

- [ ] Consultas por código, descrição, categoria e identificador
- [ ] Insumo inexistente, inativo e sem saldo suficiente
- [ ] Paginação e filtros de disponibilidade
- [ ] Respostas `200`, `400`, `401`, `403`, `404` e `500`
- [ ] Ausência de alteração, reserva ou pedido de compra

**Documentação**

- [ ] Documentar rotas, parâmetros, unidades e paginação no OpenAPI/Swagger
- [ ] Documentar exemplos com estoque suficiente e insuficiente
- [ ] Documentar a diferença entre lista vazia, item inexistente e saldo insuficiente

**Review**

- [ ] Executar testes automatizados
- [ ] Realizar Code Review
- [ ] Validar os critérios de aceite

---
