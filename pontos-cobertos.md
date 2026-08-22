---
documento: Pontos Cobertos — Checklist de Refinamento
dono: José Lázaro
versao: 1.5
atualizado_em: 2026-08-22
status: em andamento
---

# Pontos Cobertos

Checklist de todas as tarefas levantadas para a aplicação. Marcado quer dizer **documentada**
(refinamento de produto, refinamento técnico e checklist de implementação escritos), não
implementada em código.

## Resumo

| Contexto         | Documentadas | Total  |
| ---------------- | ------------ | ------ |
| Cliente          | 5            | 5      |
| Veículo          | 5            | 5      |
| Ordem de Serviço | 13           | 13     |
| Orçamento        | 4            | 4      |
| Serviços         | 4            | 4      |
| Peças            | 14           | 14     |
| Insumos          | 10           | 10     |
| **Total**        | **55**       | **55** |

---

## Cliente

- [x] Consultar cliente — [consultar-cliente.md](docs/cliente/consultar-cliente.md)
- [x] Cadastrar cliente — [cadastrar-cliente.md](docs/cliente/cadastrar-cliente.md)
- [x] Atualizar cliente — [atualizar-cliente.md](docs/cliente/atualizar-cliente.md)
- [x] Vincular cliente a veículo — [vincular-veiculo-ao-cliente.md](docs/cliente/vincular-veiculo-ao-cliente.md)
- [x] Deletar cliente — [deletar-cliente.md](docs/cliente/deletar-cliente.md)

## Veículo

- [x] Consultar veículo — [consultar-veiculo.md](docs/veiculo/consultar-veiculo.md)
- [x] Cadastrar veículo — [cadastrar-veiculo.md](docs/veiculo/cadastrar-veiculo.md), rota `POST /veiculos` aposentada; o cadastro passou para `POST /clientes/{clienteId}/veiculos`
- [x] Deletar veículo — [deletar-veiculo.md](docs/veiculo/deletar-veiculo.md)
- [x] Atualizar veículo — [atualizar-veiculo.md](docs/veiculo/atualizar-veiculo.md)
- [x] Cadastrar veículo e vincular ao cliente — [cadastrar-veiculo-e-vincular-ao-cliente.md](docs/veiculo/cadastrar-veiculo-e-vincular-ao-cliente.md)

## Ordem de Serviço

- [x] Criar ordem de serviço — [criar-ordem-de-servico.md](docs/ordem-de-servico/criar-ordem-de-servico.md)
- [x] Registrar problema relatado — [registrar-problema-relatado.md](docs/ordem-de-servico/registrar-problema-relatado.md)
- [x] Registrar problema encontrado — [registrar-problema-encontrado.md](docs/ordem-de-servico/registrar-problema-encontrado.md)
- [x] Registrar serviços necessários — [registrar-servicos-necessarios.md](docs/ordem-de-servico/registrar-servicos-necessarios.md)
- [x] Registrar peças e insumos necessários — [registrar-pecas-e-insumos-necessarios.md](docs/ordem-de-servico/registrar-pecas-e-insumos-necessarios.md)
- [x] Consultar fila de atendimentos — [consultar-fila-de-atendimento.md](docs/ordem-de-servico/consultar-fila-de-atendimento.md)
- [x] Iniciar execução — [iniciar-execucao.md](docs/ordem-de-servico/iniciar-execucao.md)
- [x] Finalizar serviço — [finalizar-servico.md](docs/ordem-de-servico/finalizar-servico.md)
- [x] Registrar entrega do veículo — [registrar-entrega-de-veiculo.md](docs/ordem-de-servico/registrar-entrega-de-veiculo.md)
- [x] Consultar ordem de serviço — [consultar-ordem-de-servico.md](docs/ordem-de-servico/consultar-ordem-de-servico.md)
- [x] Listar ordens de serviço — [listar-ordens-de-servico.md](docs/ordem-de-servico/listar-ordens-de-servico.md)
- [x] Monitorar tempo médio de execução — [monitorar-tempo-medio-de-execucao.md](docs/ordem-de-servico/monitorar-tempo-medio-de-execucao.md)
- [x] Incluir OS na fila de atendimento — [incluir-os-na-fila-de-atendimento.md](docs/ordem-de-servico/incluir-os-na-fila-de-atendimento.md)

## Orçamento

- [x] Calcular orçamento — [calcular-orcamento.md](docs/orcamento/calcular-orcamento.md)
- [x] Consultar orçamento — [consultar-orcamento.md](docs/orcamento/consultar-orcamento.md)
- [x] Aprovar orçamento — [aprovar-orcamento.md](docs/orcamento/aprovar-orcamento.md)
- [x] Recusar orçamento — [recusar-orcamento.md](docs/orcamento/recusar-orcamento.md)

## Serviços

- [x] Cadastrar serviço — [cadastrar-servico.md](docs/servicos/cadastrar-servico.md)
- [x] Consultar serviços — [consultar-servicos.md](docs/servicos/consultar-servicos.md)
- [x] Atualizar serviço — [atualizar-servico.md](docs/servicos/atualizar-servico.md)
- [x] Desativar e reativar serviço — [desativar-servico.md](docs/servicos/desativar-servico.md)

## Peças

- [x] Cadastrar peça — [cadastrar-peca.md](docs/pecas/cadastrar-peca.md)
- [x] Consultar peças — [consultar-pecas.md](docs/pecas/consultar-pecas.md)
- [x] Atualizar peça — [atualizar-peca.md](docs/pecas/atualizar-peca.md)
- [x] Deletar peça — [deletar-peca.md](docs/pecas/deletar-peca.md)
- [x] Reservar peças para OS — [reservar-peca-para-os.md](docs/pecas/reservar-peca-para-os.md)
- [x] Processar peças para reserva e compra — [processar-pecas-para-reserva-e-compra.md](docs/pecas/processar-pecas-para-reserva-e-compra.md)
- [x] Solicitar compra de peças — [solicitar-compra-de-pecas.md](docs/pecas/solicitar-compra-de-pecas.md)
- [x] Registrar entrada de peças — [registrar-entrada-de-pecas.md](docs/pecas/registrar-entrada-de-pecas.md)
- [x] Retornar peça ao estoque — [retornar-peca-ao-estoque.md](docs/pecas/retornar-peca-ao-estoque.md)
- [x] Registrar consumo e saída de peças — [registrar-consumo-e-saida-de-pecas.md](docs/pecas/registrar-consumo-e-saida-de-pecas.md)
- [x] Cadastrar fornecedor — [cadastrar-fornecedor.md](docs/pecas/cadastrar-fornecedor.md)
- [x] Consultar fornecedores — [consultar-fornecedores.md](docs/pecas/consultar-fornecedores.md)
- [x] Atualizar fornecedor — [atualizar-fornecedor.md](docs/pecas/atualizar-fornecedor.md)
- [x] Desativar e reativar fornecedor — [desativar-fornecedor.md](docs/pecas/desativar-fornecedor.md)

## Insumos

- [x] Cadastrar insumo — [cadastrar-insumo.md](docs/insumos/cadastrar-insumo.md)
- [x] Consultar insumos — [consultar-insumos.md](docs/insumos/consultar-insumos.md)
- [x] Atualizar insumo — [atualizar-insumo.md](docs/insumos/atualizar-insumo.md)
- [x] Deletar insumo — [deletar-insumo.md](docs/insumos/deletar-insumo.md)
- [x] Reservar insumos para OS — [reservar-insumo-para-os.md](docs/insumos/reservar-insumo-para-os.md)
- [x] Processar insumos para reserva e compra — [processar-insumos-para-reserva-e-compra.md](docs/insumos/processar-insumos-para-reserva-e-compra.md)
- [x] Solicitar compra de insumos — [solicitar-compra-de-insumos.md](docs/insumos/solicitar-compra-de-insumos.md)
- [x] Registrar entrada de insumos — [registrar-entrada-de-insumos.md](docs/insumos/registrar-entrada-de-insumos.md)
- [x] Retornar insumo ao estoque — [retornar-insumo-ao-estoque.md](docs/insumos/retornar-insumo-ao-estoque.md)
- [x] Registrar consumo e saída de insumos — [registrar-consumo-e-saida-de-insumos.md](docs/insumos/registrar-consumo-e-saida-de-insumos.md)

---

## Observações

- A lista original tinha 52 tarefas. Ela mudou conforme os refinamentos chegaram: quatro linhas
  viraram a tarefa **Registrar peças e insumos necessários**, e entraram tarefas que não estavam
  previstas — retornar peça e insumo ao estoque, reservar insumo, os dois processamentos de reserva
  e compra, consultar insumos e cadastrar veículo já vinculado ao cliente. O total hoje é **55**:
  entraram o CRUD de fornecedor e a baixa de consumo nos dois contextos, e saiu a consulta de itens
  faltantes, que a equipe tirou do MVP.
- **Marcado quer dizer documentado, não implementado.** Uma tarefa marcada ainda está incompleta:
  consultar peças não tem seção de autenticação. Registrar peças e insumos necessários foi
  completada em 2026-08-22 com o refinamento técnico e o checklist.
- **Contextos de Cliente, Veículo, Serviços e Orçamento fechados.** Os pontos em aberto dos três foram
  decididos e aplicados em 2026-08-22; o registro está no `pontos-em-aberto.md` de cada diretório.
  As decisões arquiteturais foram todas fechadas em 22/08/2026, e o ponto que sobrava em Veículo
  saiu junto com a D-01.
- **A maior lacuna do projeto foi fechada.** `POST /estoque/saidas` agora existe, nos dois
  contextos: a baixa consome a reserva feita na aprovação, reduz o saldo físico e devolve ao saldo
  livre o que foi reservado e não usado. Sem ela, o saldo físico nunca diminuía.
- **O CRUD de fornecedor foi escrito**, no contexto de Peças, que é o dono do agregado de Compras.
- **Peças e Insumos fecharam seus pontos em aberto** em 2026-08-22, com decisões em bloco: código
  gerado pelo sistema, `nome` mantido, estoque inicial proibido, duplicidade por descrição
  normalizada, `ativo` fora do `PUT`, fornecedor obrigatório, compra acima da necessidade, e
  **nenhuma mensageria no projeto**. A entrada de estoque voltou a ser uma rota só, e as rotas de
  reserva direta foram aposentadas: reservar virou serviço de domínio chamado pelo processamento,
  que a aprovação do orçamento dispara.
- **Quatro tarefas de Orçamento deixaram de existir** em 2026-08-22: enviar orçamento, gerar
  orçamento complementar, aprovar complementar e recusar complementar. O complementar é criado no
  registro de itens da OS, e as decisões sobre ele usam as mesmas rotas de aprovar e recusar
  (DT-34).
- **Peças & Insumos virou dois contextos** em 2026-08-22: `docs/pecas/` e `docs/insumos/`. A
  divisão resolveu a numeração duplicada de tarefas e os IDs de requisito repetidos — os prefixos
  agora são `RF-PEC` e `RF-INS` —, renomeou `consultar-estoque.md` para `consultar-pecas.md` e
  duplicou os dois documentos que serviam aos dois tipos: a entrada de estoque e o retorno ao
  estoque na recusa. O total de tarefas subiu de 55 para 58 por causa dessas duplicações.
- **Selecionar próxima OS para execução e registrar problema adicional saíram do escopo** por
  decisão do time: a primeira é coberta por Consultar Fila de Atendimento, e a segunda por
  Registrar Problema Encontrado, que já abre o orçamento complementar.
- **O contexto de Ordem de Serviço foi renumerado** na ordem do fluxo: tarefas de 1 a 13, `RF-OS-01`
  a `RF-OS-124` e `RNF-OS-01` a `RNF-OS-65`, sem repetição.
- O retrato de cada contexto está no `00-resumo.md` do respectivo diretório, e as inconsistências
  no `pontos-em-aberto.md`. As divergências que atravessam contextos estão em
  [02-decisoes-arquiteturais.md](docs/02-decisoes-arquiteturais.md), decisões D-16 a D-25.
