---
documento: Refinamento de Requisitos — Criar Ordem de Serviço
dono: Helena Miranda
versao: 0.3
atualizado_em: 2026-10-28
status: rascunho
---

# Refinamento de Requisitos — Criar Ordem de Serviço

Este documento detalha a tarefa Criar Ordem de Serviço do contexto de Ordem de Serviço.

## 1 · Criar Ordem de Serviço

### 1.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Iniciar uma Ordem de Serviço para registrar a entrada de um veículo vinculado a um cliente,
já com os serviços e as peças que o cliente solicitou na recepção.

**Problema**

A oficina precisa iniciar formalmente o atendimento do veículo para permitir o acompanhamento da OS desde sua abertura.

**Pré-condições**

- O cliente deve estar cadastrado e identificado.
- O veículo deve estar cadastrado e identificado.
- O veículo deve estar vinculado ao cliente.
- O usuário deve estar autorizado a criar Ordem de Serviço.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-01 | Permitir ao mecânico criar uma Ordem de Serviço. |
| RF-OS-02 | Associar a Ordem de Serviço ao cliente. |
| RF-OS-03 | Associar a Ordem de Serviço ao veículo. |
| RF-OS-04 | Validar se o veículo pertence ao cliente informado. |
| RF-OS-05 | Criar a Ordem de Serviço com status inicial `RECEBIDA`. |
| RF-OS-06 | Registrar a data e hora de criação da Ordem de Serviço. |
| RF-OS-07 | Permitir que a Ordem de Serviço seja acompanhada posteriormente. |
| RF-OS-137 | Permitir informar, na abertura, os serviços solicitados, que entram no orçamento `PRINCIPAL`. |
| RF-OS-138 | Permitir informar, na abertura, as peças solicitadas e suas quantidades, que entram no orçamento `PRINCIPAL`. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-01 | A operação deve ser feita por API RESTful. |
| RNF-OS-02 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-OS-03 | A criação da Ordem de Serviço deve ser persistida de forma consistente. |
| RNF-OS-04 | O vínculo entre cliente, veículo e Ordem de Serviço deve preservar o histórico de atendimento. |
| RNF-OS-05 | A criação da Ordem de Serviço deve ocorrer de forma atômica. |
| RNF-OS-66 | A abertura, o orçamento e os itens solicitados são gravados na mesma transação: um item inválido desfaz a abertura inteira. |

**Fluxo Principal**

1. O mecânico seleciona o cliente.
2. O mecânico seleciona o veículo.
3. O mecânico informa, se houver, os serviços e as peças solicitados pelo cliente.
4. O mecânico solicita a criação da Ordem de Serviço.
5. O sistema verifica se o cliente existe.
6. O sistema verifica se o veículo existe.
7. O sistema verifica se o veículo está vinculado ao cliente.
8. O sistema cria a Ordem de Serviço.
9. O sistema define o status inicial como `RECEBIDA`.
10. O sistema registra a data e hora de criação.
11. Havendo serviços ou peças, o sistema cria o orçamento `PRINCIPAL` e registra os itens nele.
12. O sistema confirma a criação da Ordem de Serviço.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente não encontrado | O sistema informa que o cliente não existe. |
| A2 | Veículo não encontrado | O sistema informa que o veículo não existe. |
| A3 | Veículo não vinculado ao cliente | O sistema impede a criação da Ordem de Serviço. |
| A4 | Usuário sem autorização | O sistema impede a operação. |
| A5 | Erro ao criar Ordem de Serviço | O sistema informa que não foi possível concluir a criação. |
| A6 | Serviço ou peça não encontrado | O sistema impede a criação; nenhuma OS fica gravada. |
| A7 | Serviço ou peça inativo, ou peça sem preço de venda | O sistema impede a criação; nenhuma OS fica gravada. |
| A8 | Serviço ou peça repetido, ou quantidade de peça inválida | O sistema impede a criação antes de consultar o banco. |

**Saída**

- Ordem de Serviço criada.
- Ordem de Serviço vinculada ao cliente e ao veículo.
- Status inicial `RECEBIDA` registrado.
- Orçamento `PRINCIPAL` em `CRIADO` com os serviços e as peças solicitados, quando informados.

**Pós-condições**

- A Ordem de Serviço passa a existir no sistema.
- A Ordem de Serviço fica vinculada ao cliente e ao veículo.
- A Ordem de Serviço fica disponível para consulta e acompanhamento.
- O fluxo pode seguir para Iniciar Diagnóstico ou Registrar problema identificado, conforme o processo definido.

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico
```

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Permitido para usuário com perfil/permissão de Mecânico.
- Requer escopo `os:escrever`.

**Entrada**

Body:

```json
{
  "clienteId": "uuid-do-cliente",
  "veiculoId": "uuid-do-veiculo",
  "servicos": [
    { "servicoId": "uuid-do-servico", "observacao": "opcional" }
  ],
  "pecas": [
    { "pecaId": "uuid-da-peca", "quantidade": 2 }
  ]
}
```

`servicos` e `pecas` são opcionais. Sem eles, a abertura funciona como antes e nenhum orçamento é
criado.

**Validações**

- Validar autenticação e permissão do mecânico.
- Validar se `clienteId` e `veiculoId` foram informados.
- Validar o formato dos identificadores, inclusive `servicoId` e `pecaId`.
- Validar se o cliente existe.
- Validar se o veículo existe.
- Validar se o veículo está vinculado ao cliente informado.
- Impedir serviço ou peça repetido na mesma abertura.
- Exigir quantidade de peça inteira e maior que zero.
- Validar que cada serviço e cada peça existe, está ativo e, no caso da peça, tem preço de venda.
- Não exigir problema relatado, serviços, peças ou insumos na criação da OS.
- Gerar o orçamento `PRINCIPAL` somente quando a abertura trouxer serviços ou peças.

> **Decisão de projeto.** A OS grava a placa vigente do veículo no momento em que é criada. O
> cadastro do veículo permite corrigir a placa depois, e sem essa cópia a correção reescreveria o
> histórico de todas as OS anteriores.

> **Decisão de projeto (Fase 2).** O enunciado da Fase 2 pede que a abertura receba os serviços e
> as peças. A OS continua nascendo em `RECEBIDA`: a máquina de estados (D-18) só leva a
> `EM_DIAGNOSTICO` pelo registro do problema relatado, e a visão geral já previa a etapa "abertura
> da OS e registro dos serviços solicitados". Os itens entram no orçamento `PRINCIPAL`, o mesmo que
> o diagnóstico completa depois; o registro de peças já criava esse orçamento no primeiro item, e
> o de serviços passou a fazer o mesmo quando a OS está em `RECEBIDA`. Calcular e enviar o
> orçamento continuam exigindo `EM_DIAGNOSTICO`.

**Processamento**

- Receber `clienteId`, `veiculoId` e, opcionalmente, `servicos` e `pecas`.
- Validar o usuário autenticado.
- Consultar o cliente e o veículo.
- Validar o vínculo entre cliente e veículo.
- Criar a Ordem de Serviço com status `RECEBIDA`.
- Associar o cliente e o veículo à Ordem de Serviço.
- Gravar na OS a **placa vigente do veículo** no momento da criação.
- Registrar a data e hora de criação.
- Havendo serviços ou peças, criar o orçamento `PRINCIPAL` e registrar os itens com as mesmas
  regras de [Registrar Serviços Necessários](registrar-servicos-necessarios.md) e
  [Registrar Peças e Insumos Necessários](registrar-pecas-e-insumos-necessarios.md).
- Persistir a OS, o orçamento e os itens em uma única transação.
- Retornar os dados da OS criada.

**Persistência**

- Consultar Cliente.
- Consultar Veículo e seu vínculo com o Cliente.
- Criar Ordem de Serviço com:
  - `id`;
  - `clienteId`;
  - `veiculoId`;
  - `status = RECEBIDA`;
  - `criadaEm`.
- Com serviços ou peças: criar `orcamento` (`PRINCIPAL`, `CRIADO`), gravar `orcamento_item` para
  cada serviço e peça, e `ordem_servico_item` para cada peça.
- Sem serviços nem peças: nenhum orçamento é criado.

**Saída da API**

```json
{
  "ordemServicoId": "uuid-da-os",
  "clienteId": "uuid-do-cliente",
  "veiculoId": "uuid-do-veiculo",
  "status": "RECEBIDA",
  "criadaEm": "2026-08-22T10:30:00-03:00"
}
```

**Códigos HTTP / Erros**

- `201 Created`: Ordem de Serviço criada com sucesso.
- `400 Bad Request`: identificadores ausentes ou inválidos; serviço ou peça repetido; quantidade
  de peça inválida.
- `401 Unauthorized`: usuário não autenticado.
- `403 Forbidden`: usuário sem permissão.
- `404 Not Found`: cliente, veículo, serviço ou peça não encontrado.
- `409 Conflict`: veículo não vinculado ao cliente; serviço ou peça inativo; peça sem preço de venda.
- `500 Internal Server Error`: erro inesperado.

**Dependências**

- Módulo de autenticação JWT.
- Módulo de autorização.
- ClienteRepository.
- VeiculoRepository.
- OrdemDeServicoRepository.
- Contexto do usuário autenticado.
- Banco de dados.

**Testes**

- Deve criar Ordem de Serviço com cliente e veículo válidos e vinculados.
- Deve criar a OS com status inicial `RECEBIDA`.
- Deve registrar a data e hora de criação.
- Deve associar a OS ao cliente e ao veículo.
- Deve permitir criar a OS sem problema relatado, serviços, peças ou insumos.
- Deve garantir que nenhum orçamento seja gerado na criação da OS sem serviços nem peças.
- Deve criar o orçamento `PRINCIPAL` com os serviços e as peças informados na abertura.
- Deve desfazer a abertura inteira quando um serviço ou peça for inexistente, inativo ou sem preço.
- Deve retornar `400 Bad Request` para `clienteId` ou `veiculoId` ausente ou inválido.
- Deve retornar `404 Not Found` para cliente ou veículo inexistente.
- Deve retornar `409 Conflict` quando o veículo não estiver vinculado ao cliente.
- Deve retornar `401 Unauthorized` sem autenticação.
- Deve retornar `403 Forbidden` sem permissão.
- Deve garantir a persistência atômica da Ordem de Serviço.
- Deve possuir teste de integração do endpoint.

### 1.3 Check-list de Implementação

**Domínio e Persistência**

- [ ] Criar/ajustar o modelo `OrdemDeServico`.
- [ ] Definir os campos necessários para abertura da Ordem de Serviço.
- [ ] Definir o status inicial como `RECEBIDA`.
- [ ] Definir os relacionamentos da Ordem de Serviço com Cliente e Veículo.
- [ ] Criar/ajustar `OrdemDeServicoRepository` para persistência da OS.
- [ ] Criar/ajustar `ClienteRepository` para consulta do cliente.
- [ ] Criar/ajustar `VeiculoRepository` para consulta do veículo e vínculo com o cliente.

**Caso de Uso**

- [ ] Implementar o caso de uso `CriarOrdemDeServico`.
- [ ] Garantir que o caso de uso receba `clienteId`, `veiculoId` e, opcionalmente, serviços e peças.
- [ ] Validar os identificadores informados.
- [ ] Validar que cliente e veículo existem.
- [ ] Validar que o veículo está vinculado ao cliente.
- [ ] Criar a OS com status `RECEBIDA`.
- [ ] Associar cliente e veículo à Ordem de Serviço.
- [ ] Registrar a data e hora de criação.
- [ ] Persistir a criação de forma atômica.
- [ ] Garantir que problema relatado, serviços, peças e insumos não sejam obrigatórios.
- [ ] Garantir que orçamento só seja gerado quando houver serviços ou peças na abertura.
- [ ] Registrar serviços e peças na mesma transação da abertura.

**API e Segurança**

- [ ] Criar DTO de entrada com `clienteId`, `veiculoId`, `servicos` e `pecas`.
- [ ] Criar DTO de resposta da Ordem de Serviço criada.
- [ ] Criar handler para `POST /ordens-servico`.
- [ ] Implementar validação do payload.
- [ ] Aplicar autenticação JWT na rota.
- [ ] Aplicar autorização para o perfil/permissão de Mecânico.
- [ ] Retornar `201` para criação bem-sucedida.
- [ ] Mapear erros para `400`, `401`, `403`, `404` e `409`.

**Testes e Documentação**

- [ ] Criar testes unitários para criação válida, identificadores ausentes ou inválidos, cliente inexistente, veículo inexistente e vínculo inválido.
- [ ] Criar testes para status inicial, data de criação e associação com cliente e veículo.
- [ ] Criar testes para garantir que problema, serviços, peças e insumos não são obrigatórios.
- [ ] Criar teste para garantir que orçamento não é gerado na criação sem itens.
- [ ] Criar teste de integração da abertura com serviços e peças, incluindo o rollback por item inválido.
- [ ] Criar testes de autenticação e autorização.
- [ ] Criar teste de integração do endpoint.
- [ ] Documentar o endpoint no Swagger/OpenAPI.
- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto.
- [ ] Executar testes automatizados, realizar code review e validar critérios de aceite da task.
