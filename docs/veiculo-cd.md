---
documento: Refinamento de Requisitos — Contexto de Veículo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos

Este documento reúne, para cada requisito levantado da aplicação, três blocos:

1. **Refinamento de Produto** — o que o usuário precisa e por quê (visão de negócio).
2. **Refinamento Técnico** — como o sistema entrega isso (contrato, processamento, testes).
3. **Checklist de Implementação** — o passo a passo verificável até o merge.

---

## 1 · Consultar Veículo

### 1.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Consultar o cadastro do veículo a partir da placa informada.

**Problema**
A oficina precisa identificar o veículo antes de seguir com o atendimento, garantindo que ele
seja corretamente vinculado ao cliente e à Ordem de Serviço. Sem essa identificação, o histórico
do veículo pode ficar incompleto ou ser associado ao cliente errado.

**Pré-condições**

- A placa do veículo deve ser informada.
- O usuário deve estar autorizado a consultar veículos.
- Deve existir cadastro de veículos disponível para consulta.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-VEI-01 | Permitir ao mecânico consultar veículo pela placa. |
| RF-VEI-02 | Validar a placa informada. |
| RF-VEI-03 | Consultar o cadastro de veículos. |
| RF-VEI-04 | Identificar o veículo quando houver cadastro correspondente. |
| RF-VEI-05 | Informar quando o veículo não for identificado. |
| RF-VEI-06 | Permitir seguir para cadastro do veículo quando ele não for encontrado. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-VEI-01 | A consulta deve ser feita por API RESTful. |
| RNF-VEI-02 | A placa do veículo deve ser validada. |
| RNF-VEI-03 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-VEI-04 | A consulta não deve alterar os dados cadastrais do veículo. |

**Fluxo Principal**

1. A placa do veículo é informada.
2. O mecânico solicita a consulta do veículo.
3. O sistema valida a placa informada.
4. O sistema consulta o cadastro de veículos.
5. O sistema encontra o veículo correspondente.
6. O sistema identifica o veículo.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Placa inválida | O sistema informa que a placa informada não é válida e não consulta o cadastro. |
| A2 | Veículo não identificado | O sistema informa que o veículo não foi identificado. |
| A3 | Veículo não encontrado | O sistema permite seguir para o cadastro do veículo. |
| A4 | Usuário sem autorização | O sistema impede a consulta. |

**Saída**

- Veículo identificado a partir da placa informada; **ou**
- Indicação de que o veículo não foi identificado.

**Pós-condições**

- Os dados do veículo permanecem inalterados.
- O veículo identificado fica disponível para vínculo com cliente.
- O veículo identificado fica disponível para criação da Ordem de Serviço.
- Caso o veículo não seja encontrado, o fluxo pode seguir para Cadastrar Veículo.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
GET /api/v1/veiculos
```

Consulta por placa via query param.

> **Decisão de projeto.** Foi adotada a rota plural com prefixo versionado
> `GET /api/v1/veiculos?placa=...`, alinhada ao padrão compartilhado do projeto. A alternativa
> `GET /veiculos` foi descartada por não usar o prefixo `/api/v1/`.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfil: `MECANICO`
- Escopo: `veiculos:ler`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Query | `placa` | string | Placa do veículo, obrigatória. |

**Validações**

- `placa` deve ser informada.
- `placa` deve possuir formato válido.
- O veículo deve existir para que a consulta retorne sucesso.
- A operação não altera dados do veículo.

**Processamento**

1. Receber a placa informada no query param `placa`.
2. Validar presença e formato da placa.
3. Consultar o cadastro de veículos pela placa.
4. Caso encontre o veículo, retornar os dados cadastrais.
5. Caso não encontre, retornar erro informando que o veículo não foi encontrado.

**Persistência**

- Consulta: agregado/dados de `Veículo`.
- Altera: nada.

**Saída da API**

```json
{
  "id": "uuid-do-veiculo",
  "placa": "ABC1D23",
  "marca": "Marca do Veículo",
  "modelo": "Modelo do Veículo",
  "ano": 2020
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Veículo encontrado. |
| `400` | Placa ausente ou inválida. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `veiculos:ler`. |
| `404` | Veículo não encontrado. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de veículos.
- `VeiculoRepository`.
- Validador de placa.
- Caso de uso Cadastrar Veículo, quando o veículo não for encontrado.

**Testes**

*Unitários*

- Retorna veículo quando placa válida existe no cadastro.
- Retorna erro quando a placa não for informada.
- Retorna erro quando a placa for inválida.
- Retorna erro quando veículo não for encontrado.
- Garante que a consulta não altera os dados do veículo.

*Integração*

- `GET` válido retorna `200` com os dados esperados do veículo.
- Veículo inexistente retorna `404`.
- Placa ausente retorna `400`.
- Placa inválida retorna `400`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `Veículo`
- [ ] Definir os campos necessários para identificação do veículo por placa
- [ ] Garantir que o veículo possua placa como identificador de negócio
- [ ] Criar validação de placa de veículo
- [ ] Garantir que a consulta não altera dados do veículo

**Caso de uso**

- [ ] Implementar `ConsultarVeiculo`
- [ ] Receber a placa como critério de consulta
- [ ] Consultar o veículo cadastrado pela placa
- [ ] Retornar erro quando o veículo não for encontrado

**Repositório**

- [ ] Criar ou ajustar `VeiculoRepository` para consulta por placa
- [ ] Implementar método de busca de veículo por placa

**Handler HTTP**

- [ ] Implementar `GET /api/v1/veiculos`
- [ ] Implementar leitura do parâmetro `placa` via query param
- [ ] Criar DTO/request para entrada da consulta
- [ ] Criar DTO/response de saída com os dados do veículo
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `veiculos:ler`
- [ ] Mapear erros de validação para os códigos HTTP documentados

**Validações**

- [ ] Validar que a placa foi informada
- [ ] Validar formato da placa
- [ ] Retornar `400` para placa ausente ou inválida
- [ ] Retornar `404` quando o veículo não for encontrado
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Testes unitários**

- [ ] Consulta de veículo existente
- [ ] Placa ausente
- [ ] Placa inválida
- [ ] Veículo não encontrado
- [ ] Consulta não altera dados do veículo

**Testes de integração**

- [ ] Endpoint retorna os dados esperados do veículo
- [ ] Endpoint retorna `400` para placa ausente ou inválida
- [ ] Endpoint retorna `404` quando o veículo não existe
- [ ] Endpoint retorna `401` sem autenticação
- [ ] Endpoint retorna `403` sem permissão
- [ ] Consulta não altera dados persistidos

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI
- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite da task
- [ ] Code Review aprovado

---

## 2 · Cadastrar Veículo

### 2.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Cadastrar um veículo no sistema quando ele não for identificado na consulta pela placa.

**Problema**
A oficina precisa manter o cadastro dos veículos atendidos para permitir vínculo com cliente,
criação de Ordem de Serviço e preservação do histórico de atendimento. Sem o cadastro, o veículo
não pode ser acompanhado corretamente ao longo dos atendimentos.

**Pré-condições**

- A placa do veículo deve ter sido informada.
- O veículo não deve ter sido identificado na consulta.
- A placa deve ser válida.
- O usuário deve estar autorizado a cadastrar veículos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-VEI-07 | Permitir ao mecânico cadastrar um novo veículo. |
| RF-VEI-08 | Validar a placa informada. |
| RF-VEI-09 | Registrar placa, marca, modelo e ano do veículo. |
| RF-VEI-10 | Impedir cadastro duplicado para a mesma placa. |
| RF-VEI-11 | Confirmar que o veículo foi cadastrado. |
| RF-VEI-12 | Permitir que o veículo cadastrado seja vinculado ao cliente. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-VEI-05 | A operação deve ser feita por API RESTful. |
| RNF-VEI-06 | A placa do veículo deve ser validada. |
| RNF-VEI-07 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-VEI-08 | O cadastro deve ser persistido de forma consistente. |
| RNF-VEI-09 | O sistema deve evitar duplicidade de veículos. |

**Fluxo Principal**

1. O mecânico consulta o veículo pela placa.
2. O sistema informa que o veículo não foi identificado.
3. O mecânico solicita o cadastro do veículo.
4. O mecânico informa placa, marca, modelo e ano.
5. O sistema valida a placa informada.
6. O sistema verifica se já existe veículo cadastrado com a mesma placa.
7. O sistema registra o novo veículo.
8. O sistema confirma que o veículo foi cadastrado.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Placa inválida | O sistema informa que a placa informada não é válida e não cadastra o veículo. |
| A2 | Veículo já cadastrado | O sistema impede novo cadastro para a mesma placa. |
| A3 | Dados obrigatórios ausentes | O sistema informa que placa, marca, modelo e ano devem ser informados. |
| A4 | Usuário sem autorização | O sistema impede o cadastro. |
| A5 | Erro ao cadastrar veículo | O sistema informa que não foi possível concluir o cadastro. |

**Saída**

- Veículo cadastrado no sistema.
- Confirmação de cadastro realizado.

**Pós-condições**

- O veículo passa a existir no cadastro da oficina.
- O veículo fica disponível para consulta.
- O veículo pode ser vinculado ao cliente.
- O fluxo pode seguir para Vincular Veículo ao Cliente.

---

### 2.2 Refinamento Técnico

**Endpoint**

```http
POST /api/v1/veiculos
```

> **Decisão de projeto.** Foi adotada a rota plural com prefixo versionado
> `POST /api/v1/veiculos`, alinhada ao padrão compartilhado do projeto. A alternativa
> `POST /veiculos` foi descartada por não usar o prefixo `/api/v1/`.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório
- Perfil: `MECANICO`
- Escopo: `veiculos:escrever`

**Entrada**

| Local | Param | Tipo | Descrição |
|---|---|---|---|
| Body | `placa` | string | Placa do veículo, obrigatória. |
| Body | `marca` | string | Marca do veículo, obrigatória. |
| Body | `modelo` | string | Modelo do veículo, obrigatório. |
| Body | `ano` | int | Ano do veículo, obrigatório. |

```json
{
  "placa": "ABC1D23",
  "marca": "Marca do Veículo",
  "modelo": "Modelo do Veículo",
  "ano": 2020
}
```

**Validações**

- `placa` deve ser informada.
- `marca` deve ser informada.
- `modelo` deve ser informado.
- `ano` deve ser informado.
- `placa` deve possuir formato válido.
- Não pode existir veículo cadastrado com a mesma placa.

**Processamento**

1. Receber os dados do veículo.
2. Validar os campos obrigatórios.
3. Validar a placa informada.
4. Consultar se já existe veículo com a mesma placa.
5. Criar o cadastro do veículo.
6. Persistir o novo veículo.
7. Retornar os dados do veículo cadastrado.

**Persistência**

- Consulta: agregado/dados de `Veículo` para verificar duplicidade.
- Altera: `Veículo` com novo registro.
- Persiste: identificador do veículo, placa, marca, modelo e ano.

**Saída da API**

```json
{
  "id": "uuid-do-veiculo",
  "placa": "ABC1D23",
  "marca": "Marca do Veículo",
  "modelo": "Modelo do Veículo",
  "ano": 2020
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Veículo cadastrado com sucesso. |
| `400` | Dados obrigatórios ausentes ou placa inválida. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `veiculos:escrever`. |
| `409` | Veículo já cadastrado com a placa informada. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de veículos.
- `VeiculoRepository`.
- Validador de placa.
- Caso de uso Consultar Veículo, para verificar se o veículo já existe.

**Testes**

*Unitários*

- Cadastra veículo quando os dados são válidos.
- Rejeita cadastro quando placa não for informada.
- Rejeita cadastro quando marca não for informada.
- Rejeita cadastro quando modelo não for informado.
- Rejeita cadastro quando ano não for informado.
- Rejeita cadastro quando placa for inválida.
- Rejeita cadastro quando já existir veículo com a mesma placa.

*Integração*

- `POST` válido retorna `201` e persiste o veículo.
- Veículo cadastrado pode ser consultado por placa.
- Placa ausente retorna `400`.
- Marca ausente retorna `400`.
- Modelo ausente retorna `400`.
- Ano ausente retorna `400`.
- Placa inválida retorna `400`.
- Veículo duplicado retorna `409`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `Veículo`
- [ ] Definir os campos necessários para cadastro do veículo
- [ ] Garantir que o veículo possua placa como identificador de negócio
- [ ] Criar validação de placa de veículo
- [ ] Impedir cadastro duplicado de veículo

**Caso de uso**

- [ ] Implementar `CadastrarVeiculo`
- [ ] Receber os dados necessários do veículo
- [ ] Verificar se já existe veículo cadastrado com a mesma placa
- [ ] Criar novo veículo
- [ ] Persistir o veículo no banco de dados

**Repositório**

- [ ] Criar ou ajustar `VeiculoRepository` para persistência do veículo
- [ ] Criar método para consultar veículo por placa
- [ ] Criar método para salvar novo veículo

**Handler HTTP**

- [ ] Implementar `POST /api/v1/veiculos`
- [ ] Criar DTO/request de entrada
- [ ] Criar DTO/response de saída com os dados do veículo cadastrado
- [ ] Implementar validação do payload
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `veiculos:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar que a placa foi informada
- [ ] Validar que a marca foi informada
- [ ] Validar que o modelo foi informado
- [ ] Validar que o ano foi informado
- [ ] Validar formato da placa
- [ ] Retornar `201` quando o veículo for cadastrado com sucesso
- [ ] Retornar `400` para dados obrigatórios ausentes ou placa inválida
- [ ] Retornar `409` quando já existir veículo com a mesma placa
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Testes unitários**

- [ ] Cadastro válido de veículo
- [ ] Placa ausente
- [ ] Marca ausente
- [ ] Modelo ausente
- [ ] Ano ausente
- [ ] Placa inválida
- [ ] Veículo duplicado

**Testes de integração**

- [ ] Endpoint cadastra veículo válido e retorna `201`
- [ ] Veículo é persistido corretamente no cadastro
- [ ] Veículo cadastrado pode ser consultado por placa
- [ ] Endpoint retorna `400` para dados obrigatórios ausentes ou placa inválida
- [ ] Endpoint retorna `409` para veículo duplicado
- [ ] Endpoint retorna `401` sem autenticação
- [ ] Endpoint retorna `403` sem permissão

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI
- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite da task
- [ ] Code Review aprovado

---

## Pontos em aberto

| # | Ponto | Responsável |
|---|---|---|
| 1 | Confirmar o dono do documento do contexto Veículo. | — |
| 2 | Confirmar se o escopo definitivo para consulta de veículos será `veiculos:ler` ou outro nome padronizado pelo time. | — |
| 3 | Definir se a resposta de consulta de veículo deve usar envelope simples, como documentado aqui, ou algum envelope padronizado para recursos únicos. | — |
| 4 | Definir o padrão definitivo de validação de placa, incluindo placa Mercosul e placa antiga. | — |
| 5 | Confirmar se o escopo definitivo para cadastro de veículos será `veiculos:escrever` ou outro nome padronizado pelo time. | — |
| 6 | Definir as validações de negócio para o campo `ano`, incluindo ano mínimo e se ano futuro é permitido. | — |
