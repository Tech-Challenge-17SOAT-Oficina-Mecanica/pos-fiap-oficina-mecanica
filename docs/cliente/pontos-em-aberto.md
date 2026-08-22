---
documento: Pontos em Aberto — Contexto de Cliente
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Cliente

## O que é este documento

Um ponto em aberto é uma **decisão que ainda não foi tomada** ou uma **inconsistência encontrada**
entre os documentos deste contexto. Enquanto o ponto estiver aberto, quem for implementar não deve
resolver sozinho: a escolha muda contrato de API, modelo de dados ou regra de negócio, e precisa
valer para o time inteiro.

Cada linha traz o ponto, o arquivo afetado e quem decide. Para fechar um ponto:

1. aplique a decisão nos documentos afetados;
2. registre o porquê — como *Decisão de projeto* no documento da tarefa, ou em
   [`02-decisoes-arquiteturais.md`](../02-decisoes-arquiteturais.md) quando a decisão valer para
   todos os contextos;
3. remova a linha desta tabela.

Divergência que atravessa contextos mora em [`02-decisoes-arquiteturais.md`](../02-decisoes-arquiteturais.md);
aqui ficam as que se resolvem dentro de Cliente. O retrato do que já está decidido está em
[`00-resumo.md`](00-resumo.md).

## Inconsistências a corrigir

| # | Ponto | Arquivo relacionado | Responsável |
|---|---|---|---|
| 1 | Três operações fazem coisas parecidas: `POST /veiculos` cadastra veículo sem vínculo, `POST /clientes/{clienteId}/veiculos` cadastra e vincula, e `POST /clientes/{clienteId}/veiculos/{veiculoId}` vincula veículo existente. Definir quais permanecem e qual é o caminho padrão. | [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) e [`../veiculo/cadastrar-veiculo-e-vincular-ao-cliente.md`](../veiculo/cadastrar-veiculo-e-vincular-ao-cliente.md) | — |
| 2 | O vínculo cliente-veículo está documentado em dois contextos ao mesmo tempo. Confirmar se ele pertence a Cliente, como está hoje, ou a Veículo. | [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) | — |
| 3 | Vincular veículo devolve `200`. Confirmar se não deveria ser `201`, já que cria um vínculo novo. | [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) | — |
| 4 | Um veículo pode pertencer a mais de um cliente ao mesmo tempo? O contrato de hoje só impede a duplicidade do mesmo par. Falta a regra de troca de dono, e não existe operação de desvincular. | [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) | — |
| 5 | Nenhuma operação de escrita usa `If-Match` com `version`, enquanto Peças & Insumos e Serviços usam controle otimista. Definir se cadastro de cliente entra nesse padrão. | [`atualizar-cliente.md`](atualizar-cliente.md) e [`vincular-veiculo-ao-cliente.md`](vincular-veiculo-ao-cliente.md) | — |
| 6 | A resposta de consulta de cliente não usa envelope. Definir o padrão para recurso único, já que a listagem paginada tem envelope definido. | [`consultar-cliente.md`](consultar-cliente.md) | — |
| 7 | O cliente não tem contato cadastrado — telefone ou e-mail —, mas orçamento, finalização e entrega preveem notificar o cliente. Sem contato, esses fluxos não fecham. | [`cadastrar-cliente.md`](cadastrar-cliente.md) | — |
| 8 | A anonimização citada no refinamento de Deletar Cliente não tem refinamento próprio nem rota no catálogo. Definir se entra no MVP e, se sim, refiná-la como tarefa. | [`deletar-cliente.md`](deletar-cliente.md) | — |
| 9 | A exclusão lógica depende de índice parcial `UNIQUE (cpf_cnpj) WHERE ativo = true` e da ausência de `ON DELETE CASCADE` nas foreign keys de OS. Confirmar com quem cuidar da migration. | [`deletar-cliente.md`](deletar-cliente.md) | — |
| 10 | Confirmar o dono do contexto e preencher o campo `dono` dos cinco documentos, hoje em "A definir". | Todos os documentos do contexto | — |
