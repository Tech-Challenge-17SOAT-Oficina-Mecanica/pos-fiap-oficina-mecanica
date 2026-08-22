---
documento: Refinamento de Requisitos — Consultar Serviços
dono: João Victor Silva de Oliveira
versao: 0.1
atualizado_em: 2026-08-20
status: rascunho
---

# Refinamento de Requisitos — Consultar Serviços

Este documento detalha a tarefa Consultar Serviços do contexto de Serviços.

## 1 · Consultar Serviços

### 1.1 Refinamento de Produto

**Persona**
Administrador/Gestor da oficina.

**Objetivo**
Consultar os serviços cadastrados no catálogo da oficina para utilização na gestão de Ordens de
Serviço e orçamentos.

**Problema**
A oficina precisa localizar rapidamente os serviços disponíveis, seus valores e seus respectivos
estados para manter a operação organizada. Sem uma consulta estruturada, a composição de
orçamentos e o registro de serviços em Ordens de Serviço ficam sujeitos a erro, retrabalho e uso
de valores desatualizados.

**Pré-condições**

- O usuário deve possuir acesso ao gerenciamento de serviços.
- O sistema deve possuir ou permitir consultar o catálogo de serviços.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-SRV-01 | Permitir listar os serviços cadastrados. |
| RF-SRV-02 | Permitir consultar os dados de um serviço específico. |
| RF-SRV-03 | Permitir identificar serviços ativos e inativos. |
| RF-SRV-04 | Permitir filtrar ou pesquisar serviços por nome, caso aplicável. |
| RF-SRV-05 | Permitir filtrar serviços por status. |
| RF-SRV-06 | Apresentar os dados necessários para identificação do serviço e composição de orçamentos. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SRV-01 | A consulta deve apresentar informações consistentes com os dados persistidos. |
| RNF-SRV-02 | A API deve retornar respostas padronizadas. |
| RNF-SRV-03 | A consulta deve respeitar as regras de autenticação e autorização. |
| RNF-SRV-04 | A operação não deve alterar os dados dos serviços. |
| RNF-SRV-05 | A consulta deve possuir desempenho adequado para o volume esperado do MVP. |

**Fluxo Principal**

1. O administrador acessa o gerenciamento de serviços.
2. O administrador solicita a consulta dos serviços.
3. O sistema verifica a autorização do usuário.
4. O sistema consulta os serviços cadastrados.
5. O sistema aplica filtros ou critérios de busca, caso informados.
6. O sistema retorna os serviços encontrados.
7. O administrador seleciona um serviço, caso queira consultar seus detalhes.
8. O sistema apresenta os dados do serviço.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhum serviço encontrado | O sistema informa que não existem serviços correspondentes aos critérios. |
| A2 | Serviço não encontrado | O sistema informa que o serviço solicitado não existe. |
| A3 | Filtro inválido | O sistema informa os critérios inválidos. |
| A4 | Usuário sem autorização | O sistema impede o acesso. |
| A5 | Falha na consulta | O sistema informa a indisponibilidade e não altera os dados. |

**Saída**

- Lista de serviços; **ou**
- Detalhes de um serviço específico apresentados ao usuário.

**Pós-condições**

- Nenhum dado do catálogo é alterado.
- O usuário obtém as informações necessárias para gestão dos serviços.
- Os serviços consultados podem ser utilizados como referência para novas OS e orçamentos.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
GET /servicos
GET /servicos/{id}
```

O primeiro endpoint lista os serviços cadastrados no catálogo da oficina. O segundo consulta os
dados de um serviço específico.

> **Decisão de projeto.** A consulta por ID foi mantida no mesmo requisito da listagem porque as
> duas operações são somente leitura, compartilham autorização e usam o mesmo recurso
> `/servicos`. A alternativa seria separar em duas tarefas, mas isso duplicaria regras de
> autorização, DTO e testes de leitura.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `GESTOR`.
- Escopo: `servicos:ler`.
- Caso clientes também possam consultar o catálogo, a autorização deve ser definida
  separadamente.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Query | `nome` | string | Filtro opcional por nome parcial do serviço. |
| Query | `status` | enum | Filtro opcional por `ATIVO` ou `INATIVO`. |
| Query | `page` | int | Página da listagem. Default `0`. |
| Query | `size` | int | Tamanho da página. Default `20`. |
| Path | `id` | UUID | Identificador do serviço, obrigatório na consulta individual. |

Exemplo:

```http
GET /servicos?nome=oleo&status=ATIVO&page=0&size=20
```

A operação não recebe corpo.

**Validações**

- `id`, quando informado, deve possuir formato válido de UUID.
- `status`, quando informado, deve ser `ATIVO` ou `INATIVO`.
- `page` deve ser maior ou igual a zero.
- `size` deve ser maior que zero e respeitar o limite máximo definido pelo projeto.
- Para consulta por ID, o serviço deve existir.

**Processamento**

1. Receber os parâmetros da requisição.
2. Identificar o usuário autenticado.
3. Validar autorização.
4. Validar formato do `id`, quando informado.
5. Validar filtros e paginação.
6. Aplicar os filtros informados.
7. Consultar o `ServicoRepository`.
8. Mapear as entidades para DTOs.
9. Montar a resposta.
10. Retornar os serviços encontrados.

**Persistência**

- Consulta: `ServicoRepository`.
- Altera: nada.
- Consulta somente leitura.

**Saída da API**

Listagem:

```json
{
  "content": [
    {
      "id": "123",
      "codigo": "SER-000001",
      "nome": "Troca de óleo",
      "descricao": "Troca de óleo e filtro",
      "valor": 150.0,
      "tempoEstimadoMinutos": 60,
      "status": "ATIVO"
    },
    {
      "id": "124",
      "codigo": "SER-000002",
      "nome": "Alinhamento",
      "descricao": "Alinhamento completo",
      "valor": 120.0,
      "tempoEstimadoMinutos": 90,
      "status": "ATIVO"
    }
  ],
  "page": 0,
  "size": 20,
  "totalElements": 2,
  "totalPages": 1
}
```

Consulta por ID:

```json
{
  "id": "123",
  "codigo": "SER-000001",
  "nome": "Troca de óleo",
  "descricao": "Troca de óleo e filtro",
  "valor": 150.0,
  "tempoEstimadoMinutos": 60,
  "status": "ATIVO"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Consulta realizada. |
| `400` | Filtro, paginação ou parâmetro inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `servicos:ler`. |
| `404` | Serviço não encontrado por ID. |
| `500` | Falha inesperada. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de serviços.
- `ServicoRepository`.
- Middleware ou política de autorização para o escopo `servicos:ler`.

**Testes**

*Unitários*

- Lista serviços cadastrados.
- Consulta serviço por ID.
- Filtra serviços por nome.
- Filtra serviços por status.
- Respeita paginação.
- Rejeita filtros inválidos.
- Retorna erro para ID inexistente.
- Garante que a consulta não altera os dados dos serviços.

*Integração*

- `GET /servicos` retorna `200` com listagem paginada.
- `GET /servicos/{id}` retorna `200` com os dados do serviço.
- Nome informado filtra os resultados.
- Status informado filtra os resultados.
- Paginação retorna `page`, `size`, `totalElements` e `totalPages`.
- Serviço inexistente retorna `404`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Definir comportamento para serviços inativos na consulta
- [ ] Garantir que a consulta não altera os dados dos serviços

**Caso de uso**

- [ ] Criar caso de uso `ConsultarServicos`
- [ ] Criar consulta por ID
- [ ] Criar consulta de listagem
- [ ] Permitir filtro por nome, caso necessário
- [ ] Permitir filtro por status
- [ ] Aplicar paginação, caso adotada

**Repositório**

- [ ] Criar método de consulta no `ServicoRepository`
- [ ] Consultar serviço por ID
- [ ] Consultar serviços por filtros e paginação

**Handler HTTP**

- [ ] Implementar `GET /servicos`
- [ ] Implementar `GET /servicos/{id}`
- [ ] Criar DTO/response
- [ ] Aplicar autenticação JWT
- [ ] Aplicar autorização para o escopo `servicos:ler`
- [ ] Retornar `404` para serviço inexistente

**Validações**

- [ ] Validar formato do `id`, quando informado
- [ ] Validar valores dos filtros
- [ ] Validar paginação
- [ ] Retornar `400` para filtro ou parâmetro inválido
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Testes unitários**

- [ ] Consulta por ID
- [ ] Listagem de serviços
- [ ] Filtro por nome
- [ ] Filtro por status
- [ ] Serviço inexistente
- [ ] Consulta sem alteração de dados

**Testes de integração**

- [ ] Endpoint de listagem retorna `200`
- [ ] Endpoint de consulta por ID retorna `200`
- [ ] Endpoint respeita filtros
- [ ] Endpoint respeita paginação
- [ ] Endpoint retorna `404` para serviço inexistente
- [ ] Endpoint impede acesso de usuário sem permissão

**Documentação**

- [ ] Documentar os endpoints no Swagger/OpenAPI

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite
- [ ] Code Review aprovado

---
