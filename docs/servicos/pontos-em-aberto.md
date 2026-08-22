---
documento: Pontos em Aberto — Contexto de Serviços
dono: João Victor Silva de Oliveira
versao: 0.3
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Serviços

## O que é este documento

Um ponto em aberto é uma **decisão que ainda não foi tomada** ou uma **inconsistência encontrada**
entre os documentos deste contexto. Enquanto o ponto estiver aberto, quem for implementar não deve
resolver sozinho: a escolha muda contrato de API, modelo de dados ou regra de negócio.

Para fechar um ponto:

1. aplique a decisão nos documentos afetados;
2. registre o porquê — como *Decisão de projeto* no documento da tarefa, ou em
   [`02-decisoes-arquiteturais.md`](../02-decisoes-arquiteturais.md) quando valer para todos os contextos;
3. remova a linha desta tabela.

O retrato do que já está decidido está em [`00-resumo.md`](00-resumo.md).

## Inconsistências a corrigir

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Numeração duplicada: Consultar Serviços e Desativar Serviço são ambas `## 1 ·`. Renumerar o contexto na ordem do fluxo. | [`consultar-servicos.md`](consultar-servicos.md) e [`desativar-servico.md`](desativar-servico.md) | — |
| 2 | Situação do serviço divergente: cadastrar, consultar e atualizar usam `status` com `ATIVO`/`INATIVO`; desativar usa o booleano `ativo`. Escolher um dos dois e alinhar os quatro documentos. | Todos os documentos do contexto | — |
| 3 | Duplicidade de IDs: `RF-SRV-01` a `RF-SRV-06` aparecem em mais de um documento do contexto. Renumerar. | Todos os documentos do contexto | — |
| 4 | Path param divergente: consultar e atualizar usam `{id}`; desativar usa `{servicoId}`. Padronizar. | [`consultar-servicos.md`](consultar-servicos.md), [`atualizar-servico.md`](atualizar-servico.md) e [`desativar-servico.md`](desativar-servico.md) | — |
| 5 | Envelope de paginação divergente: a listagem usa `content`, `page`, `size`, `totalElements` e `totalPages`, enquanto o resto do projeto usa `data`, `pagina`, `tamanho`, `totalElementos` e `totalPaginas`. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 6 | Confirmar se clientes também poderão consultar o catálogo de serviços ou se a consulta será restrita ao Gestor. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 7 | Confirmar se o escopo definitivo para consulta de serviços será `servicos:ler` ou outro nome padronizado pelo time. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 8 | Definir o limite máximo de `size` para paginação da listagem de serviços. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 9 | Confirmar se a listagem paginada de Serviços deve usar `content`, `page`, `size`, `totalElements` e `totalPages`, ou o envelope em português já usado em outros contextos. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 10 | Definir se serviços inativos aparecem na consulta padrão ou apenas quando filtrados por `status=INATIVO`. | [`consultar-servicos.md`](consultar-servicos.md) | — |
| 11 | Definir o critério de unicidade para impedir serviço duplicado: nome exato, nome normalizado, código ou combinação de campos. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 12 | Definir se `tempoEstimadoMinutos` será obrigatório no MVP ou opcional. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 13 | Definir a regra de geração do código funcional `SER-000001`, incluindo sequência, reset anual ou sequência global. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 14 | Confirmar se o escopo definitivo para cadastro de serviços será `servicos:escrever` ou outro nome padronizado pelo time. | [`cadastrar-servico.md`](cadastrar-servico.md) | — |
| 15 | Confirmar se atualização de serviço deve usar controle otimista com header `If-Match`, como outras operações de escrita do projeto. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 16 | Definir se `PATCH /servicos/{id}` aceitará atualização parcial real ou payload completo dos campos editáveis. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 17 | Confirmar quais campos são imutáveis na atualização, especialmente `codigo`, `status` e identificadores técnicos. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 18 | Definir como os valores históricos de serviços em OS e orçamentos serão preservados: cópia do valor no momento do uso ou histórico versionado do serviço. | [`atualizar-servico.md`](atualizar-servico.md) | — |
| 19 | Verbo da desativação: aqui é `PATCH /servicos/{servicoId}/desativar`, enquanto Cliente, Veículo e Peças & Insumos usam `DELETE` para a mesma transição lógica. Padronizar entre os contextos. | [`desativar-servico.md`](desativar-servico.md) | — |
| 20 | Situação do serviço: o refinamento da desativação usava `status` com `ATIVO` e `INATIVO`. Foi padronizado como o booleano `ativo`, mais `dataDesativacao` e `usuarioDesativacao`, como nos demais contextos. Confirmar e alinhar com os documentos de cadastro, consulta e atualização. | [`desativar-servico.md`](desativar-servico.md) | — |
| 21 | Remoção física quando não há vínculo (RF-SRV-06) não tem endpoint nem regra definida. Decidir se o MVP suporta os dois caminhos ou apenas a desativação lógica. | [`desativar-servico.md`](desativar-servico.md) | — |
| 22 | Reativação de serviço inativo não está prevista, embora Cliente e Veículo tenham rota de reativação. Definir se entra no MVP. | [`desativar-servico.md`](desativar-servico.md) | — |
