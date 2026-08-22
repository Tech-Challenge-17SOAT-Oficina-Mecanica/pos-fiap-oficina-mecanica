---
documento: Pontos em Aberto — Contexto de Peças & Insumos
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Peças & Insumos

## O que é este documento

Um ponto em aberto é uma **decisão que ainda não foi tomada** ou uma **inconsistência encontrada**
entre os documentos deste contexto. Enquanto o ponto estiver aberto, quem for implementar não deve
resolver sozinho: a escolha muda contrato de API, modelo de dados ou regra de negócio.

Para fechar um ponto:

1. aplique a decisão nos documentos afetados;
2. registre o porquê — como *Decisão de projeto* no documento da tarefa, ou em
   [`02-decisoes-arquiteturais.md`](../02-decisoes-arquiteturais.md) quando valer para todos os contextos;
3. remova a linha desta tabela.

Este contexto recebeu contribuições de várias pessoas em paralelo, então há mais divergência de
contrato aqui do que nos demais. O retrato do que já está decidido está em [`00-resumo.md`](00-resumo.md).

## Inconsistências a corrigir

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Numeração duplicada em cinco pares: `10` em Cadastrar Peça e Reservar Peça, `11` em Deletar Peça, Consultar Peças e Reservar Insumo, `12` em Cadastrar Insumo e Processar Peças, `13` em Deletar Insumo e Processar Insumos, `14` em Consultar Insumos e Retornar ao Estoque. Renumerar o contexto inteiro. | Todos os documentos do contexto | — |
| 2 | Duplicidade de IDs de requisito: `RF-EST-100` e a faixa `RF-EST-120` a `RF-EST-140` aparecem em mais de um documento. Renumerar junto com a numeração das tarefas. | Todos os documentos do contexto | — |
| 3 | **Três caminhos para reservar**: reserva direta de peça (`POST /estoque/reservas`), reserva direta de insumo (`POST /estoque/reservas-insumos`) e o processamento que reserva o disponível e compra o faltante (`POST /estoque/solicitacoes-compra-reserva`). Some-se a isso que o pedido de compra também cria reserva. Definir qual é o fluxo oficial. | [`reservar-peca-para-os.md`](reservar-peca-para-os.md), [`reservar-insumo-para-os.md`](reservar-insumo-para-os.md), [`processar-pecas-para-reserva-e-compra.md`](processar-pecas-para-reserva-e-compra.md) e [`solicitar-compra-de-pecas.md`](solicitar-compra-de-pecas.md) | — |
| 4 | Peça e insumo têm rotas separadas para reservar e para processar, mas compartilham a rota de compra e a mesma entidade de item. Definir se a separação por tipo vale para tudo ou para nada. | [`reservar-insumo-para-os.md`](reservar-insumo-para-os.md) e [`solicitar-compra-de-insumos.md`](solicitar-compra-de-insumos.md) | — |
| 5 | Insumo é reservado nas tarefas novas, mas a documentação anterior afirmava que insumo não tem reserva e sofre baixa direta na execução. Fechar a regra. | [`reservar-insumo-para-os.md`](reservar-insumo-para-os.md) | — |
| 6 | Consultar Peças não tem seção de autenticação e autorização: é o único documento do contexto sem escopo definido. | [`consultar-estoque.md`](consultar-estoque.md) | — |
| 7 | O arquivo `consultar-estoque.md` documenta a tarefa "Consultar Peças" e responde em `/estoque/pecas`. Renomear o arquivo para refletir o conteúdo. | [`consultar-estoque.md`](consultar-estoque.md) | — |
| 8 | A consulta unificada de itens (`GET /estoque/itens`) deixou de existir: agora são duas rotas por tipo. Confirmar se nenhuma tela precisa da visão consolidada. | [`consultar-estoque.md`](consultar-estoque.md) e [`consultar-insumos.md`](consultar-insumos.md) | — |
| 9 | Formato do código funcional: convivem `PC-0142`, `IN-0031`, `INS-0012` e `PEC-000001`. Definir um formato único, quem gera e se o usuário pode informar. | [`cadastrar-peca.md`](cadastrar-peca.md), [`cadastrar-insumo.md`](cadastrar-insumo.md) e [`consultar-insumos.md`](consultar-insumos.md) | — |
| 10 | Campo `nome`: os cadastros introduzem `nome` além de `descricao`, mas as consultas e atualizações só expõem `descricao`. Definir se o item tem os dois campos. | [`cadastrar-peca.md`](cadastrar-peca.md) e [`cadastrar-insumo.md`](cadastrar-insumo.md) | — |
| 11 | Estoque inicial: o cadastro de peça proíbe e manda usar a entrada de estoque; o de insumo aceita `saldoFisicoInicial`. Definir a regra para os dois tipos. | [`cadastrar-peca.md`](cadastrar-peca.md) e [`cadastrar-insumo.md`](cadastrar-insumo.md) | — |
| 12 | Regra de duplicidade do item: os dois cadastros dizem "conforme a regra definida pela oficina" sem definir qual é. | [`cadastrar-peca.md`](cadastrar-peca.md) e [`cadastrar-insumo.md`](cadastrar-insumo.md) | — |
| 13 | Dois caminhos para inativar o mesmo item: `PUT` com `ativo: false` e `DELETE`. Definir qual é o oficial. | [`atualizar-peca.md`](atualizar-peca.md) e [`deletar-peca.md`](deletar-peca.md) | — |
| 14 | Desativação com saldo em estoque ou com orçamento pendente: a regra está declarada como "a definir" nos dois documentos de exclusão. | [`deletar-peca.md`](deletar-peca.md) e [`deletar-insumo.md`](deletar-insumo.md) | — |
| 15 | A atualização de peça bloqueia inativação com saldo reservado; a de insumo não trata o caso. Unificar. | [`atualizar-peca.md`](atualizar-peca.md) e [`atualizar-insumo.md`](atualizar-insumo.md) | — |
| 16 | `precoVenda` na peça e `custoUnitario` no insumo, na mesma entidade. Definir como as consultas representam os dois tipos. | [`consultar-estoque.md`](consultar-estoque.md) e [`consultar-insumos.md`](consultar-insumos.md) | — |
| 17 | Custo do insumo na entrada: a entrada grava `custoUnitario` por recebimento, e a atualização trata o custo como dado cadastral. Definir se a entrada atualiza o custo, e por qual critério. | [`registrar-entrada-de-estoque.md`](registrar-entrada-de-estoque.md) e [`atualizar-insumo.md`](atualizar-insumo.md) | — |
| 18 | `Idempotency-Key` é obrigatório nas reservas e apenas recomendado na entrada. Padronizar o mecanismo e o comportamento da repetição. | [`registrar-entrada-de-estoque.md`](registrar-entrada-de-estoque.md) e [`reservar-peca-para-os.md`](reservar-peca-para-os.md) | — |
| 19 | A quantidade comprada deve ser exatamente igual à necessidade apurada nas OS, o que impede comprar lote maior para estoque. Confirmar se é essa a regra. | [`solicitar-compra-de-pecas.md`](solicitar-compra-de-pecas.md) e [`solicitar-compra-de-insumos.md`](solicitar-compra-de-insumos.md) | — |
| 20 | Fornecedor é obrigatório na compra de peças e opcional na de insumos, na mesma rota. Confirmar a assimetria. | [`solicitar-compra-de-pecas.md`](solicitar-compra-de-pecas.md) e [`solicitar-compra-de-insumos.md`](solicitar-compra-de-insumos.md) | — |
| 21 | Não existe cadastro de fornecedor, embora ele seja pré-condição das compras e origem do lead time. | [`solicitar-compra-de-pecas.md`](solicitar-compra-de-pecas.md) | — |
| 22 | A entrada de estoque muda o status das OS vinculadas ao pedido. Confirmar com Ordem de Serviço quem é o dono dessa transição. | [`registrar-entrada-de-estoque.md`](registrar-entrada-de-estoque.md) | — |
| 23 | A devolução na recusa assume que os itens da OS têm marcação de devolução e vínculo com pedido de compra. Definir onde esses campos vivem. | [`retornar-peca-e-insumo-ao-estoque.md`](retornar-peca-e-insumo-ao-estoque.md) | — |
| 24 | Existe devolução na recusa do orçamento, mas não no cancelamento da OS por outro motivo. Confirmar se o cancelamento também devolve. | [`retornar-peca-e-insumo-ao-estoque.md`](retornar-peca-e-insumo-ao-estoque.md) | — |
| 25 | Tarefas removidas e ainda não reescritas: consultar peças faltantes e registrar consumo e saída. Sem a segunda, não existe baixa de estoque na execução do serviço. | — | — |
| 26 | Confirmar o dono do contexto e preencher o campo `dono` dos documentos, hoje em "A definir" ou "Desconhecido". | Todos os documentos do contexto | — |
