---
documento: Pontos em Aberto — Contexto de Insumos
dono: A definir
versao: 2.1
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Insumos

## O que é este documento

Um ponto em aberto é uma **decisão que ainda não foi tomada** ou uma **inconsistência encontrada**
entre os documentos deste contexto. Enquanto o ponto estiver aberto, quem for implementar não deve
resolver sozinho: a escolha muda contrato de API, modelo de dados ou regra de negócio, e precisa
valer para o time inteiro.

Para fechar um ponto:

1. aplique a decisão nos documentos afetados;
2. registre o porquê — como _Decisão de projeto_ no documento da tarefa, ou em
   [02-decisoes-arquiteturais.md](../02-decisoes-arquiteturais.md) quando valer para todos os contextos;
3. mova a linha para a tabela de decisões abaixo.

O retrato do que já está decidido está em [00-resumo.md](00-resumo.md).

## Inconsistências a corrigir

Nenhuma. Os três pontos que estavam aqui foram fechados em 22/08/2026, junto com as decisões
arquiteturais, e estão registrados abaixo.

## Decisões desta rodada

Peças e Insumos decidem em conjunto, e as decisões que valem para os dois estão registradas nos
dois arquivos, cada uma escrita do ponto de vista do seu contexto.

| #   | Ponto                                                                                 | Decisão                                                                                                                                                                                                                                                                                                                                             | Onde foi aplicada                                                                                                                                                                                                                                              |
| --- | ------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | O CRUD de fornecedor estava aprovado mas sem documento.                               | **Escrito.** Quatro tarefas no contexto de Peças, dono do agregado de Compras: cadastrar, consultar, atualizar, e desativar com reativação. Documento imutável, unicidade entre ativos, exclusão lógica, e desativação bloqueada com pedido de compra em aberto. Entrou o escopo `compras:ler`.                                                     | [cadastrar-fornecedor.md](../pecas/cadastrar-fornecedor.md), [consultar-fornecedores.md](../pecas/consultar-fornecedores.md), [atualizar-fornecedor.md](../pecas/atualizar-fornecedor.md), [desativar-fornecedor.md](../pecas/desativar-fornecedor.md) e DT-48 |
| 2   | Com a D-16 fechada, as rotas de reserva direta ficaram sem chamador público.          | **Aposentadas.** Revisando a tarefa, o processamento já faz tudo o que a reserva direta fazia, e mais a compra do faltante. `POST /estoque/reservas` e `POST /estoque/reservas-insumos` saíram da API; a reserva virou **serviço de domínio** chamado pelo processamento. Os documentos continuam valendo como especificação das regras de reserva. | [reservar-peca-para-os.md](../pecas/reservar-peca-para-os.md), [reservar-insumo-para-os.md](../insumos/reservar-insumo-para-os.md), [03-endpoints.md](../03-endpoints.md) e DT-47                                                                              |
| 3   | Registrar consumo e saída continuava sem documento — a maior lacuna do projeto.       | **Escrito, em uma rota só:** `POST /estoque/saidas`, compartilhada pelos dois contextos, como a entrada e a compra. A baixa **consome a reserva**, nunca o saldo livre, reduz o saldo físico, devolve ao saldo livre o reservado e não usado, e informa o custo à OS. Agora existe baixa de estoque na execução do serviço.                         | [registrar-consumo-e-saida-de-pecas.md](../pecas/registrar-consumo-e-saida-de-pecas.md), [registrar-consumo-e-saida-de-insumos.md](../insumos/registrar-consumo-e-saida-de-insumos.md) e DT-49                                                                 |
| 4   | O mesmo produto em `L` e em `ML` continuava sendo dois itens, sem regra de conversão. | **Cada unidade de medida é um item independente, sem conversão.** Comprar 1 L não aumenta o saldo do item em mililitro, e a baixa de um não afeta o outro. A oficina cadastra o insumo na unidade em que compra e consome.                                                                                                                          | [cadastrar-insumo.md](cadastrar-insumo.md), [00-resumo.md](00-resumo.md) e DT-46                                                                                                                                                                               |
| 5   | `categoria` era obrigatória e entrava na regra de duplicidade, mas continuava texto livre. | **Virou tabela (D-09).** A escrita informa `categoriaId`; a leitura devolve `categoriaId` e o nome. A unicidade passou a ser `UNIQUE (categoria_id, unidade_medida, descricao_normalizada) WHERE ativo = true`. Entrou `GET /estoque/categorias` na lista de rotas sem documento. | [cadastrar-insumo.md](cadastrar-insumo.md), [atualizar-insumo.md](atualizar-insumo.md), [consultar-insumos.md](consultar-insumos.md), [03-endpoints.md](../03-endpoints.md) e D-09 |
| 6   | `DELETE /estoque/reservas/ordens-servico/{osId}` continuava sem documento. | **Aposentada.** A rota saiu do catálogo: a liberação da reserva acontece dentro da devolução na recusa do orçamento e dentro da baixa de consumo, que devolve ao saldo livre o reservado e não usado. | Seção *Rotas aposentadas* de [03-endpoints.md](../03-endpoints.md) |
| 7   | Faltava a consulta de insumos faltantes, equivalente à de peças. | **A consulta de itens faltantes saiu do MVP inteira**, nos dois contextos — não só a de insumos: a de peças, que já estava escrita, foi removida junto. A falta continua sendo apurada dentro do processamento disparado pela aprovação do orçamento, que reserva o disponível e abre pedido do restante. | [03-endpoints.md](../03-endpoints.md), [00-resumo.md](00-resumo.md), [../../pontos-cobertos.md](../../pontos-cobertos.md) e DT-50 |
| 8   | O `422` convivia com o `409` para a mesma classe de erro. | **O `422` saiu da API (D-01).** Insumo inativo, troca de unidade com saldo, item sem reserva e quantidade acima da reservada viraram `409`; item do tipo errado no endpoint virou `400`. | Todas as tarefas do contexto e D-01 |
| 10  | O custo do insumo era atualizado pela entrada, mas a tarefa de entrada não dizia como. | **Último custo (D-14).** Cada recebimento sobrescreve o `custo_unitario` do insumo com o valor daquela nota. A regra estava só no resumo do contexto; entrou no documento que a executa. | [registrar-entrada-de-insumos.md](registrar-entrada-de-insumos.md) e D-14 |
| 11  | A atualização de insumo exigia `If-Match` mas não dizia o que fazer quando o header não vem. | **`428`.** Mesma correção feita em peça: `412` para divergência, `428` para ausência. | [atualizar-insumo.md](atualizar-insumo.md), [../pecas/atualizar-peca.md](../pecas/atualizar-peca.md) e D-24 |
| 9   | A listagem não devolvia `version`, exigida pelo `If-Match` do `PUT`. | **Corrigido (D-10).** `GET /estoque/insumos` e o detalhe passaram a devolver `version`. Peça ganhou a rota de detalhe que insumo já tinha. | [consultar-insumos.md](consultar-insumos.md), [../pecas/consultar-pecas.md](../pecas/consultar-pecas.md) e D-10 |
