---
documento: Pontos Cobertos — Checklist de Refinamento
dono: José Lázaro
versao: 0.8
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
| Veículo | 5 | 5 |
| Ordem de Serviço | 13 | 15 |
| Orçamento | 4 | 8 |
| Serviços | 4 | 4 |
| Peças & Insumos | 16 | 18 |
| **Total** | **47** | **55** |

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
- [x] Cadastrar veículo e vincular ao cliente — [cadastrar-veiculo-e-vincular-ao-cliente.md](docs/veiculo/cadastrar-veiculo-e-vincular-ao-cliente.md)

## Ordem de Serviço

- [x] Criar ordem de serviço — [criar-ordem-de-servico.md](docs/ordem-de-servico/criar-ordem-de-servico.md)
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
- [x] Incluir OS na fila de atendimento — [incluir-os-na-fila-de-atendimento.md](docs/ordem-de-servico/incluir-os-na-fila-de-atendimento.md)
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
- [x] Consultar peças — [consultar-estoque.md](docs/pecas-e-insumos/consultar-estoque.md)
- [x] Cadastrar insumo — [cadastrar-insumo.md](docs/pecas-e-insumos/cadastrar-insumo.md)
- [x] Atualizar insumo — [atualizar-insumo.md](docs/pecas-e-insumos/atualizar-insumo.md)
- [x] Deletar insumo — [deletar-insumo.md](docs/pecas-e-insumos/deletar-insumo.md)
- [x] Consultar insumos — [consultar-insumos.md](docs/pecas-e-insumos/consultar-insumos.md)
- [x] Reservar peças para OS — [reservar-peca-para-os.md](docs/pecas-e-insumos/reservar-peca-para-os.md)
- [x] Reservar insumos para OS — [reservar-insumo-para-os.md](docs/pecas-e-insumos/reservar-insumo-para-os.md)
- [x] Processar peças para reserva e compra — [processar-pecas-para-reserva-e-compra.md](docs/pecas-e-insumos/processar-pecas-para-reserva-e-compra.md)
- [x] Processar insumos para reserva e compra — [processar-insumos-para-reserva-e-compra.md](docs/pecas-e-insumos/processar-insumos-para-reserva-e-compra.md)
- [x] Registrar entrada de estoque — [registrar-entrada-de-estoque.md](docs/pecas-e-insumos/registrar-entrada-de-estoque.md)
- [x] Solicitar compra de peças — [solicitar-compra-de-pecas.md](docs/pecas-e-insumos/solicitar-compra-de-pecas.md)
- [x] Solicitar compra de insumos — [solicitar-compra-de-insumos.md](docs/pecas-e-insumos/solicitar-compra-de-insumos.md)
- [x] Retornar peça e insumo ao estoque — [retornar-peca-e-insumo-ao-estoque.md](docs/pecas-e-insumos/retornar-peca-e-insumo-ao-estoque.md)
- [ ] Consultar peças faltantes — documento removido, será reescrito
- [ ] Registrar consumo e saída — documento removido, será reescrito

---

## Observações

- A lista original tinha 52 tarefas. Ela mudou conforme os refinamentos chegaram: quatro linhas
  viraram a tarefa **Registrar peças e insumos necessários**, e entraram tarefas que não estavam
  previstas — retornar peça e insumo ao estoque, reservar insumo, os dois processamentos de reserva
  e compra, consultar insumos e cadastrar veículo já vinculado ao cliente. O total hoje é **55**.
- **Marcado quer dizer documentado, não implementado.** Duas tarefas marcadas ainda estão
  incompletas: registrar peças e insumos necessários veio só com o refinamento de produto, e
  consultar peças não tem seção de autenticação.
- **Faltam reescrever:** consultar peças faltantes e registrar consumo e saída, em Peças & Insumos;
  enviar orçamento e gerar orçamento complementar, em Orçamento. Sem *registrar consumo e saída*
  não existe baixa de estoque na execução do serviço.
- **Nunca foram refinadas:** selecionar próxima OS para execução e registrar problema adicional, em
  Ordem de Serviço; aprovar e recusar orçamento complementar, em Orçamento.
- O retrato de cada contexto está no `00-resumo.md` do respectivo diretório, e as inconsistências
  no `pontos-em-aberto.md`. As divergências que atravessam contextos estão em
  [02-decisoes-arquiteturais.md](docs/02-decisoes-arquiteturais.md), decisões D-16 a D-25.
