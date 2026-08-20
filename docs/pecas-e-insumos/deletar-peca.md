---
documento: Refinamento de Requisitos — Deletar Peça
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Deletar Peça

Este documento detalha a tarefa Deletar Peça do contexto de Peças & Insumos.

## 11 · Deletar Peça

### 11.1 Refinamento de Produto

**Persona**

Gestor.

**Objetivo**

Retirar uma peça do catálogo de peças disponíveis para novos atendimentos, preservando seu
histórico de utilização nas Ordens de Serviço e evitando inconsistências no controle de estoque.

**Problema**

A oficina precisa impedir que peças que não estão mais disponíveis sejam selecionadas em novos
orçamentos e Ordens de Serviço. Ao mesmo tempo, não pode apagar fisicamente peças que já foram
utilizadas em atendimentos anteriores, pois isso comprometeria o histórico da oficina.

**Pré-condições**

- Deve existir uma peça cadastrada.
- A peça deve estar ativa.
- O usuário deve possuir autorização para realizar a operação.
- A peça não pode ser removida fisicamente caso possua histórico de utilização.
- Deve ser considerada a existência de saldo em estoque, conforme a regra definida pela oficina.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-70 | Permitir ao usuário autorizado desativar uma peça. |
| RF-EST-71 | Alterar a situação da peça de ativa para inativa. |
| RF-EST-72 | Impedir que uma peça inativa seja utilizada em novos orçamentos. |
| RF-EST-73 | Impedir que uma peça inativa seja adicionada a novas Ordens de Serviço. |
| RF-EST-74 | Manter a peça disponível para consulta histórica. |
| RF-EST-75 | Preservar as informações da peça utilizadas em Ordens de Serviço anteriores. |
| RF-EST-76 | Registrar a data da desativação. |
| RF-EST-77 | Registrar o usuário responsável pela desativação. |
| RF-EST-78 | Permitir identificar que a peça não está mais disponível para novos atendimentos. |
| RF-EST-79 | Não remover fisicamente a peça do histórico do sistema. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-55 | A desativação deve ser persistida de forma consistente. |
| RNF-EST-56 | Somente usuários autorizados devem poder desativar peças. |
| RNF-EST-57 | O histórico das Ordens de Serviço não pode ser alterado pela desativação. |
| RNF-EST-58 | A operação deve manter a rastreabilidade da peça. |
| RNF-EST-59 | A desativação não deve alterar valores ou itens de Ordens de Serviço já registradas. |
| RNF-EST-60 | A operação deve ter comportamento consistente em caso de erro ou concorrência. |

**Fluxo Principal**

1. O gestor acessa o cadastro de peças.
2. O sistema apresenta as peças cadastradas e suas situações.
3. O gestor seleciona uma peça ativa.
4. O gestor solicita a desativação da peça.
5. O sistema apresenta uma confirmação da operação.
6. O gestor confirma a desativação.
7. O sistema valida se a peça pode ser desativada.
8. O sistema altera a situação da peça para inativa.
9. O sistema registra a data e hora da desativação e o usuário responsável.
10. O sistema mantém a peça no histórico.
11. O sistema confirma que a peça foi desativada.
12. A peça deixa de estar disponível para novos orçamentos e atendimentos.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Peça não encontrada | Informa que a peça não existe. |
| A2 | Peça já inativa | Informa que a peça já está desativada. |
| A3 | Usuário sem autorização | Impede a operação. |
| A4 | Peça utilizada em OS anteriores | Mantém o registro e realiza apenas a desativação lógica. |
| A5 | Peça com saldo em estoque | Segue a regra de negócio definida para peças com estoque disponível. |
| A6 | Peça vinculada a orçamento pendente | Comportamento a definir: permitir a desativação ou bloquear a operação. |
| A7 | Erro na persistência | Não considera a peça desativada até que a alteração seja persistida com sucesso. |

**Saída**

- Peça inativa, permanecendo registrada no sistema e preservando seu histórico de utilização.

**Pós-condições**

- A peça passa de ativa para inativa e deixa de ser utilizável em novos orçamentos e Ordens de Serviço.
- O histórico das Ordens de Serviço que utilizaram a peça permanece preservado.
- A peça continua disponível para consultas administrativas e históricas.
- Nenhum registro histórico da peça é removido.
- O saldo de estoque permanece conforme a regra definida para o processo de desativação.

---

### 11.2 Refinamento Técnico

**Endpoint**

```http
DELETE /estoque/pecas/{pecaId}
```

> **Decisão de projeto.** O `DELETE` executa exclusão **lógica**: a peça permanece no banco e
> apenas deixa de estar disponível para novos atendimentos, porque o registro é referenciado por
> Ordens de Serviço antigas. A resposta devolve `200` com o recurso atualizado em vez de `204`
> sem corpo, para que a mudança de situação fique visível na demonstração do Swagger. Mesma
> decisão adotada em [`deletar-cliente.md`](../cliente/deletar-cliente.md) e
> [`deletar-veiculo.md`](../veiculo/deletar-veiculo.md).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `estoque:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `pecaId` | uuid | Identificador da peça. |

Não há corpo na requisição.

**Validações**

*Técnicas*

- `pecaId` em formato UUID válido.

*Negócio*

- A peça deve existir.
- A peça deve estar ativa.
- A peça não pode ser removida fisicamente caso possua histórico.
- Comportamento a definir quando houver saldo em estoque ou orçamento pendente.
- A peça não pode ser utilizada em novos orçamentos após a operação.

**Regra de domínio**

```
ativa → desativar → inativa
```

A peça continua armazenada para preservar as referências das Ordens de Serviço e do histórico.

**Processamento**

1. Receber o identificador da peça e identificar o usuário autenticado.
2. Buscar a peça.
3. Validar existência, autorização e situação atual.
4. Verificar as restrições relacionadas ao estoque.
5. Executar `peca.desativar()`.
6. Registrar data e hora da desativação e o usuário responsável.
7. Persistir a alteração.
8. Retornar o resultado da operação.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`.
- Altera: `item_estoque` (`ativo = false`, `data_desativacao`, `usuario_desativacao`).
- Não altera: `saldo_fisico`, `saldo_reservado`, histórico de Ordens de Serviço.
- O registro não é removido fisicamente do banco.

**Saída da API**

```json
{
  "id": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
  "codigo": "PEC-000001",
  "nome": "Pastilha de freio",
  "ativo": false,
  "dataDesativacao": "2026-08-18T09:30:00-03:00",
  "usuarioDesativacao": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Peça desativada, com o recurso atualizado no corpo. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `estoque:escrever`. |
| `404` | Peça inexistente. |
| `409` | Peça já inativa, ou desativação impedida por regra de negócio. |

**Dependências**

- `ItemEstoqueRepository`.
- `ReservaEstoqueRepository` (verificação de reserva ativa).
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Desativa peça ativa e a situação passa a inativa.
- Não desativa peça inexistente.
- Não desativa peça já inativa.
- Valida a regra relacionada ao estoque.

*Integração*

- `DELETE` válido retorna `200` com a peça atualizada.
- O registro não é apagado fisicamente do banco.
- Peça utilizada em OS anteriores é preservada.
- Peça inativa não pode ser usada em novos orçamentos.
- Peça inexistente retorna `404` e peça já inativa retorna `409`.
- Perfil sem escopo retorna `403`.

*Regressão*

- O histórico das Ordens de Serviço permanece intacto após a desativação.

---

### 11.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `desativar()` na entidade `Peca`
- [ ] Definir a exclusão lógica da peça, sem remoção física
- [ ] Registrar data, hora e usuário responsável pela desativação
- [ ] Garantir que peça inativa não possa ser adicionada a novos orçamentos e Ordens de Serviço

**Caso de uso**

- [ ] Implementar `DesativarPeca`
- [ ] Validar que a peça existe e está ativa
- [ ] Definir o comportamento quando houver saldo em estoque
- [ ] Verificar se a peça já foi utilizada em alguma Ordem de Serviço
- [ ] Impedir exclusão física de peça com histórico

**Repositório**

- [ ] Ajustar `ItemEstoqueRepository` para a atualização de situação
- [ ] Implementar a verificação de utilização da peça em Ordens de Serviço

**Handler HTTP**

- [ ] Implementar `DELETE /estoque/pecas/{pecaId}`
- [ ] Validar o path param `pecaId`
- [ ] Criar DTO/response com o recurso atualizado
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados
- [ ] Retornar `404` para peça inexistente e `409` quando a peça não puder ser desativada

**Testes unitários**

- [ ] Desativação válida
- [ ] Peça inexistente
- [ ] Peça já inativa
- [ ] Peça com saldo em estoque
- [ ] Peça utilizada em histórico
- [ ] Usuário sem autorização

**Testes de integração**

- [ ] `200` com a situação inativa persistida
- [ ] `404` para peça inexistente e `409` para peça já inativa
- [ ] `403` para perfil sem permissão
- [ ] Preservação do histórico das Ordens de Serviço

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI, explicando a exclusão lógica

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
