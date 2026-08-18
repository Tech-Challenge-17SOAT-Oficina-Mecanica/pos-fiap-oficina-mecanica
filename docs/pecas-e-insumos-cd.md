---
documento: Refinamento de Requisitos — Contexto de Peças & Insumos (Estoque)
dono: José Lázaro
versao: 0.9
atualizado_em: 2026-08-18
status: rascunho
---

# Refinamento de Requisitos

Este documento reúne, para cada requisito levantado da aplicação, três blocos:

1. **Refinamento de Produto** — o que o usuário precisa e por quê (visão de negócio).
2. **Refinamento Técnico** — como o sistema entrega isso (contrato, processamento, testes).
3. **Checklist de Implementação** — o passo a passo verificável até o merge.

---

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
| RF-EST-01 | Permitir consultar peças e insumos por código, nome ou categoria.                |
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
GET /api/v1/estoque/itens
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:ler`

**Entrada** — query params, todos opcionais:

| Param                | Tipo    | Descrição                                     |
| -------------------- | ------- | --------------------------------------------- |
| `busca`              | string  | Código, descrição parcial ou código de barras |
| `tipo`               | enum    | `PECA` \| `INSUMO`                            |
| `categoria`          | string  | Filtro por categoria                          |
| `somenteDisponiveis` | boolean | Default `false`                               |
| `incluirInativos`    | boolean | Default `false`                               |
| `page` / `size`      | int     | Paginação                                     |

**Validações**

- `size` não pode exceder 100.
- `tipo` deve ser um valor válido do enum.
- `busca`, quando informado, deve ter no mínimo 2 caracteres.
- Nenhuma validação de negócio — operação puramente de leitura.

**Processamento**

1. Validar e normalizar os query params.
2. Montar a query no repositório com os filtros informados.
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
      "id": "PC-0142",
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
- Normalização de filtros.

_Integração_

- `GET` sem filtro retorna a página default.
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
- [ ] Garantir que a consulta não abre transação de escrita

**Handler HTTP**

- [ ] Implementar `GET /api/v1/estoque/itens`
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

## 2 · Atualizar Peça

### 2.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Atualizar os dados cadastrais de uma peça (descrição, preço de venda, estoque mínimo,
categoria, fabricante e situação).

**Problema**
Preço de fornecedor, embalagem e ponto de reposição mudam com o tempo. Sem atualização,
o orçamento é calculado com valor defasado — a oficina vende no prejuízo ou perde o cliente
por preço fora de mercado — e o alerta de reposição dispara na hora errada.

**Pré-condições**

- A peça deve estar cadastrada.
- O usuário deve estar autorizado a manter o cadastro de peças.

**Requisitos Funcionais**

| ID        | Requisito                                                                           |
| --------- | ----------------------------------------------------------------------------------- |
| RF-EST-06 | Permitir alterar descrição, preço de venda, estoque mínimo, categoria e fabricante. |
| RF-EST-07 | Permitir inativar e reativar a peça.                                                |
| RF-EST-08 | Validar os dados informados antes de gravar.                                        |
| RF-EST-09 | Impedir descrição duplicada dentro da mesma categoria.                              |
| RF-EST-10 | Registrar o histórico de alteração de preço, com data e responsável.                |
| RF-EST-11 | Manter inalterados os valores já registrados em OS e orçamentos anteriores.         |

**Requisitos Não Funcionais**

| ID         | Requisito                                                                           |
| ---------- | ----------------------------------------------------------------------------------- |
| RNF-EST-06 | A operação deve ser feita por API RESTful.                                          |
| RNF-EST-07 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-EST-08 | A alteração deve ser auditável — quem alterou, quando e o valor anterior.           |
| RNF-EST-09 | A alteração de preço não pode ter efeito retroativo sobre documentos já emitidos.   |
| RNF-EST-10 | A operação não deve alterar saldo de estoque.                                       |

**Fluxo Principal**

1. O mecânico seleciona a peça a ser atualizada.
2. O sistema apresenta os dados atuais da peça.
3. O mecânico informa os novos dados.
4. O sistema valida os dados informados.
5. O sistema grava a alteração.
6. O sistema registra o histórico da alteração.

**Fluxos Alternativos / Exceções**

| #   | Situação                                       | Comportamento do sistema                                 |
| --- | ---------------------------------------------- | -------------------------------------------------------- |
| A1  | Peça não encontrada                            | Informa que a peça não existe.                           |
| A2  | Dados inválidos                                | Informa quais campos estão incorretos e não grava nada.  |
| A3  | Descrição duplicada na categoria               | Impede a gravação.                                       |
| A4  | Preço menor ou igual a zero                    | Impede a gravação.                                       |
| A5  | Tentativa de inativar peça com saldo reservado | Impede a inativação e informa quais OS mantêm a reserva. |
| A6  | Usuário sem autorização                        | Impede a operação.                                       |

**Saída**

- Peça atualizada com os novos dados; **ou**
- Indicação dos erros de validação encontrados.

**Pós-condições**

- Os dados cadastrais da peça refletem a última alteração.
- O saldo físico e o saldo reservado permanecem inalterados.
- OS e orçamentos anteriores mantêm o preço registrado no momento de sua emissão.
- O histórico de alteração fica disponível para auditoria.

---

### 2.2 Refinamento Técnico

**Endpoint**

```http
PUT /api/v1/estoque/pecas/{pecaId}
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:escrever`

**Entrada**

| Local  | Param           | Tipo    | Descrição                                                  |
| ------ | --------------- | ------- | ---------------------------------------------------------- |
| Path   | `pecaId`        | string  | Identificador da peça                                      |
| Header | `If-Match`      | string  | `version` atual do registro, para controle de concorrência |
| Body   | `descricao`     | string  | Obrigatório, 3 a 120 caracteres                            |
| Body   | `categoria`     | string  | Categoria da peça                                          |
| Body   | `fabricante`    | string  | Fabricante da peça                                         |
| Body   | `precoVenda`    | decimal | Obrigatório, maior que zero, até 2 casas decimais          |
| Body   | `estoqueMinimo` | int     | Maior ou igual a zero                                      |
| Body   | `ativo`         | boolean | Situação da peça                                           |

```json
{
  "descricao": "Pastilha de freio dianteira cerâmica",
  "categoria": "Freios",
  "fabricante": "Bosch",
  "precoVenda": 199.9,
  "estoqueMinimo": 6,
  "ativo": true
}
```

**Validações**

_Técnicas_

- `pecaId` existe e é do tipo `PECA`.
- `descricao` obrigatória, de 3 a 120 caracteres.
- `precoVenda` obrigatório, maior que zero, com no máximo 2 casas decimais.
- `estoqueMinimo` inteiro maior ou igual a zero.
- `If-Match` deve bater com a `version` atual do registro.

_Negócio_

- `descricao` única dentro da mesma categoria.
- Não permitir `ativo = false` quando `saldoReservado > 0`.
- Alteração de preço não propaga para OS ou orçamentos já emitidos.

**Processamento**

1. Carregar a peça por id, com lock otimista.
2. Comparar `If-Match` com a `version` atual — conflito retorna `412`.
3. Validar unicidade de descrição na categoria.
4. Se `ativo = false`, verificar reservas ativas.
5. Detectar mudança de `precoVenda`.
6. Aplicar as alterações na entidade.
7. Se o preço mudou, gravar registro em `historico_preco_item`.
8. Persistir e incrementar `version`.
9. Publicar o evento `PecaAtualizada`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`
- Altera: `item_estoque` (dados cadastrais e `version`), `historico_preco_item` (insert quando o preço muda)
- Não altera: `saldo_fisico`, `saldo_reservado`

**Saída da API**

```json
{
  "id": "PC-0142",
  "tipo": "PECA",
  "descricao": "Pastilha de freio dianteira cerâmica",
  "categoria": "Freios",
  "fabricante": "Bosch",
  "precoVenda": 199.9,
  "estoqueMinimo": 6,
  "ativo": true,
  "version": 8,
  "atualizadoEm": "2026-08-12T14:30:00-03:00",
  "atualizadoPor": "usr-018"
}
```

**Códigos HTTP / Erros**

| Código | Situação                                                                       |
| ------ | ------------------------------------------------------------------------------ |
| `200`  | Peça atualizada                                                                |
| `400`  | Body inválido — campo obrigatório ausente, preço menor ou igual a zero         |
| `401`  | Token ausente ou expirado                                                      |
| `403`  | Perfil sem permissão                                                           |
| `404`  | Peça não encontrada                                                            |
| `409`  | Descrição duplicada na categoria; tentativa de inativar peça com reserva ativa |
| `412`  | `If-Match` divergente — registro alterado por outro usuário                    |
| `422`  | Regra de negócio violada                                                       |

**Dependências**

- `ItemEstoqueRepository`
- `ReservaEstoqueRepository` (verificação de reserva ativa)
- `HistoricoPrecoRepository`
- Publicador de eventos de domínio

**Testes**

_Unitários_

- Rejeita `precoVenda` zero e negativo.
- Rejeita `descricao` com menos de 3 caracteres.
- Gera registro de histórico apenas quando o preço muda.
- Bloqueia inativação com reserva ativa.

_Integração_

- `PUT` válido retorna `200` e incrementa `version`.
- `PUT` com `If-Match` antigo retorna `412`.
- Descrição duplicada na mesma categoria retorna `409`.
- Descrição duplicada em categoria diferente é aceita.
- Inativar peça com reserva ativa retorna `409`.
- Peça inexistente retorna `404`.
- Perfil sem o escopo `estoque:escrever` recebe `403`.

_Regressão_

- Alterar o preço da peça não altera o `precoNoMomento` de OS já criadas.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `atualizarDados()` na entidade `ItemEstoque` com as regras de peça
- [ ] Implementar a regra de bloqueio de inativação quando houver saldo reservado
- [ ] Garantir que a alteração de preço não altera o `precoNoMomento` de OS já emitidas

**Caso de uso**

- [ ] Implementar `AtualizarPeca`

**Repositório**

- [ ] Implementar `ItemEstoqueRepository.salvar` com incremento de `version`
- [ ] Implementar `HistoricoPrecoRepository`
- [ ] Registrar o histórico apenas quando o preço mudar

**Handler HTTP**

- [ ] Implementar `PUT /api/v1/estoque/pecas/{pecaId}`

**Validações**

- [ ] Validar `descricao` entre 3 e 120 caracteres
- [ ] Validar `precoVenda` maior que zero, com no máximo 2 casas decimais
- [ ] Validar `estoqueMinimo` maior ou igual a zero
- [ ] Validar descrição única dentro da mesma categoria

**Concorrência**

- [ ] Implementar controle otimista comparando o header `If-Match` com a `version` atual

**Eventos**

- [ ] Publicar `PecaAtualizada`

**Testes unitários**

- [ ] Rejeição de preço zero e negativo
- [ ] Registro de histórico apenas quando o preço muda
- [ ] Bloqueio de inativação com reserva ativa

**Testes de integração**

- [ ] `PUT` válido retornando `200` com `version` incrementada
- [ ] `If-Match` divergente retornando `412`
- [ ] Descrição duplicada na mesma categoria retornando `409`
- [ ] Descrição duplicada em categoria diferente sendo aceita
- [ ] Peça inexistente retornando `404`
- [ ] Perfil sem o escopo `estoque:escrever` retornando `403`
- [ ] Regressão: OS antiga mantém o preço original após alteração no catálogo

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, incluindo o header `If-Match`

**Review**

- [ ] Code Review aprovado

---


## 3 · Atualizar Insumo

### 3.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Atualizar os dados cadastrais de um insumo (descrição, unidade de medida, custo, estoque
mínimo, categoria e situação).

**Problema**
Insumo não é cobrado item a item do cliente — entra como custo diluído no serviço. Se a
unidade de medida ou o custo estiverem errados, o custo real do serviço fica distorcido e a
oficina perde margem sem perceber.

**Pré-condições**

- O insumo deve estar cadastrado.
- O usuário deve estar autorizado a manter o cadastro de insumos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-12 | Permitir alterar descrição, unidade de medida, custo unitário, estoque mínimo e categoria. |
| RF-EST-13 | Permitir inativar e reativar o insumo. |
| RF-EST-14 | Validar os dados informados antes de gravar. |
| RF-EST-15 | Impedir alteração de unidade de medida quando houver saldo físico maior que zero. |
| RF-EST-16 | Registrar o histórico de alteração de custo, com data e responsável. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-11 | A operação deve ser feita por API RESTful. |
| RNF-EST-12 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-EST-13 | A alteração deve ser auditável. |
| RNF-EST-14 | A operação não deve alterar saldo de estoque. |
| RNF-EST-15 | A alteração de custo não pode ter efeito retroativo sobre serviços já finalizados. |

**Fluxo Principal**

1. O mecânico seleciona o insumo a ser atualizado.
2. O sistema apresenta os dados atuais do insumo.
3. O mecânico informa os novos dados.
4. O sistema valida os dados informados.
5. O sistema grava a alteração.
6. O sistema registra o histórico da alteração.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Insumo não encontrado | Informa que o insumo não existe. |
| A2 | Dados inválidos | Informa quais campos estão incorretos e não grava nada. |
| A3 | Alteração de unidade de medida com saldo em estoque | Impede a alteração e orienta a zerar o saldo antes. |
| A4 | Custo menor ou igual a zero | Impede a gravação. |
| A5 | Usuário sem autorização | Impede a operação. |

**Saída**

- Insumo atualizado com os novos dados; **ou**
- Indicação dos erros de validação encontrados.

**Pós-condições**

- Os dados cadastrais do insumo refletem a última alteração.
- O saldo físico permanece inalterado.
- Serviços já finalizados mantêm o custo registrado no momento de sua execução.
- O histórico de alteração fica disponível para auditoria.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
PUT /api/v1/estoque/insumos/{insumoId}
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Path | `insumoId` | string | Identificador do insumo |
| Header | `If-Match` | string | `version` atual do registro, para controle de concorrência |
| Body | `descricao` | string | Obrigatório, 3 a 120 caracteres |
| Body | `categoria` | string | Categoria do insumo |
| Body | `unidadeMedida` | enum | `UN` \| `L` \| `ML` \| `KG` \| `G` \| `M` |
| Body | `custoUnitario` | decimal | Obrigatório, maior que zero |
| Body | `estoqueMinimo` | decimal | Maior ou igual a zero, aceita casas decimais |
| Body | `ativo` | boolean | Situação do insumo |

```json
{
  "descricao": "Óleo lubrificante 15W40",
  "categoria": "Lubrificantes",
  "unidadeMedida": "L",
  "custoUnitario": 32.50,
  "estoqueMinimo": 20,
  "ativo": true
}
```

**Validações**

*Técnicas*

- `insumoId` existe e é do tipo `INSUMO`.
- `descricao` obrigatória, de 3 a 120 caracteres.
- `custoUnitario` maior que zero.
- `unidadeMedida` pertence ao conjunto permitido (`UN`, `L`, `ML`, `KG`, `G`, `M`).
- `estoqueMinimo` maior ou igual a zero, aceita decimal.
- `If-Match` deve bater com a `version` atual do registro.

*Negócio*

- `descricao` única dentro da mesma categoria.
- Alteração de `unidadeMedida` bloqueada quando `saldoFisico > 0` — converter unidade com saldo distorce todo o histórico.
- Alteração de custo não retroage sobre serviços finalizados.

**Processamento**

1. Carregar o insumo por id.
2. Validar `If-Match`.
3. Se `unidadeMedida` mudou, verificar `saldoFisico == 0`.
4. Validar unicidade de descrição na categoria.
5. Detectar mudança de `custoUnitario`.
6. Aplicar as alterações na entidade.
7. Gravar registro em `historico_preco_item` quando o custo mudar.
8. Persistir e incrementar `version`.
9. Publicar o evento `InsumoAtualizado`.

**Persistência**

- Consulta: `item_estoque`
- Altera: `item_estoque`, `historico_preco_item`
- Não altera: `saldo_fisico`, `saldo_reservado`

**Saída da API**

```json
{
  "id": "IN-0031",
  "tipo": "INSUMO",
  "descricao": "Óleo lubrificante 15W40",
  "categoria": "Lubrificantes",
  "unidadeMedida": "L",
  "custoUnitario": 32.50,
  "estoqueMinimo": 20,
  "saldoFisico": 47.5,
  "ativo": true,
  "version": 4,
  "atualizadoEm": "2026-08-12T14:35:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Insumo atualizado |
| `400` | Body inválido; `unidadeMedida` fora do conjunto permitido |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem permissão |
| `404` | Insumo não encontrado |
| `409` | Descrição duplicada na categoria |
| `412` | `If-Match` divergente — registro alterado por outro usuário |
| `422` | Troca de unidade de medida com saldo em estoque |

**Dependências**

- `ItemEstoqueRepository`
- `HistoricoPrecoRepository`
- Publicador de eventos de domínio

**Testes**

*Unitários*

- Rejeita `unidadeMedida` fora do enum.
- Bloqueia troca de unidade com `saldoFisico > 0`.
- Permite troca de unidade com `saldoFisico = 0`.
- Aceita `estoqueMinimo` decimal.

*Integração*

- `PUT` válido retorna `200`.
- Troca de unidade com saldo retorna `422`.
- `If-Match` divergente retorna `412`.
- Insumo inexistente retorna `404`.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `atualizarDados()` na entidade `ItemEstoque` com as regras de insumo
- [ ] Implementar o enum de unidade de medida (`UN`, `L`, `ML`, `KG`, `G`, `M`)
- [ ] Implementar o bloqueio de troca de unidade de medida com saldo físico maior que zero
- [ ] Garantir que a alteração de custo não retroage sobre serviços já finalizados

**Caso de uso**

- [ ] Implementar `AtualizarInsumo`

**Repositório**

- [ ] Registrar em `HistoricoPrecoRepository` na mudança de custo

**Handler HTTP**

- [ ] Implementar `PUT /api/v1/estoque/insumos/{insumoId}`

**Validações**

- [ ] Validar `descricao` entre 3 e 120 caracteres
- [ ] Validar `custoUnitario` maior que zero
- [ ] Validar `unidadeMedida` dentro do enum permitido
- [ ] Validar descrição única dentro da mesma categoria

**Concorrência**

- [ ] Implementar controle otimista via `If-Match`

**Eventos**

- [ ] Publicar `InsumoAtualizado`

**Testes unitários**

- [ ] Rejeição de unidade fora do enum
- [ ] Bloqueio de troca de unidade com saldo maior que zero
- [ ] Troca de unidade permitida com saldo zerado
- [ ] Aceitação de `estoqueMinimo` decimal

**Testes de integração**

- [ ] `PUT` válido retornando `200`
- [ ] Troca de unidade com saldo retornando `422`
- [ ] `If-Match` divergente retornando `412`
- [ ] Insumo inexistente retornando `404`

**Documentação**

- [ ] Documentar no Swagger/OpenAPI

**Review**

- [ ] Code Review aprovado

---


## 4 · Registrar Entrada de Estoque

### 4.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Registrar o recebimento de peças e insumos, aumentando o saldo físico do estoque.

**Problema**
Sem registro de entrada, o saldo do sistema diverge do saldo real da prateleira. A
consequência é dupla: o mecânico deixa de reservar peça que existe, ou reserva peça que já
acabou — e a oficina só descobre no momento de executar o serviço.

**Pré-condições**

- O item (peça ou insumo) deve estar cadastrado e ativo.
- Deve existir um documento de origem do recebimento (nota fiscal ou pedido de compra).
- O usuário deve estar autorizado a movimentar estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-17 | Permitir registrar entrada informando item, quantidade, custo unitário e documento de origem. |
| RF-EST-18 | Permitir registrar a entrada de vários itens em um mesmo recebimento. |
| RF-EST-19 | Validar que a quantidade informada é maior que zero. |
| RF-EST-20 | Vincular a entrada ao pedido de compra correspondente, quando houver. |
| RF-EST-21 | Atualizar o saldo físico do item. |
| RF-EST-22 | Registrar a movimentação no histórico de estoque. |
| RF-EST-23 | Atualizar a situação do pedido de compra quando o recebimento for total. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-16 | A operação deve ser feita por API RESTful. |
| RNF-EST-17 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-EST-18 | A entrada deve ser transacional — ou todos os itens do recebimento entram, ou nenhum entra. |
| RNF-EST-19 | A operação deve ser idempotente em relação ao documento de origem, para impedir dupla contagem em caso de reenvio. |
| RNF-EST-20 | A movimentação deve ser auditável e o histórico imutável. |

**Fluxo Principal**

1. O mecânico informa o documento de origem do recebimento.
2. O mecânico informa os itens, as quantidades e os custos unitários.
3. O sistema valida os itens e as quantidades.
4. O sistema atualiza o saldo físico de cada item.
5. O sistema registra a movimentação de entrada no histórico.
6. O sistema atualiza a situação do pedido de compra vinculado.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Item não encontrado ou inativo | Impede a entrada e informa qual item está irregular. |
| A2 | Quantidade menor ou igual a zero | Impede a entrada. |
| A3 | Documento de origem já registrado | Informa que o recebimento já foi lançado e não altera o saldo. |
| A4 | Recebimento parcial | Registra a quantidade recebida e mantém o pedido de compra em aberto com o saldo pendente. |
| A5 | Quantidade recebida maior que a pedida | Alerta a divergência e exige confirmação antes de gravar. |
| A6 | Usuário sem autorização | Impede a operação. |

**Saída**

- Confirmação da entrada com o saldo físico atualizado de cada item; **ou**
- Indicação do motivo pelo qual a entrada foi recusada.

**Pós-condições**

- O saldo físico dos itens recebidos está acrescido da quantidade informada.
- O saldo reservado permanece inalterado.
- A movimentação de entrada está registrada no histórico.
- O pedido de compra vinculado está atualizado — concluído ou parcialmente atendido.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
POST /api/v1/estoque/entradas
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:movimentar`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | Recomendado; impede dupla contagem em caso de reenvio |
| Body | `documentoOrigem` | string | Obrigatório; nota fiscal ou documento do recebimento |
| Body | `fornecedorId` | string | Fornecedor do recebimento |
| Body | `pedidoCompraId` | string | Pedido de compra vinculado, quando houver |
| Body | `itens[]` | array | Obrigatório, não vazio, máximo 200 linhas |
| Body | `itens[].itemId` | string | Item recebido; não pode repetir no mesmo payload |
| Body | `itens[].quantidade` | decimal | Maior que zero |
| Body | `itens[].custoUnitario` | decimal | Maior que zero |
| Body | `confirmarDivergencia` | boolean | Obrigatório quando a quantidade recebida excede a pedida |

```json
{
  "documentoOrigem": "NF-88421",
  "fornecedorId": "FOR-004",
  "pedidoCompraId": "PC-2026-0117",
  "itens": [
    { "itemId": "PC-0142", "quantidade": 10, "custoUnitario": 118.40 },
    { "itemId": "IN-0031", "quantidade": 20.0, "custoUnitario": 30.10 }
  ]
}
```

**Validações**

*Técnicas*

- `documentoOrigem` obrigatório.
- `itens` não vazio, com no máximo 200 linhas.
- `quantidade` maior que zero em cada linha.
- `custoUnitario` maior que zero.
- Sem `itemId` repetido no mesmo payload.

*Negócio*

- Todos os `itemId` existem e estão ativos.
- `documentoOrigem` ainda não registrado — chave única em `movimentacao_estoque`.
- Quando há `pedidoCompraId`, os itens devem pertencer ao pedido.
- Quantidade recebida maior que a pedida exige `confirmarDivergencia: true` no body.

**Processamento**

1. Verificar o `Idempotency-Key` — se já processada, retornar a resposta original.
2. Abrir transação.
3. Validar o payload e carregar todos os itens com `SELECT ... FOR UPDATE`.
4. Verificar duplicidade de `documentoOrigem`.
5. Se houver `pedidoCompraId`, carregar o pedido e conferir divergência de quantidade.
6. Para cada linha: `saldoFisico += quantidade`.
7. Inserir uma `movimentacao_estoque` do tipo `ENTRADA` por linha.
8. Atualizar `quantidade_recebida` em `pedido_compra_item`.
9. Recalcular o status do pedido: `ABERTO` para `PARCIAL` ou `CONCLUIDO`.
10. Commit.
11. Publicar o evento `EntradaRegistrada`.

**Persistência**

- Consulta: `item_estoque`, `pedido_compra`, `pedido_compra_item`, `chave_idempotencia`
- Altera: `item_estoque.saldo_fisico`, `movimentacao_estoque` (insert), `pedido_compra_item.quantidade_recebida`, `pedido_compra.status`, `chave_idempotencia` (insert)
- Não altera: `saldo_reservado`
- Tudo em uma transação — ou todas as linhas entram, ou nenhuma entra.

**Saída da API**

```json
{
  "entradaId": "ENT-2026-0455",
  "documentoOrigem": "NF-88421",
  "registradoEm": "2026-08-12T15:02:00-03:00",
  "registradoPor": "usr-018",
  "itens": [
    {
      "itemId": "PC-0142",
      "quantidade": 10,
      "saldoFisicoAnterior": 6,
      "saldoFisicoAtual": 16,
      "saldoDisponivel": 12
    },
    {
      "itemId": "IN-0031",
      "quantidade": 20.0,
      "saldoFisicoAnterior": 27.5,
      "saldoFisicoAtual": 47.5,
      "saldoDisponivel": 47.5
    }
  ],
  "pedidoCompra": { "id": "PC-2026-0117", "status": "CONCLUIDO" }
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Entrada registrada |
| `200` | Requisição repetida com a mesma `Idempotency-Key` — retorna a resposta original |
| `400` | Body inválido; quantidade menor ou igual a zero; item repetido no payload |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `estoque:movimentar` |
| `404` | Item ou pedido de compra não encontrado |
| `409` | `documentoOrigem` já registrado |
| `422` | Item inativo; quantidade recebida maior que a pedida sem `confirmarDivergencia` |

**Dependências**

- `ItemEstoqueRepository`
- `MovimentacaoEstoqueRepository`
- `PedidoCompraRepository`
- Serviço de idempotência
- Publicador de eventos de domínio
- Caso de uso Solicitar Compra (origem do `pedidoCompraId`)

**Testes**

*Unitários*

- Soma correta do saldo por linha.
- Rejeita quantidade zero e negativa.
- Detecta `itemId` duplicado no payload.
- Cálculo de status do pedido: parcial versus concluído.

*Integração*

- Entrada de 2 itens atualiza os dois saldos e cria 2 movimentações.
- Mesma `Idempotency-Key` duas vezes: o saldo sobe uma vez só.
- `documentoOrigem` repetido retorna `409`.
- Item inativo retorna `422` e nenhum saldo é alterado (rollback).
- Recebimento parcial mantém o pedido em `PARCIAL`.
- Recebimento total move o pedido para `CONCLUIDO`.

*Concorrência*

- Duas entradas simultâneas do mesmo item somam corretamente, sem perda de atualização.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `registrarEntrada()` na entidade `ItemEstoque` somando o saldo físico
- [ ] Implementar a invariante de quantidade maior que zero
- [ ] Implementar a entidade `MovimentacaoEstoque` do tipo `ENTRADA` como histórico imutável
- [ ] Implementar o recálculo de status do `PedidoCompra` (`ABERTO`, `PARCIAL`, `CONCLUIDO`)

**Caso de uso**

- [ ] Implementar `RegistrarEntradaEstoque` com múltiplos itens
- [ ] Implementar a regra de divergência: quantidade recebida maior que a pedida exige `confirmarDivergencia`

**Repositório**

- [ ] Implementar `MovimentacaoEstoqueRepository`
- [ ] Atualizar `quantidadeRecebida` em `PedidoCompraItem`
- [ ] Criar constraint de unicidade de `documentoOrigem` em `movimentacao_estoque`

**Handler HTTP**

- [ ] Implementar `POST /api/v1/estoque/entradas`

**Validações**

- [ ] Validar `documentoOrigem` obrigatório
- [ ] Validar `itens` não vazio e sem `itemId` repetido
- [ ] Validar `quantidade` maior que zero em cada linha
- [ ] Validar `custoUnitario` maior que zero
- [ ] Validar que todos os itens existem e estão ativos

**Transação e idempotência**

- [ ] Executar a operação inteira em uma única transação (todas as linhas entram ou nenhuma entra)
- [ ] Implementar a tabela `chave_idempotencia` e o fluxo do header `Idempotency-Key`
- [ ] Gravar a chave de idempotência dentro da mesma transação da operação

**Eventos**

- [ ] Publicar `EntradaRegistrada`

**Testes unitários**

- [ ] Soma correta do saldo por linha
- [ ] Rejeição de quantidade zero e negativa
- [ ] Detecção de `itemId` duplicado no payload
- [ ] Cálculo de status do pedido: parcial versus concluído

**Testes de integração**

- [ ] Entrada com 2 itens atualizando os dois saldos e criando 2 movimentações
- [ ] Mesma `Idempotency-Key` duas vezes somando o saldo uma única vez
- [ ] `documentoOrigem` repetido retornando `409`
- [ ] Item inativo retornando `422` com rollback total dos saldos
- [ ] Recebimento parcial mantendo o pedido em `PARCIAL`
- [ ] Recebimento total movendo o pedido para `CONCLUIDO`

**Testes de concorrência**

- [ ] Duas entradas simultâneas do mesmo item somando corretamente, sem perda de atualização

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, incluindo o header `Idempotency-Key`

**Review**

- [ ] Code Review aprovado

---


## 5 · Reservar Peça para OS

### 5.1 Refinamento de Produto

**Persona**
Sistema, acionado pela aprovação do orçamento.
Beneficiário: mecânico, que passa a ter a peça garantida para executar o serviço.

**Objetivo**
Separar as peças de uma OS no estoque, garantindo que não sejam usadas em outro atendimento.

**Problema**
Duas OS aprovadas no mesmo dia podem depender da mesma última peça em estoque. Sem reserva,
quem chegar primeiro na prateleira leva, e o segundo cliente descobre o problema com o
veículo já desmontado. A reserva transforma "tem em estoque" em "está garantido para esta OS".

**Pré-condições**

- A OS deve existir e conter itens de peça registrados.
- O orçamento da OS deve estar aprovado.
- As peças devem estar cadastradas e ativas.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-24 | Permitir reservar todas as peças de uma OS em uma única operação. |
| RF-EST-25 | Validar o saldo disponível de cada peça antes de reservar. |
| RF-EST-26 | Aumentar o saldo reservado da peça e reduzir o saldo disponível. |
| RF-EST-27 | Vincular cada reserva à OS de origem. |
| RF-EST-28 | Informar quais peças não puderam ser reservadas por falta de saldo. |
| RF-EST-29 | Permitir liberar a reserva quando a OS for cancelada ou o orçamento recusado. |
| RF-EST-30 | Registrar a movimentação de reserva no histórico. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-21 | A operação deve ser feita por API RESTful. |
| RNF-EST-22 | A operação deve ser acessível somente por usuário ou serviço autorizado. |
| RNF-EST-23 | A reserva deve ser atômica e protegida contra concorrência — duas OS não podem reservar a mesma unidade simultaneamente. |
| RNF-EST-24 | A operação deve ser transacional — ou todas as peças da OS são reservadas, ou nenhuma é. |
| RNF-EST-25 | A operação deve ser idempotente por OS, para impedir reserva em dobro em caso de reprocessamento da mensagem. |
| RNF-EST-26 | A reserva não altera o saldo físico, apenas o saldo reservado. |

**Fluxo Principal**

1. O sistema recebe o pedido de reserva das peças de uma OS aprovada.
2. O sistema valida que a OS existe e possui orçamento aprovado.
3. O sistema verifica o saldo disponível de cada peça da OS.
4. O sistema aumenta o saldo reservado de cada peça.
5. O sistema vincula as reservas à OS.
6. O sistema registra a movimentação de reserva no histórico.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Peça com saldo disponível insuficiente | Não reserva nenhuma peça da OS, informa quais faltaram e sinaliza a indisponibilidade para que a compra seja solicitada. |
| A2 | Reserva já existente para a OS | Não reserva novamente e retorna a reserva vigente. |
| A3 | OS sem orçamento aprovado | Impede a reserva. |
| A4 | OS cancelada ou orçamento recusado | Libera as reservas vinculadas e devolve o saldo ao disponível. |
| A5 | Peça inativada após o orçamento | Impede a reserva e sinaliza a necessidade de substituição do item na OS. |
| A6 | Usuário ou serviço sem autorização | Impede a operação. |

**Saída**

- Confirmação da reserva com a relação de peças e quantidades reservadas para a OS; **ou**
- Indicação das peças indisponíveis que impediram a reserva.

**Pós-condições**

- O saldo reservado das peças está acrescido das quantidades da OS.
- O saldo disponível está reduzido na mesma proporção.
- O saldo físico permanece inalterado.
- As peças ficam vinculadas à OS e indisponíveis para outros atendimentos.
- A OS está liberada para iniciar a execução.

---

### 5.2 Refinamento Técnico

**Endpoint**

```http
POST   /api/v1/estoque/reservas
DELETE /api/v1/estoque/reservas/ordens-servico/{osId}
```

O `DELETE` atende a liberação da reserva (OS cancelada ou orçamento recusado).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `SERVICO` (chamada da política de orçamento aprovado), `MECANICO`, `GESTOR`
- Escopo: `estoque:movimentar`
- O mecânico não reserva diretamente: a reserva é consequência da aprovação do orçamento

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | Obrigatório neste endpoint |
| Body | `ordemServicoId` | string | Obrigatório; OS de origem da reserva |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido |
| Body | `itens[].itemId` | string | Peça a reservar; item do tipo `INSUMO` é rejeitado |
| Body | `itens[].quantidade` | int | Inteiro maior que zero |
| Path (DELETE) | `osId` | string | OS cujas reservas ativas serão liberadas |

```json
{
  "ordemServicoId": "OS-2026-0912",
  "itens": [
    { "itemId": "PC-0142", "quantidade": 4 },
    { "itemId": "PC-0311", "quantidade": 1 }
  ]
}
```

**Validações**

*Técnicas*

- `ordemServicoId` obrigatório.
- `itens` não vazio, sem `itemId` repetido.
- `quantidade` inteira maior que zero.
- Todos os itens do tipo `PECA` — insumo não é reservável.

*Negócio*

- A OS existe e possui orçamento aprovado vigente.
- Todas as peças estão ativas.
- `saldoDisponivel >= quantidade` para todas as peças; a reserva é tudo ou nada.
- Não existe reserva `ATIVA` para a mesma OS (idempotência de negócio).

**Processamento**

*Reserva (POST)*

1. Verificar o `Idempotency-Key`; se já processada, retornar a resposta original.
2. Consultar o módulo de OS: a OS existe e tem orçamento aprovado?
3. Verificar reserva `ATIVA` existente para a OS — se houver, retornar a reserva vigente.
4. Abrir transação.
5. Carregar todas as peças com `SELECT ... FOR UPDATE`, ordenadas por `item_id` — a ordem fixa evita deadlock entre transações concorrentes.
6. Para cada peça, conferir `saldo_fisico - saldo_reservado >= quantidade`.
7. Se qualquer peça falhar: rollback, montar a lista de indisponíveis e publicar `PecaIndisponivel`.
8. Se todas passarem: `saldo_reservado += quantidade`, com guarda na cláusula `WHERE`.
9. Inserir `reserva_estoque` com status `ATIVA` por item.
10. Inserir `movimentacao_estoque` do tipo `RESERVA`.
11. Commit.
12. Publicar o evento `PecaReservada`.

*Liberação (DELETE)*

1. Carregar as reservas `ATIVAS` da OS com lock.
2. `saldo_reservado -= quantidade` de cada peça.
3. Marcar as reservas como `LIBERADA`.
4. Inserir movimentação do tipo `LIBERACAO`.
5. Publicar o evento `ReservaLiberada`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `chave_idempotencia`, módulo de OS (externo ao agregado)
- Altera: `item_estoque.saldo_reservado`, `reserva_estoque` (insert e update de status), `movimentacao_estoque` (insert)
- Não altera: `saldo_fisico`
- Isolamento mínimo: `READ COMMITTED` com lock explícito de linha.

**Saída da API**

Sucesso (`POST`):

```json
{
  "reservaId": "RSV-2026-0308",
  "ordemServicoId": "OS-2026-0912",
  "status": "ATIVA",
  "reservadoEm": "2026-08-12T15:20:00-03:00",
  "itens": [
    { "itemId": "PC-0142", "quantidade": 4, "saldoDisponivelApos": 8 },
    { "itemId": "PC-0311", "quantidade": 1, "saldoDisponivelApos": 2 }
  ]
}
```

Saldo insuficiente (`409`):

```json
{
  "type": "https://api.oficina/errors/saldo-insuficiente",
  "title": "Saldo insuficiente",
  "status": 409,
  "detail": "Não foi possível reservar as peças da OS-2026-0912",
  "erros": [{ "itemId": "PC-0311", "solicitado": 1, "disponivel": 0 }]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Reserva criada |
| `200` | Reserva já existente para a OS; repetição de `Idempotency-Key` |
| `204` | Reserva liberada (`DELETE`) |
| `400` | Body inválido; item do tipo `INSUMO` |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `estoque:movimentar` |
| `404` | OS ou peça não encontrada; nenhuma reserva ativa no `DELETE` |
| `409` | Saldo insuficiente — nenhuma peça foi reservada |
| `422` | OS sem orçamento aprovado; peça inativada após o orçamento |

**Dependências**

- `ItemEstoqueRepository`
- `ReservaEstoqueRepository`
- `MovimentacaoEstoqueRepository`
- Módulo Ordem de Serviço — consulta de status e do orçamento vigente
- Módulo Orçamento — origem do evento `OrcamentoAprovado` que dispara a chamada
- Serviço de idempotência
- Publicador de eventos de domínio

**Testes**

*Unitários*

- Reserva tudo ou nada: uma peça sem saldo impede as demais.
- Cálculo de `saldoDisponivelApos`.
- Rejeita item do tipo `INSUMO`.
- Liberação devolve exatamente a quantidade reservada.

*Integração*

- Reserva válida retorna `201`, `saldo_reservado` sobe e `saldo_fisico` não muda.
- Saldo insuficiente retorna `409` e nenhum saldo é alterado.
- OS sem orçamento aprovado retorna `422`.
- Segunda chamada com a mesma `Idempotency-Key` não duplica a reserva.
- `DELETE` devolve o saldo ao disponível e marca a reserva como `LIBERADA`.
- `DELETE` em OS sem reserva ativa retorna `404`.

*Concorrência (obrigatórios)*

- Duas OS reservando a última peça em paralelo: exatamente uma recebe `201`, a outra `409`.
- Reserva de duas OS com os mesmos itens em ordens diferentes de payload não gera deadlock (validação da ordenação por `item_id`).
- Teste de carga: N requisições simultâneas nunca deixam `saldo_reservado > saldo_fisico`.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `reservar()` na entidade `ItemEstoque` aumentando o saldo reservado
- [ ] Implementar o método `liberarReserva()` na entidade `ItemEstoque`
- [ ] Implementar a invariante de `saldoReservado` nunca maior que `saldoFisico`
- [ ] Implementar a invariante de não reservar acima do saldo disponível
- [ ] Implementar a entidade `ReservaEstoque` com status `ATIVA`, `LIBERADA` e `CONSUMIDA`

**Caso de uso**

- [ ] Implementar `ReservarPecasParaOS` com semântica tudo ou nada
- [ ] Implementar `LiberarReservaDaOS`
- [ ] Implementar o rollback total quando qualquer peça não tiver saldo, sem reservar nenhuma

**Repositório**

- [ ] Implementar `ReservaEstoqueRepository`
- [ ] Registrar `MovimentacaoEstoque` dos tipos `RESERVA` e `LIBERACAO`

**Integrações**

- [ ] Consultar o módulo de Ordem de Serviço para validar existência e orçamento aprovado
- [ ] Assinar o evento `OrcamentoAprovado` e acionar a reserva
- [ ] Assinar os eventos `OrcamentoRecusado` e `OSCancelada` e acionar a liberação

**Handler HTTP**

- [ ] Implementar `POST /api/v1/estoque/reservas`
- [ ] Implementar `DELETE /api/v1/estoque/reservas/ordens-servico/{osId}`

**Validações**

- [ ] Validar `ordemServicoId` obrigatório
- [ ] Validar `itens` não vazio e sem repetição
- [ ] Validar `quantidade` inteira maior que zero
- [ ] Rejeitar item do tipo `INSUMO` com `400`
- [ ] Validar que todas as peças estão ativas

**Concorrência e idempotência**

- [ ] Implementar `SELECT ... FOR UPDATE` nas linhas de `item_estoque` dentro da transação
- [ ] Ordenar o carregamento das linhas por `item_id` para evitar deadlock
- [ ] Implementar guarda na cláusula `WHERE` do `UPDATE` (saldo físico menos saldo reservado maior ou igual à quantidade)
- [ ] Tornar a `Idempotency-Key` obrigatória neste endpoint
- [ ] Implementar idempotência de negócio: reserva `ATIVA` existente para a OS retorna a reserva vigente com `200`

**Eventos**

- [ ] Publicar `PecaReservada`
- [ ] Publicar `PecaIndisponivel` no caminho triste
- [ ] Publicar `ReservaLiberada`

**Testes unitários**

- [ ] Semântica tudo ou nada com uma peça sem saldo
- [ ] Cálculo de `saldoDisponivelApos`
- [ ] Rejeição de item do tipo `INSUMO`
- [ ] Liberação devolvendo exatamente a quantidade reservada

**Testes de integração**

- [ ] Reserva válida com saldo reservado subindo e saldo físico inalterado
- [ ] Saldo insuficiente retornando `409` sem alterar nenhum saldo
- [ ] OS sem orçamento aprovado retornando `422`
- [ ] Segunda chamada com a mesma `Idempotency-Key` não duplicando a reserva
- [ ] `DELETE` devolvendo o saldo e marcando a reserva como `LIBERADA`
- [ ] `DELETE` em OS sem reserva ativa retornando `404`

**Testes de concorrência**

- [ ] Duas OS disputando a última peça: exatamente um `201` e um `409`
- [ ] Payloads com itens em ordem invertida sem gerar deadlock
- [ ] Teste de carga garantindo que `saldo_reservado` nunca ultrapassa `saldo_fisico`

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI, com o exemplo de erro `409` de saldo insuficiente

**Review**

- [ ] Code Review aprovado

---


## 6 · Registrar Consumo e Saída

### 6.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Dar baixa nas peças e insumos efetivamente utilizados na execução do serviço, reduzindo o
saldo físico do estoque.

**Problema**
A reserva apenas separa a peça; ela continua contando como saldo físico. Se a baixa não for
registrada quando a peça é montada no veículo, o sistema segue afirmando que ela existe na
prateleira — e o estoque acumula divergência até o inventário.

**Pré-condições**

- A OS deve estar em execução.
- As peças devem estar reservadas para a OS.
- O usuário deve estar autorizado a movimentar estoque.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-31 | Permitir registrar a saída de peças reservadas para a OS. |
| RF-EST-32 | Permitir registrar o consumo de insumos vinculado à OS. |
| RF-EST-33 | Reduzir o saldo físico e o saldo reservado da peça na mesma operação. |
| RF-EST-34 | Reduzir o saldo físico do insumo consumido. |
| RF-EST-35 | Permitir registrar quantidade consumida menor que a reservada e devolver a diferença ao saldo disponível. |
| RF-EST-36 | Registrar a movimentação de saída no histórico, vinculada à OS. |
| RF-EST-37 | Impedir saída de peça não reservada para a OS. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-27 | A operação deve ser feita por API RESTful. |
| RNF-EST-28 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-EST-29 | A operação deve ser transacional — a baixa do saldo físico e a baixa da reserva ocorrem juntas ou não ocorrem. |
| RNF-EST-30 | O saldo físico não pode ficar negativo em nenhuma hipótese. |
| RNF-EST-31 | A movimentação deve ser auditável e o histórico imutável. |
| RNF-EST-32 | A operação deve ser idempotente por item da OS, para impedir baixa em duplicidade. |

**Fluxo Principal**

1. O mecânico informa a OS e os itens efetivamente utilizados.
2. O sistema valida que a OS está em execução.
3. O sistema valida que as peças informadas estão reservadas para essa OS.
4. O sistema reduz o saldo físico e o saldo reservado das peças.
5. O sistema reduz o saldo físico dos insumos consumidos.
6. O sistema registra a movimentação de saída no histórico, vinculada à OS.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Peça não reservada para a OS | Impede a saída e orienta a registrar a peça na OS e reservá-la antes. |
| A2 | Consumo menor que o reservado | Baixa a quantidade utilizada e libera a diferença de volta ao saldo disponível. |
| A3 | Consumo maior que o reservado | Verifica o saldo disponível e, havendo saldo, registra a diferença como consumo adicional; não havendo, impede a operação e sinaliza a indisponibilidade. |
| A4 | Insumo com saldo físico insuficiente | Impede a baixa e sinaliza a necessidade de reposição. |
| A5 | Baixa já registrada para o item | Informa que a saída já foi lançada e não altera o saldo. |
| A6 | OS fora do status de execução | Impede a operação. |
| A7 | Usuário sem autorização | Impede a operação. |

**Saída**

- Confirmação da saída com os saldos atualizados dos itens movimentados; **ou**
- Indicação do motivo pelo qual a baixa foi recusada.

**Pós-condições**

- O saldo físico das peças e insumos utilizados está reduzido.
- O saldo reservado das peças baixadas está zerado para essa OS.
- Eventual diferença entre reservado e consumido voltou ao saldo disponível.
- A movimentação está registrada no histórico e vinculada à OS, permitindo apurar o custo real do serviço.

---

### 6.2 Refinamento Técnico

**Endpoint**

```http
POST /api/v1/estoque/saidas
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:movimentar`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Header | `Idempotency-Key` | uuid | Recomendado; impede baixa em duplicidade |
| Body | `ordemServicoId` | string | Obrigatório; OS em execução |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido |
| Body | `itens[].itemId` | string | Peça reservada para a OS ou insumo consumido |
| Body | `itens[].quantidade` | decimal | Maior ou igual a zero, com casas decimais compatíveis com a `unidadeMedida` |

```json
{
  "ordemServicoId": "OS-2026-0912",
  "itens": [
    { "itemId": "PC-0142", "quantidade": 4 },
    { "itemId": "PC-0311", "quantidade": 0 },
    { "itemId": "IN-0031", "quantidade": 3.5 }
  ]
}
```

`quantidade: 0` em peça reservada significa peça não utilizada: a reserva é devolvida ao saldo disponível.

**Validações**

*Técnicas*

- `ordemServicoId` obrigatório.
- `itens` não vazio, sem `itemId` repetido.
- `quantidade` maior ou igual a zero, com casas decimais compatíveis com a `unidadeMedida`.

*Negócio*

- A OS está com status `EM_EXECUCAO`.
- Toda peça informada possui reserva `ATIVA` para essa OS.
- Consumo de peça maior que o reservado exige saldo disponível para a diferença.
- Insumo não exige reserva, mas exige `saldoFisico` suficiente.
- `saldoFisico` não pode ficar negativo em nenhuma hipótese.

**Processamento**

1. Verificar o `Idempotency-Key`.
2. Consultar o módulo de OS: o status é `EM_EXECUCAO`?
3. Abrir transação.
4. Carregar os itens com `SELECT ... FOR UPDATE`, ordenados por `item_id`.
5. Para cada peça:
   - Carregar a reserva ativa da OS.
   - Consumido menor ou igual ao reservado: `saldo_fisico -= consumido`, `saldo_reservado -= reservado`, devolvendo a diferença ao disponível.
   - Consumido maior que o reservado: conferir o disponível para a diferença; havendo, baixar; não havendo, abortar com `409`.
   - Marcar a reserva como `CONSUMIDA`.
6. Para cada insumo: conferir `saldo_fisico >= consumido` e baixar.
7. Inserir `movimentacao_estoque` do tipo `SAIDA` por linha, com o `os_id`.
8. Commit.
9. Publicar os eventos `PecaBaixada` e `InsumoConsumido`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`, `chave_idempotencia`, módulo de OS
- Altera: `item_estoque.saldo_fisico` e `item_estoque.saldo_reservado`, `reserva_estoque.status`, `movimentacao_estoque` (insert)
- A baixa do físico e a baixa da reserva ocorrem na mesma transação — é o ponto onde o estoque mais diverge se houver falha parcial.

**Saída da API**

```json
{
  "saidaId": "SAI-2026-0771",
  "ordemServicoId": "OS-2026-0912",
  "registradoEm": "2026-08-12T16:40:00-03:00",
  "itens": [
    { "itemId": "PC-0142", "consumido": 4, "reservado": 4, "devolvido": 0, "saldoFisicoAtual": 12 },
    { "itemId": "PC-0311", "consumido": 0, "reservado": 1, "devolvido": 1, "saldoFisicoAtual": 3 },
    { "itemId": "IN-0031", "consumido": 3.5, "reservado": 0, "devolvido": 0, "saldoFisicoAtual": 44.0 }
  ],
  "custoTotalMateriais": 512.10
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Saída registrada |
| `200` | Repetição de `Idempotency-Key` — retorna a resposta original |
| `400` | Body inválido; quantidade negativa; decimal incompatível com a unidade |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `estoque:movimentar` |
| `404` | OS ou item não encontrado |
| `409` | Peça sem reserva para a OS; saldo insuficiente para consumo acima do reservado |
| `422` | OS fora do status `EM_EXECUCAO` |

**Dependências**

- `ItemEstoqueRepository`
- `ReservaEstoqueRepository`
- `MovimentacaoEstoqueRepository`
- Módulo Ordem de Serviço — status e itens
- Serviço de idempotência
- Caso de uso Reservar Peça para OS (pré-requisito)

**Testes**

*Unitários*

- Consumo igual ao reservado: reserva zerada, nada devolvido.
- Consumo menor que o reservado: a diferença volta ao disponível.
- Consumo maior que o reservado com saldo: baixa a diferença.
- Consumo maior que o reservado sem saldo: aborta.
- Consumo zero: devolve toda a reserva.
- Insumo com decimal na unidade correta.

*Integração*

- Saída completa reduz físico e reservado juntos.
- Peça sem reserva retorna `409` e nada é alterado.
- OS em `AGUARDANDO_APROVACAO` retorna `422`.
- Repetição de `Idempotency-Key` não baixa duas vezes.
- `saldo_fisico` nunca fica negativo, mesmo com payload malicioso.

*Regressão*

- Após a saída, `saldo_fisico - saldo_reservado` continua igual ao disponível esperado.

---

### 6.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `baixar()` na entidade `ItemEstoque` reduzindo saldo físico e reservado juntos
- [ ] Implementar a invariante de saldo físico nunca negativo
- [ ] Implementar a regra de consumo menor que o reservado devolvendo a diferença ao disponível
- [ ] Implementar a regra de consumo maior que o reservado consumindo do disponível quando houver
- [ ] Implementar a transição da `ReservaEstoque` para `CONSUMIDA`

**Caso de uso**

- [ ] Implementar `RegistrarSaidaEstoque` cobrindo peças e insumos
- [ ] Implementar o cálculo de `custoTotalMateriais` da OS

**Repositório**

- [ ] Registrar `MovimentacaoEstoque` do tipo `SAIDA` vinculada ao `osId`

**Integrações**

- [ ] Consultar o módulo de OS para validar o status `EM_EXECUCAO`

**Handler HTTP**

- [ ] Implementar `POST /api/v1/estoque/saidas`

**Validações**

- [ ] Validar `quantidade` maior ou igual a zero
- [ ] Validar decimais compatíveis com a unidade de medida do item
- [ ] Validar que toda peça informada possui reserva ativa para a OS, retornando `409` caso contrário

**Transação e idempotência**

- [ ] Executar a baixa do saldo físico e da reserva na mesma transação
- [ ] Implementar `SELECT ... FOR UPDATE` ordenado por `item_id`
- [ ] Implementar a idempotência por item da OS

**Eventos**

- [ ] Publicar `PecaBaixada`
- [ ] Publicar `InsumoConsumido`

**Testes unitários**

- [ ] Consumo igual ao reservado: reserva zerada e nada devolvido
- [ ] Consumo menor que o reservado: diferença volta ao disponível
- [ ] Consumo maior que o reservado com saldo suficiente
- [ ] Consumo maior que o reservado sem saldo: aborta
- [ ] Consumo zero devolvendo toda a reserva

**Testes de integração**

- [ ] Saída reduzindo físico e reservado juntos
- [ ] Peça sem reserva retornando `409` sem alterar nada
- [ ] OS fora do status `EM_EXECUCAO` retornando `422`
- [ ] `Idempotency-Key` repetida não baixando duas vezes
- [ ] Saldo físico nunca ficando negativo, mesmo com payload malicioso
- [ ] Regressão: saldo físico menos saldo reservado continua igual ao disponível esperado

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, explicando o comportamento de `quantidade: 0`

**Review**

- [ ] Code Review aprovado

---


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
GET /api/v1/estoque/itens/faltantes
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
      "itemId": "PC-0311",
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
        { "id": "OS-2026-0930", "status": "AGUARDANDO_APROVACAO", "quantidade": 2 }
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

- [ ] Implementar `GET /api/v1/estoque/itens/faltantes`

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
POST   /api/v1/compras/pedidos
DELETE /api/v1/compras/pedidos/{pedidoId}
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
| Body | `fornecedorId` | string | Obrigatório e existente |
| Body | `prazoPrevistoEntrega` | date | Obrigatório, no futuro |
| Body | `confirmarDuplicidade` | boolean | Obrigatório quando já houver pedido em aberto para a mesma peça |
| Body | `itens[]` | array | Obrigatório, não vazio, sem `itemId` repetido |
| Body | `itens[].itemId` | string | Peça a comprar; item do tipo `INSUMO` é rejeitado |
| Body | `itens[].quantidade` | int | Inteiro maior que zero |
| Path (DELETE) | `pedidoId` | string | Pedido a cancelar |

```json
{
  "fornecedorId": "FOR-004",
  "prazoPrevistoEntrega": "2026-08-20",
  "confirmarDuplicidade": false,
  "itens": [
    { "itemId": "PC-0311", "quantidade": 5 },
    { "itemId": "PC-0142", "quantidade": 10 }
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
  "pedidoId": "PC-2026-0118",
  "numero": "2026/0118",
  "fornecedor": { "id": "FOR-004", "nome": "Auto Peças Recife" },
  "status": "ABERTO",
  "prazoPrevistoEntrega": "2026-08-20",
  "criadoEm": "2026-08-12T17:05:00-03:00",
  "criadoPor": "usr-018",
  "itens": [
    {
      "itemId": "PC-0311",
      "descricao": "Disco de freio ventilado",
      "quantidadePedida": 5,
      "quantidadeRecebida": 0
    },
    {
      "itemId": "PC-0142",
      "descricao": "Pastilha de freio dianteira",
      "quantidadePedida": 10,
      "quantidadeRecebida": 0
    }
  ],
  "ordensServicoVinculadas": ["OS-2026-0930"]
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

- [ ] Implementar `POST /api/v1/compras/pedidos`
- [ ] Implementar `DELETE /api/v1/compras/pedidos/{pedidoId}`

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
POST /api/v1/compras/pedidos
GET  /api/v1/estoque/insumos/{insumoId}/sugestao-compra
```

O `GET` auxiliar calcula a quantidade sugerida antes de montar o pedido.

> **Decisão de projeto.** Peça e insumo usam o mesmo recurso `pedido_compra` porque o ciclo de
> vida é idêntico (`ABERTO` → `PARCIAL` → `CONCLUIDO`) e o recebimento é o mesmo; o tipo do item
> diferencia as regras. Se o time preferir rotas separadas para espelhar os dois requisitos,
> `POST /api/v1/compras/pedidos/insumos` é aceitável — mas duplica a lógica de recebimento.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `compras:escrever` (leitura da sugestão: `estoque:ler`)

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Path (GET) | `insumoId` | string | Insumo para o qual a sugestão é calculada |
| Body | `fornecedorId` | string | Obrigatório e existente |
| Body | `prazoPrevistoEntrega` | date | Obrigatório, no futuro |
| Body | `confirmarDuplicidade` | boolean | Obrigatório quando já houver pedido em aberto para o mesmo insumo |
| Body | `itens[]` | array | Obrigatório, não vazio |
| Body | `itens[].itemId` | string | Insumo a comprar; item do tipo `PECA` é rejeitado |
| Body | `itens[].quantidade` | decimal | Maior que zero, com casas decimais compatíveis com a `unidadeMedida` |

```json
{
  "fornecedorId": "FOR-009",
  "prazoPrevistoEntrega": "2026-08-19",
  "confirmarDuplicidade": false,
  "itens": [{ "itemId": "IN-0031", "quantidade": 60.0 }]
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
  "itemId": "IN-0031",
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

- [ ] Implementar `GET /api/v1/estoque/insumos/{insumoId}/sugestao-compra`
- [ ] Reaproveitar `POST /api/v1/compras/pedidos` com validação por tipo de item

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

## Pontos em aberto

| # | Ponto | Responsável |
|---|---|---|
| 1 | `categoria` é string livre na entrada, mas o checklist pedia validação de enum. Definir se vira enum/tabela de categorias ou permanece texto livre. | — |
| 2 | `abaixoDoMinimo` está sendo comparado contra o saldo **disponível**. Confirmar se a regra do negócio é essa e não contra o saldo **físico**. | — |
| 3 | A consulta (RF-EST-01) não retorna `fabricante` nem `version`, mas a atualização exige `If-Match` com a `version`. Definir se o GET passa a expor esses campos ou se existe um GET de detalhe do item. | — |
| 4 | Sobreposição entre `409` e `422`: descrição duplicada é `409` e regra de negócio genérica é `422` (requisito 2); troca de unidade com saldo é `422` (requisito 3); saldo insuficiente é `409` e status inválido da OS é `422` (requisitos 5 e 6); item do tipo errado é `422` (requisitos 8 e 9). Padronizar qual código vale para qual situação, em todos os contextos. | — |
| 5 | Peça expõe `precoVenda` e insumo expõe `custoUnitario`, mas os dois vivem em `item_estoque` e são retornados pela mesma consulta (requisito 1), que só traz `precoVenda`. Definir como a consulta representa os dois tipos. | — |
| 6 | O requisito 2 bloqueia a inativação de peça com saldo reservado; o requisito 3 não trata inativação de insumo com saldo. Confirmar se a regra vale para insumo também. | — |
| 7 | `Idempotency-Key` é obrigatório no requisito 5 e apenas recomendado nos requisitos 4 e 6, embora os RNFs exijam idempotência nos três. Some-se a isso que o requisito 6 fala em idempotência **por item da OS**, enquanto o header é por requisição. Padronizar o mecanismo e o comportamento da repetição (`200` com a resposta original ou `409`). | — |
| 8 | A entrada grava `custoUnitario` por recebimento, mas o requisito 3 trata o custo como dado cadastral do insumo. Definir se a entrada atualiza o custo do item (e por qual critério: último custo ou média) ou se os dois custos são independentes. | — |
| 9 | O requisito 5 é o único que define corpo de erro, no formato Problem Details (`type`, `title`, `status`, `detail`). Padronizar esse formato para todos os erros de todos os contextos, ou removê-lo daqui. | — |
| 10 | Os requisitos 5, 6, 7 e 8 consultam o módulo de OS de forma síncrona, e o 5 também assina o evento `OrcamentoAprovado`. Definir qual das duas vias é a fonte da verdade — se o evento já carregar os dados necessários, a consulta síncrona deixa de ser necessária e o acoplamento entre contextos cai. | — |
| 11 | O requisito 6 confirma que insumo não é reservado e tem baixa direta na execução. Falta decidir se a consulta (requisito 1) continua exibindo `saldoReservado` para insumo, já que o campo será sempre zero. | — |
| 12 | O requisito 6 devolve `custoTotalMateriais` e o checklist coloca esse cálculo no caso de uso de estoque. Definir se apurar o custo dos materiais da OS pertence a este contexto ou ao de Ordem de Serviço, e qual valor entra na conta: `precoVenda` da peça, `custoUnitario` do insumo ou o custo do último recebimento. | — |
| 13 | A quantidade sugerida de compra tem duas fórmulas diferentes: no requisito 7 é `max(estoqueMinimo - saldoDisponivel, demandaNaoAtendida)` e no requisito 9 é `(consumoMedioDiario × leadTime) + estoqueMinimo − saldoDisponivel`. Definir qual vale para insumo e o que a tela mostra quando as duas divergem. | — |
| 14 | O `leadTimeDias` usado na sugestão do requisito 9 não tem origem definida: não existe campo de lead time no cadastro de fornecedor nem no de item. Definir onde ele é cadastrado e qual o valor padrão quando ausente. | — |
| 15 | O cadastro de `fornecedor` é pré-condição dos requisitos 8 e 9, mas nenhum requisito refina o CRUD de fornecedores. Definir se entra neste contexto, em Compras ou fica fora do MVP. | — |
