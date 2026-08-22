---
documento: Pontos Cobertos — Checklist de Refinamento
dono: José Lázaro
versao: 0.7
atualizado_em: 2026-08-22
status: em andamento
---

# Pontos Cobertos

Checklist de todas as tarefas levantadas para a aplicação. Marcado quer dizer **documentada**
(refinamento de produto, refinamento técnico e checklist de implementação escritos), não
implementada em código.

## Resumo

| Contexto | Documentadas | Total |
|---|---|---|
| Cliente | 5 | 5 |
| Veículo | 4 | 4 |
| Ordem de Serviço | 11 | 15 |
| Orçamento | 4 | 8 |
| Serviços | 4 | 4 |
| Peças & Insumos | 10 | 14 |
| **Total** | **38** | **50** |

---

## Cliente

- [x] Consultar cliente — [consultar-cliente.md](docs/cliente/consultar-cliente.md)
- [x] Cadastrar cliente — [cadastrar-cliente.md](docs/cliente/cadastrar-cliente.md)
- [x] Atualizar cliente — [atualizar-cliente.md](docs/cliente/atualizar-cliente.md)
- [x] Vincular cliente a veículo — [vincular-veiculo-ao-cliente.md](docs/cliente/vincular-veiculo-ao-cliente.md)
- [x] Deletar cliente — [deletar-cliente.md](docs/cliente/deletar-cliente.md)

## Veículo

- [x] Consultar veículo — [consultar-veiculo.md](docs/veiculo/consultar-veiculo.md)
- [x] Cadastrar veículo — [cadastrar-veiculo.md](docs/veiculo/cadastrar-veiculo.md)
- [x] Deletar veículo — [deletar-veiculo.md](docs/veiculo/deletar-veiculo.md)
- [x] Atualizar veículo — [atualizar-veiculo.md](docs/veiculo/atualizar-veiculo.md)

## Ordem de Serviço

- [ ] Criar ordem de serviço — documento removido, será reescrito
- [x] Registrar problema relatado — [registrar-problema-relatado.md](docs/ordem-de-servico/registrar-problema-relatado.md)
- [x] Registrar problema encontrado — [registrar-problema-encontrado.md](docs/ordem-de-servico/registrar-problema-encontrado.md)
- [x] Registrar serviços necessários — [registrar-servicos-necessarios.md](docs/ordem-de-servico/registrar-servicos-necessarios.md)
- [x] Registrar peças e insumos necessários — [registrar-pecas-e-insumos-necessarios.md](docs/ordem-de-servico/registrar-pecas-e-insumos-necessarios.md), só com o refinamento de produto
- [x] Consultar fila de atendimentos — [consultar-fila-de-atendimento.md](docs/ordem-de-servico/consultar-fila-de-atendimento.md)
- [x] Iniciar execução — [iniciar-execucao.md](docs/ordem-de-servico/iniciar-execucao.md)
- [x] Finalizar serviço — [finalizar-servico.md](docs/ordem-de-servico/finalizar-servico.md)
- [x] Registrar entrega do veículo — [registrar-entrega-de-veiculo.md](docs/ordem-de-servico/registrar-entrega-de-veiculo.md)
- [x] Consultar ordem de serviço — [consultar-ordem-de-servico.md](docs/ordem-de-servico/consultar-ordem-de-servico.md)
- [x] Listar ordens de serviço — [listar-ordens-de-servico.md](docs/ordem-de-servico/listar-ordens-de-servico.md)
- [x] Monitorar tempo médio de execução — [monitorar-tempo-medio-de-execucao.md](docs/ordem-de-servico/monitorar-tempo-medio-de-execucao.md)
- [ ] Incluir OS na fila de atendimento — documento removido, será reescrito
- [ ] Selecionar próxima OS para execução
- [ ] Registrar problema adicional

## Orçamento

- [x] Calcular orçamento — [calcular-orcamento.md](docs/orcamento/calcular-orcamento.md)
- [x] Consultar orçamento — [consultar-orcamento.md](docs/orcamento/consultar-orcamento.md)
- [x] Aprovar orçamento — [aprovar-orcamento.md](docs/orcamento/aprovar-orcamento.md)
- [x] Recusar orçamento — [recusar-orcamento.md](docs/orcamento/recusar-orcamento.md)
- [ ] Enviar orçamento — documento removido, será reescrito
- [ ] Gerar orçamento complementar — documento removido, será reescrito
- [ ] Aprovar orçamento complementar
- [ ] Recusar orçamento complementar

## Serviços

- [x] Cadastrar serviço — [cadastrar-servico.md](docs/servicos/cadastrar-servico.md)
- [x] Consultar serviços — [consultar-servicos.md](docs/servicos/consultar-servicos.md)
- [x] Atualizar serviço — [atualizar-servico.md](docs/servicos/atualizar-servico.md)
- [x] Remover ou desativar serviço — [desativar-servico.md](docs/servicos/desativar-servico.md)

## Peças & Insumos

- [x] Cadastrar peça — [cadastrar-peca.md](docs/pecas-e-insumos/cadastrar-peca.md)
- [x] Atualizar peça — [atualizar-peca.md](docs/pecas-e-insumos/atualizar-peca.md)
- [x] Deletar peça — [deletar-peca.md](docs/pecas-e-insumos/deletar-peca.md)
- [x] Cadastrar insumo — [cadastrar-insumo.md](docs/pecas-e-insumos/cadastrar-insumo.md)
- [x] Atualizar insumo — [atualizar-insumo.md](docs/pecas-e-insumos/atualizar-insumo.md)
- [x] Deletar insumo — [deletar-insumo.md](docs/pecas-e-insumos/deletar-insumo.md)
- [x] Registrar entrada de estoque — [registrar-entrada-de-estoque.md](docs/pecas-e-insumos/registrar-entrada-de-estoque.md)
- [x] Solicitar compra de peças — [solicitar-compra-de-pecas.md](docs/pecas-e-insumos/solicitar-compra-de-pecas.md)
- [x] Solicitar compra de insumos — [solicitar-compra-de-insumos.md](docs/pecas-e-insumos/solicitar-compra-de-insumos.md)
- [x] Retornar peça e insumo ao estoque — [retornar-peca-e-insumo-ao-estoque.md](docs/pecas-e-insumos/retornar-peca-e-insumo-ao-estoque.md)
- [ ] Consultar estoque — documento removido, será reescrito
- [ ] Consultar peças faltantes — documento removido, será reescrito
- [ ] Registrar consumo e saída — documento removido, será reescrito
- [ ] Reservar peças para OS — documento removido, será reescrito

---

## Observações

- A lista original tinha 52 tarefas. Quatro linhas dela — registrar peças necessárias, registrar
  insumos necessários, registrar peças adicionais e registrar insumos adicionais — viraram uma
  tarefa só, **Registrar peças e insumos necessários**, porque o refinamento trata os quatro casos
  no mesmo fluxo, distinguindo adição principal de complementar. Entrou também **Retornar peça e
  insumo ao estoque**, que não estava na lista. O total ficou em 50.
- **Registrar peças e insumos necessários** está marcada, mas veio só com o refinamento de produto:
  o refinamento técnico e o checklist estão pendentes, e a lista do que falta está no próprio
  documento.
- **Documentos removidos e a reescrever:** criar ordem de serviço, incluir OS na fila, enviar
  orçamento, gerar orçamento complementar, consultar estoque, consultar peças faltantes, registrar
  consumo e saída, e reservar peças para OS. As rotas correspondentes estão listadas em
  [03-endpoints.md](docs/03-endpoints.md), na seção *Rotas sem documento no momento*.
- **Iniciar diagnóstico deixou de existir como tarefa.** Registrar problema relatado a substitui:
  o mesmo endpoint grava o relato do cliente, marca a data de início e muda a OS para
  `EM_DIAGNOSTICO`.
- Marcado quer dizer documentado, não implementado.
