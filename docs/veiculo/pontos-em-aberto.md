---
documento: Pontos em Aberto — Contexto de Veículo
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Veículo

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
| 1 | Duplicidade de IDs: `RF-VEI-13` a `RF-VEI-19` aparecem tanto em Deletar Veículo quanto em Atualizar Veículo. Renumerar um dos dois. | [`deletar-veiculo.md`](deletar-veiculo.md) e [`atualizar-veiculo.md`](atualizar-veiculo.md) | — |
| 2 | Três operações concorrentes para o mesmo objetivo: `POST /veiculos` cadastra sem vínculo, `POST /clientes/{clienteId}/veiculos` cadastra e vincula, e `POST /clientes/{clienteId}/veiculos/{veiculoId}` vincula veículo existente. Definir quais permanecem. | [`cadastrar-veiculo.md`](cadastrar-veiculo.md) e [`cadastrar-veiculo-e-vincular-ao-cliente.md`](cadastrar-veiculo-e-vincular-ao-cliente.md) | — |
| 3 | A operação combinada usa rota sob `/clientes`, mas está documentada no contexto de Veículo. Confirmar onde ela mora. | [`cadastrar-veiculo-e-vincular-ao-cliente.md`](cadastrar-veiculo-e-vincular-ao-cliente.md) | — |
| 4 | A operação combinada exige `veiculos:escrever` mais `clientes:ler`, e cria um vínculo no cadastro do cliente. Confirmar se não deveria exigir também `clientes:escrever`. | [`cadastrar-veiculo-e-vincular-ao-cliente.md`](cadastrar-veiculo-e-vincular-ao-cliente.md) | — |
| 5 | A persona da operação combinada é o Gestor; as demais tarefas do contexto têm o Mecânico. Confirmar quem cadastra veículo na oficina. | [`cadastrar-veiculo-e-vincular-ao-cliente.md`](cadastrar-veiculo-e-vincular-ao-cliente.md) | — |
| 6 | Formato da placa: o padrão Mercosul (`ABC1D23`) e o antigo (`ABC1234`) convivem na frota. Definir o validador. | [`consultar-veiculo.md`](consultar-veiculo.md) e [`cadastrar-veiculo.md`](cadastrar-veiculo.md) | — |
| 7 | Faixa de validação do campo `ano`: mínimo, e se ano futuro é permitido. | [`cadastrar-veiculo.md`](cadastrar-veiculo.md) | — |
| 8 | Alteração de placa em veículo com Ordens de Serviço existentes: permitida ou bloqueada? | [`atualizar-veiculo.md`](atualizar-veiculo.md) | — |
| 9 | Nenhuma operação de escrita usa `If-Match` com `version`, enquanto Peças & Insumos e Serviços usam controle otimista. Definir se entra aqui. | [`atualizar-veiculo.md`](atualizar-veiculo.md) | — |
| 10 | A reativação com cliente inativo devolve `422`. Confirmar o código depois de fechar a padronização de `409` e `422`. | [`deletar-veiculo.md`](deletar-veiculo.md) | — |
| 11 | A exclusão lógica depende de índice parcial `UNIQUE (placa) WHERE ativo = true` e da ausência de `ON DELETE CASCADE` nas foreign keys de OS. Confirmar com quem cuidar da migration. | [`deletar-veiculo.md`](deletar-veiculo.md) | — |
| 12 | Não existe histórico de proprietários: ao trocar o dono do veículo, o vínculo anterior se perde. Definir se o MVP precisa desse histórico. | — | — |
| 13 | Confirmar o dono do contexto e preencher o campo `dono` dos cinco documentos, hoje em "A definir". | Todos os documentos do contexto | — |
