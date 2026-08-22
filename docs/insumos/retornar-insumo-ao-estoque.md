---
documento: Refinamento de Requisitos — Retornar Insumo ao Estoque
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Retornar Insumo ao Estoque

Este documento detalha a tarefa Retornar Insumo ao Estoque do contexto de Insumos.

> **Escopo deste documento.** A devolução é disparada uma vez por recusa de orçamento e percorre
> todos os itens da OS. Este documento descreve a parte que toca **insumos**; a parte de peças
> está em [retornar-peca-ao-estoque.md](../pecas/retornar-peca-ao-estoque.md). O caso de uso é o mesmo e roda
> na mesma transação: o que muda são as regras de saldo e as validações de cada tipo de item.

## 9 · Retornar Insumo ao Estoque

### 9.1 Refinamento de Produto

**Persona**

Sistema, dentro do fluxo de recusa do orçamento. Beneficiário: o mecânico, que volta a enxergar
os insumos disponíveis para outras OS.

**Objetivo**

Devolver ao estoque os insumos vinculados à OS quando o orçamento é recusado, liberando as reservas e
retornando o que já havia sido baixado.

**Problema**

Quando o cliente recusa o orçamento, o serviço não vai acontecer, mas os insumos continuam
comprometidos: a reserva segue de pé e o que já saiu do estoque não volta. O resultado é estoque
que existe mas aparece como indisponível, outra OS parando por falta de um item que está na
prateleira, e compra sendo solicitada sem necessidade.

**Gatilho**

Chamada direta do caso de uso de recusa do orçamento, logo após o cancelamento da OS. Não há
endpoint próprio, tela própria nem acionamento manual.

**Pré-condições**

- A OS deve existir e estar carregada pelo caso de uso chamador.
- A OS deve possuir insumos vinculados.
- A operação ocorre dentro da transação já aberta pela recusa.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-INS-77 | Identificar os insumos a devolver. |
| RF-INS-78 | Identificar quais itens estão reservados e quais já foram baixados do estoque. |
| RF-INS-79 | Liberar as reservas ativas desses itens. |
| RF-INS-80 | Reduzir o saldo reservado na quantidade liberada. |
| RF-INS-81 | Retornar ao saldo físico as quantidades já baixadas. |
| RF-INS-82 | Desvincular da OS os itens vinculados apenas a pedido de compra ainda não recebido. |
| RF-INS-83 | Registrar as movimentações de liberação e de retorno no histórico de estoque. |
| RF-INS-84 | Marcar os itens da OS como devolvidos. |
| RF-INS-85 | Ignorar os itens já marcados como devolvidos. |
| RF-INS-86 | Devolver ao caso de uso chamador o resultado do que foi liberado, retornado e desvinculado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-INS-50 | A devolução deve ocorrer dentro da transação aberta pelo caso de uso chamador, sem abrir transação própria. |
| RNF-INS-51 | O saldo reservado nunca pode ficar negativo. |
| RNF-INS-52 | A operação não deve alterar o status da OS, o orçamento nem os itens do orçamento. |
| RNF-INS-53 | As movimentações registradas devem ser auditáveis e imutáveis. |
| RNF-INS-54 | A operação deve ser protegida contra concorrência, com lock de linha nos itens. |

**Fluxo Principal**

1. O caso de uso de recusa cancela a OS e chama a devolução.
2. O sistema carrega os itens vinculados à OS.
3. O sistema separa os itens reservados, os já baixados e os pendentes de compra.
4. O sistema libera as reservas ativas e reduz o saldo reservado.
5. O sistema devolve ao saldo físico as quantidades já baixadas.
6. O sistema desvincula os itens pendentes de compra.
7. O sistema registra as movimentações de liberação e de retorno.
8. O sistema marca os itens da OS como devolvidos.
9. O sistema retorna o resultado para o caso de uso chamador.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | OS sem itens vinculados | Conclui sem devolver nada e informa que não havia itens. |
| A2 | Todos os itens apenas reservados | Libera as reservas e não altera o saldo físico. |
| A3 | Todos os itens já baixados | Registra o retorno e aumenta o saldo físico. |
| A4 | Item vinculado somente a pedido de compra não recebido | Apenas desvincula e sinaliza que não havia o que devolver. |
| A5 | Item já devolvido | Ignora e segue com os demais. |
| A6 | Item inativado após o vínculo com a OS | Conclui a devolução e sinaliza o item como inativo. |
| A7 | Erro durante a devolução | Propaga o erro e a transação da recusa é desfeita por inteiro. |

**Saída**

- Relação das reservas liberadas, com item e quantidade.
- Relação dos itens retornados ao estoque, com item e quantidade.
- Relação dos itens desvinculados sem devolução, com o motivo.
- Itens da OS marcados como devolvidos.

**Pós-condições**

- Nenhuma reserva ativa dos itens devolvidos permanece.
- O saldo reservado dos itens reservados foi reduzido e o saldo físico dos itens já baixados foi
  restabelecido.
- Os itens pendentes de compra não estão mais vinculados à OS.
- As movimentações de liberação e de retorno estão registradas no histórico.
- A OS recusada não mantém nenhum item comprometendo o estoque.

---

### 9.2 Refinamento Técnico

**Gatilho**

Não há endpoint: é uma chamada em processo, dentro de `RecusarOrcamento`.

```
RecusarOrcamento
├── valida o orçamento e a recusa
├── altera o status da OS para CANCELADA
├── DevolverItensAoEstoque(ordemServico)   ← esta tarefa
└── confirma a transação
```

> **Decisão de projeto.** A marcação de devolução e o vínculo com o pedido de compra vivem na
> **reserva** (`reserva_estoque`), não no item da OS. A reserva já é quem conhece a quantidade
> comprometida, a OS e o pedido de origem; guardar a marcação no item da OS obrigaria o outro
> contexto a escrever num agregado que não é dele.

> **Decisão de projeto.** A devolução é disparada em **qualquer transição da OS para
> `CANCELADA`**, e não só na recusa do orçamento. Cancelamento por outro motivo deixava o insumo
> reservado para sempre, com o saldo disponível sumindo sem explicação.

> **Decisão de projeto.** Sem endpoint, sem evento e sem mensageria: a entrada e a saída são
> objetos de domínio, não payloads HTTP. O commit é do caso de uso chamador, e qualquer exceção
> aqui desfaz também a recusa e o cancelamento da OS. A alternativa — devolver por evento, depois
> do commit — abriria uma janela em que a OS está cancelada e o estoque continua comprometido.

Assinatura sugerida:

```
ResultadoDevolucao devolverItensAoEstoque(OrdemServico ordemServico)
```

**Autenticação / Autorização**

Não se aplica — a autorização já foi verificada pelo caso de uso de recusa do orçamento, que é
quem expõe o endpoint.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Interno | `ordemServico` | agregado | Ordem de Serviço já carregada pelo caso de uso chamador, dentro da transação aberta por ele. |

**Validações**

*Técnicas*

- A OS deve estar carregada e a transação, aberta pelo chamador.

*Negócio*

- A OS deve possuir insumos vinculados; sem insumos, o resultado volta vazio, sem erro.
- Quantidades de insumo são **fracionárias**: a devolução respeita as casas decimais da
  `unidadeMedida` do item, e o arredondamento nunca pode criar nem destruir saldo.
- Só são liberadas reservas com status `ATIVA` vinculadas à OS.
- Só retornam ao estoque as quantidades efetivamente baixadas.
- Itens vinculados apenas a pedido de compra não recebido são desvinculados sem movimentação de estoque.
- Itens já marcados como devolvidos são ignorados.
- A liberação de reserva nunca pode deixar o saldo reservado negativo.

> **Decisão de projeto.** O arredondamento de quantidade fracionária segue a **precisão da
> `unidadeMedida`** do item, sempre na mesma direção, em um único ponto do domínio. Sem regra
> única, dois cálculos do mesmo consumo divergem na terceira casa e o saldo nunca fecha.

**Processamento**

1. Carregar os itens vinculados à OS, ordenados por `item_id`, com lock de linha.
2. Descartar os itens já marcados como devolvidos.
3. Separar os itens em reservados, baixados e pendentes de compra.
4. Liberar as reservas ativas, alterando o status para `LIBERADA`.
5. Reduzir o `saldoReservado` dos itens liberados.
6. Inserir movimentação de estoque do tipo `LIBERACAO_RESERVA`.
7. Aumentar o `saldoFisico` dos itens já baixados.
8. Inserir movimentação de estoque do tipo `ENTRADA_RETORNO`.
9. Desvincular os itens pendentes de compra.
10. Marcar os itens como devolvidos ao estoque.
11. Retornar o `ResultadoDevolucao` para o caso de uso chamador.

**Persistência**

- Consulta: itens necessários da OS do tipo `INSUMO`, `reserva_estoque`, `pedido_compra`,
  `item_estoque`.
- Altera: `reserva_estoque` (status `LIBERADA`), `item_estoque` (`saldoReservado` e `saldoFisico`),
  `movimentacao_estoque` (insert), itens da OS (marcação de devolução e desvínculo).
- Não altera: `ordem_servico.status`, `orcamento` e `orcamento_item`.

**Resultado retornado**

```json
{
  "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
  "reservasLiberadas": [
    {
      "itemId": "b62d4f18-9e33-4a71-8c05-1d7f2ab63e90",
      "codigo": "INS-000031",
      "descricao": "Óleo lubrificante 15W40",
      "tipo": "INSUMO",
      "unidadeMedida": "L",
      "quantidade": 12.0,
      "saldoReservadoApos": 0.0
    }
  ],
  "itensRetornadosAoEstoque": [
    {
      "itemId": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
      "codigo": "INS-000031",
      "descricao": "Óleo lubrificante 15W40",
      "tipo": "INSUMO",
      "unidadeMedida": "L",
      "quantidade": 12.0,
      "saldoFisicoApos": 56.0
    }
  ],
  "itensSemDevolucao": [
    {
      "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "codigo": "INS-000077",
      "descricao": "Fluido de freio DOT 4",
      "tipo": "INSUMO",
      "unidadeMedida": "L",
      "quantidade": 1.5,
      "motivo": "PEDIDO_DE_COMPRA_NAO_RECEBIDO",
      "pedidoId": "f05a1d63-8b47-49e2-a731-6c94d2e08b57"
    }
  ],
  "totalItensProcessados": 3
}
```

O caso de uso chamador decide o que fazer com esse resultado: compor a resposta da recusa,
registrar em log, ou ambos.

**Tratamento de erros**

Não há códigos HTTP próprios. As exceções sobem para o caso de uso de recusa:

| Situação | Comportamento |
|---|---|
| Item ou reserva inconsistente | Exceção de domínio; rollback da recusa inteira. |
| Saldo reservado insuficiente para a liberação | Exceção de domínio; rollback. |
| Lock timeout ou falha de banco | Exceção de infraestrutura; rollback e nova tentativa da recusa pelo usuário. |

**Dependências**

- Caso de uso `RecusarOrcamento`, o chamador.
- `ItemEstoqueRepository`.
- `ReservaEstoqueRepository`.
- `MovimentacaoEstoqueRepository`.
- `PedidoCompraRepository`, para identificar os itens pendentes de compra.

**Testes**

*Unitários*

- Liberação de reserva reduzindo o saldo reservado e sem alterar o saldo físico.
- Retorno de item baixado aumentando o saldo físico.
- Item pendente de compra desvinculado sem movimentação.
- Item já devolvido ignorado.
- Processamento de vários insumos na mesma devolução.
- OS sem itens concluindo com resultado vazio.
- Cálculo de `saldoReservadoApos` e `saldoFisicoApos` com quantidade fracionária.
- Status da OS inalterado pelo serviço.

*Integração*

- Recusa devolvendo os itens e cancelando a OS na mesma transação.
- Reservas com status `LIBERADA` no banco.
- Saldo reservado reduzido e saldo físico inalterado para itens apenas reservados.
- Saldo físico restabelecido para itens já baixados.
- Movimentações de `LIBERACAO_RESERVA` e `ENTRADA_RETORNO` registradas.
- Erro na devolução desfazendo a recusa por inteiro.

*Concorrência*

- Itens carregados com lock de linha, ordenados por `item_id`, para reduzir risco de deadlock.
- Liberação concorrente não gera saldo reservado negativo.

---

### 9.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o status `LIBERADA` na `ReservaEstoque`
- [ ] Implementar a regra de liberação de reserva reduzindo apenas o saldo reservado
- [ ] Implementar a regra de retorno de item baixado aumentando o saldo físico
- [ ] Implementar a movimentação do tipo `LIBERACAO_RESERVA`
- [ ] Implementar a movimentação do tipo `ENTRADA_RETORNO`
- [ ] Implementar a marcação de devolução nos itens da OS
- [ ] Implementar o descarte de itens já marcados como devolvidos
- [ ] Garantir que a devolução não altera o status da OS nem o orçamento
- [ ] Garantir que o saldo reservado nunca fique negativo

**Serviço de devolução**

- [ ] Implementar `DevolverItensAoEstoque` com a assinatura definida
- [ ] Separar os itens em reservados, baixados e pendentes de compra
- [ ] Desvincular os itens pendentes de compra
- [ ] Montar e retornar o `ResultadoDevolucao`
- [ ] Executar dentro da transação do chamador, sem abrir transação própria

**Integração com a recusa**

- [ ] Chamar a devolução a partir de `RecusarOrcamento`, depois do cancelamento da OS
- [ ] Propagar as exceções para desfazer a recusa por inteiro

**Repositório**

- [ ] Criar consulta dos itens vinculados à OS
- [ ] Criar consulta das reservas ativas da OS

**Concorrência**

- [ ] Carregar os itens com lock de linha
- [ ] Ordenar os itens por `item_id` antes do lock para reduzir risco de deadlock
- [ ] Garantir que a liberação concorrente não gera saldo reservado negativo

**Testes unitários**

- [ ] Liberação de reserva reduzindo o saldo reservado
- [ ] Liberação de reserva sem alterar o saldo físico
- [ ] Retorno de item baixado aumentando o saldo físico
- [ ] Item pendente de compra desvinculado sem movimentação
- [ ] Item já devolvido ignorado
- [ ] Processamento de vários insumos na mesma devolução
- [ ] OS sem itens concluindo com resultado vazio
- [ ] Cálculo de `saldoReservadoApos` e `saldoFisicoApos` com quantidade fracionária
- [ ] Status da OS inalterado pelo serviço

**Testes de integração**

- [ ] Recusa devolvendo os itens e cancelando a OS na mesma transação
- [ ] Reservas com status `LIBERADA` no banco
- [ ] Saldo reservado reduzido e saldo físico inalterado para itens apenas reservados
- [ ] Saldo físico restabelecido para itens já baixados
- [ ] Movimentações registradas no histórico
- [ ] Erro na devolução desfazendo a recusa por inteiro

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
