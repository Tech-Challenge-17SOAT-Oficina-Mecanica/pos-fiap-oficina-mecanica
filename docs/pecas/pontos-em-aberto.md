---
documento: Pontos em Aberto — Contexto de Peças
dono: A definir
versao: 2.1
atualizado_em: 2026-08-22
status: em construcao
---

# Pontos em Aberto — Peças

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

Nenhuma. Os quatro pontos que estavam aqui foram fechados em 22/08/2026, junto com as decisões
arquiteturais, e estão registrados abaixo.

## Decisões desta rodada

Peças e Insumos decidem em conjunto, e as decisões que valem para os dois estão registradas nos
dois arquivos, cada uma escrita do ponto de vista do seu contexto.

| #   | Ponto                                                                           | Decisão                                                                                                                                                                                                                                                                                                                                             | Onde foi aplicada                                                                                                                                                                                                                                              |
| --- | ------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | O CRUD de fornecedor estava aprovado mas sem documento.                         | **Escrito.** Quatro tarefas no contexto de Peças, dono do agregado de Compras: cadastrar, consultar, atualizar, e desativar com reativação. Documento imutável, unicidade entre ativos, exclusão lógica, e desativação bloqueada com pedido de compra em aberto. Entrou o escopo `compras:ler`.                                                     | [cadastrar-fornecedor.md](../pecas/cadastrar-fornecedor.md), [consultar-fornecedores.md](../pecas/consultar-fornecedores.md), [atualizar-fornecedor.md](../pecas/atualizar-fornecedor.md), [desativar-fornecedor.md](../pecas/desativar-fornecedor.md) e DT-48 |
| 2   | Com a D-16 fechada, as rotas de reserva direta ficaram sem chamador público.    | **Aposentadas.** Revisando a tarefa, o processamento já faz tudo o que a reserva direta fazia, e mais a compra do faltante. `POST /estoque/reservas` e `POST /estoque/reservas-insumos` saíram da API; a reserva virou **serviço de domínio** chamado pelo processamento. Os documentos continuam valendo como especificação das regras de reserva. | [reservar-peca-para-os.md](../pecas/reservar-peca-para-os.md), [reservar-insumo-para-os.md](../insumos/reservar-insumo-para-os.md), [03-endpoints.md](../03-endpoints.md) e DT-47                                                                              |
| 3   | Registrar consumo e saída continuava sem documento — a maior lacuna do projeto. | **Escrito, em uma rota só:** `POST /estoque/saidas`, compartilhada pelos dois contextos, como a entrada e a compra. A baixa **consome a reserva**, nunca o saldo livre, reduz o saldo físico, devolve ao saldo livre o reservado e não usado, e informa o custo à OS. Agora existe baixa de estoque na execução do serviço.                         | [registrar-consumo-e-saida-de-pecas.md](../pecas/registrar-consumo-e-saida-de-pecas.md), [registrar-consumo-e-saida-de-insumos.md](../insumos/registrar-consumo-e-saida-de-insumos.md) e DT-49                                                                 |
| 4   | Consultar peças faltantes continuava sem documento. | **Escrita e depois removida.** O documento chegou a existir, com `GET /estoque/pecas/faltantes`, saldo, mínimo, necessidade apurada e sugestão de compra — e a fórmula da sugestão fechou a D-06. Na rodada seguinte a equipe tirou a consulta de itens faltantes do MVP, e ela foi removida junto com o documento. Ver o ponto 7. | D-06 e DT-50 |
| 5   | `categoria` era obrigatória e entrava na regra de duplicidade, mas continuava texto livre. | **Virou tabela (D-09).** A escrita informa `categoriaId`; a leitura devolve `categoriaId` e o nome. A unicidade passou a ser `UNIQUE (categoria_id, descricao_normalizada) WHERE ativo = true`. Entrou `GET /estoque/categorias` na lista de rotas sem documento: sem ela ninguém descobre o identificador. | [cadastrar-peca.md](cadastrar-peca.md), [atualizar-peca.md](atualizar-peca.md), [consultar-pecas.md](consultar-pecas.md), [03-endpoints.md](../03-endpoints.md) e D-09 |
| 6   | `DELETE /estoque/reservas/ordens-servico/{osId}` continuava sem documento. | **Aposentada.** A rota saiu do catálogo: a liberação da reserva acontece dentro da devolução na recusa do orçamento e dentro da baixa de consumo, que devolve ao saldo livre o reservado e não usado. | Seção *Rotas aposentadas* de [03-endpoints.md](../03-endpoints.md) |
| 7   | Faltava a consulta de insumos faltantes, equivalente à de peças. | **A consulta de itens faltantes saiu do MVP inteira**, nos dois contextos. `GET /estoque/pecas/faltantes` foi aposentada e o documento removido. A falta continua sendo apurada onde importa: dentro do processamento disparado pela aprovação do orçamento, que reserva o disponível e abre pedido do restante. | [03-endpoints.md](../03-endpoints.md), [00-resumo.md](00-resumo.md), [../../pontos-cobertos.md](../../pontos-cobertos.md) e DT-50 |
| 8   | O escopo `compras:ler` foi criado para a consulta de fornecedores sem passar pelo time. | **Confirmado.** O nome fica, e já está na lista oficial de escopos. | Seção 8 do [01-guia-de-documentacao.md](../01-guia-de-documentacao.md) |
| 9   | O `422` convivia com o `409` para a mesma classe de erro. | **O `422` saiu da API (D-01).** Peça inativa, quantidade menor que a necessidade e item sem reserva viraram `409`; item do tipo errado no endpoint virou `400`. | Todas as tarefas do contexto e D-01 |
| 10  | As listagens não devolviam `version`, e peça não tinha rota de detalhe. | **Corrigido (D-10).** `GET /estoque/pecas` passou a devolver `version`, e nasceu `GET /estoque/pecas/{pecaId}`, espelhando a de insumo. A paginação da consulta também foi corrigida de `page`/`size` para `pagina`/`tamanho`. | [consultar-pecas.md](consultar-pecas.md), [03-endpoints.md](../03-endpoints.md), D-10 e D-21 |
| 12  | A atualização de peça exigia `If-Match` mas não dizia o que fazer quando o header não vem. | **`428`.** A D-24 já previa o par `412` para divergência e `428` para ausência; faltava o segundo nas tarefas de atualização de peça e de insumo. | [atualizar-peca.md](atualizar-peca.md), [../insumos/atualizar-insumo.md](../insumos/atualizar-insumo.md) e D-24 |
| 11  | O prazo de entrega do fornecedor não existia em lugar nenhum. | **Entrou como `prazoEntregaDias`**, em português, no cadastro de fornecedor, com padrão de 7 dias. É informativo para quem compra: a sugestão de quantidade não depende dele. | [cadastrar-fornecedor.md](cadastrar-fornecedor.md), [atualizar-fornecedor.md](atualizar-fornecedor.md), [consultar-fornecedores.md](consultar-fornecedores.md), D-07 e D-08 |
