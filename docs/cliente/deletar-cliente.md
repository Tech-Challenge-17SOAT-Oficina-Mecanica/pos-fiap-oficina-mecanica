---
documento: Refinamento de Requisitos — Deletar Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Deletar Cliente

Este documento detalha a tarefa Deletar Cliente do contexto de Cliente.

## 5 · Deletar Cliente

### 5.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Remover do cadastro ativo um cliente que não é mais atendido pela oficina, sem perder o
histórico de atendimentos já realizados.

**Problema**

Cadastros duplicados, clientes que nunca voltaram e registros criados por engano poluem a busca
e fazem o mecânico selecionar o cliente errado ao abrir uma OS. Ao mesmo tempo, apagar o
registro de verdade quebraria o histórico de OS, os relatórios de faturamento e a rastreabilidade
fiscal — então a oficina precisa de uma forma de tirar o cliente de circulação sem destruir o
passado.

**Pré-condições**

- O cliente deve estar cadastrado e ativo.
- O usuário deve estar autorizado a manter o cadastro de clientes.
- O cliente não pode possuir Ordem de Serviço em aberto.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-CLI-28 | Permitir inativar um cliente do cadastro. |
| RF-CLI-29 | Impedir a inativação quando o cliente possuir OS em aberto. |
| RF-CLI-30 | Inativar automaticamente os veículos vinculados ao cliente. |
| RF-CLI-31 | Manter o cliente e seu histórico de OS acessíveis para consulta e relatórios. |
| RF-CLI-32 | Excluir o cliente inativo das buscas e listagens padrão. |
| RF-CLI-33 | Permitir reativar um cliente inativado. |
| RF-CLI-34 | Permitir que o mesmo CPF/CNPJ seja cadastrado novamente após a inativação. |
| RF-CLI-35 | Registrar quem inativou, quando e por qual motivo. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-CLI-21 | A operação deve ser feita por API RESTful. |
| RNF-CLI-22 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-CLI-23 | A exclusão deve ser lógica, nunca física — o registro permanece no banco. |
| RNF-CLI-24 | A operação deve ser auditável, com registro de responsável, data e motivo. |
| RNF-CLI-25 | A operação não deve alterar nenhuma Ordem de Serviço já emitida. |
| RNF-CLI-26 | A operação deve ser idempotente: inativar um cliente já inativo não gera erro nem efeito adicional. |
| RNF-CLI-27 | A unicidade de CPF/CNPJ deve valer apenas entre clientes ativos. |

**Fluxo Principal**

1. O mecânico localiza o cliente no cadastro.
2. O mecânico solicita a exclusão do cliente.
3. O sistema verifica se o usuário está autorizado.
4. O sistema verifica se o cliente possui OS em aberto.
5. O sistema inativa o cliente.
6. O sistema inativa os veículos vinculados ao cliente.
7. O sistema registra a operação na trilha de auditoria.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente com OS em aberto | Impede a exclusão e informa quais OS estão em andamento. |
| A2 | Cliente já inativo | Informa que o cliente já está inativo e não realiza nova alteração. |
| A3 | Cliente não encontrado | Informa que o cliente não existe. |
| A4 | Cliente com veículos vinculados | Inativa os veículos junto com o cliente. |
| A5 | Recadastro do mesmo CPF/CNPJ | Permite cadastrar um cliente novo com o documento de um cliente inativo. |
| A6 | Reativação | Permite reativar o cliente, desde que não exista outro cliente ativo com o mesmo CPF/CNPJ. Os veículos precisam ser reativados individualmente. |
| A7 | Usuário sem autorização | Impede a operação. |

**Saída**

- Confirmação de que o cliente foi inativado, com a relação de veículos inativados junto; **ou**
- Indicação do motivo pelo qual a exclusão foi impedida.

**Pós-condições**

- O cliente está marcado como inativo e não aparece nas buscas e listagens padrão.
- Os veículos vinculados estão inativos.
- As Ordens de Serviço históricas permanecem íntegras e continuam referenciando o cliente.
- Não é possível abrir nova OS para esse cliente enquanto ele estiver inativo.
- O CPF/CNPJ fica liberado para um novo cadastro.
- A operação está registrada na trilha de auditoria.

---

### 5.2 Refinamento Técnico

**Endpoint**

```http
DELETE /clientes/{clienteId}
POST   /clientes/{clienteId}/reativacao
```

> **Decisão de projeto.** O `DELETE` executa exclusão **lógica**. O verbo é mantido por ser o
> semanticamente correto do ponto de vista de quem consome a API — quem chama quer remover o
> cliente do cadastro, e como isso é implementado é detalhe interno. A alternativa seria
> `PATCH /clientes/{id}/inativacao`, descartada por expor a implementação no contrato.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `clientes:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `clienteId` | uuid | Identificador do cliente. |
| Query | `motivo` | string | Opcional; motivo da inativação, até 200 caracteres. |
| Path (reativação) | `clienteId` | uuid | Identificador do cliente a reativar. |

Nenhuma das duas operações tem corpo na requisição.

**Validações**

*Técnicas*

- `clienteId` em formato UUID válido.
- Cliente existe na base.
- `motivo` com no máximo 200 caracteres.

*Negócio*

- O cliente não possui OS com status diferente de `ENTREGUE` ou `CANCELADA`.
- Cliente já inativo torna a operação idempotente: retorna `204` sem alterar nada.
- A reativação é bloqueada se já existir outro cliente ativo com o mesmo CPF/CNPJ.

**Processamento**

*Exclusão lógica (`DELETE`)*

1. Carregar o cliente por identificador.
2. Se já estiver inativo, retornar `204` sem efeito.
3. Consultar o módulo de Ordem de Serviço buscando OS em aberto do cliente.
4. Havendo OS em aberto, abortar com `409` e a lista das OS.
5. Abrir transação.
6. Marcar `cliente.ativo = false` e gravar `inativadoEm`, `inativadoPor` e `motivoInativacao`.
7. Registrar na trilha de auditoria.
8. Publicar `ClienteInativado`.
9. Commit.
10. A política "cliente inativado, então inativar veículos" consome o evento e inativa cada veículo vinculado.

*Reativação (`POST /reativacao`)*

1. Carregar o cliente e verificar que está inativo.
2. Verificar que não existe outro cliente ativo com o mesmo CPF/CNPJ; havendo, retornar `409`.
3. Marcar `ativo = true` e limpar os campos de inativação.
4. Registrar na trilha de auditoria.
5. Publicar `ClienteReativado`.

Os veículos não são reativados em cascata: a reativação é individual e deliberada, para não
trazer de volta carros que o cliente já não tem.

**Persistência**

- Consulta: `cliente`, `veiculo`, módulo de Ordem de Serviço (OS em aberto).
- Altera: `cliente.ativo`, `cliente.inativado_em`, `cliente.inativado_por`,
  `cliente.motivo_inativacao`, `auditoria` (insert).
- Não altera: nenhuma tabela de Ordem de Serviço, Orçamento ou financeiro.

Índices e constraints:

- A unicidade de `cpf_cnpj` deve valer apenas entre clientes ativos — índice parcial
  `UNIQUE (cpf_cnpj) WHERE ativo = true`. Sem isso, o mesmo CPF não pode ser recadastrado depois
  de uma inativação.
- Nenhuma foreign key de OS para cliente pode ter `ON DELETE CASCADE`.

**Saída da API**

`DELETE` — `200`:

```json
{
  "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "nome": "João Ribeiro",
  "ativo": false,
  "inativadoEm": "2026-08-12T18:10:00-03:00",
  "inativadoPor": "4c1d8e62-9b07-4a53-8f16-2d7e5a90c3b1",
  "motivo": "Cadastro duplicado",
  "veiculosInativados": [
    { "id": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607", "placa": "PEB1D23" }
  ],
  "documentoLiberadoParaNovoCadastro": true
}
```

`DELETE` em cliente já inativo — `204`, sem corpo.

`POST /reativacao` — `200`:

```json
{
  "clienteId": "c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "nome": "João Ribeiro",
  "ativo": true,
  "reativadoEm": "2026-08-13T09:05:00-03:00",
  "reativadoPor": "4c1d8e62-9b07-4a53-8f16-2d7e5a90c3b1",
  "veiculosReativados": 0
}
```

Conflito por OS em aberto — `409`:

```json
{
  "type": "https://api.oficina/errors/cliente-com-os-aberta",
  "title": "Cliente possui Ordem de Serviço em aberto",
  "status": 409,
  "detail": "Não é possível excluir o cliente enquanto houver OS em andamento",
  "instance": "/clientes/c7f3a9b2-1e4d-4c8a-9f21-0b6d5e2a7c14",
  "erros": [
    {
      "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
      "status": "EM_EXECUCAO"
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Cliente inativado ou reativado. |
| `204` | Cliente já estava inativo — operação idempotente. |
| `400` | `clienteId` fora do formato UUID; `motivo` acima do limite. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `clientes:escrever`. |
| `404` | Cliente não encontrado. |
| `409` | Cliente possui OS em aberto; reativação com CPF/CNPJ já usado por cliente ativo. |

**Dependências**

- `ClienteRepository`.
- `VeiculoRepository` (inativação em cascata).
- `AuditoriaRepository`.
- Módulo Ordem de Serviço — consulta de OS em aberto.
- Publicador de eventos de domínio.
- Caso de uso Deletar Veículo (acionado pela política de cascata).
- Caso de uso Consultar Cliente (não deve retornar inativos).
- Caso de uso Cadastrar Cliente (depende do índice parcial para permitir recadastro).

**Testes**

*Unitários*

- Cliente ativo sem OS em aberto pode ser inativado.
- Cliente com OS `EM_EXECUCAO` não pode ser inativado.
- Cliente com OS apenas em `ENTREGUE` e `CANCELADA` pode ser inativado.
- Inativar cliente já inativo não altera os campos de auditoria originais.
- Reativação bloqueada quando há outro cliente ativo com o mesmo CPF/CNPJ.

*Integração*

- `DELETE` válido retorna `200` e marca `ativo = false`.
- `DELETE` em cliente já inativo retorna `204`.
- `DELETE` em cliente com OS aberta retorna `409` com a lista de OS.
- `DELETE` em cliente inexistente retorna `404`.
- Perfil sem o escopo `clientes:escrever` retorna `403`.
- `POST /reativacao` válido retorna `200` e marca `ativo = true`.
- Reativação com CPF/CNPJ em uso por cliente ativo retorna `409`.
- Após o `DELETE`, os veículos vinculados ficam inativos.
- Após a reativação, os veículos continuam inativos.

*Regressão*

- Após inativar o cliente, as OS históricas continuam retornando o nome do cliente normalmente.
- Após inativar o cliente, o relatório de faturamento do período anterior não muda.
- Após inativar, o mesmo CPF/CNPJ pode ser cadastrado novamente em um cliente novo.
- Cliente inativo não aparece em `GET /clientes` sem filtro explícito.
- Não é possível abrir OS para cliente inativo.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `inativar()` na entidade `Cliente` com os campos `inativadoEm`, `inativadoPor` e `motivo`
- [ ] Implementar o método `reativar()` na entidade `Cliente`
- [ ] Implementar a regra que impede inativação com OS em aberto
- [ ] Implementar a idempotência: inativar cliente já inativo não altera nada
- [ ] Garantir que a inativação não toca em nenhuma tabela de OS, orçamento ou financeiro

**Caso de uso**

- [ ] Implementar `InativarCliente`
- [ ] Implementar `ReativarCliente`

**Repositório**

- [ ] Implementar `ClienteRepository.inativar` e `ClienteRepository.reativar`
- [ ] Ajustar as consultas padrão para filtrar somente clientes ativos
- [ ] Criar índice parcial `UNIQUE (cpf_cnpj) WHERE ativo = true`
- [ ] Remover qualquer `ON DELETE CASCADE` em foreign key que aponte para cliente
- [ ] Implementar registro na trilha de auditoria

**Integrações**

- [ ] Consultar o módulo de Ordem de Serviço para verificar OS em aberto
- [ ] Implementar a política "cliente inativado, então inativar veículos vinculados"

**Handler HTTP**

- [ ] Implementar `DELETE /clientes/{clienteId}`
- [ ] Implementar `POST /clientes/{clienteId}/reativacao`
- [ ] Criar DTO/response de saída das duas operações
- [ ] Aplicar autenticação JWT e autorização por escopo nas rotas
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar `clienteId` em formato UUID
- [ ] Validar `motivo` com no máximo 200 caracteres
- [ ] Retornar `204` quando o cliente já estiver inativo
- [ ] Retornar `409` quando houver OS em aberto
- [ ] Retornar `409` na reativação com CPF/CNPJ em uso por cliente ativo

**Eventos**

- [ ] Publicar `ClienteInativado`
- [ ] Publicar `ClienteReativado`

**Testes unitários**

- [ ] Cliente sem OS em aberto pode ser inativado
- [ ] Cliente com OS `EM_EXECUCAO` não pode ser inativado
- [ ] Cliente com OS apenas `ENTREGUE` e `CANCELADA` pode ser inativado
- [ ] Inativar cliente já inativo não altera os campos de auditoria originais

**Testes de integração**

- [ ] `DELETE` válido retornando `200` com `ativo` falso
- [ ] `DELETE` em cliente já inativo retornando `204`
- [ ] `DELETE` em cliente com OS aberta retornando `409` com a lista de OS
- [ ] `DELETE` em cliente inexistente retornando `404`
- [ ] Perfil sem escopo retornando `403`
- [ ] Veículos vinculados ficando inativos após o `DELETE`
- [ ] Veículos continuando inativos após a reativação do cliente

**Testes de regressão**

- [ ] OS históricas continuam retornando o nome do cliente após a inativação
- [ ] Relatório de faturamento do período anterior não muda após a inativação
- [ ] O mesmo CPF/CNPJ pode ser cadastrado novamente após a inativação
- [ ] Cliente inativo não aparece em `GET /clientes` sem filtro explícito
- [ ] Não é possível abrir OS para cliente inativo

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI
- [ ] Documentar explicitamente que o `DELETE` é exclusão lógica
- [ ] Documentar o exemplo de erro `409` `cliente-com-os-aberta`

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Migration versionada e reversível

---
