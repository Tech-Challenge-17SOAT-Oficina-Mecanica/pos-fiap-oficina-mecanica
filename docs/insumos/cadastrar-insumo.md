---
documento: Refinamento de Requisitos — Cadastrar Insumo
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Insumo

Este documento detalha a tarefa Cadastrar Insumo do contexto de Insumos.

## 1 · Cadastrar Insumo

### 1.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Cadastrar um novo insumo utilizado pela oficina, permitindo seu controle de estoque e associação
às Ordens de Serviço.

**Problema**

Além das peças, a oficina utiliza insumos durante a execução dos serviços e precisa controlar sua
disponibilidade e consumo para evitar falta de materiais e inconsistências no estoque.

**Pré-condições**

- O usuário deve possuir autorização para manter o cadastro de insumos.
- Os dados obrigatórios do insumo devem estar disponíveis.
- O insumo não deve estar duplicado no cadastro, conforme a regra de identificação definida.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-INS-01 | Permitir cadastrar um insumo. |
| RF-INS-02 | Registrar o nome e a descrição do insumo. |
| RF-INS-03 | Registrar a unidade de medida. |
| RF-INS-04 | Registrar o custo unitário. |
| RF-INS-05 | Não aceitar estoque inicial no cadastro: o saldo entra pela movimentação de entrada. |
| RF-INS-06 | Permitir definir as informações necessárias para o controle de estoque, incluindo o estoque mínimo. |
| RF-INS-07 | Validar os dados informados. |
| RF-INS-08 | Impedir duplicidade: a descrição normalizada é única dentro da mesma categoria e unidade de medida, entre insumos ativos. |
| RF-INS-09 | Disponibilizar o insumo para utilização em Ordens de Serviço. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-INS-01 | O cadastro deve ser persistido de forma consistente. |
| RNF-INS-02 | O controle de estoque deve evitar saldos incorretos. |
| RNF-INS-03 | Os valores monetários devem ser armazenados de forma adequada. |
| RNF-INS-04 | Somente usuário autorizado poderá realizar o cadastro. |
| RNF-INS-05 | A operação deve preservar o histórico das movimentações de estoque. |
| RNF-INS-06 | O cadastro não deve alterar outros insumos já existentes. |

**Fluxo Principal**

1. O gestor acessa o gerenciamento de peças e insumos.
2. O gestor solicita o cadastro de um novo insumo.
3. O sistema apresenta o formulário de cadastro.
4. O gestor informa os dados do insumo.
5. O gestor informa a quantidade inicial, quando aplicável.
6. O sistema valida os campos obrigatórios.
7. O sistema verifica se já existe um insumo equivalente.
8. O sistema registra o insumo.
9. O sistema registra o estoque inicial, quando informado.
10. O sistema confirma o cadastro.
11. O insumo fica disponível para utilização nas Ordens de Serviço.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Dados obrigatórios ausentes | Informa os campos que precisam ser preenchidos. |
| A2 | Insumo duplicado | Impede o cadastro. |
| A3 | Unidade de medida inválida | Solicita uma unidade válida. |
| A4 | Quantidade inválida | Impede o registro do estoque. |
| A5 | Custo inválido | Rejeita o valor informado. |
| A6 | Usuário sem autorização | Impede a operação. |
| A7 | Falha no registro de estoque | Não conclui o cadastro até garantir a consistência entre o insumo e seu estoque. |

**Saída**

- Insumo cadastrado e disponível para controle de estoque e utilização em Ordens de Serviço.

**Pós-condições**

- O insumo passa a existir no catálogo e pode ser associado a uma Ordem de Serviço.
- O estoque inicial fica registrado, quando aplicável.
- As movimentações de estoque podem ser registradas ao longo da utilização do insumo.
- O cadastro fica disponível para composição de orçamentos e execução dos serviços.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/insumos
```

> **Decisão de projeto.** Insumo admite **quantidade fracionada**, diferente da peça: um saldo de
> 10,0 L com utilização de 3,5 L resulta em 6,5 L. Por isso saldo, estoque mínimo e quantidade
> consumida são decimais no insumo, e inteiros na peça.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `estoque:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `nome` | string | Obrigatório; nome do insumo. |
| Body | `descricao` | string | Obrigatória; descrição usada na regra de duplicidade. |
| Body | `categoriaId` | uuid | Obrigatório; categoria ativa do catálogo. |
| Body | `unidadeMedida` | enum | Obrigatória; `UN` \| `L` \| `ML` \| `KG` \| `G` \| `M`. |
| Body | `custoUnitario` | decimal | Obrigatório; não pode ser negativo. |
| Body | `estoqueMinimo` | decimal | Maior ou igual a zero; aceita casas decimais. |

```json
{
  "nome": "Óleo 5W30",
  "descricao": "Óleo sintético 5W30 API SN",
  "categoriaId": "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724",
  "unidadeMedida": "L",
  "custoUnitario": 45.0,
  "estoqueMinimo": 20.0
}
```

O cliente **não** informa `id`, `codigo`, `ativo`, `dataCriacao` nem saldo: são responsabilidade do
sistema ou de outros fluxos.

> **Decisão de projeto.** O `codigo` é **gerado pelo sistema**, no formato `INS-000001`, em
> sequência global, sem reset, com seis dígitos. Antes o cliente enviava o código, e conviviam os
> formatos `IN-0031`, `INS-0012` e `OLEO-001`. Mesmo padrão do `PEC-000001` de Peças e do
> `SER-000001` de Serviços.

> **Decisão de projeto.** O cadastro **não aceita estoque inicial**. `saldoFisicoInicial` saiu do
> contrato: todo saldo entra por movimentação de entrada, que é o único lugar que gera histórico
> auditável. O cadastro de peça já funcionava assim.

> **Decisão de projeto.** A duplicidade é decidida por **descrição normalizada dentro da mesma
> categoria e unidade de medida**, entre insumos **ativos**, por índice parcial.

> **Decisão de projeto.** **Cada unidade de medida é um item de estoque independente, sem
> conversão.** O mesmo óleo em `L` e em `ML` são dois cadastros, com saldos próprios: comprar 1 L
> não aumenta o saldo do item em mililitro, e a baixa de um não afeta o outro. Converter unidades
> exigiria fator de conversão por família, arredondamento e reconciliação de saldo — complexidade
> que o MVP não paga. Na prática, a oficina cadastra o insumo na unidade em que compra e consome.

> **Decisão de projeto.** O insumo mantém **`nome` e `descricao`**, e os dois aparecem no cadastro,
> na consulta e na atualização.

**Validações**

*Técnicas*

- `nome` obrigatório.
- `descricao` obrigatória.
- `categoriaId` obrigatório, no formato uuid.
- `unidadeMedida` obrigatória e pertencente ao enum. Não há conversão entre unidades: cada uma é
  um item independente.
- `custoUnitario` não pode ser negativo.
- `estoqueMinimo` não pode ser negativo.

*Negócio*

- O `codigo` gerado é único no catálogo.
- Não pode existir outro insumo **ativo** com a mesma descrição normalizada — sem acento, sem
  espaço duplo, em minúsculas — na mesma categoria e unidade de medida.
- O cadastro não movimenta estoque e não aceita saldo inicial.

**Processamento**

1. Validar o payload.
2. Carregar a categoria pelo `categoriaId`, validar que existe e está ativa, normalizar a
   descrição e verificar duplicidade entre insumos ativos da mesma categoria e
   unidade de medida.
3. Gerar o `id` técnico e o `codigo` funcional.
4. Criar o insumo com `tipo = INSUMO`.
5. Definir `ativo = true` e saldos zerados.
6. Persistir.
7. Retornar o insumo cadastrado.

**Persistência**

- Consulta: `item_estoque` (verificação de duplicidade e de unicidade do código).
- Altera: `item_estoque` (insert, com `descricao_normalizada` e saldos zerados).
- Não altera: nenhum saldo de estoque nem `movimentacao_estoque`.
- Chave estrangeira `categoria_id` para a tabela `categoria`.
- Índice parcial `UNIQUE (categoria_id, unidade_medida, descricao_normalizada) WHERE ativo = true`.

**Saída da API**

```json
{
  "id": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
  "codigo": "INS-000001",
  "tipo": "INSUMO",
  "nome": "Óleo 5W30",
  "descricao": "Óleo sintético 5W30 API SN",
  "categoriaId": "e4b7a1c6-90d5-4f2b-8a37-1c5e6d09b724",
  "categoria": "Lubrificantes",
  "unidadeMedida": "L",
  "custoUnitario": 45.0,
  "estoqueMinimo": 20.0,
  "saldoFisico": 0.0,
  "ativo": true,
  "dataCriacao": "2026-08-19T19:55:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Insumo cadastrado. |
| `400` | Dados inválidos; unidade fora do enum; custo ou estoque mínimo negativo. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `estoque:escrever`. |
| `409` | Insumo duplicado, ou código já usado no catálogo. |

**Dependências**

- `ItemEstoqueRepository`.
- `MovimentacaoEstoqueRepository` (registro do saldo inicial).
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Cadastra insumo com dados válidos.
- Aceita quantidade decimal no saldo inicial e no estoque mínimo.
- Rejeita estoque negativo.
- Rejeita custo negativo.
- Rejeita unidade de medida inválida.
- Rejeita duplicidade.
- O insumo nasce ativo.

*Integração*

- `POST` válido retorna `201` e persiste o insumo.
- Saldo inicial informado é persistido junto com a movimentação correspondente.
- Código duplicado retorna `409`.
- Unidade inválida retorna `400`.
- Sem token retorna `401` e perfil sem escopo retorna `403`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar a entidade `Insumo` no agregado de item de estoque
- [ ] Definir código, nome, descrição, categoria, unidade de medida, custo unitário, estoque mínimo e situação
- [ ] Garantir que o cadastro nasce com saldos zerados e não aceita estoque inicial
- [ ] Gerar o `codigo` no formato `INS-000001`, em sequência global, sem reset
- [ ] Garantir que saldo, estoque mínimo e consumo aceitem valores decimais
- [ ] Definir a situação inicial como ativa
- [ ] Impedir quantidade negativa

**Caso de uso**

- [ ] Implementar `CadastrarInsumo`
- [ ] Implementar a normalização da descrição para a checagem de duplicidade

**Repositório**

- [ ] Implementar a persistência do insumo em `ItemEstoqueRepository`
- [ ] Implementar a consulta de duplicidade por descrição normalizada, categoria e unidade
- [ ] Criar a tabela `categoria` e a chave estrangeira `categoria_id` na migration
- [ ] Criar o índice parcial `UNIQUE (categoria_id, unidade_medida, descricao_normalizada) WHERE ativo = true` na migration

**Handler HTTP**

- [ ] Implementar `POST /estoque/insumos`
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar `nome` obrigatório
- [ ] Validar `descricao` e `categoriaId` obrigatórios
- [ ] Validar `unidadeMedida` dentro do enum
- [ ] Validar `custoUnitario` não negativo
- [ ] Validar `estoqueMinimo` não negativo
- [ ] Rejeitar `codigo` e saldo enviados pelo cliente
- [ ] Retornar `400` para dados inválidos e `409` para duplicidade

**Testes unitários**

- [ ] Cadastro válido
- [ ] Quantidade decimal aceita
- [ ] Estoque mínimo negativo rejeitado
- [ ] Custo negativo rejeitado
- [ ] Unidade inválida rejeitada
- [ ] Descrição repetida na mesma categoria e unidade rejeitada
- [ ] Descrição igual à de um insumo inativo aceita
- [ ] Situação inicial ativa

**Testes de integração**

- [ ] `201` com o insumo persistido, saldos zerados e `codigo` gerado
- [ ] `409` para duplicidade
- [ ] `400` para dados inválidos
- [ ] `401` sem autenticação e `403` sem permissão

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
