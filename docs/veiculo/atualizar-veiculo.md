---
documento: Refinamento de Requisitos — Atualizar Veículo
dono: A definir
versao: 0.1
atualizado_em: 2026-08-20
status: rascunho
---

# Refinamento de Requisitos — Atualizar Veículo

Este documento detalha a tarefa Atualizar Veículo do contexto de Veículo.

## 4 · Atualizar Veículo

### 4.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Atualizar os dados cadastrais de um veículo já existente no sistema.

**Problema**
A oficina precisa manter os dados do veículo corretos e atualizados para garantir identificação,
vínculo com cliente, criação de Ordem de Serviço e preservação do histórico de atendimento. Dados
incorretos podem gerar vínculos errados, perda de rastreabilidade e abertura de OS para o veículo
equivocado.

**Pré-condições**

- O veículo deve estar cadastrado no sistema.
- O veículo deve ter sido identificado.
- O usuário deve estar autorizado a atualizar veículos.
- Os novos dados informados devem ser válidos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-VEI-13 | Permitir ao mecânico atualizar dados de um veículo cadastrado. |
| RF-VEI-14 | Consultar o veículo antes da atualização. |
| RF-VEI-15 | Validar os dados alterados. |
| RF-VEI-16 | Validar a placa quando esse dado for informado ou alterado. |
| RF-VEI-17 | Persistir as alterações no cadastro do veículo. |
| RF-VEI-18 | Confirmar que o veículo foi atualizado. |
| RF-VEI-19 | Manter o vínculo do veículo com cliente e Ordens de Serviço. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-VEI-10 | A operação deve ser feita por API RESTful. |
| RNF-VEI-11 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-VEI-12 | A placa do veículo deve ser validada. |
| RNF-VEI-13 | A atualização deve ser persistida de forma consistente. |
| RNF-VEI-14 | A operação não deve remover o histórico do veículo. |

**Fluxo Principal**

1. O mecânico consulta o veículo.
2. O sistema identifica o veículo.
3. O mecânico solicita a atualização dos dados do veículo.
4. O mecânico informa os dados que devem ser alterados.
5. O sistema valida os dados informados.
6. O sistema verifica se a placa informada já pertence a outro veículo.
7. O sistema atualiza o cadastro do veículo.
8. O sistema confirma que o veículo foi atualizado.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Veículo não encontrado | O sistema informa que o veículo não existe. |
| A2 | Veículo não identificado | O sistema impede a atualização até que o veículo seja identificado. |
| A3 | Placa inválida | O sistema informa que a placa informada não é válida. |
| A4 | Placa já vinculada a outro veículo | O sistema impede a atualização para evitar duplicidade. |
| A5 | Dados obrigatórios ausentes | O sistema informa quais dados precisam ser preenchidos. |
| A6 | Usuário sem autorização | O sistema impede a atualização. |
| A7 | Erro ao atualizar veículo | O sistema informa que não foi possível concluir a atualização. |

**Saída**

- Veículo atualizado no sistema.
- Confirmação de atualização realizada.

**Pós-condições**

- Os dados do veículo ficam atualizados.
- O veículo permanece vinculado ao cliente e às Ordens de Serviço existentes.
- O veículo atualizado fica disponível para consulta e continuidade do atendimento.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
PUT /veiculos/{veiculoId}
```

O endpoint atualiza os dados cadastrais de um veículo existente e preserva os vínculos existentes
com cliente e Ordens de Serviço.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `veiculos:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `veiculoId` | UUID | Identificador do veículo. |
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

- `veiculoId` deve ser informado e possuir formato válido.
- Deve existir veículo cadastrado para o `veiculoId`.
- `placa` deve ser informada.
- `marca` deve ser informada.
- `modelo` deve ser informado.
- `ano` deve ser informado.
- `placa` deve possuir formato válido.
- A placa informada não pode estar vinculada a outro veículo.
- A operação deve impedir duplicidade de placa entre veículos.

**Processamento**

1. Receber o identificador do veículo.
2. Receber os dados informados para atualização.
3. Validar campos obrigatórios.
4. Consultar o veículo pelo identificador.
5. Validar a placa informada.
6. Verificar se a placa já pertence a outro veículo.
7. Atualizar os dados do veículo.
8. Persistir as alterações.
9. Retornar os dados atualizados.

**Persistência**

- Consulta: agregado/dados de `Veículo`.
- Consulta: `Veículo` por placa para verificar duplicidade.
- Altera: registro de `Veículo`.
- Persiste: placa, marca, modelo e ano.
- Não altera: vínculos existentes do veículo com cliente e Ordens de Serviço.

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
| `200` | Veículo atualizado com sucesso. |
| `400` | Identificador ausente, dados obrigatórios ausentes ou placa inválida. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `veiculos:escrever`. |
| `404` | Veículo não encontrado. |
| `409` | Placa já vinculada a outro veículo. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de veículos.
- `VeiculoRepository`.
- Validador de placa.
- Caso de uso Consultar Veículo, para localizar o veículo antes da atualização.

**Testes**

*Unitários*

- Atualiza veículo quando os dados são válidos.
- Rejeita atualização quando `veiculoId` não for informado.
- Rejeita atualização quando veículo não existir.
- Rejeita atualização quando placa não for informada.
- Rejeita atualização quando marca não for informada.
- Rejeita atualização quando modelo não for informado.
- Rejeita atualização quando ano não for informado.
- Rejeita atualização quando placa for inválida.
- Rejeita atualização quando placa pertencer a outro veículo.
- Preserva vínculos existentes com cliente e Ordens de Serviço.

*Integração*

- `PUT /veiculos/{veiculoId}` válido retorna `200`.
- Dados atualizados são persistidos no banco.
- Veículo inexistente retorna `404`.
- `veiculoId` ausente ou inválido retorna `400`.
- Placa ausente retorna `400`.
- Marca ausente retorna `400`.
- Modelo ausente retorna `400`.
- Ano ausente retorna `400`.
- Placa inválida retorna `400`.
- Placa pertencente a outro veículo retorna `409`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.
- Vínculos com cliente e Ordens de Serviço são preservados.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Criar ou ajustar o modelo `Veículo`
- [ ] Definir quais campos do veículo podem ser atualizados
- [ ] Garantir que o veículo possua placa como identificador de negócio
- [ ] Criar ou ajustar validação de placa de veículo
- [ ] Impedir duplicidade de placa entre veículos
- [ ] Garantir que vínculos existentes com Cliente e Ordens de Serviço sejam preservados

**Caso de uso**

- [ ] Implementar `AtualizarVeiculo`
- [ ] Receber o identificador do veículo
- [ ] Receber os dados atualizados do veículo
- [ ] Validar que o veículo existe
- [ ] Verificar se a placa informada já pertence a outro veículo
- [ ] Atualizar os dados do veículo
- [ ] Persistir as alterações no banco de dados

**Repositório**

- [ ] Criar ou ajustar `VeiculoRepository` para busca e persistência do veículo
- [ ] Criar método para consultar veículo por identificador
- [ ] Criar método para consultar veículo por placa
- [ ] Criar método para salvar alterações do veículo

**Handler HTTP**

- [ ] Implementar `PUT /veiculos/{veiculoId}`
- [ ] Criar DTO/request de entrada
- [ ] Criar DTO/response de saída com os dados atualizados do veículo
- [ ] Implementar validação do parâmetro `veiculoId`
- [ ] Implementar validação do payload
- [ ] Aplicar autenticação JWT na rota
- [ ] Aplicar autorização para o escopo `veiculos:escrever`
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Validações**

- [ ] Validar que o identificador do veículo foi informado
- [ ] Validar formato do identificador do veículo
- [ ] Validar que o veículo existe
- [ ] Validar que a placa foi informada
- [ ] Validar que a marca foi informada
- [ ] Validar que o modelo foi informado
- [ ] Validar que o ano foi informado
- [ ] Validar formato da placa
- [ ] Retornar `200` quando o veículo for atualizado com sucesso
- [ ] Retornar `400` para identificador ausente, dados obrigatórios ausentes ou placa inválida
- [ ] Retornar `404` quando o veículo não for encontrado
- [ ] Retornar `409` quando a placa pertencer a outro veículo
- [ ] Retornar `401` quando não houver autenticação
- [ ] Retornar `403` quando o usuário não tiver permissão

**Concorrência**

- [ ] Definir e implementar estratégia para conflito de atualização

**Testes unitários**

- [ ] Atualização válida de veículo
- [ ] Identificador ausente
- [ ] Veículo inexistente
- [ ] Placa ausente
- [ ] Marca ausente
- [ ] Modelo ausente
- [ ] Ano ausente
- [ ] Placa inválida
- [ ] Placa já vinculada a outro veículo
- [ ] Preservação dos vínculos com Cliente e Ordens de Serviço

**Testes de integração**

- [ ] Endpoint atualiza veículo válido e retorna `200`
- [ ] Dados atualizados são persistidos no banco
- [ ] Endpoint retorna `400` para identificador ausente, dados obrigatórios ausentes ou placa inválida
- [ ] Endpoint retorna `404` quando o veículo não existe
- [ ] Endpoint retorna `409` quando a placa pertence a outro veículo
- [ ] Endpoint retorna `401` sem autenticação
- [ ] Endpoint retorna `403` sem permissão
- [ ] Vínculos com Cliente e Ordens de Serviço são preservados após a atualização

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI
- [ ] Revisar nomes usando a Linguagem Ubíqua definida no projeto

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite da task
- [ ] Code Review aprovado

---
