---
documento: Refinamento de Requisitos — Atualizar Peça
dono: José Lázaro
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Atualizar Peça

Este documento detalha a tarefa Atualizar Peça do contexto de Peças & Insumos.

## 2 · Atualizar Peça

### 2.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Atualizar os dados cadastrais de uma peça (descrição, preço de venda, estoque mínimo,
categoria, fabricante e situação).

**Problema**
Preço de fornecedor, embalagem e ponto de reposição mudam com o tempo. Sem atualização,
o orçamento é calculado com valor defasado — a oficina vende no prejuízo ou perde o cliente
por preço fora de mercado — e o alerta de reposição dispara na hora errada.

**Pré-condições**

- A peça deve estar cadastrada.
- O usuário deve estar autorizado a manter o cadastro de peças.

**Requisitos Funcionais**

| ID        | Requisito                                                                           |
| --------- | ----------------------------------------------------------------------------------- |
| RF-EST-06 | Permitir alterar descrição, preço de venda, estoque mínimo, categoria e fabricante. |
| RF-EST-07 | Permitir inativar e reativar a peça.                                                |
| RF-EST-08 | Validar os dados informados antes de gravar.                                        |
| RF-EST-09 | Impedir descrição duplicada dentro da mesma categoria.                              |
| RF-EST-10 | Registrar o histórico de alteração de preço, com data e responsável.                |
| RF-EST-11 | Manter inalterados os valores já registrados em OS e orçamentos anteriores.         |

**Requisitos Não Funcionais**

| ID         | Requisito                                                                           |
| ---------- | ----------------------------------------------------------------------------------- |
| RNF-EST-06 | A operação deve ser feita por API RESTful.                                          |
| RNF-EST-07 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-EST-08 | A alteração deve ser auditável — quem alterou, quando e o valor anterior.           |
| RNF-EST-09 | A alteração de preço não pode ter efeito retroativo sobre documentos já emitidos.   |
| RNF-EST-10 | A operação não deve alterar saldo de estoque.                                       |

**Fluxo Principal**

1. O mecânico seleciona a peça a ser atualizada.
2. O sistema apresenta os dados atuais da peça.
3. O mecânico informa os novos dados.
4. O sistema valida os dados informados.
5. O sistema grava a alteração.
6. O sistema registra o histórico da alteração.

**Fluxos Alternativos / Exceções**

| #   | Situação                                       | Comportamento do sistema                                 |
| --- | ---------------------------------------------- | -------------------------------------------------------- |
| A1  | Peça não encontrada                            | Informa que a peça não existe.                           |
| A2  | Dados inválidos                                | Informa quais campos estão incorretos e não grava nada.  |
| A3  | Descrição duplicada na categoria               | Impede a gravação.                                       |
| A4  | Preço menor ou igual a zero                    | Impede a gravação.                                       |
| A5  | Tentativa de inativar peça com saldo reservado | Impede a inativação e informa quais OS mantêm a reserva. |
| A6  | Usuário sem autorização                        | Impede a operação.                                       |

**Saída**

- Peça atualizada com os novos dados; **ou**
- Indicação dos erros de validação encontrados.

**Pós-condições**

- Os dados cadastrais da peça refletem a última alteração.
- O saldo físico e o saldo reservado permanecem inalterados.
- OS e orçamentos anteriores mantêm o preço registrado no momento de sua emissão.
- O histórico de alteração fica disponível para auditoria.

---

### 2.2 Refinamento Técnico

**Endpoint**

```http
PUT /estoque/pecas/{pecaId}
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfis: `MECANICO`, `GESTOR`
- Escopo: `estoque:escrever`

**Entrada**

| Local  | Param           | Tipo    | Descrição                                                  |
| ------ | --------------- | ------- | ---------------------------------------------------------- |
| Path   | `pecaId`        | uuid    | Identificador da peça                                      |
| Header | `If-Match`      | string  | `version` atual do registro, para controle de concorrência |
| Body   | `descricao`     | string  | Obrigatório, 3 a 120 caracteres                            |
| Body   | `categoria`     | string  | Categoria da peça                                          |
| Body   | `fabricante`    | string  | Fabricante da peça                                         |
| Body   | `precoVenda`    | decimal | Obrigatório, maior que zero, até 2 casas decimais          |
| Body   | `estoqueMinimo` | int     | Maior ou igual a zero                                      |
| Body   | `ativo`         | boolean | Situação da peça                                           |

```json
{
  "descricao": "Pastilha de freio dianteira cerâmica",
  "categoria": "Freios",
  "fabricante": "Bosch",
  "precoVenda": 199.9,
  "estoqueMinimo": 6,
  "ativo": true
}
```

**Validações**

_Técnicas_

- `pecaId` existe e é do tipo `PECA`.
- `descricao` obrigatória, de 3 a 120 caracteres.
- `precoVenda` obrigatório, maior que zero, com no máximo 2 casas decimais.
- `estoqueMinimo` inteiro maior ou igual a zero.
- `If-Match` deve bater com a `version` atual do registro.

_Negócio_

- `descricao` única dentro da mesma categoria.
- O `codigo` é o identificador de negócio da peça e não é alterado por esta operação.
- Não permitir `ativo = false` quando `saldoReservado > 0`.
- Alteração de preço não propaga para OS ou orçamentos já emitidos.

**Processamento**

1. Carregar a peça por id, com lock otimista.
2. Comparar `If-Match` com a `version` atual — conflito retorna `412`.
3. Validar unicidade de descrição na categoria.
4. Se `ativo = false`, verificar reservas ativas.
5. Detectar mudança de `precoVenda`.
6. Aplicar as alterações na entidade.
7. Se o preço mudou, gravar registro em `historico_preco_item`.
8. Persistir e incrementar `version`.
9. Publicar o evento `PecaAtualizada`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`
- Altera: `item_estoque` (dados cadastrais e `version`), `historico_preco_item` (insert quando o preço muda)
- Não altera: `saldo_fisico`, `saldo_reservado`

**Saída da API**

```json
{
  "id": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
  "codigo": "PC-0142",
  "tipo": "PECA",
  "descricao": "Pastilha de freio dianteira cerâmica",
  "categoria": "Freios",
  "fabricante": "Bosch",
  "precoVenda": 199.9,
  "estoqueMinimo": 6,
  "ativo": true,
  "version": 8,
  "atualizadoEm": "2026-08-12T14:30:00-03:00",
  "atualizadoPor": "0e93b571-2ac6-4d18-95f7-8b40e6c31a29"
}
```

**Códigos HTTP / Erros**

| Código | Situação                                                                       |
| ------ | ------------------------------------------------------------------------------ |
| `200`  | Peça atualizada                                                                |
| `400`  | Body inválido — campo obrigatório ausente, preço menor ou igual a zero         |
| `401`  | Token ausente ou expirado                                                      |
| `403`  | Perfil sem permissão                                                           |
| `404`  | Peça não encontrada                                                            |
| `409`  | Descrição duplicada na categoria; tentativa de inativar peça com reserva ativa |
| `412`  | `If-Match` divergente — registro alterado por outro usuário                    |
| `422`  | Regra de negócio violada                                                       |

**Dependências**

- `ItemEstoqueRepository`
- `ReservaEstoqueRepository` (verificação de reserva ativa)
- `HistoricoPrecoRepository`
- Publicador de eventos de domínio

**Testes**

_Unitários_

- Rejeita `precoVenda` zero e negativo.
- Rejeita `descricao` com menos de 3 caracteres.
- Gera registro de histórico apenas quando o preço muda.
- Bloqueia inativação com reserva ativa.

_Integração_

- `PUT` válido retorna `200` e incrementa `version`.
- `PUT` com `If-Match` antigo retorna `412`.
- Descrição duplicada na mesma categoria retorna `409`.
- Descrição duplicada em categoria diferente é aceita.
- Inativar peça com reserva ativa retorna `409`.
- Peça inexistente retorna `404`.
- Perfil sem o escopo `estoque:escrever` recebe `403`.

_Regressão_

- Alterar o preço da peça não altera o `precoNoMomento` de OS já criadas.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `atualizarDados()` na entidade `ItemEstoque` com as regras de peça
- [ ] Implementar a regra de bloqueio de inativação quando houver saldo reservado
- [ ] Garantir que a alteração de preço não altera o `precoNoMomento` de OS já emitidas

**Caso de uso**

- [ ] Implementar `AtualizarPeca`

**Repositório**

- [ ] Implementar `ItemEstoqueRepository.salvar` com incremento de `version`
- [ ] Implementar `HistoricoPrecoRepository`
- [ ] Registrar o histórico apenas quando o preço mudar

**Handler HTTP**

- [ ] Implementar `PUT /estoque/pecas/{pecaId}`

**Validações**

- [ ] Validar `descricao` entre 3 e 120 caracteres
- [ ] Validar `precoVenda` maior que zero, com no máximo 2 casas decimais
- [ ] Validar `estoqueMinimo` maior ou igual a zero
- [ ] Validar descrição única dentro da mesma categoria

**Concorrência**

- [ ] Implementar controle otimista comparando o header `If-Match` com a `version` atual

**Eventos**

- [ ] Publicar `PecaAtualizada`

**Testes unitários**

- [ ] Rejeição de preço zero e negativo
- [ ] Registro de histórico apenas quando o preço muda
- [ ] Bloqueio de inativação com reserva ativa

**Testes de integração**

- [ ] `PUT` válido retornando `200` com `version` incrementada
- [ ] `If-Match` divergente retornando `412`
- [ ] Descrição duplicada na mesma categoria retornando `409`
- [ ] Descrição duplicada em categoria diferente sendo aceita
- [ ] Peça inexistente retornando `404`
- [ ] Perfil sem o escopo `estoque:escrever` retornando `403`
- [ ] Regressão: OS antiga mantém o preço original após alteração no catálogo

**Documentação**

- [ ] Documentar no Swagger/OpenAPI, incluindo o header `If-Match`

**Review**

- [ ] Code Review aprovado

---
