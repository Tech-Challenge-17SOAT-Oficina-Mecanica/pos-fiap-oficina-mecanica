---
documento: Refinamento de Requisitos — Cadastrar Insumo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Insumo

Este documento detalha a tarefa Cadastrar Insumo do contexto de Peças & Insumos.

## 12 · Cadastrar Insumo

### 12.1 Refinamento de Produto

**Persona**

Gestor.

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
| RF-EST-80 | Permitir cadastrar um insumo. |
| RF-EST-81 | Registrar o nome e a descrição do insumo. |
| RF-EST-82 | Registrar a unidade de medida. |
| RF-EST-83 | Registrar o custo unitário. |
| RF-EST-84 | Registrar a quantidade disponível como estoque inicial, quando informada. |
| RF-EST-85 | Permitir definir as informações necessárias para o controle de estoque, incluindo o estoque mínimo. |
| RF-EST-86 | Validar os dados informados. |
| RF-EST-87 | Impedir duplicidade do insumo. |
| RF-EST-88 | Disponibilizar o insumo para utilização em Ordens de Serviço. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-EST-61 | O cadastro deve ser persistido de forma consistente. |
| RNF-EST-62 | O controle de estoque deve evitar saldos incorretos. |
| RNF-EST-63 | Os valores monetários devem ser armazenados de forma adequada. |
| RNF-EST-64 | Somente usuário autorizado poderá realizar o cadastro. |
| RNF-EST-65 | A operação deve preservar o histórico das movimentações de estoque. |
| RNF-EST-66 | O cadastro não deve alterar outros insumos já existentes. |

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

### 12.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/insumos
```

> **Decisão de projeto.** Insumo admite **quantidade fracionada**, diferente da peça: um saldo de
> 10,0 L com utilização de 3,5 L resulta em 6,5 L. Por isso saldo, estoque mínimo e quantidade
> consumida são decimais no insumo, e inteiros na peça.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `estoque:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `codigo` | string | Código funcional do insumo; único no catálogo. |
| Body | `nome` | string | Obrigatório; nome do insumo. |
| Body | `descricao` | string | Opcional; descrição complementar. |
| Body | `categoria` | string | Categoria do insumo. |
| Body | `unidadeMedida` | enum | Obrigatória; `UN` \| `L` \| `ML` \| `KG` \| `G` \| `M`. |
| Body | `custoUnitario` | decimal | Obrigatório; não pode ser negativo. |
| Body | `estoqueMinimo` | decimal | Maior ou igual a zero; aceita casas decimais. |
| Body | `saldoFisicoInicial` | decimal | Opcional; maior ou igual a zero; aceita casas decimais. |

```json
{
  "codigo": "OLEO-001",
  "nome": "Óleo 5W30",
  "descricao": "Óleo sintético",
  "categoria": "Lubrificantes",
  "unidadeMedida": "L",
  "custoUnitario": 45.0,
  "estoqueMinimo": 20.0,
  "saldoFisicoInicial": 5.0
}
```

**Validações**

*Técnicas*

- `nome` obrigatório.
- `unidadeMedida` obrigatória e pertencente ao enum.
- `custoUnitario` não pode ser negativo.
- `saldoFisicoInicial` e `estoqueMinimo` não podem ser negativos.

*Negócio*

- `codigo` único no catálogo.
- Não pode existir insumo duplicado, conforme a regra de identificação adotada.

**Processamento**

1. Validar o payload.
2. Verificar duplicidade de código e de insumo equivalente.
3. Criar o insumo com `tipo = INSUMO`.
4. Definir `ativo = true`.
5. Registrar o estoque inicial, quando informado, com a movimentação correspondente.
6. Persistir.
7. Retornar o insumo cadastrado.

**Persistência**

- Consulta: `item_estoque` (verificação de duplicidade).
- Altera: `item_estoque` (insert), `movimentacao_estoque` (insert do saldo inicial, quando informado).
- O insumo e seu estoque inicial são persistidos na mesma transação.

**Saída da API**

```json
{
  "id": "c48e7d05-2a19-4b63-9f27-6e5a1c930b48",
  "codigo": "OLEO-001",
  "tipo": "INSUMO",
  "nome": "Óleo 5W30",
  "descricao": "Óleo sintético",
  "categoria": "Lubrificantes",
  "unidadeMedida": "L",
  "custoUnitario": 45.0,
  "estoqueMinimo": 20.0,
  "saldoFisico": 5.0,
  "ativo": true,
  "dataCriacao": "2026-08-19T19:55:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Insumo cadastrado. |
| `400` | Dados inválidos; unidade fora do enum; custo ou estoque negativo. |
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

### 12.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar a entidade `Insumo` no agregado de item de estoque
- [ ] Definir código, nome, descrição, categoria, unidade de medida, custo unitário, estoque mínimo e situação
- [ ] Definir o estoque inicial como dado opcional do cadastro
- [ ] Garantir que saldo, estoque mínimo e consumo aceitem valores decimais
- [ ] Definir a situação inicial como ativa
- [ ] Impedir quantidade negativa

**Caso de uso**

- [ ] Implementar `CadastrarInsumo`
- [ ] Implementar o registro do estoque inicial, quando informado

**Repositório**

- [ ] Implementar a persistência do insumo em `ItemEstoqueRepository`
- [ ] Implementar a consulta de duplicidade e a verificação de unicidade do código
- [ ] Registrar a movimentação do estoque inicial

**Handler HTTP**

- [ ] Implementar `POST /estoque/insumos`
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar `nome` obrigatório
- [ ] Validar `codigo` único
- [ ] Validar `unidadeMedida` dentro do enum
- [ ] Validar `custoUnitario` não negativo
- [ ] Validar `estoqueMinimo` e `saldoFisicoInicial` não negativos
- [ ] Retornar `400` para dados inválidos e `409` para duplicidade

**Testes unitários**

- [ ] Cadastro válido
- [ ] Quantidade decimal aceita
- [ ] Estoque negativo rejeitado
- [ ] Custo negativo rejeitado
- [ ] Unidade inválida rejeitada
- [ ] Código duplicado rejeitado
- [ ] Situação inicial ativa

**Testes de integração**

- [ ] `201` com persistência do insumo e do saldo inicial
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
