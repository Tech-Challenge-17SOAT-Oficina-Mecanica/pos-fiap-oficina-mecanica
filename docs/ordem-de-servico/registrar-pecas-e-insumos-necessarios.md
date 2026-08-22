---
documento: Refinamento de Requisitos — Registrar Peças e Insumos Necessários na OS
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Registrar Peças e Insumos Necessários na OS

Este documento detalha a tarefa Registrar Peças e Insumos Necessários na OS do contexto de Ordem
de Serviço.

> **Refinamento técnico e checklist pendentes.** O material recebido traz apenas o refinamento de
> produto. As seções 13.2 e 13.3 estão marcadas como pendentes, com a lista do que falta definir —
> o documento só fica completo, conforme o [guia](../01-guia-de-documentacao.md), quando elas forem
> escritas.

## 13 · Registrar Peças e Insumos Necessários na OS

### 13.1 Refinamento de Produto

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
- Os itens devem possuir reserva ativa ou pedido de compra vinculado à OS.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-129 | Permitir registrar as peças necessárias na OS. |
| RF-OS-130 | Permitir registrar os insumos necessários na OS. |
| RF-OS-131 | Tratar os dois registros como independentes: registrar peças não exige registrar insumos, e vice-versa. |
| RF-OS-132 | Vincular cada item registrado ao orçamento da OS. |
| RF-OS-133 | Criar o orçamento da OS na primeira adição, caso ainda não exista. |
| RF-OS-134 | Classificar como `PRINCIPAL` os itens registrados enquanto o orçamento ainda não foi aprovado, ou seja, enquanto a OS não entrou em execução. |
| RF-OS-135 | Classificar como `COMPLEMENTAR` os itens registrados com a OS já em execução. |
| RF-OS-136 | Permitir mais de um lançamento na adição principal, enquanto o orçamento não for aprovado. |
| RF-OS-137 | Identificar cada adição complementar separadamente, permitindo mais de uma ao longo da execução. |
| RF-OS-138 | Deixar cada adição complementar pendente de aprovação do cliente. |
| RF-OS-139 | Registrar quantidade e valor unitário de cada item no momento do registro. |
| RF-OS-140 | Calcular o valor de cada item e o valor da adição. |
| RF-OS-141 | Atualizar o valor do orçamento, mantendo separados o valor aprovado e o valor complementar pendente. |
| RF-OS-142 | Manter o histórico das adições, identificando o que entrou como principal e o que entrou como complementar. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-70 | A operação deve ser feita por API RESTful. |
| RNF-OS-71 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-OS-72 | O registro dos itens e a atualização do orçamento devem ocorrer na mesma operação. |
| RNF-OS-73 | O registro não altera saldo de estoque nem reservas — ambos já foram tratados nos fluxos de reserva e compra. |
| RNF-OS-74 | O valor unitário deve ser gravado no orçamento como cópia do valor vigente, para que alterações futuras de preço não mudem orçamentos já registrados. |
| RNF-OS-75 | A operação deve ser auditável, com registro de quem lançou e quando. |
| RNF-OS-76 | O tipo da adição deve ser determinado pelo sistema, nunca informado pelo usuário. |
| RNF-OS-77 | Itens de uma adição complementar não podem ser tratados como aprovados enquanto o cliente não reavaliar. |

**Fluxo Principal**

1. O mecânico acessa a OS e seleciona as peças ou os insumos a registrar.
2. O mecânico informa as quantidades necessárias.
3. O sistema valida a OS e o status em que ela se encontra.
4. O sistema valida os itens informados, o tipo, a situação e as quantidades.
5. O sistema verifica que cada item possui reserva ativa ou pedido de compra vinculado à OS.
6. O sistema localiza o orçamento da OS, ou o cria caso seja a primeira adição.
7. O sistema determina o tipo da adição: `PRINCIPAL` se a OS ainda não entrou em execução,
   `COMPLEMENTAR` se já está em execução.
8. O sistema registra os itens na OS com quantidade e valor unitário vigente.
9. O sistema vincula os itens à adição correspondente do orçamento.
10. O sistema recalcula os valores do orçamento.
11. Sendo adição complementar, o sistema a deixa pendente de aprovação e sinaliza que o cliente
    precisa reavaliar.
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
| A7 | Item sem reserva ativa nem pedido de compra vinculado à OS | Impede o registro e sinaliza que o fluxo de reserva ou compra ainda não foi executado para aquele item. |
| A8 | Item sem valor unitário vigente | Impede o registro. |
| A9 | Novo lançamento antes da aprovação do orçamento | Soma os itens à adição principal. |
| A10 | Novo lançamento com a OS em execução | Cria uma nova adição complementar, mesmo que já existam outras. |
| A11 | Item já registrado na OS | Aceita a nova quantidade dentro da adição vigente. |
| A12 | Usuário sem autorização | Impede a operação. |

**Saída**

- Relação dos itens registrados, com quantidade, valor unitário e valor do item.
- Identificação e tipo da adição: `PRINCIPAL` ou `COMPLEMENTAR`.
- Situação da adição: aprovada, no caso da principal já aprovada, ou pendente de aprovação, no
  caso da complementar.
- Valor da adição, valor aprovado do orçamento e valor complementar pendente.
- Ou indicação do motivo pelo qual o registro foi recusado.

**Pós-condições**

- As peças e os insumos registrados estão vinculados à OS e a uma adição do orçamento.
- A adição está classificada como `PRINCIPAL` ou `COMPLEMENTAR`.
- O valor unitário está congelado nos itens do orçamento.
- Os valores do orçamento refletem o que está aprovado e o que aguarda reavaliação do cliente.
- Sendo adição complementar, o orçamento fica com complemento pendente de aprovação.
- O saldo de estoque e as reservas permanecem inalterados.

---

### 13.2 Refinamento Técnico

**Pendente.** O material recebido não trouxe esta seção. Para fechá-la, falta definir:

- **Endpoint ou endpoints.** O fluxo alternativo A4 fala em "item de tipo divergente do endpoint
  utilizado", o que sugere rotas separadas para peça e insumo, algo como
  `POST /ordens-servico/{osId}/pecas` e `POST /ordens-servico/{osId}/insumos`. Confirmar se são
  duas rotas ou uma só com o tipo no corpo.
- **Contrato de entrada**: lista de itens com identificador e quantidade, e se a observação por
  item existe aqui como existe no registro de serviços.
- **Contrato de saída**: itens registrados, adição criada e os valores do orçamento separados
  entre aprovado e complementar pendente.
- **Códigos HTTP**, em especial o que responder quando o item não tem reserva nem pedido de compra
  vinculado à OS (exceção A7).
- **Persistência**: como a adição é representada (`orcamento_adicao`, citado no documento de
  retorno ao estoque), e onde ficam os itens de peça e insumo da OS.
- **Regra de determinação do tipo**: qual campo da OS ou do orçamento decide entre `PRINCIPAL` e
  `COMPLEMENTAR`, considerando que o contexto de Orçamento também usa `status` do orçamento para
  isso.

---

### 13.3 Checklist de Implementação

**Pendente.** Será escrito junto com o refinamento técnico, a partir dos requisitos e das regras
já definidos na seção 13.1.

---
