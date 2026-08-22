---
documento: Refinamento de Requisitos — Registrar Problema Encontrado
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Registrar Problema Encontrado

Este documento detalha a tarefa Registrar Problema Encontrado do contexto de Ordem de Serviço.

## 11 · Registrar Problema Encontrado

### 11.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar os problemas identificados durante o diagnóstico ou durante a execução dos serviços,
associando-os ao orçamento correspondente.

**Problema**

Durante a análise ou a execução, o mecânico pode identificar problemas que precisam ser registrados
para depois definir os serviços, peças e insumos necessários e compor o orçamento ao cliente. O
problema é único como conceito, independentemente de ter sido identificado no diagnóstico ou na
execução: quem tem tipo é o orçamento, não o problema.

**Pré-condições**

- A Ordem de Serviço deve existir.
- A OS deve estar `EM_DIAGNOSTICO` ou `EM_EXECUCAO`.
- O usuário deve possuir permissão para registrar problemas.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-99 | Permitir registrar mais de um problema para a mesma OS. |
| RF-OS-100 | Permitir informar a descrição do problema. |
| RF-OS-101 | Permitir informar observações adicionais. |
| RF-OS-102 | Vincular todo problema a um orçamento. |
| RF-OS-103 | Definir automaticamente o tipo do orçamento de acordo com o status da OS. |
| RF-OS-104 | Utilizar o orçamento `PRINCIPAL` quando a OS estiver `EM_DIAGNOSTICO`. |
| RF-OS-105 | Criar o orçamento `PRINCIPAL` caso ainda não exista. |
| RF-OS-106 | Permitir que vários problemas sejam vinculados ao mesmo orçamento `PRINCIPAL`. |
| RF-OS-107 | Utilizar um orçamento `COMPLEMENTAR` com status `CRIADO` quando a OS estiver `EM_EXECUCAO`. |
| RF-OS-108 | Criar um novo orçamento `COMPLEMENTAR` quando não existir um complementar com status `CRIADO`. |
| RF-OS-109 | Permitir vários problemas no mesmo orçamento `COMPLEMENTAR` enquanto ele estiver `CRIADO`. |
| RF-OS-110 | Permitir que uma OS possua mais de um orçamento `COMPLEMENTAR` ao longo da execução. |
| RF-OS-111 | Não permitir adicionar novos problemas a orçamentos `APROVADO` ou `RECUSADO`. |
| RF-OS-112 | Manter o problema sem classificação própria de tipo. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-56 | Apenas usuários autorizados devem realizar o registro. |
| RNF-OS-57 | O tipo do orçamento não deve ser informado pelo usuário. |
| RNF-OS-58 | O registro do problema e seu vínculo com o orçamento devem ocorrer de forma consistente. |
| RNF-OS-59 | Em caso de criação automática do orçamento, problema, orçamento e vínculo devem ser registrados na mesma operação. |

**Fluxo Principal — problema identificado no diagnóstico**

1. O mecânico identifica um problema durante o diagnóstico.
2. O mecânico informa a descrição e, opcionalmente, observações.
3. O sistema identifica que a OS está `EM_DIAGNOSTICO`.
4. O sistema busca o orçamento `PRINCIPAL` da OS.
5. Não existindo, o sistema cria o orçamento `PRINCIPAL` com status `CRIADO`.
6. O sistema registra o problema e o vincula ao orçamento `PRINCIPAL`.
7. O sistema confirma o registro.

**Fluxo Principal — problema identificado durante a execução**

1. O mecânico identifica um novo problema durante a execução.
2. O mecânico informa a descrição e, opcionalmente, observações.
3. O sistema identifica que a OS está `EM_EXECUCAO`.
4. O sistema busca um orçamento `COMPLEMENTAR` com status `CRIADO`.
5. Existindo, o novo problema é vinculado a esse orçamento; não existindo, o sistema cria um novo
   orçamento `COMPLEMENTAR` com status `CRIADO`.
6. O sistema registra e vincula o problema ao orçamento correspondente.
7. O sistema confirma o registro.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | Informa que a OS não existe. |
| A2 | Descrição do problema não informada | Impede o registro. |
| A3 | OS não está `EM_DIAGNOSTICO` nem `EM_EXECUCAO` | Impede o registro. |
| A4 | Usuário sem permissão | Impede a operação. |
| A5 | Falha ao registrar o problema ou vinculá-lo ao orçamento | Não deixa alteração parcial salva. |

**Saída**

- Problema registrado e vinculado ao orçamento correspondente.

**Pós-condições**

- O problema está registrado e vinculado a um orçamento: ao `PRINCIPAL`, se identificado no
  diagnóstico; ao `COMPLEMENTAR` aberto, se identificado na execução.
- O orçamento permanece `CRIADO` e disponível para receber os serviços, peças e insumos
  relacionados aos problemas.

---

### 11.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/problemas
```

> **Decisão de projeto.** O problema **não** tem classificação entre principal e adicional: o tipo
> pertence ao orçamento (`PRINCIPAL` ou `COMPLEMENTAR`), e é deduzido do status da OS. Esta tarefa
> também é o ponto onde o orçamento nasce: registrar o primeiro problema do diagnóstico cria o
> orçamento `PRINCIPAL`, e registrar um problema durante a execução abre um `COMPLEMENTAR` se não
> houver nenhum aberto.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`.
- Escopo: `os:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |
| Body | `descricao` | string | Obrigatória; não pode estar vazia. |
| Body | `observacoes` | string | Opcional. |

```json
{
  "descricao": "Pastilhas de freio desgastadas",
  "observacoes": "Desgaste acima do limite recomendado"
}
```

**Validações**

*Técnicas*

- `osId` em formato UUID válido.
- `descricao` obrigatória e não vazia; `observacoes` opcional.

*Negócio*

- A OS deve existir e estar `EM_DIAGNOSTICO` ou `EM_EXECUCAO`.
- O problema deve ser vinculado a um orçamento, cujo tipo vem do status da OS.
- Um orçamento `CRIADO` pode receber novos problemas; `APROVADO` e `RECUSADO`, não.
- Uma OS possui no máximo um orçamento `PRINCIPAL`.
- Uma OS possui no máximo um orçamento `COMPLEMENTAR` com status `CRIADO` por vez, e pode ter
  vários complementares ao longo da execução.

**Regra de domínio**

```
OS EM_DIAGNOSTICO → orçamento PRINCIPAL   (cria se não existir) → vincula problema
OS EM_EXECUCAO    → orçamento COMPLEMENTAR CRIADO (cria se não houver) → vincula problema
```

**Processamento**

1. Buscar a OS pelo identificador e validar sua existência.
2. Validar se está `EM_DIAGNOSTICO` ou `EM_EXECUCAO`.
3. Validar a descrição do problema.
4. Determinar o tipo do orçamento pelo status da OS.
5. Buscar o orçamento aplicável e criar um com status `CRIADO` quando necessário.
6. Registrar o problema.
7. Vincular o problema ao orçamento.
8. Persistir as alterações na mesma operação.
9. Retornar o problema e o orçamento vinculado.

**Persistência**

- Consulta: `ordem_servico` (validação de status e vínculo).
- Altera: `orcamento` (consulta e, quando necessário, insert), `problema` (insert),
  `orcamento_problema` (insert do vínculo).

Campos de `orcamento`, quando criado: `id`, `ordem_servico_id`, `tipo`, `status = CRIADO`,
`created_at`, `updated_at`. Campos de `problema`: `id`, `descricao`, `observacoes`, `created_at`,
`updated_at`. Campos de `orcamento_problema`: `orcamento_id`, `problema_id`, `created_at`.

Regras de persistência: no máximo um orçamento `PRINCIPAL` por OS; no máximo um `COMPLEMENTAR`
com status `CRIADO` por OS; o problema é vinculado a um orçamento no momento da criação; problema,
eventual criação do orçamento e vínculo são persistidos na mesma transação, sem alteração parcial
em caso de erro.

**Saída da API**

```json
{
  "problemaId": "a3f60c81-7d24-4e59-b016-8c5f2b93ea47",
  "descricao": "Pastilhas de freio desgastadas",
  "observacoes": "Desgaste acima do limite recomendado",
  "orcamento": {
    "id": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
    "tipo": "PRINCIPAL",
    "status": "CRIADO"
  }
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Problema registrado com sucesso. |
| `400` | Payload inválido ou descrição vazia. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:escrever`. |
| `404` | Ordem de Serviço não encontrada. |
| `409` | Status atual da OS não permite registrar problema. |

**Dependências**

- `OrdemDeServicoRepository`.
- `OrcamentoRepository`.
- Repositório de problemas e do vínculo `orcamento_problema`.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Registra problema com a OS `EM_DIAGNOSTICO` e com a OS `EM_EXECUCAO`.
- Cria o orçamento `PRINCIPAL` quando ainda não existir e reutiliza o existente.
- Reutiliza o `COMPLEMENTAR` com status `CRIADO` e cria um novo quando não houver aberto.
- Garante no máximo um `PRINCIPAL` por OS e no máximo um `COMPLEMENTAR` `CRIADO` por vez.
- Rejeita descrição vazia.
- Rejeita orçamento `APROVADO` ou `RECUSADO`.

*Integração*

- `POST` válido retorna `201` com o problema e o orçamento vinculado.
- OS inexistente retorna `404` e OS em status não permitido retorna `409`.
- Criação de novo complementar depois de o anterior estar `APROVADO` ou `RECUSADO`.
- Rollback completo em caso de falha de persistência.

---

### 11.3 Checklist de Implementação

**Domínio**

- [ ] Determinar automaticamente o tipo do orçamento pelo status da OS
- [ ] Definir `PRINCIPAL` quando a OS estiver `EM_DIAGNOSTICO`
- [ ] Definir `COMPLEMENTAR` quando a OS estiver `EM_EXECUCAO`
- [ ] Garantir no máximo um orçamento `PRINCIPAL` por OS
- [ ] Garantir no máximo um `COMPLEMENTAR` com status `CRIADO` por OS
- [ ] Permitir múltiplos orçamentos `COMPLEMENTAR` por OS
- [ ] Impedir a inclusão de problemas em orçamento `APROVADO` ou `RECUSADO`
- [ ] Manter o problema sem classificação própria de tipo

**Caso de uso**

- [ ] Implementar `RegistrarProblemaEncontrado`
- [ ] Validar que a OS exista e esteja `EM_DIAGNOSTICO` ou `EM_EXECUCAO`
- [ ] Validar que `descricao` seja obrigatória e não esteja vazia
- [ ] Permitir `observacoes` como campo opcional
- [ ] Criar o orçamento aplicável com status `CRIADO` quando necessário
- [ ] Registrar o problema e vinculá-lo ao orçamento correspondente

**Repositório**

- [ ] Persistir o eventual novo orçamento
- [ ] Persistir o novo problema
- [ ] Persistir o vínculo em `orcamento_problema`

**Transação**

- [ ] Executar criação do orçamento, do problema e do vínculo na mesma transação
- [ ] Garantir rollback completo em caso de erro

**Handler HTTP**

- [ ] Criar o handler para `POST /ordens-servico/{osId}/problemas`
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação e autorização na rota
- [ ] Mapear os erros para os códigos HTTP definidos

**Testes unitários**

- [ ] Problema registrado durante `EM_DIAGNOSTICO`
- [ ] Criação automática do primeiro orçamento `PRINCIPAL`
- [ ] Reutilização do orçamento `PRINCIPAL`
- [ ] Problema registrado durante `EM_EXECUCAO`
- [ ] Criação de orçamento `COMPLEMENTAR`
- [ ] Reutilização de `COMPLEMENTAR` / `CRIADO`
- [ ] Criação de novo complementar depois de o anterior estar `APROVADO`
- [ ] Criação de novo complementar depois de o anterior estar `RECUSADO`
- [ ] Descrição vazia
- [ ] OS inexistente
- [ ] OS em status não permitido

**Testes de integração**

- [ ] Endpoint registrando o problema e devolvendo o orçamento vinculado
- [ ] Rollback em caso de falha de persistência

**Documentação**

- [ ] Documentar o endpoint no OpenAPI/Swagger

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite da task

---
