---
documento: Refinamento de Requisitos — Consultar Ordem de Serviço
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Ordem de Serviço

Este documento detalha a tarefa Consultar Ordem de Serviço do contexto de Ordem de Serviço.

## 9 · Consultar Ordem de Serviço

### 9.1 Refinamento de Produto

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
| RF-OS-63 | Permitir consultar uma Ordem de Serviço pelo identificador da OS. |
| RF-OS-64 | Permitir consultar Ordem de Serviço pelo CPF/CNPJ do cliente. |
| RF-OS-65 | Retornar o status atual da Ordem de Serviço. |
| RF-OS-66 | Retornar os dados do cliente vinculado à Ordem de Serviço. |
| RF-OS-67 | Retornar os dados do veículo vinculado à Ordem de Serviço. |
| RF-OS-68 | Retornar os problemas vinculados à Ordem de Serviço, sem expor o tipo do problema na resposta. |
| RF-OS-69 | Retornar os orçamentos vinculados à Ordem de Serviço. |
| RF-OS-70 | Retornar os itens de cada orçamento. |
| RF-OS-71 | Identificar cada orçamento como `PRINCIPAL` ou `COMPLEMENTAR`. |
| RF-OS-72 | Retornar o vínculo entre orçamento complementar e orçamento original, quando houver. |
| RF-OS-73 | Retornar o valor total geral dos orçamentos. |
| RF-OS-74 | Retornar os eventos da Ordem de Serviço em `event_data`. |
| RF-OS-75 | Permitir que o cliente acompanhe o progresso da Ordem de Serviço via API. |
| RF-OS-76 | Permitir o detalhamento administrativo da Ordem de Serviço. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-39 | A consulta deve ser feita por API RESTful. |
| RNF-OS-40 | A operação deve respeitar autenticação e autorização. |
| RNF-OS-41 | A consulta não deve alterar os dados da Ordem de Serviço. |
| RNF-OS-42 | A consulta não deve alterar dados de orçamento, cliente, veículo ou eventos. |
| RNF-OS-43 | A resposta deve refletir o status atual da Ordem de Serviço. |
| RNF-OS-44 | O `event_data` deve funcionar como histórico técnico e de negócio, sem substituir o status atual da OS. |

**Fluxo Principal**

1. O usuário solicita a consulta da Ordem de Serviço.
2. O sistema valida se a consulta foi feita pelo identificador da OS ou pelo CPF/CNPJ do cliente.
3. O sistema valida a permissão de acesso.
4. O sistema verifica se a Ordem de Serviço existe.
5. O sistema consulta os dados da Ordem de Serviço.
6. O sistema consulta os dados do cliente e do veículo vinculados.
7. O sistema consulta os problemas vinculados à OS.
8. O sistema consulta os orçamentos vinculados e os itens de cada um.
9. O sistema consulta os eventos da OS em `event_data`.
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
| A7 | Nenhum evento encontrado | Retorna a OS com `event_data` vazio. |
| A8 | Erro ao consultar | Informa que não foi possível concluir a consulta. |

**Saída**

- Dados da Ordem de Serviço, com status atual, cliente, veículo, problemas, orçamentos e seus
  itens, valor total geral e eventos em `event_data`.

**Pós-condições**

- A Ordem de Serviço, os orçamentos e os eventos permanecem inalterados.
- O usuário passa a conhecer o status atual, os orçamentos e o histórico de eventos da OS.
- A Ordem de Serviço pode seguir seu fluxo conforme o status atual.

---

### 9.2 Refinamento Técnico

**Endpoint**

```http
GET /ordens-servico/{osId}
```

A consulta pelo CPF/CNPJ do cliente é atendida pela listagem, em
[`listar-ordens-de-servico.md`](listar-ordens-de-servico.md), com o filtro `documento`.

> **Decisão de projeto.** O `event_data` é o histórico técnico e de negócio da OS: cada registro
> guarda o agregado de origem, o tipo de evento, a transição de status e o payload do que
> aconteceu. Ele **não** substitui o status atual da OS — o status vive na própria Ordem de
> Serviço, e o `event_data` explica como ela chegou até ele.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`. O cliente pode consultar o progresso da própria OS.
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
7. Carregar os eventos da OS em `event_data`.
8. Calcular o `valorTotalGeral` somando os valores totais dos orçamentos retornados.
9. Retornar os dados consolidados da OS.

**Persistência**

- Consulta: `ordem_servico`, `cliente`, `veiculo`, problemas da OS, `orcamento`, `orcamento_item`,
  `event_data`.
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
  "event_data": [
    {
      "id": "2f8c1a47-6b03-4d92-9e75-40ab3c186de5",
      "aggregateType": "ORDEM_SERVICO",
      "aggregateId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "eventType": "ORDEM_SERVICO_CRIADA",
      "statusAnterior": null,
      "statusNovo": "RECEBIDA",
      "etapa": "ATENDIMENTO",
      "payload": {
        "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
        "veiculoId": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607"
      },
      "metadata": {
        "usuarioId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
      },
      "occurredAt": "2026-08-18T09:30:00-03:00",
      "createdAt": "2026-08-18T09:30:01-03:00"
    },
    {
      "id": "7e05b93c-8f14-42a6-b0d7-51c9a2e63f80",
      "aggregateType": "ORCAMENTO",
      "aggregateId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "eventType": "ORCAMENTO_GERADO",
      "statusAnterior": null,
      "statusNovo": null,
      "etapa": "ORCAMENTO",
      "payload": {
        "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
        "tipo": "PRINCIPAL",
        "valorTotal": 200.0
      },
      "metadata": {
        "usuarioId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
      },
      "occurredAt": "2026-08-18T10:30:00-03:00",
      "createdAt": "2026-08-18T10:30:01-03:00"
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
- Problemas retornados sem o campo de tipo.
- `event_data` não substitui o status atual da OS.

*Integração*

- Consulta pelo identificador retorna `200` com os dados consolidados.
- A resposta traz cliente, veículo, problemas, orçamentos, itens e eventos.
- OS sem orçamento retorna lista de orçamentos vazia.
- OS sem evento retorna `event_data` vazio.
- Identificador inválido retorna `400`.
- OS inexistente retorna `404`.
- Sem token retorna `401`.
- Cliente consultando OS de outro cliente retorna `403`.
- A consulta não altera dados persistidos.

---

### 9.3 Checklist de Implementação

**Domínio**

- [ ] Garantir que a Ordem de Serviço possua identificador único e status atual
- [ ] Garantir que a OS mantenha vínculo com `Cliente` e com `Veiculo`
- [ ] Criar ou ajustar o modelo `ProblemaDaOS`, permitindo retorno sem o campo de tipo
- [ ] Garantir que `Orcamento` tenha `tipo` `PRINCIPAL` ou `COMPLEMENTAR` e possa referenciar o orçamento original
- [ ] Criar ou ajustar o modelo `ItemOrcamento`
- [ ] Criar ou ajustar o modelo `EventData` como histórico técnico e de negócio da OS
- [ ] Garantir que `EventData` não substitua o status atual da OS

**Caso de uso**

- [ ] Implementar `ConsultarOrdemDeServico`
- [ ] Validar que o identificador da OS foi informado e tem formato válido
- [ ] Validar a permissão de acesso à Ordem de Serviço
- [ ] Consultar cliente, veículo, problemas, orçamentos, itens e eventos
- [ ] Calcular o valor total geral dos orçamentos retornados
- [ ] Garantir que a consulta não altere dados persistidos

**Repositório**

- [ ] Criar o método que busca Ordem de Serviço por identificador
- [ ] Criar o método que carrega os dados vinculados à OS
- [ ] Criar ou ajustar `OrcamentoRepository` para os orçamentos da OS
- [ ] Criar ou ajustar `EventDataRepository` para os eventos da OS

**Handler HTTP**

- [ ] Implementar `GET /ordens-servico/{osId}`
- [ ] Implementar a validação do path param `osId`
- [ ] Criar DTO/response com os dados detalhados da Ordem de Serviço
- [ ] Aplicar autenticação JWT e autorização na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Testes unitários**

- [ ] Consulta válida por identificador
- [ ] Identificador ausente ou inválido
- [ ] Ordem de Serviço inexistente
- [ ] Usuário sem permissão
- [ ] Retorno do status atual, do cliente e do veículo
- [ ] Problemas retornados sem o campo de tipo
- [ ] Orçamento principal e complementar com os tipos corretos
- [ ] Vínculo do complementar com o orçamento original
- [ ] Cálculo do valor total geral
- [ ] Retorno dos eventos em `event_data`

**Testes de integração**

- [ ] Consulta por identificador da OS
- [ ] OS sem orçamento e OS sem evento
- [ ] A consulta não altera dados persistidos

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
