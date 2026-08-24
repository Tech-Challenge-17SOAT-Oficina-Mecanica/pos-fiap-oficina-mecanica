# Guia de Documentação — Contextos Delimitados

> **Leia este guia antes de criar ou alterar qualquer documento de contexto.**
> Ele define o nome do arquivo, a estrutura das seções e as convenções de escrita que
> todos os integrantes seguem. O objetivo é simples: qualquer pessoa (ou agente de IA)
> deve conseguir abrir qualquer documento do projeto e encontrar a mesma coisa no mesmo lugar.
>
> **Documento de referência:** [`pecas/consultar-pecas.md`](pecas/consultar-pecas.md) — é o exemplo
> real e aprovado do padrão. Em caso de dúvida entre o guia e o exemplo, siga o exemplo e
> abra um PR corrigindo o guia.

---

## 1. Nome e localização dos arquivos — regra obrigatória

**Cada contexto delimitado é uma pasta em `docs/`, e cada tarefa é um arquivo dentro dela.**
Não existe mais contexto em arquivo único: mesmo um contexto pequeno vira pasta, para que o
diretório cresça sem precisar ser dividido depois.

```text
docs/
└── ordem-de-servico/
    ├── 00-resumo.md                       ← o que o contexto cobre, rotas e tipos
    ├── criar-ordem-de-servico.md          ← uma tarefa por arquivo
    ├── registrar-problema-relatado.md
    ├── registrar-servicos-necessarios.md
    └── pontos-em-aberto.md                ← inconsistências e decisões pendentes
```

Regras de formação, tanto para a pasta quanto para os arquivos:

- tudo em **minúsculas**;
- **sem acentos** e sem cedilha (`servicos`, não `serviços`);
- palavras separadas por **hífen**;
- o arquivo recebe o nome da tarefa, começando pelo verbo: `cadastrar-peca.md`,
  `consultar-fila-de-atendimento.md`;
- sem prefixo numérico, sem data, sem nome do autor — as duas exceções são `00-resumo.md` e
  `pontos-em-aberto.md`, que têm nome fixo.

| Contexto delimitado | Pasta |
|---|---|
| Cliente | `cliente/` |
| Veículo | `veiculo/` |
| Ordem de Serviço | `ordem-de-servico/` |
| Orçamento | `orcamento/` |
| Serviços | `servicos/` |
| Peças | `pecas/` |
| Insumos | `insumos/` |

Exemplos **errados**: `Pecas.md`, `estoque_cd.md`, `04-pecas.md`, `peças-e-insumos.md`,
`lazaro-estoque.md`, `pecas-e-insumos-cd.md` (o sufixo `-cd` foi aposentado).

**Os dois arquivos de nome fixo**

- `00-resumo.md` — o retrato do contexto: as tarefas com rota, escopo e link, os tipos e enums,
  as convenções em vigor e o que o contexto **não** faz. É a porta de entrada de quem chega.
- `pontos-em-aberto.md` — tudo que ainda não foi decidido ou que está inconsistente entre
  documentos, com o motivo de ser um problema e uma sugestão de correção. Quando o time decide,
  a linha sai da tabela e a decisão é registrada no documento da tarefa ou em
  [02-decisoes-arquiteturais.md](02-decisoes-arquiteturais.md).

As pastas de contexto ficam diretamente em `docs/`. Material de apoio (PDFs, exports do Miro e
imagens) vai em `docs/files/`.

---

## 2. Frontmatter obrigatório

Todo documento começa com este bloco, sem exceção:

```yaml
---
documento: Refinamento de Requisitos — Contexto de <Nome do Contexto>
dono: <Nome do integrante>
versao: 0.1
atualizado_em: AAAA-MM-DD
status: rascunho | em revisao | aprovado
---
```

- `dono` — quem escreveu e quem revisa qualquer alteração futura naquele arquivo.
- `versao` — sobe `0.1` a cada rodada de mudança relevante; vira `1.0` quando aprovado.
- `atualizado_em` — data absoluta, formato `AAAA-MM-DD`. Nunca "hoje", "semana passada".
- `status` — `rascunho` enquanto escreve, `em revisao` no PR, `aprovado` após o merge.

---

## 3. Estrutura do documento

Depois do frontmatter, o documento segue sempre esta ordem:

```
# Refinamento de Requisitos                  ← título fixo
<parágrafo de introdução, explicando os 3 blocos>

## 1 · <Nome do Requisito>
### 1.1 Refinamento de Produto
### 1.2 Refinamento Técnico
### 1.3 Checklist de Implementação

## 2 · <Nome do Próximo Requisito>
### 2.1 Refinamento de Produto
### 2.2 Refinamento Técnico
### 2.3 Checklist de Implementação

## Pontos em aberto                          ← sempre a última seção
```

Regras para contexto em arquivo único:

- **Um documento por contexto delimitado**, com **todos** os requisitos daquele contexto dentro.
- Cada requisito é uma seção `##` numerada, com o separador `·` (`1 · Consultar Estoque`).
- Cada requisito tem **os três blocos**, sempre nessa ordem e sempre com esses nomes.
- Requisito sem os três blocos não entra no documento — entra em **Pontos em aberto**.
- Separe cada requisito com uma linha `---`.

Regras para contexto em pasta dedicada:

- Cada tarefa fica em um arquivo e mantém os três blocos obrigatórios.
- Os números e IDs dos requisitos continuam únicos dentro do contexto.
- Pontos pendentes não ficam repetidos nos arquivos das tarefas; são centralizados em
  `pontos-em-aberto.md`.
- A tabela de pontos em aberto contém `#`, `Ponto`, `Arquivo relacionado` e `Responsável`.

---

## 4. Bloco 1 — Refinamento de Produto

Visão de negócio. Nada de tabela de banco, nome de classe ou verbo HTTP aqui.
Campos, nesta ordem:

| Campo | O que escrever |
|---|---|
| **Persona** | Quem executa a ação (Mecânico, Cliente). Um por linha. |
| **Objetivo** | Uma frase: o que a pessoa quer conseguir e para quê. |
| **Problema** | A dor real de hoje, com a consequência concreta (retrabalho, atraso, perda de histórico). É o que justifica o requisito existir. |
| **Pré-condições** | Lista do que precisa ser verdade antes de começar. |
| **Requisitos Funcionais** | Tabela com ID + requisito. O que o sistema faz. |
| **Requisitos Não Funcionais** | Tabela com ID + requisito. Como o sistema se comporta (segurança, performance, paginação, ausência de efeito colateral). |
| **Fluxo Principal** | Lista numerada, do caminho feliz, do início ao fim. |
| **Fluxos Alternativos / Exceções** | Tabela com ID (`A1`, `A2`...), situação e comportamento do sistema. |
| **Saída** | O que a pessoa recebe ao final. |
| **Pós-condições** | O estado do sistema depois da operação. Diga explicitamente o que **não** mudou. |

Escreva sempre em termos do negócio: *"o mecânico consulta a disponibilidade"*, não
*"o endpoint retorna a lista"*.

---

## 5. Bloco 2 — Refinamento Técnico

Como o sistema entrega o que o bloco 1 pediu. Campos, nesta ordem:

| Campo | O que escrever |
|---|---|
| **Endpoint** | Método e rota em bloco `http`, sem prefixo de versão. Um requisito pode expor mais de um endpoint (por exemplo, criar e liberar) — liste todos no mesmo bloco e explique em uma linha o papel de cada um. |
| **Autenticação / Autorização** | Token exigido, perfis permitidos e escopo. |
| **Entrada** | Tabela: parâmetro, tipo, descrição. Diga o que é obrigatório e os defaults. |
| **Validações** | Lista objetiva de limites e formatos. Se não houver regra de negócio, diga isso explicitamente. |
| **Processamento** | Lista numerada do passo a passo interno, da validação até a montagem da resposta. |
| **Persistência** | O que é **consultado**, o que é **alterado** e observações (somente leitura, transação, réplica). |
| **Saída da API** | Exemplo real em bloco `json`, com valores plausíveis e coerentes entre si. |
| **Códigos HTTP / Erros** | Tabela: código e situação. Cubra 4xx de autenticação e autorização. |
| **Dependências** | Repositórios, middlewares e outros contextos de que este fluxo depende. |
| **Testes** | Subdividido em *Unitários* (regra de domínio isolada) e *Integração* (contrato da API ponta a ponta). |

O exemplo de JSON precisa ser **consistente**: se `saldoFisico` é 6 e `saldoReservado` é 4,
então `saldoDisponivel` é 2 e `disponivel` é `true`. Revisor confere isso.

**Decisão de projeto.** Quando o contrato embutir uma escolha de arquitetura — rota
compartilhada entre dois requisitos, reaproveitamento de um recurso existente, read model
materializado — registre-a em uma citação (`>`) logo abaixo do endpoint, dizendo qual foi a
alternativa considerada e o custo dela. Decisão sem alternativa registrada vira discussão de
novo três semanas depois.

---

## 6. Bloco 3 — Checklist de Implementação

Tarefas verificáveis, em checkbox markdown (`- [ ]`), agrupadas **sempre** nestes grupos e
nesta ordem (pule o grupo que não se aplicar):

```
**Domínio**              → regra de negócio pura, cálculo, invariante
**Caso de uso**          → orquestração do fluxo
**Repositório**          → acesso a dados
**Integrações**          → consulta a outro contexto ou chamada direta a outro caso de uso
**Handler HTTP**         → rota, envelope de resposta
**Validações**           → limites de entrada
**Concorrência**         → controle otimista (`If-Match` / `version`), quando houver escrita
**Transação e idempotência** → transação única e `Idempotency-Key`, quando a operação movimenta saldo
**Auditoria**            → registro em trilha de auditoria, quando a operação precisa de rastro
**Testes unitários**     → espelham os testes do bloco 2
**Testes de integração** → espelham os testes do bloco 2
**Testes de concorrência**   → execuções simultâneas, quando duas pessoas disputam o mesmo registro
**Documentação**         → Swagger/OpenAPI
**Review**               → Code Review aprovado
```

Cada item é uma ação concreta que dá para marcar como feita. `- [ ] Implementar
ConsultarEstoque com filtros e paginação` serve; `- [ ] Fazer o estoque` não serve.

---

## 7. Seção final — Pontos em aberto

Última seção do documento, em tabela:

| # | Ponto | Responsável |
|---|---|---|
| 1 | O que ainda não foi decidido, com contexto suficiente para outra pessoa decidir | Nome ou `—` |

**Não invente regra para preencher lacuna.** Se o time não decidiu, o lugar é aqui.
Documento sem nenhum ponto em aberto é possível, mas raro num primeiro rascunho.

---

## 8. Convenções de escrita

**IDs.** Todo requisito e exceção recebe ID, para ser citado em código, teste e PR:

| Tipo | Formato | Exemplo |
|---|---|---|
| Requisito funcional | `RF-<CTX>-NN` | `RF-PEC-01` |
| Requisito não funcional | `RNF-<CTX>-NN` | `RNF-PEC-03` |
| Fluxo alternativo | `A<N>` | `A2` |

Sigla de contexto (`<CTX>`) — use sempre a mesma:

| Contexto | Sigla |
|---|---|
| Cliente | `CLI` |
| Veículo | `VEI` |
| Ordem de Serviço | `OS` |
| Orçamento | `ORC` |
| Peças | `PEC` |
| Insumos | `INS` |
| Serviços | `SRV` |

**Linguagem.** Prosa em português. Nomes técnicos (campo, classe, enum, rota, escopo)
em `crase`, exatamente como aparecem no código. Termo de negócio sempre igual em todos os
documentos: `Ordem de Serviço`, nunca "chamado", "ticket" ou "pedido".

**Padrões de API compartilhados** — valem para todos os contextos:

- Rotas sem prefixo de versão: o recurso começa na raiz, por exemplo `/clientes`
- Toda rota nova entra no [`03-endpoints.md`](03-endpoints.md) no mesmo PR do documento da tarefa
- Autenticação: `Bearer <JWT>` nas APIs administrativas
- Escopo no formato `recurso:acao`, escolhido da **lista oficial** logo abaixo — escopo novo só entra depois de acrescentado a ela
- **Recurso único devolve o objeto direto, sem envelope.** O envelope é só de listagem
- Envelope paginado: `data`, `pagina`, `tamanho`, `totalElementos`, `totalPaginas`
- **Limites da paginação (D-26)**, iguais em todos os contextos: `pagina` inicia em zero, padrão `0`; `tamanho` tem padrão `20` e teto `50`. Fora da faixa é `400`
- Lista vazia é `200` com `"data": []`, **nunca** `404`
- **O `422` não existe nesta API (D-01).** `400` é entrada inválida — formato, campo obrigatório ausente, item de tipo errado — e `409` é qualquer conflito com o estado atual: duplicidade, saldo insuficiente, status incompatível, registro inativo
- `400` entrada inválida · `401` token ausente/expirado · `403` sem escopo · `404` recurso inexistente
- Operações de escrita: `201` recurso criado · `204` operação sem corpo de resposta · `409` conflito de estado · `412` `If-Match` divergente · `428` `If-Match` ausente
- **Corpo de erro em Problem Details, RFC 9457 (D-03)**: `type`, `title`, `status`, `detail` e a lista `erros` para falhas por campo ou por item, produzido por um handler global de exceções — não escreva formato próprio de erro no seu contexto
- Operação que movimenta saldo é transacional (tudo ou nada) e idempotente: header `Idempotency-Key`, com a repetição devolvendo `200` e a resposta original
- Escrita em recurso que pode ser alterado por duas pessoas usa controle otimista: header `If-Match` comparado com o campo `version` do registro. `412` quando divergir, `428` quando o header não vier, e a consulta de detalhe expõe `version`
- **Sem mensageria e sem eventos de domínio**: integração entre contextos é consulta síncrona ou chamada direta dentro da mesma transação

**Escopos oficiais** — esta é a lista completa; não invente escopo fora dela:

| Escopo | O que autoriza |
|---|---|
| `mecanicos:escrever` | Cadastrar mecânicos e definir seus escopos |
| `clientes:ler` | Consultar cliente e seus veículos |
| `clientes:escrever` | Cadastrar, atualizar, inativar e reativar cliente, e vincular veículo |
| `veiculos:ler` | Consultar veículo |
| `veiculos:escrever` | Cadastrar, atualizar, inativar e reativar veículo |
| `os:ler` | Consultar OS, listar OS, fila de atendimento e indicadores de tempo |
| `os:escrever` | Criar OS e registrar problemas, serviços, peças, insumos, execução, finalização e entrega |
| `orcamentos:ler` | Consultar orçamento |
| `orcamentos:escrever` | Calcular e montar orçamento |
| `orcamentos:decidir` | Aprovar ou recusar orçamento — decisão do cliente |
| `servicos:ler` | Consultar o catálogo de serviços |
| `servicos:escrever` | Cadastrar, atualizar, inativar e reativar serviço |
| `estoque:ler` | Consultar peças e insumos |
| `estoque:escrever` | Cadastrar, atualizar, inativar e reativar peça e insumo |
| `estoque:movimentar` | Reservar, liberar, dar entrada e dar baixa em saldo |
| `compras:ler` | Consultar fornecedores |
| `compras:escrever` | Solicitar e cancelar pedido de compra; manter o cadastro de fornecedor |

**Perfis** — `MECANICO` (a oficina), `CLIENTE` (aprova e acompanha a própria OS) e `SERVICO`
(chamadas máquina a máquina). Não existem `GESTOR` nem `ESTOQUISTA`. O corte de permissão é feito
pelo **escopo**, não pelo perfil.

Se o seu contexto precisar divergir de algum item acima, isso é uma decisão do time:
registre em **Pontos em aberto** antes de divergir por conta própria.

**Formatação.** Tabelas markdown de verdade (nada de texto grudado colado do Notion/Word),
blocos de código com a linguagem declarada (` ```json `, ` ```http `), listas numeradas
para fluxos e checkbox só no bloco 3.

**Sem ícones.** Não use emoji nem ícones em nenhum ponto do documento — nem para destacar
item crítico, nem em título, nem em checklist. Destaque se faz com **negrito** ou com o texto
da própria frase. Item crítico que precisa de atenção vira uma linha explícita no checklist
ou um registro em **Pontos em aberto**.

---

## 9. Checklist antes de abrir o PR

- [ ] Arquivo está na pasta do contexto, com o nome da tarefa, minúsculo e sem acento.
- [ ] Frontmatter completo (documento, dono, versao, atualizado_em, status).
- [ ] Todos os requisitos têm os três blocos, na ordem certa.
- [ ] Todo RF e RNF tem ID com a sigla correta do contexto.
- [ ] Fluxo principal e fluxos alternativos cobrem também os casos de erro e de autorização.
- [ ] Os testes do bloco 2 e os do bloco 3 batem entre si.
- [ ] O JSON de exemplo é internamente consistente.
- [ ] Rota, envelope e códigos HTTP seguem os padrões compartilhados da seção 8.
- [ ] Nada inventado: o que não foi decidido está em **Pontos em aberto**.
- [ ] Um integrante de outro contexto leu e entendeu sem explicação verbal.

---

## 10. Fluxo de contribuição

1. Crie a branch `docs/<contexto>-<tarefa>`.
2. Copie o modelo da seção 11 (ou um arquivo de [pecas/](pecas/)) e preencha.
   Atualize `00-resumo.md` e [03-endpoints.md](03-endpoints.md) no mesmo PR.
3. Abra o PR com `status: em revisao`, marcando dois revisores — um deles do contexto vizinho.
4. Divergência de termo entre contextos se resolve escolhendo **um** termo, e os dois
   documentos são atualizados no mesmo PR.
5. Após o merge: `status: aprovado` e versão incrementada.
6. Mudou regra depois? Mesmo fluxo, com `atualizado_em` novo. PR que muda regra de negócio
   e não atualiza o documento não é aprovado.

---

## 11. Modelo pronto para copiar

````markdown
---
documento: Refinamento de Requisitos — Contexto de <Nome do Contexto>
dono: <Seu nome>
versao: 0.1
atualizado_em: AAAA-MM-DD
status: rascunho
---

# Refinamento de Requisitos

Este documento reúne, para cada requisito levantado da aplicação, três blocos:

1. **Refinamento de Produto** — o que o usuário precisa e por quê (visão de negócio).
2. **Refinamento Técnico** — como o sistema entrega isso (contrato, processamento, testes).
3. **Checklist de Implementação** — o passo a passo verificável até o merge.

---

## 1 · <Nome do Requisito>

### 1.1 Refinamento de Produto

**Persona**
<quem executa>

**Objetivo**
<o que quer conseguir e para quê>

**Problema**
<a dor de hoje e a consequência concreta>

**Pré-condições**

- <o que precisa existir antes>

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-XXX-01 | <o que o sistema faz> |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-XXX-01 | <como o sistema se comporta> |

**Fluxo Principal**

1. <passo>

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | <situação> | <o que o sistema faz> |

**Saída**

- <o que a pessoa recebe>

**Pós-condições**

- <o que mudou — e o que explicitamente não mudou>

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
GET /<recurso>
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `<PERFIL>`
- Escopo: `<recurso>:<acao>`

**Entrada**

| Param | Tipo | Descrição |
|---|---|---|
| `<param>` | `<tipo>` | <descrição, default, obrigatoriedade> |

**Validações**

- <limite ou formato>

**Processamento**

1. <passo interno>

**Persistência**

- Consulta: `<tabela>`
- Altera: `<tabela ou nada>`

**Saída da API**

```json
{
  "data": []
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | <situação> |
| `400` | <entrada inválida> |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem o escopo `<recurso>:<acao>` |

**Dependências**

- `<Repositório / middleware / outro contexto>`

**Testes**

*Unitários*

- <regra de domínio isolada>

*Integração*

- <contrato da API ponta a ponta>

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] <regra ou cálculo>

**Caso de uso**

- [ ] <orquestração>

**Repositório**

- [ ] <acesso a dados>

**Handler HTTP**

- [ ] <rota e envelope>

**Validações**

- [ ] <limite de entrada>

**Testes unitários**

- [ ] <caso>

**Testes de integração**

- [ ] <caso>

**Documentação**

- [ ] Documentar no Swagger/OpenAPI

**Review**

- [ ] Code Review aprovado

---

## Pontos em aberto

| # | Ponto | Responsável |
|---|---|---|
| 1 | <o que ainda não foi decidido> | — |
````

---

## 12. Erros comuns

- **Arquivo solto em `docs/`.** Fora da pasta do contexto, o documento não é encontrado por quem chega depois.
- **Misturar duas tarefas no mesmo arquivo.** Um arquivo, uma tarefa.
- **Rota nova sem entrar no catálogo.** Toda rota entra em [03-endpoints.md](03-endpoints.md) no mesmo PR.
- **Tabela colada de outra ferramenta.** Cabeçalho grudado (`ParamTipoDescrição`) não é tabela — remonte em markdown.
- **Bloco técnico invadindo o de produto.** Verbo HTTP e nome de tabela não aparecem no bloco 1.
- **Checklist genérico.** "Fazer o estoque" não é tarefa; "Implementar `ItemEstoqueRepository.buscarPorFiltro`" é.
- **Divergência entre o contrato e o checklist.** Rota, campo e código HTTP precisam ser os mesmos nos dois blocos.
- **Regra inventada para não deixar lacuna.** Lacuna vai para **Pontos em aberto**.
- **Ícone no lugar de texto.** Emoji de alerta não sobrevive a copiar e colar, não é pesquisável e não diz qual é o risco. Escreva a frase.
- **Persona que não bate com o perfil de acesso.** Se a persona é o Estoquista, o perfil autorizado no bloco técnico precisa refletir isso.
- **Termos concorrentes.** "Peça", "insumo", "item", "produto" para a mesma coisa: escolha um e use sempre.
