---
documento: Refinamento de Requisitos — Registrar Problema Relatado
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Registrar Problema Relatado

Este documento detalha a tarefa Registrar Problema Relatado do contexto de Ordem de Serviço.

> Esta tarefa **substitui** a antiga tarefa Iniciar Diagnóstico: o início do diagnóstico passou a
> ser consequência do registro do relato do cliente, e não uma operação separada.

## 1 · Registrar Problema Relatado

### 1.1 Refinamento de Produto

**Persona**

Mecânico, no atendimento.

**Objetivo**

Registrar na Ordem de Serviço o problema relatado pelo cliente e encaminhar o veículo para a etapa
de diagnóstico.

**Problema**

Ao receber o veículo, a oficina precisa registrar o motivo apresentado pelo cliente para orientar
o diagnóstico que será realizado depois.

**Pré-condições**

- A Ordem de Serviço deve existir.
- A OS deve estar com status `RECEBIDA`.
- O problema relatado ainda não deve ter sido registrado.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-82 | Permitir registrar a descrição do problema relatado pelo cliente. |
| RF-OS-83 | Permitir registrar observações adicionais. |
| RF-OS-84 | Vincular o relato à Ordem de Serviço. |
| RF-OS-85 | Iniciar a etapa de diagnóstico ao registrar o relato. |
| RF-OS-86 | Alterar o status da OS de `RECEBIDA` para `EM_DIAGNOSTICO`. |
| RF-OS-87 | Registrar a data e hora de início do diagnóstico. |
| RF-OS-88 | Impedir que um novo problema inicial seja registrado depois do início do diagnóstico. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-49 | Apenas usuários autorizados devem realizar o registro. |
| RNF-OS-50 | O relato e a mudança de status devem ser registrados na mesma operação. |
| RNF-OS-51 | A operação deve manter a consistência dos dados da Ordem de Serviço. |

**Fluxo Principal**

1. O mecânico acessa uma OS com status `RECEBIDA`.
2. O cliente informa o problema apresentado pelo veículo.
3. O mecânico registra a descrição e, opcionalmente, as observações.
4. O sistema valida as informações.
5. O sistema registra o problema relatado.
6. O sistema registra o início do diagnóstico.
7. O sistema altera a OS para `EM_DIAGNOSTICO`.
8. O sistema confirma o registro.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | Informa que a OS não existe. |
| A2 | Descrição do problema não informada | Impede o registro e informa o campo obrigatório. |
| A3 | Problema relatado já registrado | Impede novo registro. |
| A4 | OS não está mais com status `RECEBIDA` | Impede o registro. |
| A5 | Usuário sem permissão | Impede a operação. |

**Saída**

- Problema relatado registrado e Ordem de Serviço atualizada para `EM_DIAGNOSTICO`.

**Pós-condições**

- O problema informado pelo cliente está vinculado à OS.
- A data de início do diagnóstico está registrada.
- A OS encontra-se em `EM_DIAGNOSTICO`.
- O veículo está disponível para a continuidade do diagnóstico.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/problema-relatado
```

> **Decisão de projeto.** O início do diagnóstico deixou de ser um endpoint próprio
> (`PATCH /ordens-servico/{osId}/diagnostico/iniciar`) e passou a ser consequência do registro do
> relato: o mesmo POST grava o relato, grava `dataInicioDiagnostico` e muda o status para
> `EM_DIAGNOSTICO`. A alternativa — duas chamadas — permitia OS parada em `RECEBIDA` com relato
> registrado, um estado sem significado no negócio.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `os:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |
| Body | `descricao` | string | Obrigatória; não pode estar vazia. |
| Body | `observacoes` | string | Opcional. |

```json
{
  "descricao": "Veículo apresenta ruído ao frear",
  "observacoes": "Cliente relata que o problema começou há aproximadamente uma semana"
}
```

**Validações**

*Técnicas*

- `osId` em formato UUID válido.
- `descricao` obrigatória e não vazia.
- `observacoes` opcional.

*Negócio*

- A OS deve existir e estar com status `RECEBIDA`.
- Não pode existir problema relatado já registrado para a OS.
- Ao registrar o problema, o sistema grava `dataInicioDiagnostico` e altera o status para
  `EM_DIAGNOSTICO`, na mesma operação.

**Regra de domínio**

```
RECEBIDA → registrar problema relatado → EM_DIAGNOSTICO
```

**Processamento**

1. Buscar a Ordem de Serviço pelo identificador.
2. Validar se a OS existe.
3. Validar se a OS está `RECEBIDA`.
4. Validar se ainda não existe problema relatado.
5. Validar a descrição recebida.
6. Registrar o problema relatado.
7. Registrar a data e hora de início do diagnóstico.
8. Alterar o status da OS para `EM_DIAGNOSTICO`.
9. Persistir as alterações.
10. Retornar a OS atualizada.

**Persistência**

- Consulta: `ordem_servico`.
- Altera: `ordem_servico` (relato do cliente, `data_inicio_diagnostico`, `status`).

**Saída da API**

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "problemaRelatado": {
    "descricao": "Veículo apresenta ruído ao frear",
    "observacoes": "Cliente relata que o problema começou há aproximadamente uma semana"
  },
  "status": "EM_DIAGNOSTICO",
  "dataInicioDiagnostico": "2026-08-21T20:30:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Problema relatado registrado e diagnóstico iniciado. |
| `400` | Payload inválido ou descrição ausente. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:escrever`. |
| `404` | Ordem de Serviço não encontrada. |
| `409` | OS não está em `RECEBIDA`; problema relatado já registrado. |

**Dependências**

- `OrdemDeServicoRepository`.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Registra o problema relatado em uma OS `RECEBIDA`.
- Permite observações opcionais.
- Rejeita descrição vazia.
- Rejeita OS fora de `RECEBIDA`.
- Impede registro duplicado do problema relatado.
- Registra `dataInicioDiagnostico`.
- Altera o status de `RECEBIDA` para `EM_DIAGNOSTICO`.

*Integração*

- `POST` válido retorna `201` com a OS atualizada.
- OS inexistente retorna `404`.
- OS fora de `RECEBIDA` retorna `409`.
- Problema já registrado retorna `409`.
- Sem token retorna `401` e perfil sem escopo retorna `403`.
- Relato, data e mudança de status são persistidos na mesma operação.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Registrar o problema relatado na Ordem de Serviço
- [ ] Registrar `data_inicio_diagnostico` com a data e hora da operação
- [ ] Alterar o status da OS de `RECEBIDA` para `EM_DIAGNOSTICO`
- [ ] Garantir que relato, data e mudança de status sejam persistidos na mesma operação

**Caso de uso**

- [ ] Implementar `RegistrarProblemaRelatado`
- [ ] Validar que a OS exista e esteja com status `RECEBIDA`
- [ ] Validar que ainda não exista problema relatado registrado
- [ ] Validar que `descricao` seja obrigatória e não esteja vazia
- [ ] Permitir `observacoes` como campo opcional

**Repositório**

- [ ] Criar ou ajustar `OrdemDeServicoRepository` para o registro do relato e a transição de status

**Handler HTTP**

- [ ] Criar o handler para `POST /ordens-servico/{osId}/problema-relatado`
- [ ] Validar o path param e o payload
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação e autorização na rota
- [ ] Mapear os erros para os códigos HTTP definidos

**Validações**

- [ ] Retornar `404` para OS inexistente
- [ ] Retornar `409` para OS fora de `RECEBIDA`
- [ ] Retornar `409` quando já existir problema relatado

**Testes unitários**

- [ ] Registro válido
- [ ] Descrição vazia
- [ ] OS inexistente
- [ ] OS fora de `RECEBIDA`
- [ ] Problema já registrado
- [ ] Transição `RECEBIDA` para `EM_DIAGNOSTICO`
- [ ] Preenchimento de `data_inicio_diagnostico`

**Testes de integração**

- [ ] Endpoint retornando `201` com a OS atualizada
- [ ] Persistência do relato, da data e do status na mesma operação

**Documentação**

- [ ] Documentar o endpoint no OpenAPI/Swagger

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite da task

---
