---
documento: Refinamento de Requisitos — Atualizar Peça
dono: José Lázaro
versao: 0.4
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Atualizar Peça

Este documento detalha a tarefa Atualizar Peça do contexto de Peças.

## 3 · Atualizar Peça

### 3.1 Refinamento de Produto

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
| RF-PEC-25 | Permitir alterar descrição, preço de venda, estoque mínimo, categoria — pelo `categoriaId` — e fabricante. |
| RF-PEC-26 | Permitir inativar e reativar a peça.                                                |
| RF-PEC-27 | Validar os dados informados antes de gravar.                                        |
| RF-PEC-28 | Impedir descrição duplicada dentro da mesma categoria.                              |
| RF-PEC-29 | Registrar o histórico de alteração de preço, com data e responsável.                |
| RF-PEC-30 | Manter inalterados os valores já registrados em OS e orçamentos anteriores.         |

**Requisitos Não Funcionais**

| ID         | Requisito                                                                           |
| ---------- | ----------------------------------------------------------------------------------- |
| RNF-PEC-15 | A operação deve ser feita por API RESTful.                                          |
| RNF-PEC-16 | A operação deve ser acessível somente por usuário autorizado com perfil de estoque. |
| RNF-PEC-17 | A alteração deve ser auditável — quem alterou, quando e o valor anterior.           |
| RNF-PEC-18 | A alteração de preço não pode ter efeito retroativo sobre documentos já emitidos.   |
| RNF-PEC-19 | A operação não deve alterar saldo de estoque.                                       |

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

### 3.2 Refinamento Técnico

**Endpoint**

```http
PUT /estoque/pecas/{pecaId}
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfil: `MECANICO`
- Escopo: `estoque:escrever`

**Entrada**

| Local  | Param           | Tipo    | Descrição                                                  |
| ------ | --------------- | ------- | ---------------------------------------------------------- |
| Path   | `pecaId`        | uuid    | Identificador da peça                                      |
| Header | `If-Match`      | string  | **Obrigatório.** `version` atual do registro, para controle de concorrência |
| Body   | `nome`          | string  | Obrigatório; nome curto da peça                            |
| Body   | `descricao`     | string  | Obrigatório, 3 a 120 caracteres                            |
| Body   | `categoria`     | string  | Categoria da peça                                          |
| Body   | `fabricante`    | string  | Fabricante da peça                                         |
| Body   | `precoVenda`    | decimal | Obrigatório, maior que zero, até 2 casas decimais          |
| Body   | `estoqueMinimo` | int     | Maior ou igual a zero                                      |
| Body   | `ativo`         | —       | **Não aceito.** A inativação é feita pelo `DELETE`         |

```json
{
  "nome": "Pastilha de freio",
  "descricao": "Pastilha de freio dianteira cerâmica",
  "categoriaId": "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94",
  "fabricante": "Bosch",
  "precoVenda": 199.9,
  "estoqueMinimo": 6
}
```

> **Decisão de projeto.** `ativo` **não é aceito** neste endpoint. Existiam dois caminhos para
> inativar a mesma peça — `PUT` com `ativo: false` e `DELETE` —, com validações diferentes, o que
> permitia burlar a checagem de saldo usando o caminho mais permissivo. A situação passa a mudar
> apenas pelo `DELETE`, onde as validações estão. Mesma regra já adotada em Serviços.

> **Decisão de projeto.** `nome` passa a ser atualizável e retornado, junto com `descricao`: o
> campo era gravado no cadastro e sumia daqui e da consulta.

**Validações**

_Técnicas_

- `pecaId` existe e é do tipo `PECA`.
- `nome` obrigatório.
- `descricao` obrigatória, de 3 a 120 caracteres.
- `precoVenda` obrigatório, maior que zero, com no máximo 2 casas decimais.
- `estoqueMinimo` inteiro maior ou igual a zero.
- `If-Match` é obrigatório e deve bater com a `version` atual do registro.

_Negócio_

- `descricao` normalizada única dentro da mesma categoria, entre peças **ativas**.
- O `codigo` é o identificador de negócio da peça e não é alterado por esta operação.
- A operação não altera `ativo`: qualquer valor enviado para esse campo retorna `400`.
- Alteração de preço não propaga para OS ou orçamentos já emitidos.

**Processamento**

1. Carregar a peça por id, com lock otimista.
2. Exigir o header `If-Match` — ausente, retorna `428`.
3. Comparar `If-Match` com a `version` atual — divergente, retorna `412`.
3. Rejeitar `ativo` no corpo, se vier.
4. Carregar a categoria pelo `categoriaId`, validar que existe e está ativa, e normalizar a
   descrição para validar unicidade na categoria, entre peças ativas.
5. Detectar mudança de `precoVenda`.
6. Aplicar as alterações na entidade.
7. Se o preço mudou, gravar registro em `historico_preco_item`.
8. Persistir e incrementar `version`.

**Persistência**

- Consulta: `item_estoque`, `reserva_estoque`
- Altera: `item_estoque` (dados cadastrais e `version`), `historico_preco_item` (insert quando o preço muda)
- Não altera: `saldo_fisico`, `saldo_reservado`

**Saída da API**

```json
{
  "id": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
  "codigo": "PEC-000142",
  "tipo": "PECA",
  "nome": "Pastilha de freio",
  "descricao": "Pastilha de freio dianteira cerâmica",
  "categoriaId": "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94",
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
| `400`  | Body inválido — campo obrigatório ausente, preço menor ou igual a zero, ou `ativo` enviado |
| `401`  | Token ausente ou expirado                                                      |
| `403`  | Perfil sem permissão                                                           |
| `404`  | Peça não encontrada                                                            |
| `409`  | Descrição duplicada na categoria entre peças ativas                             |
| `412`  | `If-Match` divergente — registro alterado por outro usuário                    |
| `428`  | `If-Match` ausente                                                             |

**Dependências**

- `ItemEstoqueRepository`
- `HistoricoPrecoRepository`

**Testes**

_Unitários_

- Rejeita `precoVenda` zero e negativo.
- Rejeita `descricao` com menos de 3 caracteres.
- Gera registro de histórico apenas quando o preço muda.
- Rejeita `ativo` enviado no corpo.
- Atualiza `nome` e `descricao` juntos.

_Integração_

- `PUT` válido retorna `200` e incrementa `version`.
- `PUT` com `If-Match` antigo retorna `412`, e sem o header retorna `428`.
- Descrição duplicada na mesma categoria retorna `409`.
- Descrição duplicada em categoria diferente é aceita.
- Descrição igual à de uma peça inativa é aceita.
- `ativo` no corpo retorna `400`.
- Peça inexistente retorna `404`.
- Perfil sem o escopo `estoque:escrever` recebe `403`.

_Regressão_

- Alterar o preço da peça não altera o `precoNoMomento` de OS já criadas.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `atualizarDados()` na entidade `ItemEstoque` com as regras de peça
- [ ] Rejeitar `ativo` no payload: a situação muda apenas pelo `DELETE`
- [ ] Implementar a normalização da descrição para a checagem de duplicidade
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

**Auditoria**

- [ ] Registrar a atualização na trilha de auditoria

**Testes unitários**

- [ ] Rejeição de preço zero e negativo
- [ ] Registro de histórico apenas quando o preço muda
- [ ] Bloqueio de inativação com reserva ativa

**Testes de integração**

- [ ] `PUT` válido retornando `200` com `version` incrementada
- [ ] `If-Match` divergente retornando `412` e ausente retornando `428`
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
