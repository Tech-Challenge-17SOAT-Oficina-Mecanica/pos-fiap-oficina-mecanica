---
documento: Refinamento de Requisitos — Deletar Veículo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Deletar Veículo

Este documento detalha a tarefa Deletar Veículo do contexto de Veículo.

## 3 · Deletar Veículo

### 3.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Remover do cadastro ativo um veículo que a oficina não atende mais, sem perder o histórico de
manutenção já realizado nele.

**Problema**

Veículos cadastrados por engano, com placa digitada errada ou que o cliente não possui mais
poluem a busca e fazem abrir OS no carro errado. Apagar o registro, porém, destruiria o histórico
de manutenção do veículo — a informação que permite ao mecânico saber o que já foi trocado e
quando, e que sustenta qualquer discussão de garantia.

**Pré-condições**

- O veículo deve estar cadastrado e ativo.
- O usuário deve estar autorizado a manter o cadastro de veículos.
- O veículo não pode possuir Ordem de Serviço em aberto.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-VEI-13 | Permitir inativar um veículo do cadastro. |
| RF-VEI-14 | Impedir a inativação quando o veículo possuir OS em aberto. |
| RF-VEI-15 | Manter o veículo e seu histórico de OS acessíveis para consulta. |
| RF-VEI-16 | Excluir o veículo inativo das buscas e listagens padrão. |
| RF-VEI-17 | Permitir reativar um veículo inativado. |
| RF-VEI-18 | Permitir que a mesma placa seja cadastrada novamente após a inativação. |
| RF-VEI-19 | Registrar quem inativou, quando e por qual motivo. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-VEI-10 | A operação deve ser feita por API RESTful. |
| RNF-VEI-11 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-VEI-12 | A exclusão deve ser lógica, nunca física — o registro permanece no banco. |
| RNF-VEI-13 | A operação deve ser auditável, com registro de responsável, data e motivo. |
| RNF-VEI-14 | A operação não deve alterar nenhuma Ordem de Serviço já emitida. |
| RNF-VEI-15 | A operação deve ser idempotente: inativar um veículo já inativo não gera erro nem efeito adicional. |
| RNF-VEI-16 | A unicidade de placa deve valer apenas entre veículos ativos. |

**Fluxo Principal**

1. O mecânico localiza o veículo no cadastro, pela placa ou pelo cliente.
2. O mecânico solicita a exclusão do veículo.
3. O sistema verifica se o usuário está autorizado.
4. O sistema verifica se o veículo possui OS em aberto.
5. O sistema inativa o veículo.
6. O sistema registra a operação na trilha de auditoria.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Veículo com OS em aberto | Impede a exclusão e informa quais OS estão em andamento. |
| A2 | Veículo já inativo | Informa que o veículo já está inativo e não realiza nova alteração. |
| A3 | Veículo não encontrado | Informa que o veículo não existe. |
| A4 | Inativação em cascata | Quando o cliente é inativado, seus veículos são inativados automaticamente pela política do agregado Cliente. |
| A5 | Recadastro da mesma placa | Permite cadastrar um veículo novo com a placa de um veículo inativo. |
| A6 | Reativação | Permite reativar o veículo, desde que não exista outro veículo ativo com a mesma placa e que o cliente proprietário esteja ativo. |
| A7 | Usuário sem autorização | Impede a operação. |

**Saída**

- Confirmação de que o veículo foi inativado, com a quantidade de OS históricas preservadas; **ou**
- Indicação do motivo pelo qual a exclusão foi impedida.

**Pós-condições**

- O veículo está marcado como inativo e não aparece nas buscas e listagens padrão.
- As Ordens de Serviço históricas permanecem íntegras e continuam referenciando o veículo.
- Não é possível abrir nova OS para esse veículo enquanto ele estiver inativo.
- A placa fica liberada para um novo cadastro.
- A operação está registrada na trilha de auditoria.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
DELETE /veiculos/{veiculoId}
POST   /veiculos/{veiculoId}/reativacao
```

> **Decisão de projeto.** O `DELETE` executa exclusão **lógica**. O verbo é mantido por ser o
> semanticamente correto do ponto de vista de quem consome a API — quem chama quer remover o
> veículo do cadastro, e como isso é implementado é detalhe interno. Mesma decisão adotada em
> [`deletar-cliente.md`](../cliente/deletar-cliente.md).

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `veiculos:escrever`.
- O identificador do usuário responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `veiculoId` | uuid | Identificador do veículo. |
| Query | `motivo` | string | Opcional; motivo da inativação, até 200 caracteres. |
| Path (reativação) | `veiculoId` | uuid | Identificador do veículo a reativar. |

Nenhuma das duas operações tem corpo na requisição.

**Validações**

*Técnicas*

- `veiculoId` em formato UUID válido.
- Veículo existe na base.
- `motivo` com no máximo 200 caracteres.

*Negócio*

- O veículo não possui OS com status diferente de `ENTREGUE` ou `CANCELADA`.
- Veículo já inativo torna a operação idempotente: retorna `204` sem alterar nada.
- A reativação é bloqueada se já existir outro veículo ativo com a mesma placa.
- A reativação é bloqueada se o cliente proprietário estiver inativo.

**Processamento**

*Exclusão lógica (`DELETE`)*

1. Carregar o veículo por identificador.
2. Se já estiver inativo, retornar `204` sem efeito.
3. Consultar o módulo de Ordem de Serviço buscando OS em aberto do veículo.
4. Havendo OS em aberto, abortar com `409` e a lista das OS.
5. Abrir transação.
6. Marcar `veiculo.ativo = false` e gravar `inativadoEm`, `inativadoPor` e `motivoInativacao`.
7. Registrar na trilha de auditoria.
8. Publicar `VeiculoInativado`.
9. Commit.

*Reativação (`POST /reativacao`)*

1. Carregar o veículo e verificar que está inativo.
2. Verificar que não existe outro veículo ativo com a mesma placa; havendo, retornar `409`.
3. Verificar que o cliente proprietário está ativo; caso contrário, retornar `422`.
4. Marcar `ativo = true` e limpar os campos de inativação.
5. Registrar na trilha de auditoria.
6. Publicar `VeiculoReativado`.

*Consumo do evento `ClienteInativado`*

1. Receber o evento do agregado Cliente.
2. Buscar os veículos ativos do cliente.
3. Inativar cada um com o motivo "Cliente inativado".

**Persistência**

- Consulta: `veiculo`, `cliente`, módulo de Ordem de Serviço (OS em aberto).
- Altera: `veiculo.ativo`, `veiculo.inativado_em`, `veiculo.inativado_por`,
  `veiculo.motivo_inativacao`, `auditoria` (insert).
- Não altera: nenhuma tabela de Ordem de Serviço, Orçamento ou financeiro.

Índices e constraints:

- Unicidade de placa apenas entre ativos — índice parcial `UNIQUE (placa) WHERE ativo = true`.
  Sem isso, uma placa inativada fica bloqueada para sempre e o carro nunca pode ser recadastrado.
- Nenhuma foreign key de OS para veículo pode ter `ON DELETE CASCADE`.

**Saída da API**

`DELETE` — `200`:

```json
{
  "veiculoId": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
  "placa": "PEB1D23",
  "ativo": false,
  "inativadoEm": "2026-08-12T18:25:00-03:00",
  "inativadoPor": "4c1d8e62-9b07-4a53-8f16-2d7e5a90c3b1",
  "motivo": "Placa cadastrada incorretamente",
  "placaLiberadaParaNovoCadastro": true
}
```

`DELETE` em veículo já inativo — `204`, sem corpo.

`POST /reativacao` — `200`:

```json
{
  "veiculoId": "1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
  "placa": "PEB1D23",
  "ativo": true,
  "reativadoEm": "2026-08-13T09:10:00-03:00",
  "reativadoPor": "4c1d8e62-9b07-4a53-8f16-2d7e5a90c3b1"
}
```

Conflito por OS em aberto — `409`:

```json
{
  "type": "https://api.oficina/errors/veiculo-com-os-aberta",
  "title": "Veículo possui Ordem de Serviço em aberto",
  "status": 409,
  "detail": "Não é possível excluir o veículo enquanto houver OS em andamento",
  "instance": "/veiculos/1a2b3c44-5d6e-4f70-8a91-b2c3d4e5f607",
  "erros": [
    {
      "ordemServicoId": "e21b7c46-0d95-4f83-a6b1-3c5d92e74801",
      "status": "AGUARDANDO_APROVACAO"
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Veículo inativado ou reativado. |
| `204` | Veículo já estava inativo — operação idempotente. |
| `400` | `veiculoId` fora do formato UUID; `motivo` acima do limite. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `veiculos:escrever`. |
| `404` | Veículo não encontrado. |
| `409` | Veículo possui OS em aberto; reativação com placa já usada por veículo ativo. |
| `422` | Reativação com cliente proprietário inativo. |

**Dependências**

- `VeiculoRepository`.
- `ClienteRepository` (verificação da situação do proprietário na reativação).
- `AuditoriaRepository`.
- Módulo Ordem de Serviço — consulta de OS em aberto.
- Publicador e consumidor de eventos de domínio.
- Caso de uso Deletar Cliente — aciona este caso de uso via política de cascata.
- Caso de uso Consultar Veículo (não deve retornar inativos).
- Caso de uso Cadastrar Veículo (depende do índice parcial para permitir recadastro).

**Testes**

*Unitários*

- Veículo sem OS em aberto pode ser inativado.
- Veículo com OS `AGUARDANDO_APROVACAO` não pode ser inativado.
- Veículo com OS apenas `ENTREGUE` pode ser inativado.
- Inativar veículo já inativo não altera os campos de auditoria originais.
- Reativação bloqueada quando há outro veículo ativo com a mesma placa.
- Reativação bloqueada quando o cliente proprietário está inativo.

*Integração*

- `DELETE` válido retorna `200` e marca `ativo = false`.
- `DELETE` em veículo já inativo retorna `204`.
- `DELETE` em veículo com OS aberta retorna `409` com a lista de OS.
- `DELETE` em veículo inexistente retorna `404`.
- Perfil sem o escopo `veiculos:escrever` retorna `403`.
- `POST /reativacao` válido retorna `200` e marca `ativo = true`.
- Reativação com placa em uso por veículo ativo retorna `409`.
- Reativação com cliente inativo retorna `422`.

*Regressão*

- Após inativar o veículo, as OS históricas continuam retornando a placa e o modelo.
- Após inativar, a mesma placa pode ser cadastrada em um veículo novo.
- Veículo inativo não aparece em `GET /veiculos` sem filtro explícito.
- Não é possível abrir OS para veículo inativo.
- Inativar o cliente inativa os veículos dele em cascata.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Implementar o método `inativar()` na entidade `Veiculo` com os campos `inativadoEm`, `inativadoPor` e `motivo`
- [ ] Implementar o método `reativar()` na entidade `Veiculo`
- [ ] Implementar a regra que impede inativação com OS em aberto
- [ ] Implementar a regra que impede reativação com placa em uso por veículo ativo
- [ ] Implementar a regra que impede reativação com cliente proprietário inativo
- [ ] Implementar a idempotência: inativar veículo já inativo não altera nada
- [ ] Garantir que a inativação não toca em nenhuma tabela de OS, orçamento ou financeiro

**Caso de uso**

- [ ] Implementar `InativarVeiculo`
- [ ] Implementar `ReativarVeiculo`

**Repositório**

- [ ] Implementar `VeiculoRepository.inativar` e `VeiculoRepository.reativar`
- [ ] Implementar a consulta de veículo ativo por placa
- [ ] Ajustar as consultas padrão para filtrar somente veículos ativos
- [ ] Criar índice parcial `UNIQUE (placa) WHERE ativo = true`
- [ ] Remover qualquer `ON DELETE CASCADE` em foreign key que aponte para `veiculo`
- [ ] Implementar registro na trilha de auditoria

**Integrações**

- [ ] Consultar o módulo de Ordem de Serviço para verificar OS em aberto
- [ ] Consumir o evento `ClienteInativado` e inativar os veículos vinculados em cascata
- [ ] Validar a situação do cliente proprietário junto ao `ClienteRepository` na reativação

**Handler HTTP**

- [ ] Implementar `DELETE /veiculos/{veiculoId}`
- [ ] Implementar `POST /veiculos/{veiculoId}/reativacao`
- [ ] Criar DTO/response de saída das duas operações
- [ ] Aplicar autenticação JWT e autorização por escopo nas rotas
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar `veiculoId` em formato UUID
- [ ] Validar `motivo` com no máximo 200 caracteres
- [ ] Bloquear reativação quando houver outro veículo ativo com a mesma placa
- [ ] Bloquear reativação quando o cliente proprietário estiver inativo
- [ ] Restringir a operação ao escopo `veiculos:escrever`

**Eventos**

- [ ] Publicar `VeiculoInativado`
- [ ] Publicar `VeiculoReativado`

**Testes unitários**

- [ ] Veículo sem OS em aberto pode ser inativado
- [ ] Veículo com OS `AGUARDANDO_APROVACAO` não pode ser inativado
- [ ] Veículo com OS apenas `ENTREGUE` pode ser inativado
- [ ] Inativar veículo já inativo não altera os campos de auditoria originais
- [ ] Reativação bloqueada com placa já usada por veículo ativo
- [ ] Reativação bloqueada com cliente proprietário inativo

**Testes de integração**

- [ ] `DELETE` válido retornando `200` com `ativo` falso
- [ ] `DELETE` em veículo já inativo retornando `204`
- [ ] `DELETE` em veículo com OS aberta retornando `409` com a lista de OS
- [ ] `DELETE` em veículo inexistente retornando `404`
- [ ] Perfil sem escopo retornando `403`
- [ ] `POST /reativacao` válido retornando `200` com `ativo` verdadeiro
- [ ] Reativação com placa em uso retornando `409`
- [ ] Reativação com cliente inativo retornando `422`

**Testes de regressão**

- [ ] OS históricas continuam retornando placa e modelo após a inativação
- [ ] A mesma placa pode ser cadastrada em um veículo novo após a inativação
- [ ] Veículo inativo não aparece em `GET /veiculos` sem filtro explícito
- [ ] Não é possível abrir OS para veículo inativo
- [ ] Inativar o cliente inativa os veículos dele em cascata

**Documentação**

- [ ] Documentar os dois endpoints no Swagger/OpenAPI
- [ ] Documentar explicitamente que o `DELETE` é exclusão lógica
- [ ] Documentar o exemplo de erro `409` `veiculo-com-os-aberta`

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Migration versionada e reversível

---
