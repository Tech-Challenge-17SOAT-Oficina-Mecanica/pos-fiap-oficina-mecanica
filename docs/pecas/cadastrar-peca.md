---
documento: Refinamento de Requisitos — Cadastrar Peça
dono: A definir
versao: 0.3
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Peça

Este documento detalha a tarefa Cadastrar Peça do contexto de Peças.

## 1 · Cadastrar Peça

### 1.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Cadastrar uma nova peça no catálogo da oficina, permitindo sua identificação, consulta e
utilização posterior nas Ordens de Serviço e nos orçamentos.

**Problema**

A oficina precisa manter um catálogo estruturado das peças utilizadas nos atendimentos. Sem um
cadastro centralizado existe risco de duplicidade, informação inconsistente e dificuldade para
identificar corretamente as peças usadas nas Ordens de Serviço. O cadastro também precisa
preservar uma identificação estável para a peça, permitindo que ela seja referenciada depois sem
comprometer seu histórico.

**Pré-condições**

- O usuário deve estar autenticado.
- O usuário deve possuir permissão para cadastrar peças.
- Os dados obrigatórios da peça devem ser informados.
- Não deve existir outra peça ativa com a mesma descrição normalizada dentro da mesma categoria, conforme a regra
  definida pela oficina.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-01 | Permitir cadastrar uma nova peça. |
| RF-PEC-123 | Permitir informar a categoria, pelo identificador, e o estoque mínimo da peça. |
| RF-PEC-02 | Permitir informar o nome da peça. |
| RF-PEC-03 | Permitir informar a descrição da peça. |
| RF-PEC-04 | Permitir informar o fabricante, quando aplicável. |
| RF-PEC-05 | Permitir informar o valor atual da peça. |
| RF-PEC-06 | Gerar automaticamente um código único para a peça. |
| RF-PEC-07 | Gerar uma identificação técnica única para a peça. |
| RF-PEC-08 | Registrar a peça inicialmente como ativa. |
| RF-PEC-09 | Disponibilizar a peça para consultas após o cadastro. |
| RF-PEC-10 | Permitir que a peça seja utilizada em Ordens de Serviço e orçamentos enquanto estiver ativa. |
| RF-PEC-11 | Manter o código da peça estável após sua criação. |
| RF-PEC-12 | Não registrar estoque inicial durante o cadastro da peça. |
| RF-PEC-13 | Permitir que o estoque da peça seja controlado depois, pelos fluxos de movimentação de estoque. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-01 | O cadastro deve ser persistido de forma consistente. |
| RNF-PEC-02 | O código da peça deve ser único no catálogo. |
| RNF-PEC-03 | O identificador técnico da peça deve ser único e imutável. |
| RNF-PEC-04 | O valor monetário deve usar representação decimal adequada. |
| RNF-PEC-05 | Somente usuários autorizados devem poder cadastrar peças. |
| RNF-PEC-06 | O cadastro deve manter rastreabilidade da data de criação e do usuário responsável. |
| RNF-PEC-07 | O cadastro da peça não deve alterar o estoque automaticamente. |
| RNF-PEC-08 | O cadastro deve preservar a separação entre identificação técnica (`id`) e identificação funcional (`codigo`). |

**Fluxo Principal**

1. O gestor acessa o cadastro de peças.
2. O sistema apresenta o formulário de cadastro.
3. O gestor informa o nome da peça.
4. O gestor informa a descrição, quando aplicável.
5. O gestor informa o fabricante, quando aplicável.
6. O gestor informa o valor atual da peça.
7. O gestor solicita o cadastro.
8. O sistema valida os dados informados.
9. O sistema verifica se não existe peça duplicada, conforme a regra definida.
10. O sistema gera o identificador técnico da peça.
11. O sistema gera o código funcional da peça.
12. O sistema define a situação inicial como ativa.
13. O sistema registra a data e hora da criação.
14. O sistema associa o usuário responsável pelo cadastro.
15. O sistema persiste a peça.
16. O sistema confirma o cadastro e apresenta os dados da peça criada.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Dados obrigatórios não informados | Solicita o preenchimento dos campos necessários. |
| A2 | Valor inválido | Impede o cadastro caso o valor seja negativo ou inválido. |
| A3 | Peça duplicada | Informa que já existe peça cadastrada com os mesmos critérios de identificação. |
| A4 | Usuário sem autorização | Impede a operação. |
| A5 | Falha na geração do código | Não conclui o cadastro até que seja possível gerar um código válido e único. |
| A6 | Falha na persistência | Não considera a peça cadastrada até que a operação seja concluída com sucesso. |
| A7 | Cadastro realizado sem estoque | Cria a peça normalmente; a quantidade em estoque será registrada depois, por uma movimentação de estoque. |

**Saída**

- Peça cadastrada, com identificador técnico, código funcional, dados cadastrais, valor atual e
  situação ativa.
- A peça fica disponível no catálogo para consultas e utilização em novos atendimentos.

**Pós-condições**

- A peça fica registrada no sistema, com `id` técnico único e `codigo` funcional único gerado pelo sistema.
- A peça fica ativa e pode ser utilizada em novos orçamentos e Ordens de Serviço.
- O cadastro não cria nem altera estoque; a movimentação acontece depois, pelo fluxo específico
  de entrada e saída.
- O histórico da peça passa a existir a partir do momento do cadastro.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /estoque/pecas
```

> **Decisão de projeto.** A peça tem dois identificadores com responsabilidades diferentes, e
> eles não são o mesmo atributo. O `id` é técnico: gerado pelo sistema, UUID, único, imutável, e
> é o que aparece nas referências entre entidades. O `codigo` é funcional: gerado pelo sistema,
> único no catálogo, e é o que o negócio usa para identificar e buscar a peça. A mesma separação
> vale para insumo, e está descrita em [consultar-pecas.md](consultar-pecas.md).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `estoque:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `nome` | string | Obrigatório; nome da peça. |
| Body | `descricao` | string | Obrigatória; respeita o limite de tamanho definido. |
| Body | `categoriaId` | uuid | Obrigatório; categoria ativa do catálogo. Usada na regra de duplicidade. |
| Body | `fornecedorId` | uuid | Opcional; fornecedor padrão ativo para compra automática. |
| Body | `fabricante` | string | Opcional; respeita o limite de tamanho definido. |
| Body | `precoVenda` | decimal | Obrigatório; maior ou igual a zero. |
| Body | `estoqueMinimo` | int | Opcional; inteiro maior ou igual a zero. Default `0`. |

```json
{
  "nome": "Pastilha de freio",
  "descricao": "Pastilha de freio dianteira",
  "categoriaId": "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94",
  "fornecedorId": "60000000-0000-0000-0000-000000000001",
  "fabricante": "Fabricante X",
  "precoVenda": 180.0,
  "estoqueMinimo": 4
}
```

O cliente **não** informa `id`, `codigo`, `ativo`, `dataCriacao` nem saldo de estoque: esses
dados são responsabilidade do sistema ou de outros fluxos.

> **Decisão de projeto.** A peça mantém **`nome` e `descricao`**: `nome` é o termo curto que o
> mecânico usa no balcão, `descricao` é o detalhamento que sai no orçamento. Os dois aparecem no
> cadastro, na consulta e na atualização — não faz sentido gravar um campo que some depois.

> **Decisão de projeto.** O `codigo` segue **um formato só**: `PEC-000001`, gerado pelo sistema em
> sequência global, sem reset, com seis dígitos. Os formatos antigos `PC-0142` e variações saem da
> documentação. Mesmo padrão do `SER-000001` de Serviços e do `INS-000001` de Insumos.

> **Decisão de projeto.** A duplicidade é decidida por **descrição normalizada dentro da mesma
> categoria**, entre peças **ativas**, por índice parcial. O `codigo` é gerado pelo sistema e
> sempre único, então não serve de critério para barrar cadastro repetido. Foi preciso trazer
> `categoria` para o payload do cadastro, que antes só a tinha na atualização.

> **Decisão de projeto — D-09.** A categoria deixou de ser texto livre e virou **tabela**,
> referenciada por `categoriaId`. A escrita informa o identificador; a leitura devolve o
> identificador e o nome. Sem isso, "Freios", "freios" e "Sistema de freios" seriam três
> categorias e a regra de duplicidade não pegaria nada.

> **Decisão de projeto.** O cadastro **não aceita estoque inicial**, e o de insumo também não. Todo
> saldo entra por movimentação de entrada, que é o único lugar que gera histórico auditável.

**Validações**

*Técnicas*

- `nome` obrigatório.
- `descricao` obrigatória, dentro do limite de tamanho definido.
- `categoriaId` obrigatório, no formato uuid.
- `fornecedorId`, quando informado, no formato uuid.
- `precoVenda` maior ou igual a zero.
- `estoqueMinimo`, quando informado, inteiro maior ou igual a zero.
- `fabricante`, quando informado, dentro do limite de tamanho definido.

*Negócio*

- Não pode existir outra peça **ativa** com a mesma descrição normalizada — sem acento, sem espaço
  duplo, em minúsculas — dentro da mesma categoria.
- O fornecedor padrão informado deve existir e estar ativo.
- O `codigo` gerado deve ser único no catálogo.
- O cadastro não movimenta estoque e não aceita saldo inicial.

**Regra de domínio**

```
dados válidos → criar peça → gerar id técnico → gerar código funcional → definir como ativa → peça cadastrada
```

O estoque é tratado por um fluxo separado:

```
cadastrar peça → peça ativa → registrar entrada de estoque → quantidade disponível
```

**Processamento**

1. Receber o payload e identificar o usuário autenticado.
2. Validar os dados de entrada e a autorização.
3. Carregar a categoria pelo `categoriaId` e validar que existe e está ativa.
4. Normalizar a descrição e verificar duplicidade entre peças ativas da mesma categoria.
4. Gerar o `id` técnico.
5. Gerar o `codigo` funcional.
6. Criar a entidade `Peca` com `tipo = PECA`.
7. Definir `ativo = true`.
8. Registrar data e hora de criação e o usuário responsável.
9. Persistir a peça.
10. Retornar a peça criada.

**Persistência**

- Consulta: `item_estoque` (verificação de duplicidade e de unicidade do código).
- Altera: `item_estoque` (insert de `id`, `codigo`, `nome`, `descricao`, `descricao_normalizada`,
  `categoria_id`, `fabricante`, `preco_venda`, `estoque_minimo`, `ativo`, `data_criacao`,
  `usuario_criacao`).
- Não altera: nenhum saldo de estoque.

Restrições recomendadas: `id` como PRIMARY KEY; `codigo` como `UNIQUE NOT NULL`; `nome`,
`descricao`, `categoria_id`, `preco_venda` e `ativo` como `NOT NULL`; chave estrangeira para
`categoria`; índice parcial `UNIQUE (categoria_id, descricao_normalizada) WHERE ativo = true`.

**Saída da API**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "codigo": "PEC-000001",
  "tipo": "PECA",
  "nome": "Pastilha de freio",
  "descricao": "Pastilha de freio dianteira",
  "categoriaId": "7c1b4d09-2f83-4a51-9e6c-3d0a75b21e94",
  "categoria": "Freios",
  "fabricante": "Fabricante X",
  "precoVenda": 180.0,
  "estoqueMinimo": 4,
  "ativo": true,
  "dataCriacao": "2026-08-19T19:50:00-03:00"
}
```

A resposta traz informações que não vieram na requisição — `id`, `codigo`, `ativo` e
`dataCriacao` — porque são geradas pelo sistema.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Peça cadastrada com sucesso. |
| `400` | Dados de entrada inválidos, ou categoria inexistente ou inativa. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `estoque:escrever`. |
| `409` | Já existe peça ativa com a mesma descrição na categoria, ou conflito na geração do código. |

**Dependências**

- `ItemEstoqueRepository`.
- Componente responsável pela geração do código funcional, quando separado.
- Middleware de autenticação/autorização.
- Trilha de auditoria, quando adotada.

O cadastro não depende do fluxo de estoque para ser concluído.

**Testes**

*Unitários*

- Cadastra peça com dados válidos.
- Gera o `id` técnico e o `codigo` funcional.
- O código gerado é único.
- A peça nasce ativa.
- Persiste a data de criação.
- Rejeita nome vazio.
- Rejeita valor negativo.
- Impede descrição repetida na mesma categoria entre peças ativas.
- Aceita descrição igual à de uma peça inativa.
- Normaliza a descrição antes de comparar.
- Não exige estoque inicial.

*Integração*

- `POST` válido retorna `201` com a peça persistida, incluindo `id` e `codigo` gerados.
- Payload inválido retorna `400`.
- Descrição repetida na categoria retorna `409`.
- Descrição igual à de peça inativa retorna `201`.
- Sem token retorna `401` e perfil sem escopo retorna `403`.

*Regressão*

- O cadastro não altera nenhum saldo de estoque.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar a entidade `Peca` no agregado de item de estoque
- [ ] Definir o `id` técnico como identificador da entidade, gerado pelo sistema em formato UUID
- [ ] Definir o `codigo` como identificador funcional do catálogo, gerado pelo sistema
- [ ] Gerar o `codigo` no formato `PEC-000001`, em sequência global, sem reset
- [ ] Garantir a unicidade do `id` e do `codigo`
- [ ] Definir a situação inicial como ativa
- [ ] Definir o valor monetário com precisão decimal
- [ ] Definir os campos obrigatórios e opcionais
- [ ] Criar o método de domínio de criação da peça
- [ ] Implementar a normalização da descrição (sem acento, sem espaço duplo, minúscula)
- [ ] Validar duplicidade de descrição normalizada na categoria, entre peças ativas
- [ ] Criar a tabela `categoria` e a chave estrangeira `categoria_id` na migration
- [ ] Criar o índice parcial `UNIQUE (categoria_id, descricao_normalizada) WHERE ativo = true` na migration
- [ ] Garantir que o cadastro não inclui estoque inicial nem movimenta estoque

**Caso de uso**

- [ ] Implementar `CadastrarPeca`
- [ ] Registrar data, hora de criação e usuário responsável

**Repositório**

- [ ] Implementar a persistência da peça em `ItemEstoqueRepository`
- [ ] Implementar a consulta de duplicidade e a verificação de unicidade do código

**Handler HTTP**

- [ ] Implementar `POST /estoque/pecas`
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados
- [ ] Retornar `id`, `codigo` e situação inicial na resposta

**Validações**

- [ ] Validar `nome` obrigatório
- [ ] Validar `precoVenda` maior ou igual a zero
- [ ] Validar os limites de tamanho de `descricao` e `fabricante`
- [ ] Retornar `400` para dados inválidos
- [ ] Retornar `409` para duplicidade ou conflito na geração do código

**Testes unitários**

- [ ] Cadastro válido
- [ ] Nome obrigatório
- [ ] Valor negativo
- [ ] Payload inválido
- [ ] Duplicidade
- [ ] Geração do `id` e do `codigo`
- [ ] Unicidade do código
- [ ] Situação inicial ativa
- [ ] Estoque não alterado

**Testes de integração**

- [ ] `201` com persistência da peça no banco
- [ ] Resposta trazendo `id` e `codigo` gerados
- [ ] `400` para dados inválidos e `409` para duplicidade
- [ ] `401` sem autenticação e `403` sem permissão

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI
- [ ] Documentar a separação entre `id` técnico e `codigo` funcional
- [ ] Documentar que o estoque é responsabilidade de outro fluxo

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
