---
documento: Refinamento de Requisitos — Monitorar Tempo Médio de Execução
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Monitorar Tempo Médio de Execução

Este documento detalha a tarefa Monitorar Tempo Médio de Execução dos Serviços do contexto de
Ordem de Serviço.

## 6 · Monitorar Tempo Médio de Execução

### 6.1 Refinamento de Produto

**Persona**

Gestor.

**Objetivo**

Acompanhar o tempo médio gasto na execução das Ordens de Serviço finalizadas.

**Problema**

A oficina precisa monitorar quanto tempo, em média, as Ordens de Serviço permanecem em execução,
para acompanhar a eficiência operacional. É o indicador exigido pelo enunciado do Tech Challenge.

**Pré-condições**

- Devem existir Ordens de Serviço com início e finalização registrados.
- O usuário deve possuir acesso ao indicador.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-38 | Permitir consultar o tempo médio de execução das Ordens de Serviço. |
| RF-OS-39 | Considerar apenas OS que possuam `dataInicioExecucao` e `dataFinalizacao`. |
| RF-OS-40 | Exibir a quantidade de Ordens de Serviço utilizadas no cálculo. |
| RF-OS-41 | Desconsiderar OS ainda em execução ou sem as datas necessárias. |
| RF-OS-42 | Calcular o indicador a partir dos dados já registrados na OS. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-23 | O cálculo deve ser realizado sem alterar os dados das Ordens de Serviço. |
| RNF-OS-24 | O indicador deve ser retornado de forma consistente para o mesmo conjunto de dados. |
| RNF-OS-25 | Apenas usuários autorizados devem acessar a informação. |

**Fluxo Principal**

1. O gestor acessa o monitoramento do tempo médio de execução.
2. O sistema identifica as Ordens de Serviço elegíveis para o cálculo.
3. O sistema calcula o tempo de execução de cada OS pela diferença entre a data de início e a de finalização.
4. O sistema calcula a média dos tempos encontrados.
5. O sistema apresenta o tempo médio e a quantidade de OS consideradas.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Não existem OS finalizadas com as datas necessárias | Retorna o indicador com nenhuma OS considerada. |
| A2 | Existem OS ainda em execução | São ignoradas no cálculo. |
| A3 | Existem registros incompletos, sem data de início ou de finalização | São ignorados no cálculo. |
| A4 | Usuário sem permissão | Impede a consulta do indicador. |

**Saída**

- Tempo médio de execução das Ordens de Serviço e quantidade de OS utilizadas no cálculo.

**Pós-condições**

- Nenhuma Ordem de Serviço é alterada.
- Nenhum dado de tempo médio precisa ser persistido: o indicador vem das datas já armazenadas.

---

### 6.2 Refinamento Técnico

**Endpoint**

```http
GET /ordens-servico/{osId}/tempo-execucao
GET /ordens-servico/tempos-execucao
```

O primeiro devolve o tempo de execução de uma OS específica; o segundo lista, de forma paginada,
os tempos das OS finalizadas, com o tempo médio do conjunto filtrado.

> **Decisão de projeto.** O tempo de execução é sempre **calculado** por
> `dataFinalizacao - dataInicioExecucao`. Não existem colunas `tempo_execucao` nem `tempo_medio`
> no banco: persistir o indicador criaria um segundo lugar para a mesma verdade, que ficaria
> defasado a cada correção de data.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `GESTOR`.
- Escopo: `os:ler`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da OS, na consulta individual. |
| Query | `dataInicio` | date | Opcional; início do período. |
| Query | `dataFim` | date | Opcional; fim do período. |
| Query | `page` / `size` | int | Paginação da listagem; `size` com máximo definido pela aplicação. |

**Validações**

*Técnicas*

- `osId` em formato UUID válido.
- Parâmetros de paginação válidos, respeitando o limite de `size`.
- Quando informados, `dataInicio` não pode ser posterior a `dataFim`.

*Negócio*

- Para o cálculo individual, a OS deve possuir `dataInicioExecucao` e `dataFinalizacao` preenchidas.
- OS ainda em execução não entram na consulta geral.
- Registros sem data de início ou de finalização são desconsiderados.
- O tempo médio considera **todas** as OS elegíveis do filtro aplicado, não apenas os itens da página atual.
- Nenhuma informação de tempo é persistida.

**Processamento**

*Consulta individual*

1. Buscar a OS pelo identificador e validar sua existência.
2. Validar se possui data de início e data de finalização.
3. Calcular o tempo de execução.
4. Retornar os dados da OS e o tempo calculado.

*Consulta geral*

1. Validar os parâmetros de paginação e os filtros de período.
2. Buscar as OS elegíveis e aplicar os filtros solicitados.
3. Calcular o tempo de execução de cada OS.
4. Calcular o tempo médio usando todas as OS elegíveis.
5. Paginar os registros da listagem.
6. Retornar o indicador geral e os itens da página.

**Persistência**

- Consulta: `ordem_servico` (`data_inicio_execucao`, `data_finalizacao`).
- Altera: nada.

**Saída da API**

Consulta individual:

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "dataInicioExecucao": "2026-08-10T09:00:00-03:00",
  "dataFinalizacao": "2026-08-10T12:30:00-03:00",
  "tempoExecucaoMinutos": 210
}
```

Consulta geral:

```json
{
  "tempoMedioExecucaoMinutos": 185,
  "totalElementos": 48,
  "pagina": 0,
  "tamanho": 20,
  "totalPaginas": 3,
  "data": [
    {
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "dataInicioExecucao": "2026-08-10T09:00:00-03:00",
      "dataFinalizacao": "2026-08-10T12:30:00-03:00",
      "tempoExecucaoMinutos": 210
    },
    {
      "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
      "dataInicioExecucao": "2026-08-11T08:15:00-03:00",
      "dataFinalizacao": "2026-08-11T10:45:00-03:00",
      "tempoExecucaoMinutos": 150
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Tempo calculado ou listagem retornada com sucesso. |
| `400` | OS sem dados suficientes para o cálculo; paginação ou filtros inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:ler`. |
| `404` | Ordem de Serviço não encontrada, na consulta individual. |

> Nenhuma OS elegível é `200` com `"data": []` e o indicador zerado, nunca `404`.

**Dependências**

- `OrdemDeServicoRepository`.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Cálculo do tempo de execução de uma OS.
- Cálculo da média com múltiplas OS.
- OS em execução é ignorada.
- OS sem data de início é ignorada.
- OS sem data de finalização é ignorada.
- Cenário sem OS elegíveis retorna indicador vazio, sem erro.

*Integração*

- Consulta individual retorna `200` com o tempo calculado.
- OS inexistente retorna `404`.
- OS sem as datas necessárias retorna `400` na consulta individual.
- Listagem retorna as OS elegíveis paginadas.
- Filtro por período funciona e período inválido retorna `400`.
- O tempo médio considera todas as OS do filtro, independentemente da página.
- Sem token retorna `401` e perfil sem escopo retorna `403`.

---

### 6.3 Checklist de Implementação

**Domínio**

- [ ] Definir o cálculo do tempo de execução como `dataFinalizacao - dataInicioExecucao`
- [ ] Garantir que OS ainda em execução sejam desconsideradas
- [ ] Garantir que registros sem data de início ou de finalização sejam desconsiderados
- [ ] Não criar coluna `tempo_execucao` no banco
- [ ] Não criar coluna `tempo_medio` no banco

**Caso de uso**

- [ ] Implementar `ConsultarTempoExecucaoDaOS`
- [ ] Implementar `ConsultarTempoMedioExecucao`
- [ ] Calcular a média dos tempos considerando todas as OS elegíveis do filtro
- [ ] Retornar a quantidade de OS utilizadas no cálculo
- [ ] Definir o comportamento quando nenhuma OS for elegível

**Repositório**

- [ ] Criar a consulta que recupera OS com `data_inicio_execucao` e `data_finalizacao` preenchidas
- [ ] Implementar o filtro opcional por período

**Handler HTTP**

- [ ] Implementar `GET /ordens-servico/{osId}/tempo-execucao`
- [ ] Implementar `GET /ordens-servico/tempos-execucao`
- [ ] Criar DTO/response do indicador e da listagem
- [ ] Implementar o envelope de resposta paginado
- [ ] Aplicar autenticação JWT e autorização por escopo nas rotas

**Validações**

- [ ] Validar os parâmetros de paginação
- [ ] Validar que `dataInicio` não é posterior a `dataFim`
- [ ] Retornar `400` para OS sem dados suficientes na consulta individual
- [ ] Retornar `404` para OS inexistente

**Testes unitários**

- [ ] Cálculo com uma OS
- [ ] Cálculo com múltiplas OS
- [ ] OS em execução ignorada
- [ ] OS sem data de início ignorada
- [ ] OS sem data de finalização ignorada
- [ ] Cenário sem OS elegíveis

**Testes de integração**

- [ ] Consulta individual e consulta geral
- [ ] Paginação
- [ ] Filtros de período, incluindo período inválido
- [ ] `401` sem autenticação e `403` sem permissão

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
