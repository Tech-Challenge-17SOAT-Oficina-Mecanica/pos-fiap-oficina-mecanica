---
documento: Decisões Arquiteturais — Pauta de Discussão
dono: José Lázaro
versao: 0.9
atualizado_em: 2026-08-22
status: em discussao
---

# Decisões Arquiteturais

Este documento reúne as decisões que precisam ser tomadas em equipe. Cada uma nasceu de uma
dúvida real levantada durante o refinamento do contexto de Peças & Insumos — hoje dividido em
[pecas/](pecas/) e [insumos/](insumos/) —, mas quase todas valem para os demais
contextos: quem decide sozinho gera retrabalho no contexto do vizinho.

**Como usar:** cada decisão traz a pergunta, onde ela aparece, as opções com a consequência
de cada uma e uma recomendação já pronta. Na reunião, o time só precisa escolher e escrever a
linha **Decisão**. Depois de decidida, a mudança volta para o documento do contexto e o ponto
sai da lista de **Pontos em aberto**.

**Legenda de prioridade**

- **Bloqueante** — trava código já na primeira sprint.
- **Alta** — precisa estar decidido antes de fechar o contrato da API.
- **Média** — dá para decidir durante a implementação, sem retrabalho grande.

**Situação em 22/08/2026.** As vinte e cinco decisões da pauta estão **fechadas**. O que a equipe
escolheu está escrito na linha **Decisão** de cada uma, com as opções descartadas preservadas —
elas explicam o porquê. As decisões que já vieram prontas do refinamento, sem passar por esta
pauta, estão na tabela **Decisões já tomadas**, no fim do documento.

---

## Resumo

| #    | Decisão                                                 | Prioridade | O que ficou decidido                                              |
| ---- | ------------------------------------------------------- | ---------- | ----------------------------------------------------------------- |
| D-01 | Padrão de códigos HTTP: `409` x `422`                   | Bloqueante | Sem `422`: `400` para entrada inválida, `409` para conflito       |
| D-02 | Idempotência das operações de saldo                     | Bloqueante | `Idempotency-Key` obrigatório em toda operação de saldo           |
| D-03 | Formato do corpo de erro                                | Bloqueante | Problem Details (RFC 9457) em toda a API                          |
| D-04 | Consulta síncrona à OS x evento                         | Alta       | Consulta síncrona; sem eventos e sem mensageria no projeto        |
| D-05 | Quem apura o custo dos materiais da OS                  | Alta       | Estoque devolve o custo da saída; a OS consolida                  |
| D-06 | Fórmula da quantidade sugerida de compra                | Alta       | Uma função no domínio, descontando o pedido em aberto             |
| D-07 | Origem do `leadTimeDias`                                | Alta       | `prazoEntregaDias` no fornecedor, padrão de 7 dias                |
| D-08 | Cadastro de fornecedor no MVP                           | Alta       | CRUD mínimo dentro do contexto de Peças                           |
| D-09 | `categoria`: enum, tabela ou texto livre                | Média      | Tabela de categorias, validada por referência                     |
| D-10 | Campos ausentes na consulta de itens                    | Média      | `version` nas listagens e rota de detalhe também para peça        |
| D-11 | `precoVenda` x `custoUnitario` no mesmo recurso         | Média      | Os dois campos, nulos quando não se aplicam                       |
| D-12 | Base de cálculo do `abaixoDoMinimo`                     | Média      | Saldo disponível                                                  |
| D-13 | Inativação de item com saldo                            | Média      | Bloqueia só com saldo reservado; físico livre pode inativar       |
| D-14 | Custo do insumo atualizado pela entrada                 | Média      | Último custo no MVP; média ponderada depois                       |
| D-15 | `saldoReservado` exibido para insumo                    | Média      | Campo real nos dois tipos: insumo é reservado (DT-03 revogada)    |
| D-16 | Reserva: três caminhos concorrentes                     | Bloqueante | Um fluxo só, disparado pela aprovação do orçamento                |
| D-17 | Orçamento complementar: orçamentos separados x adições  | Bloqueante | Orçamentos separados, com `tipoOrcamento` e `orcamentoOriginalId` |
| D-18 | Máquina de estados da OS                                | Bloqueante | Nove status, com a máquina desenhada no resumo do contexto        |
| D-19 | Situação de cadastro: `status` x `ativo`                | Alta       | Booleano `ativo`, mais data e usuário da desativação              |
| D-20 | Verbo da exclusão lógica: `DELETE` x `PATCH /desativar` | Alta       | `DELETE`, com rota de reativação onde faz sentido                 |
| D-21 | Envelope de listagem paginada                           | Alta       | `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`     |
| D-22 | Separação de peça e insumo nas rotas                    | Alta       | Rotas por tipo no catálogo; rota única nas operações comuns       |
| D-23 | Escopos de decisão do cliente                           | Média      | `orcamentos:decidir` nas duas rotas                               |
| D-24 | Controle otimista com `If-Match`                        | Média      | Obrigatório em toda atualização de cadastro                       |
| D-25 | Pagamento no MVP                                        | Média      | Fora do MVP; a entrega não bloqueia por pagamento                 |
| D-26 | Limites de `pagina` e `tamanho`                         | Média      | `pagina` padrão `0`; `tamanho` padrão `20`, teto `50` em toda a API |

---

## D-01 · Padrão de códigos HTTP: `409` x `422`

**Prioridade:** Bloqueante · **Ponto em aberto:** 4 · **Afeta:** todos os contextos

**Situação.** Hoje o mesmo tipo de problema recebe códigos diferentes: descrição duplicada é
`409` e regra de negócio genérica é `422` (requisito 2); troca de unidade com saldo é `422`
(requisito 3); saldo insuficiente é `409` e status inválido da OS é `422` (requisitos 5 e 6);
item do tipo errado é `422` (requisitos 8 e 9). Nenhum dev consegue acertar sem consultar tabela.

**Opções**

| Opção | Regra                                                                                                                                              | Consequência                                                       |
| ----- | -------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------ |
| A     | Sem `422`: `400` para entrada inválida (incluindo tipo de item errado) e `409` para qualquer conflito com o estado atual                           | Regra que se aplica sem pensar; menos expressiva para quem consome |
| B     | Manter os três, com regra escrita: `400` formato, `422` conteúdo válido mas recusado pela regra, `409` conflito com outro registro ou com o estado | Mais expressivo; exige disciplina e revisão constante              |

**Recomendação: A.** Para um MVP com cinco pessoas escrevendo handlers em paralelo, uma regra
sem zona cinzenta vale mais que expressividade. `409` cobre duplicidade, saldo insuficiente,
status incompatível e item inativo — todos são conflito com o estado atual.

**Decisão:** opção A. O `422` sai da API. `400` para entrada inválida — formato, campo faltando,
tipo de item errado — e `409` para qualquer conflito com o estado atual: duplicidade, saldo
insuficiente, status incompatível, registro inativo. A regra vale para todos os contextos e
entrou na seção 8 do guia.
**Data:** 2026-08-22

---

## D-02 · Idempotência das operações que mexem em saldo

**Prioridade:** Bloqueante · **Ponto em aberto:** 7 · **Afeta:** Peças & Insumos, Ordem de Serviço

**Situação.** `Idempotency-Key` é obrigatório na reserva (requisito 5) e apenas "recomendado"
na entrada (4) e na saída (6), embora os três RNFs exijam idempotência. Pior: o requisito 6
fala em idempotência _por item da OS_ e o header é _por requisição_. Sem decisão, uma retentativa
de rede duplica saldo.

**Opções**

| Opção | Regra                                                                                               | Consequência                                                                           |
| ----- | --------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| A     | `Idempotency-Key` obrigatório nas três operações; a repetição devolve a resposta original com `200` | Mecanismo único, testável; exige a tabela `chave_idempotencia` desde a primeira sprint |
| B     | Idempotência só por chave de negócio (`documentoOrigem` na entrada, reserva ativa na reserva)       | Menos infraestrutura; não cobre a saída, que não tem chave natural                     |

**Recomendação: A**, com a chave de negócio permanecendo como segunda barreira (a constraint de
`documentoOrigem` continua). A idempotência "por item da OS" do requisito 6 sai do texto: quem
garante o não-duplicado é a chave da requisição somada ao status da reserva.

**Decisão:** opção A. `Idempotency-Key` obrigatório nas operações que movimentam saldo — reserva,
processamento, entrada e baixa. A repetição devolve a resposta original. A chave de negócio
continua como segunda barreira.
**Data:** 2026-08-22

---

## D-03 · Formato do corpo de erro

**Prioridade:** Bloqueante · **Ponto em aberto:** 9 · **Afeta:** todos os contextos

**Situação.** Só o requisito 5 define corpo de erro, no formato Problem Details
(`type`, `title`, `status`, `detail`, mais uma lista `erros`). Os demais endpoints não dizem
o que devolvem no 4xx.

**Opções**

| Opção | Regra                                                                                           | Consequência                                                         |
| ----- | ----------------------------------------------------------------------------------------------- | -------------------------------------------------------------------- |
| A     | Problem Details (RFC 9457) em toda a API, com a lista `erros` para falhas por campo ou por item | Padrão conhecido, bom no Swagger; exige um handler global de exceção |
| B     | Formato próprio e simples (`{ "mensagem": "...", "erros": [...] }`)                             | Menos cerimônia; cada time acaba inventando o seu                    |

**Recomendação: A**, implementado uma única vez em um handler global de exceções, compartilhado
pelos cinco contextos. É o tipo de coisa que só é barata se for feita antes de existirem 40 endpoints.

**Decisão:** opção A. Problem Details (RFC 9457) em toda a API — `type`, `title`, `status`,
`detail` e a lista `erros` para falhas por campo ou por item —, implementado uma única vez em um
handler global de exceções.
**Data:** 2026-08-22

---

## D-04 · Consulta síncrona à OS x integração por evento

**Prioridade:** Alta · **Ponto em aberto:** 10 · **Afeta:** Peças & Insumos, Ordem de Serviço, Orçamento

**Situação.** Os requisitos 5, 6, 7 e 8 consultam o módulo de OS de forma síncrona para saber
status e orçamento; o requisito 5 ainda assina o evento `OrcamentoAprovado`. Existem duas vias
para a mesma informação e nenhuma foi eleita como fonte da verdade.

**Opções**

| Opção | Regra                                                                                                        | Consequência                                                                             |
| ----- | ------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------- |
| A     | Consulta síncrona é a fonte da verdade; o evento só **dispara** o caso de uso, sem carregar dados de decisão | Simples e correto num monolito; acopla Estoque a OS em tempo de execução                 |
| B     | O evento carrega tudo (itens, aprovação, status) e o Estoque nunca consulta OS                               | Desacopla de verdade; exige contrato de evento versionado e tolerância a evento atrasado |

**Recomendação: A.** O enunciado do Tech Challenge pede um monolito em camadas: dentro do mesmo
processo, a consulta síncrona é mais simples e não tem inconsistência temporal. O evento fica
como gatilho, e a decisão continua sendo tomada com dado lido na hora.

**Decisão:** consulta síncrona, e **sem evento de disparo**. Não há mensageria neste projeto:
onde havia publicação de evento, há chamada direta na mesma transação, e o histórico virou
trilha de auditoria.
**Data:** 2026-08-22

---

## D-05 · Quem apura o custo dos materiais da OS

**Prioridade:** Alta · **Ponto em aberto:** 12 · **Afeta:** Peças & Insumos, Ordem de Serviço, Orçamento

**Situação.** O requisito 6 devolve `custoTotalMateriais` e coloca o cálculo no caso de uso de
estoque. Mas o custo total do serviço é assunto da OS, que também tem mão de obra. Falta ainda
definir qual valor entra na conta.

**Opções**

| Opção | Regra                                                                                        | Consequência                                                                               |
| ----- | -------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------ |
| A     | Estoque devolve o custo **daquela saída**; a OS acumula e consolida o custo total do serviço | Cada contexto responde pelo que é seu; a OS vira o único lugar que sabe o custo do serviço |
| B     | Estoque mantém o total acumulado por OS                                                      | Estoque passa a guardar estado da OS — vazamento de responsabilidade                       |

**Recomendação: A.** Sobre qual valor usar: `precoVenda` para peça (é o que o cliente paga) e
`custoUnitario` para insumo (é custo diluído, não cobrado item a item), ambos congelados no
momento da saída — coerente com RF-EST-11 e RNF-EST-15, que proíbem efeito retroativo.

**Decisão:** opção A. A baixa apura e devolve o custo **daquela saída** (`custoTotalSaida`) e
informa o valor à OS; a OS acumula e consolida o custo total do serviço, material mais mão de
obra. Peça sai por `precoVenda`, insumo por `custoUnitario`, os dois congelados no momento da
baixa. Como a baixa pode acontecer mais de uma vez ao longo da execução, quem soma é a OS — o
estoque nunca guarda total por OS.
**Data:** 2026-08-22

---

## D-06 · Fórmula da quantidade sugerida de compra

**Prioridade:** Alta · **Ponto em aberto:** 13 · **Afeta:** Peças & Insumos

**Situação.** Existem duas fórmulas para o mesmo número. Requisito 7:
`max(estoqueMinimo - saldoDisponivel, demandaNaoAtendida)`. Requisito 9:
`(consumoMedioDiario × leadTime) + estoqueMinimo − saldoDisponivel`. Para insumo, as duas se
aplicam e vão divergir na tela.

**Opções**

| Opção | Regra                                                                                                                                                        | Consequência                                                        |
| ----- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------- |
| A     | Uma função de domínio `calcularQuantidadeSugerida(item)`, que aplica a fórmula de peça ou de insumo conforme o tipo; os dois endpoints chamam a mesma função | Um só lugar para mudar; a tela nunca mostra dois números diferentes |
| B     | Manter as duas fórmulas, cada endpoint com a sua                                                                                                             | Divergência visível para o usuário na primeira semana               |

**Recomendação: A**, com a divisão explícita: **peça** usa demanda de OS mais reposição de
mínimo; **insumo** usa consumo médio, lead time e mínimo. A diferença é legítima — o que não
pode é o mesmo item receber dois números.

**Decisão:** opção A — uma função de domínio única, em um ponto só. Com a saída da consulta de
itens faltantes (DT-50), o único consumidor que sobrou é o **processamento disparado pela
aprovação do orçamento**, que apura por item `quantidadeReservada` e `quantidadePendenteCompra` a
partir de `necessidade - saldoDisponivel`, nunca negativa, descontando o que já está pedido em
aberto. A fórmula com `estoqueMinimo` e consumo médio, que existia para a consulta de faltantes,
saiu junto com ela.
**Data:** 2026-08-22

---

## D-07 · Origem do `leadTimeDias`

**Prioridade:** Alta · **Ponto em aberto:** 14 · **Afeta:** Peças & Insumos

**Situação.** A sugestão de compra do requisito 9 multiplica o consumo médio pelo lead time do
fornecedor, mas esse campo não existe em nenhum cadastro documentado.

**Opções**

| Opção | Regra                                                                                      | Consequência                                         |
| ----- | ------------------------------------------------------------------------------------------ | ---------------------------------------------------- |
| A     | `leadTimeDias` no cadastro de fornecedor, com um padrão global configurável quando ausente | Simples; depende de D-08 (cadastro de fornecedor)    |
| B     | `leadTimeDias` por item, calculado do histórico de recebimentos                            | Mais preciso; sem histórico no MVP, não calcula nada |

**Recomendação: A**, com padrão de 7 dias quando o fornecedor não tiver o campo preenchido, e o
valor usado sempre devolvido na resposta da sugestão (como já está no contrato), para o usuário
entender de onde veio o número.

**Decisão:** opção A, com o nome em português: **`prazoEntregaDias`** no cadastro de fornecedor,
com padrão de 7 dias quando não vier preenchido. O campo informa quem compra; a fórmula de
sugestão fechada na D-06 não depende dele.
**Data:** 2026-08-22

---

## D-08 · Cadastro de fornecedor no MVP

**Prioridade:** Alta · **Ponto em aberto:** 15 · **Afeta:** Peças & Insumos

**Situação.** Fornecedor é pré-condição dos requisitos 8 e 9 e origem do lead time (D-07), mas
nenhum requisito refina o CRUD. O enunciado do Tech Challenge não pede fornecedor explicitamente.

**Opções**

| Opção | Regra                                                                                                      | Consequência                                                      |
| ----- | ---------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------- |
| A     | CRUD mínimo de fornecedor dentro do contexto de Peças & Insumos (nome, documento, contato, `leadTimeDias`) | Fecha a dependência; custo baixo, é um CRUD simples               |
| B     | Fora do MVP: o pedido guarda o nome do fornecedor como texto livre                                         | Menos código agora; perde o lead time e a consulta por fornecedor |

**Recomendação: A.** É um CRUD pequeno que destrava dois requisitos e uma fórmula. Vira um
requisito 10 no documento do contexto, seguindo o mesmo padrão dos demais.

**Decisão:** opção A. CRUD mínimo de fornecedor dentro do contexto de **Peças**, dono do agregado
de Compras: cadastrar, consultar, atualizar e desativar com reativação. Insumos apenas referencia.
O campo de prazo entrou com nome em português, `prazoEntregaDias` (D-07).
**Data:** 2026-08-22

---

## D-09 · `categoria`: enum, tabela ou texto livre

**Prioridade:** Média · **Ponto em aberto:** 1 · **Afeta:** Peças & Insumos

**Situação.** `categoria` é string livre na entrada da consulta, mas o checklist pedia validação
de enum. Além disso, a unicidade de descrição é validada **dentro da categoria** — ou seja, a
categoria é chave de regra de negócio, não só um filtro.

**Opções**

| Opção | Regra                                                         | Consequência                                                                           |
| ----- | ------------------------------------------------------------- | -------------------------------------------------------------------------------------- |
| A     | Tabela `categoria`, referenciada por id; item aponta para ela | Nova categoria sem deploy; filtros e relatórios confiáveis                             |
| B     | Enum em código                                                | Simples; qualquer categoria nova exige deploy                                          |
| C     | Texto livre                                                   | Zero esforço; "Freios", "freios" e "Freio" convivem e quebram a unicidade de descrição |

**Recomendação: A.** Como a categoria participa de uma invariante (descrição única por
categoria), texto livre é risco real de dado sujo.

**Decisão:** opção A. Tabela de categorias, referenciada por id, validada na entrada. Como a
categoria participa da invariante de descrição única, texto livre ficaria sujo em uma semana.
**Data:** 2026-08-22

---

## D-10 · Campos ausentes na consulta de itens

**Prioridade:** Média · **Ponto em aberto:** 3 · **Afeta:** Peças & Insumos

**Situação.** O `PUT` exige `If-Match` com a `version`, mas o `GET` de listagem não devolve
`version` nem `fabricante`. Hoje não há como o cliente montar uma atualização sem adivinhar.

**Opções**

| Opção | Regra                                                                                    | Consequência                                    |
| ----- | ---------------------------------------------------------------------------------------- | ----------------------------------------------- |
| A     | Expor `version` na listagem e criar `GET /estoque/itens/{itemId}` com o detalhe completo | Contrato coerente; um endpoint a mais           |
| B     | Colocar todos os campos na listagem                                                      | Sem endpoint novo; payload de lista mais pesado |

**Recomendação: A.** É o desenho usual de catálogo: lista enxuta para a tela de busca, detalhe
completo para a tela de edição.

**Decisão:** opção A. As listagens de peça e de insumo passam a expor `version`, e peça ganha
rota de detalhe, espelhando a que insumo já tem. Sem isso não há como montar o `If-Match` que a
D-24 tornou obrigatório no `PUT`.
**Data:** 2026-08-22

---

## D-11 · `precoVenda` x `custoUnitario` no mesmo recurso

**Prioridade:** Média · **Ponto em aberto:** 5 · **Afeta:** Peças & Insumos

**Situação.** Peça tem `precoVenda`, insumo tem `custoUnitario`, os dois vivem em `item_estoque`
e saem pela mesma consulta — que hoje só traz `precoVenda`.

**Opções**

| Opção | Regra                                                          | Consequência                                                      |
| ----- | -------------------------------------------------------------- | ----------------------------------------------------------------- |
| A     | Manter os dois campos no contrato, nulos quando não se aplicam | Explícito, sem ambiguidade sobre o que o número significa         |
| B     | Um campo `valorUnitario` genérico                              | Payload menor; o consumidor não sabe se é preço de venda ou custo |

**Recomendação: A.** Preço de venda e custo são conceitos diferentes do negócio — juntá-los num
campo só é justamente o tipo de economia que confunde relatório depois.

**Decisão:** opção A, com um efeito colateral da D-22: como peça e insumo passaram a ter recursos
próprios, não existe mais um payload compartilhado onde os dois campos convivam nulos. A regra
sobrevive na forma que interessa — **peça tem `precoVenda` e insumo tem `custoUnitario`, e nenhum
dos dois herda o campo do outro**, nem no cadastro, nem na consulta, nem no orçamento.
**Data:** 2026-08-22

---

## D-12 · Base de cálculo do `abaixoDoMinimo`

**Prioridade:** Média · **Ponto em aberto:** 2 · **Afeta:** Peças & Insumos

**Situação.** O sinalizador compara o **estoque mínimo** com o **saldo disponível**. Falta
confirmar se a regra do negócio não é comparar com o saldo físico.

**Opções**

| Opção | Regra                           | Consequência                                                              |
| ----- | ------------------------------- | ------------------------------------------------------------------------- |
| A     | Comparar com o saldo disponível | Alerta antecipado: peça reservada para outra OS já não conta como estoque |
| B     | Comparar com o saldo físico     | Alerta só quando a peça sai da prateleira — perto demais da ruptura       |

**Recomendação: A**, coerente com RNF-EST-37, que manda calcular falta pelo disponível.

**Decisão:** opção A. `abaixoDoMinimo = saldoDisponivel < estoqueMinimo`, nos dois contextos.
O que está reservado para outra OS não conta como estoque para efeito de alerta.
**Data:** 2026-08-22

---

## D-13 · Inativação de item com saldo

**Prioridade:** Média · **Ponto em aberto:** 6 · **Afeta:** Peças & Insumos

**Situação.** O requisito 2 bloqueia inativar peça com saldo reservado. O requisito 3 não diz
nada sobre insumo — e insumo não tem reserva (decidido no requisito 6).

**Opções**

| Opção | Regra                                                                                                                  | Consequência                                                                  |
| ----- | ---------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| A     | Bloquear inativação apenas com saldo **reservado**; item com saldo físico pode ser inativado e some da consulta padrão | Regra única para os dois tipos; permite descontinuar item que ainda tem sobra |
| B     | Bloquear também com saldo físico maior que zero                                                                        | Impede descontinuar item encalhado sem antes dar baixa manual                 |

**Recomendação: A.** Inativar é decisão de catálogo ("não compramos mais"), não de prateleira.

**Decisão:** opção A. A inativação é bloqueada apenas com **saldo reservado**; saldo físico livre
não impede inativar.
**Data:** 2026-08-22

---

## D-14 · Custo do insumo atualizado pela entrada

**Prioridade:** Média · **Ponto em aberto:** 8 · **Afeta:** Peças & Insumos

**Situação.** A entrada grava `custoUnitario` por recebimento, enquanto o requisito 3 trata o
custo como dado cadastral do insumo. Não está dito se um atualiza o outro.

**Opções**

| Opção | Regra                                                                                               | Consequência                                                                  |
| ----- | --------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------- |
| A     | A entrada atualiza o custo do item com o **último custo** recebido, gravando `historico_preco_item` | Simples e auditável; oscila a cada compra                                     |
| B     | Custo médio ponderado, recalculado a cada entrada                                                   | Mais fiel ao custo real; exige recálculo e cuidado com arredondamento         |
| C     | Independentes: o cadastro só muda por edição manual                                                 | Custo do cadastro envelhece — é exatamente o problema descrito no requisito 3 |

**Recomendação: A** para o MVP, deixando **B** registrado como evolução. O histórico de preço já
existe, então trocar para média ponderada depois não exige migração de dado.

**Decisão:** opção A. O custo do insumo é o **último custo de entrada** no MVP. Média ponderada
fica para quando existir histórico que justifique.
**Data:** 2026-08-22

---

## D-15 · `saldoReservado` exibido para insumo

**Prioridade:** Média · **Ponto em aberto:** 11 · **Afeta:** Peças & Insumos

**Situação.** Insumo não é reservado (decidido no requisito 6), mas a consulta devolve
`saldoReservado` para todos os itens — para insumo, sempre zero.

**Opções**

| Opção | Regra                                   | Consequência                                                  |
| ----- | --------------------------------------- | ------------------------------------------------------------- |
| A     | Manter o campo, sempre zero para insumo | Contrato uniforme; cliente não precisa ramificar por tipo     |
| B     | Omitir o campo quando o item for insumo | Payload mais honesto; obriga o cliente a tratar campo ausente |

**Recomendação: A.** Campo estável vale mais que payload enxuto num contrato que cinco pessoas
vão consumir.

**Decisão:** o campo é **real nos dois tipos**. A premissa da pergunta estava errada — o requisito
de reserva de insumo sempre existiu —, e a DT-03, que dizia que insumo não é reservado, foi
**revogada**. Insumo é reservado como peça, e a baixa acontece na execução, sobre a reserva.
**Data:** 2026-08-22

---

## D-16 · Reserva: três caminhos concorrentes

**Prioridade:** Bloqueante · **Afeta:** Peças & Insumos, Orçamento, Ordem de Serviço

**Situação.** Existem hoje quatro formas de comprometer estoque para uma OS, cada uma em um
documento diferente: reserva direta de peça (`POST /estoque/reservas`), reserva direta de insumo
(`POST /estoque/reservas-insumos`), processamento que reserva o disponível e compra o faltante
(`POST /estoque/solicitacoes-compra-reserva`) e o próprio pedido de compra, que também cria reserva.
Nenhum documento diz qual dispara qual, nem se elas se excluem.

**Opções**

| Opção | Regra                                                                                                                         | Consequência                                                                                 |
| ----- | ----------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------- |
| A     | Um único ponto de entrada: a aprovação do orçamento chama o processamento, que reserva o disponível e abre compra do faltante | Um fluxo só, alinhado ao que o negócio faz; as rotas de reserva direta viram detalhe interno |
| B     | Manter reserva direta e processamento como rotas separadas                                                                    | Flexível; exige regra explícita de quem chama o quê e abre espaço para reserva duplicada     |

**Recomendação: A.** O negócio tem um momento só em que o estoque é comprometido — quando o
cliente aprova. Ter três portas para a mesma coisa é a maior fonte de inconsistência do projeto hoje.

**Decisão:** opção A. Um único fluxo, disparado pela **aprovação do orçamento**, que chama
`ProcessarPecas` e `ProcessarInsumos` na mesma transação. As rotas de reserva direta perderam o
chamador público e foram aposentadas; a reserva virou serviço de domínio.
**Data:** 2026-08-22

---

## D-17 · Orçamento complementar: orçamentos separados ou adições

**Prioridade:** Bloqueante · **Afeta:** Orçamento, Ordem de Serviço

**Situação.** Duas modelagens convivem. Em Ordem de Serviço, o complementar é uma **adição** dentro
do mesmo orçamento (`orcamento_adicao`, com uma principal e várias complementares). Em Orçamento, o
complementar é um **orçamento separado**, com `tipo` e `orcamentoOriginalId`. As duas aparecem em
documentos aprovados, e o cálculo do valor total geral depende de qual vale.

**Opções**

| Opção | Regra                                                                | Consequência                                                                                |
| ----- | -------------------------------------------------------------------- | ------------------------------------------------------------------------------------------- |
| A     | Orçamentos separados por tipo, vinculados pelo `orcamentoOriginalId` | Cada decisão do cliente recai sobre um documento inteiro; mais simples de aprovar e recusar |
| B     | Um orçamento com várias adições                                      | Um único total; exige aprovação por adição, que não está desenhada                          |

**Recomendação: A**, que é o que a maioria dos documentos já usa e o que o fluxo de aprovação
espera. A ideia de adição pode voltar como agrupamento visual, sem virar entidade.

**Atualização de 2026-08-22.** O refinamento técnico de
[registrar-pecas-e-insumos-necessarios.md](ordem-de-servico/registrar-pecas-e-insumos-necessarios.md)
foi escrito seguindo a **opção B**, com `orcamento`, `orcamento_adicao` e `orcamento_item`, a
derivação do `tipoAdicao` pelo momento do registro e a situação `PENDENTE_APROVACAO` por adição —
inclusive resolvendo a aprovação por adição, que era o ponto fraco apontado na opção B. Agora as
duas modelagens têm refinamento técnico completo, em documentos do mesmo contexto. **A escolha
ficou mais urgente, não menos:** quem implementar primeiro obriga o outro lado a refazer.

**Decisão:** opção A. Orçamentos **separados por tipo** — a mesma tabela, mas dois registros —,
cada um com identificador próprio, `tipoOrcamento` (`PRINCIPAL` ou `COMPLEMENTAR`) e
`orcamentoOriginalId`. `orcamento_adicao` deixou de existir.
**Data:** 2026-08-22

---

## D-18 · Máquina de estados da Ordem de Serviço

**Prioridade:** Bloqueante · **Afeta:** todos os contextos

**Situação.** O enunciado lista seis status: `Recebida`, `Em diagnóstico`, `Aguardando aprovação`,
`Em execução`, `Finalizada` e `Entregue`. Os documentos usam nove, com `AGUARDANDO_RECURSOS`,
`AGUARDANDO_EXECUCAO` e `CANCELADA`. Não existe lista oficial nem diagrama de transições, e cada
tarefa valida o status que acha que precisa.

**Opções**

| Opção | Regra                                                                              | Consequência                                                                                               |
| ----- | ---------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| A     | Adotar os nove status e documentar as transições em um único lugar                 | Reflete o fluxo real, incluindo espera por peça; exige justificar a diferença na entrega do Tech Challenge |
| B     | Ficar nos seis do enunciado e tratar espera por recursos como atributo, não status | Alinhado ao enunciado; perde visibilidade de por que a OS está parada                                      |

**Recomendação: A**, com a máquina de estados desenhada no resumo do contexto de Ordem de Serviço
e citada no documento de entrega, explicando por que os três status extras existem.

**Decisão:** opção A. Os nove status são oficiais e a máquina de estados está desenhada em
[ordem-de-servico/00-resumo.md](ordem-de-servico/00-resumo.md), com as transições válidas e quem
dispara cada uma. Os três status além dos seis do enunciado existem por motivo concreto:
`AGUARDANDO_RECURSOS` mostra a OS parada esperando peça comprada, `AGUARDANDO_EXECUCAO` separa
orçamento aprovado de serviço começado, e `CANCELADA` é o destino da recusa do principal.
**Data:** 2026-08-22

---

## D-19 · Situação de cadastro: `status` ou `ativo`

**Prioridade:** Alta · **Afeta:** Serviços, Peças & Insumos, Cliente, Veículo

**Situação.** Cliente, Veículo e Estoque usam o booleano `ativo` com `dataDesativacao` e
`usuarioDesativacao`. Em Serviços, três documentos usam `status` com `ATIVO`/`INATIVO` e um usa
`ativo`. O mesmo conceito, escrito de duas formas, dentro do mesmo contexto.

**Recomendação:** booleano `ativo`, mais `dataDesativacao` e `usuarioDesativacao`. Se um dia
existir um terceiro estado, aí sim vira enum — mas hoje não existe.

**Decisão:** aprovada. O enum `status` saiu dos três documentos de Serviços que o usavam; o
projeto inteiro usa o booleano `ativo`.
**Data:** 2026-08-22

---

## D-20 · Verbo da exclusão lógica

**Prioridade:** Alta · **Afeta:** Serviços, Peças & Insumos, Cliente, Veículo

**Situação.** Cliente, Veículo e Estoque usam `DELETE /recurso/{id}` para a exclusão lógica.
Serviços usa `PATCH /servicos/{id}/desativar`. Os dois desenhos são defensáveis; conviver é que não.

**Opções**

| Opção | Regra                                                | Consequência                                                        |
| ----- | ---------------------------------------------------- | ------------------------------------------------------------------- |
| A     | `DELETE` em todos, com a exclusão lógica documentada | Contrato uniforme; o consumidor não precisa saber que é lógica      |
| B     | `PATCH .../desativar` em todos                       | Explícito de que o registro permanece; exige mudar quatro contextos |

**Recomendação: A**, que já é maioria e é o que o consumidor da API espera.

**Decisão:** aprovada. `PATCH /servicos/{servicoId}/desativar` foi aposentada em favor de
`DELETE /servicos/{servicoId}`, e Serviços ganhou `POST /servicos/{servicoId}/reativacao`, como
Cliente e Veículo.
**Data:** 2026-08-22

---

## D-21 · Envelope de listagem paginada

**Prioridade:** Alta · **Afeta:** todos os contextos

**Situação.** A maior parte das listagens usa `data`, `pagina`, `tamanho`, `totalElementos` e
`totalPaginas`. As de Serviços e de Orçamento usam `content`, `page`, `size`, `totalElements` e
`totalPages`. São dois contratos de paginação na mesma API.

**Recomendação:** o envelope em português, que já é maioria e está no guia. Alinhar as duas
listagens divergentes.

**Decisão:** aprovada a recomendação, com a regra explícita de que **recurso único devolve o objeto
direto, sem envelope**, e só listagem usa envelope. Registrada no guia, seção 8, e aplicada em
Cliente e Serviços. Falta Orçamento trocar `content`/`page`/`size`.
**Data:** 2026-08-22

---

## D-22 · Separação de peça e insumo nas rotas

**Prioridade:** Alta · **Afeta:** Peças & Insumos

**Situação.** Peça e insumo são a mesma entidade, diferenciada por `tipo`. Mesmo assim, hoje há
rotas separadas para consultar, reservar e processar, e rota compartilhada para comprar. A
consulta unificada `GET /estoque/itens` deixou de existir.

**Opções**

| Opção | Regra                                                      | Consequência                                                                 |
| ----- | ---------------------------------------------------------- | ---------------------------------------------------------------------------- |
| A     | Separar por tipo em toda a API                             | Contratos específicos, validações mais simples; mais rotas e mais duplicação |
| B     | Unificar tudo sob `/estoque/itens`, com `tipo` como filtro | Menos rotas; validações precisam ramificar por tipo                          |

**Recomendação: A**, porque as regras já são diferentes — insumo aceita fração e não tem preço de
venda —, mas então a compra também deve separar, para não ficar meio a meio.

**Decisão:** opção A. Rotas por tipo em toda a API: `/estoque/pecas` e `/estoque/insumos`, com a
consulta unificada removida. As operações que atendem os dois — entrada, saída e compra —
continuam em rota única, por serem a mesma operação de negócio.
**Data:** 2026-08-22

---

## D-23 · Escopos de decisão do cliente

**Prioridade:** Média · **Afeta:** Orçamento

**Situação.** Aprovar usa `orcamentos:aprovar`; recusar usa `orcamentos:recusar`. É a mesma pessoa,
no mesmo momento, decidindo sobre o mesmo documento.

**Recomendação:** um escopo só — `orcamentos:decidir` ou `orcamentos:aprovar` — porque quem pode
aprovar necessariamente pode recusar.

**Decisão:** a recomendação. Um escopo só, **`orcamentos:decidir`**, nas rotas de aprovar e de
recusar. A lista oficial de escopos do guia foi atualizada.
**Data:** 2026-08-22

---

## D-24 · Controle otimista com `If-Match`

**Prioridade:** Média · **Afeta:** todos os contextos

**Situação.** Peças & Insumos e Serviços usam `If-Match` com `version` na atualização de cadastro.
Cliente e Veículo não usam nada. Duas pessoas editando o mesmo cliente sobrescrevem uma à outra
sem aviso.

**Recomendação:** exigir `If-Match` em toda atualização de cadastro, com `version` exposto na
consulta de detalhe.

**Decisão:** aprovada. `PUT /clientes/{clienteId}`, `PUT /veiculos/{veiculoId}` e
`PATCH /servicos/{servicoId}` exigem `If-Match`, devolvem `412` quando diverge e `428` quando o
header não vem, e as consultas expõem `version`.
**Data:** 2026-08-22

---

## D-25 · Pagamento no MVP

**Prioridade:** Média · **Afeta:** Ordem de Serviço

**Situação.** A entrega do veículo exige confirmação de pagamento e apresenta o valor final ao
cliente, mas não existe contexto, rota, entidade nem refinamento de pagamento em lugar nenhum.

**Opções**

| Opção | Regra                                                                     | Consequência                                                               |
| ----- | ------------------------------------------------------------------------- | -------------------------------------------------------------------------- |
| A     | Fora do MVP: a entrega registra o valor final, sem bloquear por pagamento | Entrega implementável já; a oficina controla o recebimento fora do sistema |
| B     | Pagamento entra como contexto novo                                        | Fiel ao fluxo real; é escopo que o enunciado não pede                      |

**Recomendação: A**, com o campo de valor final mantido na entrega e a regra de bloqueio removida
até existir contexto de pagamento.

**Decisão:** opção A. Pagamento fica **fora do MVP**. A entrega apresenta e registra o valor final
acordado com o cliente, sem bloquear a retirada do veículo — o recebimento é controlado fora do
sistema.
**Data:** 2026-08-22

---

## D-26 · Limites de `pagina` e `tamanho`

**Prioridade:** Média · **Afeta:** todos os contextos

**Situação.** A D-21 padronizou os **nomes** do envelope paginado, mas não fixou os limites. Os
documentos ficaram com dois tetos diferentes para `tamanho`: **100** em Peças, Insumos e nas três
listagens de Ordem de Serviço, e **50** em Serviços (DT-23) e em Fornecedores. O padrão `20` e a
`pagina` iniciada em zero, por outro lado, já eram unânimes em todos os contextos. Como a
implementação do envelope é compartilhada, o teto variável exigiria passar o limite em cada
chamada e abriria espaço para um contexto herdar o teto errado por descuido.

**Opções**

| Opção | Regra | Consequência |
| ----- | ----- | ------------ |
| A     | Teto único de **50** para toda a API | Uma constante compartilhada; alinha 5 documentos ao teto que Serviços e Fornecedores já usavam |
| B     | Teto por contexto: 100 na maioria, 50 em Serviços e Fornecedores | Fiel ao que estava escrito; cada handler precisa declarar seu teto, e esquecer significa servir 100 onde deveria ser 50 |

**Recomendação: A.** O teto existe para limitar o custo da consulta, e não há caso de uso no
projeto que precise de mais de 50 itens por página — a diferença entre 50 e 100 não muda nenhuma
tela, mas dobra o pior caso de cada listagem.

**Decisão:** opção A. Para **toda listagem da API**: `pagina` inicia em zero, com padrão `0`;
`tamanho` tem padrão `20` e teto `50`, devolvendo `400` fora da faixa. Os cinco documentos que
declaravam teto 100 — [consultar-pecas.md](pecas/consultar-pecas.md),
[consultar-insumos.md](insumos/consultar-insumos.md),
[consultar-fila-de-atendimento.md](ordem-de-servico/consultar-fila-de-atendimento.md),
[listar-ordens-de-servico.md](ordem-de-servico/listar-ordens-de-servico.md) e
[monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) —
foram alinhados. Os limites vivem em constantes únicas no pacote de HTTP compartilhado, e a
validação é feita uma vez, junto do envelope.
**Data:** 2026-08-24

---

---

## Decisões já tomadas

Registradas aqui para não voltarem à pauta. A numeração é estável: a `DT-NN` é citada nos
documentos de contexto e nas tabelas de pontos em aberto.

| #     | Decisão | Onde está documentada |
| ----- | ------- | --------------------- |
| DT-01 | Não existem os perfis `ESTOQUISTA` nem `GESTOR`. Os perfis do sistema são `MECANICO`, `CLIENTE` e `SERVICO`, e o corte de permissão é feito por **escopo** (`estoque:ler`, `estoque:escrever`, `estoque:movimentar`, `compras:escrever`), não por perfil. | Todas as tarefas, seção *Autenticação / Autorização*                                                     |
| DT-02 | Compras **não** é contexto delimitado separado: `pedido_compra` pertence ao contexto de Peças & Insumos, e o recebimento atualiza o pedido na mesma transação. | [solicitar-compra-de-pecas.md](pecas/solicitar-compra-de-pecas.md) |
| DT-03 | ~~Insumo não é reservado: a baixa acontece direto na execução do serviço.~~ **Revogada em 2026-08-22:** insumo é reservado como peça, e a baixa acontece na execução, sobre a reserva. | [reservar-insumo-para-os.md](insumos/reservar-insumo-para-os.md) |
| DT-04 | Peça e insumo compartilham o recurso `pedido_compra`; o tipo do item diferencia as regras.                                                                                                                                                       | [solicitar-compra-de-insumos.md](insumos/solicitar-compra-de-insumos.md) |
| DT-05 | `POST /veiculos` foi **aposentada**. O cadastro de veículo acontece sempre por `POST /clientes/{clienteId}/veiculos`. O contrato em [cadastrar-veiculo.md](veiculo/cadastrar-veiculo.md) segue valendo como regra de negócio do cadastro. | [cadastrar-veiculo.md](veiculo/cadastrar-veiculo.md) e [00-resumo.md](veiculo/00-resumo.md) |
| DT-06 | O vínculo cliente-veículo devolve `201` quando é criado e `409` quando já existe. | [vincular-veiculo-ao-cliente.md](cliente/vincular-veiculo-ao-cliente.md) |
| DT-07 | Desvincular veículo fica **fora do MVP**, junto com a regra de um proprietário ativo por vez. | [vincular-veiculo-ao-cliente.md](cliente/vincular-veiculo-ao-cliente.md) |
| DT-08 | O cliente passa a ter contato: ao menos um entre `telefone` e `email` é obrigatório no cadastro e na atualização. | [cadastrar-cliente.md](cliente/cadastrar-cliente.md) e [atualizar-cliente.md](cliente/atualizar-cliente.md) |
| DT-09 | Anonimização e apagamento de dados pessoais ficam **fora do MVP**. O `DELETE` é sempre exclusão lógica. | [00-resumo.md](cliente/00-resumo.md) do contexto de Cliente |
| DT-10 | O vínculo cliente-veículo pertence ao contexto de **Cliente**; Veículo apenas referencia. | [vincular-veiculo-ao-cliente.md](cliente/vincular-veiculo-ao-cliente.md) |
| DT-11 | O perfil `GESTOR` **deixou de existir**. Os perfis do sistema são `MECANICO`, `CLIENTE` e `SERVICO`, e o corte de permissão continua sendo por escopo. | Seção *Autenticação / Autorização* de todas as tarefas e [00-visao-geral.md](00-visao-geral.md) |
| DT-12 | A placa aceita **os dois formatos**, Mercosul `ABC1D23` e antigo `ABC1234`, normalizada para maiúsculas, sem hífen e sem espaço. | [cadastrar-veiculo.md](veiculo/cadastrar-veiculo.md) e [atualizar-veiculo.md](veiculo/atualizar-veiculo.md) |
| DT-13 | O campo `ano` aceita de `1900` até o ano corrente mais um. | [cadastrar-veiculo.md](veiculo/cadastrar-veiculo.md) |
| DT-14 | A placa **pode ser corrigida** mesmo com OS existentes; a OS grava a placa vigente no momento da criação, para o histórico não mudar junto. | [atualizar-veiculo.md](veiculo/atualizar-veiculo.md) e [criar-ordem-de-servico.md](ordem-de-servico/criar-ordem-de-servico.md) |
| DT-15 | A operação de cadastrar e vincular exige **apenas** `veiculos:escrever`; não há validação cruzada contra o contexto de Cliente, porque a permissão vem nos escopos do JWT. | [cadastrar-veiculo-e-vincular-ao-cliente.md](veiculo/cadastrar-veiculo-e-vincular-ao-cliente.md) |
| DT-16 | Não existe histórico de proprietários no MVP: ao trocar o dono, o vínculo anterior não é preservado. | [00-resumo.md](veiculo/00-resumo.md) do contexto de Veículo |
| DT-17 | Serviços ganhou **reativação**: `POST /servicos/{servicoId}/reativacao`, como Cliente e Veículo. | [desativar-servico.md](servicos/desativar-servico.md) |
| DT-18 | A **remoção física** de serviço saiu do MVP. A exclusão é sempre lógica. | [desativar-servico.md](servicos/desativar-servico.md) |
| DT-19 | Unicidade do serviço por **nome normalizado** — sem acento, sem espaço duplo, minúsculo — apenas entre ativos, por índice parcial. | [cadastrar-servico.md](servicos/cadastrar-servico.md) |
| DT-20 | O `codigo` do serviço segue `SER-000001`, em sequência global, sem reset, com seis dígitos. | [cadastrar-servico.md](servicos/cadastrar-servico.md) |
| DT-21 | `tempoEstimadoMinutos` é obrigatório, com mínimo de 1 minuto. | [cadastrar-servico.md](servicos/cadastrar-servico.md) |
| DT-22 | `id`, `codigo` e `dataCriacao` são imutáveis; enviados no corpo do `PATCH`, retornam `400`. `PATCH` é atualização parcial de verdade. | [atualizar-servico.md](servicos/atualizar-servico.md) |
| DT-23 | Serviços inativos ficam fora da listagem por padrão; aparecem com `incluirInativos=true`. O teto de `tamanho` na paginação é **50** — teto que a D-26 estendeu depois a toda a API. | [consultar-servicos.md](servicos/consultar-servicos.md) e D-26 |
| DT-24 | O catálogo de serviços é restrito à oficina, perfil `MECANICO`. O cliente vê os serviços pelo orçamento. | [consultar-servicos.md](servicos/consultar-servicos.md) |
| DT-25 | Path param padronizado como `{servicoId}`, alinhado a `{clienteId}`, `{veiculoId}` e `{pecaId}`. | Todas as tarefas de Serviços |
| DT-26 | **Peças & Insumos virou dois contextos**: `docs/pecas/` e `docs/insumos/`, com prefixos de requisito `RF-PEC` e `RF-INS`. Os documentos que serviam aos dois tipos — entrada de estoque e retorno na recusa — foram duplicados e adaptados. | [00-resumo.md](pecas/00-resumo.md) e [00-resumo.md](insumos/00-resumo.md) |
| DT-27 | O recebimento é **uma rota só**, `POST /estoque/entradas`, compartilhada por peça e insumo. A divisão por tipo chegou a ser escrita e foi desfeita: receber é a mesma operação de negócio para os dois, e a nota fiscal costuma trazer peça e insumo juntos. | [registrar-entrada-de-pecas.md](pecas/registrar-entrada-de-pecas.md) e [registrar-entrada-de-insumos.md](insumos/registrar-entrada-de-insumos.md) |
| DT-28 | `consultar-estoque.md` foi renomeado para `consultar-pecas.md`: o arquivo sempre documentou a consulta de peças, não uma consulta unificada. | [consultar-pecas.md](pecas/consultar-pecas.md) |
| DT-29 | Aprovar orçamento **complementar** devolve a OS para a fila, igual ao principal. Recusar o complementar marca só aquele orçamento como `RECUSADO`, devolve seus itens ao estoque e mantém o serviço aprovado. Não há tarefas separadas para o complementar. | [aprovar-orcamento.md](orcamento/aprovar-orcamento.md) e [recusar-orcamento.md](orcamento/recusar-orcamento.md) |
| DT-30 | O cliente se autentica por **token de escopo reduzido**, emitido no envio do orçamento e válido apenas para aquela OS. Serve para consultar, aprovar e recusar. | [aprovar-orcamento.md](orcamento/aprovar-orcamento.md) |
| DT-31 | O `status` do **orçamento** é a fonte da verdade da decisão do cliente; o `status` da **OS**, a da etapa do atendimento. A transição da OS é consequência da decisão. | [calcular-orcamento.md](orcamento/calcular-orcamento.md) |
| DT-32 | O cálculo do orçamento é acionado **duas vezes**: ao fim do diagnóstico, antes do envio, e ao fechar cada complementar. Registrar item não recalcula sozinho. | [calcular-orcamento.md](orcamento/calcular-orcamento.md) |
| DT-33 | Não existe **recusa parcial** nem estado de **renegociação** no MVP. A decisão é sobre o orçamento inteiro; principal recusado cancela a OS. | [recusar-orcamento.md](orcamento/recusar-orcamento.md) |
| DT-34 | Enviar orçamento, gerar complementar, aprovar complementar e recusar complementar **deixaram de ser tarefas próprias**: o envio sai do fluxo do orçamento, o complementar é criado no registro de itens da OS, e as decisões usam as rotas de aprovar e recusar já existentes. | [00-resumo.md](orcamento/00-resumo.md) |
| DT-35 | **O projeto não usa mensageria nem eventos de domínio.** Onde havia publicação de evento há chamada direta na mesma transação: a inativação de cliente inativa os veículos, a entrada de estoque muda o status das OS, e o resto virou registro em trilha de auditoria. | Seção *Auditoria* das tarefas, no lugar da antiga seção *Eventos* |
| DT-36 | O `codigo` do item é **gerado pelo sistema**, em sequência global sem reset, com seis dígitos: `PEC-000001` e `INS-000001`. Os formatos `PC-0142`, `IN-0031`, `INS-0012` e `OLEO-001` saíram da documentação. | [cadastrar-peca.md](pecas/cadastrar-peca.md) e [cadastrar-insumo.md](insumos/cadastrar-insumo.md) |
| DT-37 | O item mantém **`nome` e `descricao`**, e os dois aparecem no cadastro, na consulta e na atualização. | Tarefas de cadastro, consulta e atualização de Peças e Insumos |
| DT-38 | Nenhum cadastro aceita **estoque inicial**. `saldoFisicoInicial` saiu do cadastro de insumo; todo saldo entra por movimentação de entrada. | [cadastrar-insumo.md](insumos/cadastrar-insumo.md) |
| DT-39 | Duplicidade por **descrição normalizada**, entre itens ativos: em Peças, dentro da categoria; em Insumos, dentro da categoria e da unidade de medida. Índice parcial na migration. | [cadastrar-peca.md](pecas/cadastrar-peca.md) e [cadastrar-insumo.md](insumos/cadastrar-insumo.md) |
| DT-40 | `ativo` **não é aceito** no `PUT` de peça nem no de insumo: a inativação acontece apenas pelo `DELETE`, onde estão as validações de saldo. | [atualizar-peca.md](pecas/atualizar-peca.md) e [atualizar-insumo.md](insumos/atualizar-insumo.md) |
| DT-41 | O **fornecedor é obrigatório** na compra, para peça e para insumo. | [solicitar-compra-de-pecas.md](pecas/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](insumos/solicitar-compra-de-insumos.md) |
| DT-42 | A compra pode ser **acima da necessidade** apurada: reserva-se para as OS apenas o necessário e o excedente fica como saldo livre. Comprar menos que a necessidade continua bloqueado. | [solicitar-compra-de-pecas.md](pecas/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](insumos/solicitar-compra-de-insumos.md) |
| DT-43 | A marcação de devolução e o vínculo com o pedido de compra vivem na **reserva**, não no item da OS. A devolução dispara em **qualquer transição da OS para `CANCELADA`**, não só na recusa. | [retornar-peca-ao-estoque.md](pecas/retornar-peca-ao-estoque.md) e [retornar-insumo-ao-estoque.md](insumos/retornar-insumo-ao-estoque.md) |
| DT-44 | O arredondamento de quantidade fracionária segue a precisão da `unidadeMedida`, sempre na mesma direção, em um único ponto do domínio. | [retornar-insumo-ao-estoque.md](insumos/retornar-insumo-ao-estoque.md) |
| DT-45 | `GET /estoque/itens`, a consulta unificada, foi **removida de propósito**: quem precisa de peça consulta `/estoque/pecas`, quem precisa de insumo consulta `/estoque/insumos`. | [03-endpoints.md](03-endpoints.md) |
| DT-46 | **Cada unidade de medida é um item de estoque independente, sem conversão.** O mesmo produto em `L` e em `ML` são dois cadastros, com saldos próprios. | [cadastrar-insumo.md](insumos/cadastrar-insumo.md) |
| DT-47 | As rotas de **reserva direta** foram aposentadas. `POST /estoque/reservas` e `POST /estoque/reservas-insumos` saíram da API: a reserva virou **serviço de domínio** chamado pelo processamento, que a aprovação do orçamento dispara. | [reservar-peca-para-os.md](pecas/reservar-peca-para-os.md) e [reservar-insumo-para-os.md](insumos/reservar-insumo-para-os.md) |
| DT-48 | O **CRUD de fornecedor foi escrito**, no contexto de Peças: `/fornecedores`, com documento imutável, unicidade entre ativos, exclusão lógica e reativação. Escopo novo `compras:ler` para a consulta. | [cadastrar-fornecedor.md](pecas/cadastrar-fornecedor.md) e seção 8 do [01-guia-de-documentacao.md](01-guia-de-documentacao.md) |
| DT-49 | A **baixa de consumo é uma rota só**, `POST /estoque/saidas`, compartilhada pelos dois contextos. Ela **consome a reserva**, nunca o saldo livre, e devolve ao saldo livre o reservado e não usado. | [registrar-consumo-e-saida-de-pecas.md](pecas/registrar-consumo-e-saida-de-pecas.md) e [registrar-consumo-e-saida-de-insumos.md](insumos/registrar-consumo-e-saida-de-insumos.md) |
| DT-50 | A **consulta de itens faltantes saiu do MVP**, nos dois contextos. Ela chegou a ser escrita para peças e foi removida: a falta é apurada dentro do processamento disparado pela aprovação do orçamento, que reserva o disponível e abre pedido do restante. | [03-endpoints.md](03-endpoints.md), seção *Rotas aposentadas* |
| DT-51 | As tarefas do contexto de **Ordem de Serviço foram renumeradas na ordem do fluxo** (1 a 13), e com elas os identificadores `RF-OS-NN` e `RNF-OS-NN`. Quem citar requisito de OS em outro documento precisa conferir o número. | [ordem-de-servico/00-resumo.md](ordem-de-servico/00-resumo.md) |
| DT-52 | A **baixa de estoque acontece durante a execução**, e pode ocorrer mais de uma vez na mesma OS. A finalização do serviço só passa com todas as reservas baixadas: se sobrar item pendente, responde `409` com a lista. | [iniciar-execucao.md](ordem-de-servico/iniciar-execucao.md) e [finalizar-servico.md](ordem-de-servico/finalizar-servico.md) |
| DT-53 | A **notificação de conclusão ao cliente é por e-mail**, com o resultado do envio gravado. Falha no envio não desfaz a finalização. | [finalizar-servico.md](ordem-de-servico/finalizar-servico.md) |
| DT-54 | O histórico devolvido na consulta da OS é **trilha de auditoria, não mensageria**. O campo passou a se chamar `eventos`, com os atributos em português (`agregado`, `agregadoId`, `tipoEvento`, `dados`, `metadados`, `ocorridoEm`, `registradoEm`). | [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| DT-55 | **Problema da OS não tem tipo próprio.** O que distingue o problema relatado do encontrado é o momento do registro; a consulta devolve `orcamentoId` para ligar o problema ao orçamento que o cobre. | [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| DT-56 | Peça e insumo podem ter **fornecedor padrão** em `item_estoque.fornecedor_id`, usado como origem para compra automática quando o item exigir reposição. | [cadastrar-peca.md](pecas/cadastrar-peca.md) e [cadastrar-insumo.md](insumos/cadastrar-insumo.md) |
| DT-56 | **Listagem e detalhe da OS são contratos distintos e ambos ficam:** a listagem é enxuta, para a tela de acompanhamento; o detalhe traz orçamentos, itens e trilha de auditoria. | [listar-ordens-de-servico.md](ordem-de-servico/listar-ordens-de-servico.md) e [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| DT-57 | As **duas consultas de orçamento convivem**, com papéis separados: `GET /ordens-servico/{osId}/orcamento` devolve todos os orçamentos daquela OS, e `GET /orcamentos` devolve um orçamento isolado ou os orçamentos de um cliente. | [registrar-pecas-e-insumos-necessarios.md](ordem-de-servico/registrar-pecas-e-insumos-necessarios.md) e [consultar-orcamento.md](orcamento/consultar-orcamento.md) |
| DT-58 | **Selecionar a próxima OS da fila** e **registrar problema adicional** saíram do escopo: não serão implementadas e foram removidas do checklist de cobertura. | [pontos-cobertos.md](../pontos-cobertos.md) |

---

## Depois da reunião

Para cada decisão fechada:

1. Escrever a linha **Decisão** e a **Data** aqui, mantendo as opções descartadas visíveis — elas explicam o porquê.
2. Aplicar a mudança no documento do contexto afetado, atualizando `versao` e `atualizado_em`.
3. Remover o ponto correspondente da tabela **Pontos em aberto**.
4. Se a decisão virar padrão para todos (D-01, D-02, D-03, D-04), levar também para o
   [01-guia-de-documentacao.md](01-guia-de-documentacao.md), seção 8.
