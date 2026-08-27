---
documento: Catálogo de Endpoints
dono: José Lázaro
versao: 1.5
atualizado_em: 2026-08-23
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

## Segurança

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/autenticacao/login` | Autentica mecânico e emite JWT | — | [autenticar-mecanico.md](seguranca/autenticar-mecanico.md) |

## Mecânico

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/mecanicos` | Cadastra mecânico e cria sua conta de acesso | `mecanicos:escrever` | [cadastrar-mecanico.md](mecanico/cadastrar-mecanico.md) |
| `PUT` | `/mecanicos/{mecanicoId}` | Atualiza cadastro e escopos de um mecânico | `mecanicos:escrever` | [atualizar-mecanico.md](mecanico/atualizar-mecanico.md) |

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
| `POST` | `/clientes/{clienteId}/veiculos` | Cadastra um veículo e o vincula ao cliente na mesma operação | `veiculos:escrever` | [cadastrar-veiculo-e-vincular-ao-cliente.md](veiculo/cadastrar-veiculo-e-vincular-ao-cliente.md) |
| `PUT` | `/veiculos/{veiculoId}` | Atualiza os dados cadastrais do veículo | `veiculos:escrever` | [atualizar-veiculo.md](veiculo/atualizar-veiculo.md) |
| `DELETE` | `/veiculos/{veiculoId}` | Inativa o veículo (exclusão lógica) | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |
| `POST` | `/veiculos/{veiculoId}/reativacao` | Reativa um veículo inativado | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |

## Ordem de Serviço

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/ordens-servico` | Cria a Ordem de Serviço para um cliente e veículo | `os:escrever` | [criar-ordem-de-servico.md](ordem-de-servico/criar-ordem-de-servico.md) |
| `POST` | `/ordens-servico/{osId}/problema-relatado` | Registra o relato do cliente e inicia o diagnóstico | `os:escrever` | [registrar-problema-relatado.md](ordem-de-servico/registrar-problema-relatado.md) |
| `POST` | `/ordens-servico/{osId}/problemas` | Registra problema encontrado e vincula ao orçamento aplicável | `os:escrever` | [registrar-problema-encontrado.md](ordem-de-servico/registrar-problema-encontrado.md) |
| `POST` | `/ordens-servico/{osId}/servicos` | Registra os serviços necessários no orçamento vigente da OS | `os:escrever` | [registrar-servicos-necessarios.md](ordem-de-servico/registrar-servicos-necessarios.md) |
| — | *(sem endpoint)* | Coloca a OS na fila após a aprovação do orçamento; caso de uso interno | — | [incluir-os-na-fila-de-atendimento.md](ordem-de-servico/incluir-os-na-fila-de-atendimento.md) |
| `GET` | `/fila-atendimento` | Lista as OS aptas para execução, com as do mecânico responsável primeiro | `os:ler` | [consultar-fila-de-atendimento.md](ordem-de-servico/consultar-fila-de-atendimento.md) |
| `POST` | `/ordens-servico/{osId}/execucao/iniciar` | Inicia a execução dos serviços autorizados | `os:escrever` | [iniciar-execucao.md](ordem-de-servico/iniciar-execucao.md) |
| `POST` | `/ordens-servico/{osId}/finalizar` | Finaliza os serviços autorizados e notifica o cliente | `os:escrever` | [finalizar-servico.md](ordem-de-servico/finalizar-servico.md) |
| `POST` | `/ordens-servico/{osId}/entrega` | Registra o valor final e a entrega do veículo, encerrando a OS | `os:escrever` | [registrar-entrega-de-veiculo.md](ordem-de-servico/registrar-entrega-de-veiculo.md) |
| `GET` | `/ordens-servico/{osId}` | Detalha a OS com cliente, veículo, problemas, orçamentos e histórico de eventos | `os:ler` | [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| `GET` | `/ordens-servico` | Lista OS com filtro por status, documento do cliente e placa | `os:ler` | [listar-ordens-de-servico.md](ordem-de-servico/listar-ordens-de-servico.md) |
| `POST` | `/ordens-servico/{osId}/pecas` | Registra as peças necessárias na OS e no orçamento vigente | `os:escrever` | [registrar-pecas-e-insumos-necessarios.md](ordem-de-servico/registrar-pecas-e-insumos-necessarios.md) |
| `POST` | `/ordens-servico/{osId}/insumos` | Registra os insumos necessários na OS e no orçamento vigente | `os:escrever` | [registrar-pecas-e-insumos-necessarios.md](ordem-de-servico/registrar-pecas-e-insumos-necessarios.md) |
| `GET` | `/ordens-servico/{osId}/orcamento` | Devolve todos os orçamentos da OS — principal e complementares — com seus itens | `os:ler` ou `orcamentos:ler` | [consultar-orcamento.md](orcamento/consultar-orcamento.md) |
| `GET` | `/ordens-servico/{osId}/tempo-execucao` | Retorna o tempo de execução de uma OS | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |
| `GET` | `/ordens-servico/tempos-execucao` | Lista os tempos de execução e o tempo médio do período | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |

## Orçamento

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/orcamentos/{orcamentoId}/calcular` | Calcula os itens, o valor total geral e a estimativa de entrega do orçamento | `orcamentos:escrever` | [calcular-orcamento.md](orcamento/calcular-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/aprovar` | Cliente aprova o orçamento, que dispara reserva e compra dos itens | `orcamentos:decidir` | [aprovar-orcamento.md](orcamento/aprovar-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/recusar` | Cliente recusa o orçamento; o principal cancela a OS, o complementar só é descartado | `orcamentos:decidir` | [recusar-orcamento.md](orcamento/recusar-orcamento.md) |

## Serviços

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/servicos` | Lista serviços do catálogo, com filtros e paginação | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `GET` | `/servicos/{servicoId}` | Consulta os detalhes de um serviço | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `POST` | `/servicos` | Cadastra um novo serviço no catálogo | `servicos:escrever` | [cadastrar-servico.md](servicos/cadastrar-servico.md) |
| `PATCH` | `/servicos/{servicoId}` | Atualiza os dados cadastrais de um serviço | `servicos:escrever` | [atualizar-servico.md](servicos/atualizar-servico.md) |
| `DELETE` | `/servicos/{servicoId}` | Inativa o serviço (exclusão lógica), preservando o histórico | `servicos:escrever` | [desativar-servico.md](servicos/desativar-servico.md) |
| `POST` | `/servicos/{servicoId}/reativacao` | Reativa um serviço inativado | `servicos:escrever` | [desativar-servico.md](servicos/desativar-servico.md) |

## Peças

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/estoque/pecas` | Consulta peças com saldo físico, reservado e disponível | `estoque:ler` | [consultar-pecas.md](pecas/consultar-pecas.md) |
| `GET` | `/estoque/pecas/{pecaId}` | Consulta uma peça específica, com `version` para o `If-Match` | `estoque:ler` | [consultar-pecas.md](pecas/consultar-pecas.md) |
| `POST` | `/estoque/pecas` | Cadastra uma nova peça no catálogo | `estoque:escrever` | [cadastrar-peca.md](pecas/cadastrar-peca.md) |
| `PUT` | `/estoque/pecas/{pecaId}` | Atualiza os dados cadastrais da peça | `estoque:escrever` | [atualizar-peca.md](pecas/atualizar-peca.md) |
| `DELETE` | `/estoque/pecas/{pecaId}` | Desativa a peça (exclusão lógica) | `estoque:escrever` | [deletar-peca.md](pecas/deletar-peca.md) |
| `POST` | `/estoque/solicitacoes-compra-reserva` | Reserva peças disponíveis e solicita compra do saldo faltante para uma OS | `estoque:movimentar` | [processar-pecas-para-reserva-e-compra.md](pecas/processar-pecas-para-reserva-e-compra.md) |
| — | *(sem endpoint)* | Reserva peças disponíveis para a OS; serviço de domínio chamado pelo processamento | — | [reservar-peca-para-os.md](pecas/reservar-peca-para-os.md) |
| — | *(sem endpoint)* | Devolve peças ao estoque na recusa do orçamento; chamada em processo dentro de `RecusarOrcamento` | — | [retornar-peca-ao-estoque.md](pecas/retornar-peca-ao-estoque.md) |

## Insumos

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/estoque/insumos` | Consulta insumos por filtros, quantidade desejada e disponibilidade | `estoque:ler` | [consultar-insumos.md](insumos/consultar-insumos.md) |
| `GET` | `/estoque/insumos/{insumoId}` | Consulta um insumo específico e sua disponibilidade | `estoque:ler` | [consultar-insumos.md](insumos/consultar-insumos.md) |
| `POST` | `/estoque/insumos` | Cadastra um novo insumo no catálogo | `estoque:escrever` | [cadastrar-insumo.md](insumos/cadastrar-insumo.md) |
| `PUT` | `/estoque/insumos/{insumoId}` | Atualiza os dados cadastrais do insumo | `estoque:escrever` | [atualizar-insumo.md](insumos/atualizar-insumo.md) |
| `DELETE` | `/estoque/insumos/{insumoId}` | Desativa o insumo (exclusão lógica) | `estoque:escrever` | [deletar-insumo.md](insumos/deletar-insumo.md) |
| `POST` | `/estoque/solicitacoes-compra-reserva-insumos` | Reserva insumos disponíveis e solicita compra do saldo faltante para uma OS | `estoque:movimentar` | [processar-insumos-para-reserva-e-compra.md](insumos/processar-insumos-para-reserva-e-compra.md) |
| — | *(sem endpoint)* | Reserva insumos disponíveis para a OS; serviço de domínio chamado pelo processamento | — | [reservar-insumo-para-os.md](insumos/reservar-insumo-para-os.md) |
| — | *(sem endpoint)* | Devolve insumos ao estoque na recusa do orçamento; chamada em processo dentro de `RecusarOrcamento` | — | [retornar-insumo-ao-estoque.md](insumos/retornar-insumo-ao-estoque.md) |

## Compras e recebimento

Rotas **compartilhadas** entre Peças e Insumos: o mesmo `pedido_compra` e o mesmo recebimento
atendem os dois tipos, porque o pedido e a nota fiscal reais misturam peça e insumo. O dono do
agregado de Compras é o contexto de **Peças**; Insumos referencia (DT-27 e DT-41).

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/estoque/entradas` | Registra o recebimento de peças e insumos, efetiva as reservas do pedido e libera as OS sem itens pendentes | `estoque:movimentar` | [registrar-entrada-de-pecas.md](pecas/registrar-entrada-de-pecas.md) e [registrar-entrada-de-insumos.md](insumos/registrar-entrada-de-insumos.md) |
| `POST` | `/compras/pedidos` | Cria pedido de compra de peças ou insumos, reserva os itens para as OS e as coloca em `AGUARDANDO_RECURSOS` | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](insumos/solicitar-compra-de-insumos.md) |
| `DELETE` | `/compras/pedidos/{pedidoId}` | Cancela um pedido de compra ainda não recebido e libera as reservas | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas/solicitar-compra-de-pecas.md) |
| `POST` | `/estoque/saidas` | Dá baixa das peças e insumos usados na execução, consumindo a reserva | `estoque:movimentar` | [registrar-consumo-e-saida-de-pecas.md](pecas/registrar-consumo-e-saida-de-pecas.md) e [registrar-consumo-e-saida-de-insumos.md](insumos/registrar-consumo-e-saida-de-insumos.md) |

## Fornecedor

Cadastro do agregado de **Compras**, cujo dono é o contexto de Peças. Atende os pedidos de compra
dos dois tipos de item.

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/fornecedores` | Lista fornecedores, com filtros e paginação | `compras:ler` | [consultar-fornecedores.md](pecas/consultar-fornecedores.md) |
| `GET` | `/fornecedores/{fornecedorId}` | Consulta um fornecedor específico | `compras:ler` | [consultar-fornecedores.md](pecas/consultar-fornecedores.md) |
| `POST` | `/fornecedores` | Cadastra um novo fornecedor | `compras:escrever` | [cadastrar-fornecedor.md](pecas/cadastrar-fornecedor.md) |
| `PUT` | `/fornecedores/{fornecedorId}` | Atualiza razão social, nome fantasia e contato | `compras:escrever` | [atualizar-fornecedor.md](pecas/atualizar-fornecedor.md) |
| `DELETE` | `/fornecedores/{fornecedorId}` | Inativa o fornecedor (exclusão lógica) | `compras:escrever` | [desativar-fornecedor.md](pecas/desativar-fornecedor.md) |
| `POST` | `/fornecedores/{fornecedorId}/reativacao` | Reativa um fornecedor inativado | `compras:escrever` | [desativar-fornecedor.md](pecas/desativar-fornecedor.md) |

---

## Resumo

| Contexto | Endpoints |
|---|---|
| Cliente | 6 |
| Veículo | 5 |
| Ordem de Serviço | 15 |
| Orçamento | 4 |
| Serviços | 6 |
| Peças | 6 |
| Insumos | 6 |
| Fornecedor (compartilhado) | 6 |
| Compras e recebimento (compartilhados) | 4 |
| **Total** | **58** |

---

## Rotas sem documento no momento

Estas rotas já existiram no catálogo, mas o documento que as define foi retirado do repositório
para ser reescrito. Elas voltam para a tabela junto com o documento novo — e se alguma deixar de
existir, é uma decisão que precisa ser registrada.

`POST /ordens-servico/{osId}/orcamentos-complementares` saiu desta lista: o complementar passou a
ser criado no registro de itens da OS, sem rota própria (DT-34).

| Rota | Tarefa | Contexto | Situação |
|---|---|---|---|
| `GET /estoque/categorias` | Consulta das categorias de item | Peças e Insumos | Nasceu da D-09, que trocou o texto livre por tabela de categorias: sem esta consulta ninguém descobre o `categoriaId` que o cadastro exige. Precisa de documento. |

`DELETE /estoque/reservas/ordens-servico/{osId}` saiu desta lista: a rota foi aposentada, porque a
liberação da reserva acontece dentro da devolução na recusa e da baixa de consumo.

---

## Rotas aposentadas

Rotas que saíram da API por decisão do time. Ficam registradas aqui para que ninguém as
reintroduza sem rever a decisão.

| Rota | Contexto | Substituída por | Decisão |
|---|---|---|---|
| `POST /veiculos` | Veículo | `POST /clientes/{clienteId}/veiculos` | Era o único caminho que permitia cadastrar veículo sem dono. O cadastro passa a acontecer sempre dentro do cliente. |
| `PATCH /servicos/{servicoId}/desativar` | Serviços | `DELETE /servicos/{servicoId}` | A exclusão lógica passou a usar o mesmo verbo dos demais contextos (D-20). |
| `GET /estoque/itens` | Peças e Insumos | `GET /estoque/pecas` e `GET /estoque/insumos` | A consulta unificada foi removida de propósito com a divisão dos contextos (DT-45). |
| `POST /estoque/reservas` | Peças | `POST /estoque/solicitacoes-compra-reserva` | Ficou sem chamador público com a D-16: a aprovação do orçamento chama o processamento, não a reserva direta. A reserva virou serviço de domínio. |
| `POST /estoque/reservas-insumos` | Insumos | `POST /estoque/solicitacoes-compra-reserva-insumos` | Mesma razão da rota de peças. |
| `GET /estoque/pecas/faltantes` | Peças | — | A consulta de itens faltantes saiu do MVP. A falta é apurada dentro do processamento disparado pela aprovação do orçamento, que reserva o disponível e abre pedido do restante. |
| `DELETE /estoque/reservas/ordens-servico/{osId}` | Peças e Insumos | — | Nunca teve documento nem caso de uso. A liberação de reserva acontece dentro da devolução na recusa e da baixa de consumo. |

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
| 1 | `GET /fila-atendimento` está na raiz, fora de `/ordens-servico`. Definir se a fila é recurso próprio ou visão da OS. | [consultar-fila-de-atendimento.md](ordem-de-servico/consultar-fila-de-atendimento.md) |
| 2 | Nenhuma listagem de peça ou insumo devolve `version`, e peça não tem rota de detalhe — sem isso não há como montar o `If-Match` que o `PUT` exige (D-10). | [consultar-pecas.md](pecas/consultar-pecas.md) e [consultar-insumos.md](insumos/consultar-insumos.md) |

Fechados nesta rodada: a convivência de `GET /ordens-servico/{osId}/orcamento` com `GET /orcamentos`
(as duas ficam, com papéis distintos — DT-57), o bloqueio da entrega por pagamento (fora do MVP —
D-25) e o escopo `compras:ler`, que já entrou na lista oficial da seção 8 do guia.
