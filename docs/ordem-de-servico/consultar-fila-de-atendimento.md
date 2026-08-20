---
documento: Refinamento de Requisitos — Consultar Fila de Atendimento
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Fila de Atendimento

Este documento detalha a tarefa Consultar Fila de Atendimento do contexto de Ordem de Serviço.

## 4 · Consultar Fila de Atendimento

### 4.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Consultar as Ordens de Serviço disponíveis para execução.

**Problema**

O mecânico precisa visualizar as OS aptas para escolher qual serviço vai executar. Sem uma fila
confiável, a priorização vira ordem de chegada na memória de alguém — e OS sem peça disponível
entram em execução e travam o box.

**Pré-condições**

- O mecânico deve estar autenticado e autorizado.
- Devem existir OS cadastradas.
- As OS disponíveis devem estar com status `AGUARDANDO_EXECUCAO`.
- As peças e insumos necessários para a OS devem estar disponíveis ou reservados.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-21 | Permitir ao mecânico consultar as OS disponíveis. |
| RF-OS-22 | Apresentar as informações necessárias para identificar a OS. |
| RF-OS-23 | Retornar somente OS em `AGUARDANDO_EXECUCAO`. |
| RF-OS-24 | Retornar somente OS com peças e insumos necessários disponíveis para execução. |
| RF-OS-25 | Ordenar as OS pela data e hora de entrada na fila, exibindo a mais antiga primeiro. |
| RF-OS-26 | Permitir paginação da lista. |
| RF-OS-27 | Não apresentar OS canceladas, entregues ou já assumidas por outro mecânico. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-15 | A consulta deve refletir o estado atual das OS e do estoque. |
| RNF-OS-16 | A consulta não deve alterar os dados da OS nem do estoque. |
| RNF-OS-17 | O acesso deve exigir autenticação e autorização. |
| RNF-OS-18 | A ordenação deve ser determinística. |

**Fluxo Principal**

1. O mecânico acessa a fila de atendimento.
2. O sistema valida a autorização do mecânico.
3. O sistema consulta as OS em `AGUARDANDO_EXECUCAO`.
4. O sistema valida se as peças e insumos necessários estão disponíveis ou reservados.
5. O sistema ordena as OS pela data e hora de entrada na fila.
6. O sistema aplica a paginação.
7. O sistema apresenta a lista ao mecânico.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Não existem OS disponíveis | Retorna lista vazia. |
| A2 | Peça ou insumo necessário não está disponível | Não apresenta a OS na fila. |
| A3 | OS deixou de estar apta | Não a apresenta na fila. |
| A4 | Mecânico sem autorização | Impede o acesso. |
| A5 | Erro na consulta | Informa que não foi possível consultar a fila. |

**Saída**

- Lista de OS disponíveis para execução, ordenada da mais antiga para a mais recente.

**Pós-condições**

- O mecânico identifica as OS disponíveis.
- Nenhum dado é alterado durante a consulta.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
GET /fila-atendimento
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `os:ler`.
- O identificador do mecânico é obtido do token.

**Entrada** — query params, todos opcionais:

| Param | Tipo | Descrição |
|---|---|---|
| `page` / `size` | int | Paginação; `size` com máximo de 100. |

**Validações**

*Técnicas*

- Parâmetros de paginação válidos, com `size` dentro do limite.

*Negócio*

- Retornar somente OS em `AGUARDANDO_EXECUCAO`.
- Não retornar OS que já possuam mecânico responsável.
- Retornar somente OS cujas peças e insumos necessários estejam disponíveis ou reservados.

**Processamento**

1. Receber os parâmetros da consulta e identificar o mecânico autenticado.
2. Validar a autorização.
3. Consultar as OS com status `AGUARDANDO_EXECUCAO`.
4. Excluir as OS já associadas a um mecânico responsável.
5. Consultar a disponibilidade ou reserva das peças e insumos necessários.
6. Excluir as OS que possuam peça ou insumo pendente.
7. Ordenar por `dataEntradaFila` crescente, usando o identificador da OS como desempate.
8. Aplicar a paginação.
9. Montar e retornar a lista de OS.

**Persistência**

- Consulta: `ordem_servico`, itens de peça e insumo vinculados à OS, disponibilidade ou reserva de estoque.
- Filtros: `status = AGUARDANDO_EXECUCAO`, `mecanico_responsavel_id` nulo, peças e insumos disponíveis ou reservados.
- Ordenação: `data_entrada_fila`, depois `id`.
- Altera: nada.

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
      "dataEntradaFila": "2026-08-18T10:30:00-03:00"
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
| `200` | Consulta realizada, com ou sem OS disponível. |
| `400` | Parâmetros de paginação inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:ler`. |

> Fila vazia é `200` com `"data": []`, nunca `404`.

**Dependências**

- `OrdemDeServicoRepository`.
- Repositório ou serviço de peças e insumos.
- Repositório ou serviço de estoque.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Ordenação por `dataEntradaFila` crescente, com o identificador da OS como desempate.
- Normalização dos parâmetros de paginação.

*Integração*

- Retorna somente OS em `AGUARDANDO_EXECUCAO`.
- Não retorna OS com mecânico responsável.
- Não retorna OS com peça ou insumo necessário indisponível.
- Retorna OS com peças e insumos disponíveis ou reservados.
- Retorna primeiro a OS mais antiga na fila.
- Retorna lista vazia quando não houver OS disponível.
- Aplica a paginação corretamente.
- Parâmetros inválidos retornam `400`.
- Sem token retorna `401` e perfil sem escopo retorna `403`.
- A consulta não altera dados persistidos.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o campo `dataEntradaFila` na Ordem de Serviço
- [ ] Definir a regra de aptidão da OS para a fila: status, mecânico responsável e disponibilidade de itens

**Caso de uso**

- [ ] Implementar `ConsultarFilaAtendimento`
- [ ] Implementar a ordenação por data e hora de entrada na fila
- [ ] Implementar o identificador da OS como critério de desempate
- [ ] Implementar a paginação da consulta
- [ ] Garantir que a consulta não altera dados persistidos

**Repositório**

- [ ] Criar ou ajustar a consulta de OS por status `AGUARDANDO_EXECUCAO`
- [ ] Garantir que OS com mecânico responsável não apareçam na fila
- [ ] Implementar a consulta de peças e insumos vinculados à OS

**Integrações**

- [ ] Implementar a consulta de disponibilidade ou reserva de estoque
- [ ] Garantir que OS com peça ou insumo pendente não apareçam na fila

**Handler HTTP**

- [ ] Implementar `GET /fila-atendimento`
- [ ] Criar DTO/response com os dados da OS e do veículo
- [ ] Implementar o envelope de resposta paginado
- [ ] Aplicar autenticação JWT e autorização por escopo na rota

**Validações**

- [ ] Validar os query params `page` e `size`
- [ ] Retornar `400` para parâmetros inválidos
- [ ] Retornar `401` sem autenticação e `403` sem permissão

**Testes unitários**

- [ ] Ordenação por data de entrada na fila
- [ ] Desempate pelo identificador da OS
- [ ] Normalização dos parâmetros de paginação

**Testes de integração**

- [ ] Lista com OS aptas e lista vazia
- [ ] Exclusão de OS com mecânico responsável
- [ ] Exclusão de OS com peça ou insumo indisponível
- [ ] Paginação
- [ ] `400`, `401` e `403`
- [ ] A consulta não altera dados persistidos

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
