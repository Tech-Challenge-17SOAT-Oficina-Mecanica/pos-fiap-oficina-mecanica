---
documento: Visão Geral do Projeto
dono: José Lázaro
versao: 0.6
atualizado_em: 2026-08-22
status: em construcao
---

# Sistema Integrado de Atendimento e Execução de Serviços

Documento de entrada do projeto. Explica o que estamos construindo, por quê, como o time se
organiza e onde encontrar cada coisa. É o primeiro arquivo que alguém novo — pessoa ou agente
de IA — deve ler.

Este documento cresce junto com o projeto. Os sete contextos delimitados já estão refinados, cada
um no seu diretório, com resumo, uma tarefa por arquivo e a lista de pontos em aberto. O que ainda
muda aqui são as decisões de padrão de API que continuam na pauta.

---

## 1. O que é

MVP do back-end de um sistema de gestão para uma **oficina mecânica de médio porte**,
especializada em manutenção de veículos. O sistema cobre o ciclo completo do atendimento:
identificação do cliente, cadastro do veículo, abertura da ordem de serviço, diagnóstico,
orçamento, aprovação pelo cliente, execução, controle de peças e insumos, e entrega.

É o **Tech Challenge da Fase 1** da pós-graduação em Arquitetura de Software da FIAP (15SOAT),
desenvolvido em grupo de cinco integrantes, aplicando Domain-Driven Design e boas práticas de
qualidade e segurança.

---

## 2. O problema que estamos resolvendo

Hoje o atendimento, o diagnóstico, a execução e a entrega são controlados com anotações
manuais e planilhas. O levantamento dos fluxos atuais (em [`files/Cliente (1).pdf`](files/)) mostra
a rotina real da oficina:

- a fila de atendimento é por ordem de chegada, consultada em planilha;
- o cadastro do cliente é preenchido em formulário de papel e depois digitado;
- o orçamento é calculado à mão pelo mecânico e escrito em folha;
- a aprovação do cliente acontece por conversa, sem registro;
- quando falta peça, o mecânico anota o pedido de compra na planilha e avisa o cliente do atraso;
- o acompanhamento do serviço depende do cliente ligar e alguém interromper o trabalho para conferir;
- no encerramento, a ordem de serviço é preenchida à mão e arquivada.

Disso vêm os problemas que o sistema precisa resolver:

- erros na priorização dos atendimentos;
- falhas no controle de peças e insumos;
- dificuldade em acompanhar o status dos serviços;
- perda de histórico de clientes e veículos;
- ineficiência no fluxo de orçamentos e autorizações.

O objetivo do sistema é permitir que o cliente acompanhe o andamento em tempo real e autorize
reparos adicionais, e que a oficina tenha uma gestão interna organizada e segura.

---

## 3. O fluxo do negócio, ponta a ponta

O caminho que um atendimento percorre, e quem responde por cada etapa:

| Etapa | Contexto responsável |
|---|---|
| Identificação do cliente por CPF/CNPJ | Cliente |
| Cadastro do veículo (placa, marca, modelo, ano) e vínculo com o cliente | Veículo |
| Abertura da OS e registro dos serviços solicitados | Ordem de Serviço |
| Diagnóstico e identificação de problemas | Ordem de Serviço |
| Consulta de disponibilidade de peças e insumos | Peças / Insumos |
| Geração automática do orçamento a partir de serviços e peças | Orçamento |
| Envio do orçamento e aprovação pelo cliente | Orçamento |
| Reserva das peças e insumos da OS aprovada | Peças / Insumos |
| Execução do serviço e baixa das peças e insumos usados | Peças / Insumos / Ordem de Serviço |
| Problema adicional, orçamento complementar e nova aprovação | Orçamento / Ordem de Serviço |
| Reposição de estoque: pedido de compra e recebimento | Peças / Insumos |
| Finalização e entrega do veículo | Ordem de Serviço |

**Status da OS.** O enunciado lista seis — `Recebida`, `Em diagnóstico`, `Aguardando aprovação`,
`Em execução`, `Finalizada` e `Entregue` — e o refinamento chegou a nove, acrescentando
`AGUARDANDO_RECURSOS` (parada esperando peça comprada), `AGUARDANDO_EXECUCAO` (orçamento
aprovado, serviço ainda não iniciado) e `CANCELADA` (destino da recusa do orçamento). A máquina
de estados completa, com as transições e quem dispara cada uma, está no
[resumo de Ordem de Serviço](ordem-de-servico/00-resumo.md).

---

## 4. Contextos delimitados

O sistema foi dividido em contextos delimitados, cada um com um dono responsável pela
documentação e pela manutenção das regras.

| Contexto | Agregados principais | Documento | Situação |
|---|---|---|---|
| Cliente | `Cliente` | [`cliente/`](cliente/) | 5 tarefas — ver [resumo](cliente/00-resumo.md) |
| Veículo | `Veículo` | [`veiculo/`](veiculo/) | 5 tarefas — ver [resumo](veiculo/00-resumo.md) |
| Ordem de Serviço | `Ordem de Serviço`, `Problema`, `Evento da OS` | [`ordem-de-servico/`](ordem-de-servico/) | 13 tarefas — ver [resumo](ordem-de-servico/00-resumo.md) |
| Orçamento | `Orçamento`, `Item de Orçamento` | [`orcamento/`](orcamento/) | 4 tarefas — ver [resumo](orcamento/00-resumo.md) |
| Serviços | `Serviço` | [`servicos/`](servicos/) | 4 tarefas — ver [resumo](servicos/00-resumo.md) |
| Peças | `Item de Estoque` (`PECA`), `Reserva`, `Movimentação`, `Pedido de Compra`, `Fornecedor` | [`pecas/`](pecas/) | 15 tarefas — ver [resumo](pecas/00-resumo.md) |
| Insumos | `Item de Estoque` (`INSUMO`), `Reserva`, `Movimentação`, `Pedido de Compra` | [`insumos/`](insumos/) | 10 tarefas — ver [resumo](insumos/00-resumo.md) |

A divisão nasceu do Event Storming feito pelo grupo (board exportado em
[`files/Designs – Software Architecture _ FIAP (1).pdf`](files/)). O board tem inconsistências
conhecidas: os documentos de contexto são o lugar de corrigir e explicar as divergências, não
de copiar o board sem crítica.

### Peças e Insumos

Eram um contexto só e foram **divididos em dois**: [`pecas/`](pecas/) e [`insumos/`](insumos/).
Os dois seguem o mesmo desenho — catálogo, saldos, reserva para ordens de serviço, baixa no
consumo e ciclo de compras — e cada um tem as suas rotas, `/estoque/pecas` e `/estoque/insumos`.
A diferença é de negócio: a peça é cobrada do cliente item a item, pelo `precoVenda`; o insumo
entra no custo do serviço, pelo `custoUnitario`.

O agregado de **Compras** — `Pedido de Compra` e `Fornecedor` — não é contexto separado: ele
pertence a Peças, e Insumos apenas o referencia. Por isso `/fornecedores`, `/compras/pedidos`,
`/estoque/entradas` e `/estoque/saidas` são rotas compartilhadas pelos dois.

O conceito central é a separação de saldos: **saldo físico** (o que está na prateleira),
**saldo reservado** (o que já tem dono, uma OS aprovada) e **saldo disponível** (a diferença
entre os dois, que é o que pode ser prometido a um novo atendimento).

---

## 5. Linguagem ubíqua (parcial)

Termos já estabelecidos. Cada contexto amplia esta lista no seu próprio documento.

| Termo | Significado |
|---|---|
| Ordem de Serviço (OS) | Registro do atendimento de um veículo, do recebimento à entrega |
| Orçamento | Valor proposto ao cliente a partir dos serviços e peças da OS, sujeito a aprovação |
| Peça | Item cobrado do cliente e aplicado no veículo; é reservável |
| Insumo | Material de consumo diluído no custo do serviço; é reservado como a peça, com baixa na execução |
| Saldo físico | Quantidade existente no estoque |
| Saldo reservado | Quantidade já comprometida com OS aprovadas |
| Saldo disponível | Saldo físico menos saldo reservado |
| Estoque mínimo | Ponto de reposição do item |
| Reserva | Vínculo entre uma quantidade de peça e uma OS específica |
| Movimentação de estoque | Registro imutável de entrada, saída, reserva ou liberação |
| Pedido de compra | Solicitação formal de reposição a um fornecedor |
| Entrada | Recebimento que aumenta o saldo físico |
| Saída | Baixa do que foi efetivamente utilizado no serviço |

---

## 6. Decisões técnicas

### Definidas pelo enunciado

- Back-end **monolítico**, em arquitetura em camadas (é um MVP).
- Banco de dados de escolha livre, com justificativa registrada.
- APIs **RESTful documentadas** via Swagger ou equivalente.
- **Dockerfile** e **docker-compose.yml** para subir o ambiente completo.
- Testes automatizados com **cobertura mínima de 80%** nos domínios críticos.
- **JWT** nas APIs administrativas; validação de dados sensíveis (CPF/CNPJ, placa).
- README explicativo para execução local e repositório privado com acesso ao usuário `soat-architecture`.

### Acordadas pelo time até agora

- Rotas sem prefixo de versão: o recurso começa na raiz, por exemplo `/clientes`.
- Autorização por **escopo** no token (`estoque:ler`, `estoque:escrever`, `estoque:movimentar`,
  `compras:escrever`), e não por perfil. Os perfis existentes são `MECANICO`, `CLIENTE` e
  `SERVICO`; **não existem os perfis `ESTOQUISTA` nem `GESTOR`**.
- Envelope de listagem paginada: `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`.
  Lista vazia é `200` com `"data": []`, nunca `404`.
- Escrita concorrente em cadastro usa **lock otimista**: header `If-Match` comparado com o
  campo `version`.
- Operação que movimenta saldo é **transacional** (tudo ou nada) e protegida por
  `SELECT ... FOR UPDATE` com ordenação fixa por `item_id`, para evitar deadlock.
- Compras **não** é contexto separado: `pedido_compra` e `fornecedor` pertencem a Peças, e
  Insumos os referencia.
- **Insumo é reservado como a peça.** A reserva dos dois nasce da aprovação do orçamento e é
  consumida pela baixa, em `POST /estoque/saidas`, durante a execução do serviço.
- **Sem mensageria e sem eventos de domínio.** A conversa entre contextos é consulta síncrona ou
  chamada direta dentro da mesma transação; o histórico da OS é trilha de auditoria.

### Pauta fechada

As vinte e cinco decisões de [`02-decisoes-arquiteturais.md`](02-decisoes-arquiteturais.md) foram
fechadas em 22/08/2026, com as opções descartadas preservadas para explicar o porquê de cada
escolha. As últimas a cair mudaram contrato de API em todos os contextos: o `422` saiu (`400` para
entrada inválida, `409` para conflito de estado), o corpo de erro passou a ser Problem Details
vindo de um handler global, `categoria` virou tabela referenciada por identificador, e as
listagens de peça e insumo passaram a devolver `version`.

---

## 7. Documentação do projeto

| Arquivo | Para que serve |
|---|---|
| [`00-visao-geral.md`](00-visao-geral.md) | Este documento: o que é o projeto e onde está cada coisa |
| [`01-guia-de-documentacao.md`](01-guia-de-documentacao.md) | Como escrever um documento de contexto: nome do arquivo, estrutura, convenções |
| [`cliente/`](cliente/) | Contexto de Cliente: refinamentos separados por tarefa |
| [`veiculo/`](veiculo/) | Contexto de Veículo: refinamentos separados por tarefa |
| [`ordem-de-servico/`](ordem-de-servico/) | Contexto de Ordem de Serviço: refinamentos separados por tarefa |
| [`orcamento/`](orcamento/) | Contexto de Orçamento: refinamentos separados por tarefa |
| [`pecas/`](pecas/) | Contexto de Peças: um arquivo por tarefa, mais o resumo e os pontos em aberto |
| [`insumos/`](insumos/) | Contexto de Insumos: um arquivo por tarefa, mais o resumo e os pontos em aberto |
| [`02-decisoes-arquiteturais.md`](02-decisoes-arquiteturais.md) | Decisões pendentes, com opções e recomendação, para discussão em equipe |
| [`03-endpoints.md`](03-endpoints.md) | Catálogo de todas as rotas da API, com método, caminho, escopo e documento de origem |
| `<contexto>/00-resumo.md` | O que o contexto cobre: tarefas, rotas, tipos e convenções em vigor |
| `<contexto>/pontos-em-aberto.md` | Decisões pendentes e inconsistências a corrigir naquele contexto |
| [`files/`](files/) | Material de apoio: enunciado do Tech Challenge, board de Event Storming, levantamento dos fluxos atuais |

Cada contexto delimitado tem um arquivo próprio, nomeado como
`<nome-do-contexto>-cd.md` — a regra completa está na seção 1 do guia.

Cada requisito, dentro do documento do seu contexto, é descrito em três blocos: **Refinamento
de Produto** (o que o usuário precisa e por quê), **Refinamento Técnico** (contrato,
processamento, testes) e **Checklist de Implementação** (o passo a passo até o merge).

---

## 8. Entregáveis da Fase 1

- Vídeo de até 15 minutos demonstrando o sistema.
- Documentação DDD: Event Storming completo dos fluxos de OS e de peças e insumos, diagramas e
  linguagem ubíqua aplicada.
- Código-fonte no repositório privado, com APIs, Dockerfile e docker-compose configurados e
  README completo.
- Relatório com análise de vulnerabilidades, incluindo o resultado do scan do código.
- Documento de entrega em PDF com nome do grupo, participantes e usernames no Discord, link da
  documentação, link do repositório e o relatório de vulnerabilidades.

---

## 9. Próximos passos

1. Documentar `GET /estoque/categorias`, a única rota do catálogo ainda sem documento — ela nasceu
   da D-09 e sem ela ninguém descobre o `categoriaId` que o cadastro exige.
2. Decidir se a fila de atendimento é recurso próprio ou visão da OS, último ponto em aberto do
   catálogo de rotas.
3. Definir os donos de cada contexto e preencher a coluna de responsáveis.
4. Consolidar a linguagem ubíqua num glossário único, agora que os sete contextos estão escritos.
5. Escolher e justificar o banco de dados.
6. Iniciar a implementação pelos requisitos com checklist pronto.
