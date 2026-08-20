---
documento: Refinamento de Requisitos — Atualizar Insumo
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Atualizar Insumo

Este documento detalha a tarefa Atualizar Insumo do contexto de Peças & Insumos.

## 3 · Atualizar Insumo

### 3.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Atualizar os dados cadastrais de um insumo (descrição, unidade de medida, custo, estoque
mínimo, categoria e situação).

**Problema**
Insumo não é cobrado item a item do cliente — entra como custo diluído no serviço. Se a
unidade de medida ou o custo estiverem errados, o custo real do serviço fica distorcido e a
oficina perde margem sem perceber.

**Pré-condições**

- O insumo deve estar cadastrado.
- O usuário deve estar autorizado a manter o cadastro de insumos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-EST-12 | Permitir alterar descrição, unidade de medida, custo unitário, estoque mínimo e categoria. |
| RF-EST-13 | Permitir inativar e reativar o insumo. |
| RF-EST-14 | Validar os dados informados antes de gravar. |
| RF-EST-15 | Impedir alteração de unidade de medida quando houver saldo físico maior que zero. |
| RF-EST-16 | Registrar o histórico de alteração de custo, com data e responsável. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-11 | A operação deve ser feita por API RESTful. |
| RNF-EST-12 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-EST-13 | A alteração deve ser auditável. |
| RNF-EST-14 | A operação não deve alterar saldo de estoque. |
| RNF-EST-15 | A alteração de custo não pode ter efeito retroativo sobre serviços já finalizados. |

**Fluxo Principal**

1. O mecânico seleciona o insumo a ser atualizado.
2. O sistema apresenta os dados atuais do insumo.
3. O mecânico informa os novos dados.
4. O sistema valida os dados informados.
5. O sistema grava a alteração.
6. O sistema registra o histórico da alteração.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Insumo não encontrado | Informa que o insumo não existe. |
| A2 | Dados inválidos | Informa quais campos estão incorretos e não grava nada. |
| A3 | Alteração de unidade de medida com saldo em estoque | Impede a alteração e orienta a zerar o saldo antes. |
| A4 | Custo menor ou igual a zero | Impede a gravação. |
| A5 | Usuário sem autorização | Impede a operação. |

**Saída**

- Insumo atualizado com os novos dados; **ou**
- Indicação dos erros de validação encontrados.

**Pós-condições**

- Os dados cadastrais do insumo refletem a última alteração.
- O saldo físico permanece inalterado.
- Serviços já finalizados mantêm o custo registrado no momento de sua execução.
- O histórico de alteração fica disponível para auditoria.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
PUT /estoque/insumos/{insumoId}
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Path | `insumoId` | uuid   | Identificador do insumo |
| Header | `If-Match` | string | `version` atual do registro, para controle de concorrência |
| Body | `descricao` | string | Obrigatório, 3 a 120 caracteres |
| Body | `categoria` | string | Categoria do insumo |
| Body | `unidadeMedida` | enum | `UN` \| `L` \| `ML` \| `KG` \| `G` \| `M` |
| Body | `custoUnitario` | decimal | Obrigatório, maior que zero |
| Body | `estoqueMinimo` | decimal | Maior ou igual a zero, aceita casas decimais |
| Body | `ativo` | boolean | Situação do insumo |

```json
{
  "descricao": "Óleo lubrificante 15W40",
  "categoria": "Lubrificantes",
  "unidadeMedida": "L",
  "custoUnitario": 32.50,
  "estoqueMinimo": 20,
  "ativo": true
}
```

**Validações**

*Técnicas*

- `insumoId` existe e é do tipo `INSUMO`.
- `descricao` obrigatória, de 3 a 120 caracteres.
- `custoUnitario` maior que zero.
- `unidadeMedida` pertence ao conjunto permitido (`UN`, `L`, `ML`, `KG`, `G`, `M`).
- `estoqueMinimo` maior ou igual a zero, aceita decimal.
- `If-Match` deve bater com a `version` atual do registro.

*Negócio*

- `descricao` única dentro da mesma categoria.
- O `codigo` é o identificador de negócio do insumo e não é alterado por esta operação.
- Alteração de `unidadeMedida` bloqueada quando `saldoFisico > 0` — converter unidade com saldo distorce todo o histórico.
- Alteração de custo não retroage sobre serviços finalizados.

**Processamento**

1. Carregar o insumo por id.
2. Validar `If-Match`.
3. Se `unidadeMedida` mudou, verificar `saldoFisico == 0`.
4. Validar unicidade de descrição na categoria.
5. Detectar mudança de `custoUnitario`.
6. Aplicar as alterações na entidade.
7. Gravar registro em `historico_preco_item` quando o custo mudar.
8. Persistir e incrementar `version`.
9. Publicar o evento `InsumoAtualizado`.

**Persistência**

- Consulta: `item_estoque`
- Altera: `item_estoque`, `historico_preco_item`
- Não altera: `saldo_fisico`, `saldo_reservado`

**Saída da API**

```json
{
  "id": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
  "codigo": "IN-0031",
  "tipo": "INSUMO",
  "descricao": "Óleo lubrificante 15W40",
  "categoria": "Lubrificantes",
  "unidadeMedida": "L",
  "custoUnitario": 32.50,
  "estoqueMinimo": 20,
  "saldoFisico": 47.5,
  "ativo": true,
  "version": 4,
  "atualizadoEm": "2026-08-12T14:35:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Insumo atualizado |
| `400` | Body inválido; `unidadeMedida` fora do conjunto permitido |
| `401` | Token ausente ou expirado |
| `403` | Perfil sem permissão |
| `404` | Insumo não encontrado |
| `409` | Descrição duplicada na categoria |
| `412` | `If-Match` divergente — registro alterado por outro usuário |
| `422` | Troca de unidade de medida com saldo em estoque |

**Dependências**

- `ItemEstoqueRepository`
- `HistoricoPrecoRepository`
- Publicador de eventos de domínio

**Testes**

*Unitários*

- Rejeita `unidadeMedida` fora do enum.
- Bloqueia troca de unidade com `saldoFisico > 0`.
- Permite troca de unidade com `saldoFisico = 0`.
- Aceita `estoqueMinimo` decimal.

*Integração*

- `PUT` válido retorna `200`.
- Troca de unidade com saldo retorna `422`.
- `If-Match` divergente retorna `412`.
- Insumo inexistente retorna `404`.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `atualizarDados()` na entidade `ItemEstoque` com as regras de insumo
- [ ] Implementar o enum de unidade de medida (`UN`, `L`, `ML`, `KG`, `G`, `M`)
- [ ] Implementar o bloqueio de troca de unidade de medida com saldo físico maior que zero
- [ ] Garantir que a alteração de custo não retroage sobre serviços já finalizados

**Caso de uso**

- [ ] Implementar `AtualizarInsumo`

**Repositório**

- [ ] Registrar em `HistoricoPrecoRepository` na mudança de custo

**Handler HTTP**

- [ ] Implementar `PUT /estoque/insumos/{insumoId}`

**Validações**

- [ ] Validar `descricao` entre 3 e 120 caracteres
- [ ] Validar `custoUnitario` maior que zero
- [ ] Validar `unidadeMedida` dentro do enum permitido
- [ ] Validar descrição única dentro da mesma categoria

**Concorrência**

- [ ] Implementar controle otimista via `If-Match`

**Eventos**

- [ ] Publicar `InsumoAtualizado`

**Testes unitários**

- [ ] Rejeição de unidade fora do enum
- [ ] Bloqueio de troca de unidade com saldo maior que zero
- [ ] Troca de unidade permitida com saldo zerado
- [ ] Aceitação de `estoqueMinimo` decimal

**Testes de integração**

- [ ] `PUT` válido retornando `200`
- [ ] Troca de unidade com saldo retornando `422`
- [ ] `If-Match` divergente retornando `412`
- [ ] Insumo inexistente retornando `404`

**Documentação**

- [ ] Documentar no Swagger/OpenAPI

**Review**

- [ ] Code Review aprovado

---
