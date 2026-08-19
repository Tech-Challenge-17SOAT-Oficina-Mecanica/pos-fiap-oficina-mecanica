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
