---
documento: Refinamento de Requisitos — Atualizar Insumo
dono: José Lázaro
versao: 0.4
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Atualizar Insumo

Este documento detalha a tarefa Atualizar Insumo do contexto de Insumos.

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
| RF-INS-24 | Permitir alterar descrição, unidade de medida, custo unitário, estoque mínimo e categoria, pelo `categoriaId`. |
| RF-INS-25 | Permitir inativar e reativar o insumo. |
| RF-INS-26 | Validar os dados informados antes de gravar. |
| RF-INS-27 | Impedir alteração de unidade de medida quando houver saldo físico maior que zero. |
| RF-INS-28 | Registrar o histórico de alteração de custo, com data e responsável. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-INS-14 | A operação deve ser feita por API RESTful. |
| RNF-INS-15 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-INS-16 | A alteração deve ser auditável. |
| RNF-INS-17 | A operação não deve alterar saldo de estoque. |
| RNF-INS-18 | A alteração de custo não pode ter efeito retroativo sobre serviços já finalizados. |

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
- Perfil: `MECANICO`
- Escopo: `estoque:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Path | `insumoId` | uuid   | Identificador do insumo |
| Header | `If-Match` | string | **Obrigatório.** `version` atual do registro, para controle de concorrência |
| Body | `nome` | string | Obrigatório; nome curto do insumo |
| Body | `descricao` | string | Obrigatório, 3 a 120 caracteres |
| Body | `categoriaId` | uuid | Categoria ativa do catálogo |
| Body | `unidadeMedida` | enum | `UN` \| `L` \| `ML` \| `KG` \| `G` \| `M` |
| Body | `custoUnitario` | decimal | Obrigatório, maior que zero |
| Body | `estoqueMinimo` | decimal | Maior ou igual a zero, aceita casas decimais |
| Body | `ativo` | — | **Não aceito.** A inativação é feita pelo `DELETE` |

```json
{
  "nome": "Óleo 15W40",
  "descricao": "Óleo lubrificante 15W40",
  "categoriaId": "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724",
  "unidadeMedida": "L",
  "custoUnitario": 32.50,
  "estoqueMinimo": 20
}
```

> **Decisão de projeto.** `ativo` **não é aceito** neste endpoint. A situação muda apenas pelo
> `DELETE`, onde ficam as validações de saldo reservado e de orçamento pendente — a peça já
> funcionava assim, e o insumo não tratava o caso, o que deixava reserva órfã.

> **Decisão de projeto.** `nome` passa a ser atualizável e retornado, junto com `descricao`.

> **Decisão de projeto.** O `custoUnitario` informado aqui é o **custo cadastral de referência**. O
> custo efetivo é atualizado pela **entrada de estoque**, com o último custo recebido, e cada
> alteração grava `historico_preco_item`. Média ponderada fica para depois (D-14).

**Validações**

*Técnicas*

- `insumoId` existe e é do tipo `INSUMO`.
- `nome` obrigatório.
- `descricao` obrigatória, de 3 a 120 caracteres.
- `custoUnitario` maior que zero.
- `unidadeMedida` pertence ao conjunto permitido (`UN`, `L`, `ML`, `KG`, `G`, `M`).
- `estoqueMinimo` maior ou igual a zero, aceita decimal.
- `If-Match` é obrigatório e deve bater com a `version` atual do registro.

*Negócio*

- `descricao` normalizada única dentro da mesma categoria e unidade de medida, entre insumos **ativos**.
- O `codigo` é o identificador de negócio do insumo e não é alterado por esta operação.
- Alteração de `unidadeMedida` bloqueada quando `saldoFisico > 0` — converter unidade com saldo distorce todo o histórico.
- Alteração de custo não retroage sobre serviços finalizados.
- A operação não altera `ativo`: qualquer valor enviado para esse campo retorna `400`.
- O último custo gravado pela entrada de estoque prevalece sobre o custo cadastral.

**Processamento**

1. Carregar o insumo por id.
2. Validar `If-Match`: ausente retorna `428`, divergente retorna `412`.
3. Rejeitar `ativo` no corpo, se vier.
4. Se `unidadeMedida` mudou, verificar `saldoFisico == 0`.
5. Carregar a categoria pelo `categoriaId`, validar que existe e está ativa, normalizar a
   descrição e validar unicidade na categoria e unidade, entre insumos ativos.
6. Detectar mudança de `custoUnitario`.
7. Aplicar as alterações na entidade.
8. Gravar registro em `historico_preco_item` quando o custo mudar.
9. Persistir e incrementar `version`.
10. Registrar a atualização na trilha de auditoria.

**Persistência**

- Consulta: `item_estoque`
- Altera: `item_estoque`, `historico_preco_item`
- Não altera: `saldo_fisico`, `saldo_reservado`

**Saída da API**

```json
{
  "id": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
  "codigo": "INS-000031",
  "tipo": "INSUMO",
  "nome": "Óleo 15W40",
  "descricao": "Óleo lubrificante 15W40",
  "categoriaId": "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724",
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
| `409` | Descrição duplicada na mesma categoria e unidade, entre insumos ativos |
| `412` | `If-Match` divergente — registro alterado por outro usuário |
| `428` | `If-Match` ausente |
| `409` | Troca de unidade de medida com saldo em estoque |

**Dependências**

- `ItemEstoqueRepository`
- `HistoricoPrecoRepository`
- Trilha de auditoria

**Testes**

*Unitários*

- Rejeita `unidadeMedida` fora do enum.
- Bloqueia troca de unidade com `saldoFisico > 0`.
- Permite troca de unidade com `saldoFisico = 0`.
- Aceita `estoqueMinimo` decimal.

*Integração*

- `PUT` válido retorna `200`.
- Troca de unidade com saldo retorna `409`.
- `If-Match` divergente retorna `412`, e ausente retorna `428`.
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

- [ ] Validar `nome` obrigatório
- [ ] Validar `descricao` entre 3 e 120 caracteres
- [ ] Rejeitar `ativo` no payload: a situação muda apenas pelo `DELETE`
- [ ] Validar `custoUnitario` maior que zero
- [ ] Validar `unidadeMedida` dentro do enum permitido
- [ ] Validar descrição normalizada única na mesma categoria e unidade, entre insumos ativos

**Concorrência**

- [ ] Implementar controle otimista via `If-Match`

**Auditoria**

- [ ] Registrar a atualização na trilha de auditoria

**Testes unitários**

- [ ] Rejeição de unidade fora do enum
- [ ] Bloqueio de troca de unidade com saldo maior que zero
- [ ] Troca de unidade permitida com saldo zerado
- [ ] Aceitação de `estoqueMinimo` decimal

**Testes de integração**

- [ ] `PUT` válido retornando `200`
- [ ] Troca de unidade com saldo retornando `409`
- [ ] `If-Match` divergente retornando `412` e ausente retornando `428`
- [ ] Insumo inexistente retornando `404`

**Documentação**

- [ ] Documentar no Swagger/OpenAPI

**Review**

- [ ] Code Review aprovado

---
