---
documento: Refinamento de Requisitos — Consultar Orçamento
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Consultar Orçamento

Este documento detalha a tarefa Consultar Orçamento do contexto de Orçamento.

## 2 · Consultar Orçamento

### 2.1 Refinamento de Produto

**Persona**

Cliente.

**Objetivo**

Consultar os serviços, peças, insumos, valores, status e estimativa de entrega dos orçamentos
vinculados às suas OS.

**Problema**

O cliente precisa analisar o orçamento principal e seus complementares, incluindo os valores, o
status de cada orçamento e a estimativa de entrega, antes de aprová-los ou recusá-los.

**Pré-condições**

- Deve existir uma OS vinculada ao cliente.
- Deve existir um orçamento principal associado à OS.
- Orçamentos complementares são opcionais e devem estar vinculados ao orçamento principal da mesma OS.
- O cliente deve estar autenticado e autorizado.
- Deve ser informado o identificador do orçamento ou o documento do cliente.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-ORC-08 | Permitir ao cliente consultar orçamento por identificador. |
| RF-ORC-09 | Permitir ao cliente consultar orçamentos pelo seu documento. |
| RF-ORC-10 | Apresentar o identificador e o documento do cliente. |
| RF-ORC-11 | Apresentar serviços, peças e insumos. |
| RF-ORC-12 | Apresentar quantidades, valores unitários e totais. |
| RF-ORC-13 | Apresentar o valor total de cada orçamento. |
| RF-ORC-14 | Apresentar orçamento principal e orçamentos complementares, quando existirem. |
| RF-ORC-15 | Apresentar o valor total geral da OS. |
| RF-ORC-16 | Apresentar o status atual da OS. |
| RF-ORC-17 | Impedir acesso a orçamento de outro cliente. |
| RF-ORC-33 | Apresentar o `tipoOrcamento` de cada orçamento: `PRINCIPAL` ou `COMPLEMENTAR`. |
| RF-ORC-34 | Apresentar o `statusOrcamento` de cada orçamento: `CRIADO`, `APROVADO` ou `RECUSADO`. |
| RF-ORC-35 | Apresentar a estimativa de entrega de cada orçamento em dias. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-04 | A consulta deve exigir autenticação e autorização. |
| RNF-ORC-05 | A consulta não deve alterar os dados do orçamento. |
| RNF-ORC-06 | O cliente deve visualizar somente informações vinculadas às suas OS. |
| RNF-ORC-07 | O documento deve ser validado por se tratar de dado sensível. |
| RNF-ORC-15 | A estimativa deve ser apresentada em dias, sem data exata de entrega. |

**Fluxo Principal**

1. O cliente informa o identificador do orçamento ou seu documento.
2. O sistema valida a identificação do cliente.
3. O sistema valida os critérios de consulta.
4. O sistema localiza a OS e os orçamentos associados.
5. O sistema valida se a OS pertence ao cliente.
6. O sistema consulta o orçamento principal e os complementares, quando existirem.
7. O sistema apresenta itens, valores, tipo, status e estimativa de cada orçamento.
8. O sistema apresenta o valor total geral e o status atual da OS.
9. O cliente consulta as informações.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente não autenticado | Solicita autenticação. |
| A2 | Nenhum critério informado | Informa que é necessário informar orçamento ou documento. |
| A3 | Documento inválido | Informa que o documento não é válido. |
| A4 | Orçamento não encontrado | Informa que não existe orçamento. |
| A5 | Cliente sem acesso | Impede a consulta. |
| A6 | OS em outra etapa | Apresenta os dados sem permitir alteração. |
| A7 | OS sem orçamento complementar | Apresenta somente o orçamento principal. |

**Saída**

- Orçamentos principal e complementares, quando existirem, consultados pelo cliente.

**Pós-condições**

- Cliente visualiza os dados dos orçamentos.
- Cliente visualiza tipo, status e estimativa de entrega de cada orçamento.
- Nenhum dado do orçamento ou da OS é alterado.

---

### 2.2 Refinamento Técnico

**Endpoint**

```http
GET /orcamentos
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Permitido apenas para o cliente vinculado à OS dos orçamentos.
- Escopo: `orcamentos:ler`.
- A operação é somente leitura.

**Entrada**

| Local | Parâmetro | Tipo | Obrigatório | Descrição |
|---|---|---|---|---|
| Query | `orcamentoId` | uuid | Não* | Identificador do orçamento. |
| Query | `documento` | string | Não* | CPF ou CNPJ do cliente. |
| Query | `page` | inteiro | Não | Página da consulta por documento. Padrão: `0`. |
| Query | `size` | inteiro | Não | Quantidade de registros por página. Padrão: `20`. |

*Deve ser informado ao menos um dos parâmetros `orcamentoId` ou `documento`.*

```http
GET /orcamentos?orcamentoId=uuid-do-orcamento
```

```http
GET /orcamentos?documento=00000000000&page=0&size=20
```

**Validações**

*Técnicas*

- `orcamentoId`, quando informado, deve possuir formato UUID válido.
- `page` e `size` devem possuir valores válidos.
- O CPF/CNPJ deve possuir formato válido, quando informado.

*Negócio*

- Deve ser informado `orcamentoId` ou `documento`.
- O orçamento deve existir, quando informado.
- A OS vinculada deve pertencer ao cliente autenticado.
- O documento informado deve pertencer ao cliente autenticado.
- `tipoOrcamento` deve ser `PRINCIPAL` ou `COMPLEMENTAR`.
- `statusOrcamento` deve ser `CRIADO`, `APROVADO` ou `RECUSADO`.
- Quando existir orçamento complementar, seu vínculo com o orçamento principal da mesma OS deve ser válido.
- A consulta não deve alterar dados.

**Processamento**

1. Receber os critérios da consulta.
2. Identificar o cliente autenticado.
3. Validar os critérios informados.
4. Consultar os orçamentos pelo identificador ou documento.
5. Consultar a OS vinculada.
6. Validar o vínculo entre cliente, OS e orçamento.
7. Consultar o orçamento principal e os complementares vinculados à mesma OS.
8. Consultar os itens de cada orçamento.
9. Consultar `tipoOrcamento`, `statusOrcamento` e `estimativaEntregaDias` de cada orçamento.
10. Calcular o valor total geral dos orçamentos da OS.
11. Montar e retornar a resposta com o status atual da OS.

**Persistência**

- Consultar Cliente.
- Consultar os dados de Orçamento.
- Consultar os itens de cada Orçamento.
- Consultar a OS vinculada.
- Consultar os orçamentos complementares associados ao orçamento principal.
- Nenhum dado deve ser alterado.

**Saída da API**

```json
{
  "content": [
    {
      "cliente": {
        "clienteId": "uuid-do-cliente",
        "documento": "00000000000"
      },
      "ordemServicoId": "uuid-da-os",
      "statusOrdemServico": "AGUARDANDO_APROVACAO",
      "orcamentos": [
        {
          "orcamentoId": "uuid-orcamento-principal",
          "tipoOrcamento": "PRINCIPAL",
          "statusOrcamento": "CRIADO",
          "itens": [
            {
              "tipo": "SERVICO",
              "descricao": "Troca de óleo",
              "quantidade": 1,
              "valorUnitario": 150.00,
              "valorTotal": 150.00
            },
            {
              "tipo": "PECA",
              "descricao": "Filtro de óleo",
              "quantidade": 1,
              "valorUnitario": 50.00,
              "valorTotal": 50.00
            }
          ],
          "valorTotal": 200.00,
          "estimativaEntregaDias": 4,
          "dataGeracao": "2026-08-22T10:30:00-03:00"
        },
        {
          "orcamentoId": "uuid-orcamento-complementar",
          "tipoOrcamento": "COMPLEMENTAR",
          "orcamentoOriginalId": "uuid-orcamento-principal",
          "statusOrcamento": "CRIADO",
          "itens": [
            {
              "tipo": "PECA",
              "descricao": "Correia dentada",
              "quantidade": 1,
              "valorUnitario": 150.00,
              "valorTotal": 150.00
            }
          ],
          "valorTotal": 150.00,
          "estimativaEntregaDias": 6,
          "dataGeracao": "2026-08-22T11:30:00-03:00"
        }
      ],
      "valorTotalGeral": 350.00
    }
  ],
  "page": 0,
  "size": 20,
  "totalElements": 1,
  "totalPages": 1
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| 200 OK | Orçamentos consultados com sucesso, inclusive quando a lista estiver vazia. |
| 400 Bad Request | Nenhum critério informado, documento, identificador ou paginação inválidos. |
| 401 Unauthorized | Cliente não autenticado. |
| 403 Forbidden | Cliente sem acesso aos orçamentos. |
| 404 Not Found | Orçamento ou cliente não encontrado. |
| 500 Internal Server Error | Erro inesperado. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de autorização.
- `ClienteRepository`.
- `OrcamentoRepository`.
- `OrcamentoItemRepository`.
- `OrdemDeServicoRepository`.
- Validador de CPF/CNPJ.
- Contexto do cliente autenticado.
- Banco de dados.

**Testes**

*Unitários*

- Deve consultar orçamento pelo `orcamentoId`.
- Deve consultar orçamentos pelo documento do cliente.
- Deve retornar orçamento principal e complementares, quando existirem.
- Deve retornar `tipoOrcamento` como `PRINCIPAL` ou `COMPLEMENTAR`.
- Deve retornar `statusOrcamento` como `CRIADO`, `APROVADO` ou `RECUSADO`.
- Deve retornar a estimativa de entrega em dias.
- Deve retornar o valor total de cada orçamento e o valor total geral da OS.
- Deve retornar somente o orçamento principal quando não existirem complementares.
- Deve garantir que a consulta não altera dados do orçamento ou da OS.

*Integração*

- Deve retornar `clienteId` e documento na resposta.
- Deve retornar `400` quando nenhum critério for informado, ou para documento inválido.
- Deve retornar `401` sem autenticação.
- Deve retornar `403` para orçamento de outro cliente.
- Deve retornar `404` para orçamento ou cliente inexistente.

---

### 2.3 Check-list de Implementação

**Caso de uso**

- [ ] Implementar o caso de uso `ConsultarOrcamento`.
- [ ] Implementar busca de orçamento por identificador.
- [ ] Implementar busca de orçamentos pelo documento do cliente.
- [ ] Implementar busca de orçamento principal e complementares.
- [ ] Validar o vínculo entre orçamento, OS e cliente autenticado.
- [ ] Validar o CPF/CNPJ informado.
- [ ] Calcular e retornar o valor total geral.

**Repositório**

- [ ] Criar/ajustar `ClienteRepository`.
- [ ] Criar/ajustar `OrcamentoRepository`.
- [ ] Criar/ajustar `OrcamentoItemRepository`.
- [ ] Criar/ajustar `OrdemDeServicoRepository`.

**DTOs**

- [ ] Criar DTO de resposta com cliente, OS, orçamentos e itens.
- [ ] Incluir `clienteId` e documento na resposta.
- [ ] Incluir orçamento principal e complementares na resposta.
- [ ] Incluir tipo, status e estimativa de entrega de cada orçamento na resposta.
- [ ] Incluir o status atual da OS na resposta.

**Handler HTTP**

- [ ] Criar handler para `GET /orcamentos`.
- [ ] Validar query params `orcamentoId`, `documento`, `page` e `size`.
- [ ] Aplicar autenticação JWT na rota.
- [ ] Aplicar autorização para o cliente vinculado à OS.
- [ ] Retornar erros `400`, `401`, `403` e `404`.

**Testes unitários**

- [ ] Criar testes para orçamento principal, complementar e ausência de complementar.
- [ ] Criar testes para status, tipo e estimativa no retorno.

**Testes de integração**

- [ ] Criar testes para documento inválido, cliente sem acesso e contrato do endpoint.

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Review**

- [ ] Executar testes automatizados.
- [ ] Realizar code review.
- [ ] Validar critérios de aceite da task.
