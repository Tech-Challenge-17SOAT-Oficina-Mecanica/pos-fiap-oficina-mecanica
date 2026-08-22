---
documento: Catálogo de Endpoints
dono: José Lázaro
versao: 0.4
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
| `PUT` | `/veiculos/{veiculoId}` | Atualiza os dados cadastrais do veículo | `veiculos:escrever` | [atualizar-veiculo.md](veiculo/atualizar-veiculo.md) |
| `DELETE` | `/veiculos/{veiculoId}` | Inativa o veículo (exclusão lógica) | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |
| `POST` | `/veiculos/{veiculoId}/reativacao` | Reativa um veículo inativado | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |

## Ordem de Serviço

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/ordens-servico` | Cria a Ordem de Serviço para um cliente e veículo | `os:escrever` | [criar-ordem-de-servico.md](ordem-de-servico/criar-ordem-de-servico.md) |
| `POST` | `/ordens-servico/{osId}/problema-relatado` | Registra o relato do cliente e inicia o diagnóstico | `os:escrever` | [registrar-problema-relatado.md](ordem-de-servico/registrar-problema-relatado.md) |
| `POST` | `/ordens-servico/{osId}/problemas` | Registra problema encontrado e vincula ao orçamento aplicável | `os:escrever` | [registrar-problema-encontrado.md](ordem-de-servico/registrar-problema-encontrado.md) |
| `POST` | `/ordens-servico/{osId}/servicos` | Registra os serviços necessários na OS | `os:escrever` | [registrar-servicos-necessarios.md](ordem-de-servico/registrar-servicos-necessarios.md) |
| `GET` | `/fila-atendimento` | Lista as OS aptas para execução, com as do mecânico responsável primeiro | `os:ler` | [consultar-fila-de-atendimento.md](ordem-de-servico/consultar-fila-de-atendimento.md) |
| — | *(sem endpoint)* | Coloca a OS na fila após a aprovação de um orçamento; processamento interno | — | [incluir-os-na-fila-de-atendimento.md](ordem-de-servico/incluir-os-na-fila-de-atendimento.md) |
| `POST` | `/ordens-servico/{osId}/execucao/iniciar` | Inicia a execução dos serviços autorizados | `os:escrever` | [iniciar-execucao.md](ordem-de-servico/iniciar-execucao.md) |
| `GET` | `/ordens-servico/{osId}/tempo-execucao` | Retorna o tempo de execução de uma OS | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |
| `GET` | `/ordens-servico/tempos-execucao` | Lista os tempos de execução e o tempo médio do período | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |
| `POST` | `/ordens-servico/{osId}/finalizar` | Finaliza os serviços autorizados e notifica o cliente | `os:escrever` | [finalizar-servico.md](ordem-de-servico/finalizar-servico.md) |
| `POST` | `/ordens-servico/{osId}/entrega` | Registra o pagamento e a entrega do veículo, encerrando a OS | `os:escrever` | [registrar-entrega-de-veiculo.md](ordem-de-servico/registrar-entrega-de-veiculo.md) |
| `GET` | `/ordens-servico/{osId}` | Detalha a OS com cliente, veículo, problemas, orçamentos e histórico de eventos | `os:ler` | [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| `GET` | `/ordens-servico` | Lista OS com filtro por status, documento do cliente e placa | `os:ler` | [listar-ordens-de-servico.md](ordem-de-servico/listar-ordens-de-servico.md) |

## Orçamento

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/orcamentos/{orcamentoId}/calcular` | Calcula os itens, o valor total geral e a estimativa de entrega de um orçamento existente | `orcamentos:escrever` | [calcular-orcamento.md](orcamento/calcular-orcamento.md) |
| `GET` | `/orcamentos` | Consulta orçamentos por identificador ou pelo documento do cliente | `orcamentos:ler` | [consultar-orcamento.md](orcamento/consultar-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/aprovar` | Cliente aprova o orçamento e libera a OS para execução | `orcamentos:aprovar` | [aprovar-orcamento.md](orcamento/aprovar-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/recusar` | Cliente recusa o orçamento e a OS é cancelada | `orcamentos:aprovar` | [recusar-orcamento.md](orcamento/recusar-orcamento.md) |
| `POST` | `/ordens-servico/{osId}/orcamentos-complementares` | Gera orçamento complementar a partir dos itens adicionais da OS | `orcamentos:escrever` | [gerar-orcamento-complementar.md](orcamento/gerar-orcamento-complementar.md) |
| — | *(sem endpoint)* | Envia o orçamento ao cliente; processamento interno após o cálculo | — | [enviar-orcamento.md](orcamento/enviar-orcamento.md) |

## Peças & Insumos

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/estoque/itens` | Consulta peças e insumos com saldo físico, reservado e disponível | `estoque:ler` | [consultar-estoque.md](pecas-e-insumos/consultar-estoque.md) |
| `GET` | `/estoque/itens/faltantes` | Lista itens abaixo do mínimo ou demandados por OS sem saldo | `estoque:ler` | [consultar-pecas-faltantes.md](pecas-e-insumos/consultar-pecas-faltantes.md) |
| `POST` | `/estoque/pecas` | Cadastra uma nova peça no catálogo | `estoque:escrever` | [cadastrar-peca.md](pecas-e-insumos/cadastrar-peca.md) |
| `PUT` | `/estoque/pecas/{pecaId}` | Atualiza os dados cadastrais da peça | `estoque:escrever` | [atualizar-peca.md](pecas-e-insumos/atualizar-peca.md) |
| `DELETE` | `/estoque/pecas/{pecaId}` | Desativa a peça (exclusão lógica) | `estoque:escrever` | [deletar-peca.md](pecas-e-insumos/deletar-peca.md) |
| `POST` | `/estoque/insumos` | Cadastra um novo insumo no catálogo | `estoque:escrever` | [cadastrar-insumo.md](pecas-e-insumos/cadastrar-insumo.md) |
| `PUT` | `/estoque/insumos/{insumoId}` | Atualiza os dados cadastrais do insumo | `estoque:escrever` | [atualizar-insumo.md](pecas-e-insumos/atualizar-insumo.md) |
| `DELETE` | `/estoque/insumos/{insumoId}` | Desativa o insumo (exclusão lógica) | `estoque:escrever` | [deletar-insumo.md](pecas-e-insumos/deletar-insumo.md) |
| `POST` | `/estoque/entradas` | Registra o recebimento de peças e insumos | `estoque:movimentar` | [registrar-entrada-de-estoque.md](pecas-e-insumos/registrar-entrada-de-estoque.md) |
| `POST` | `/estoque/reservas` | Reserva as peças de uma OS aprovada | `estoque:movimentar` | [reservar-peca-para-os.md](pecas-e-insumos/reservar-peca-para-os.md) |
| `DELETE` | `/estoque/reservas/ordens-servico/{osId}` | Libera as reservas ativas de uma OS | `estoque:movimentar` | [reservar-peca-para-os.md](pecas-e-insumos/reservar-peca-para-os.md) |
| `POST` | `/estoque/saidas` | Dá baixa nas peças e insumos usados no serviço | `estoque:movimentar` | [registrar-consumo-e-saida.md](pecas-e-insumos/registrar-consumo-e-saida.md) |
| `POST` | `/compras/pedidos` | Cria pedido de compra de peças ou insumos, reserva os itens para as OS e as coloca em `AGUARDANDO_RECURSOS` | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](pecas-e-insumos/solicitar-compra-de-insumos.md) |
| `DELETE` | `/compras/pedidos/{pedidoId}` | Cancela um pedido de compra ainda não recebido | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) |

## Serviços

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/servicos` | Lista serviços cadastrados no catálogo, com filtros e paginação | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `GET` | `/servicos/{id}` | Consulta os detalhes de um serviço específico | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `POST` | `/servicos` | Cadastra um novo serviço no catálogo | `servicos:escrever` | [cadastrar-servico.md](servicos/cadastrar-servico.md) |
| `PATCH` | `/servicos/{id}` | Atualiza os dados cadastrais de um serviço | `servicos:escrever` | [atualizar-servico.md](servicos/atualizar-servico.md) |
| `PATCH` | `/servicos/{servicoId}/desativar` | Desativa o serviço no catálogo, preservando o histórico | `servicos:escrever` | [desativar-servico.md](servicos/desativar-servico.md) |
---

## Resumo

| Contexto | Endpoints |
|---|---|
| Cliente | 6 |
| Veículo | 5 |
| Ordem de Serviço | 12 |
| Orçamento | 5 |
| Peças & Insumos | 14 |
| Serviços | 5 |
| **Total** | **47** |

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
| 2 | `POST /compras/pedidos` atende peça e insumo na mesma rota, diferenciando pelo tipo do item. Confirmar se fica assim ou se haverá rotas separadas. | [solicitar-compra-de-insumos.md](pecas-e-insumos/solicitar-compra-de-insumos.md) |
| 3 | A desativação usa `DELETE` em Peças & Insumos, Cliente e Veículo, mas Serviços usa `PATCH /servicos/{servicoId}/desativar`. Padronizar o verbo da exclusão lógica entre os contextos. | [deletar-peca.md](pecas-e-insumos/deletar-peca.md) e [desativar-servico.md](servicos/desativar-servico.md) |
| 4 | Gerar Orçamento não tem endpoint: é disparado internamente ao fim do diagnóstico. Confirmar se o MVP precisa de uma rota administrativa para reprocessar a geração. | [gerar-orcamento.md](orcamento/gerar-orcamento.md) |
| 5 | `GET /ordens-servico` serve a listagem e também a busca das OS de um cliente por CPF/CNPJ, enquanto `GET /ordens-servico/{osId}` faz o detalhamento. Confirmar essa divisão entre listar e consultar. | [listar-ordens-de-servico.md](ordem-de-servico/listar-ordens-de-servico.md) e [consultar-ordem-de-servico.md](ordem-de-servico/consultar-ordem-de-servico.md) |
| 6 | A entrega do veículo depende da confirmação do pagamento, mas pagamento não é um contexto documentado nem tem rota. Definir se entra no MVP. | [registrar-entrega-de-veiculo.md](ordem-de-servico/registrar-entrega-de-veiculo.md) |
| 7 | O path param da OS aparecia como `{id}`, `{ordemServicoId}` e `{osId}` nos refinamentos. Foi padronizado `{osId}` — alinhar os documentos que estão sendo reenviados. | Todas as rotas de `/ordens-servico` |
| 8 | `PATCH /ordens-servico/{osId}/diagnostico/iniciar` saiu da tabela porque o início do diagnóstico virou consequência de `POST /ordens-servico/{osId}/problema-relatado`, mas o documento `iniciar-diagnostico.md` continua no repositório. Confirmar qual das duas tarefas vale e remover a outra. | [registrar-problema-relatado.md](ordem-de-servico/registrar-problema-relatado.md) e [iniciar-diagnostico.md](ordem-de-servico/iniciar-diagnostico.md) |
| 9 | O documento de compra de peças propunha `POST /compras/pedidos/pecas`, enquanto o de insumos manteve a rota compartilhada `POST /compras/pedidos`. Ficou a rota compartilhada — confirmar. | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) |
| 10 | `GET /estoque/insumos/{insumoId}/sugestao-compra` saiu do catálogo: a nova versão da compra de insumos tira o cálculo por estoque mínimo e consumo médio, e usa a necessidade apurada nas OS. Confirmar se a sugestão volta como tarefa própria. | [solicitar-compra-de-insumos.md](pecas-e-insumos/solicitar-compra-de-insumos.md) |