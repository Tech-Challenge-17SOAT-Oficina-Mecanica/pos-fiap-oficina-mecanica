---
documento: Refinamento de Requisitos — Atualizar Fornecedor
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Atualizar Fornecedor

Este documento detalha a tarefa Atualizar Fornecedor do contexto de Peças.

> **Escopo.** Fornecedor pertence ao agregado de **Compras**, cujo dono é o contexto de Peças.

## 12 · Atualizar Fornecedor

### 12.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Corrigir os dados cadastrais do fornecedor — razão social, nome fantasia e contato — sem afetar os
pedidos de compra já emitidos.

**Problema**

Fornecedor troca de telefone, muda de razão social e passa a atender por outro e-mail. Sem
atualização, o cadastro envelhece e o contato registrado deixa de funcionar justamente quando a
oficina precisa cobrar um pedido atrasado.

**Pré-condições**

- O fornecedor deve existir e estar ativo.
- O usuário deve estar autorizado a manter o cadastro de compras.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-102 | Permitir atualizar razão social, nome fantasia, contato e prazo de entrega do fornecedor. |
| RF-PEC-103 | Manter o `documento` imutável após o cadastro. |
| RF-PEC-104 | Manter ao menos um contato preenchido após a atualização. |
| RF-PEC-105 | Preservar os pedidos de compra já emitidos para o fornecedor. |
| RF-PEC-106 | Impedir que a atualização altere a situação do fornecedor. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-62 | A operação deve ser protegida contra escrita concorrente. |
| RNF-PEC-63 | A operação deve registrar data, hora e usuário responsável pela alteração. |
| RNF-PEC-64 | A operação deve ser acessível somente por usuário autorizado. |

**Fluxo Principal**

1. O mecânico seleciona o fornecedor e informa os dados a alterar.
2. O sistema valida os campos e a `version` enviada.
3. O sistema aplica as alterações.
4. O sistema confirma e devolve o fornecedor atualizado.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Fornecedor inexistente | Retorna `404`. |
| A2 | Fornecedor inativo | Impede a alteração e informa que é preciso reativá-lo antes. |
| A3 | `documento` enviado no corpo | Impede a operação: o documento é imutável. |
| A4 | `ativo` enviado no corpo | Impede a operação: a situação muda pelo `DELETE` e pela reativação. |
| A5 | Nenhum contato após a alteração | Impede a operação. |
| A6 | `If-Match` divergente | Impede a alteração: outro usuário alterou o registro. |
| A7 | Usuário sem autorização | Impede a operação. |

**Saída**

- Fornecedor com os dados atualizados.
- Ou indicação do motivo pelo qual a alteração foi recusada.

**Pós-condições**

- Os dados cadastrais do fornecedor estão atualizados e a `version` foi incrementada.
- Os pedidos de compra já emitidos permanecem inalterados.

---

### 12.2 Refinamento Técnico

**Endpoint**

```http
PUT /fornecedores/{fornecedorId}
```

> **Decisão de projeto.** O `documento` é **imutável**: trocar o CNPJ de um fornecedor é, na
> prática, outro fornecedor, e mudá-lo reescreveria a contra-parte de pedidos já emitidos. Se o
> documento estava errado, o caminho é inativar e cadastrar de novo.

> **Decisão de projeto.** `ativo` **não é aceito** neste endpoint, como em peça, insumo e serviço:
> a situação muda apenas pelo `DELETE` e pela reativação, onde estão as validações.

> **Decisão de projeto.** A atualização usa controle otimista com `If-Match` e `version` (D-24).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `compras:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `fornecedorId` | uuid | Identificador do fornecedor. |
| Header | `If-Match` | string | `version` atual do registro, obrigatório. |
| Body | `razaoSocial` | string | Obrigatória; 3 a 120 caracteres. |
| Body | `nomeFantasia` | string | Opcional. |
| Body | `telefone` | string | Somente dígitos. Obrigatório se `email` não for informado. |
| Body | `email` | string | Obrigatório se `telefone` não for informado. |
| Body | `prazoEntregaDias` | inteiro | Opcional; de 1 a 365. |
| Body | `documento` | — | **Não aceito.** Imutável após o cadastro. |
| Body | `ativo` | — | **Não aceito.** A situação muda pelo `DELETE` e pela reativação. |

```json
{
  "razaoSocial": "Auto Peças Bandeirantes Ltda",
  "nomeFantasia": "Bandeirantes Autopeças",
  "telefone": "1133224466",
  "email": "compras@bandeirantes.com.br",
  "prazoEntregaDias": 10
}
```

**Validações**

*Técnicas*

- `fornecedorId` em formato UUID válido.
- `razaoSocial` obrigatória, de 3 a 120 caracteres.
- `telefone`, quando informado, com 10 ou 11 dígitos.
- `email`, quando informado, em formato válido.
- `prazoEntregaDias`, quando informado, inteiro de 1 a 365.
- `If-Match` obrigatório e igual à `version` atual.

*Negócio*

- O fornecedor deve existir e estar ativo.
- Pelo menos um entre `telefone` e `email` deve permanecer preenchido.
- `documento` e `ativo` no corpo retornam `400`.
- A alteração não afeta pedidos de compra já emitidos.

**Processamento**

1. Carregar o fornecedor por identificador, com lock otimista.
2. Comparar `If-Match` com a `version` atual — divergência retorna `412`, ausência retorna `428`.
3. Rejeitar `documento` e `ativo`, se vierem no corpo.
4. Validar os campos e a regra de contato.
5. Aplicar as alterações.
6. Registrar data, hora e usuário responsável.
7. Persistir, incrementar `version` e registrar na trilha de auditoria.
8. Retornar o fornecedor atualizado.

**Persistência**

- Consulta: `fornecedor`.
- Altera: `fornecedor` (`razao_social`, `nome_fantasia`, `telefone`, `email`,
  `prazo_entrega_dias`, `data_atualizacao`, `usuario_atualizacao`, `version`).
- Não altera: `documento`, `ativo`, `pedido_compra`.

**Saída da API**

```json
{
  "id": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
  "razaoSocial": "Auto Peças Bandeirantes Ltda",
  "nomeFantasia": "Bandeirantes Autopeças",
  "documento": "12345678000190",
  "tipoDocumento": "CNPJ",
  "telefone": "1133224466",
  "email": "compras@bandeirantes.com.br",
  "prazoEntregaDias": 10,
  "ativo": true,
  "version": 4,
  "dataAtualizacao": "2026-08-22T11:40:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Fornecedor atualizado. |
| `400` | Dados inválidos, nenhum contato informado, ou `documento`/`ativo` enviados. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `compras:escrever`. |
| `404` | Fornecedor não encontrado. |
| `409` | Fornecedor inativo. |
| `412` | `If-Match` divergente — registro alterado por outro usuário. |
| `428` | `If-Match` não informado. |

**Dependências**

- `FornecedorRepository`.
- Middleware de autenticação e autorização.
- Trilha de auditoria.

**Testes**

*Unitários*

- Atualiza razão social e contato.
- Rejeita `documento` no corpo.
- Rejeita `ativo` no corpo.
- Rejeita atualização que deixaria o fornecedor sem contato.
- Rejeita `If-Match` divergente.
- Incrementa `version`.

*Integração*

- `PUT` válido retorna `200` e persiste as alterações.
- Fornecedor inexistente retorna `404`.
- Fornecedor inativo retorna `409`.
- `If-Match` antigo retorna `412`; ausência retorna `428`.
- `documento` no corpo retorna `400`.
- Pedidos de compra já emitidos permanecem inalterados.

---

### 12.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método de atualização em `Fornecedor`
- [ ] Declarar `documento` e `ativo` como imutáveis neste fluxo
- [ ] Garantir que ao menos um contato permaneça preenchido

**Caso de uso**

- [ ] Implementar `AtualizarFornecedor`
- [ ] Comparar `If-Match` com a `version` atual antes de aplicar
- [ ] Registrar data, hora e usuário responsável

**Repositório**

- [ ] Criar método para consultar fornecedor por identificador
- [ ] Persistir alterações incrementando `version`

**Handler HTTP**

- [ ] Implementar `PUT /fornecedores/{fornecedorId}`
- [ ] Ler o header `If-Match` e devolver `428` quando ausente
- [ ] Aplicar autenticação JWT e autorização pelo escopo `compras:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Rejeitar com `400` os campos imutáveis enviados no corpo
- [ ] Validar a regra de contato obrigatório
- [ ] Retornar `409` para fornecedor inativo
- [ ] Retornar `412` para `If-Match` divergente

**Concorrência**

- [ ] Implementar controle otimista com `If-Match` comparado ao campo `version`

**Auditoria**

- [ ] Registrar a atualização na trilha de auditoria

**Testes unitários**

- [ ] Atualização válida
- [ ] Campo imutável enviado no corpo
- [ ] Atualização sem contato
- [ ] `If-Match` divergente

**Testes de integração**

- [ ] Endpoint atualiza fornecedor e retorna `200`
- [ ] Endpoint retorna `404`, `409`, `412` e `428` nos casos previstos
- [ ] Pedidos já emitidos preservados

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
