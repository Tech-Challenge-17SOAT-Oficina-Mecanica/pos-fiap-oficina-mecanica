---
documento: Refinamento de Requisitos — Cadastrar Veículo e Vincular ao Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Veículo e Vincular ao Cliente

Este documento detalha a tarefa Cadastrar Veículo e Vincular ao Cliente, do contexto de Veículo.

## 5 · Cadastrar Veículo e Vincular ao Cliente

### 5.1 Refinamento de Produto

**Persona**

Gestor da oficina.

**Objetivo**

Cadastrar um veículo e vinculá-lo corretamente a um cliente para que possa ser utilizado nas
Ordens de Serviço e no histórico de atendimentos.

**Problema**

A oficina precisa manter os veículos associados aos respectivos clientes, evitando informações
inconsistentes e garantindo que cada atendimento seja realizado para o veículo correto. Se o
cadastro e o vínculo forem operações independentes, uma falha intermediária pode deixar um
veículo sem proprietário no sistema.

**Pré-condições**

- O gestor deve possuir acesso ao gerenciamento de veículos.
- O cliente deve estar previamente cadastrado e ativo.
- O cliente deve poder ser localizado por CPF/CNPJ.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-VEI-20 | Permitir localizar o cliente por CPF/CNPJ antes do cadastro do veículo. |
| RF-VEI-21 | Permitir cadastrar um veículo informando placa, marca, modelo e ano. |
| RF-VEI-22 | Vincular o veículo ao cliente durante o cadastro. |
| RF-VEI-23 | Validar se o cliente existe e está ativo. |
| RF-VEI-24 | Validar os dados obrigatórios do veículo. |
| RF-VEI-25 | Impedir o cadastro de veículo ativo com placa já cadastrada. |
| RF-VEI-26 | Disponibilizar o veículo vinculado para criação de Ordem de Serviço e histórico de atendimento. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-VEI-17 | Retornar mensagens claras e respostas padronizadas em caso de erro. |
| RNF-VEI-18 | Persistir o veículo e o vínculo com o cliente na mesma transação. |
| RNF-VEI-19 | Restringir a operação a usuário autenticado e autorizado. |
| RNF-VEI-20 | Impedir que uma falha deixe veículo cadastrado sem vínculo ou vínculo sem veículo. |
| RNF-VEI-21 | Preservar a unicidade da placa entre veículos ativos. |

**Fluxo Principal**

1. O gestor acessa o gerenciamento de veículos.
2. O gestor informa o CPF/CNPJ do cliente.
3. O sistema valida o documento e localiza o cliente.
4. O sistema apresenta o cliente encontrado.
5. O gestor informa placa, marca, modelo e ano do veículo.
6. O sistema valida os dados informados.
7. O sistema verifica se a placa já está cadastrada para um veículo ativo.
8. O sistema cadastra o veículo.
9. O sistema vincula o veículo ao cliente na mesma operação.
10. O sistema confirma o cadastro e o vínculo.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Cliente não encontrado ou inativo | Impede o cadastro e informa que o cliente não está disponível para vínculo. |
| A2 | CPF/CNPJ inválido na consulta prévia | Informa que o documento é inválido e não segue para o cadastro. |
| A3 | Placa já cadastrada para veículo ativo | Impede o cadastro duplicado e informa o conflito. |
| A4 | Dados obrigatórios ausentes | Informa os campos que devem ser preenchidos. |
| A5 | Dados do veículo inválidos | Informa os campos que precisam ser corrigidos. |
| A6 | Usuário sem autenticação ou autorização | Impede a operação. |
| A7 | Falha ao cadastrar ou vincular | Reverte toda a operação e não deixa dados parcialmente persistidos. |

**Saída**

- Veículo cadastrado e vinculado ao cliente; ou
- Indicação do motivo pelo qual a operação não pôde ser concluída.

**Pós-condições**

- O veículo passa a fazer parte do cadastro da oficina.
- O veículo fica vinculado ao cliente informado.
- O veículo pode ser utilizado na criação de uma Ordem de Serviço.
- O histórico de atendimentos pode ser associado ao veículo.

---

### 5.2 Refinamento Técnico

**Endpoint**

```http
POST /clientes/{clienteId}/veiculos
```

O endpoint cria um veículo e seu vínculo com um cliente existente na mesma operação.

> **Decisão de projeto.** Esta rota representa a criação de um veículo subordinado ao cliente.
> `POST /veiculos` permanece responsável pelo cadastro sem vínculo, enquanto
> `POST /clientes/{clienteId}/veiculos/{veiculoId}` vincula um veículo já existente. A
> alternativa de reutilizar `POST /veiculos` com `clienteId` no corpo criaria dois contratos
> diferentes para a mesma rota.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil: `GESTOR`.
- Escopo da operação: `veiculos:escrever`.
- A consulta prévia por CPF/CNPJ exige `clientes:ler`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `clienteId` | UUID | Identificador obrigatório do cliente previamente localizado. |
| Body | `placa` | string | Placa obrigatória do veículo. |
| Body | `marca` | string | Marca obrigatória do veículo. |
| Body | `modelo` | string | Modelo obrigatório do veículo. |
| Body | `ano` | inteiro | Ano obrigatório do veículo. |

```json
{
  "placa": "ABC1D23",
  "marca": "Toyota",
  "modelo": "Corolla",
  "ano": 2024
}
```

O cliente é localizado previamente por:

```http
GET /clientes?documento={cpfOuCnpj}
```

O identificador retornado nessa consulta é usado como `clienteId` na rota de cadastro e vínculo.

**Validações**

*Técnicas*

- `clienteId` deve ser um UUID válido.
- `placa`, `marca`, `modelo` e `ano` são obrigatórios.
- A placa deve ser normalizada antes da validação e da verificação de duplicidade.
- O ano deve possuir formato numérico válido.

*Negócio*

- O cliente deve existir e estar ativo.
- A placa deve respeitar o padrão de placas aceito pelo sistema.
- Não pode existir outro veículo ativo com a mesma placa normalizada.
- Marca e modelo não podem ficar vazios.
- O ano deve respeitar os limites definidos pelo negócio.
- O usuário deve possuir autorização para cadastrar veículos.

**Processamento**

1. Validar `clienteId`, o corpo e a autorização do usuário.
2. Consultar o cliente pelo identificador.
3. Validar se o cliente existe e está ativo.
4. Normalizar e validar os dados do veículo.
5. Verificar duplicidade da placa entre veículos ativos.
6. Criar a entidade `Veiculo`.
7. Associar o veículo ao cliente.
8. Persistir o veículo e o vínculo na mesma transação.
9. Montar e retornar o DTO do veículo cadastrado.

**Persistência**

- Consulta: cliente pelo identificador e veículo ativo pela placa normalizada.
- Altera: `veiculo` e o relacionamento entre veículo e cliente.
- Não altera: dados cadastrais do cliente ou Ordens de Serviço existentes.
- A criação e o vínculo são atômicos: falha em qualquer etapa reverte toda a operação.

**Saída da API**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "clienteId": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
  "placa": "ABC1D23",
  "marca": "Toyota",
  "modelo": "Corolla",
  "ano": 2024,
  "ativo": true
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Veículo cadastrado e vinculado ao cliente. |
| `400` | Identificador ou dados do veículo ausentes ou inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `veiculos:escrever`. |
| `404` | Cliente não encontrado. |
| `409` | Placa já cadastrada para outro veículo ativo. |
| `422` | Cliente inativo ou ano fora das regras de negócio. |
| `500` | Falha inesperada, sem persistência parcial do veículo ou vínculo. |

**Dependências**

- `VeiculoRepository`.
- `ClienteRepository`.
- Validador de CPF/CNPJ usado na consulta prévia.
- Validador e normalizador de placa.
- Middlewares de autenticação e autorização.
- Gerenciador de transações.

**Testes**

*Unitários*

- Cadastra veículo válido e associa ao cliente.
- Rejeita cliente inexistente ou inativo.
- Valida placa, marca, modelo, ano e campos obrigatórios.
- Rejeita placa duplicada entre veículos ativos.
- Reverte cadastro quando o vínculo falha.

*Integração*

- Requisição válida retorna `201` com veículo e `clienteId`.
- Cliente inexistente retorna `404`.
- Placa duplicada retorna `409`.
- Dados inválidos retornam `400` ou `422`, conforme a regra.
- Usuário sem autorização retorna `403`.
- Veículo cadastrado aparece vinculado na consulta do cliente.
- Falha de persistência não deixa veículo sem vínculo.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar a entidade `Veiculo`
- [ ] Definir o relacionamento entre `Veiculo` e `Cliente`
- [ ] Implementar a normalização e a unicidade da placa entre veículos ativos

**Caso de uso**

- [ ] Implementar `CadastrarVeiculoEVincularAoCliente`
- [ ] Consultar e validar o cliente
- [ ] Criar o veículo e realizar o vínculo atomicamente
- [ ] Retornar o veículo cadastrado com `clienteId`

**Repositório**

- [ ] Implementar ou ajustar `VeiculoRepository`
- [ ] Integrar a consulta ao `ClienteRepository`
- [ ] Verificar duplicidade por placa normalizada
- [ ] Persistir veículo e vínculo na mesma transação

**Integrações**

- [ ] Integrar a consulta prévia do cliente por CPF/CNPJ

**Handler HTTP**

- [ ] Implementar `POST /clientes/{clienteId}/veiculos`
- [ ] Criar DTO de entrada e DTO de resposta
- [ ] Aplicar autenticação e autorização
- [ ] Mapear erros para os códigos HTTP documentados

**Validações**

- [ ] Validar `clienteId`, placa, marca, modelo e ano
- [ ] Validar existência e situação ativa do cliente
- [ ] Validar formato e unicidade da placa
- [ ] Validar limites de negócio do ano

**Transação e idempotência**

- [ ] Garantir atomicidade entre o cadastro e o vínculo
- [ ] Garantir rollback integral quando cadastro ou vínculo falhar

**Testes unitários**

- [ ] Cadastro e vínculo válidos
- [ ] Cliente inexistente ou inativo
- [ ] Placa inválida ou duplicada
- [ ] Campos obrigatórios e ano inválido
- [ ] Rollback quando o vínculo falhar

**Testes de integração**

- [ ] Respostas `201`, `400`, `401`, `403`, `404`, `409`, `422` e `500`
- [ ] Veículo vinculado aparece na consulta do cliente
- [ ] Ausência de persistência parcial diante de falha

**Documentação**

- [ ] Documentar o endpoint no OpenAPI/Swagger

**Review**

- [ ] Code Review aprovado

---
