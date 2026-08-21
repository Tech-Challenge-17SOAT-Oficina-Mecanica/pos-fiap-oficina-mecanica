---
documento: Pontos em Aberto — Contexto de Serviços
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Pontos em Aberto — Serviços

Este documento centraliza as decisões pendentes das tarefas do contexto de Serviços.

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Confirmar o dono do contexto de Serviços. | — | — |
| 2 | Faltam refinar três tarefas do contexto: cadastrar serviço, consultar serviços e atualizar serviço. A numeração das tarefas provavelmente muda quando elas chegarem. | — | — |
| 3 | Verbo da desativação: aqui é `PATCH /servicos/{servicoId}/desativar`, enquanto Cliente, Veículo e Peças & Insumos usam `DELETE` para a mesma transição lógica. Padronizar entre os contextos. | [`desativar-servico.md`](desativar-servico.md) | — |
| 4 | Situação do serviço: o refinamento usava `status` com `ATIVO` e `INATIVO`. Foi padronizado como o booleano `ativo`, mais `dataDesativacao` e `usuarioDesativacao`, como nos demais contextos. Confirmar. | [`desativar-servico.md`](desativar-servico.md) | — |
| 5 | Formato do código do serviço (`SER-000001`): definir quem gera, se é único e se pode ser alterado — mesma discussão do `codigo` de peças e insumos. | [`desativar-servico.md`](desativar-servico.md) | — |
| 6 | Remoção física quando não há vínculo (RF-SRV-06) não tem endpoint nem regra definida. Decidir se o MVP suporta os dois caminhos ou apenas a desativação lógica. | [`desativar-servico.md`](desativar-servico.md) | — |
| 7 | Reativação de serviço inativo não está prevista, embora Cliente e Veículo tenham rota de reativação. Definir se entra no MVP. | [`desativar-servico.md`](desativar-servico.md) | — |
| 8 | Nomes dos escopos (`servicos:ler`, `servicos:escrever`) foram derivados do padrão `recurso:acao` usado nos demais contextos. Confirmar a lista oficial. | [`desativar-servico.md`](desativar-servico.md) | — |
