---
documento: Refinamento de Requisitos — Iniciar Diagnóstico
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Iniciar Diagnóstico

Este documento detalha a tarefa Iniciar Diagnóstico do contexto de Ordem de Serviço.

## 1 · Iniciar Diagnóstico

### 1.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Iniciar formalmente o processo de diagnóstico do veículo associado a uma Ordem de Serviço,
permitindo acompanhar corretamente o atendimento e sua situação.

**Problema**

A oficina precisa registrar quando o diagnóstico de um veículo começou. Sem esse registro, não
é possível acompanhar com segurança o andamento da Ordem de Serviço, identificar quando ela
entrou nessa etapa nem impedir transições de situação indevidas.

**Pré-condições**

- A Ordem de Serviço deve existir.
- A Ordem de Serviço deve estar na situação `RECEBIDA`.
- O cliente e o veículo devem estar vinculados à Ordem de Serviço.
- O mecânico deve estar autorizado a iniciar o diagnóstico.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-01 | Permitir ao mecânico iniciar o diagnóstico de uma Ordem de Serviço. |
| RF-OS-02 | Validar se a Ordem de Serviço está apta para iniciar o diagnóstico. |
| RF-OS-03 | Alterar a situação da Ordem de Serviço de `RECEBIDA` para `EM_DIAGNOSTICO`. |
| RF-OS-04 | Registrar a data e a hora de início do diagnóstico. |
| RF-OS-05 | Impedir que o diagnóstico da mesma Ordem de Serviço seja iniciado mais de uma vez. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-01 | Restringir a operação a usuário autenticado e autorizado. |
| RNF-OS-02 | Persistir a mudança de situação e a data e hora de início na mesma transação. |
| RNF-OS-03 | Evitar que alterações concorrentes deixem a Ordem de Serviço em situação inconsistente. |
| RNF-OS-04 | Preservar os vínculos da Ordem de Serviço com o cliente e o veículo. |

**Fluxo Principal**

1. O mecânico seleciona uma Ordem de Serviço com situação `RECEBIDA`.
2. O mecânico solicita o início do diagnóstico.
3. O sistema verifica se a Ordem de Serviço existe.
4. O sistema verifica se o mecânico está autorizado.
5. O sistema verifica se a Ordem de Serviço está na situação `RECEBIDA` e se o diagnóstico
   ainda não foi iniciado.
6. O sistema registra a data e a hora de início do diagnóstico.
7. O sistema altera a situação da Ordem de Serviço para `EM_DIAGNOSTICO`.
8. O sistema confirma que o diagnóstico foi iniciado.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | Informa que a Ordem de Serviço não existe e não registra nenhuma alteração. |
| A2 | Ordem de Serviço em situação diferente de `RECEBIDA` | Impede o início do diagnóstico e mantém a Ordem de Serviço inalterada. |
| A3 | Diagnóstico já iniciado | Impede um novo início e mantém a data e a hora registradas anteriormente. |
| A4 | Usuário não autenticado | Impede a operação e solicita autenticação. |
| A5 | Usuário sem autorização | Impede a operação e mantém a Ordem de Serviço inalterada. |
| A6 | A Ordem de Serviço foi alterada por outro usuário | Impede a sobrescrita da versão mais recente e solicita nova consulta. |

**Saída**

- Ordem de Serviço atualizada, com situação `EM_DIAGNOSTICO` e data e hora de início do
  diagnóstico; ou
- Indicação do motivo pelo qual o diagnóstico não pôde ser iniciado.

**Pós-condições**

- A Ordem de Serviço permanece vinculada ao mesmo cliente e veículo.
- A situação da Ordem de Serviço passa a ser `EM_DIAGNOSTICO`.
- A data e a hora de início do diagnóstico ficam registradas.
- A Ordem de Serviço fica apta para o próximo caso de uso: Registrar Diagnóstico.
- Nenhum dado de estoque, orçamento, cliente ou veículo é alterado.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
PATCH /api/v1/ordens-servico/{id}/diagnostico/iniciar
```

O endpoint executa a transição da Ordem de Serviço de `RECEBIDA` para `EM_DIAGNOSTICO`.

> **Decisão de projeto.** Foi adotada uma ação explícita na rota porque iniciar o diagnóstico é
> uma transição de domínio, e não uma alteração genérica de campos. A alternativa seria usar
> `PATCH /api/v1/ordens-servico/{id}` com uma nova situação, mas isso permitiria que o cliente da
> API tentasse controlar diretamente a máquina de estados.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `MECANICO`.
- Escopo: `os:escrever`.
- O identificador do usuário é obtido do token e não é recebido no corpo da requisição.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `id` | string | Identificador obrigatório da Ordem de Serviço. |
| Header | `If-Match` | inteiro | Versão atual da Ordem de Serviço, obrigatória para controle de concorrência. |

A operação não recebe corpo.

**Validações**

*Técnicas*

- `id` deve ser informado e possuir formato válido conforme o padrão de identificação da Ordem
  de Serviço.
- `If-Match` deve ser informado e conter uma versão válida.

*Negócio*

- A Ordem de Serviço deve existir.
- A situação atual deve ser `RECEBIDA`.
- A data de início do diagnóstico ainda não pode estar preenchida.
- Situações como `EM_DIAGNOSTICO`, `EM_EXECUCAO`, `FINALIZADA` e `ENTREGUE` não permitem iniciar
  o diagnóstico.
- O usuário deve possuir autorização para executar a operação.

A transição de domínio permitida é:

```text
RECEBIDA → EM_DIAGNOSTICO
```

**Processamento**

1. Validar os parâmetros e identificar o usuário autenticado.
2. Buscar a Ordem de Serviço pelo identificador.
3. Verificar a autorização do usuário.
4. Comparar o header `If-Match` com a versão atual da Ordem de Serviço.
5. Solicitar ao agregado a operação `IniciarDiagnostico`, fornecendo a data e a hora atuais.
6. No domínio, validar a situação atual e a ausência de início anterior.
7. Alterar a situação para `EM_DIAGNOSTICO` e registrar a data e a hora de início.
8. Persistir todas as alterações na mesma transação.
9. Retornar a representação atualizada da Ordem de Serviço.

A regra da transição pertence ao agregado de Ordem de Serviço e não ao handler HTTP.

**Persistência**

- Consulta: `ordem_servico` pelo identificador.
- Altera: `ordem_servico.status`, `ordem_servico.data_inicio_diagnostico` e
  `ordem_servico.version`.
- Não altera: vínculos com cliente e veículo, estoque, orçamento ou cadastro de serviços.
- A mudança de situação, a data e a hora e o incremento da versão são persistidos em uma única
  transação.

**Saída da API**

```json
{
  "id": "OS-2026-0123",
  "status": "EM_DIAGNOSTICO",
  "dataInicioDiagnostico": "2026-08-14T20:10:00-03:00",
  "version": 2
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Diagnóstico iniciado e Ordem de Serviço atualizada. |
| `400` | Identificador ou header `If-Match` ausente ou inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `os:escrever`. |
| `404` | Ordem de Serviço não encontrada. |
| `409` | A situação atual não permite iniciar o diagnóstico ou o diagnóstico já foi iniciado. |
| `412` | A versão informada em `If-Match` diverge da versão atual. |

**Dependências**

- `OrdemDeServicoRepository` para consultar e persistir a Ordem de Serviço.
- Middleware de autenticação JWT para identificar o usuário.
- Middleware ou política de autorização para validar o escopo `os:escrever`.
- Relógio da aplicação para fornecer a data e a hora de maneira testável.
- Gerenciador de transações.

O caso de uso não consulta os contextos de Estoque, Orçamento, Cliente ou Veículo.

**Testes**

*Unitários*

- Inicia o diagnóstico de uma Ordem de Serviço na situação `RECEBIDA`.
- Altera a situação para `EM_DIAGNOSTICO`.
- Registra a data e a hora de início fornecidas pelo relógio da aplicação.
- Rejeita o início quando o diagnóstico já foi iniciado.
- Rejeita o início nas situações `EM_DIAGNOSTICO`, `EM_EXECUCAO`, `FINALIZADA` e `ENTREGUE`.
- Mantém a situação e a data de início inalteradas quando a transição é rejeitada.

*Integração*

- Requisição válida retorna `200` com situação `EM_DIAGNOSTICO`, data de início e nova versão.
- Ordem de Serviço inexistente retorna `404`.
- Diagnóstico já iniciado retorna `409`.
- Ordem de Serviço em situação incompatível retorna `409`.
- Token ausente ou expirado retorna `401`.
- Usuário sem o escopo necessário retorna `403`.
- Versão divergente em `If-Match` retorna `412`.
- Falha durante a persistência não grava parcialmente a situação nem a data de início.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Implementar a transição `RECEBIDA` para `EM_DIAGNOSTICO` no agregado `OrdemDeServico`
- [ ] Implementar `OrdemDeServico.IniciarDiagnostico`
- [ ] Registrar a data e a hora de início do diagnóstico no agregado
- [ ] Rejeitar novo início e transições originadas de situações diferentes de `RECEBIDA`

**Caso de uso**

- [ ] Implementar `IniciarDiagnostico`
- [ ] Obter a data e a hora por uma abstração de relógio
- [ ] Retornar a Ordem de Serviço atualizada

**Repositório**

- [ ] Implementar a busca da Ordem de Serviço pelo identificador
- [ ] Persistir situação, data de início e nova versão atomicamente

**Handler HTTP**

- [ ] Implementar `PATCH /api/v1/ordens-servico/{id}/diagnostico/iniciar`
- [ ] Obter o identificador do usuário a partir do JWT
- [ ] Retornar a representação atualizada com status `200`
- [ ] Mapear erros para os códigos HTTP documentados

**Validações**

- [ ] Validar o formato do identificador da Ordem de Serviço
- [ ] Validar a presença e o formato do header `If-Match`
- [ ] Validar o escopo `os:escrever`

**Concorrência**

- [ ] Comparar `If-Match` com `ordem_servico.version`
- [ ] Retornar `412` sem sobrescrever a Ordem de Serviço quando a versão divergir

**Transação e idempotência**

- [ ] Persistir a situação, a data e a hora de início e a versão na mesma transação
- [ ] Garantir rollback integral em caso de falha

**Testes unitários**

- [ ] Início válido a partir de `RECEBIDA`
- [ ] Alteração para `EM_DIAGNOSTICO`
- [ ] Registro da data e da hora de início
- [ ] Rejeição de diagnóstico já iniciado
- [ ] Rejeição nas situações `EM_DIAGNOSTICO`, `EM_EXECUCAO`, `FINALIZADA` e `ENTREGUE`
- [ ] Ausência de alteração quando a transição é rejeitada

**Testes de integração**

- [ ] Resposta `200` com a Ordem de Serviço atualizada
- [ ] Resposta `404` para Ordem de Serviço inexistente
- [ ] Resposta `409` para diagnóstico já iniciado ou situação incompatível
- [ ] Resposta `401` para token ausente ou expirado
- [ ] Resposta `403` para usuário sem autorização
- [ ] Resposta `412` para versão divergente
- [ ] Rollback integral diante de falha de persistência

**Testes de concorrência**

- [ ] Duas solicitações com a mesma versão não iniciam o diagnóstico duas vezes

**Documentação**

- [ ] Documentar o endpoint e seus erros no Swagger/OpenAPI

**Review**

- [ ] Code Review aprovado

---
