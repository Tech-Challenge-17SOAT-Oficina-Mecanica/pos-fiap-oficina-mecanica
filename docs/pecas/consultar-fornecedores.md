---
documento: Refinamento de Requisitos — Consultar Fornecedores
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Consultar Fornecedores

Este documento detalha a tarefa Consultar Fornecedores do contexto de Peças.

> **Escopo.** Fornecedor pertence ao agregado de **Compras**, cujo dono é o contexto de Peças.

## 11 · Consultar Fornecedores

### 11.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Localizar um fornecedor pelo nome ou pelo documento, para informá-lo no pedido de compra e para
saber com quem falar quando o pedido atrasa.

**Problema**

Sem consulta, o cadastro de fornecedor é uma caixa-preta: o mecânico não sabe se o fornecedor já
existe, cadastra de novo, e o mesmo CNPJ acaba com dois registros. Na hora de cobrar um pedido,
ninguém acha o telefone.

**Pré-condições**

- O usuário deve estar autorizado a consultar o cadastro de compras.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-96 | Permitir listar os fornecedores cadastrados. |
| RF-PEC-97 | Permitir consultar um fornecedor pelo identificador. |
| RF-PEC-98 | Permitir filtrar por razão social ou nome fantasia, parcial. |
| RF-PEC-99 | Permitir filtrar pelo documento exato. |
| RF-PEC-100 | Ocultar fornecedores inativos por padrão. |
| RF-PEC-101 | Apresentar os dados necessários para contato e para o pedido de compra. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-59 | A operação é somente leitura e não altera nenhum dado. |
| RNF-PEC-60 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-PEC-61 | A listagem deve ser paginada. |

**Fluxo Principal**

1. O mecânico informa o filtro desejado.
2. O sistema valida os parâmetros e a paginação.
3. O sistema consulta os fornecedores, ocultando os inativos por padrão.
4. O sistema devolve os fornecedores encontrados.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhum fornecedor encontrado | Retorna `200` com `data` vazio. |
| A2 | Identificador inexistente na consulta por id | Retorna `404`. |
| A3 | `tamanho` acima do teto | Retorna `400`. |
| A4 | Usuário sem autorização | Impede a operação. |

**Saída**

- Relação de fornecedores com razão social, documento, contato e situação.
- Ou o fornecedor específico, na consulta por identificador.

**Pós-condições**

- Nenhum dado é alterado.

---

### 11.2 Refinamento Técnico

**Endpoint**

```http
GET /fornecedores
GET /fornecedores/{fornecedorId}
```

O primeiro lista com filtros e paginação; o segundo devolve um fornecedor específico.

> **Decisão de projeto.** A listagem usa o envelope padrão do projeto — `data`, `pagina`,
> `tamanho`, `totalElementos` e `totalPaginas` — e o recurso único vai direto, sem envelope (D-21).
> A consulta por identificador expõe `version`, que a atualização envia no `If-Match`.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `MECANICO`.
- Escopo: `compras:ler`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Query | `nome` | string | Filtro parcial por razão social ou nome fantasia. |
| Query | `documento` | string | Filtro exato, somente dígitos. |
| Query | `incluirInativos` | boolean | Default `false`. |
| Query | `pagina` | int | Default `0`. |
| Query | `tamanho` | int | Default `20`, máximo `50`. |
| Path | `fornecedorId` | uuid | Obrigatório na consulta individual. |

```http
GET /fornecedores?nome=bandeirantes&incluirInativos=false&pagina=0&tamanho=20
```

A operação não recebe corpo.

**Validações**

- `fornecedorId`, quando informado, em formato UUID válido.
- `documento`, quando informado, somente dígitos.
- `pagina` maior ou igual a zero.
- `tamanho` maior que zero e no máximo `50`.

**Processamento**

1. Receber e validar os parâmetros.
2. Aplicar os filtros informados, ocultando inativos quando `incluirInativos` for `false`.
3. Consultar o `FornecedorRepository`.
4. Montar e devolver a resposta.

**Persistência**

- Consulta: `fornecedor`.
- Altera: nada.

**Saída da API**

Listagem:

```json
{
  "data": [
    {
      "id": "a17d3e92-5c48-4b60-9f31-2e6a8d045cb7",
      "razaoSocial": "Auto Peças Bandeirantes Ltda",
      "nomeFantasia": "Bandeirantes Peças",
      "documento": "12345678000190",
      "tipoDocumento": "CNPJ",
      "telefone": "1133224455",
      "email": "vendas@bandeirantes.com.br",
      "prazoEntregaDias": 5,
      "ativo": true
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 1,
  "totalPaginas": 1
}
```

Consulta por identificador:

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
  "version": 3,
  "dataCriacao": "2026-08-22T09:10:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Consulta realizada, inclusive quando a lista estiver vazia. |
| `400` | Filtro ou paginação inválidos; `tamanho` acima de 50. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `compras:ler`. |
| `404` | Fornecedor não encontrado, na consulta por identificador. |

**Dependências**

- `FornecedorRepository`.
- Middleware de autenticação e autorização.

**Testes**

*Unitários*

- Lista fornecedores cadastrados.
- Filtra por nome parcial.
- Filtra por documento exato.
- Oculta inativos por padrão.
- Inclui inativos quando `incluirInativos=true`.
- Rejeita `tamanho` acima de 50.

*Integração*

- `GET /fornecedores` retorna `200` com o envelope padrão.
- `GET /fornecedores/{fornecedorId}` retorna `200` com o objeto direto e `version`.
- Fornecedor inexistente retorna `404`.
- Lista vazia retorna `200` com `data` vazio.
- Sem token retorna `401`; perfil sem escopo retorna `403`.

---

### 11.3 Checklist de Implementação

**Caso de uso**

- [ ] Implementar `ConsultarFornecedores` com filtros e paginação
- [ ] Implementar `ConsultarFornecedorPorId`
- [ ] Ocultar fornecedores inativos por padrão

**Repositório**

- [ ] Criar método de consulta por filtros e paginação
- [ ] Criar método de consulta por identificador

**Handler HTTP**

- [ ] Implementar `GET /fornecedores`
- [ ] Implementar `GET /fornecedores/{fornecedorId}`
- [ ] Criar DTO/response com o envelope `data`, `pagina`, `tamanho`, `totalElementos` e `totalPaginas`
- [ ] Devolver o objeto direto, sem envelope, na consulta por identificador
- [ ] Expor `version` na consulta por identificador
- [ ] Aplicar autenticação JWT e autorização pelo escopo `compras:ler`

**Validações**

- [ ] Validar formato do `fornecedorId`
- [ ] Validar paginação e o teto de `tamanho` em 50
- [ ] Retornar `404` para fornecedor inexistente

**Testes unitários**

- [ ] Listagem com filtros
- [ ] Consulta por identificador
- [ ] Filtro `incluirInativos`
- [ ] `tamanho` acima do teto

**Testes de integração**

- [ ] Endpoint de listagem retorna `200` com o envelope padrão
- [ ] Endpoint de consulta por identificador retorna `200`
- [ ] Endpoint retorna `404` para fornecedor inexistente
- [ ] Endpoint bloqueia usuário sem permissão

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
