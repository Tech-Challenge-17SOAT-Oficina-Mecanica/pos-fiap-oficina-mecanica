---
documento: Refinamento de Requisitos — Atualizar Serviço
dono: João Victor Silva de Oliveira
versao: 0.1
atualizado_em: 2026-08-20
status: rascunho
---

# Refinamento de Requisitos — Atualizar Serviço

Este documento detalha a tarefa Atualizar Serviço do contexto de Serviços.

## 3 · Atualizar Serviço

### 3.1 Refinamento de Produto

**Persona**
Administrador/Gestor da oficina.

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
PATCH /servicos/{id}
```

O endpoint atualiza os dados cadastrais de um serviço existente sem alterar os valores históricos
utilizados em Ordens de Serviço já registradas.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `GESTOR`.
- Escopo: `servicos:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `id` | UUID | Identificador do serviço. |
| Body | `nome` | string | Nome atualizado do serviço. |
| Body | `descricao` | string | Descrição atualizada do serviço. |
| Body | `valor` | decimal | Valor atualizado do serviço; não pode ser negativo. |
| Body | `tempoEstimadoMinutos` | int | Tempo estimado atualizado, caso adotado pelo time. |

```json
{
  "nome": "Troca de óleo e filtro",
  "descricao": "Troca de óleo, filtro e inspeção",
  "valor": 180.0,
  "tempoEstimadoMinutos": 75
}
```

O `codigo` não deve ser alterado por esse fluxo caso seja definido como identificador funcional
imutável.

**Validações**

- `id` deve possuir formato válido de UUID.
- O serviço deve existir.
- `nome` deve ser válido.
- `valor` não pode ser negativo.
- `tempoEstimadoMinutos`, caso adotado, deve ser válido.
- A atualização não pode gerar conflito com outro serviço.
- O usuário deve possuir autorização para atualizar serviços.

**Processamento**

1. Receber o `id`.
2. Identificar o usuário autenticado.
3. Validar autorização.
4. Buscar o serviço.
5. Validar existência.
6. Validar payload.
7. Validar duplicidade e conflitos.
8. Executar `servico.Atualizar(...)`.
9. Registrar data e hora da atualização.
10. Registrar usuário responsável pela atualização.
11. Persistir.
12. Retornar o serviço atualizado.

**Persistência**

- Consulta: `ServicoRepository` para buscar o serviço e validar duplicidade.
- Altera: `Servico`.
- Persiste: `nome`, `descricao`, `valor`, `tempo_estimado_minutos`, `data_atualizacao` e
  `usuario_atualizacao`.
- Não altera: `id`, `codigo` e histórico das OS ou orçamentos já registrados.

**Saída da API**

```json
{
  "id": "123",
  "codigo": "SER-000001",
  "nome": "Troca de óleo e filtro",
  "descricao": "Troca de óleo, filtro e inspeção",
  "valor": 180.0,
  "tempoEstimadoMinutos": 75,
  "status": "ATIVO",
  "dataAtualizacao": "2026-08-19T20:20:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200` | Serviço atualizado. |
| `400` | Dados inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `servicos:escrever`. |
| `404` | Serviço inexistente. |
| `409` | Conflito com outro serviço. |
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
- Rejeita duplicidade.
- Preserva valores históricos das OS.

*Integração*

- `PATCH /servicos/{id}` válido retorna `200`.
- Serviço atualizado é persistido corretamente no banco.
- Serviço inexistente retorna `404`.
- Dados inválidos retornam `400`.
- Conflito com outro serviço retorna `409`.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.

---

### 3.3 Checklist de Implementação

**Domínio**

- [ ] Criar método de domínio `Atualizar` em `Servico`
- [ ] Definir quais campos podem ser alterados
- [ ] Validar campos obrigatórios
- [ ] Validar valor
- [ ] Validar duplicidade de nome
- [ ] Permitir alteração dos dados previstos
- [ ] Preservar histórico das OS existentes
- [ ] Não alterar retroativamente valores utilizados em OS

**Caso de uso**

- [ ] Criar caso de uso `AtualizarServico`
- [ ] Buscar serviço existente
- [ ] Validar existência
- [ ] Registrar data e hora de atualização
- [ ] Registrar usuário responsável, caso aplicável

**Repositório**

- [ ] Consultar serviço por identificador no `ServicoRepository`
- [ ] Consultar duplicidade de nome ou critério equivalente
- [ ] Persistir alterações do serviço

**Handler HTTP**

- [ ] Implementar `PATCH /servicos/{id}`
- [ ] Criar DTO/request
- [ ] Criar DTO/response
- [ ] Aplicar autenticação JWT
- [ ] Aplicar autorização para o escopo `servicos:escrever`
- [ ] Retornar `404` para serviço inexistente
- [ ] Retornar `409` para conflito

**Validações**

- [ ] Validar formato do `id`
- [ ] Validar nome
- [ ] Validar valor não negativo
- [ ] Validar tempo estimado, caso adotado
- [ ] Validar duplicidade de serviço

**Concorrência**

- [ ] Definir e implementar estratégia para conflito de atualização

**Testes unitários**

- [ ] Atualização de serviço existente
- [ ] Serviço inexistente
- [ ] Nome duplicado
- [ ] Valor negativo
- [ ] Preservação do histórico
- [ ] Manutenção de `id` e `codigo`

**Testes de integração**

- [ ] Endpoint atualiza serviço válido e retorna `200`
- [ ] Endpoint retorna `404` para serviço inexistente
- [ ] Endpoint retorna `400` para dados inválidos
- [ ] Endpoint retorna `409` para conflito
- [ ] Endpoint bloqueia usuário sem autorização

**Documentação**

- [ ] Documentar no Swagger/OpenAPI
- [ ] Revisar linguagem ubíqua

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite
- [ ] Code Review aprovado

---
