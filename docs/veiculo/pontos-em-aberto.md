---
documento: Pontos em Aberto — Contexto de Veículo
dono: A definir
versao: 1.0
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Veículo

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

Nenhuma. O último ponto aberto do contexto dependia da D-01 e foi fechado com ela, em 22/08/2026.

## Decisões desta rodada

| # | Ponto | Decisão | Onde foi aplicada |
|---|---|---|---|
| 1 | Duplicidade de IDs: `RF-VEI-13` a `RF-VEI-19` e `RNF-VEI-10` a `RNF-VEI-14` apareciam em dois documentos. | **Renumerado.** Atualizar Veículo passou para `RF-VEI-27` a `RF-VEI-33` e `RNF-VEI-22` a `RNF-VEI-26`; Deletar Veículo ficou como estava. O contexto inteiro foi conferido e não há mais ID repetido. | [atualizar-veiculo.md](atualizar-veiculo.md) |
| 2 | Três operações concorrentes de cadastro e vínculo de veículo. | **`POST /veiculos` aposentada.** Já aplicada junto com a revisão do contexto de Cliente; conferida aqui. | [cadastrar-veiculo.md](cadastrar-veiculo.md), [00-resumo.md](00-resumo.md), seção *Rotas aposentadas* de [../03-endpoints.md](../03-endpoints.md) e DT-05 |
| 3 | A operação combinada usa rota sob `/clientes`, mas está documentada no contexto de Veículo. | **Fica onde está.** Com o cadastro avulso aposentado, esta passou a ser a única forma de cadastrar veículo, e o refinamento pertence a quem conhece as regras do veículo. | [cadastrar-veiculo-e-vincular-ao-cliente.md](cadastrar-veiculo-e-vincular-ao-cliente.md) |
| 4 | A operação combinada exige `veiculos:escrever` mais `clientes:ler`. | **Exige apenas `veiculos:escrever`.** Não há validação cruzada contra o contexto de Cliente: a permissão vem nos escopos do próprio JWT. | [cadastrar-veiculo-e-vincular-ao-cliente.md](cadastrar-veiculo-e-vincular-ao-cliente.md) e DT-15 |
| 5 | A persona da operação combinada é o Gestor; as demais tarefas têm o Mecânico. | **O perfil `GESTOR` deixou de existir.** Ficam `MECANICO`, `CLIENTE` e `SERVICO`. A troca foi aplicada em todos os contextos, não só neste. | Todas as tarefas do projeto, [../00-visao-geral.md](../00-visao-geral.md) e DT-11 |
| 6 | Formato da placa: Mercosul e antigo convivem na frota. | **Aceita os dois**, `ABC1D23` e `ABC1234`, com normalização para maiúsculas, sem hífen e sem espaço antes de validar e de gravar. | [cadastrar-veiculo.md](cadastrar-veiculo.md), [atualizar-veiculo.md](atualizar-veiculo.md), [consultar-veiculo.md](consultar-veiculo.md) e DT-12 |
| 7 | Faixa de validação do campo `ano` não definida. | **De `1900` até o ano corrente mais um**, o que cobre o modelo do ano seguinte vendido no fim do ano. | [cadastrar-veiculo.md](cadastrar-veiculo.md), [atualizar-veiculo.md](atualizar-veiculo.md) e DT-13 |
| 8 | Alteração de placa em veículo com Ordens de Serviço existentes. | **Permitida.** A OS grava a placa vigente no momento da criação, então a correção do cadastro não reescreve o histórico. | [atualizar-veiculo.md](atualizar-veiculo.md), [../ordem-de-servico/criar-ordem-de-servico.md](../ordem-de-servico/criar-ordem-de-servico.md) e DT-14 |
| 9 | Nenhuma operação de escrita usava `If-Match` com `version`. | **`If-Match` obrigatório** no `PUT /veiculos/{veiculoId}`: `412` quando a `version` diverge, `428` quando o header não vem. A consulta expõe `version`. | [atualizar-veiculo.md](atualizar-veiculo.md), [consultar-veiculo.md](consultar-veiculo.md) e D-24 |
| 10 | A exclusão lógica depende do índice parcial `UNIQUE (placa) WHERE ativo = true` e da ausência de `ON DELETE CASCADE`. | **Viraram itens obrigatórios do review da primeira migration**, junto com as exigências equivalentes do contexto de Cliente. | [deletar-veiculo.md](deletar-veiculo.md) |
| 11 | Não existe histórico de proprietários: ao trocar o dono, o vínculo anterior se perde. | **Fica como está no MVP.** Sem histórico de propriedade. | [00-resumo.md](00-resumo.md) e DT-16 |
| 12 | O campo `dono` dos documentos está em "A definir". | **Sem ação por ora.** Fica para a definição de donos por contexto. | — |
| 13 | A reativação com cliente inativo devolvia `422`, enquanto conflitos de estado em outros contextos usavam `409`. | **`409`.** A D-01 tirou o `422` da API: entrada inválida é `400`, conflito com o estado atual é `409`. Cliente proprietário inativo é conflito de estado. | [deletar-veiculo.md](deletar-veiculo.md), [cadastrar-veiculo-e-vincular-ao-cliente.md](cadastrar-veiculo-e-vincular-ao-cliente.md), [00-resumo.md](00-resumo.md) e D-01 |
