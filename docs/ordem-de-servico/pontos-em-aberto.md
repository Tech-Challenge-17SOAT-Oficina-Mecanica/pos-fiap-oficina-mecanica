---
documento: Pontos em Aberto — Contexto de Ordem de Serviço
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Pontos em Aberto — Ordem de Serviço

Este documento centraliza as decisões pendentes das tarefas do contexto de Ordem de Serviço.

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Confirmar se o início do diagnóstico deve registrar também o identificador do mecânico responsável. | [`iniciar-diagnostico.md`](iniciar-diagnostico.md) | — |
| 2 | Confirmar se `os:escrever` será o escopo compartilhado das alterações de Ordem de Serviço ou se haverá um escopo específico para diagnóstico. | [`iniciar-diagnostico.md`](iniciar-diagnostico.md) e [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 3 | Definir o formato definitivo do identificador da Ordem de Serviço para concluir sua validação técnica. | [`iniciar-diagnostico.md`](iniciar-diagnostico.md) e [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 4 | Confirmar se o início do diagnóstico deve publicar um evento de domínio, como `DiagnosticoIniciado`. | [`iniciar-diagnostico.md`](iniciar-diagnostico.md) | — |
| 5 | Confirmar se um serviço duplicado deve ser rejeitado ou se o registro existente deve ser atualizado. O contrato atual adota rejeição com `409`. | [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 6 | Definir se serviços da Ordem de Serviço possuem quantidade ou se cada serviço pode aparecer apenas uma vez. | [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 7 | Definir se nome e preço do serviço devem ser copiados para a Ordem de Serviço no momento da associação, preservando o histórico contra alterações futuras no catálogo. | [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 8 | Definir os limites de tamanho e o conteúdo permitido no campo `observacao`. | [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 9 | Confirmar se a criação de Ordem de Serviço deve ser renumerada como requisito 1 do contexto, já que acontece antes de Iniciar Diagnóstico, ou se a numeração atual deve ser mantida para evitar retrabalho nos arquivos existentes. | [`criar-ordem-de-servico.md`](criar-ordem-de-servico.md), [`iniciar-diagnostico.md`](iniciar-diagnostico.md) e [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 10 | Confirmar se serviços solicitados e peças/insumos necessários devem ser associados já na criação da Ordem de Serviço ou se devem ficar apenas em tarefas posteriores do diagnóstico. | [`criar-ordem-de-servico.md`](criar-ordem-de-servico.md) e [`registrar-servicos-necessarios.md`](registrar-servicos-necessarios.md) | — |
| 11 | Definir os limites de tamanho e o conteúdo permitido no campo `problemaRelatado`. | [`criar-ordem-de-servico.md`](criar-ordem-de-servico.md) | — |
| 12 | Confirmar se a criação de Ordem de Serviço deve publicar evento de domínio, como `OrdemDeServicoCriada`. | [`criar-ordem-de-servico.md`](criar-ordem-de-servico.md) | — |
| 13 | Confirmar se a criação de Ordem de Serviço deve exigir idempotência por `Idempotency-Key`. | [`criar-ordem-de-servico.md`](criar-ordem-de-servico.md) | — |
| 14 | A fila de atendimento foi documentada em `GET /fila-atendimento`. Definir se o recurso fica na raiz ou sob `/ordens-servico/fila-atendimento`, já que a fila é uma visão da OS. | [`consultar-fila-de-atendimento.md`](consultar-fila-de-atendimento.md) | — |
| 15 | O refinamento da fila e do tempo médio usava o envelope `content`, `page`, `size`, `totalElements`, `totalPages`. Foi padronizado para `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`, como no resto do projeto. Confirmar. | [`consultar-fila-de-atendimento.md`](consultar-fila-de-atendimento.md) e [`monitorar-tempo-medio-de-execucao.md`](monitorar-tempo-medio-de-execucao.md) | — |
| 16 | A fila exige que peças e insumos estejam disponíveis ou reservados, o que faz o contexto de OS consultar estoque a cada listagem. Definir se essa checagem é síncrona ou se a OS guarda um indicador atualizado por evento do estoque. | [`consultar-fila-de-atendimento.md`](consultar-fila-de-atendimento.md) | — |
| 17 | O status `AGUARDANDO_EXECUCAO` aparece na fila e no início da execução, mas não está na lista de status do enunciado (`Recebida`, `Em diagnóstico`, `Aguardando aprovação`, `Em execução`, `Finalizada`, `Entregue`). Confirmar a máquina de estados completa da OS, incluindo `CANCELADA`. | [`consultar-fila-de-atendimento.md`](consultar-fila-de-atendimento.md) e [`iniciar-execucao.md`](iniciar-execucao.md) | — |
| 18 | O indicador de tempo médio foi restrito ao perfil `GESTOR`. Confirmar se o mecânico também pode consultar o próprio tempo médio. | [`monitorar-tempo-medio-de-execucao.md`](monitorar-tempo-medio-de-execucao.md) | — |
| 19 | A listagem de tempos de execução usava `page` começando em 1 no refinamento original, enquanto o resto do projeto usa página começando em 0. Foi padronizado em 0. Confirmar. | [`monitorar-tempo-medio-de-execucao.md`](monitorar-tempo-medio-de-execucao.md) | — |
