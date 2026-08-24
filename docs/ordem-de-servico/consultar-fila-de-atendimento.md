---
documento: Refinamento de Requisitos — Consultar Fila de Atendimento
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Consultar Fila de Atendimento

Este documento detalha a tarefa Consultar Fila de Atendimento do contexto de Ordem de Serviço.

## 7 · Consultar Fila de Atendimento

### 7.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Consultar as Ordens de Serviço disponíveis para execução e identificar aquelas que já possuem
mecânico responsável.

**Problema**

O mecânico precisa visualizar as OS aptas para execução de forma ordenada, garantindo que os
atendimentos já vinculados a um mecânico sejam apresentados primeiro.

**Pré-condições**

- O mecânico deve estar autenticado e autorizado.
- A OS deve estar em `AGUARDANDO_EXECUCAO`.
- A OS deve possuir data de entrada na fila.
- As peças e insumos necessários devem estar disponíveis ou reservados.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-61 | Permitir consultar as OS disponíveis para execução. |
| RF-OS-62 | Considerar como fila as OS em `AGUARDANDO_EXECUCAO` com data de entrada registrada. |
| RF-OS-63 | Exibir OS com mecânico responsável antes das OS sem mecânico. |
| RF-OS-64 | Manter na consulta as OS que já possuem mecânico responsável. |
| RF-OS-65 | Ordenar as OS pela data de entrada na fila dentro de cada grupo. |
| RF-OS-66 | Permitir paginação. |
| RF-OS-67 | Apresentar dados suficientes para identificação da OS e do veículo. |
| RF-OS-68 | Não apresentar OS que não estejam aptas por falta de recursos necessários. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-30 | A consulta deve refletir o estado atual das Ordens de Serviço. |
| RNF-OS-31 | A consulta não deve alterar dados. |
| RNF-OS-32 | O acesso deve exigir autenticação e autorização. |
| RNF-OS-33 | A ordenação deve ser determinística. |
| RNF-OS-34 | A fila não deve exigir estrutura própria de persistência. |

**Fluxo Principal**

1. O mecânico acessa a fila de atendimento.
2. O sistema valida sua autorização.
3. O sistema identifica as OS em `AGUARDANDO_EXECUCAO` com data de entrada na fila.
4. O sistema valida a disponibilidade ou reserva dos recursos necessários.
5. O sistema remove da consulta as OS que não estão aptas.
6. O sistema apresenta primeiro as OS que possuem mecânico responsável.
7. O sistema ordena as OS de cada grupo pela data de entrada.
8. O sistema aplica a paginação.
9. O sistema apresenta a fila ao mecânico.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Não existem OS disponíveis | Retorna lista vazia. |
| A2 | Peça ou insumo necessário indisponível | A OS não aparece na fila. |
| A3 | OS deixa de estar em `AGUARDANDO_EXECUCAO` | A OS deixa de aparecer. |
| A4 | Usuário sem autorização | Impede o acesso. |
| A5 | Erro de consulta | Informa a falha na operação. |

**Saída**

- Lista paginada de Ordens de Serviço aptas para execução.

**Pós-condições**

- O mecânico consegue identificar as OS disponíveis.
- OS com mecânico responsável são apresentadas primeiro.
- Nenhum dado é alterado.

---

### 7.2 Refinamento Técnico

**Endpoint**

```http
GET /fila-atendimento
```

> **Decisão de projeto.** A fila é obtida diretamente das Ordens de Serviço, sem tabela própria:
> pertence à fila a OS com `status = AGUARDANDO_EXECUCAO` e `dataEntradaFila` preenchida. OS que
> já têm mecânico responsável **permanecem** na fila e aparecem **primeiro** — o mecânico precisa
> reencontrar o atendimento que já é dele, especialmente quando a OS voltou da aprovação de um
> orçamento complementar.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `os:ler`.

**Entrada** — query params, todos opcionais:

| Param | Tipo | Descrição |
|---|---|---|
| `pagina` / `tamanho` | inteiro | Paginação; `pagina` inicia em zero e `tamanho` tem máximo de 50. |

**Validações**

*Técnicas*

- Parâmetros de paginação válidos, respeitando o limite de `tamanho`.

*Negócio*

- Uma OS pertence à fila quando tem `status = AGUARDANDO_EXECUCAO` e `dataEntradaFila` preenchida.
- As peças e insumos necessários devem estar disponíveis ou reservados.
- OS com mecânico responsável permanecem na fila e aparecem antes das demais.
- A consulta não altera nenhum dado.

**Regras de ordenação**

Aplicadas nesta sequência:

1. OS com `mecanicoResponsavelId` preenchido.
2. OS sem mecânico responsável.
3. Dentro de cada grupo, `dataEntradaFila` crescente.
4. Em caso de empate, `id` crescente.

Exemplo: a OS A, com mecânico vinculado e entrada às 14:00, aparece antes da OS B, sem mecânico e
com entrada às 09:00, que por sua vez aparece antes da OS C, sem mecânico e com entrada às 10:00.

**Processamento**

1. Receber `pagina` e `tamanho` e validar a autenticação, a autorização e os parâmetros.
2. Consultar OS com `status = AGUARDANDO_EXECUCAO` e `dataEntradaFila` não nula.
3. Consultar as peças e insumos necessários das OS.
4. Validar a disponibilidade ou reserva dos itens e excluir as OS com recurso indisponível.
5. Priorizar as OS com mecânico responsável.
6. Ordenar por `dataEntradaFila` crescente, usando o `id` como desempate.
7. Aplicar a paginação e retornar a lista.

**Persistência**

- Operação somente de leitura sobre `ordem_servico`, com os relacionamentos de veículo, peças e
  insumos necessários e a disponibilidade ou reserva de estoque.
- Filtros: `status = AGUARDANDO_EXECUCAO`, `data_entrada_fila IS NOT NULL`, recursos necessários
  disponíveis ou reservados.
- Ordenação: `mecanico_responsavel_id IS NOT NULL` primeiro, depois `data_entrada_fila` e `id`.
- Não existe tabela `fila_atendimento` e nenhum dado é alterado.

**Saída da API**

```json
{
  "data": [
    {
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "veiculo": {
        "placa": "ABC1D23",
        "marca": "Marca",
        "modelo": "Modelo"
      },
      "status": "AGUARDANDO_EXECUCAO",
      "mecanicoResponsavelId": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29",
      "dataEntradaFila": "2026-08-21T14:00:00-03:00"
    },
    {
      "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
      "veiculo": {
        "placa": "DEF4G56",
        "marca": "Marca",
        "modelo": "Modelo"
      },
      "status": "AGUARDANDO_EXECUCAO",
      "mecanicoResponsavelId": null,
      "dataEntradaFila": "2026-08-21T09:00:00-03:00"
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 2,
  "totalPaginas": 1
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Consulta realizada, inclusive quando a lista estiver vazia. |
| `400` | Parâmetros de paginação inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:ler`. |

> Fila vazia é `200` com `"data": []`, nunca `404`.

**Dependências**

- `OrdemDeServicoRepository`.
- Consulta de peças e insumos necessários da OS.
- Consulta de disponibilidade ou reserva de estoque.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Ordenação com OS de mecânico antes das demais.
- Ordenação por `dataEntradaFila` crescente dentro de cada grupo.
- Desempate por `id` crescente.
- Normalização dos parâmetros de paginação.

*Integração*

- Retorna somente OS em `AGUARDANDO_EXECUCAO`.
- Ignora OS sem `dataEntradaFila`.
- OS com mecânico responsável permanece na fila e aparece primeiro.
- Não retorna OS com peça ou insumo necessário indisponível.
- Retorna lista vazia quando não houver OS aptas.
- Aplica a paginação corretamente.
- Parâmetros inválidos retornam `400`; sem token, `401`; sem escopo, `403`.
- A consulta não altera dados persistidos.

---

### 7.3 Checklist de Implementação

**Domínio**

- [ ] Garantir que a fila seja derivada de `ordem_servico`, sem dependência de tabela `fila_atendimento`
- [ ] Definir a regra de pertencimento à fila: `AGUARDANDO_EXECUCAO` com `data_entrada_fila` preenchida

**Caso de uso**

- [ ] Implementar `ConsultarFilaAtendimento`
- [ ] Filtrar `status = AGUARDANDO_EXECUCAO`
- [ ] Filtrar `data_entrada_fila IS NOT NULL`
- [ ] Excluir OS com recurso necessário indisponível
- [ ] Manter na consulta as OS com mecânico responsável
- [ ] Implementar a prioridade de leitura para `mecanicoResponsavelId` preenchido
- [ ] Ordenar por `dataEntradaFila` crescente dentro de cada grupo
- [ ] Implementar `id` crescente como critério de desempate
- [ ] Implementar a paginação
- [ ] Garantir que a consulta não altere dados persistidos

**Repositório**

- [ ] Criar ou ajustar `OrdemDeServicoRepository`
- [ ] Consultar peças e insumos necessários da OS

**Integrações**

- [ ] Criar ou ajustar a consulta ao estoque para disponibilidade e reserva

**Handler HTTP**

- [ ] Criar o handler para `GET /fila-atendimento`
- [ ] Validar `pagina` e `tamanho`
- [ ] Criar DTO de resposta com dados da OS e do veículo, incluindo `mecanicoResponsavelId`
- [ ] Implementar o envelope de resposta paginado
- [ ] Aplicar autenticação JWT e autorização na rota
- [ ] Retornar `400` para paginação inválida, `401` sem autenticação e `403` sem permissão

**Testes unitários**

- [ ] Somente OS em `AGUARDANDO_EXECUCAO`
- [ ] OS sem `dataEntradaFila` ignorada
- [ ] OS com mecânico permanece na fila
- [ ] OS com mecânico aparece antes das demais
- [ ] Ordenação por `dataEntradaFila`
- [ ] Desempate por `id`

**Testes de integração**

- [ ] Indisponibilidade de peças e de insumos
- [ ] Lista vazia
- [ ] Paginação
- [ ] A consulta não altera dados persistidos

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite

---
