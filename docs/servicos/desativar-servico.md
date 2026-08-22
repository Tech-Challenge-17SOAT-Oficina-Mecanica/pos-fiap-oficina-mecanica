---
documento: Refinamento de Requisitos — Desativar e Reativar Serviço
dono: A definir
versao: 0.2
atualizado_em: 2026-08-19
status: em revisao
---

# Refinamento de Requisitos — Desativar e Reativar Serviço

Este documento detalha a tarefa Desativar Serviço, e sua contrapartida Reativar Serviço, do
contexto de Serviços.

## 4 · Desativar e Reativar Serviço

### 4.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Retirar um serviço do catálogo de utilização sem comprometer o histórico das Ordens de Serviço
existentes.

**Problema**

Um serviço pode deixar de ser oferecido pela oficina, mas sua remoção física comprometeria
referências históricas em OS e orçamentos. Por isso a exclusão é sempre lógica, e o serviço pode
voltar ao catálogo por uma operação de reativação.

**Pré-condições**

- O usuário deve possuir autorização para manter o catálogo de serviços.
- O serviço deve existir.
- Deve ser possível determinar se o serviço possui vínculos com Ordens de Serviço ou orçamentos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-SRV-20 | Permitir desativar um serviço. |
| RF-SRV-21 | Verificar se o serviço possui vínculos históricos. |
| RF-SRV-22 | Impedir a utilização de serviço desativado em novas operações. |
| RF-SRV-23 | Preservar o serviço quando houver necessidade de manter histórico. |
| RF-SRV-24 | Permitir identificar serviços ativos e inativos. |
| RF-SRV-25 | Permitir reativar um serviço inativo. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SRV-16 | A operação deve preservar a integridade referencial. |
| RNF-SRV-17 | A desativação deve ser persistida de forma consistente. |
| RNF-SRV-18 | O histórico de OS e orçamentos não deve ser perdido. |
| RNF-SRV-19 | Somente usuário autorizado poderá realizar a operação. |
| RNF-SRV-20 | A operação deve ser auditável. |

**Fluxo Principal**

1. O mecânico consulta o catálogo de serviços.
2. O mecânico seleciona o serviço.
3. O sistema verifica se o serviço existe.
4. O sistema verifica se existem vínculos com Ordens de Serviço ou orçamentos.
5. O mecânico solicita a desativação.
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
| A4 | Reativação de serviço já ativo | Informa que o serviço já está ativo e não altera nada. |
| A6 | Reativação com nome já usado por outro serviço ativo | Impede a reativação, porque a unicidade do nome vale entre os ativos. |
| A5 | Usuário sem autorização | Impede a operação. |

**Saída**

- Serviço desativado e indisponível para novas utilizações, com seu histórico preservado.

**Pós-condições**

- O serviço fica marcado como inativo e não pode ser utilizado em novos registros que exijam
  serviços ativos.
- As Ordens de Serviço e orçamentos antigos continuam preservados.
- O catálogo mantém o histórico do serviço.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
DELETE /servicos/{servicoId}
POST   /servicos/{servicoId}/reativacao
```

O `DELETE` inativa o serviço; o `POST /reativacao` traz o serviço de volta ao catálogo.

> **Decisão de projeto.** A operação usa `DELETE`, e não `PATCH /servicos/{servicoId}/desativar`.
> A exclusão continua sendo **lógica** — o registro permanece no banco, com `ativo = false` —, e o
> verbo passa a ser o mesmo de Cliente, Veículo, Peças e Insumos (D-20). A documentação do
> endpoint precisa deixar explícito que nada é removido fisicamente, já que o verbo, sozinho,
> sugere o contrário.

> **Decisão de projeto.** Existe **reativação**, espelhando o que Cliente e Veículo já fazem. Sem
> ela, um serviço desativado por engano só voltaria pelo banco.

> **Decisão de projeto.** A **remoção física foi retirada do MVP**. Ela não tinha endpoint nem
> regra definida, e a exclusão lógica cobre o caso de uso da oficina.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `servicos:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `servicoId` | uuid | Identificador do serviço. |

Não há corpo em nenhuma das duas requisições.

**Validações**

*Técnicas*

- `servicoId` em formato UUID válido.

*Negócio — desativação*

- O serviço deve existir.
- O serviço deve estar ativo.
- O registro nunca é removido fisicamente.
- O histórico de OS deve permanecer preservado.

*Negócio — reativação*

- O serviço deve existir.
- O serviço deve estar inativo.
- Não pode existir outro serviço **ativo** com o mesmo nome normalizado.

**Regra de domínio**

```
ativo → DELETE → inativo → POST /reativacao → ativo
```

- Serviço inativo não pode ser utilizado em novas OS nem em novos orçamentos.
- O histórico das OS permanece intacto.
- A consulta administrativa continua permitindo localizar o serviço inativo.

**Processamento**

*Desativação*

1. Receber o identificador do serviço e identificar o usuário autenticado.
2. Buscar o serviço e validar existência, autorização e situação atual.
3. Executar `servico.desativar()`.
4. Registrar data e hora da desativação e o usuário responsável.
5. Persistir a alteração.
6. Retornar o serviço atualizado.

*Reativação*

1. Receber o identificador do serviço e identificar o usuário autenticado.
2. Buscar o serviço e validar existência, autorização e situação atual.
3. Verificar que nenhum outro serviço ativo usa o mesmo nome normalizado.
4. Executar `servico.reativar()`.
5. Limpar `data_desativacao` e `usuario_desativacao`.
6. Persistir a alteração.
7. Retornar o serviço atualizado.

**Persistência**

- Consulta: `servico`, vínculos com Ordens de Serviço e orçamentos.
- Altera, na desativação: `servico` (`ativo = false`, `data_desativacao`, `usuario_desativacao`).
- Altera, na reativação: `servico` (`ativo = true`, `data_desativacao` e `usuario_desativacao`
  nulos).
- O registro continua armazenado no banco nas duas operações.
- O índice parcial `UNIQUE (nome_normalizado) WHERE ativo = true` é o que garante a unicidade
  entre ativos, tanto no cadastro quanto na reativação.

**Saída da API**

`DELETE` — `200`:

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

`POST /reativacao` — `200`:

```json
{
  "id": "4b8e2c17-95a3-4f60-b7d1-6e0c58a3f942",
  "codigo": "SER-000001",
  "nome": "Troca de óleo",
  "ativo": true,
  "dataDesativacao": null,
  "usuarioDesativacao": null
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Serviço desativado ou reativado, com o recurso atualizado no corpo. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `servicos:escrever`. |
| `404` | Serviço inexistente. |
| `409` | Serviço já inativo no `DELETE`, já ativo na reativação, ou nome já usado por outro serviço ativo. |

**Dependências**

- `ServicoRepository`.
- Consulta de vínculos com Ordens de Serviço e orçamentos.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Desativa serviço ativo e `ativo` passa a `false`.
- Reativa serviço inativo e `ativo` passa a `true`.
- Rejeita serviço inexistente.
- Rejeita desativação de serviço já inativo.
- Rejeita reativação de serviço já ativo.
- Rejeita reativação quando outro serviço ativo usa o mesmo nome normalizado.

*Integração*

- `DELETE` válido retorna `200` com o serviço atualizado e `ativo` em `false`.
- `POST /reativacao` válido retorna `200` com `ativo` em `true`.
- O registro não é removido fisicamente do banco.
- Serviço inativo não pode ser usado em nova OS ou orçamento.
- Serviço inexistente retorna `404`; serviço já inativo no `DELETE` retorna `409`.
- Reativação com nome já usado por outro serviço ativo retorna `409`.
- Serviço reativado volta a aparecer na listagem padrão.
- Perfil sem escopo retorna `403`.

*Regressão*

- O histórico das OS que usaram o serviço permanece intacto.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Definir a desativação lógica em vez de exclusão física
- [ ] Implementar os métodos de domínio `desativar()` e `reativar()` em `Servico`
- [ ] Registrar data, hora e usuário responsável pela desativação
- [ ] Impedir a utilização de serviço inativo em novos orçamentos e Ordens de Serviço
- [ ] Preservar o serviço utilizado em OS antigas

**Caso de uso**

- [ ] Implementar `DesativarServico`
- [ ] Implementar `ReativarServico`
- [ ] Validar a existência do serviço
- [ ] Retornar `409` ao desativar serviço já inativo e ao reativar serviço já ativo
- [ ] Validar, na reativação, a unicidade do nome normalizado entre ativos
- [ ] Verificar vínculos com Ordens de Serviço e orçamentos

**Repositório**

- [ ] Criar ou ajustar `ServicoRepository` para a transição de situação
- [ ] Implementar a consulta de vínculos históricos

**Handler HTTP**

- [ ] Implementar `DELETE /servicos/{servicoId}`
- [ ] Implementar `POST /servicos/{servicoId}/reativacao`
- [ ] Implementar a validação do path param `servicoId`
- [ ] Criar DTO/response com o recurso atualizado
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Retornar `404` para serviço inexistente e `409` para serviço já inativo

**Testes unitários**

- [ ] Desativação válida
- [ ] Reativação válida
- [ ] Serviço inexistente
- [ ] Serviço já inativo no `DELETE`
- [ ] Serviço já ativo na reativação
- [ ] Reativação com nome já usado por outro serviço ativo
- [ ] Usuário sem autorização

**Testes de integração**

- [ ] `DELETE` retornando `200` com `ativo` em `false` persistido
- [ ] `POST /reativacao` retornando `200` com `ativo` em `true` persistido
- [ ] Registro não removido fisicamente
- [ ] Serviço reativado voltando à listagem padrão
- [ ] Preservação do histórico de OS e orçamentos

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI, explicando que o `DELETE` é exclusão lógica

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
