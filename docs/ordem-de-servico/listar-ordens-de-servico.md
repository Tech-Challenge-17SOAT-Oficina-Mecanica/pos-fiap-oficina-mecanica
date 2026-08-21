---
documento: Refinamento de Requisitos — Listar Ordens de Serviço
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Listar Ordens de Serviço

Este documento detalha a tarefa Listar Ordens de Serviço do contexto de Ordem de Serviço.

## 10 · Listar Ordens de Serviço

### 10.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Listar Ordens de Serviço para acompanhar e gerenciar os atendimentos da oficina.

**Problema**

A oficina precisa acompanhar o status dos serviços, organizar atendimentos e reduzir falhas
causadas por controle manual ou planilhas.

**Pré-condições**

- Deve existir acesso à listagem de Ordens de Serviço.
- O usuário deve estar autorizado a listar Ordens de Serviço.
- As Ordens de Serviço devem estar registradas no sistema para serem exibidas.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-77 | Permitir listar Ordens de Serviço. |
| RF-OS-78 | Permitir detalhar Ordens de Serviço a partir da listagem. |
| RF-OS-79 | Exibir o status atual das Ordens de Serviço. |
| RF-OS-80 | Permitir o acompanhamento do andamento dos serviços. |
| RF-OS-81 | Permitir a consulta administrativa das Ordens de Serviço, com filtro por status, cliente e veículo. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-45 | A listagem deve ser feita por API RESTful. |
| RNF-OS-46 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-OS-47 | A listagem não deve alterar dados das Ordens de Serviço. |
| RNF-OS-48 | A resposta deve ser consistente com os status atuais das Ordens de Serviço. |

**Fluxo Principal**

1. O mecânico acessa a listagem de Ordens de Serviço.
2. O sistema consulta as Ordens de Serviço registradas, aplicando os filtros informados.
3. O sistema retorna a lista de Ordens de Serviço.
4. O sistema exibe o status atual de cada Ordem de Serviço.
5. O mecânico seleciona uma Ordem de Serviço para detalhamento, se necessário.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhuma Ordem de Serviço encontrada | Retorna a listagem vazia. |
| A2 | Status informado inválido | Informa que o status não é válido. |
| A3 | Documento ou placa inválidos | Informa que o critério informado não é válido. |
| A4 | Cliente ou veículo do filtro não encontrado | Informa que o registro do filtro não existe. |
| A5 | Usuário sem autorização | Impede a listagem. |
| A6 | Erro ao listar | Informa que não foi possível concluir a listagem. |

**Saída**

- Lista de Ordens de Serviço, com o status atual de cada uma e a identificação das que podem ser
  detalhadas.

**Pós-condições**

- As Ordens de Serviço permanecem inalteradas.
- O mecânico pode selecionar uma Ordem de Serviço para consulta detalhada e acompanhar o andamento
  dos atendimentos registrados.

---

### 10.2 Refinamento Técnico

**Endpoint**

```http
GET /ordens-servico
```

> **Decisão de projeto.** Esta é a mesma rota usada para localizar as OS de um cliente pelo
> CPF/CNPJ: a busca por `documento` é um filtro da listagem, e não um endpoint próprio. O
> detalhamento de uma OS específica fica em `GET /ordens-servico/{osId}`, descrito em
> [`consultar-ordem-de-servico.md`](consultar-ordem-de-servico.md).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `os:ler`.

**Entrada** — query params, todos opcionais:

| Param | Tipo | Descrição |
|---|---|---|
| `status` | enum | Filtra por status da OS |
| `documento` | string | Filtra pelo CPF/CNPJ do cliente |
| `placa` | string | Filtra pela placa do veículo |
| `page` / `size` | int | Paginação; `size` com máximo de 100 |

Valores válidos de `status`: `RECEBIDA`, `EM_DIAGNOSTICO`, `AGUARDANDO_APROVACAO`,
`AGUARDANDO_EXECUCAO`, `EM_EXECUCAO`, `FINALIZADA`, `ENTREGUE`, `CANCELADA`.

**Validações**

*Técnicas*

- `status`, quando informado, deve pertencer ao enum.
- `documento`, quando informado, deve ter formato válido de CPF/CNPJ.
- `placa`, quando informada, deve ter formato válido.
- Parâmetros de paginação válidos.

*Negócio*

- Deve existir cliente cadastrado com o documento informado.
- Deve existir veículo cadastrado com a placa informada.
- A listagem não altera dados.

**Processamento**

1. Receber os filtros informados e validar filtros e paginação.
2. Quando `documento` for informado, consultar o cliente pelo CPF/CNPJ.
3. Quando `placa` for informada, consultar o veículo pela placa.
4. Consultar as Ordens de Serviço conforme os filtros: por status, por cliente e por veículo.
5. Aplicar a paginação.
6. Retornar a lista de Ordens de Serviço encontradas.

**Persistência**

- Consulta: `ordem_servico`, `cliente` (filtro por documento), `veiculo` (filtro por placa).
- Altera: nada.

**Saída da API**

```json
{
  "data": [
    {
      "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
      "cliente": {
        "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
        "nome": "Nome do Cliente",
        "documento": "00000000000"
      },
      "veiculo": {
        "veiculoId": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
        "placa": "ABC1D23",
        "marca": "Marca do Veículo",
        "modelo": "Modelo do Veículo"
      },
      "status": "RECEBIDA"
    },
    {
      "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
      "cliente": {
        "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
        "nome": "Nome do Cliente",
        "documento": "00000000000"
      },
      "veiculo": {
        "veiculoId": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
        "placa": "ABC1D23",
        "marca": "Marca do Veículo",
        "modelo": "Modelo do Veículo"
      },
      "status": "EM_DIAGNOSTICO"
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 2,
  "totalPaginas": 1
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Listagem retornada com sucesso. |
| `400` | Filtros ou parâmetros de paginação inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:ler`. |
| `404` | Cliente ou veículo informado no filtro não encontrado. |

> Listagem vazia é `200` com `"data": []`, nunca `404`.

**Dependências**

- `OrdemDeServicoRepository`.
- `ClienteRepository` e `VeiculoRepository`.
- Validador de CPF/CNPJ e validador de placa.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Normalização e validação dos filtros recebidos.
- Status fora do enum é rejeitado.
- Documento e placa inválidos são rejeitados.
- Paginação inválida é rejeitada.

*Integração*

- Listagem sem filtros retorna a página default.
- Filtro por status retorna apenas as OS naquele status.
- Filtro por CPF/CNPJ retorna apenas as OS do cliente.
- Filtro por placa retorna apenas as OS do veículo.
- Filtros combinados de status, documento e placa funcionam juntos.
- Sem resultado, retorna lista vazia com `200`.
- Cliente ou veículo inexistente no filtro retorna `404`.
- Sem token retorna `401` e perfil sem escopo retorna `403`.
- A listagem traz o status atual de cada OS e não altera dados persistidos.

---

### 10.3 Checklist de Implementação

**Domínio**

- [ ] Garantir que a Ordem de Serviço possua identificador único e status atual
- [ ] Garantir que a OS mantenha vínculo com `Cliente` e com `Veiculo`

**Caso de uso**

- [ ] Implementar `ListarOrdensDeServico` recebendo filtros opcionais
- [ ] Validar os filtros de status, cliente, veículo e paginação
- [ ] Consultar as Ordens de Serviço conforme os filtros recebidos
- [ ] Retornar lista vazia quando não houver registros para os filtros informados
- [ ] Garantir que a listagem não altere dados das Ordens de Serviço

**Repositório**

- [ ] Criar o método de listagem de Ordens de Serviço
- [ ] Criar a listagem por status
- [ ] Criar a listagem por cliente
- [ ] Criar a listagem por veículo
- [ ] Implementar a paginação na consulta

**Handler HTTP**

- [ ] Implementar `GET /ordens-servico`
- [ ] Implementar a leitura e a validação dos query params
- [ ] Criar DTO/request de entrada para filtros e paginação
- [ ] Criar DTO/response com a lista de Ordens de Serviço
- [ ] Implementar o envelope de resposta paginado
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Retornar `400` para filtros ou paginação inválidos
- [ ] Retornar `404` quando cliente ou veículo do filtro não existir
- [ ] Retornar `401` sem autenticação e `403` sem permissão

**Testes unitários**

- [ ] Listagem sem filtros
- [ ] Listagem por status
- [ ] Listagem por cliente
- [ ] Listagem por veículo
- [ ] Lista vazia
- [ ] Status inválido
- [ ] Cliente inexistente no filtro
- [ ] Veículo inexistente no filtro
- [ ] Paginação inválida

**Testes de integração**

- [ ] Endpoint retornando o status atual de cada Ordem de Serviço
- [ ] Filtros combinados
- [ ] A listagem não altera dados persistidos

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar critérios de aceite da task

---
