---
documento: Pontos Cobertos — Checklist de Refinamento
dono: José Lázaro
versao: 0.4
atualizado_em: 2026-08-19
status: em andamento
---

# Pontos Cobertos

Checklist de todas as tarefas levantadas para a aplicação. Marcado quer dizer **documentada**
(refinamento de produto, refinamento técnico e checklist de implementação escritos), não
implementada em código.

## Resumo

| Contexto | Documentadas | Aguardando reenvio | Total |
|---|---|---|---|
| Cliente | 5 | 0 | 5 |
| Veículo | 3 | 0 | 3 |
| Ordem de Serviço | 5 | 5 | 18 |
| Orçamento | 0 | 5 | 8 |
| Serviços | 1 | 0 | 4 |
| Peças & Insumos | 6 | 7 | 13 |
| **Total** | **20** | **17** | **51** |

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

- [ ] Criar ordem de serviço — `criar-ordem-de-servico.md` aguardando reenvio
- [ ] Iniciar diagnóstico — `iniciar-diagnostico.md` aguardando reenvio
- [ ] Registrar serviços necessários — `registrar-servicos-necessarios.md` aguardando reenvio
- [ ] Consultar fila de atendimentos — `consultar-fila-de-atendimento.md` aguardando reenvio
- [ ] Iniciar execução — `iniciar-execucao.md` aguardando reenvio
- [x] Monitorar tempo médio de execução — [monitorar-tempo-medio-de-execucao.md](docs/ordem-de-servico/monitorar-tempo-medio-de-execucao.md)
- [ ] Registrar peças necessárias
- [ ] Registrar insumos necessários
- [ ] Registrar diagnóstico
- [ ] Incluir OS na fila de atendimento
- [ ] Selecionar próxima OS para execução
- [ ] Registrar problema adicional
- [ ] Registrar peças adicionais
- [ ] Registrar insumos adicionais
- [x] Finalizar serviço — [finalizar-servico.md](docs/ordem-de-servico/finalizar-servico.md)
- [x] Registrar entrega do veículo — [registrar-entrega-de-veiculo.md](docs/ordem-de-servico/registrar-entrega-de-veiculo.md)
- [x] Consultar ordem de serviço — [consultar-ordem-de-servico.md](docs/ordem-de-servico/consultar-ordem-de-servico.md)
- [x] Listar ordens de serviço — [listar-ordens-de-servico.md](docs/ordem-de-servico/listar-ordens-de-servico.md)

## Orçamento

- [ ] Gerar orçamento — `gerar-orcamento.md` aguardando reenvio
- [ ] Consultar orçamento — `consultar-orcamento.md` aguardando reenvio
- [ ] Aprovar orçamento — `aprovar-orcamento.md` aguardando reenvio
- [ ] Recusar orçamento — `recusar-orcamento.md` aguardando reenvio
- [ ] Gerar orçamento complementar — `gerar-orcamento-complementar.md` aguardando reenvio
- [ ] Enviar orçamento
- [ ] Aprovar orçamento complementar
- [ ] Recusar orçamento complementar

## Serviços

- [ ] Cadastrar serviço
- [ ] Consultar serviços
- [ ] Atualizar serviço
- [x] Remover ou desativar serviço — [desativar-servico.md](docs/servicos/desativar-servico.md)

## Peças & Insumos

- [ ] Consultar estoque — `consultar-estoque.md` aguardando reenvio
- [x] Cadastrar peça — [cadastrar-peca.md](docs/pecas-e-insumos/cadastrar-peca.md)
- [x] Atualizar peça — [atualizar-peca.md](docs/pecas-e-insumos/atualizar-peca.md)
- [x] Deletar peça — [deletar-peca.md](docs/pecas-e-insumos/deletar-peca.md)
- [x] Cadastrar insumo — [cadastrar-insumo.md](docs/pecas-e-insumos/cadastrar-insumo.md)
- [x] Atualizar insumo — [atualizar-insumo.md](docs/pecas-e-insumos/atualizar-insumo.md)
- [x] Deletar insumo — [deletar-insumo.md](docs/pecas-e-insumos/deletar-insumo.md)
- [ ] Registrar entrada de estoque — `registrar-entrada-de-estoque.md` aguardando reenvio
- [ ] Reservar peças para OS — `reservar-peca-para-os.md` aguardando reenvio
- [ ] Registrar consumo e saída — `registrar-consumo-e-saida.md` aguardando reenvio
- [ ] Consultar peças faltantes — `consultar-pecas-faltantes.md` aguardando reenvio
- [ ] Solicitar compra de peças — `solicitar-compra-de-pecas.md` aguardando reenvio
- [ ] Solicitar compra de insumos — `solicitar-compra-de-insumos.md` aguardando reenvio

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
- **Reenvio em andamento.** Dezenove documentos foram retirados do repositório para serem
  substituídos por versões atualizadas: a pasta de Orçamento inteira, cinco tarefas de Ordem de
  Serviço e sete de Peças & Insumos, mais os `pontos-em-aberto.md` de Orçamento e de Ordem de
  Serviço. As 17 tarefas correspondentes estão **desmarcadas** e sinalizadas como *aguardando
  reenvio*, para voltarem a contar quando os arquivos novos chegarem. As rotas dessas tarefas
  continuam listadas em [03-endpoints.md](docs/03-endpoints.md), na seção *Documentos aguardando
  reenvio*.
- O que falta refinar de fato: oito tarefas de Ordem de Serviço (registro do diagnóstico, peças e
  insumos necessários e adicionais, problema adicional, fila e seleção da próxima OS), três de
  Orçamento (enviar orçamento, aprovar e recusar complementar) e três de Serviços (cadastrar,
  consultar e atualizar).
