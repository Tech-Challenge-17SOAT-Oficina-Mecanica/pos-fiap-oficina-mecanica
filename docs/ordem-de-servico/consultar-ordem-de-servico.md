---
documento: Refinamento de Requisitos — Consultar Ordem de Serviço
dono: A definir
versao: 0.3
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Ordem de Serviço

Este documento detalha a tarefa Consultar Ordem de Serviço do contexto de Ordem de Serviço.

## 11 · Consultar Ordem de Serviço

### 11.1 Refinamento de Produto

**Persona**

Mecânico. Também o cliente, quando consulta o progresso da própria Ordem de Serviço.

**Objetivo**

Consultar os detalhes, o status atual, os orçamentos e o histórico de eventos de uma Ordem de Serviço.

**Problema**

A oficina precisa permitir o acompanhamento do status dos serviços, dos orçamentos gerados e dos
eventos ocorridos durante o fluxo da Ordem de Serviço. É o que sustenta a exigência do enunciado
de o cliente acompanhar o progresso via API.

**Pré-condições**

- A Ordem de Serviço deve existir.
- O usuário deve estar autorizado a consultar a Ordem de Serviço.
- Quando a consulta for feita pelo cliente, a OS deve pertencer a ele.
- Quando a consulta for feita por documento, deve existir cliente vinculado ao CPF/CNPJ informado.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-101 | Permitir consultar uma Ordem de Serviço pelo identificador da OS. |
| RF-OS-102 | Permitir consultar Ordem de Serviço pelo CPF/CNPJ do cliente. |
| RF-OS-103 | Retornar o status atual da Ordem de Serviço. |
| RF-OS-104 | Retornar os dados do cliente vinculado à Ordem de Serviço. |
| RF-OS-105 | Retornar os dados do veículo vinculado à Ordem de Serviço. |
| RF-OS-106 | Retornar os problemas da OS com o orçamento a que cada um está vinculado. |
| RF-OS-107 | Retornar os orçamentos vinculados à Ordem de Serviço. |
| RF-OS-108 | Retornar os itens de cada orçamento. |
| RF-OS-109 | Identificar cada orçamento como `PRINCIPAL` ou `COMPLEMENTAR`. |
| RF-OS-110 | Retornar o vínculo entre orçamento complementar e orçamento original, quando houver. |
| RF-OS-111 | Retornar o valor total geral dos orçamentos. |
| RF-OS-112 | Retornar os eventos da Ordem de Serviço em `eventos`. |
| RF-OS-113 | Permitir que o cliente acompanhe o progresso da Ordem de Serviço via API. |
| RF-OS-114 | Permitir o detalhamento administrativo da Ordem de Serviço. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-53 | A consulta deve ser feita por API RESTful. |
| RNF-OS-54 | A operação deve respeitar autenticação e autorização. |
| RNF-OS-55 | A consulta não deve alterar os dados da Ordem de Serviço. |
| RNF-OS-56 | A consulta não deve alterar dados de orçamento, cliente, veículo ou eventos. |
| RNF-OS-57 | A resposta deve refletir o status atual da Ordem de Serviço. |
| RNF-OS-58 | O `eventos` deve funcionar como histórico técnico e de negócio, sem substituir o status atual da OS. |

**Fluxo Principal**

1. O usuário solicita a consulta da Ordem de Serviço.
2. O sistema valida se a consulta foi feita pelo identificador da OS ou pelo CPF/CNPJ do cliente.
3. O sistema valida a permissão de acesso.
4. O sistema verifica se a Ordem de Serviço existe.
5. O sistema consulta os dados da Ordem de Serviço.
6. O sistema consulta os dados do cliente e do veículo vinculados.
7. O sistema consulta os problemas vinculados à OS.
8. O sistema consulta os orçamentos vinculados e os itens de cada um.
9. O sistema consulta os eventos da OS em `eventos`.
10. O sistema calcula o valor total geral dos orçamentos.
11. O sistema retorna os detalhes consolidados da Ordem de Serviço.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | Informa que a Ordem de Serviço não existe. |
| A2 | Cliente não encontrado pelo documento | Informa que não existe cliente para o CPF/CNPJ informado. |
| A3 | Documento inválido | Informa que o CPF/CNPJ informado é inválido. |
| A4 | Usuário sem autorização | Impede a consulta. |
| A5 | Cliente consultando OS de outro cliente | Impede o acesso. |
| A6 | Nenhum orçamento encontrado | Retorna a OS com lista de orçamentos vazia. |
| A7 | Nenhum evento encontrado | Retorna a OS com `eventos` vazio. |
| A8 | Erro ao consultar | Informa que não foi possível concluir a consulta. |

**Saída**

- Dados da Ordem de Serviço, com status atual, cliente, veículo, problemas, orçamentos e seus
  itens, valor total geral e eventos em `eventos`.

**Pós-condições**

- A Ordem de Serviço, os orçamentos e os eventos permanecem inalterados.
- O usuário passa a conhecer o status atual, os orçamentos e o histórico de eventos da OS.
- A Ordem de Serviço pode seguir seu fluxo conforme o status atual.

---

### 11.2 Refinamento Técnico

**Endpoint**

```http
GET /ordens-servico/{osId}
```

A consulta pelo CPF/CNPJ do cliente é atendida pela listagem, em
[`listar-ordens-de-servico.md`](listar-ordens-de-servico.md), com o filtro `documento`.

> **Decisão de projeto.** Esta rota **detalha uma OS**; a listagem com filtros é
> [listar-ordens-de-servico.md](listar-ordens-de-servico.md), em `GET /ordens-servico`. As duas
> tarefas chegaram propondo a mesma rota, e esta é a divisão confirmada.

> **Decisão de projeto.** O bloco chama-se **`eventos`**, e não `event_data`: era o único campo em
> `snake_case` numa resposta toda em `camelCase`, e os campos internos estavam em inglês. Foram
> traduzidos junto — `agregado`, `agregadoId`, `tipoEvento`, `dados`, `metadados`, `ocorridoEm` e
> `registradoEm`.

> **Decisão de projeto.** Este histórico é **trilha de auditoria**, não mensageria: ele sobreviveu à
> decisão de não usar eventos de domínio no projeto (DT-35). Ninguém publica nem consome esses
> registros; eles são gravados para contar o que aconteceu com a OS.

> **Decisão de projeto.** O `eventos` é o histórico técnico e de negócio da OS: cada registro
> guarda o agregado de origem, o tipo de evento, a transição de status e o payload do que
> aconteceu. Ele **não** substitui o status atual da OS — o status vive na própria Ordem de
> Serviço, e o `eventos` explica como ela chegou até ele.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO` e `CLIENTE`. O cliente consulta apenas o progresso da própria OS.

> **Decisão de projeto.** O cliente se autentica por **token de escopo reduzido**, emitido no envio
> do orçamento e válido apenas para aquela OS. É o mesmo mecanismo já adotado em Orçamento para
> aprovar e recusar, e evita criar cadastro com senha para cliente no MVP (DT-30).
- Escopo: `os:ler`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |

**Validações**

*Técnicas*

- `osId` informado e em formato UUID válido.

*Negócio*

- A Ordem de Serviço deve existir.
- O usuário deve ter permissão para consultar a OS; o cliente só acessa a própria.
- A consulta não altera dados da OS.

**Processamento**

1. Receber o identificador da OS.
2. Validar o parâmetro informado e a permissão de acesso.
3. Consultar a Ordem de Serviço.
4. Carregar os dados do cliente e do veículo.
5. Carregar os problemas vinculados à OS.
6. Carregar os orçamentos vinculados e os itens de cada orçamento.
7. Carregar os eventos da OS em `eventos`.
8. Calcular o `valorTotalGeral` somando os valores totais dos orçamentos retornados.
9. Retornar os dados consolidados da OS.

**Persistência**

- Consulta: `ordem_servico`, `cliente`, `veiculo`, problemas da OS, `orcamento`, `orcamento_item`,
  `eventos`.
- Altera: nada.

**Saída da API**

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "statusOrdemServico": "AGUARDANDO_APROVACAO",
  "cliente": {
    "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
    "documento": "00000000000"
  },
  "veiculo": {
    "veiculoId": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
    "placa": "ABC1D23",
    "marca": "Marca",
    "modelo": "Modelo",
    "ano": 2020
  },
  "problemas": [
    {
      "problemaId": "a3f60c81-7d24-4e59-b016-8c5f2b93ea47",
      "descricao": "Barulho ao frear",
      "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
      "exigeNovaAprovacao": false,
      "identificadoEm": "2026-08-18T10:00:00-03:00"
    }
  ],
  "orcamentos": [
    {
      "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
      "tipo": "PRINCIPAL",
      "orcamentoOriginalId": null,
      "itens": [
        {
          "tipo": "SERVICO",
          "descricao": "Troca de óleo",
          "quantidade": 1,
          "valorUnitario": 150.0,
          "valorTotal": 150.0
        },
        {
          "tipo": "PECA",
          "descricao": "Filtro de óleo",
          "quantidade": 1,
          "valorUnitario": 50.0,
          "valorTotal": 50.0
        }
      ],
      "valorTotal": 200.0,
      "dataGeracao": "2026-08-18T10:30:00-03:00"
    },
    {
      "orcamentoId": "b1d47c60-92fe-4a38-8c15-73e0a6b5d284",
      "tipo": "COMPLEMENTAR",
      "orcamentoOriginalId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
      "itens": [
        {
          "tipo": "PECA",
          "descricao": "Correia dentada",
          "quantidade": 1,
          "valorUnitario": 150.0,
          "valorTotal": 150.0
        }
      ],
      "valorTotal": 150.0,
      "dataGeracao": "2026-08-19T10:30:00-03:00"
    }
  ],
  "valorTotalGeral": 350.0,
  "eventos": [
    {
      "id": "2f8c1a47-6b03-4d92-9e75-40ab3c186de5",
      "agregado": "ORDEM_SERVICO",
      "agregadoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "tipoEvento": "ORDEM_SERVICO_CRIADA",
      "statusAnterior": null,
      "statusNovo": "RECEBIDA",
      "etapa": "ATENDIMENTO",
      "dados": {
        "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
        "veiculoId": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607"
      },
      "metadados": {
        "usuarioId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
      },
      "ocorridoEm": "2026-08-18T09:30:00-03:00",
      "registradoEm": "2026-08-18T09:30:01-03:00"
    },
    {
      "id": "7e05b93c-8f14-42a6-b0d7-51c9a2e63f80",
      "agregado": "ORCAMENTO",
      "agregadoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "tipoEvento": "ORCAMENTO_GERADO",
      "statusAnterior": null,
      "statusNovo": null,
      "etapa": "ORCAMENTO",
      "dados": {
        "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
        "tipo": "PRINCIPAL",
        "valorTotal": 200.0
      },
      "metadados": {
        "usuarioId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
      },
      "ocorridoEm": "2026-08-18T10:30:00-03:00",
      "registradoEm": "2026-08-18T10:30:01-03:00"
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Consulta realizada com sucesso. |
| `400` | Identificador ausente ou inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem permissão para consultar a OS. |
| `404` | Ordem de Serviço não encontrada. |

**Dependências**

- `OrdemDeServicoRepository`.
- `ClienteRepository` e `VeiculoRepository`.
- Repositório de problemas da OS.
- `OrcamentoRepository` e `OrcamentoItemRepository`.
- `EventDataRepository`.
- Validador de CPF/CNPJ, na consulta por documento.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Cálculo do `valorTotalGeral` como soma dos orçamentos retornados.
- Orçamento principal com tipo `PRINCIPAL` e `orcamentoOriginalId` nulo.
- Orçamento complementar com tipo `COMPLEMENTAR` e `orcamentoOriginalId` preenchido.
- Cada problema retornado com o `orcamentoId` a que está vinculado.

> **Decisão de projeto.** O problema **não tem tipo próprio** — quem tem tipo é o orçamento
> (`PRINCIPAL` ou `COMPLEMENTAR`), deduzido do status da OS, como definido em
> [registrar-problema-encontrado.md](registrar-problema-encontrado.md). Para o consumidor saber se
> o problema veio do diagnóstico ou apareceu na execução, a resposta expõe o `orcamentoId` do
> vínculo: o tipo do orçamento correspondente responde a pergunta. A tabela `orcamento_problema` já
> guarda esse vínculo; ele só não vinha na consulta.
- `eventos` não substitui o status atual da OS.

*Integração*

- Consulta pelo identificador retorna `200` com os dados consolidados.
- A resposta traz cliente, veículo, problemas, orçamentos, itens e eventos.
- OS sem orçamento retorna lista de orçamentos vazia.
- OS sem evento retorna `eventos` vazio.
- Identificador inválido retorna `400`.
- OS inexistente retorna `404`.
- Sem token retorna `401`.
- Cliente consultando OS de outro cliente retorna `403`.
- A consulta não altera dados persistidos.

---

### 11.3 Checklist de Implementação

> **Nota de implementação (2026-08-27).** Implementado em `internal/domain/ordemservico`,
> `internal/application/ordemservico` e `internal/infrastructure/ordemservico` (reaproveitando o
> repositório existente). Três desvios do refinamento original:
> `eventos` não tem colunas separadas de `statusAnterior`/`statusNovo` no schema
> (`auditoria_ordem_servico`) — quando presentes, esses dados vivem dentro de `dados`, não como
> campos próprios da resposta; o campo `exigeNovaAprovacao` do problema não existe em nenhuma
> fonte de dados atual e foi omitido; e a consulta por CPF/CNPJ do cliente não foi implementada
> aqui — conforme o próprio doc diz, ela é atendida por `GET /ordens-servico` (listar-ordens-de-servico.md),
> que não foi implementado nesta entrega.

**Domínio**

- [x] Garantir que a Ordem de Serviço possua identificador único e status atual
- [x] Garantir que a OS mantenha vínculo com `Cliente` e com `Veiculo`
- [x] Retornar cada problema com o `orcamentoId` do vínculo em `problema_ordem_servico.orcamento_id`
- [x] Garantir que `Orcamento` tenha `tipo` `PRINCIPAL` ou `COMPLEMENTAR` e possa referenciar o orçamento original
- [x] Criar ou ajustar o modelo `ItemOrcamento` (`ItemOrcamentoConsulta`)
- [x] Criar ou ajustar o modelo `EventData` como histórico técnico e de negócio da OS (`EventoConsulta`, a partir de `auditoria_ordem_servico`)
- [x] Garantir que `EventData` não substitua o status atual da OS

**Caso de uso**

- [x] Implementar `ConsultarOrdemDeServico`
- [x] Validar que o identificador da OS foi informado e tem formato válido
- [x] Validar a permissão de acesso à Ordem de Serviço
- [x] Consultar cliente, veículo, problemas, orçamentos, itens e eventos
- [x] Calcular o valor total geral dos orçamentos retornados
- [x] Garantir que a consulta não altere dados persistidos

**Repositório**

- [x] Criar o método que busca Ordem de Serviço por identificador
- [x] Criar o método que carrega os dados vinculados à OS
- [x] Criar ou ajustar `OrcamentoRepository` para os orçamentos da OS (consulta própria em `internal/infrastructure/ordemservico`, sem reaproveitar o repositório de `orcamento`)
- [x] Criar ou ajustar `EventDataRepository` para os eventos da OS

**Handler HTTP**

- [x] Implementar `GET /ordens-servico/{osId}`
- [x] Implementar a validação do path param `osId`
- [x] Criar DTO/response com os dados detalhados da Ordem de Serviço
- [x] Aplicar autenticação JWT e autorização na rota (escopo `os:ler` ou `orcamentos:ler`, com checagem de dono para token de cliente)
- [x] Mapear erros de domínio para os códigos HTTP documentados

**Testes unitários**

- [x] Consulta válida por identificador (delegação ao repositório)
- [ ] Identificador ausente ou inválido, OS inexistente, usuário sem permissão como testes unitários isolados — cobertos pelo teste do handler e pelo teste de integração, pois a regra vive na query SQL do repositório
- [x] Retorno do status atual, do cliente e do veículo (teste do handler)
- [x] Problemas retornados com o `orcamentoId` do vínculo (teste de integração)
- [x] Orçamento principal e complementar com os tipos corretos (reaproveita a query já testada em `orcamento`)
- [x] Vínculo do complementar com o orçamento original
- [x] Cálculo do valor total geral (teste de integração)
- [x] Retorno dos eventos em `eventos` (teste de integração)

**Testes de integração**

- [x] Consulta por identificador da OS
- [x] OS sem orçamento e OS sem evento (listas vazias, não nulas)
- [x] A consulta não altera dados persistidos (nenhum `UPDATE`/`INSERT` no fluxo de leitura)

**Documentação**

- [x] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [x] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [x] Executar testes automatizados (unitários; integração escrita, não executada nesta sessão por falta de Docker)
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
