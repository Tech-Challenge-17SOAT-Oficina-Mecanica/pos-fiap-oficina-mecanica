---
documento: Refinamento de Requisitos — Consultar Estoque
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Estoque

Este documento detalha a tarefa Consultar Estoque do contexto de Peças & Insumos.

## 1 · Consultar Estoque

### 1.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Consultar a disponibilidade de peças e insumos durante o diagnóstico, para saber se o
serviço pode ser executado com o que existe na oficina.

**Problema**
Sem saber o que há em estoque no momento do diagnóstico, o mecânico registra na OS peças
que não existem. O orçamento é aprovado e só então se descobre a falta, gerando retrabalho,
atraso na entrega e retorno ao cliente para renegociar prazo.

**Pré-condições**

- Deve existir cadastro de peças e insumos disponível para consulta.
- O usuário deve estar autorizado a consultar o estoque.

**Requisitos Funcionais**

| ID        | Requisito                                                                        |
| --------- | -------------------------------------------------------------------------------- |
| RF-EST-01 | Permitir consultar peças e insumos pelo código, pelo nome ou pela categoria, sendo o código uma busca exata — o mecânico costuma digitar o código em vez do nome. |
| RF-EST-02 | Exibir, para cada item, saldo físico, saldo reservado e saldo disponível.        |
| RF-EST-03 | Indicar visualmente os itens sem saldo disponível.                               |
| RF-EST-04 | Permitir filtrar somente itens ativos.                                           |
| RF-EST-05 | Permitir seguir para a solicitação de compra quando o item estiver indisponível. |

**Requisitos Não Funcionais**

| ID         | Requisito                                                                   |
| ---------- | --------------------------------------------------------------------------- |
| RNF-EST-01 | A consulta deve ser feita por API RESTful.                                  |
| RNF-EST-02 | A operação deve ser acessível somente por usuário autorizado.               |
| RNF-EST-03 | A consulta não deve alterar saldo nem gerar reserva.                        |
| RNF-EST-04 | A listagem deve ser paginada, por se tratar de catálogo com volume elevado. |
| RNF-EST-05 | O saldo exibido deve refletir o estado do momento da consulta, sem cache.   |

**Fluxo Principal**

1. O mecânico informa o critério de busca (código, nome ou categoria).
2. O sistema valida o critério informado.
3. O sistema consulta o cadastro de peças e insumos.
4. O sistema calcula o saldo disponível de cada item.
5. O sistema retorna a lista com saldo físico, reservado e disponível.

**Fluxos Alternativos / Exceções**

| #   | Situação                             | Comportamento do sistema                                                                 |
| --- | ------------------------------------ | ---------------------------------------------------------------------------------------- |
| A1  | Nenhum item encontrado               | Informa que não há item correspondente ao critério.                                      |
| A2  | Item encontrado sem saldo disponível | Retorna o item sinalizado como indisponível e permite seguir para solicitação de compra. |
| A3  | Item inativo                         | Não retorna o item na consulta padrão.                                                   |
| A4  | Usuário sem autorização              | Impede a consulta.                                                                       |

**Saída**

- Lista de peças e insumos com saldo físico, saldo reservado e saldo disponível; **ou**
- Indicação de que nenhum item foi encontrado.

**Pós-condições**

- Os saldos permanecem inalterados: a consulta não reserva nem movimenta estoque.
- A informação de disponibilidade fica disponível para o mecânico registrar as peças
  necessárias na OS.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
GET /estoque/itens
```

> **Decisão de projeto.** O item de estoque tem dois identificadores com papéis distintos: o
> `id`, um **UUID** gerado pelo sistema, usado em rotas, vínculos e payloads, alinhado ao padrão
> das demais APIs do projeto; e o `codigo`, o **código de negócio** do item (`PC-0142`,
> `IN-0031`), digitado e reconhecido pela oficina, usado na busca e na conferência de prateleira.
> O `codigo` é campo de busca de primeira classe: o mecânico costuma digitar o código da peça em
> vez do nome. Nenhuma referência entre recursos, porém, usa o `codigo` — para isso vale o `id`.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:ler`

**Entrada** — query params, todos opcionais:

| Param                | Tipo    | Descrição                                     |
| -------------------- | ------- | --------------------------------------------- |
| `codigo`             | string  | Busca exata pelo código do item, sem diferenciar maiúsculas de minúsculas |
| `busca`              | string  | Busca parcial por `codigo`, descrição ou código de barras |
| `tipo`               | enum    | `PECA` \| `INSUMO`                            |
| `categoria`          | string  | Filtro por categoria                          |
| `somenteDisponiveis` | boolean | Default `false`                               |
| `incluirInativos`    | boolean | Default `false`                               |
| `page` / `size`      | int     | Paginação                                     |

**Validações**

- `size` não pode exceder 100.
- `tipo` deve ser um valor válido do enum.
- `busca`, quando informado, deve ter no mínimo 2 caracteres.
- `codigo`, quando informado, dispensa o mínimo de caracteres: a correspondência é exata.
- `codigo` e `busca` podem vir na mesma requisição; nesse caso os dois filtros se somam.
- Nenhuma validação de negócio — operação puramente de leitura.

**Processamento**

1. Validar e normalizar os query params, incluindo o `codigo` (remover espaços em branco e normalizar para maiúsculas).
2. Montar a query no repositório com os filtros informados: `codigo` por igualdade, `busca` por correspondência parcial em código e descrição.
3. Buscar a página de itens.
4. Calcular `saldoDisponivel = saldoFisico - saldoReservado` para cada item.
5. Marcar `disponivel = saldoDisponivel > 0`.
6. Montar o envelope paginado.

**Persistência**

- Consulta: `item_estoque`
- Altera: nada
- Consulta somente leitura; pode rodar em réplica de leitura, se houver.

**Saída da API**

```json
{
  "data": [
    {
      "id": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "codigo": "PC-0142",
      "tipo": "PECA",
      "descricao": "Pastilha de freio dianteira",
      "categoria": "Freios",
      "unidadeMedida": "UN",
      "precoVenda": 189.9,
      "saldoFisico": 6,
      "saldoReservado": 4,
      "saldoDisponivel": 2,
      "estoqueMinimo": 4,
      "abaixoDoMinimo": true,
      "disponivel": true,
      "ativo": true
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 1,
  "totalPaginas": 1
}
```

**Códigos HTTP / Erros**

| Código | Situação                                                 |
| ------ | -------------------------------------------------------- |
| `200`  | Consulta realizada — a lista pode vir vazia              |
| `400`  | Query param inválido (`size` > 100, `tipo` desconhecido) |
| `401`  | Token ausente ou expirado                                |
| `403`  | Perfil sem o escopo `estoque:ler`                        |

> Lista vazia é `200` com `"data": []`, nunca `404`.

**Dependências**

- `ItemEstoqueRepository`
- Middleware de autenticação/autorização

**Testes**

_Unitários_

- Cálculo de `saldoDisponivel` com reservado zero, parcial e igual ao físico.
- `abaixoDoMinimo` quando o disponível é menor, igual e maior que o mínimo.
- Normalização de filtros, incluindo `codigo` com espaços e em minúsculas.

_Integração_

- `GET` sem filtro retorna a página default.
- `codigo=PC-0142` retorna exatamente o item correspondente.
- `codigo=pc-0142` retorna o mesmo item, sem diferenciar maiúsculas de minúsculas.
- `codigo` inexistente retorna `200` com lista vazia.
- `busca` com parte do código retorna os itens cujo código contém o trecho.
- `tipo=PECA` não retorna insumos.
- `somenteDisponiveis=true` exclui item com disponível zero.
- `incluirInativos=false` exclui item inativo.
- `size=500` retorna `400`.
- Sem token retorna `401`.
- Token de perfil sem escopo retorna `403`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o cálculo de `saldoDisponivel` (saldo físico menos saldo reservado) como método do domínio
- [ ] Implementar as flags `abaixoDoMinimo` e `disponivel`

**Caso de uso**

- [ ] Implementar `ConsultarEstoque` com filtros e paginação

**Repositório**

- [ ] Implementar `ItemEstoqueRepository.buscarPorFiltro` com busca por código, descrição e categoria
- [ ] Implementar a busca exata por `codigo`, com índice único na coluna
- [ ] Garantir que a consulta não abre transação de escrita

**Handler HTTP**

- [ ] Implementar `GET /estoque/itens`
- [ ] Implementar o envelope de resposta paginado

**Validações**

- [ ] Validar `size` com máximo de 100
- [ ] Validar o enum de `tipo`
- [ ] Validar `busca` com no mínimo 2 caracteres

**Testes unitários**

- [ ] Cálculo de saldo disponível com reservado zero, parcial e total
- [ ] Flag `abaixoDoMinimo` nos três limites (abaixo, igual e acima)
- [ ] Normalização dos filtros recebidos

**Testes de integração**

- [ ] Filtros `tipo`, `categoria`, `somenteDisponiveis` e `incluirInativos`
- [ ] Lista vazia retornando `200` e não `404`
- [ ] `size` acima do limite retornando `400`
- [ ] Requisição sem token retornando `401`
- [ ] Perfil sem o escopo `estoque:ler` retornando `403`

**Documentação**

- [ ] Documentar no Swagger/OpenAPI com exemplo de resposta paginada

**Review**

- [ ] Code Review aprovado

---
