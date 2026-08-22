---
documento: Refinamento de Requisitos — Consultar Veículo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Consultar Veículo

Este documento detalha a tarefa Consultar Veículo do contexto de Veículo.

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
GET /veiculos
```

Consulta por placa via query param.

> **Decisão de projeto.** A rota segue o padrão compartilhado do projeto: recurso no plural,
> em minúsculas e **sem prefixo de versão**. A alternativa com prefixo `/api/v1` foi descartada
> para manter todas as rotas do sistema no mesmo formato.

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
2. Normalizar e validar presença e formato da placa, nos padrões Mercosul `ABC1D23` e antigo `ABC1234`.
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
  "ano": 2020,
  "version": 3
}
```

> **Decisão de projeto.** A consulta expõe `version` para que a atualização possa enviá-la no
> header `If-Match`. A resposta de recurso único vai **direta**, sem envelope (D-21).

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

- [ ] Implementar `GET /veiculos`
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


