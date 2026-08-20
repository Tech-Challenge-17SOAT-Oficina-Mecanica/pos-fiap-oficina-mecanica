---
documento: Visão Geral do Projeto
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-18
status: em construcao
---

# Sistema Integrado de Atendimento e Execução de Serviços

Documento de entrada do projeto. Explica o que estamos construindo, por quê, como o time se
organiza e onde encontrar cada coisa. É o primeiro arquivo que alguém novo — pessoa ou agente
de IA — deve ler.

Este documento cresce junto com o projeto: hoje ele descreve em profundidade apenas o contexto
de Peças & Insumos, que é o único já refinado. À medida que os outros contextos delimitados
forem documentados, as seções correspondentes deixam de ser esboço.

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
| Consulta de disponibilidade de peças e insumos | Peças & Insumos |
| Geração automática do orçamento a partir de serviços e peças | Orçamento |
| Envio do orçamento e aprovação pelo cliente | Orçamento |
| Reserva das peças da OS aprovada | Peças & Insumos |
| Execução do serviço e baixa das peças e insumos usados | Peças & Insumos / Ordem de Serviço |
| Problema adicional, orçamento complementar e nova aprovação | Orçamento / Ordem de Serviço |
| Reposição de estoque: itens faltantes, pedido de compra e recebimento | Peças & Insumos |
| Finalização e entrega do veículo | Ordem de Serviço |

**Status da OS**, conforme o enunciado: `Recebida` → `Em diagnóstico` → `Aguardando aprovação`
→ `Em execução` → `Finalizada` → `Entregue`. O board de Event Storming também prevê o
cancelamento da OS; o comportamento exato do cancelamento ainda será definido pelo contexto
de Ordem de Serviço.

---

## 4. Contextos delimitados

O sistema foi dividido em contextos delimitados, cada um com um dono responsável pela
documentação e pela manutenção das regras.

| Contexto | Agregados principais | Documento | Situação |
|---|---|---|---|
| Cliente | `Cliente` | [`cliente/`](cliente/) | 5 tarefas refinadas |
| Veículo | `Veículo` | [`veiculo/`](veiculo/) | 3 tarefas refinadas |
| Ordem de Serviço | `Ordem de Serviço` | [`ordem-de-servico/`](ordem-de-servico/) | 6 de 18 tarefas refinadas |
| Orçamento | `Orçamento`, `Item de Orçamento` | [`orcamento/`](orcamento/) | 5 de 8 tarefas refinadas |
| Serviços | `Serviço` | `servicos/` | a documentar |
| Peças & Insumos | `Item de Estoque`, `Reserva`, `Movimentação`, `Pedido de Compra` | [`pecas-e-insumos/`](pecas-e-insumos/) | 13 tarefas refinadas |

A divisão nasceu do Event Storming feito pelo grupo (board exportado em
[`files/Designs – Software Architecture _ FIAP (1).pdf`](files/)). O board tem inconsistências
conhecidas: os documentos de contexto são o lugar de corrigir e explicar as divergências, não
de copiar o board sem crítica.

### Peças & Insumos (único contexto já refinado)

Cobre o catálogo de peças e insumos, os saldos, a reserva para ordens de serviço, a baixa no
consumo e o ciclo de compras. Nove requisitos documentados:

1. Consultar Estoque
2. Atualizar Peça
3. Atualizar Insumo
4. Registrar Entrada de Estoque
5. Reservar Peça para OS
6. Registrar Consumo e Saída
7. Consultar Peças Faltantes
8. Solicitar Compra de Peças
9. Solicitar Compra de Insumos

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
| Insumo | Material de consumo diluído no custo do serviço; não é reservado, tem baixa direta |
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
  `compras:escrever`), e não por perfil. Os perfis existentes são `MECANICO`, `GESTOR` e
  `SERVICO`; **não existe perfil `ESTOQUISTA`**.
- Envelope de listagem paginada: `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`.
  Lista vazia é `200` com `"data": []`, nunca `404`.
- Escrita concorrente em cadastro usa **lock otimista**: header `If-Match` comparado com o
  campo `version`.
- Operação que movimenta saldo é **transacional** (tudo ou nada) e protegida por
  `SELECT ... FOR UPDATE` com ordenação fixa por `item_id`, para evitar deadlock.
- Compras **não** é contexto separado: `pedido_compra` pertence a Peças & Insumos.
- Insumo **não** é reservado; a baixa acontece direto na execução do serviço.

### Ainda em aberto

Quinze decisões estão pendentes de discussão em equipe, três delas bloqueantes: o padrão de
códigos HTTP (`409` x `422`), o mecanismo de idempotência e o formato do corpo de erro. Todas
estão em [`02-decisoes-arquiteturais.md`](02-decisoes-arquiteturais.md), com opções e recomendação
para cada uma.

---

## 7. Documentação do projeto

| Arquivo | Para que serve |
|---|---|
| [`00-visao-geral.md`](00-visao-geral.md) | Este documento: o que é o projeto e onde está cada coisa |
| [`01-guia-de-documentacao.md`](01-guia-de-documentacao.md) | Como escrever um documento de contexto: nome do arquivo, estrutura, convenções |
| [`pecas-e-insumos/`](pecas-e-insumos/) | Contexto de Peças & Insumos: um arquivo por tarefa, mais os pontos em aberto do contexto |
| [`02-decisoes-arquiteturais.md`](02-decisoes-arquiteturais.md) | Decisões pendentes, com opções e recomendação, para discussão em equipe |
| [`03-endpoints.md`](03-endpoints.md) | Catálogo de todas as rotas da API, com método, caminho, escopo e documento de origem |
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

1. Fechar as decisões bloqueantes de [`02-decisoes-arquiteturais.md`](02-decisoes-arquiteturais.md)
   (D-01, D-02 e D-03) — elas mudam o contrato de todos os endpoints.
2. Documentar os contextos restantes seguindo o
   [guia](01-guia-de-documentacao.md): Cliente, Veículo, Ordem de Serviço, Orçamento e Serviços.
3. Definir os donos de cada contexto e preencher a coluna de responsáveis.
4. Consolidar a linguagem ubíqua num glossário único, quando os cinco contextos estiverem escritos.
5. Escolher e justificar o banco de dados.
6. Iniciar a implementação pelos requisitos com checklist pronto.
