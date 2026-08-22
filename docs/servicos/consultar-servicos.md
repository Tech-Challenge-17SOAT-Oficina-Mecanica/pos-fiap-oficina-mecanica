---
documento: Refinamento de Requisitos — Consultar Serviços
dono: João Victor Silva de Oliveira
versao: 0.2
atualizado_em: 2026-08-20
status: em revisao
---

# Refinamento de Requisitos — Consultar Serviços

Este documento detalha a tarefa Consultar Serviços do contexto de Serviços.

## 2 · Consultar Serviços

### 2.1 Refinamento de Produto

**Persona**
Mecânico.

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
| RF-SRV-08 | Permitir listar os serviços cadastrados. |
| RF-SRV-09 | Permitir consultar os dados de um serviço específico. |
| RF-SRV-10 | Permitir identificar serviços ativos e inativos. |
| RF-SRV-11 | Permitir filtrar ou pesquisar serviços por nome parcial. |
| RF-SRV-12 | Ocultar serviços inativos por padrão e exibi-los com `incluirInativos=true`. |
| RF-SRV-13 | Apresentar os dados necessários para identificação do serviço e composição de orçamentos. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SRV-06 | A consulta deve apresentar informações consistentes com os dados persistidos. |
| RNF-SRV-07 | A API deve retornar respostas padronizadas. |
| RNF-SRV-08 | A consulta deve respeitar as regras de autenticação e autorização. |
| RNF-SRV-09 | A operação não deve alterar os dados dos serviços. |
| RNF-SRV-10 | A consulta deve possuir desempenho adequado para o volume esperado do MVP. |

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

### 2.2 Refinamento Técnico

**Endpoint**

```http
GET /servicos
GET /servicos/{servicoId}
```

O primeiro endpoint lista os serviços cadastrados no catálogo da oficina. O segundo consulta os
dados de um serviço específico.

> **Decisão de projeto.** A consulta por ID foi mantida no mesmo requisito da listagem porque as
> duas operações são somente leitura, compartilham autorização e usam o mesmo recurso
> `/servicos`. A alternativa seria separar em duas tarefas, mas isso duplicaria regras de
> autorização, DTO e testes de leitura.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `MECANICO`.
- Escopo: `servicos:ler`.

> **Decisão de projeto.** A consulta do catálogo é **restrita à oficina**, com perfil `MECANICO`.
> O cliente não consulta o catálogo: ele vê os serviços pelo orçamento que recebe.

> **Decisão de projeto.** O path param é `{servicoId}`, e não `{id}`, alinhado a `{clienteId}`,
> `{veiculoId}` e `{pecaId}` dos demais contextos.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Query | `nome` | string | Filtro opcional por nome parcial do serviço. |
| Query | `incluirInativos` | boolean | Opcional. Default `false`: serviços inativos ficam fora da listagem. |
| Query | `pagina` | int | Página da listagem. Default `0`. |
| Query | `tamanho` | int | Tamanho da página. Default `20`, máximo `50`. |
| Path | `servicoId` | uuid | Identificador do serviço, obrigatório na consulta individual. |

Exemplo:

```http
GET /servicos?nome=oleo&incluirInativos=false&pagina=0&tamanho=20
```

A operação não recebe corpo.

**Validações**

- `servicoId`, quando informado, deve possuir formato válido de UUID.
- `incluirInativos`, quando informado, deve ser booleano.
- `pagina` deve ser maior ou igual a zero.
- `tamanho` deve ser maior que zero e no máximo `50`; acima disso, `400`.
- Para consulta por ID, o serviço deve existir.

**Processamento**

1. Receber os parâmetros da requisição.
2. Identificar o usuário autenticado.
3. Validar autorização.
4. Validar formato do `servicoId`, quando informado.
5. Validar filtros e paginação.
6. Aplicar os filtros informados, ocultando serviços inativos quando `incluirInativos` for `false`.
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
  "data": [
    {
      "id": "4b8e2c17-95a3-4f60-b7d1-6e0c58a3f942",
      "codigo": "SER-000001",
      "nome": "Troca de óleo",
      "descricao": "Troca de óleo e filtro",
      "valor": 150.0,
      "tempoEstimadoMinutos": 60,
      "ativo": true
    },
    {
      "id": "e7c15d09-3a26-4b8f-90d4-51fa62e7c3b8",
      "codigo": "SER-000002",
      "nome": "Alinhamento",
      "descricao": "Alinhamento completo",
      "valor": 120.0,
      "tempoEstimadoMinutos": 90,
      "ativo": true
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 2,
  "totalPaginas": 1
}
```

Consulta por identificador:

```json
{
  "id": "4b8e2c17-95a3-4f60-b7d1-6e0c58a3f942",
  "codigo": "SER-000001",
  "nome": "Troca de óleo",
  "descricao": "Troca de óleo e filtro",
  "valor": 150.0,
  "tempoEstimadoMinutos": 60,
  "ativo": true,
  "version": 3
}
```

> **Decisão de projeto.** A listagem usa o envelope padrão do projeto — `data`, `pagina`,
> `tamanho`, `totalElementos` e `totalPaginas` —, e o recurso único vai **direto, sem envelope**
> (D-21). A consulta por identificador expõe `version`, que a atualização envia no `If-Match`.

> **Decisão de projeto.** Serviços inativos ficam **fora da listagem por padrão** e aparecem com
> `incluirInativos=true`, mesmo parâmetro já usado na consulta de peças. O teto de `tamanho` é
> **50**.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Consulta realizada. |
| `400` | Filtro, paginação ou parâmetro inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `servicos:ler`. |
| `404` | Serviço não encontrado. |
| `500` | Falha inesperada. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de serviços.
- `ServicoRepository`.
- Middleware ou política de autorização para o escopo `servicos:ler`.

**Testes**

*Unitários*

- Lista serviços cadastrados.
- Consulta serviço por identificador.
- Filtra serviços por nome.
- Oculta serviços inativos por padrão.
- Inclui serviços inativos quando `incluirInativos=true`.
- Respeita paginação.
- Rejeita filtros inválidos.
- Rejeita `tamanho` acima de 50.
- Retorna erro para identificador inexistente.
- Garante que a consulta não altera os dados dos serviços.

*Integração*

- `GET /servicos` retorna `200` com listagem paginada.
- `GET /servicos/{servicoId}` retorna `200` com o objeto direto, sem envelope.
- Nome informado filtra os resultados.
- `incluirInativos=true` traz também os serviços inativos.
- `tamanho` acima de 50 retorna `400`.
- Paginação retorna `pagina`, `tamanho`, `totalElementos` e `totalPaginas`.
- Serviço inexistente retorna `404`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Ocultar serviços inativos por padrão na listagem
- [ ] Garantir que a consulta não altera os dados dos serviços

**Caso de uso**

- [ ] Criar caso de uso `ConsultarServicos`
- [ ] Criar consulta por identificador
- [ ] Criar consulta de listagem
- [ ] Permitir filtro por nome parcial
- [ ] Permitir o filtro `incluirInativos`
- [ ] Aplicar paginação com o envelope padrão do projeto

**Repositório**

- [ ] Criar método de consulta no `ServicoRepository`
- [ ] Consultar serviço por identificador
- [ ] Consultar serviços por filtros e paginação

**Handler HTTP**

- [ ] Implementar `GET /servicos`
- [ ] Implementar `GET /servicos/{servicoId}`
- [ ] Criar DTO/response com o envelope `data`, `pagina`, `tamanho`, `totalElementos` e `totalPaginas` na listagem
- [ ] Devolver o objeto direto, sem envelope, na consulta por identificador
- [ ] Expor `version` na consulta por identificador
- [ ] Aplicar autenticação JWT
- [ ] Aplicar autorização para o escopo `servicos:ler`
- [ ] Retornar `404` para serviço inexistente

**Validações**

- [ ] Validar formato do `servicoId`, quando informado
- [ ] Validar valores dos filtros
- [ ] Validar paginação e o teto de `tamanho` em 50
- [ ] Retornar `400` para filtro ou parâmetro inválido
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Testes unitários**

- [ ] Consulta por identificador
- [ ] Listagem de serviços
- [ ] Filtro por nome
- [ ] Filtro `incluirInativos`
- [ ] Serviço inexistente
- [ ] Consulta sem alteração de dados

**Testes de integração**

- [ ] Endpoint de listagem retorna `200`
- [ ] Endpoint de consulta por identificador retorna `200` com o objeto direto
- [ ] Endpoint respeita filtros
- [ ] Endpoint respeita paginação e devolve o envelope padrão
- [ ] Endpoint retorna `400` para `tamanho` acima de 50
- [ ] Endpoint retorna `404` para serviço inexistente
- [ ] Endpoint impede acesso de usuário sem permissão

**Documentação**

- [ ] Documentar os endpoints no Swagger/OpenAPI

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite
- [ ] Code Review aprovado

---
