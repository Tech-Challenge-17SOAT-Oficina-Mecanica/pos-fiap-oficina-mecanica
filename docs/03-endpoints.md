---
documento: Catálogo de Endpoints
dono: José Lázaro
versao: 0.5
atualizado_em: 2026-08-22
status: em construcao
---

# Catálogo de Endpoints

Visão única de todas as rotas da API, com o método, o caminho, o que a rota faz, o escopo exigido
e o documento onde ela está refinada.

**Este documento é vivo.** Toda tarefa nova que expõe uma rota entra aqui no mesmo PR em que o
documento da tarefa é criado. Ele existe para responder rápido a três perguntas: essa rota já
existe? quem é o dono dela? o caminho está no padrão?

**Convenções em vigor**

- As rotas **não usam prefixo de versão**: o recurso começa na raiz, como `/clientes`.
- Recursos no plural, em minúsculas, com hífen quando houver mais de uma palavra: `/ordens-servico`.
- Identificadores em rota são sempre **UUID**.
- Ações que não são CRUD aparecem como sub-recurso com verbo de negócio, como
  `/ordens-servico/{osId}/execucao/iniciar`.
- Autenticação por `Bearer <JWT>` em todas as rotas, com autorização por escopo no formato
  `recurso:acao`.

---

## Cliente

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/clientes` | Consulta cliente por CPF/CNPJ, com os veículos vinculados | `clientes:ler` | [consultar-cliente.md](cliente/consultar-cliente.md) |
| `POST` | `/clientes` | Cadastra um novo cliente | `clientes:escrever` | [cadastrar-cliente.md](cliente/cadastrar-cliente.md) |
| `PUT` | `/clientes/{clienteId}` | Atualiza os dados cadastrais do cliente | `clientes:escrever` | [atualizar-cliente.md](cliente/atualizar-cliente.md) |
| `DELETE` | `/clientes/{clienteId}` | Inativa o cliente (exclusão lógica) e os veículos vinculados | `clientes:escrever` | [deletar-cliente.md](cliente/deletar-cliente.md) |
| `POST` | `/clientes/{clienteId}/reativacao` | Reativa um cliente inativado | `clientes:escrever` | [deletar-cliente.md](cliente/deletar-cliente.md) |
| `POST` | `/clientes/{clienteId}/veiculos/{veiculoId}` | Vincula um veículo ao cliente | `clientes:escrever` | [vincular-veiculo-ao-cliente.md](cliente/vincular-veiculo-ao-cliente.md) |

## Veículo

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/veiculos` | Consulta veículo pela placa | `veiculos:ler` | [consultar-veiculo.md](veiculo/consultar-veiculo.md) |
| `POST` | `/veiculos` | Cadastra um novo veículo | `veiculos:escrever` | [cadastrar-veiculo.md](veiculo/cadastrar-veiculo.md) |
| `POST` | `/clientes/{clienteId}/veiculos` | Cadastra um veículo e o vincula ao cliente na mesma operação | `veiculos:escrever` | [cadastrar-veiculo-e-vincular-ao-cliente.md](veiculo/cadastrar-veiculo-e-vincular-ao-cliente.md) |
| `PUT` | `/veiculos/{veiculoId}` | Atualiza os dados cadastrais do veículo | `veiculos:escrever` | [atualizar-veiculo.md](veiculo/atualizar-veiculo.md) |
| `DELETE` | `/veiculos/{veiculoId}` | Inativa o veículo (exclusão lógica) | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |
| `POST` | `/veiculos/{veiculoId}/reativacao` | Reativa um veículo inativado | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |

## Ordem de Serviço

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/ordens-servico/{osId}/problema-relatado` | Registra o relato do cliente e inicia o diagnóstico | `os:escrever` | [registrar-problema-relatado.md](ordem-de-servico/registrar-problema-relatado.md) |
| `POST` | `/ordens-servico/{osId}/problemas` | Registra problema encontrado e vincula ao orçamento aplicável | `os:escrever` | [registrar-problema-encontrado.md](ordem-de-servico/registrar-problema-encontrado.md) |
| `POST` | `/ordens-servico/{osId}/servicos` | Registra os serviços necessários no orçamento da etapa atual | `os:escrever` | [registrar-servicos-necessarios.md](ordem-de-servico/registrar-servicos-necessarios.md) |
| `GET` | `/fila-atendimento` | Lista as OS aptas para execução, com as do mecânico responsável primeiro | `os:ler` | [consultar-fila-de-atendimento.md](ordem-de-servico/consultar-fila-de-atendimento.md) |
| `POST` | `/ordens-servico/{osId}/execucao/iniciar` | Inicia a execução dos serviços autorizados | `os:escrever` | [iniciar-execucao.md](ordem-de-servico/iniciar-execucao.md) |
| `POST` | `/ordens-servico/{osId}/finalizar` | Finaliza os serviços autorizados e notifica o cliente | `os:escrever` | [finalizar-servico.md](ordem-de-servico/finalizar-servico.md) |
| `POST` | `/ordens-servico/{osId}/entrega` | Registra o pagamento e a entrega do veículo, encerrando a OS | `os:escrever` | [registrar-entrega-de-veiculo.md](ordem-de-servico/registrar-entrega-de-veiculo.md) |
| `GET` | `/ordens-servico/{osId}` | Detalha a OS com cliente, veículo, problemas, orçamentos e histórico de eventos | `os:ler` | [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| `GET` | `/ordens-servico` | Lista OS com filtro por status, documento do cliente e placa | `os:ler` | [listar-ordens-de-servico.md](ordem-de-servico/listar-ordens-de-servico.md) |
| `?` | *(a definir)* | Registra peças e insumos necessários na OS e na adição do orçamento; endpoint ainda não refinado | `os:escrever` | [registrar-pecas-e-insumos-necessarios.md](ordem-de-servico/registrar-pecas-e-insumos-necessarios.md) |
| `GET` | `/ordens-servico/{osId}/tempo-execucao` | Retorna o tempo de execução de uma OS | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |
| `GET` | `/ordens-servico/tempos-execucao` | Lista os tempos de execução e o tempo médio do período | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |

## Orçamento

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/orcamentos/{orcamentoId}/calcular` | Calcula os itens, o valor total geral e a estimativa de entrega do orçamento | `orcamentos:escrever` | [calcular-orcamento.md](orcamento/calcular-orcamento.md) |
| `GET` | `/orcamentos` | Consulta orçamentos por identificador ou pelo documento do cliente | `orcamentos:ler` | [consultar-orcamento.md](orcamento/consultar-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/aprovar` | Cliente aprova o orçamento e libera a OS para execução | `orcamentos:aprovar` | [aprovar-orcamento.md](orcamento/aprovar-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/recusar` | Cliente recusa o orçamento e a OS é cancelada | `orcamentos:recusar` | [recusar-orcamento.md](orcamento/recusar-orcamento.md) |

## Serviços

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/servicos` | Lista serviços do catálogo, com filtros e paginação | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `GET` | `/servicos/{servicoId}` | Consulta os detalhes de um serviço | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `POST` | `/servicos` | Cadastra um novo serviço no catálogo | `servicos:escrever` | [cadastrar-servico.md](servicos/cadastrar-servico.md) |
| `PATCH` | `/servicos/{servicoId}` | Atualiza os dados cadastrais de um serviço | `servicos:escrever` | [atualizar-servico.md](servicos/atualizar-servico.md) |
| `PATCH` | `/servicos/{servicoId}/desativar` | Desativa o serviço, preservando o histórico | `servicos:escrever` | [desativar-servico.md](servicos/desativar-servico.md) |

## Peças & Insumos

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/estoque/pecas` | Consulta peças com saldo físico, reservado e disponível | `estoque:ler` | [consultar-estoque.md](pecas-e-insumos/consultar-estoque.md) |
| `POST` | `/estoque/pecas` | Cadastra uma nova peça no catálogo | `estoque:escrever` | [cadastrar-peca.md](pecas-e-insumos/cadastrar-peca.md) |
| `PUT` | `/estoque/pecas/{pecaId}` | Atualiza os dados cadastrais da peça | `estoque:escrever` | [atualizar-peca.md](pecas-e-insumos/atualizar-peca.md) |
| `DELETE` | `/estoque/pecas/{pecaId}` | Desativa a peça (exclusão lógica) | `estoque:escrever` | [deletar-peca.md](pecas-e-insumos/deletar-peca.md) |
| `POST` | `/estoque/insumos` | Cadastra um novo insumo no catálogo | `estoque:escrever` | [cadastrar-insumo.md](pecas-e-insumos/cadastrar-insumo.md) |
| `GET` | `/estoque/insumos` | Consulta insumos por filtros, quantidade desejada e disponibilidade | `estoque:ler` | [consultar-insumos.md](pecas-e-insumos/consultar-insumos.md) |
| `GET` | `/estoque/insumos/{insumoId}` | Consulta um insumo específico e sua disponibilidade | `estoque:ler` | [consultar-insumos.md](pecas-e-insumos/consultar-insumos.md) |
| `PUT` | `/estoque/insumos/{insumoId}` | Atualiza os dados cadastrais do insumo | `estoque:escrever` | [atualizar-insumo.md](pecas-e-insumos/atualizar-insumo.md) |
| `DELETE` | `/estoque/insumos/{insumoId}` | Desativa o insumo (exclusão lógica) | `estoque:escrever` | [deletar-insumo.md](pecas-e-insumos/deletar-insumo.md) |
| `POST` | `/estoque/entradas` | Registra o recebimento, efetiva as reservas do pedido e libera as OS sem itens pendentes | `estoque:movimentar` | [registrar-entrada-de-estoque.md](pecas-e-insumos/registrar-entrada-de-estoque.md) |
| — | *(sem endpoint)* | Devolve peças e insumos ao estoque na recusa do orçamento; chamada em processo dentro de `RecusarOrcamento` | — | [retornar-peca-e-insumo-ao-estoque.md](pecas-e-insumos/retornar-peca-e-insumo-ao-estoque.md) |
| `POST` | `/compras/pedidos` | Cria pedido de compra de peças ou insumos, reserva os itens para as OS e as coloca em `AGUARDANDO_RECURSOS` | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](pecas-e-insumos/solicitar-compra-de-insumos.md) |
| `DELETE` | `/compras/pedidos/{pedidoId}` | Cancela um pedido de compra ainda não recebido e libera as reservas | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) |

---

## Resumo

| Contexto | Endpoints |
|---|---|
| Cliente | 6 |
| Veículo | 6 |
| Ordem de Serviço | 11 |
| Orçamento | 4 |
| Serviços | 5 |
| Peças & Insumos | 11 |
| **Total** | **43** |

---

## Rotas sem documento no momento

Estas rotas já existiram no catálogo, mas o documento que as define foi retirado do repositório
para ser reescrito. Elas voltam para a tabela junto com o documento novo — e se alguma deixar de
existir, é uma decisão que precisa ser registrada.

| Rota | Tarefa | Contexto |
|---|---|---|
| `POST /ordens-servico` | Criar ordem de serviço | Ordem de Serviço |
| `POST /ordens-servico/{osId}/orcamentos-complementares` | Gerar orçamento complementar | Orçamento |
| *(sem endpoint)* | Enviar orçamento | Orçamento |
| `GET /estoque/itens` | Consultar estoque | Peças & Insumos |
| `GET /estoque/itens/faltantes` | Consultar peças faltantes | Peças & Insumos |
| `POST /estoque/saidas` | Registrar consumo e saída | Peças & Insumos |
| `POST /estoque/reservas` | Reservar peça para OS | Peças & Insumos |
| `DELETE /estoque/reservas/ordens-servico/{osId}` | Reservar peça para OS (liberação) | Peças & Insumos |

---

## Como manter este arquivo

1. Ao refinar uma tarefa nova, copie o método e a rota exatamente como estão no bloco `http` do
   documento da tarefa.
2. Preencha o escopo com o mesmo valor da seção **Autenticação / Autorização** do documento.
3. Aponte o link para o documento da tarefa, não para a pasta do contexto.
4. Uma tarefa que expõe duas rotas gera duas linhas.
5. Atualize o resumo e a data do frontmatter.
6. Antes de criar uma rota nova, procure aqui: se já existir algo parecido, reaproveite o recurso
   em vez de abrir um caminho paralelo.

## Pontos em aberto sobre rotas

| # | Ponto | Onde |
|---|---|---|
| 1 | `GET /fila-atendimento` está na raiz, fora de `/ordens-servico`. Definir se a fila é recurso próprio ou uma visão da OS. | [consultar-fila-de-atendimento.md](ordem-de-servico/consultar-fila-de-atendimento.md) |
| 2 | `POST /compras/pedidos` atende peça e insumo na mesma rota, diferenciando pelo tipo do item, e o fornecedor é obrigatório para peça e opcional para insumo. Confirmar a assimetria. | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](pecas-e-insumos/solicitar-compra-de-insumos.md) |
| 3 | A desativação usa `DELETE` em Cliente, Veículo e Peças & Insumos, mas Serviços usa `PATCH /servicos/{servicoId}/desativar`. Padronizar o verbo da exclusão lógica. | [deletar-peca.md](pecas-e-insumos/deletar-peca.md) e [desativar-servico.md](servicos/desativar-servico.md) |
| 4 | `GET /ordens-servico` serve a listagem e também a busca das OS de um cliente por CPF/CNPJ, enquanto `GET /ordens-servico/{osId}` faz o detalhamento. Confirmar essa divisão. | [listar-ordens-de-servico.md](ordem-de-servico/listar-ordens-de-servico.md) e [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| 5 | A entrega do veículo depende da confirmação do pagamento, mas pagamento não é um contexto documentado nem tem rota. Definir se entra no MVP. | [registrar-entrega-de-veiculo.md](ordem-de-servico/registrar-entrega-de-veiculo.md) |
| 6 | O path param aparece como `{id}` nos documentos de Serviços e como `{servicoId}` na desativação. A tabela padronizou `{servicoId}` — alinhar os documentos. | [consultar-servicos.md](servicos/consultar-servicos.md) e [atualizar-servico.md](servicos/atualizar-servico.md) |
| 7 | Escopos de decisão do cliente: aprovar usa `orcamentos:aprovar` e recusar usa `orcamentos:recusar`. Confirmar se são dois escopos mesmo ou um só. | [aprovar-orcamento.md](orcamento/aprovar-orcamento.md) e [recusar-orcamento.md](orcamento/recusar-orcamento.md) |
| 8 | Não existe rota para **criar** a Ordem de Serviço enquanto o documento não voltar, embora todas as outras rotas de OS dependam de uma OS existente. | Ordem de Serviço |
| 9 | O registro de peças e insumos necessários na OS ainda não tem endpoint definido: o documento sugere rotas separadas por tipo de item, mas o refinamento técnico não foi escrito. | [registrar-pecas-e-insumos-necessarios.md](ordem-de-servico/registrar-pecas-e-insumos-necessarios.md) |
| 10 | Nenhuma rota consulta o estoque hoje, mas a fila de atendimento e a finalização da OS validam disponibilidade de peças e insumos. | Peças & Insumos |
