---
documento: Refinamento de Requisitos — Cadastrar Fornecedor
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Fornecedor

Este documento detalha a tarefa Cadastrar Fornecedor do contexto de Peças.

> **Escopo.** Fornecedor pertence ao agregado de **Compras**, cujo dono é o contexto de Peças. O
> mesmo cadastro atende os pedidos de compra de peça e de insumo — Insumos apenas o referencia.

## 10 · Cadastrar Fornecedor

### 10.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar no sistema o fornecedor de quem a oficina compra peças e insumos, para que o pedido de
compra tenha uma contra-parte identificada.

**Problema**

Hoje o pedido de compra exige fornecedor, mas não existe cadastro: o nome do fornecedor vive na
memória de quem comprou. Quando a peça atrasa, ninguém sabe para quem ligar; quando chega errada,
ninguém sabe de quem cobrar; e o mesmo fornecedor aparece escrito de três formas diferentes em
pedidos distintos.

**Pré-condições**

- O usuário deve estar autorizado a manter o cadastro de compras.
- O fornecedor deve possuir CNPJ ou CPF válido.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-88 | Permitir cadastrar um fornecedor. |
| RF-PEC-89 | Registrar razão social e nome fantasia do fornecedor. |
| RF-PEC-90 | Registrar o documento do fornecedor, CNPJ ou CPF. |
| RF-PEC-91 | Registrar ao menos um contato: telefone ou e-mail. |
| RF-PEC-92 | Validar o documento informado. |
| RF-PEC-93 | Impedir cadastro duplicado para o mesmo documento entre fornecedores ativos. |
| RF-PEC-94 | Registrar o fornecedor como ativo. |
| RF-PEC-95 | Disponibilizar o fornecedor para uso nos pedidos de compra. |
| RF-PEC-124 | Registrar o prazo de entrega do fornecedor, em dias, com padrão de 7 quando não informado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-55 | A operação deve ser feita por API RESTful. |
| RNF-PEC-56 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-PEC-57 | O cadastro deve manter rastreabilidade da data de criação e do usuário responsável. |
| RNF-PEC-58 | O documento deve ser armazenado apenas com dígitos, sem máscara. |

**Fluxo Principal**

1. O mecânico solicita o cadastro de um fornecedor.
2. O mecânico informa razão social, documento e contato.
3. O sistema valida os campos obrigatórios e o documento.
4. O sistema verifica se já existe fornecedor ativo com o mesmo documento.
5. O sistema registra o fornecedor como ativo.
6. O sistema confirma o cadastro e devolve o fornecedor criado.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Documento inválido | Impede o cadastro e informa que o CNPJ ou CPF não é válido. |
| A2 | Fornecedor já cadastrado e ativo com o mesmo documento | Impede o cadastro e devolve o fornecedor existente. |
| A3 | Documento igual ao de um fornecedor inativo | Aceita o cadastro: a unicidade vale apenas entre ativos. |
| A4 | Nenhum contato informado | Impede o cadastro e informa que telefone ou e-mail é obrigatório. |
| A5 | Usuário sem autorização | Impede a operação. |

**Saída**

- Fornecedor cadastrado e disponível para uso nos pedidos de compra.
- Ou indicação do motivo pelo qual o cadastro foi recusado.

**Pós-condições**

- O fornecedor passa a existir no cadastro da oficina, ativo.
- O fornecedor pode ser informado em `POST /compras/pedidos`.
- Nenhum saldo de estoque é alterado.

---

### 10.2 Refinamento Técnico

**Endpoint**

```http
POST /fornecedores
```

> **Decisão de projeto.** O recurso fica na raiz, `/fornecedores`, e não sob `/compras`, porque o
> fornecedor existe independentemente de haver pedido — é cadastro, não item do pedido. A rota
> pertence ao contexto de Peças, dono do agregado de Compras.

> **Decisão de projeto.** A unicidade do documento vale **apenas entre fornecedores ativos**, por
> índice parcial `UNIQUE (documento) WHERE ativo = true` — mesma regra já usada em cliente, veículo
> e catálogo de serviços.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `compras:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `razaoSocial` | string | Obrigatória; 3 a 120 caracteres. |
| Body | `nomeFantasia` | string | Opcional; até 120 caracteres. |
| Body | `documento` | string | Obrigatório; CNPJ ou CPF, somente dígitos. |
| Body | `tipoDocumento` | enum | Obrigatório; `CNPJ` \| `CPF`. |
| Body | `telefone` | string | Somente dígitos. Obrigatório se `email` não for informado. |
| Body | `email` | string | Obrigatório se `telefone` não for informado. |
| Body | `prazoEntregaDias` | inteiro | Opcional; de 1 a 365. Padrão 7 quando ausente. |

```json
{
  "razaoSocial": "Auto Peças Bandeirantes Ltda",
  "nomeFantasia": "Bandeirantes Peças",
  "documento": "12345678000190",
  "tipoDocumento": "CNPJ",
  "telefone": "1133224455",
  "email": "vendas@bandeirantes.com.br",
  "prazoEntregaDias": 5
}
```

O cliente não informa `id`, `ativo` nem `dataCriacao`: são gerados pelo sistema.

> **Decisão de projeto — D-07.** O prazo de entrega do fornecedor entrou com nome em português,
> **`prazoEntregaDias`**, e padrão de 7 dias quando não vier preenchido. Ele é informativo, para
> quem compra: a sugestão de quantidade fechada na D-06 não depende dele.

**Validações**

*Técnicas*

- `razaoSocial` obrigatória, de 3 a 120 caracteres.
- `documento` obrigatório, somente dígitos, com 11 ou 14 posições.
- `tipoDocumento` obrigatório e pertencente ao enum.
- `telefone`, quando informado, com 10 ou 11 dígitos.
- `email`, quando informado, em formato válido.
- `prazoEntregaDias`, quando informado, inteiro de 1 a 365.

*Negócio*

- O `documento` deve ser válido conforme o `tipoDocumento`.
- Não pode existir outro fornecedor **ativo** com o mesmo documento.
- Pelo menos um entre `telefone` e `email` deve ser informado.

**Processamento**

1. Receber o payload e identificar o usuário autenticado.
2. Normalizar o documento, removendo máscara.
3. Validar os campos obrigatórios e o documento.
4. Verificar se já existe fornecedor ativo com o mesmo documento.
5. Gerar o `id`.
6. Criar o fornecedor com `ativo = true` e `prazoEntregaDias`, usando 7 quando não vier informado.
7. Registrar data, hora de criação e usuário responsável.
8. Persistir e retornar o fornecedor criado.

**Persistência**

- Consulta: `fornecedor` (verificação de duplicidade entre ativos).
- Altera: `fornecedor` (insert de `id`, `razao_social`, `nome_fantasia`, `documento`,
  `tipo_documento`, `telefone`, `email`, `prazo_entrega_dias`, `ativo`, `data_criacao`,
  `usuario_criacao`).
- Não altera: `pedido_compra`, `item_estoque` nem qualquer saldo.
- Índice parcial `UNIQUE (documento) WHERE ativo = true`.

**Saída da API**

```json
{
  "id": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
  "razaoSocial": "Auto Peças Bandeirantes Ltda",
  "nomeFantasia": "Bandeirantes Peças",
  "documento": "12345678000190",
  "tipoDocumento": "CNPJ",
  "telefone": "1133224455",
  "email": "vendas@bandeirantes.com.br",
  "prazoEntregaDias": 5,
  "ativo": true,
  "version": 1,
  "dataCriacao": "2026-08-22T09:10:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Fornecedor cadastrado com sucesso. |
| `400` | Dados obrigatórios ausentes, documento inválido ou nenhum contato informado. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `compras:escrever`. |
| `409` | Já existe fornecedor ativo com o mesmo documento. |

**Dependências**

- `FornecedorRepository`.
- Validador de CNPJ e CPF.
- Middleware de autenticação e autorização.
- Trilha de auditoria.

**Testes**

*Unitários*

- Cadastra fornecedor com dados válidos.
- Normaliza o documento, removendo máscara.
- Rejeita documento inválido.
- Rejeita cadastro sem telefone e sem e-mail.
- Rejeita documento já usado por fornecedor ativo.
- Aceita documento igual ao de fornecedor inativo.
- Nasce com `ativo = true`.

*Integração*

- `POST /fornecedores` válido retorna `201` com o fornecedor persistido.
- Documento inválido retorna `400`.
- Cadastro sem contato retorna `400`.
- Documento duplicado entre ativos retorna `409`.
- Sem token retorna `401`; perfil sem escopo retorna `403`.
- Fornecedor cadastrado pode ser usado em `POST /compras/pedidos`.

---

### 10.3 Checklist de Implementação

**Domínio**

- [ ] Criar a entidade `Fornecedor` com razão social, nome fantasia, documento, contato e situação
- [ ] Definir o `id` como UUID gerado pelo sistema
- [ ] Implementar a validação de CNPJ e CPF
- [ ] Implementar a normalização do documento, removendo máscara
- [ ] Exigir ao menos um contato entre telefone e e-mail
- [ ] Aplicar o padrão de 7 dias em `prazoEntregaDias` quando o campo não vier
- [ ] Definir a situação inicial como ativa
- [ ] Adicionar o campo `version` para controle otimista

**Caso de uso**

- [ ] Implementar `CadastrarFornecedor`
- [ ] Verificar duplicidade de documento entre fornecedores ativos
- [ ] Registrar data, hora de criação e usuário responsável

**Repositório**

- [ ] Criar `FornecedorRepository`
- [ ] Criar método para consultar fornecedor por documento entre ativos
- [ ] Criar método para salvar novo fornecedor
- [ ] Criar o índice parcial `UNIQUE (documento) WHERE ativo = true` na migration

**Handler HTTP**

- [ ] Implementar `POST /fornecedores`
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização pelo escopo `compras:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar `razaoSocial` obrigatória, de 3 a 120 caracteres
- [ ] Validar `documento` e `tipoDocumento` obrigatórios
- [ ] Validar o documento conforme o tipo
- [ ] Validar que ao menos um contato foi informado
- [ ] Retornar `409` para documento já usado por fornecedor ativo

**Auditoria**

- [ ] Registrar o cadastro na trilha de auditoria

**Testes unitários**

- [ ] Cadastro válido
- [ ] Documento inválido
- [ ] Documento duplicado entre ativos
- [ ] Documento igual ao de fornecedor inativo aceito
- [ ] Cadastro sem contato rejeitado

**Testes de integração**

- [ ] Endpoint cadastra fornecedor válido e retorna `201`
- [ ] Endpoint retorna `400` para dados inválidos
- [ ] Endpoint retorna `409` para documento duplicado entre ativos
- [ ] Endpoint retorna `401` sem autenticação e `403` sem permissão

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
