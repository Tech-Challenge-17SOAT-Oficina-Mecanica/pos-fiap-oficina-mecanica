---
documento: Catálogo de Endpoints
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-20
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
| `DELETE` | `/veiculos/{veiculoId}` | Inativa o veículo (exclusão lógica) | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |
| `POST` | `/veiculos/{veiculoId}/reativacao` | Reativa um veículo inativado | `veiculos:escrever` | [deletar-veiculo.md](veiculo/deletar-veiculo.md) |

## Ordem de Serviço

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `POST` | `/ordens-servico` | Cria a Ordem de Serviço para um cliente e veículo | `os:escrever` | [criar-ordem-de-servico.md](ordem-de-servico/criar-ordem-de-servico.md) |
| `PATCH` | `/ordens-servico/{osId}/diagnostico/iniciar` | Inicia o diagnóstico da OS | `os:escrever` | [iniciar-diagnostico.md](ordem-de-servico/iniciar-diagnostico.md) |
| `POST` | `/ordens-servico/{osId}/servicos` | Registra os serviços necessários na OS | `os:escrever` | [registrar-servicos-necessarios.md](ordem-de-servico/registrar-servicos-necessarios.md) |
| `GET` | `/fila-atendimento` | Lista as OS aptas para execução, da mais antiga para a mais recente | `os:ler` | [consultar-fila-de-atendimento.md](ordem-de-servico/consultar-fila-de-atendimento.md) |
| `POST` | `/ordens-servico/{osId}/execucao/iniciar` | Inicia a execução dos serviços autorizados | `os:escrever` | [iniciar-execucao.md](ordem-de-servico/iniciar-execucao.md) |
| `GET` | `/ordens-servico/{osId}/tempo-execucao` | Retorna o tempo de execução de uma OS | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |
| `GET` | `/ordens-servico/tempos-execucao` | Lista os tempos de execução e o tempo médio do período | `os:ler` | [monitorar-tempo-medio-de-execucao.md](ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |

## Orçamento

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| — | *(sem endpoint)* | Gera o orçamento principal; disparado internamente ao fim do diagnóstico | — | [gerar-orcamento.md](orcamento/gerar-orcamento.md) |
| `GET` | `/orcamentos` | Consulta orçamentos por identificador ou pelo documento do cliente | `orcamentos:ler` | [consultar-orcamento.md](orcamento/consultar-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/aprovar` | Cliente aprova o orçamento e libera a OS para execução | `orcamentos:aprovar` | [aprovar-orcamento.md](orcamento/aprovar-orcamento.md) |
| `POST` | `/orcamentos/{orcamentoId}/recusar` | Cliente recusa o orçamento e a OS é cancelada | `orcamentos:aprovar` | [recusar-orcamento.md](orcamento/recusar-orcamento.md) |
| `POST` | `/ordens-servico/{osId}/orcamentos-complementares` | Gera orçamento complementar a partir dos itens adicionais da OS | `orcamentos:escrever` | [gerar-orcamento-complementar.md](orcamento/gerar-orcamento-complementar.md) |

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
| `GET` | `/estoque/insumos/{insumoId}/sugestao-compra` | Calcula a quantidade sugerida de compra do insumo | `estoque:ler` | [solicitar-compra-de-insumos.md](pecas-e-insumos/solicitar-compra-de-insumos.md) |
| `POST` | `/estoque/entradas` | Registra o recebimento de peças e insumos | `estoque:movimentar` | [registrar-entrada-de-estoque.md](pecas-e-insumos/registrar-entrada-de-estoque.md) |
| `POST` | `/estoque/reservas` | Reserva as peças de uma OS aprovada | `estoque:movimentar` | [reservar-peca-para-os.md](pecas-e-insumos/reservar-peca-para-os.md) |
| `DELETE` | `/estoque/reservas/ordens-servico/{osId}` | Libera as reservas ativas de uma OS | `estoque:movimentar` | [reservar-peca-para-os.md](pecas-e-insumos/reservar-peca-para-os.md) |
| `POST` | `/estoque/saidas` | Dá baixa nas peças e insumos usados no serviço | `estoque:movimentar` | [registrar-consumo-e-saida.md](pecas-e-insumos/registrar-consumo-e-saida.md) |
| `POST` | `/compras/pedidos` | Cria pedido de compra de peças ou de insumos | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](pecas-e-insumos/solicitar-compra-de-insumos.md) |
| `DELETE` | `/compras/pedidos/{pedidoId}` | Cancela um pedido de compra ainda não recebido | `compras:escrever` | [solicitar-compra-de-pecas.md](pecas-e-insumos/solicitar-compra-de-pecas.md) |

## Serviços

| Método | Rota | O que faz | Escopo | Documento |
|---|---|---|---|---|
| `GET` | `/servicos` | Lista serviços cadastrados no catálogo, com filtros e paginação | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `GET` | `/servicos/{id}` | Consulta os detalhes de um serviço específico | `servicos:ler` | [consultar-servicos.md](servicos/consultar-servicos.md) |
| `POST` | `/servicos` | Cadastra um novo serviço no catálogo | `servicos:escrever` | [cadastrar-servico.md](servicos/cadastrar-servico.md) |
| `PATCH` | `/servicos/{id}` | Atualiza os dados cadastrais de um serviço | `servicos:escrever` | [atualizar-servico.md](servicos/atualizar-servico.md) |

---

## Resumo

| Contexto | Endpoints |
|---|---|
| Cliente | 6 |
| Veículo | 4 |
| Ordem de Serviço | 7 |
| Orçamento | 4 |
| Peças & Insumos | 15 |
| Serviços | 4 |
| **Total** | **40** |

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
| 3 | A desativação usa `DELETE` em Peças & Insumos, Cliente e Veículo, mas o refinamento de Serviços propunha `PATCH /servicos/{id}/desativar`. Padronizar quando o contexto de Serviços for escrito. | [deletar-peca.md](pecas-e-insumos/deletar-peca.md) |
| 4 | Gerar Orçamento não tem endpoint: é disparado internamente ao fim do diagnóstico. Confirmar se o MVP precisa de uma rota administrativa para reprocessar a geração. | [gerar-orcamento.md](orcamento/gerar-orcamento.md) |
| 5 | Não existe rota de consulta ou listagem de OS, embora sejam tarefas previstas e o cliente precise acompanhar o progresso pela API. | [pontos-cobertos.md](../pontos-cobertos.md) |
