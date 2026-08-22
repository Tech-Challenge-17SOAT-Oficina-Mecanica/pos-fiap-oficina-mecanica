---
documento: Refinamento de Requisitos — Registrar Peças e Insumos Necessários na OS
dono: A definir
versao: 0.5
atualizado_em: 2026-08-22
status: em revisao
---

# Refinamento de Requisitos — Registrar Peças e Insumos Necessários na OS

Este documento detalha a tarefa Registrar Peças e Insumos Necessários na OS do contexto de Ordem
de Serviço.

## 5 · Registrar Peças e Insumos Necessários na OS

### 5.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar na Ordem de Serviço as peças e os insumos necessários à execução do serviço,
vinculando-os ao orçamento da OS e atualizando o valor do orçamento, distinguindo o que veio do
diagnóstico inicial do que foi encontrado com a OS já em execução.

**Problema**

O que o mecânico encontra no diagnóstico inicial é o que o cliente aprovou. Quando, no meio da
execução, aparece um problema novo, os itens correspondentes não podem se misturar ao que já foi
aprovado: eles precisam ser identificados separadamente para que o dono do veículo reavalie e
decida se autoriza o serviço adicional. Sem essa separação, o orçamento cresce sem explicação e a
oficina não consegue mostrar o que mudou nem por quê.

**Pré-condições**

- A OS deve existir e estar em um status que permita registro de itens.
- As peças e os insumos devem estar cadastrados e ativos.
- Os itens **não** precisam de reserva nem de pedido de compra: o comprometimento de estoque
  acontece depois, na aprovação do orçamento.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-39 | Permitir registrar as peças necessárias na OS. |
| RF-OS-40 | Permitir registrar os insumos necessários na OS. |
| RF-OS-41 | Tratar os dois registros como independentes: registrar peças não exige registrar insumos, e vice-versa. |
| RF-OS-42 | Vincular cada item registrado ao orçamento da OS. |
| RF-OS-43 | Criar o orçamento `PRINCIPAL` da OS no primeiro registro, caso ainda não exista. |
| RF-OS-44 | Registrar no orçamento `PRINCIPAL` os itens lançados enquanto ele não foi aprovado, ou seja, enquanto a OS não entrou em execução. |
| RF-OS-45 | Criar um orçamento `COMPLEMENTAR` para os itens lançados com a OS já em execução. |
| RF-OS-46 | Permitir mais de um lançamento no orçamento principal, enquanto ele não for aprovado. |
| RF-OS-47 | Identificar cada orçamento complementar separadamente, permitindo mais de um ao longo da execução. |
| RF-OS-48 | Deixar cada orçamento complementar em `CRIADO`, pendente de decisão do cliente. |
| RF-OS-49 | Registrar quantidade e valor unitário de cada item no momento do registro. |
| RF-OS-50 | Calcular o valor de cada item e o valor do orçamento afetado. |
| RF-OS-51 | Manter o valor total geral da OS como a soma do principal aprovado com os complementares. |
| RF-OS-52 | Manter o histórico dos orçamentos da OS, identificando principal e complementares pelo `tipoOrcamento`. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-17 | A operação deve ser feita por API RESTful. |
| RNF-OS-18 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-OS-19 | O registro dos itens e a atualização do orçamento devem ocorrer na mesma operação. |
| RNF-OS-20 | O registro não altera saldo de estoque nem reservas — ambos já foram tratados nos fluxos de reserva e compra. |
| RNF-OS-21 | O valor unitário deve ser gravado no orçamento como cópia do valor vigente, para que alterações futuras de preço não mudem orçamentos já registrados. |
| RNF-OS-22 | A operação deve ser auditável, com registro de quem lançou e quando. |
| RNF-OS-23 | O tipo do orçamento deve ser determinado pelo sistema, nunca informado pelo usuário. |
| RNF-OS-24 | Itens de um orçamento complementar não podem ser tratados como aprovados enquanto o cliente não reavaliar. |

**Fluxo Principal**

1. O mecânico acessa a OS e seleciona as peças ou os insumos a registrar.
2. O mecânico informa as quantidades necessárias.
3. O sistema valida a OS e o status em que ela se encontra.
4. O sistema valida os itens informados, o tipo, a situação e as quantidades.
5. O sistema verifica que cada item tem valor unitário vigente.
6. O sistema localiza o orçamento principal da OS, ou o cria caso seja o primeiro registro.
7. O sistema determina o orçamento de destino: o **principal**, se a OS ainda não entrou em
   execução, ou um **complementar novo**, se já está em execução.
8. O sistema registra os itens na OS com quantidade e valor unitário vigente.
9. O sistema vincula os itens ao orçamento correspondente.
10. O sistema recalcula o valor do orçamento e o valor total geral da OS.
11. Sendo orçamento complementar, o sistema o deixa em `CRIADO` e sinaliza que o cliente precisa
    reavaliar.
12. O sistema confirma o registro e devolve o orçamento atualizado.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | OS não encontrada | Impede a operação. |
| A2 | OS em status que não permite registro de itens, como finalizada, entregue ou cancelada | Impede a operação. |
| A3 | Item não encontrado ou inativo | Impede o registro do item. |
| A4 | Item de tipo divergente do endpoint utilizado | Impede o registro. |
| A5 | Quantidade menor ou igual a zero | Impede o registro. |
| A6 | Item repetido na mesma requisição | Impede o registro. |
| A7 | Item sem saldo em estoque | Registra normalmente: a falta é resolvida na aprovação do orçamento, que reserva o disponível e compra o faltante. |
| A8 | Item sem valor unitário vigente | Impede o registro. |
| A9 | Novo lançamento antes da aprovação do orçamento | Soma os itens ao orçamento principal. |
| A10 | Novo lançamento com a OS em execução | Cria um novo orçamento complementar, mesmo que já existam outros. |
| A11 | Item já registrado na OS | Aceita a nova quantidade dentro do orçamento vigente. |
| A12 | Usuário sem autorização | Impede a operação. |

**Saída**

- Relação dos itens registrados, com quantidade, valor unitário e valor do item.
- Identificação, tipo e situação do orçamento afetado: `PRINCIPAL` ou `COMPLEMENTAR`, em `CRIADO`
  ou `APROVADO`.
- Valor do orçamento afetado e valor total geral da OS.
- Ou indicação do motivo pelo qual o registro foi recusado.

**Pós-condições**

- As peças e os insumos registrados estão vinculados à OS e a um orçamento.
- O orçamento está classificado como `PRINCIPAL` ou `COMPLEMENTAR`.
- O valor unitário está congelado nos itens do orçamento.
- O valor total geral da OS reflete o principal mais os complementares.
- Sendo orçamento complementar, ele fica em `CRIADO`, esperando a decisão do cliente.
- O saldo de estoque e as reservas permanecem inalterados.

---

### 5.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/pecas
POST /ordens-servico/{osId}/insumos
GET  /ordens-servico/{osId}/orcamento
```

Os dois `POST` registram itens na OS e gravam no orçamento vigente — um para peça, outro
para insumo. O `GET` devolve os orçamentos da OS — o principal e seus complementares — com os
itens de cada um.

> **Decisão de projeto.** As duas rotas de consulta de orçamento **ficam**, com papéis distintos:
> `GET /ordens-servico/{osId}/orcamento` devolve **todos os orçamentos da OS** — o principal e seus
> complementares, com os itens de cada um —, e é a visão que a oficina e o cliente usam para
> acompanhar; `GET /orcamentos`, do contexto de Orçamento, devolve **um orçamento isolado** por
> identificador ou por documento do cliente, e é a que sustenta a decisão de aprovar e recusar.

> **Decisão de projeto.** O complementar é um **orçamento separado**, com identificador próprio,
> `tipoOrcamento` e `orcamentoOriginalId` apontando para o principal — e não uma adição dentro de
> um orçamento único (D-17). Peça e insumo **não** geram orçamentos distintos: cabem no mesmo
> orçamento vigente. Os endpoints são separados porque as regras de item divergem — insumo aceita
> fração conforme a unidade de medida e não tem preço de venda —, mas ambos gravam no mesmo
> orçamento. A alternativa de uma rota só, com o tipo do item no corpo, foi descartada por misturar
> duas validações diferentes no mesmo contrato.

> **Decisão de projeto.** O `tipoOrcamento` é **derivado do momento do registro**, nunca informado
> pelo cliente. OS que ainda não entrou em execução, ou seja, orçamento principal ainda não
> aprovado: os itens entram no **principal**, e vários lançamentos nesse período engordam o mesmo
> orçamento. OS já em execução: cada lançamento cria um **orçamento complementar novo**, em
> `CRIADO`, pendente de decisão do cliente.

> **Decisão de projeto.** O registro **não exige reserva nem pedido de compra**. A pré-condição
> anterior invertia o fluxo: o item precisava estar reservado para entrar na OS, mas a reserva só
> acontece quando o cliente aprova o orçamento — que por sua vez só existe depois deste registro.
> Com a D-16 fechada, a ordem ficou explícita: registrar item → calcular orçamento → cliente aprova
> → processamento reserva o disponível e compra o faltante. Aqui só se registra o que o serviço
> precisa; nada é comprometido no estoque.

> **Decisão de projeto.** Quem decide sobre o complementar são as mesmas rotas do principal:
> `POST /orcamentos/{orcamentoId}/aprovar` e `/recusar`. Aprovar um complementar devolve a OS para
> a fila; recusar marca aquele orçamento como `RECUSADO`, devolve os itens dele ao estoque e mantém
> o serviço já aprovado. Não existem tarefas separadas de aprovar e recusar complementar.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `os:escrever` nos dois `POST`; `os:ler` no `GET`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço, obrigatório. |
| Body | `itens[]` | array | Obrigatório; ao menos um item, sem repetição de `itemId`. |
| Body | `itens[].itemId` | uuid | Obrigatório; peça no endpoint de peças, insumo no de insumos. |
| Body | `itens[].quantidade` | decimal | Obrigatório; maior que zero. Em peça, inteiro; em insumo, com as casas decimais da `unidadeMedida` do item. |

```json
{
  "itens": [
    { "itemId": "3f1c92a7-58de-4b03-9c7a-6d24e05b91f8", "quantidade": 5 },
    { "itemId": "a7d4e1b0-2c65-4f38-91ab-8e50c7d63a24", "quantidade": 10 }
  ]
}
```

O mesmo formato vale para insumos, com `itemId` do tipo `INSUMO`.

**Validações**

*Técnicas*

- `osId` obrigatório e em formato UUID válido.
- `itens` não vazio, sem repetição de `itemId`.
- `quantidade` maior que zero.
- No endpoint de insumos, `quantidade` com casas decimais compatíveis com a `unidadeMedida` do
  item.

*Negócio*

- A OS deve existir e estar em status que permita registro de itens.
- Todos os itens existem e estão ativos.
- No endpoint de peças, todos os itens são do tipo `PECA`; no de insumos, do tipo `INSUMO`.
- O registro **não** exige reserva nem pedido de compra: o estoque é comprometido apenas na
  aprovação do orçamento.
- Cada item deve possuir valor unitário vigente.
- O `tipoOrcamento` é derivado pelo sistema e nunca aceito do cliente.
- Enquanto o orçamento principal não for aprovado, os lançamentos entram nele.
- Com a OS em execução, cada lançamento cria um orçamento complementar em `CRIADO`, vinculado ao
  principal por `orcamentoOriginalId`.
- O registro de peças e o de insumos são independentes entre si.

**Processamento**

1. Validar o payload e carregar a Ordem de Serviço.
2. Validar o status da OS.
3. Carregar e validar os itens, conferindo tipo, situação, quantidade e unidade de medida.
4. Validar que cada item tem valor unitário vigente.
5. Carregar o orçamento principal da OS ou criá-lo, caso seja o primeiro registro.
6. Determinar o `tipoOrcamento` a partir do estado da OS e da situação do orçamento principal.
7. Usar o orçamento principal, quando ele ainda não foi aprovado, ou criar um orçamento
   complementar novo, em `CRIADO`, quando a OS já está em execução.
8. Obter o valor unitário vigente de cada item.
9. Registrar os itens na OS com quantidade e valor unitário.
10. Vincular os itens ao orçamento correspondente.
11. Recalcular o valor do orçamento afetado e o valor total geral da OS.
12. Persistir as alterações na mesma transação.
13. Registrar o lançamento na trilha de auditoria.
14. Retornar o orçamento atualizado.

**Persistência**

- Consulta: `ordem_servico`, `item_estoque`, `orcamento`.
- Altera: `orcamento` (insert do principal no primeiro registro, insert de um complementar por
  lançamento durante a execução, update dos valores), `orcamento_item` (insert), itens necessários
  da OS (insert).
- Não altera: saldo físico, saldo reservado, reservas, pedidos de compra e status da OS.
- Tudo em **uma única transação**: registro dos itens e recálculo do orçamento não podem ficar
  parciais.

**Saída da API**

Registro no orçamento principal:

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "orcamento": {
    "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
    "tipoOrcamento": "PRINCIPAL",
    "statusOrcamento": "CRIADO"
  },
  "itensRegistrados": [
    {
      "itemId": "3f1c92a7-58de-4b03-9c7a-6d24e05b91f8",
      "codigo": "PEC-000311",
      "descricao": "Disco de freio ventilado",
      "tipo": "PECA",
      "quantidade": 5,
      "valorUnitario": 210.0,
      "valorItem": 1050.0
    }
  ],
  "valorOrcamento": 1050.0,
  "valorTotalGeral": 1050.0,
  "registradoEm": "2026-08-22T10:15:00-03:00",
  "registradoPor": "6b21f8d4-7e93-4c05-a1b8-40d6e2c93571"
}
```

Registro que cria um orçamento complementar:

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "orcamento": {
    "orcamentoId": "b83f5e27-4109-4ad6-92c3-7f16b0d84e59",
    "tipoOrcamento": "COMPLEMENTAR",
    "orcamentoOriginalId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
    "statusOrcamento": "CRIADO"
  },
  "itensRegistrados": [
    {
      "itemId": "a7d4e1b0-2c65-4f38-91ab-8e50c7d63a24",
      "codigo": "PEC-000142",
      "descricao": "Pastilha de freio dianteira",
      "tipo": "PECA",
      "quantidade": 10,
      "valorUnitario": 89.9,
      "valorItem": 899.0
    }
  ],
  "valorOrcamento": 899.0,
  "valorTotalGeral": 1949.0,
  "registradoEm": "2026-08-23T09:40:00-03:00",
  "registradoPor": "6b21f8d4-7e93-4c05-a1b8-40d6e2c93571"
}
```

No segundo exemplo o orçamento principal, de `1.050,00`, já foi aprovado; o complementar recém-criado
vale `899,00` e fica em `CRIADO`, esperando a decisão do cliente. O `valorTotalGeral` é a soma dos
orçamentos da OS.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Itens registrados e orçamento atualizado. |
| `200` | Orçamento consultado, no `GET`. |
| `400` | Body inválido, item repetido, quantidade menor ou igual a zero, decimal incompatível com a unidade de medida, ou item de tipo divergente do endpoint. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:escrever`, ou `os:ler` no `GET`. |
| `404` | Ordem de Serviço ou item não encontrado. |
| `409` | OS em status que não permite registro de itens; item inativo; item sem valor unitário vigente. |

> **Decisão de projeto — D-01.** O `422` saiu da API. Entrada inválida — incluindo item de tipo
> divergente do endpoint — é `400`; conflito com o estado atual é `409`.

**Dependências**

- `OrdemServicoRepository`.
- `OrcamentoRepository`.
- `ItemEstoqueRepository`.
- Trilha de auditoria.
- Caso de uso Aprovar Orçamento, destino tanto do principal quanto dos complementares.

**Testes**

*Unitários*

- Registro antes da aprovação gravado no orçamento principal.
- Segundo lançamento antes da aprovação somado ao mesmo orçamento principal.
- Registro com a OS em execução criando orçamento complementar.
- Segundo complementar criado com identificador próprio e `orcamentoOriginalId` correto.
- Orçamento complementar criado em `CRIADO`.
- Cálculo do valor do item e do valor do orçamento.
- Cálculo do valor total geral da OS a partir dos orçamentos.
- Rejeita quantidade zero.
- Rejeita item repetido na mesma requisição.
- Rejeita item de tipo divergente do endpoint.
- Aceita item sem saldo em estoque: a falta é resolvida na aprovação.
- Valor unitário congelado no item do orçamento.

*Integração*

- Registro de peças antes da aprovação retorna `201` e grava no orçamento `PRINCIPAL`.
- Registro de insumos retorna `201`.
- Registro de insumos sem peças registradas retorna `201`.
- Novo lançamento antes da aprovação mantém o mesmo orçamento principal.
- Registro com a OS em execução retorna `201` e cria um orçamento `COMPLEMENTAR` em `CRIADO`.
- Dois lançamentos durante a execução criam dois orçamentos complementares distintos.
- `valorTotalGeral` reflete a soma dos orçamentos da OS.
- Peça enviada ao endpoint de insumos retorna `400`.
- Item sem saldo em estoque retorna `201`: o registro não depende de estoque.
- Item inativo retorna `409`.
- OS inexistente retorna `404`.
- OS finalizada ou cancelada retorna `409`.
- Registro não altera saldo físico, saldo reservado, reservas nem status da OS.
- `GET` do orçamento retorna `200` com as adições e seus itens.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Implementar a entidade `Orcamento` com vínculo à OS, `tipoOrcamento`, `orcamentoOriginalId`, `statusOrcamento` e valor
- [ ] Implementar a entidade `OrcamentoItem` com item, quantidade, valor unitário e valor do item
- [ ] Implementar o enum `TipoOrcamento` com `PRINCIPAL` e `COMPLEMENTAR`
- [ ] Implementar a derivação do `tipoOrcamento` a partir do momento do registro
- [ ] Implementar o acúmulo de lançamentos no orçamento principal enquanto ele não for aprovado
- [ ] Implementar a criação de um orçamento complementar a cada lançamento durante a execução
- [ ] Criar o orçamento complementar em `CRIADO`, vinculado ao principal por `orcamentoOriginalId`
- [ ] Implementar o registro de peças necessárias na OS
- [ ] Implementar o registro de insumos necessários na OS
- [ ] Garantir que os dois registros sejam independentes entre si
- [ ] Implementar o congelamento do valor unitário no item do orçamento
- [ ] Implementar o cálculo do valor do item e do valor do orçamento
- [ ] Implementar o cálculo do valor total geral da OS a partir dos orçamentos
- [ ] Garantir que o registro não altera saldo de estoque, reservas nem status da OS

**Caso de uso**

- [ ] Implementar `RegistrarPecasNecessariasNaOrdemServico`
- [ ] Implementar `RegistrarInsumosNecessariosNaOrdemServico`
- [ ] Implementar `ConsultarOrcamentoDaOrdemServico`
- [ ] Implementar a criação do orçamento principal no primeiro registro
- [ ] Implementar a resolução do orçamento de destino: principal ou complementar novo

**Repositório**

- [ ] Implementar `OrcamentoRepository`
- [ ] Implementar a consulta do orçamento e das adições da OS
- [ ] Implementar a consulta do orçamento vigente da OS

**Integrações**

- [ ] Integrar com `OrdemServicoRepository`
- [ ] Integrar com `ItemEstoqueRepository`

**Handler HTTP**

- [ ] Implementar `POST /ordens-servico/{osId}/pecas`
- [ ] Implementar `POST /ordens-servico/{osId}/insumos`
- [ ] Implementar `GET /ordens-servico/{osId}/orcamento`
- [ ] Criar DTO/request de entrada
- [ ] Criar DTO/response de saída com o orçamento afetado, `valorOrcamento` e `valorTotalGeral`
- [ ] Validar o parâmetro `osId`
- [ ] Validar o payload
- [ ] Aplicar autenticação JWT nas rotas
- [ ] Aplicar autorização por escopo `os:escrever` e `os:ler`
- [ ] Mapear os erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar que a OS existe
- [ ] Validar que a OS está em status que permite registro de itens
- [ ] Validar que os itens existem e estão ativos
- [ ] Validar o tipo do item conforme o endpoint utilizado
- [ ] Validar quantidade maior que zero e sem repetição de item
- [ ] Validar decimais compatíveis com a unidade de medida, no endpoint de insumos
- [ ] Validar que cada item possui valor unitário vigente
- [ ] Ignorar `tipoOrcamento` informado pelo cliente

**Transação e idempotência**

- [ ] Executar o registro dos itens e o recálculo do orçamento na mesma transação
- [ ] Garantir que uma falha no recálculo desfaz o registro dos itens

**Auditoria**

- [ ] Registrar o lançamento na trilha de auditoria
- [ ] Devolver na resposta o orçamento complementar criado, para a oficina enviá-lo ao cliente

**Testes unitários**

- [ ] Registro antes da aprovação gravado no orçamento principal
- [ ] Segundo lançamento antes da aprovação somado ao mesmo orçamento principal
- [ ] Registro com a OS em execução criando orçamento complementar
- [ ] Complementares sucessivos com identificadores próprios e `orcamentoOriginalId` correto
- [ ] Orçamento complementar criado em `CRIADO`
- [ ] Cálculo do valor do item e do valor do orçamento
- [ ] Cálculo do valor total geral da OS
- [ ] Rejeição de quantidade zero
- [ ] Rejeição de item repetido na mesma requisição
- [ ] Rejeição de item de tipo divergente do endpoint
- [ ] Registro aceito para item sem saldo em estoque
- [ ] Congelamento do valor unitário no item do orçamento

**Testes de integração**

- [ ] Registro de peças antes da aprovação retornando `201` no orçamento `PRINCIPAL`
- [ ] Registro de insumos retornando `201`
- [ ] Registro de insumos sem peças registradas retornando `201`
- [ ] Novo lançamento antes da aprovação mantendo o mesmo orçamento principal
- [ ] Registro com a OS em execução retornando `201` com orçamento `COMPLEMENTAR` em `CRIADO`
- [ ] Dois lançamentos durante a execução criando dois orçamentos complementares
- [ ] `valorTotalGeral` correto após os lançamentos
- [ ] Peça no endpoint de insumos retornando `400`
- [ ] Item sem saldo em estoque retornando `201`
- [ ] Item inativo retornando `409`
- [ ] OS inexistente retornando `404`
- [ ] OS finalizada ou cancelada retornando `409`
- [ ] Nenhum saldo, reserva ou status de OS alterado após o registro
- [ ] `GET` do orçamento retornando `200` com as adições e seus itens

**Documentação**

- [ ] Documentar os três endpoints no Swagger/OpenAPI, incluindo a regra de derivação do
      `tipoOrcamento` e o vínculo do complementar com o principal

**Review**

- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto
- [ ] Executar testes automatizados
- [ ] Validar os critérios de aceite da tarefa
- [ ] Code Review aprovado

---

## Pontos em aberto

| # | Ponto | Responsável |
|---|---|---|

---
