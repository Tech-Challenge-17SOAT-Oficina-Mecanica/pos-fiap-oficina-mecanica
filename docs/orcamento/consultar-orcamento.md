---
documento: Refinamento de Requisitos — Consultar Orçamento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Orçamento

Este documento detalha a tarefa Consultar Orçamento do contexto de Orçamento.

## 2 · Consultar Orçamento

### 2.1 Refinamento de Produto

**Persona**

Cliente.

**Objetivo**

Consultar os serviços, peças, insumos e valores dos orçamentos vinculados às suas Ordens de Serviço.

**Problema**

O cliente precisa analisar o orçamento principal e seus complementares antes de aprová-los ou
recusá-los.

**Pré-condições**

- Deve existir uma OS vinculada ao cliente.
- Deve existir ao menos um orçamento associado à OS.
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
| RF-ORC-14 | Apresentar orçamento principal e orçamentos complementares. |
| RF-ORC-15 | Apresentar o valor total geral da OS. |
| RF-ORC-16 | Apresentar o status atual da OS. |
| RF-ORC-17 | Impedir acesso a orçamento de outro cliente. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-04 | A consulta deve exigir autenticação e autorização. |
| RNF-ORC-05 | A consulta não deve alterar os dados do orçamento. |
| RNF-ORC-06 | O cliente deve visualizar somente informações vinculadas às suas OS. |
| RNF-ORC-07 | O documento deve ser validado por se tratar de dado sensível. |

**Fluxo Principal**

1. O cliente informa o identificador do orçamento ou seu documento.
2. O sistema valida a identificação do cliente.
3. O sistema valida os critérios de consulta.
4. O sistema localiza a OS e os orçamentos associados.
5. O sistema valida se a OS pertence ao cliente.
6. O sistema apresenta o orçamento principal e os complementares.
7. O sistema apresenta itens, valores e o valor total geral.
8. O sistema apresenta o status atual da OS.
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

**Saída**

- Orçamentos principal e complementares consultados pelo cliente.

**Pós-condições**

- O cliente visualiza os dados dos orçamentos.
- Nenhum dado do orçamento ou da OS é alterado.

---

### 2.2 Refinamento Técnico

**Endpoint**

```http
GET /orcamentos
```

Aceita consulta por identificador do orçamento ou pelo documento do cliente. Deve ser informado
ao menos um dos dois.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Permitido apenas para o cliente vinculado à OS dos orçamentos.
- Escopo: `orcamentos:ler`.
- Operação somente leitura.

**Entrada** — query params:

| Param | Tipo | Descrição |
|---|---|---|
| `orcamentoId` | uuid | Identificador do orçamento. |
| `documento` | string | CPF ou CNPJ do cliente. |
| `page` / `size` | int | Paginação, quando a busca for por documento. |

Exemplos: `GET /orcamentos?orcamentoId=9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913` e
`GET /orcamentos?documento=00000000000&page=0&size=20`.

**Validações**

*Técnicas*

- Deve ser informado `orcamentoId` ou `documento`.
- Formato válido de CPF/CNPJ, quando o documento for informado.
- Parâmetros de paginação válidos.

*Negócio*

- O orçamento deve existir, quando informado.
- A OS vinculada deve pertencer ao cliente autenticado.
- O documento informado deve pertencer ao cliente autenticado.
- A consulta não altera dados.

**Processamento**

1. Receber os critérios da consulta e identificar o cliente autenticado.
2. Validar os critérios informados.
3. Consultar os orçamentos pelo identificador ou pelo documento.
4. Consultar a OS vinculada.
5. Validar o vínculo entre cliente, OS e orçamento.
6. Consultar o orçamento principal e os complementares vinculados à mesma OS.
7. Consultar os itens de cada orçamento.
8. Calcular o valor total geral dos orçamentos da OS.
9. Montar e retornar a resposta, com o status atual da OS.

**Persistência**

- Consulta: `cliente`, `orcamento`, `orcamento_item`, `ordem_servico`.
- Altera: nada.

**Saída da API**

```json
{
  "data": [
    {
      "cliente": {
        "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
        "documento": "00000000000"
      },
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "statusOrdemServico": "AGUARDANDO_APROVACAO",
      "orcamentos": [
        {
          "orcamentoId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
          "tipo": "PRINCIPAL",
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
      "valorTotalGeral": 350.0
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 1,
  "totalPaginas": 1
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Orçamentos consultados com sucesso. |
| `400` | Nenhum critério informado; documento inválido; paginação inválida. |
| `401` | Token ausente ou expirado. |
| `403` | Cliente sem acesso aos orçamentos. |
| `404` | Orçamento ou cliente não encontrado. |

**Dependências**

- `ClienteRepository`.
- `OrcamentoRepository`.
- `OrcamentoItemRepository`.
- `OrdemDeServicoRepository`.
- Validador de CPF/CNPJ.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Cálculo do valor total de cada orçamento e do valor total geral.
- Rejeita consulta sem nenhum critério informado.
- Rejeita documento inválido.

*Integração*

- Consulta pelo `orcamentoId` retorna o orçamento correspondente.
- Consulta pelo documento retorna os orçamentos do cliente.
- A resposta traz `clienteId`, `documento` e o status atual da OS.
- A resposta traz orçamento principal e complementares.
- Nenhum critério informado retorna `400`.
- Documento inválido retorna `400`.
- Orçamento ou cliente inexistente retorna `404`.
- Sem token retorna `401` e orçamento de outro cliente retorna `403`.
- A consulta não altera dados do orçamento nem da OS.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar os campos `tipo` e `orcamentoOriginalId` no orçamento
- [ ] Definir a regra de composição do valor total geral da OS

**Caso de uso**

- [ ] Implementar `ConsultarOrcamento`
- [ ] Implementar a busca de orçamento por identificador
- [ ] Implementar a busca de orçamentos pelo documento do cliente
- [ ] Implementar a busca do orçamento principal e dos complementares
- [ ] Validar o vínculo entre orçamento, OS e cliente autenticado
- [ ] Calcular e retornar o valor total geral
- [ ] Garantir que a consulta não altera dados persistidos

**Repositório**

- [ ] Criar ou ajustar `ClienteRepository`
- [ ] Criar ou ajustar `OrcamentoRepository`
- [ ] Criar ou ajustar `OrcamentoItemRepository`

**Handler HTTP**

- [ ] Implementar `GET /orcamentos`
- [ ] Criar DTO de resposta com cliente, OS, orçamentos e itens
- [ ] Implementar o envelope de resposta paginado
- [ ] Aplicar autenticação JWT e autorização para o cliente vinculado à OS

**Validações**

- [ ] Validar os query params `orcamentoId`, `documento`, `page` e `size`
- [ ] Validar o CPF/CNPJ informado
- [ ] Retornar `400`, `401`, `403` e `404` conforme documentado

**Testes unitários**

- [ ] Cálculo do valor total geral
- [ ] Consulta sem critério informado
- [ ] Documento inválido

**Testes de integração**

- [ ] Orçamento principal e complementar na resposta
- [ ] Consulta por identificador e por documento
- [ ] Cliente sem acesso
- [ ] Consulta não altera dados persistidos

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
