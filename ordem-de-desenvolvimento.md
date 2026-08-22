---
documento: Ordem de Desenvolvimento — Caminho de Implementação
dono: José Lázaro
versao: 1.0
atualizado_em: 2026-08-22
status: em andamento
---

# Ordem de Desenvolvimento

Este documento define **em que ordem a equipe implementa as 55 tarefas documentadas**, e por quê.
Ele não substitui o refinamento: cada tarefa continua tendo o seu documento em `docs/`, com o
contrato da rota e o checklist de implementação. O que está aqui é a sequência — o que vem antes,
o que pode andar em paralelo, e o que trava se for feito fora de hora.

**Como ler.** O trabalho está dividido em **oito etapas**. Cada uma traz o objetivo, as tarefas
com o documento correspondente, o que ela destrava e o critério para considerá-la pronta. Dentro
de uma etapa, as frentes marcadas como paralelas podem ser distribuídas entre as cinco pessoas
sem conflito. Entre etapas há dependência real: pular a ordem gera retrabalho, não velocidade.

**Regra de ouro.** O bloco `N.3 Checklist de Implementação` do documento da tarefa **é a
definição de pronto**. Se um item do checklist não foi feito, a tarefa não está pronta, mesmo que
o endpoint responda `200`.

---

## Onde o projeto está hoje

| Item | Situação |
| --- | --- |
| Documentação | 55 de 55 tarefas refinadas, em sete contextos — ver [pontos-cobertos.md](pontos-cobertos.md) |
| Decisões arquiteturais | 25 de 25 fechadas — ver [docs/02-decisoes-arquiteturais.md](docs/02-decisoes-arquiteturais.md) |
| Catálogo de rotas | 58 rotas — ver [docs/03-endpoints.md](docs/03-endpoints.md) |
| Código | Esqueleto em Go 1.24: `cmd/oficina-mecanica/main.go` com `GET /health`, e os diretórios de camada criados com `estrutura.txt` |
| Dependências | Nenhuma: `go.sum` está vazio, o roteamento usa `net/http` puro |
| Banco | Postgres 16 já configurado no `docker-compose.yml` |
| Camadas existentes | `domain`, `application`, `infrastructure`, `presentation`, com pastas para cliente, veículo, serviço, ordem de serviço e segurança |

**Falta criar as pastas de contexto** de `orcamento`, `pecas`, `insumos` e `compras` nas quatro
camadas, e em `tests/integration`.

---

## Três coisas para resolver antes da primeira linha de código

Nenhuma delas é grande, mas todas travam alguém na Etapa 1 se ficarem para depois.

| # | Pendência | Por que trava | Onde decidir |
| --- | --- | --- | --- |
| 1 | `GET /estoque/categorias` não tem documento | A D-09 trocou `categoria` de texto livre para tabela referenciada por `categoriaId`. Sem essa consulta, nenhuma tela descobre o identificador que o cadastro de peça e de insumo exige. É a única rota do catálogo sem refinamento. | Escrever o documento em `docs/pecas/`, seguindo o [guia](docs/01-guia-de-documentacao.md) |
| 2 | A fila de atendimento é recurso próprio ou visão da OS | `GET /fila-atendimento` está na raiz, fora de `/ordens-servico`. Muda a rota e o pacote onde o handler vive. | Ponto em aberto 1 de [docs/03-endpoints.md](docs/03-endpoints.md) |
| 3 | A escolha do banco não está justificada | O `docker-compose.yml` já sobe Postgres 16, então a escolha está feita na prática. Falta o parágrafo de justificativa que o enunciado pede. | Seção 6 de [docs/00-visao-geral.md](docs/00-visao-geral.md) |

---

## Etapa 0 — Fundação técnica

**Ninguém trabalha em paralelo aqui.** É a etapa mais curta e a mais importante: tudo o que for
feito nela aparece em 58 rotas depois. Se cada pessoa resolver do seu jeito, o retrabalho é
garantido.

O que entra:

1. **Estrutura de pacotes e injeção de dependência.** Fechar como o `main.go` monta os casos de
   uso e os repositórios, e como cada contexto se registra no roteador. Criar as pastas que
   faltam para orçamento, peças, insumos e compras.
2. **Handler global de erros em Problem Details (D-03).** `type`, `title`, `status`, `detail` e a
   lista `erros` para falhas por campo ou por item. Escrito uma vez, usado por todos. Junto vem a
   regra de código da D-01: **o `422` não existe nesta API** — `400` é entrada inválida, `409` é
   conflito com o estado atual.
3. **Autenticação e autorização.** Validação do JWT e o corte por **escopo**, nunca por perfil. A
   lista oficial de escopos está na seção 8 do [guia](docs/01-guia-de-documentacao.md) e não se
   inventa escopo fora dela. Inclui o **token de escopo reduzido do cliente** (DT-30), emitido no
   envio do orçamento e válido só para aquela OS.
4. **Envelope de listagem paginada (D-21).** `data`, `pagina`, `tamanho`, `totalElementos`,
   `totalPaginas`. Recurso único devolve o objeto direto, sem envelope. Lista vazia é `200` com
   `"data": []`, nunca `404`.
5. **Controle otimista (D-24).** Um mecanismo só: header `If-Match` comparado com o campo
   `version`, `412` quando diverge e `428` quando o header não vem. Toda consulta de detalhe
   expõe `version`.
6. **Idempotência (D-02).** Tabela `chave_idempotencia` e o middleware que lê o header
   `Idempotency-Key`, guarda a resposta e devolve a original na repetição. Obrigatório em toda
   operação que movimenta saldo.
7. **Trilha de auditoria.** Onde outros projetos publicariam evento, este grava registro de
   auditoria — **não há mensageria nem eventos de domínio** (D-04). É o que alimenta o campo
   `eventos` da consulta da OS.
8. **Migration base e ferramental de teste.** Escolher a biblioteca de migration, subir o esquema
   inicial e deixar rodando a suíte com banco real em container, porque metade das regras deste
   projeto é de concorrência e não se testa com banco em memória.

**Destrava:** tudo.

**Pronto quando:** uma rota de exemplo passa ponta a ponta pelo middleware de autenticação, pelo
handler de erro e pelo envelope de listagem, com teste de integração subindo o Postgres do
compose.

---

## Etapa 1 — Cadastros independentes

Sete frentes que não dependem umas das outras. É aqui que a equipe de cinco pessoas anda mais
rápido, e é a etapa que dá o vocabulário para todo o resto.

| Frente | Tarefas | Documentos |
| --- | --- | --- |
| A | Categorias de item | a escrever, ver pendência 1 |
| B | Cliente: consultar, cadastrar, atualizar, deletar e reativar | [docs/cliente/](docs/cliente/) |
| C | Veículo: consultar, cadastrar, atualizar, deletar e reativar, mais cadastrar e vincular ao cliente | [docs/veiculo/](docs/veiculo/) |
| D | Serviços: cadastrar, consultar, atualizar, desativar e reativar | [docs/servicos/](docs/servicos/) |
| E | Fornecedor: cadastrar, consultar, atualizar, desativar e reativar | [docs/pecas/](docs/pecas/), tarefas 10 a 13 |
| F | Peça: cadastrar, consultar, atualizar, deletar | [docs/pecas/](docs/pecas/), tarefas 1 a 4 |
| G | Insumo: cadastrar, consultar, atualizar, deletar | [docs/insumos/](docs/insumos/), tarefas 1 a 4 |

**A frente A vem primeiro**, mesmo sendo pequena: peça e insumo referenciam `categoriaId` no
cadastro, e sem a tabela de categorias as frentes F e G param na primeira migration.

**O vínculo cliente-veículo pertence ao contexto de Cliente** (DT-10), e Veículo apenas
referencia. `POST /veiculos` está aposentada: o cadastro acontece sempre dentro do cliente.

Padrões que se repetem nas sete frentes, e que valem conferir uma vez para não errar sete vezes:

- **Exclusão lógica** com índice parcial: `UNIQUE (documento) WHERE ativo = true` em Cliente e
  Fornecedor, `UNIQUE (placa) WHERE ativo = true` em Veículo,
  `UNIQUE (categoria_id, descricao_normalizada) WHERE ativo = true` em Peça, com a unidade de
  medida somada à chave em Insumo.
- **Situação é o booleano `ativo`** com data e usuário da desativação, nunca um enum `status`
  (D-19). O verbo é `DELETE`, e a volta é `POST .../reativacao` (D-20).
- **Código gerado pelo sistema**, em sequência global de seis dígitos: `SER-000001`, `PEC-000001`,
  `INS-000001`.
- **Nenhum cadastro aceita saldo inicial.** Todo saldo entra por movimentação de entrada (DT-38).

**Destrava:** a abertura da OS, que precisa de cliente, veículo e catálogo de serviços.

**Pronto quando:** as sete frentes têm CRUD completo, com `If-Match` na atualização, `version` na
consulta e os índices parciais criados na migration.

---

## Etapa 2 — Abertura da OS e montagem do orçamento

A partir daqui a ordem é o fluxo do atendimento, e o paralelismo diminui. As tarefas 1 a 5 do
contexto de Ordem de Serviço são sequenciais entre si, porque cada uma depende do estado que a
anterior deixou.

| Ordem | Tarefa | Rota | Documento |
| --- | --- | --- | --- |
| 1 | Criar Ordem de Serviço | `POST /ordens-servico` | [criar-ordem-de-servico.md](docs/ordem-de-servico/criar-ordem-de-servico.md) |
| 2 | Registrar Problema Relatado | `POST /ordens-servico/{osId}/problema-relatado` | [registrar-problema-relatado.md](docs/ordem-de-servico/registrar-problema-relatado.md) |
| 3 | Registrar Problema Encontrado | `POST /ordens-servico/{osId}/problemas` | [registrar-problema-encontrado.md](docs/ordem-de-servico/registrar-problema-encontrado.md) |
| 4 | Registrar Serviços Necessários | `POST /ordens-servico/{osId}/servicos` | [registrar-servicos-necessarios.md](docs/ordem-de-servico/registrar-servicos-necessarios.md) |
| 5 | Registrar Peças e Insumos Necessários | `POST /ordens-servico/{osId}/pecas`, `POST /ordens-servico/{osId}/insumos` e `GET /ordens-servico/{osId}/orcamento` | [registrar-pecas-e-insumos-necessarios.md](docs/ordem-de-servico/registrar-pecas-e-insumos-necessarios.md) |

Em paralelo, uma segunda pessoa implementa **Calcular Orçamento**
([calcular-orcamento.md](docs/orcamento/calcular-orcamento.md)) e **Consultar Orçamento**
([consultar-orcamento.md](docs/orcamento/consultar-orcamento.md)), que dependem apenas do modelo
de orçamento, não do fluxo da OS.

Três armadilhas desta etapa:

- **Registrar item não reserva estoque.** O comprometimento acontece só na aprovação do orçamento.
  Quem implementar a tarefa 5 tentando reservar vai inverter o fluxo e travar na Etapa 4.
- **O complementar é orçamento separado** (D-17), com `tipoOrcamento` e `orcamentoOriginalId`, na
  mesma tabela. Não existe `orcamento_adicao`.
- **O cálculo é um passo explícito** (DT-32), acionado duas vezes: ao fim do diagnóstico e ao
  fechar cada complementar. Registrar item não recalcula sozinho.

**Destrava:** a decisão do cliente.

**Pronto quando:** é possível abrir uma OS, registrar problema, serviços, peças e insumos, e
calcular o orçamento com o valor total geral.

---

## Etapa 3 — Núcleo transacional do estoque

**A etapa mais difícil do projeto, e a que concentra o risco.** Tudo aqui mexe em saldo, e saldo
errado é o tipo de defeito que só aparece em produção, com duas pessoas usando ao mesmo tempo.
Recomendação: as duas pessoas mais confortáveis com transação e concorrência assumem esta etapa,
e ela não corre em paralelo com a Etapa 4.

| Ordem | Tarefa | Natureza | Documento |
| --- | --- | --- | --- |
| 1 | Reservar Peça para OS | serviço de domínio, sem endpoint | [reservar-peca-para-os.md](docs/pecas/reservar-peca-para-os.md) |
| 2 | Reservar Insumo para OS | serviço de domínio, sem endpoint | [reservar-insumo-para-os.md](docs/insumos/reservar-insumo-para-os.md) |
| 3 | Solicitar Compra de Peças e de Insumos | `POST /compras/pedidos` e `DELETE /compras/pedidos/{pedidoId}` | [solicitar-compra-de-pecas.md](docs/pecas/solicitar-compra-de-pecas.md) e [solicitar-compra-de-insumos.md](docs/insumos/solicitar-compra-de-insumos.md) |
| 4 | Processar Peças para Reserva e Compra | `POST /estoque/solicitacoes-compra-reserva` | [processar-pecas-para-reserva-e-compra.md](docs/pecas/processar-pecas-para-reserva-e-compra.md) |
| 5 | Processar Insumos para Reserva e Compra | `POST /estoque/solicitacoes-compra-reserva-insumos` | [processar-insumos-para-reserva-e-compra.md](docs/insumos/processar-insumos-para-reserva-e-compra.md) |
| 6 | Registrar Entrada | `POST /estoque/entradas`, rota única | [registrar-entrada-de-pecas.md](docs/pecas/registrar-entrada-de-pecas.md) e [registrar-entrada-de-insumos.md](docs/insumos/registrar-entrada-de-insumos.md) |
| 7 | Retornar ao Estoque | serviço de domínio, sem endpoint | [retornar-peca-ao-estoque.md](docs/pecas/retornar-peca-ao-estoque.md) e [retornar-insumo-ao-estoque.md](docs/insumos/retornar-insumo-ao-estoque.md) |

O que precisa estar certo, e é fácil errar:

- **Três saldos, não um.** `saldoFisico` é o que está na prateleira, `saldoReservado` é o que já
  tem dono, e `saldoDisponivel` é a diferença — é ele que pode ser prometido a um novo
  atendimento. **Insumo é reservado como peça** (D-15); a ideia contrária foi revogada.
- **Lock de linha com ordem fixa.** `SELECT ... FOR UPDATE` ordenado por `item_id`, sempre. É o
  que evita deadlock quando duas operações pegam os mesmos itens em ordens diferentes.
- **Idempotência obrigatória** em todas as sete tarefas, pelo mecanismo da Etapa 0.
- **A compra pode ser acima da necessidade** (DT-42): reserva-se para as OS o necessário e o
  excedente vira saldo livre. Comprar menos que a necessidade continua bloqueado.
- **O fornecedor é obrigatório na compra** (DT-41).
- **A entrada muda o status das OS vinculadas** ao pedido, por chamada direta na mesma transação,
  nunca por evento: a OS sem itens pendentes volta de `AGUARDANDO_RECURSOS` para
  `AGUARDANDO_EXECUCAO`.
- **Cada unidade de medida é um item independente** (DT-46). Comprar 1 L não aumenta o saldo do
  item cadastrado em mililitro.

**Destrava:** a aprovação do orçamento, que é quem chama o processamento.

**Pronto quando:** os testes de concorrência passam — duas reservas simultâneas do mesmo item não
deixam saldo negativo, e a repetição de uma entrada com a mesma `Idempotency-Key` não duplica
saldo.

---

## Etapa 4 — A decisão do cliente

Etapa curta em número de tarefas e central em consequência: **a aprovação é o ponto onde todos os
contextos se encontram**. Ela só pode ser implementada depois da Etapa 3, porque chama o
processamento de peças e de insumos dentro da própria transação.

| Ordem | Tarefa | Rota | Documento |
| --- | --- | --- | --- |
| 1 | Aprovar Orçamento | `POST /orcamentos/{orcamentoId}/aprovar` | [aprovar-orcamento.md](docs/orcamento/aprovar-orcamento.md) |
| 2 | Recusar Orçamento | `POST /orcamentos/{orcamentoId}/recusar` | [recusar-orcamento.md](docs/orcamento/recusar-orcamento.md) |
| 3 | Incluir OS na Fila de Atendimento | caso de uso interno, sem endpoint | [incluir-os-na-fila-de-atendimento.md](docs/ordem-de-servico/incluir-os-na-fila-de-atendimento.md) |

O fluxo da aprovação, que o documento detalha:

```
AprovarOrcamento
├── valida cliente, OS e orçamento
├── marca o orçamento como APROVADO
├── ProcessarPecas(os, itens do tipo PECA)      → reserva o disponível, abre pedido do faltante
├── ProcessarInsumos(os, itens do tipo INSUMO)  → reserva o disponível, abre pedido do faltante
├── define o status da OS pelo resultado:
│     nada pendente        → AGUARDANDO_EXECUCAO
│     algum item comprado  → AGUARDANDO_RECURSOS
└── confirma a transação
```

A recusa é o caminho inverso e igualmente transacional: devolve os itens ao estoque na mesma
transação, com o recorte por tipo — **recusa do principal cancela a OS e devolve tudo; recusa de
complementar descarta só aquele orçamento e devolve só os itens dele**.

O escopo das duas rotas é `orcamentos:decidir` (D-23), e quem chama é o cliente, com o token de
escopo reduzido da Etapa 0.

**Destrava:** a execução.

**Pronto quando:** aprovar um orçamento com item em falta deixa a OS em `AGUARDANDO_RECURSOS` com
pedido de compra aberto, e recusar devolve o saldo reservado ao disponível.

---

## Etapa 5 — Execução, baixa e entrega

O fluxo do mecânico, do momento em que ele pega a OS até o cliente levar o carro.

| Ordem | Tarefa | Rota | Documento |
| --- | --- | --- | --- |
| 1 | Consultar Fila de Atendimento | `GET /fila-atendimento` | [consultar-fila-de-atendimento.md](docs/ordem-de-servico/consultar-fila-de-atendimento.md) |
| 2 | Iniciar Execução | `POST /ordens-servico/{osId}/execucao/iniciar` | [iniciar-execucao.md](docs/ordem-de-servico/iniciar-execucao.md) |
| 3 | Registrar Consumo e Saída | `POST /estoque/saidas`, rota única | [registrar-consumo-e-saida-de-pecas.md](docs/pecas/registrar-consumo-e-saida-de-pecas.md) e [registrar-consumo-e-saida-de-insumos.md](docs/insumos/registrar-consumo-e-saida-de-insumos.md) |
| 4 | Finalizar Serviço | `POST /ordens-servico/{osId}/finalizar` | [finalizar-servico.md](docs/ordem-de-servico/finalizar-servico.md) |
| 5 | Registrar Entrega de Veículo | `POST /ordens-servico/{osId}/entrega` | [registrar-entrega-de-veiculo.md](docs/ordem-de-servico/registrar-entrega-de-veiculo.md) |

A baixa é o elo que fecha o ciclo do estoque e merece atenção:

- Ela **consome a reserva**, nunca o saldo livre. Item sem reserva ativa para aquela OS é recusado.
- **Acontece durante a execução e pode ocorrer mais de uma vez** na mesma OS: o mecânico registra
  o que usou conforme usa.
- Devolve ao saldo livre o que foi reservado e não usado, na mesma operação.
- **Informa o custo à OS**, que acumula e consolida o total do serviço (D-05).
- **A finalização bloqueia com `409`** enquanto houver reserva ativa sem baixa, devolvendo a lista
  dos itens pendentes. Sem isso o inventário nunca fecha.

Duas regras que costumam ser esquecidas: a **notificação de conclusão é por e-mail**, com o
resultado do envio gravado, e a falha do envio não desfaz a finalização; e a **entrega não
bloqueia por pagamento** (D-25), que ficou fora do MVP — ela apresenta e registra o valor final.

**Pronto quando:** uma OS percorre o ciclo inteiro, de `RECEBIDA` a `ENTREGUE`, com o estoque
batendo no fim.

---

## Etapa 6 — Leitura, indicadores e acompanhamento

Depende de tudo o que veio antes, porque devolve o histórico. Pode ser feita em paralelo com o
fechamento da Etapa 5 por quem terminar primeiro.

| Tarefa | Rota | Documento |
| --- | --- | --- |
| Consultar Ordem de Serviço | `GET /ordens-servico/{osId}` | [consultar-ordem-de-servico.md](docs/ordem-de-servico/consultar-ordem-de-servico.md) |
| Listar Ordens de Serviço | `GET /ordens-servico` | [listar-ordens-de-servico.md](docs/ordem-de-servico/listar-ordens-de-servico.md) |
| Monitorar Tempo Médio de Execução | `GET /ordens-servico/{osId}/tempo-execucao` e `GET /ordens-servico/tempos-execucao` | [monitorar-tempo-medio-de-execucao.md](docs/ordem-de-servico/monitorar-tempo-medio-de-execucao.md) |

**Listagem e detalhe são contratos distintos** (DT-56): a listagem é enxuta, para a tela de
acompanhamento, e o detalhe traz orçamentos, itens e a trilha de auditoria no campo `eventos`.
Esse campo é auditoria, não mensageria — é ele que sustenta a exigência de acompanhamento do
enunciado.

**Pronto quando:** o cliente consegue acompanhar a própria OS com o token de escopo reduzido, e a
oficina consegue listar por status, documento do cliente e placa.

---

## Etapa 7 — Fechamento da entrega

Não é burocracia de fim de projeto: metade destes itens é exigência direta do enunciado e vale
nota. Começar na última semana é o erro clássico.

| Item | Exigência | Onde |
| --- | --- | --- |
| Swagger ou equivalente | APIs RESTful documentadas | gerado a partir dos handlers, cobrindo as 58 rotas |
| Cobertura mínima de 80% | nos domínios críticos | estoque e orçamento são os críticos: é onde estão as invariantes de saldo |
| Dockerfile e docker-compose | ambiente completo com um comando | já existem; validar que sobem a aplicação e o Postgres juntos |
| README de execução local | explicativo | [README.md](README.md), hoje com quatro linhas |
| Relatório de vulnerabilidades | com o resultado do scan do código | a produzir |
| Repositório privado | com acesso ao usuário `soat-architecture` | a conferir |
| Vídeo de até 15 minutos | demonstrando o sistema | a gravar |
| Documento de entrega em PDF | grupo, participantes, usernames no Discord, links e relatório | a montar |

Sugestão prática: **Swagger e cobertura andam junto com as etapas**, não depois. Cada tarefa que
fecha já entra com o teste e a anotação da rota. A Etapa 7 vira consolidação, não mutirão.

---

## Como dividir cinco pessoas

O gargalo do projeto não é a quantidade de tarefas, é a **Etapa 3**: ela é difícil, concentra o
risco e não paraleliza bem. O desenho abaixo protege esse gargalo.

| Etapa | Distribuição sugerida |
| --- | --- |
| 0 | Todos juntos, decidindo em conjunto. É a única etapa em que reunião vale mais que código. |
| 1 | Cinco frentes em paralelo, uma por pessoa; quem terminar antes pega a sexta e a sétima. |
| 2 | Duas pessoas: uma no fluxo da OS, outra no cálculo e na consulta do orçamento. |
| 3 | Duas pessoas, as mais confortáveis com transação e concorrência. As outras três avançam em Swagger, testes e no que sobrou da Etapa 1. |
| 4 | As mesmas duas pessoas da Etapa 3, porque a aprovação chama o que elas escreveram. |
| 5 | Três pessoas: fila e execução, baixa, finalização e entrega. |
| 6 | Uma pessoa, em paralelo com o fim da Etapa 5. |
| 7 | Todos, com cada um fechando a documentação e os testes do que escreveu. |

---

## Dependências em uma imagem

```
Etapa 0  Fundação técnica
   │     erro, autenticação, escopo, paginação, If-Match, idempotência, auditoria, migration
   ▼
Etapa 1  Cadastros independentes ─── categorias → peça e insumo
   │     cliente · veículo · serviços · fornecedor · peça · insumo
   ▼
Etapa 2  Abertura da OS ──────────── em paralelo: cálculo e consulta do orçamento
   │     criar OS → problema relatado → problema encontrado → serviços → peças e insumos
   ▼
Etapa 3  Núcleo transacional do estoque
   │     reserva · compra · processamento · entrada · retorno
   ▼
Etapa 4  Decisão do cliente
   │     aprovar (chama o processamento) · recusar (devolve ao estoque) · fila
   ▼
Etapa 5  Execução, baixa e entrega
   │     fila → iniciar → saídas → finalizar → entregar
   ▼
Etapa 6  Leitura e indicadores ───── pode começar junto com o fim da Etapa 5
   │
   ▼
Etapa 7  Fechamento da entrega
```

---

## O que não vai ser implementado

Registrado para ninguém gastar tempo com isso, e para a decisão não voltar à mesa:

- **Pagamento** (D-25). A entrega registra o valor final e não bloqueia por confirmação.
- **Consulta de itens faltantes** (DT-50), nos dois contextos. A falta é apurada dentro do
  processamento disparado pela aprovação.
- **Recusa parcial e renegociação de orçamento** (DT-33). A decisão é sobre o orçamento inteiro.
- **Histórico de proprietários do veículo** (DT-16).
- **Anonimização e apagamento de dados pessoais** (DT-09). O `DELETE` é sempre exclusão lógica.
- **Selecionar a próxima OS da fila e registrar problema adicional** (DT-58).
- **Mensageria e eventos de domínio** (D-04). Integração entre contextos é consulta síncrona ou
  chamada direta na mesma transação.

---

## Manutenção deste documento

Quando uma etapa fechar, marque-a aqui e registre o que mudou de plano — a ordem real quase nunca
é idêntica à planejada, e o valor deste arquivo está em refletir a decisão do time, não a
intenção inicial. Decisão nova que apareça durante a implementação vai para
[docs/02-decisoes-arquiteturais.md](docs/02-decisoes-arquiteturais.md), e o ponto que ficar sem
resposta vai para o `pontos-em-aberto.md` do contexto afetado.
