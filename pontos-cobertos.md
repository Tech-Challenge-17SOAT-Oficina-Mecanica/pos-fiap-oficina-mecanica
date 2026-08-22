---
documento: Pontos Cobertos — Checklist de Refinamento
dono: José Lázaro
versao: 0.2
atualizado_em: 2026-08-19
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
| Veículo | 3 | 3 |
| Ordem de Serviço | 6 | 18 |
| Orçamento | 5 | 8 |
| Serviços | 0 | 4 |
| Peças & Insumos | 13 | 13 |
| **Total** | **32** | **51** |

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

## Ordem de Serviço

- [x] Criar ordem de serviço — [criar-ordem-de-servico.md](docs/ordem-de-servico/criar-ordem-de-servico.md)
- [x] Iniciar diagnóstico — [iniciar-diagnostico.md](docs/ordem-de-servico/iniciar-diagnostico.md)
- [x] Registrar serviços necessários — [registrar-servicos-necessarios.md](docs/ordem-de-servico/registrar-servicos-necessarios.md)
- [x] Consultar fila de atendimentos — [consultar-fila-de-atendimento.md](docs/ordem-de-servico/consultar-fila-de-atendimento.md)
- [x] Iniciar execução — [iniciar-execucao.md](docs/ordem-de-servico/iniciar-execucao.md)
- [x] Monitorar tempo médio de execução — [monitorar-tempo-medio-de-execucao.md](docs/ordem-de-servico/monitorar-tempo-medio-de-execucao.md)
- [ ] Registrar peças necessárias
- [ ] Registrar insumos necessários
- [ ] Registrar diagnóstico
- [ ] Incluir OS na fila de atendimento
- [ ] Selecionar próxima OS para execução
- [ ] Registrar problema adicional
- [ ] Registrar peças adicionais
- [ ] Registrar insumos adicionais
- [ ] Finalizar serviço
- [ ] Registrar entrega do veículo
- [ ] Consultar ordem de serviço
- [ ] Listar ordens de serviço

## Orçamento

- [x] Calcular orçamento — [calcular-orcamento.md](docs/orcamento/calcular-orcamento.md)
- [x] Consultar orçamento — [consultar-orcamento.md](docs/orcamento/consultar-orcamento.md)
- [x] Aprovar orçamento — [aprovar-orcamento.md](docs/orcamento/aprovar-orcamento.md)
- [x] Recusar orçamento — [recusar-orcamento.md](docs/orcamento/recusar-orcamento.md)
- [x] Gerar orçamento complementar — [gerar-orcamento-complementar.md](docs/orcamento/gerar-orcamento-complementar.md)
- [ ] Enviar orçamento
- [ ] Aprovar orçamento complementar
- [ ] Recusar orçamento complementar

## Serviços

- [ ] Cadastrar serviço
- [ ] Consultar serviços
- [ ] Atualizar serviço
- [ ] Remover ou desativar serviço

## Peças & Insumos

- [x] Consultar estoque — [consultar-estoque.md](docs/pecas-e-insumos/consultar-estoque.md)
- [x] Cadastrar peça — [cadastrar-peca.md](docs/pecas-e-insumos/cadastrar-peca.md)
- [x] Atualizar peça — [atualizar-peca.md](docs/pecas-e-insumos/atualizar-peca.md)
- [x] Deletar peça — [deletar-peca.md](docs/pecas-e-insumos/deletar-peca.md)
- [x] Cadastrar insumo — [cadastrar-insumo.md](docs/pecas-e-insumos/cadastrar-insumo.md)
- [x] Atualizar insumo — [atualizar-insumo.md](docs/pecas-e-insumos/atualizar-insumo.md)
- [x] Deletar insumo — [deletar-insumo.md](docs/pecas-e-insumos/deletar-insumo.md)
- [x] Registrar entrada de estoque — [registrar-entrada-de-estoque.md](docs/pecas-e-insumos/registrar-entrada-de-estoque.md)
- [x] Reservar peças para OS — [reservar-peca-para-os.md](docs/pecas-e-insumos/reservar-peca-para-os.md)
- [x] Registrar consumo e saída — [registrar-consumo-e-saida.md](docs/pecas-e-insumos/registrar-consumo-e-saida.md)
- [x] Consultar peças faltantes — [consultar-pecas-faltantes.md](docs/pecas-e-insumos/consultar-pecas-faltantes.md)
- [x] Solicitar compra de peças — [solicitar-compra-de-pecas.md](docs/pecas-e-insumos/solicitar-compra-de-pecas.md)
- [x] Solicitar compra de insumos — [solicitar-compra-de-insumos.md](docs/pecas-e-insumos/solicitar-compra-de-insumos.md)

---

## Observações

- A lista original repetia "iniciar execução" duas vezes e trazia "registrar entrega de veículo"
  e "registrar entrega" como itens separados. Foram unificados, o que leva o total de 52 linhas
  para 50 tarefas distintas.
- **Monitorar tempo médio de execução** não estava na lista original, mas foi refinada e entrou no
  checklist — é o indicador exigido pelo enunciado do Tech Challenge. Com ela, o total sobe para 51.
- "Iniciar diagnóstico" e "registrar diagnóstico" seguem separados: o primeiro é a transição de
  status da OS e já está documentado; o segundo é o registro do resultado do diagnóstico e
  continua pendente. Se forem a mesma tarefa, vale unificar.
- "Incluir OS na fila de atendimento" e "selecionar próxima OS para execução" continuam pendentes,
  mas já aparecem como pré-condição e como fluxo em
  [consultar-fila-de-atendimento.md](docs/ordem-de-servico/consultar-fila-de-atendimento.md) e
  [iniciar-execucao.md](docs/ordem-de-servico/iniciar-execucao.md). Vale confirmar se são tarefas
  próprias ou parte dessas duas.
- Aprovar e recusar **orçamento complementar** ainda não têm documento próprio. Os documentos de
  aprovar e recusar cobrem só o orçamento principal, e o efeito da decisão sobre um complementar
  está registrado como ponto em aberto do contexto de Orçamento.
- Todos os contextos documentados seguem a mesma estrutura: uma pasta por contexto, com um arquivo
  por tarefa e um `pontos-em-aberto.md`.
- O maior bloco pendente é Ordem de Serviço, com 12 tarefas — a maior parte do fluxo de execução,
  do registro do diagnóstico até a entrega do veículo.
