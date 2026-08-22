---
documento: Refinamento de Requisitos — Desativar Serviço
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Desativar Serviço

Este documento detalha a tarefa Remover ou Desativar Serviço do contexto de Serviços.

## 1 · Desativar Serviço

### 1.1 Refinamento de Produto

**Persona**

Gestor.

**Objetivo**

Retirar um serviço do catálogo de utilização sem comprometer o histórico das Ordens de Serviço
existentes.

**Problema**

Um serviço pode deixar de ser oferecido pela oficina, mas sua remoção física comprometeria
referências históricas em OS e orçamentos. Por isso a desativação é preferível quando o serviço já
foi utilizado.

**Pré-condições**

- O usuário deve possuir autorização para manter o catálogo de serviços.
- O serviço deve existir.
- Deve ser possível determinar se o serviço possui vínculos com Ordens de Serviço ou orçamentos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-SRV-01 | Permitir desativar um serviço. |
| RF-SRV-02 | Verificar se o serviço possui vínculos históricos. |
| RF-SRV-03 | Impedir a utilização de serviço desativado em novas operações. |
| RF-SRV-04 | Preservar o serviço quando houver necessidade de manter histórico. |
| RF-SRV-05 | Permitir identificar serviços ativos e inativos. |
| RF-SRV-06 | Permitir remoção física somente quando não houver vínculos que comprometam a integridade dos dados. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SRV-01 | A operação deve preservar a integridade referencial. |
| RNF-SRV-02 | A desativação deve ser persistida de forma consistente. |
| RNF-SRV-03 | O histórico de OS e orçamentos não deve ser perdido. |
| RNF-SRV-04 | Somente usuário autorizado poderá realizar a operação. |
| RNF-SRV-05 | A operação deve ser auditável. |

**Fluxo Principal**

1. O gestor consulta o catálogo de serviços.
2. O gestor seleciona o serviço.
3. O sistema verifica se o serviço existe.
4. O sistema verifica se existem vínculos com Ordens de Serviço ou orçamentos.
5. O gestor solicita a desativação.
6. O sistema altera a situação do serviço para inativa.
7. O sistema impede que o serviço seja utilizado em novos registros.
8. O sistema mantém os vínculos históricos existentes.
9. O sistema confirma a desativação.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Serviço não encontrado | Informa que o serviço não existe. |
| A2 | Serviço já inativo | Informa que o serviço já está desativado. |
| A3 | Existem vínculos históricos | Realiza a desativação lógica, preservando os registros existentes. |
| A4 | Tentativa de remoção física com vínculos | Impede a exclusão, para preservar a integridade e o histórico. |
| A5 | Usuário sem autorização | Impede a operação. |

**Saída**

- Serviço desativado e indisponível para novas utilizações, com seu histórico preservado.

**Pós-condições**

- O serviço fica marcado como inativo e não pode ser utilizado em novos registros que exijam
  serviços ativos.
- As Ordens de Serviço e orçamentos antigos continuam preservados.
- O catálogo mantém o histórico do serviço.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
PATCH /servicos/{servicoId}/desativar
```

> **Decisão de projeto.** A operação é uma transição de situação, não uma exclusão física, porque
> o serviço já pode estar associado a Ordens de Serviço históricas. A rota usa o verbo de negócio
> (`/desativar`) em vez de `DELETE /servicos/{servicoId}`, para deixar explícito que o registro
> permanece. Nos contextos de Cliente, Veículo e Peças & Insumos a mesma transição foi modelada
> como `DELETE` — os contextos precisam convergir (ver [`pontos-em-aberto.md`](pontos-em-aberto.md)).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `GESTOR`.
- Escopo: `servicos:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `servicoId` | uuid | Identificador do serviço. |

Não há corpo na requisição.

**Validações**

*Técnicas*

- `servicoId` em formato UUID válido.

*Negócio*

- O serviço deve existir.
- O serviço deve estar ativo.
- O serviço não deve ser excluído fisicamente quando houver vínculo histórico.
- O histórico de OS deve permanecer preservado.

**Regra de domínio**

```
ativo → desativar → inativo
```

- Serviço inativo não pode ser utilizado em novas OS nem em novos orçamentos.
- O histórico das OS permanece intacto.
- A consulta administrativa continua permitindo localizar o serviço inativo.

**Processamento**

1. Receber o identificador do serviço e identificar o usuário autenticado.
2. Buscar o serviço e validar existência, autorização e situação atual.
3. Executar `servico.desativar()`.
4. Registrar data e hora da desativação e o usuário responsável.
5. Persistir a alteração.
6. Retornar o serviço atualizado.

**Persistência**

- Consulta: `servico`, vínculos com Ordens de Serviço e orçamentos.
- Altera: `servico` (`ativo = false`, `data_desativacao`, `usuario_desativacao`).
- O registro continua armazenado no banco.

**Saída da API**

```json
{
  "id": "4b8e2c17-95a3-4f60-b7d1-6e0c58a3f942",
  "codigo": "SER-000001",
  "nome": "Troca de óleo",
  "ativo": false,
  "dataDesativacao": "2026-08-19T20:30:00-03:00",
  "usuarioDesativacao": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Serviço desativado, com o recurso atualizado no corpo. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `servicos:escrever`. |
| `404` | Serviço inexistente. |
| `409` | Serviço já inativo, ou outra regra de domínio impede a operação. |

**Dependências**

- `ServicoRepository`.
- Consulta de vínculos com Ordens de Serviço e orçamentos.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Desativa serviço ativo e a situação passa a inativa.
- Rejeita serviço inexistente.
- Rejeita serviço já inativo.

*Integração*

- `PATCH` válido retorna `200` com o serviço atualizado.
- O registro não é removido fisicamente do banco.
- Serviço inativo não pode ser usado em nova OS ou orçamento.
- Serviço inexistente retorna `404` e serviço já inativo retorna `409`.
- Perfil sem escopo retorna `403`.

*Regressão*

- O histórico das OS que usaram o serviço permanece intacto.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Definir a desativação lógica em vez de exclusão física
- [ ] Implementar o método de domínio `desativar()` em `Servico`
- [ ] Registrar data, hora e usuário responsável pela desativação
- [ ] Impedir a utilização de serviço inativo em novos orçamentos e Ordens de Serviço
- [ ] Preservar o serviço utilizado em OS antigas

**Caso de uso**

- [ ] Implementar `DesativarServico`
- [ ] Validar a existência do serviço
- [ ] Definir o comportamento ao desativar serviço já inativo
- [ ] Verificar vínculos com Ordens de Serviço e orçamentos

**Repositório**

- [ ] Criar ou ajustar `ServicoRepository` para a transição de situação
- [ ] Implementar a consulta de vínculos históricos

**Handler HTTP**

- [ ] Implementar `PATCH /servicos/{servicoId}/desativar`
- [ ] Implementar a validação do path param `servicoId`
- [ ] Criar DTO/response com o recurso atualizado
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Retornar `404` para serviço inexistente e `409` para serviço já inativo

**Testes unitários**

- [ ] Desativação válida
- [ ] Serviço inexistente
- [ ] Serviço já inativo
- [ ] Usuário sem autorização

**Testes de integração**

- [ ] `200` com a situação inativa persistida
- [ ] Registro não removido fisicamente
- [ ] Preservação do histórico de OS e orçamentos

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI, explicando a desativação lógica

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
