---
documento: Refinamento de Requisitos — Deletar Insumo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Deletar Insumo

Este documento detalha a tarefa Deletar Insumo do contexto de Peças & Insumos.

## 13 · Deletar Insumo

### 13.1 Refinamento de Produto

**Persona**

Gestor.

**Objetivo**

Retirar um insumo do catálogo de itens disponíveis para novos atendimentos, preservando seu
histórico de utilização e evitando que um insumo descontinuado seja usado em novas Ordens de Serviço.

**Problema**

A oficina precisa controlar os insumos disponíveis para uso nos serviços. Quando um insumo deixa
de ser utilizado ou comercializado, ele não deve continuar aparecendo como opção para novos
atendimentos. Entretanto, o histórico de consumo precisa ser preservado para manter a
rastreabilidade das Ordens de Serviço e do estoque.

**Pré-condições**

- Deve existir um insumo cadastrado.
- O insumo deve estar ativo.
- O usuário deve possuir autorização para realizar a operação.
- O insumo não pode ser removido fisicamente caso possua histórico de utilização.
- Deve ser considerada a existência de saldo em estoque, conforme a regra definida pela oficina.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-89 | Permitir ao usuário autorizado desativar um insumo. |
| RF-EST-90 | Alterar a situação do insumo de ativo para inativo. |
| RF-EST-91 | Impedir que um insumo inativo seja utilizado em novos orçamentos. |
| RF-EST-92 | Impedir que um insumo inativo seja adicionado a novas Ordens de Serviço. |
| RF-EST-93 | Manter o insumo disponível para consulta histórica. |
| RF-EST-94 | Preservar as informações de consumo do insumo em Ordens de Serviço anteriores. |
| RF-EST-95 | Registrar a data da desativação. |
| RF-EST-96 | Registrar o usuário responsável pela desativação. |
| RF-EST-97 | Permitir identificar que o insumo não está mais disponível para novos atendimentos. |
| RF-EST-98 | Não remover fisicamente o insumo do histórico do sistema. |
| RF-EST-99 | Preservar a unidade de medida e as informações históricas relacionadas ao insumo. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-67 | A desativação deve ser persistida de forma consistente. |
| RNF-EST-68 | Somente usuários autorizados devem poder desativar insumos. |
| RNF-EST-69 | O histórico de utilização do insumo não pode ser alterado pela desativação. |
| RNF-EST-70 | A operação deve manter a rastreabilidade do insumo. |
| RNF-EST-71 | A desativação não deve modificar Ordens de Serviço já registradas. |
| RNF-EST-72 | A operação deve considerar corretamente quantidades de estoque, inclusive valores decimais. |
| RNF-EST-73 | A operação deve ter comportamento consistente em caso de erro ou concorrência. |

**Fluxo Principal**

1. O gestor acessa o cadastro de insumos.
2. O sistema apresenta os insumos cadastrados e suas situações.
3. O gestor seleciona um insumo ativo.
4. O gestor solicita a desativação do insumo.
5. O sistema apresenta uma confirmação da operação.
6. O gestor confirma a desativação.
7. O sistema valida se o insumo pode ser desativado.
8. O sistema altera a situação do insumo para inativa.
9. O sistema registra a data e hora da desativação e o usuário responsável.
10. O sistema mantém o insumo no histórico.
11. O sistema confirma que o insumo foi desativado.
12. O insumo deixa de estar disponível para novos orçamentos e atendimentos.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Insumo não encontrado | Informa que o insumo não existe. |
| A2 | Insumo já inativo | Informa que o insumo já está desativado. |
| A3 | Usuário sem autorização | Impede a operação. |
| A4 | Insumo utilizado em OS anteriores | Mantém o registro e realiza apenas a desativação lógica. |
| A5 | Insumo com saldo em estoque | Segue a regra de negócio definida para insumos com estoque disponível. |
| A6 | Insumo vinculado a orçamento pendente | Comportamento a definir: permitir a desativação ou bloquear a operação. |
| A7 | Quantidade decimal em estoque | Considera corretamente o saldo fracionado na validação. |
| A8 | Erro na persistência | Não considera o insumo desativado até que a alteração seja persistida com sucesso. |

**Saída**

- Insumo inativo, permanecendo registrado no sistema e preservando seu histórico de utilização e movimentação.

**Pós-condições**

- O insumo passa de ativo para inativo e deixa de ser utilizável em novos orçamentos e Ordens de Serviço.
- O histórico das Ordens de Serviço que utilizaram o insumo permanece preservado.
- O insumo continua disponível para consultas administrativas e históricas.
- Nenhum registro histórico do insumo é removido.
- O saldo de estoque permanece conforme a regra definida para o processo de desativação.

---

### 13.2 Refinamento Técnico

**Endpoint**

```http
DELETE /estoque/insumos/{insumoId}
```

> **Decisão de projeto.** Vale a mesma decisão de [`deletar-peca.md`](deletar-peca.md): exclusão
> lógica, com `200` e o recurso atualizado no corpo em vez de `204` sem corpo.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `estoque:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `insumoId` | uuid | Identificador do insumo. |

Não há corpo na requisição.

**Validações**

*Técnicas*

- `insumoId` em formato UUID válido.

*Negócio*

- O insumo deve existir.
- O insumo deve estar ativo.
- O insumo não pode ser removido fisicamente caso possua histórico.
- A validação de estoque deve considerar saldo fracionado (por exemplo, 5,5 L).
- Comportamento a definir quando houver saldo em estoque ou orçamento pendente.

**Regra de domínio**

```
ativo → desativar → inativo
```

O insumo permanece armazenado para preservar seu histórico.

**Processamento**

1. Receber o identificador do insumo e identificar o usuário autenticado.
2. Buscar o insumo.
3. Validar existência, autorização e situação atual.
4. Verificar a regra relacionada ao estoque, considerando o saldo decimal.
5. Executar `insumo.desativar()`.
6. Registrar data e hora da desativação e o usuário responsável.
7. Persistir a alteração.
8. Retornar o resultado da operação.

**Persistência**

- Consulta: `item_estoque`.
- Altera: `item_estoque` (`ativo = false`, `data_desativacao`, `usuario_desativacao`).
- Não altera: `saldo_fisico`, histórico de Ordens de Serviço.
- O registro não é removido fisicamente do banco.

**Saída da API**

```json
{
  "id": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
  "codigo": "OLEO-001",
  "nome": "Óleo 5W30",
  "unidadeMedida": "L",
  "ativo": false,
  "dataDesativacao": "2026-08-18T09:40:00-03:00",
  "usuarioDesativacao": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Insumo desativado, com o recurso atualizado no corpo. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `estoque:escrever`. |
| `404` | Insumo inexistente. |
| `409` | Insumo já inativo, ou desativação impedida por regra de negócio. |

**Dependências**

- `ItemEstoqueRepository`.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Desativa insumo ativo e a situação passa a inativa.
- Não desativa insumo inexistente.
- Não desativa insumo já inativo.
- Valida a regra relacionada ao estoque, com saldo decimal.

*Integração*

- `DELETE` válido retorna `200` com o insumo atualizado.
- O registro não é apagado fisicamente do banco.
- O histórico de utilização é preservado.
- Insumo inativo não pode ser usado em novos orçamentos.
- Insumo inexistente retorna `404` e insumo já inativo retorna `409`.
- Perfil sem escopo retorna `403`.

---

### 13.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `desativar()` na entidade `Insumo`
- [ ] Definir a exclusão lógica do insumo, sem remoção física
- [ ] Registrar data, hora e usuário responsável pela desativação
- [ ] Garantir que insumo inativo não possa ser usado em novos orçamentos e Ordens de Serviço

**Caso de uso**

- [ ] Implementar `DesativarInsumo`
- [ ] Validar que o insumo existe e está ativo
- [ ] Validar a quantidade disponível em estoque, considerando saldo decimal
- [ ] Definir o comportamento quando houver saldo em estoque
- [ ] Verificar se o insumo já foi utilizado em alguma Ordem de Serviço

**Repositório**

- [ ] Ajustar `ItemEstoqueRepository` para a atualização de situação
- [ ] Implementar a verificação de utilização do insumo em Ordens de Serviço

**Handler HTTP**

- [ ] Implementar `DELETE /estoque/insumos/{insumoId}`
- [ ] Validar o path param `insumoId`
- [ ] Criar DTO/response com o recurso atualizado
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Testes unitários**

- [ ] Desativação válida
- [ ] Insumo inexistente
- [ ] Insumo já inativo
- [ ] Insumo com saldo em estoque, inclusive decimal
- [ ] Insumo utilizado em histórico

**Testes de integração**

- [ ] `200` com a situação inativa persistida
- [ ] `404` para insumo inexistente e `409` para insumo já inativo
- [ ] `403` para perfil sem permissão
- [ ] Preservação do histórico de utilização

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI, explicando a exclusão lógica

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
