---
documento: Pontos em Aberto — Contexto de Serviços
dono: João Victor Silva de Oliveira
versao: 1.0
atualizado_em: 2026-08-22
status: fechado
---

# Pontos em Aberto — Serviços

## O que é este documento

Um ponto em aberto é uma **decisão que ainda não foi tomada** ou uma **inconsistência encontrada**
entre os documentos deste contexto. Enquanto o ponto estiver aberto, quem for implementar não deve
resolver sozinho: a escolha muda contrato de API, modelo de dados ou regra de negócio, e precisa
valer para o time inteiro.

Para fechar um ponto:

1. aplique a decisão nos documentos afetados;
2. registre o porquê — como *Decisão de projeto* no documento da tarefa, ou em
   [02-decisoes-arquiteturais.md](../02-decisoes-arquiteturais.md) quando valer para todos os contextos;
3. mova a linha para a tabela de decisões abaixo.

O retrato do que já está decidido está em [00-resumo.md](00-resumo.md).

## Inconsistências a corrigir

Nenhuma. Os dezenove pontos levantados nesta rodada foram decididos e aplicados — a tabela abaixo
registra o que ficou valendo. Ponto novo entra aqui, com o motivo de ser um problema e uma
sugestão de correção.

## Decisões desta rodada

| # | Ponto | Decisão | Onde foi aplicada |
|---|---|---|---|
| 1 | Numeração duplicada: Consultar Serviços e Desativar Serviço eram ambas `## 1 ·`. | **Renumerado na ordem do fluxo:** 1 cadastrar, 2 consultar, 3 atualizar, 4 desativar e reativar. | Os quatro documentos e [00-resumo.md](00-resumo.md) |
| 2 | Situação do serviço divergente: três documentos usavam `status` com `ATIVO`/`INATIVO`, um usava o booleano `ativo`. | **Booleano `ativo`** em tudo, com `dataDesativacao` e `usuarioDesativacao`, como nos demais contextos. | [cadastrar-servico.md](cadastrar-servico.md), [consultar-servicos.md](consultar-servicos.md), [atualizar-servico.md](atualizar-servico.md) e D-19 |
| 3 | Duplicidade de IDs: `RF-SRV-01` a `RF-SRV-06` apareciam em mais de um documento. | **Renumerado o contexto inteiro** na ordem do fluxo: `RF-SRV-01` a `RF-SRV-25` e `RNF-SRV-01` a `RNF-SRV-21`, sem repetição. | Os quatro documentos |
| 4 | Path param divergente: consultar e atualizar usavam `{id}`; desativar usava `{servicoId}`. | **`{servicoId}` em tudo**, alinhado a `{clienteId}`, `{veiculoId}` e `{pecaId}`. | [consultar-servicos.md](consultar-servicos.md), [atualizar-servico.md](atualizar-servico.md) e DT-25 |
| 5 | Envelope de paginação divergente: a listagem usava `content`, `page`, `size`, `totalElements` e `totalPages`. | **Envelope padrão do projeto:** `data`, `pagina`, `tamanho`, `totalElementos` e `totalPaginas`. Recurso único vai direto, sem envelope. | [consultar-servicos.md](consultar-servicos.md) e D-21 |
| 6 | Verbo da desativação: `PATCH /servicos/{servicoId}/desativar`, enquanto os demais contextos usam `DELETE`. | **`DELETE /servicos/{servicoId}`**, com a exclusão lógica documentada de forma explícita. A rota antiga foi para a lista de rotas aposentadas. | [desativar-servico.md](desativar-servico.md), [../03-endpoints.md](../03-endpoints.md) e D-20 |
| 7 | Não existia reativação de serviço, embora Cliente e Veículo tenham. | **Criada `POST /servicos/{servicoId}/reativacao`**, com validação de unicidade de nome contra os serviços ativos. | [desativar-servico.md](desativar-servico.md) e DT-17 |
| 8 | Remoção física quando não há vínculo (antigo `RF-SRV-06`) não tinha endpoint nem regra. | **Fora do MVP.** O requisito saiu do documento; a exclusão é sempre lógica. | [desativar-servico.md](desativar-servico.md) e DT-18 |
| 9 | Critério de unicidade do serviço não definido. | **Nome normalizado** — sem acento, sem espaço duplo, minúsculo — único **entre serviços ativos**, por índice parcial `UNIQUE (nome_normalizado) WHERE ativo = true`. | [cadastrar-servico.md](cadastrar-servico.md), [atualizar-servico.md](atualizar-servico.md), [desativar-servico.md](desativar-servico.md) e DT-19 |
| 10 | Regra de geração do código `SER-000001` não definida. | **Sequência global, sem reset, com seis dígitos** — mesma regra proposta para o código de peças e insumos. | [cadastrar-servico.md](cadastrar-servico.md) e DT-20 |
| 11 | `tempoEstimadoMinutos` obrigatório ou opcional? | **Obrigatório**, com mínimo de 1 minuto. É o que alimenta a estimativa de entrega do orçamento. | [cadastrar-servico.md](cadastrar-servico.md), [atualizar-servico.md](atualizar-servico.md) e DT-21 |
| 12 | Campos imutáveis na atualização não estavam definidos. | **`id`, `codigo` e `dataCriacao` são imutáveis**, e enviados no corpo retornam `400`. `ativo` também não é alterado pelo `PATCH`: a situação muda pelas rotas de desativação e reativação. | [atualizar-servico.md](atualizar-servico.md) e DT-22 |
| 13 | `PATCH` aceita atualização parcial real ou payload completo? | **Atualização parcial de verdade:** campo ausente não é alterado. | [atualizar-servico.md](atualizar-servico.md) e DT-22 |
| 14 | Preservação do valor histórico do serviço em OS e orçamentos. | **Cópia no momento do uso:** `descricao` e `valorUnitario` são copiados para o item do orçamento, que é o que os contextos de OS e Orçamento já descrevem. Sem versionamento do serviço. | [atualizar-servico.md](atualizar-servico.md) e [00-resumo.md](00-resumo.md) |
| 15 | Controle otimista com `If-Match` na atualização: usar ou não? | **Usar.** `PATCH /servicos/{servicoId}` exige `If-Match`, devolve `412` quando a `version` diverge e `428` quando o header não vem. A consulta por identificador expõe `version`. | [atualizar-servico.md](atualizar-servico.md), [consultar-servicos.md](consultar-servicos.md) e D-24 |
| 16 | Serviços inativos aparecem na consulta padrão? | **Ocultos por padrão**, exibidos com `incluirInativos=true` — mesmo parâmetro da consulta de peças. | [consultar-servicos.md](consultar-servicos.md) e DT-23 |
| 17 | Limite máximo de `tamanho` na paginação não definido. | **Teto de 50**, devolvendo `400` acima disso. Nasceu menor que o teto de 100 dos demais contextos e depois virou o padrão de toda a API (D-26). | [consultar-servicos.md](consultar-servicos.md), DT-23 e D-26 |
| 18 | O cliente pode consultar o catálogo de serviços? | **Restrito à oficina**, perfil `MECANICO`. O cliente vê os serviços pelo orçamento, não pelo catálogo. | [consultar-servicos.md](consultar-servicos.md) e DT-24 |
| 19 | Nomes dos escopos foram derivados do padrão, não confirmados. | **Lista oficial de escopos consolidada no guia**, seção 8, junto com os perfis do sistema. Escopo novo só existe depois de entrar nessa lista. | Seção 8 do [../01-guia-de-documentacao.md](../01-guia-de-documentacao.md) |
