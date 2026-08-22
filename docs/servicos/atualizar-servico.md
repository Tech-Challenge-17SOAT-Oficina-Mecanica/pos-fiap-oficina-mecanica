---
documento: Refinamento de Requisitos — Atualizar Serviço
dono: João Victor Silva de Oliveira
versao: 0.2
atualizado_em: 2026-08-20
status: em revisao
---

# Refinamento de Requisitos — Atualizar Serviço

Este documento detalha a tarefa Atualizar Serviço do contexto de Serviços.

## 3 · Atualizar Serviço

### 3.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Atualizar os dados de um serviço existente no catálogo da oficina.

**Problema**
Valores, descrições e demais informações dos serviços podem precisar ser atualizados para
refletir as condições atuais da oficina. Sem atualização controlada, o catálogo fica defasado e
pode gerar orçamentos incorretos, mantendo ainda o risco de alterar indevidamente históricos já
registrados em Ordens de Serviço e orçamentos.

**Pré-condições**

- O usuário deve possuir autorização administrativa.
- O serviço deve existir.
- O serviço não pode estar sendo atualizado simultaneamente de forma conflitante.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-SRV-14 | Permitir localizar um serviço existente. |
| RF-SRV-15 | Permitir alterar seus dados cadastrais. |
| RF-SRV-16 | Validar os dados atualizados. |
| RF-SRV-17 | Impedir alterações que gerem duplicidade, quando aplicável. |
| RF-SRV-18 | Persistir as alterações realizadas. |
| RF-SRV-19 | Manter o serviço disponível para novas utilizações após a atualização. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SRV-11 | A atualização deve ocorrer de forma consistente. |
| RNF-SRV-12 | Os valores monetários devem ser validados e armazenados corretamente. |
| RNF-SRV-13 | Somente usuário autorizado poderá atualizar o serviço. |
| RNF-SRV-14 | A atualização não deve modificar retroativamente valores já registrados em Ordens de Serviço ou orçamentos, caso esses valores sejam armazenados como histórico da operação. |
| RNF-SRV-15 | A operação deve evitar atualização parcial dos dados. |

**Fluxo Principal**

1. O administrador consulta o catálogo de serviços.
2. O administrador seleciona o serviço que deseja alterar.
3. O sistema verifica se o serviço existe.
4. O sistema apresenta os dados atuais.
5. O administrador altera os dados necessários.
6. O sistema valida os novos dados.
7. O sistema verifica possíveis conflitos ou duplicidades.
8. O sistema persiste as alterações.
9. O sistema confirma a atualização.
10. O sistema apresenta os dados atualizados.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Serviço não encontrado | O sistema informa que o serviço não existe. |
| A2 | Dados inválidos | O sistema impede a atualização e apresenta os erros. |
| A3 | Duplicidade | O sistema impede a atualização caso ela gere um serviço duplicado. |
| A4 | Conflito de atualização | O sistema solicita nova consulta dos dados antes de realizar a alteração. |
| A5 | Usuário sem autorização | O sistema impede a operação. |

**Saída**

- Serviço existente atualizado com os novos dados.

**Pós-condições**

- Os dados atuais do serviço ficam atualizados.
- O serviço continua disponível conforme seu status.
- Registros históricos já vinculados a OS ou orçamentos permanecem preservados conforme as regras de histórico do sistema.

---

### 3.2 Refinamento Técnico

**Endpoint**

```http
PATCH /servicos/{servicoId}
```

O endpoint atualiza os dados cadastrais de um serviço existente sem alterar os valores históricos
utilizados em Ordens de Serviço já registradas.

> **Decisão de projeto.** `PATCH` é **atualização parcial de verdade**: campo ausente no corpo não
> é alterado. Enviar o recurso inteiro continua funcionando, mas não é exigido.

> **Decisão de projeto.** `id`, `codigo` e `dataCriacao` são **imutáveis**. Se vierem no corpo, a
> requisição é rejeitada com `400` — em vez de serem ignorados em silêncio, o que esconderia erro
> do cliente da API.

> **Decisão de projeto.** A atualização usa controle otimista com `If-Match` e `version`, como em
> Cliente, Veículo, Peças e Insumos (D-24).

> **Decisão de projeto.** O path param é `{servicoId}`, e não `{id}`, alinhado aos demais
> contextos.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `MECANICO`.
- Escopo: `servicos:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `servicoId` | uuid | Identificador do serviço. |
| Header | `If-Match` | string | `version` atual do registro, obrigatório. |
| Body | `nome` | string | Opcional. Nome atualizado do serviço. |
| Body | `descricao` | string | Opcional. Descrição atualizada do serviço. |
| Body | `valor` | decimal | Opcional. Não pode ser negativo. |
| Body | `tempoEstimadoMinutos` | int | Opcional. Mínimo de 1 minuto. |
| Body | `ativo` | — | **Não aceito.** A desativação e a reativação têm rotas próprias. |

```json
{
  "nome": "Troca de óleo e filtro",
  "descricao": "Troca de óleo, filtro e inspeção",
  "valor": 180.0,
  "tempoEstimadoMinutos": 75
}
```

Como a atualização é parcial, o exemplo acima altera apenas os quatro campos enviados. Campos
imutáveis (`id`, `codigo`, `dataCriacao`) e a situação (`ativo`) não são aceitos no corpo.

**Validações**

- `servicoId` deve possuir formato válido de UUID.
- O serviço deve existir.
- `nome`, quando informado, deve ser válido e não vazio.
- `valor`, quando informado, não pode ser negativo.
- `tempoEstimadoMinutos`, quando informado, deve ser maior ou igual a 1.
- `id`, `codigo`, `dataCriacao` e `ativo` no corpo retornam `400`.
- O nome normalizado resultante não pode pertencer a outro serviço **ativo**.
- `If-Match` deve ser informado e bater com a `version` atual do registro.
- O usuário deve possuir autorização para atualizar serviços.

**Processamento**

1. Receber o `servicoId`.
2. Identificar o usuário autenticado.
3. Validar autorização.
4. Buscar o serviço, com lock otimista.
5. Validar existência.
6. Comparar `If-Match` com a `version` atual — divergência retorna `412`, ausência retorna `428`.
7. Validar o payload e rejeitar campos imutáveis.
8. Normalizar o nome resultante e validar duplicidade entre serviços ativos.
9. Aplicar apenas os campos presentes no corpo.
10. Registrar data, hora e usuário responsável pela atualização.
11. Persistir e incrementar `version`.
12. Retornar o serviço atualizado.

**Persistência**

- Consulta: `ServicoRepository` para buscar o serviço e validar duplicidade.
- Altera: `Servico`.
- Persiste: apenas os campos enviados, mais `nome_normalizado`, `data_atualizacao`,
  `usuario_atualizacao` e `version` incrementada.
- Não altera: `id`, `codigo`, `data_criacao`, `ativo` e o histórico das OS ou orçamentos já
  registrados.

**Saída da API**

```json
{
  "id": "4b8e2c17-95a3-4f60-b7d1-6e0c58a3f942",
  "codigo": "SER-000001",
  "nome": "Troca de óleo e filtro",
  "descricao": "Troca de óleo, filtro e inspeção",
  "valor": 180.0,
  "tempoEstimadoMinutos": 75,
  "ativo": true,
  "version": 4,
  "dataAtualizacao": "2026-08-19T20:20:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Serviço atualizado. |
| `400` | Dados inválidos, ou campo imutável enviado no corpo. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `servicos:escrever`. |
| `404` | Serviço inexistente. |
| `409` | Nome já usado por outro serviço ativo. |
| `412` | `If-Match` divergente — o serviço foi alterado por outro usuário. |
| `428` | `If-Match` não informado. |
| `500` | Falha inesperada. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de serviços.
- `ServicoRepository`.
- Relógio da aplicação para registrar data e hora de atualização.

**Testes**

*Unitários*

- Atualiza serviço existente.
- Mantém o mesmo `id`.
- Mantém o mesmo `codigo`.
- Rejeita valor negativo.
- Rejeita serviço inexistente.
- Rejeita nome já usado por outro serviço ativo.
- Rejeita `codigo` enviado no corpo.
- Rejeita `ativo` enviado no corpo.
- Rejeita `tempoEstimadoMinutos` menor que 1.
- Atualiza apenas os campos enviados, preservando os demais.
- Preserva valores históricos das OS.

*Integração*

- `PATCH /servicos/{servicoId}` válido retorna `200`.
- `PATCH` com um campo só altera apenas esse campo.
- `PATCH` com `If-Match` antigo retorna `412`.
- `PATCH` sem `If-Match` retorna `428`.
- `PATCH` com `codigo` no corpo retorna `400`.
- Serviço atualizado é persistido corretamente no banco.
- Serviço inexistente retorna `404`.
- Dados inválidos retornam `400`.
- Nome já usado por outro serviço ativo retorna `409`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Criar método de domínio `Atualizar` em `Servico`
- [ ] Definir quais campos podem ser alterados: `nome`, `descricao`, `valor` e `tempoEstimadoMinutos`
- [ ] Declarar `id`, `codigo`, `dataCriacao` e `ativo` como imutáveis neste fluxo
- [ ] Adicionar o campo `version` ao modelo `Servico` para controle otimista
- [ ] Validar campos obrigatórios
- [ ] Validar valor
- [ ] Validar duplicidade de nome normalizado entre serviços ativos
- [ ] Permitir alteração dos dados previstos
- [ ] Preservar histórico das OS existentes
- [ ] Não alterar retroativamente valores utilizados em OS

**Caso de uso**

- [ ] Criar caso de uso `AtualizarServico`
- [ ] Buscar serviço existente
- [ ] Validar existência
- [ ] Aplicar apenas os campos presentes no corpo
- [ ] Comparar `If-Match` com a `version` atual antes de aplicar as alterações
- [ ] Registrar data e hora de atualização
- [ ] Registrar usuário responsável, caso aplicável

**Repositório**

- [ ] Consultar serviço por identificador no `ServicoRepository`
- [ ] Consultar duplicidade de nome ou critério equivalente
- [ ] Persistir alterações do serviço e incrementar `version`

**Handler HTTP**

- [ ] Implementar `PATCH /servicos/{servicoId}`
- [ ] Criar DTO/request
- [ ] Criar DTO/response
- [ ] Aplicar autenticação JWT
- [ ] Aplicar autorização para o escopo `servicos:escrever`
- [ ] Retornar `404` para serviço inexistente
- [ ] Retornar `409` para nome já usado por outro serviço ativo
- [ ] Retornar `412` quando o `If-Match` divergir da `version` atual
- [ ] Retornar `428` quando o `If-Match` não for informado

**Validações**

- [ ] Validar formato do `servicoId`
- [ ] Validar nome quando informado
- [ ] Validar valor não negativo quando informado
- [ ] Validar tempo estimado maior ou igual a 1 quando informado
- [ ] Rejeitar com `400` os campos imutáveis enviados no corpo
- [ ] Validar duplicidade de nome normalizado entre serviços ativos

**Concorrência**

- [ ] Implementar controle otimista com `If-Match` comparado ao campo `version`
- [ ] Incrementar `version` a cada atualização persistida

**Testes unitários**

- [ ] Atualização de serviço existente
- [ ] Serviço inexistente
- [ ] Nome já usado por outro serviço ativo
- [ ] Valor negativo
- [ ] Campo imutável enviado no corpo
- [ ] Atualização parcial preservando os campos não enviados
- [ ] `If-Match` divergente
- [ ] Preservação do histórico
- [ ] Manutenção de `id` e `codigo`

**Testes de integração**

- [ ] Endpoint atualiza serviço válido e retorna `200`
- [ ] Endpoint retorna `404` para serviço inexistente
- [ ] Endpoint retorna `400` para dados inválidos
- [ ] Endpoint retorna `409` para nome já usado por outro serviço ativo
- [ ] Endpoint retorna `412` com `If-Match` antigo
- [ ] Endpoint retorna `428` sem `If-Match`
- [ ] Endpoint bloqueia usuário sem autorização

**Documentação**

- [ ] Documentar no Swagger/OpenAPI
- [ ] Revisar linguagem ubíqua

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite
- [ ] Code Review aprovado

---
