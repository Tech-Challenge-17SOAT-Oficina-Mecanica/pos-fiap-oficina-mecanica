---
documento: Refinamento de Requisitos — Desativar e Reativar Fornecedor
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Desativar e Reativar Fornecedor

Este documento detalha a tarefa Desativar Fornecedor, e sua contrapartida Reativar Fornecedor, do
contexto de Peças.

> **Escopo.** Fornecedor pertence ao agregado de **Compras**, cujo dono é o contexto de Peças.

## 13 · Desativar e Reativar Fornecedor

### 13.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Tirar de circulação um fornecedor com quem a oficina não compra mais, sem perder o histórico dos
pedidos já feitos a ele — e poder trazê-lo de volta.

**Problema**

Fornecedor que fechou, que passou a atender mal ou que a oficina deixou de usar continua
aparecendo na lista de escolha e vira erro de digitação em pedido novo. Apagar o registro não é
opção: os pedidos antigos perderiam a contra-parte e o histórico de compra ficaria órfão.

**Pré-condições**

- O fornecedor deve existir.
- O usuário deve estar autorizado a manter o cadastro de compras.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-107 | Permitir desativar um fornecedor. |
| RF-PEC-108 | Permitir reativar um fornecedor inativo. |
| RF-PEC-109 | Impedir a desativação enquanto houver pedido de compra em aberto para o fornecedor. |
| RF-PEC-110 | Impedir que fornecedor inativo seja informado em novo pedido de compra. |
| RF-PEC-111 | Preservar os pedidos de compra já emitidos. |
| RF-PEC-112 | Impedir a reativação quando existir outro fornecedor ativo com o mesmo documento. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-65 | A exclusão é lógica: o registro permanece no banco. |
| RNF-PEC-66 | A operação deve ser auditável. |
| RNF-PEC-67 | A operação deve ser acessível somente por usuário autorizado. |

**Fluxo Principal**

1. O mecânico seleciona o fornecedor e solicita a desativação.
2. O sistema verifica se existe pedido de compra em aberto para o fornecedor.
3. Não havendo, o sistema marca o fornecedor como inativo.
4. O sistema registra data, hora e usuário responsável.
5. O sistema confirma a desativação.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Fornecedor inexistente | Retorna `404`. |
| A2 | Fornecedor já inativo | Retorna `409`. |
| A3 | Pedido de compra em aberto para o fornecedor | Impede a desativação e informa os pedidos. |
| A4 | Reativação de fornecedor já ativo | Retorna `409`. |
| A5 | Reativação com documento já usado por outro fornecedor ativo | Impede a reativação. |
| A6 | Usuário sem autorização | Impede a operação. |

**Saída**

- Fornecedor inativo e indisponível para novos pedidos, com o histórico preservado.
- Ou o fornecedor de volta ao cadastro ativo, na reativação.

**Pós-condições**

- O fornecedor fica marcado como inativo e não pode ser informado em novo pedido.
- Os pedidos de compra já emitidos continuam apontando para ele.
- Nenhum saldo de estoque é alterado.

---

### 13.2 Refinamento Técnico

**Endpoint**

```http
DELETE /fornecedores/{fornecedorId}
POST   /fornecedores/{fornecedorId}/reativacao
```

O `DELETE` inativa o fornecedor; o `POST /reativacao` traz o fornecedor de volta.

> **Decisão de projeto.** A exclusão é **lógica**, com `DELETE`, e o registro permanece no banco —
> o mesmo verbo e a mesma semântica de cliente, veículo, peça, insumo e serviço (D-20).

> **Decisão de projeto.** A desativação é **bloqueada com pedido de compra em aberto**. O pedido em
> aberto é uma obrigação viva: existe reserva atrelada a ele e OS esperando o recebimento. Saldo em
> pedido já concluído não bloqueia nada.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `compras:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `fornecedorId` | uuid | Identificador do fornecedor. |

Não há corpo em nenhuma das duas requisições.

**Validações**

*Técnicas*

- `fornecedorId` em formato UUID válido.

*Negócio — desativação*

- O fornecedor deve existir e estar ativo.
- Não pode existir pedido de compra em `ABERTO` ou `PARCIAL` para o fornecedor.

*Negócio — reativação*

- O fornecedor deve existir e estar inativo.
- Não pode existir outro fornecedor **ativo** com o mesmo documento.

**Regra de domínio**

```
ativo → DELETE → inativo → POST /reativacao → ativo
```

**Processamento**

*Desativação*

1. Carregar o fornecedor e validar existência e situação.
2. Consultar pedidos de compra em `ABERTO` ou `PARCIAL` para o fornecedor.
3. Havendo pedidos, abortar com `409` e a lista dos pedidos.
4. Marcar `ativo = false` e gravar `inativadoEm` e `inativadoPor`.
5. Registrar na trilha de auditoria e persistir.

*Reativação*

1. Carregar o fornecedor e validar que está inativo.
2. Verificar que não existe outro fornecedor ativo com o mesmo documento.
3. Marcar `ativo = true` e limpar os campos de inativação.
4. Registrar na trilha de auditoria e persistir.

**Persistência**

- Consulta: `fornecedor`, `pedido_compra`.
- Altera, na desativação: `fornecedor` (`ativo = false`, `inativado_em`, `inativado_por`).
- Altera, na reativação: `fornecedor` (`ativo = true`, campos de inativação nulos).
- Não altera: `pedido_compra`, `item_estoque` nem qualquer saldo.
- O índice parcial `UNIQUE (documento) WHERE ativo = true` é o que garante a unicidade na
  reativação.

**Saída da API**

`DELETE` — `200`:

```json
{
  "id": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
  "razaoSocial": "Auto Peças Bandeirantes Ltda",
  "documento": "12345678000190",
  "ativo": false,
  "inativadoEm": "2026-08-22T12:05:00-03:00",
  "inativadoPor": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
}
```

`POST /reativacao` — `200`:

```json
{
  "id": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
  "razaoSocial": "Auto Peças Bandeirantes Ltda",
  "documento": "12345678000190",
  "ativo": true,
  "inativadoEm": null,
  "inativadoPor": null
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Fornecedor desativado ou reativado, com o recurso atualizado no corpo. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `compras:escrever`. |
| `404` | Fornecedor não encontrado. |
| `409` | Fornecedor já inativo no `DELETE`, já ativo na reativação, pedido de compra em aberto, ou documento já usado por outro fornecedor ativo. |

**Dependências**

- `FornecedorRepository`.
- `PedidoCompraRepository`, para verificar pedidos em aberto.
- Middleware de autenticação e autorização.
- Trilha de auditoria.

**Testes**

*Unitários*

- Desativa fornecedor ativo sem pedido em aberto.
- Rejeita desativação com pedido em `ABERTO` ou `PARCIAL`.
- Rejeita desativação de fornecedor já inativo.
- Reativa fornecedor inativo.
- Rejeita reativação de fornecedor já ativo.
- Rejeita reativação com documento já usado por outro fornecedor ativo.

*Integração*

- `DELETE` válido retorna `200` com `ativo` em `false`.
- `POST /reativacao` válido retorna `200` com `ativo` em `true`.
- Pedido em aberto retorna `409` com a lista dos pedidos.
- O registro não é removido fisicamente do banco.
- Fornecedor inativo não pode ser informado em `POST /compras/pedidos`.
- Pedidos de compra já emitidos permanecem apontando para o fornecedor.

---

### 13.3 Checklist de Implementação

**Domínio**

- [ ] Implementar os métodos `desativar()` e `reativar()` em `Fornecedor`
- [ ] Registrar data, hora e usuário responsável pela inativação
- [ ] Impedir o uso de fornecedor inativo em novo pedido de compra

**Caso de uso**

- [ ] Implementar `DesativarFornecedor`
- [ ] Implementar `ReativarFornecedor`
- [ ] Verificar pedidos de compra em aberto antes de desativar
- [ ] Validar, na reativação, a unicidade do documento entre ativos

**Repositório**

- [ ] Criar consulta de pedidos de compra em aberto por fornecedor
- [ ] Persistir a transição de situação

**Integrações**

- [ ] Integrar com `PedidoCompraRepository`

**Handler HTTP**

- [ ] Implementar `DELETE /fornecedores/{fornecedorId}`
- [ ] Implementar `POST /fornecedores/{fornecedorId}/reativacao`
- [ ] Aplicar autenticação JWT e autorização pelo escopo `compras:escrever`
- [ ] Retornar `404` e `409` nos casos previstos

**Auditoria**

- [ ] Registrar a inativação e a reativação na trilha de auditoria

**Testes unitários**

- [ ] Desativação válida
- [ ] Desativação com pedido em aberto bloqueada
- [ ] Reativação válida
- [ ] Reativação com documento duplicado bloqueada

**Testes de integração**

- [ ] `DELETE` retornando `200` com a situação inativa persistida
- [ ] `POST /reativacao` retornando `200`
- [ ] Registro não removido fisicamente
- [ ] Fornecedor inativo recusado em novo pedido de compra

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI, explicando a exclusão lógica

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
