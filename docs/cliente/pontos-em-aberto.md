---
documento: Pontos em Aberto — Contexto de Cliente
dono: A definir
versao: 1.0
atualizado_em: 2026-08-22
status: fechado
---

# Pontos em Aberto — Cliente

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

Nenhuma. Os dez pontos levantados nesta rodada foram decididos e aplicados — a tabela abaixo
registra o que ficou valendo. Ponto novo entra aqui, com o motivo de ser um problema e uma
sugestão de correção.

## Decisões desta rodada

| # | Ponto | Decisão | Onde foi aplicada |
|---|---|---|---|
| 1 | Três operações faziam a mesma coisa: `POST /veiculos`, `POST /clientes/{clienteId}/veiculos` e `POST /clientes/{clienteId}/veiculos/{veiculoId}`. | **`POST /veiculos` foi aposentada.** O cadastro de veículo acontece sempre dentro do cliente. Ficam as duas rotas sob `/clientes`: uma cadastra e vincula, a outra vincula veículo já cadastrado. | [../veiculo/cadastrar-veiculo.md](../veiculo/cadastrar-veiculo.md), [../veiculo/00-resumo.md](../veiculo/00-resumo.md), seção *Rotas aposentadas* de [../03-endpoints.md](../03-endpoints.md) e DT-05 |
| 2 | O vínculo cliente-veículo está documentado em dois contextos ao mesmo tempo. | **Sem ação.** A duplicação de descrição entre Cliente e Veículo foi considerada aceitável; o time não vê risco de as duas implementações divergirem. | — |
| 3 | Vincular veículo devolvia `200`, não `201`. | **Devolve `201`** quando o vínculo é criado, e `409` quando já existe. | [vincular-veiculo-ao-cliente.md](vincular-veiculo-ao-cliente.md) e DT-06 |
| 4 | Um veículo pode pertencer a mais de um cliente, e não existe operação de desvincular. | **Desvínculo fica para depois do MVP**, junto com a regra de um proprietário ativo por vez. `DELETE /clientes/{clienteId}/veiculos/{veiculoId}` não entra agora. | [vincular-veiculo-ao-cliente.md](vincular-veiculo-ao-cliente.md), [00-resumo.md](00-resumo.md) e DT-07 |
| 5 | Nenhuma operação de escrita usava `If-Match` com `version`, enquanto outros contextos já usavam controle otimista. | **`If-Match` obrigatório** no `PUT /clientes/{clienteId}`: `412` quando a `version` diverge, `428` quando o header não vem. A consulta expõe `version`. | [atualizar-cliente.md](atualizar-cliente.md), [consultar-cliente.md](consultar-cliente.md) e D-24 |
| 6 | A consulta de cliente não usava envelope, enquanto a listagem paginada tem envelope definido. | **Recurso único devolve o objeto direto**, sem envelope; envelope paginado apenas em listagem. A regra virou padrão do projeto. | [consultar-cliente.md](consultar-cliente.md), seção 8 do [../01-guia-de-documentacao.md](../01-guia-de-documentacao.md) e D-21 |
| 7 | O cliente não tinha contato cadastrado — telefone ou e-mail. | **`telefone` e `email` entram no cadastro**, com pelo menos um obrigatório, validados no cadastro e na atualização e devolvidos na consulta. | [cadastrar-cliente.md](cadastrar-cliente.md), [atualizar-cliente.md](atualizar-cliente.md), [consultar-cliente.md](consultar-cliente.md), [00-resumo.md](00-resumo.md) e DT-08 |
| 8 | A anonimização citada no refinamento de Deletar Cliente não tinha refinamento próprio nem rota. | **Fora do MVP.** A menção sai do documento e fica explícito que o MVP não atende pedido de apagamento de dados pessoais: o `DELETE` é sempre exclusão lógica. | [00-resumo.md](00-resumo.md) e DT-09 |
| 9 | A exclusão lógica depende do índice parcial `UNIQUE (cpf_cnpj) WHERE ativo = true` e da ausência de `ON DELETE CASCADE`. | **Viraram itens obrigatórios do review da primeira migration**, além de constarem no checklist de repositório. | [deletar-cliente.md](deletar-cliente.md) |
| 10 | O campo `dono` dos cinco documentos está em "A definir". | **Sem ação por ora.** Fica para a definição de donos por contexto, fora desta rodada. | — |
