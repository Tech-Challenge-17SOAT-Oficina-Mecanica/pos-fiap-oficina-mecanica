---
documento: Decisões Arquiteturais — Pauta de Discussão
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-18
status: em discussao
---

# Decisões Arquiteturais

Este documento reúne as decisões que precisam ser tomadas em equipe. Cada uma nasceu de uma
dúvida real levantada durante o refinamento do contexto de Peças & Insumos
([`pecas-e-insumos-cd.md`](pecas-e-insumos-cd.md)), mas quase todas valem para os cinco
contextos — quem decidir sozinho vai gerar retrabalho no contexto do vizinho.

**Como usar:** cada decisão traz a pergunta, onde ela aparece, as opções com a consequência
de cada uma e uma recomendação já pronta. Na reunião, o time só precisa escolher e escrever a
linha **Decisão**. Depois de decidida, a mudança volta para o documento do contexto e o ponto
sai da lista de **Pontos em aberto**.

**Legenda de prioridade**

- **Bloqueante** — trava código já na primeira sprint.
- **Alta** — precisa estar decidido antes de fechar o contrato da API.
- **Média** — dá para decidir durante a implementação, sem retrabalho grande.

---

## Resumo

| # | Decisão | Prioridade | Recomendação |
|---|---|---|---|
| D-01 | Padrão de códigos HTTP: `409` x `422` | Bloqueante | Abandonar o `422`; usar `400` e `409` |
| D-02 | Idempotência das operações de saldo | Bloqueante | `Idempotency-Key` obrigatório nas três operações |
| D-03 | Formato do corpo de erro | Bloqueante | Problem Details (RFC 9457) em toda a API |
| D-04 | Consulta síncrona à OS x evento | Alta | Consulta síncrona como fonte da verdade; evento só dispara |
| D-05 | Quem apura o custo dos materiais da OS | Alta | Estoque devolve o custo da saída; OS consolida |
| D-06 | Fórmula da quantidade sugerida de compra | Alta | Uma função no domínio, com fórmula por tipo de item |
| D-07 | Origem do `leadTimeDias` | Alta | Campo no fornecedor, com padrão global configurável |
| D-08 | Cadastro de fornecedor no MVP | Alta | CRUD mínimo dentro do contexto de Peças & Insumos |
| D-09 | `categoria`: enum, tabela ou texto livre | Média | Tabela de categorias, validada por referência |
| D-10 | Campos ausentes na consulta de itens | Média | Expor `version` na listagem e criar GET de detalhe |
| D-11 | `precoVenda` x `custoUnitario` no mesmo recurso | Média | Manter os dois campos, nulos quando não se aplicam |
| D-12 | Base de cálculo do `abaixoDoMinimo` | Média | Saldo disponível |
| D-13 | Inativação de item com saldo | Média | Bloquear só com saldo reservado; físico pode inativar |
| D-14 | Custo do insumo atualizado pela entrada | Média | Último custo no MVP; média ponderada depois |
| D-15 | `saldoReservado` exibido para insumo | Média | Manter o campo, sempre zero |

---

## D-01 · Padrão de códigos HTTP: `409` x `422`

**Prioridade:** Bloqueante · **Ponto em aberto:** 4 · **Afeta:** todos os contextos

**Situação.** Hoje o mesmo tipo de problema recebe códigos diferentes: descrição duplicada é
`409` e regra de negócio genérica é `422` (requisito 2); troca de unidade com saldo é `422`
(requisito 3); saldo insuficiente é `409` e status inválido da OS é `422` (requisitos 5 e 6);
item do tipo errado é `422` (requisitos 8 e 9). Nenhum dev consegue acertar sem consultar tabela.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Sem `422`: `400` para entrada inválida (incluindo tipo de item errado) e `409` para qualquer conflito com o estado atual | Regra que se aplica sem pensar; menos expressiva para quem consome |
| B | Manter os três, com regra escrita: `400` formato, `422` conteúdo válido mas recusado pela regra, `409` conflito com outro registro ou com o estado | Mais expressivo; exige disciplina e revisão constante |

**Recomendação: A.** Para um MVP com cinco pessoas escrevendo handlers em paralelo, uma regra
sem zona cinzenta vale mais que expressividade. `409` cobre duplicidade, saldo insuficiente,
status incompatível e item inativo — todos são conflito com o estado atual.

**Decisão:**
**Data:**

---

## D-02 · Idempotência das operações que mexem em saldo

**Prioridade:** Bloqueante · **Ponto em aberto:** 7 · **Afeta:** Peças & Insumos, Ordem de Serviço

**Situação.** `Idempotency-Key` é obrigatório na reserva (requisito 5) e apenas "recomendado"
na entrada (4) e na saída (6), embora os três RNFs exijam idempotência. Pior: o requisito 6
fala em idempotência *por item da OS* e o header é *por requisição*. Sem decisão, uma retentativa
de rede duplica saldo.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | `Idempotency-Key` obrigatório nas três operações; a repetição devolve a resposta original com `200` | Mecanismo único, testável; exige a tabela `chave_idempotencia` desde a primeira sprint |
| B | Idempotência só por chave de negócio (`documentoOrigem` na entrada, reserva ativa na reserva) | Menos infraestrutura; não cobre a saída, que não tem chave natural |

**Recomendação: A**, com a chave de negócio permanecendo como segunda barreira (a constraint de
`documentoOrigem` continua). A idempotência "por item da OS" do requisito 6 sai do texto: quem
garante o não-duplicado é a chave da requisição somada ao status da reserva.

**Decisão:**
**Data:**

---

## D-03 · Formato do corpo de erro

**Prioridade:** Bloqueante · **Ponto em aberto:** 9 · **Afeta:** todos os contextos

**Situação.** Só o requisito 5 define corpo de erro, no formato Problem Details
(`type`, `title`, `status`, `detail`, mais uma lista `erros`). Os demais endpoints não dizem
o que devolvem no 4xx.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Problem Details (RFC 9457) em toda a API, com a lista `erros` para falhas por campo ou por item | Padrão conhecido, bom no Swagger; exige um handler global de exceção |
| B | Formato próprio e simples (`{ "mensagem": "...", "erros": [...] }`) | Menos cerimônia; cada time acaba inventando o seu |

**Recomendação: A**, implementado uma única vez em um handler global de exceções, compartilhado
pelos cinco contextos. É o tipo de coisa que só é barata se for feita antes de existirem 40 endpoints.

**Decisão:**
**Data:**

---

## D-04 · Consulta síncrona à OS x integração por evento

**Prioridade:** Alta · **Ponto em aberto:** 10 · **Afeta:** Peças & Insumos, Ordem de Serviço, Orçamento

**Situação.** Os requisitos 5, 6, 7 e 8 consultam o módulo de OS de forma síncrona para saber
status e orçamento; o requisito 5 ainda assina o evento `OrcamentoAprovado`. Existem duas vias
para a mesma informação e nenhuma foi eleita como fonte da verdade.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Consulta síncrona é a fonte da verdade; o evento só **dispara** o caso de uso, sem carregar dados de decisão | Simples e correto num monolito; acopla Estoque a OS em tempo de execução |
| B | O evento carrega tudo (itens, aprovação, status) e o Estoque nunca consulta OS | Desacopla de verdade; exige contrato de evento versionado e tolerância a evento atrasado |

**Recomendação: A.** O enunciado do Tech Challenge pede um monolito em camadas: dentro do mesmo
processo, a consulta síncrona é mais simples e não tem inconsistência temporal. O evento fica
como gatilho, e a decisão continua sendo tomada com dado lido na hora.

**Decisão:**
**Data:**

---

## D-05 · Quem apura o custo dos materiais da OS

**Prioridade:** Alta · **Ponto em aberto:** 12 · **Afeta:** Peças & Insumos, Ordem de Serviço, Orçamento

**Situação.** O requisito 6 devolve `custoTotalMateriais` e coloca o cálculo no caso de uso de
estoque. Mas o custo total do serviço é assunto da OS, que também tem mão de obra. Falta ainda
definir qual valor entra na conta.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Estoque devolve o custo **daquela saída**; a OS acumula e consolida o custo total do serviço | Cada contexto responde pelo que é seu; a OS vira o único lugar que sabe o custo do serviço |
| B | Estoque mantém o total acumulado por OS | Estoque passa a guardar estado da OS — vazamento de responsabilidade |

**Recomendação: A.** Sobre qual valor usar: `precoVenda` para peça (é o que o cliente paga) e
`custoUnitario` para insumo (é custo diluído, não cobrado item a item), ambos congelados no
momento da saída — coerente com RF-EST-11 e RNF-EST-15, que proíbem efeito retroativo.

**Decisão:**
**Data:**

---

## D-06 · Fórmula da quantidade sugerida de compra

**Prioridade:** Alta · **Ponto em aberto:** 13 · **Afeta:** Peças & Insumos

**Situação.** Existem duas fórmulas para o mesmo número. Requisito 7:
`max(estoqueMinimo - saldoDisponivel, demandaNaoAtendida)`. Requisito 9:
`(consumoMedioDiario × leadTime) + estoqueMinimo − saldoDisponivel`. Para insumo, as duas se
aplicam e vão divergir na tela.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Uma função de domínio `calcularQuantidadeSugerida(item)`, que aplica a fórmula de peça ou de insumo conforme o tipo; os dois endpoints chamam a mesma função | Um só lugar para mudar; a tela nunca mostra dois números diferentes |
| B | Manter as duas fórmulas, cada endpoint com a sua | Divergência visível para o usuário na primeira semana |

**Recomendação: A**, com a divisão explícita: **peça** usa demanda de OS mais reposição de
mínimo; **insumo** usa consumo médio, lead time e mínimo. A diferença é legítima — o que não
pode é o mesmo item receber dois números.

**Decisão:**
**Data:**

---

## D-07 · Origem do `leadTimeDias`

**Prioridade:** Alta · **Ponto em aberto:** 14 · **Afeta:** Peças & Insumos

**Situação.** A sugestão de compra do requisito 9 multiplica o consumo médio pelo lead time do
fornecedor, mas esse campo não existe em nenhum cadastro documentado.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | `leadTimeDias` no cadastro de fornecedor, com um padrão global configurável quando ausente | Simples; depende de D-08 (cadastro de fornecedor) |
| B | `leadTimeDias` por item, calculado do histórico de recebimentos | Mais preciso; sem histórico no MVP, não calcula nada |

**Recomendação: A**, com padrão de 7 dias quando o fornecedor não tiver o campo preenchido, e o
valor usado sempre devolvido na resposta da sugestão (como já está no contrato), para o usuário
entender de onde veio o número.

**Decisão:**
**Data:**

---

## D-08 · Cadastro de fornecedor no MVP

**Prioridade:** Alta · **Ponto em aberto:** 15 · **Afeta:** Peças & Insumos

**Situação.** Fornecedor é pré-condição dos requisitos 8 e 9 e origem do lead time (D-07), mas
nenhum requisito refina o CRUD. O enunciado do Tech Challenge não pede fornecedor explicitamente.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | CRUD mínimo de fornecedor dentro do contexto de Peças & Insumos (nome, documento, contato, `leadTimeDias`) | Fecha a dependência; custo baixo, é um CRUD simples |
| B | Fora do MVP: o pedido guarda o nome do fornecedor como texto livre | Menos código agora; perde o lead time e a consulta por fornecedor |

**Recomendação: A.** É um CRUD pequeno que destrava dois requisitos e uma fórmula. Vira um
requisito 10 no documento do contexto, seguindo o mesmo padrão dos demais.

**Decisão:**
**Data:**

---

## D-09 · `categoria`: enum, tabela ou texto livre

**Prioridade:** Média · **Ponto em aberto:** 1 · **Afeta:** Peças & Insumos

**Situação.** `categoria` é string livre na entrada da consulta, mas o checklist pedia validação
de enum. Além disso, a unicidade de descrição é validada **dentro da categoria** — ou seja, a
categoria é chave de regra de negócio, não só um filtro.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Tabela `categoria`, referenciada por id; item aponta para ela | Nova categoria sem deploy; filtros e relatórios confiáveis |
| B | Enum em código | Simples; qualquer categoria nova exige deploy |
| C | Texto livre | Zero esforço; "Freios", "freios" e "Freio" convivem e quebram a unicidade de descrição |

**Recomendação: A.** Como a categoria participa de uma invariante (descrição única por
categoria), texto livre é risco real de dado sujo.

**Decisão:**
**Data:**

---

## D-10 · Campos ausentes na consulta de itens

**Prioridade:** Média · **Ponto em aberto:** 3 · **Afeta:** Peças & Insumos

**Situação.** O `PUT` exige `If-Match` com a `version`, mas o `GET` de listagem não devolve
`version` nem `fabricante`. Hoje não há como o cliente montar uma atualização sem adivinhar.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Expor `version` na listagem e criar `GET /api/v1/estoque/itens/{itemId}` com o detalhe completo | Contrato coerente; um endpoint a mais |
| B | Colocar todos os campos na listagem | Sem endpoint novo; payload de lista mais pesado |

**Recomendação: A.** É o desenho usual de catálogo: lista enxuta para a tela de busca, detalhe
completo para a tela de edição.

**Decisão:**
**Data:**

---

## D-11 · `precoVenda` x `custoUnitario` no mesmo recurso

**Prioridade:** Média · **Ponto em aberto:** 5 · **Afeta:** Peças & Insumos

**Situação.** Peça tem `precoVenda`, insumo tem `custoUnitario`, os dois vivem em `item_estoque`
e saem pela mesma consulta — que hoje só traz `precoVenda`.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Manter os dois campos no contrato, nulos quando não se aplicam | Explícito, sem ambiguidade sobre o que o número significa |
| B | Um campo `valorUnitario` genérico | Payload menor; o consumidor não sabe se é preço de venda ou custo |

**Recomendação: A.** Preço de venda e custo são conceitos diferentes do negócio — juntá-los num
campo só é justamente o tipo de economia que confunde relatório depois.

**Decisão:**
**Data:**

---

## D-12 · Base de cálculo do `abaixoDoMinimo`

**Prioridade:** Média · **Ponto em aberto:** 2 · **Afeta:** Peças & Insumos

**Situação.** O sinalizador compara o **estoque mínimo** com o **saldo disponível**. Falta
confirmar se a regra do negócio não é comparar com o saldo físico.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Comparar com o saldo disponível | Alerta antecipado: peça reservada para outra OS já não conta como estoque |
| B | Comparar com o saldo físico | Alerta só quando a peça sai da prateleira — perto demais da ruptura |

**Recomendação: A**, coerente com RNF-EST-37, que manda calcular falta pelo disponível.

**Decisão:**
**Data:**

---

## D-13 · Inativação de item com saldo

**Prioridade:** Média · **Ponto em aberto:** 6 · **Afeta:** Peças & Insumos

**Situação.** O requisito 2 bloqueia inativar peça com saldo reservado. O requisito 3 não diz
nada sobre insumo — e insumo não tem reserva (decidido no requisito 6).

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Bloquear inativação apenas com saldo **reservado**; item com saldo físico pode ser inativado e some da consulta padrão | Regra única para os dois tipos; permite descontinuar item que ainda tem sobra |
| B | Bloquear também com saldo físico maior que zero | Impede descontinuar item encalhado sem antes dar baixa manual |

**Recomendação: A.** Inativar é decisão de catálogo ("não compramos mais"), não de prateleira.

**Decisão:**
**Data:**

---

## D-14 · Custo do insumo atualizado pela entrada

**Prioridade:** Média · **Ponto em aberto:** 8 · **Afeta:** Peças & Insumos

**Situação.** A entrada grava `custoUnitario` por recebimento, enquanto o requisito 3 trata o
custo como dado cadastral do insumo. Não está dito se um atualiza o outro.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | A entrada atualiza o custo do item com o **último custo** recebido, gravando `historico_preco_item` | Simples e auditável; oscila a cada compra |
| B | Custo médio ponderado, recalculado a cada entrada | Mais fiel ao custo real; exige recálculo e cuidado com arredondamento |
| C | Independentes: o cadastro só muda por edição manual | Custo do cadastro envelhece — é exatamente o problema descrito no requisito 3 |

**Recomendação: A** para o MVP, deixando **B** registrado como evolução. O histórico de preço já
existe, então trocar para média ponderada depois não exige migração de dado.

**Decisão:**
**Data:**

---

## D-15 · `saldoReservado` exibido para insumo

**Prioridade:** Média · **Ponto em aberto:** 11 · **Afeta:** Peças & Insumos

**Situação.** Insumo não é reservado (decidido no requisito 6), mas a consulta devolve
`saldoReservado` para todos os itens — para insumo, sempre zero.

**Opções**

| Opção | Regra | Consequência |
|---|---|---|
| A | Manter o campo, sempre zero para insumo | Contrato uniforme; cliente não precisa ramificar por tipo |
| B | Omitir o campo quando o item for insumo | Payload mais honesto; obriga o cliente a tratar campo ausente |

**Recomendação: A.** Campo estável vale mais que payload enxuto num contrato que cinco pessoas
vão consumir.

**Decisão:**
**Data:**

---

## Decisões já tomadas

Registradas aqui para não voltarem à pauta.

| # | Decisão | Onde está documentada |
|---|---|---|
| DT-01 | Não existe perfil `ESTOQUISTA`. As operações de estoque são feitas pelo `MECANICO` e pelo `GESTOR`, com o corte de permissão feito por **escopo** (`estoque:ler`, `estoque:escrever`, `estoque:movimentar`, `compras:escrever`), não por perfil. | Requisitos 2 a 9 de `pecas-e-insumos-cd.md` |
| DT-02 | Compras **não** é contexto delimitado separado: `pedido_compra` pertence ao contexto de Peças & Insumos, e o recebimento atualiza o pedido na mesma transação. | Decisão de projeto no requisito 8 de `pecas-e-insumos-cd.md` |
| DT-03 | Insumo não é reservado: a baixa acontece direto na execução do serviço. | Requisito 6 de `pecas-e-insumos-cd.md` |
| DT-04 | Peça e insumo compartilham o recurso `pedido_compra`; o tipo do item diferencia as regras. | Decisão de projeto no requisito 9 de `pecas-e-insumos-cd.md` |

---

## Depois da reunião

Para cada decisão fechada:

1. Escrever a linha **Decisão** e a **Data** aqui, mantendo as opções descartadas visíveis — elas explicam o porquê.
2. Aplicar a mudança no documento do contexto afetado, atualizando `versao` e `atualizado_em`.
3. Remover o ponto correspondente da tabela **Pontos em aberto**.
4. Se a decisão virar padrão para todos (D-01, D-02, D-03, D-04), levar também para o
   [`01-guia-de-documentacao.md`](01-guia-de-documentacao.md), seção 8.
