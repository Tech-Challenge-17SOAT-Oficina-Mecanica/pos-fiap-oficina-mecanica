---
documento: Refinamento de Requisitos — Cadastrar Serviço
dono: João Victor Silva de Oliveira
versao: 0.2
atualizado_em: 2026-08-20
status: em revisao
---

# Refinamento de Requisitos — Cadastrar Serviço

Este documento detalha a tarefa Cadastrar Serviço do contexto de Serviços.

## 1 · Cadastrar Serviço

### 1.1 Refinamento de Produto

**Persona**
Mecânico.

**Objetivo**
Cadastrar um novo serviço no catálogo da oficina para que ele possa ser utilizado em
diagnósticos, Ordens de Serviço e orçamentos.

**Problema**
A oficina precisa manter um catálogo estruturado de serviços, com informações padronizadas,
evitando registros manuais e inconsistentes durante a criação das OS e dos orçamentos. Sem esse
catálogo, os serviços podem ser registrados com nomes diferentes, valores incorretos e sem
rastreabilidade.

**Pré-condições**

- O usuário deve possuir autorização administrativa.
- Os dados obrigatórios do serviço devem estar disponíveis.
- O serviço não pode possuir identificação que gere duplicidade no catálogo, conforme as regras definidas.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-SRV-01 | Permitir cadastrar um serviço. |
| RF-SRV-02 | Registrar nome ou descrição do serviço. |
| RF-SRV-03 | Registrar valor do serviço. |
| RF-SRV-04 | Registrar demais informações necessárias para sua utilização no sistema. |
| RF-SRV-05 | Validar os dados informados. |
| RF-SRV-06 | Impedir o cadastro de serviço duplicado: o nome normalizado deve ser único entre os serviços ativos. |
| RF-SRV-07 | Disponibilizar o serviço cadastrado para utilização em novas OS e orçamentos. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SRV-01 | A operação deve ser persistida de forma consistente. |
| RNF-SRV-02 | Os valores monetários devem ser armazenados corretamente. |
| RNF-SRV-03 | A operação deve exigir autenticação e autorização administrativa. |
| RNF-SRV-04 | Os dados devem ser validados antes da persistência. |
| RNF-SRV-05 | O cadastro não deve alterar serviços já existentes. |
| RNF-SRV-21 | O `codigo` deve ser gerado pelo sistema, em sequência global, sem reset. |

**Fluxo Principal**

1. O administrador acessa o gerenciamento de serviços.
2. O administrador solicita o cadastro de um novo serviço.
3. O sistema apresenta o formulário de cadastro.
4. O administrador informa os dados do serviço.
5. O sistema valida os campos obrigatórios.
6. O sistema valida o valor informado.
7. O sistema verifica a existência de serviço duplicado.
8. O sistema registra o novo serviço.
9. O sistema confirma o cadastro.
10. O serviço passa a estar disponível no catálogo.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Dados obrigatórios ausentes | O sistema informa os campos que precisam ser preenchidos. |
| A2 | Valor inválido | O sistema impede o cadastro. |
| A3 | Serviço duplicado | O sistema informa que já existe um serviço ativo com o mesmo nome normalizado. |
| A4 | Dados inválidos | O sistema rejeita o cadastro e informa os erros encontrados. |
| A5 | Usuário sem autorização | O sistema impede a operação. |

**Saída**

- Novo serviço cadastrado e disponível no catálogo da oficina.

**Pós-condições**

- O serviço passa a existir no catálogo.
- Seus dados ficam persistidos no banco de dados.
- O serviço pode ser associado a novas Ordens de Serviço e orçamentos.
- O serviço pode posteriormente ser consultado, atualizado ou desativado.

---

### 1.2 Refinamento Técnico

**Endpoint**

```http
POST /servicos
```

O endpoint cadastra um novo serviço no catálogo da oficina, gerando seu identificador técnico e
código funcional.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `MECANICO`.
- Escopo: `servicos:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `nome` | string | Nome do serviço, obrigatório. |
| Body | `descricao` | string | Descrição do serviço. |
| Body | `valor` | decimal | Valor do serviço, obrigatório, maior ou igual a zero. |
| Body | `tempoEstimadoMinutos` | int | Tempo estimado de execução, obrigatório, mínimo de 1 minuto. |

```json
{
  "nome": "Troca de óleo",
  "descricao": "Troca de óleo e filtro",
  "valor": 150.0,
  "tempoEstimadoMinutos": 60
}
```

O cliente da API não informa `id`, `codigo`, `ativo` nem `dataCriacao`; esses dados são gerados
pelo sistema.

> **Decisão de projeto.** A situação do serviço é o booleano `ativo`, com `dataDesativacao` e
> `usuarioDesativacao`, como nos demais contextos. O enum `status` com `ATIVO`/`INATIVO` foi
> descartado para o projeto inteiro ter uma representação só (D-19).

> **Decisão de projeto.** A unicidade é por **nome normalizado** — sem acento, sem espaço duplo,
> em minúsculas — e vale **apenas entre serviços ativos**, por índice parcial. Assim um serviço
> desativado não bloqueia o cadastro de outro com o mesmo nome.

> **Decisão de projeto.** O `codigo` segue o formato `SER-000001`, gerado pelo sistema em
> **sequência global, sem reset por ano**, com seis dígitos — a mesma regra proposta para o código
> de peças e insumos, para os dois contextos seguirem o mesmo padrão.

> **Decisão de projeto.** `tempoEstimadoMinutos` é **obrigatório**, com mínimo de 1 minuto. Ele
> alimenta a estimativa de entrega do orçamento, que fica sem base quando o campo é opcional.

**Validações**

- `nome` deve ser informado.
- `valor` deve ser maior ou igual a zero.
- `tempoEstimadoMinutos` deve ser informado e ser maior ou igual a 1.
- Não deve existir outro serviço **ativo** com o mesmo nome normalizado.
- O usuário deve possuir autorização para cadastrar serviços.

**Processamento**

1. Receber o payload.
2. Identificar o usuário autenticado.
3. Validar autorização.
4. Validar os dados informados.
5. Normalizar o nome e verificar duplicidade entre serviços ativos.
6. Gerar o `id` (UUID).
7. Gerar o `codigo` a partir da sequência global.
8. Criar a entidade `Servico`.
9. Definir `ativo = true`.
10. Registrar data e hora de criação.
11. Persistir o serviço.
12. Retornar o serviço criado.

**Persistência**

- Consulta: `ServicoRepository` para verificar duplicidade.
- Altera: `Servico` com novo registro.
- Persiste: `id`, `codigo`, `nome`, `nome_normalizado`, `descricao`, `valor`,
  `tempo_estimado_minutos`, `ativo`, `version`, `data_criacao` e `usuario_criacao`.
- Índice parcial `UNIQUE (nome_normalizado) WHERE ativo = true`.

**Saída da API**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "codigo": "SER-000001",
  "nome": "Troca de óleo",
  "descricao": "Troca de óleo e filtro",
  "valor": 150.0,
  "tempoEstimadoMinutos": 60,
  "ativo": true,
  "version": 1,
  "dataCriacao": "2026-08-19T20:00:00-03:00"
}
```

A resposta deve representar o serviço efetivamente cadastrado, incluindo os dados gerados pelo
sistema.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Serviço cadastrado. |
| `400` | Dados inválidos. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `servicos:escrever`. |
| `409` | Já existe serviço ativo com o mesmo nome normalizado. |
| `500` | Falha inesperada. |

**Dependências**

- Módulo de autenticação JWT.
- Módulo de serviços.
- `ServicoRepository`.
- Gerador de UUID.
- Gerador de código funcional de serviço.
- Relógio da aplicação para registrar data e hora de criação.

**Testes**

*Unitários*

- Cadastra serviço válido.
- Gera `id`.
- Gera `codigo`.
- Inicia com `ativo = true`.
- Normaliza o nome antes de verificar duplicidade.
- Rejeita `tempoEstimadoMinutos` ausente.
- Rejeita `tempoEstimadoMinutos` menor que 1.
- Aceita cadastro com nome igual ao de um serviço inativo.
- Rejeita nome vazio.
- Rejeita valor negativo.
- Rejeita nome já usado por serviço ativo.
- Não altera serviços já existentes.

*Integração*

- `POST /servicos` válido retorna `201`.
- Serviço é persistido corretamente no banco.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.
- Serviço duplicado retorna `409`.
- Dados inválidos retornam `400`.
- Nome igual ao de serviço inativo retorna `201`.

---

### 1.3 Checklist de Implementação

**Domínio**

- [ ] Criar entidade/agregado `Servico`
- [ ] Definir atributos obrigatórios
- [ ] Criar método de domínio para criação
- [ ] Validar nome obrigatório
- [ ] Validar descrição, caso aplicável
- [ ] Validar valor do serviço
- [ ] Impedir valor negativo
- [ ] Validar `tempoEstimadoMinutos` obrigatório e maior ou igual a 1
- [ ] Definir `ativo = true` no cadastro
- [ ] Implementar a normalização do nome (sem acento, sem espaço duplo, minúsculo)
- [ ] Validar duplicidade de nome normalizado entre serviços ativos

**Caso de uso**

- [ ] Criar caso de uso `CadastrarServico`
- [ ] Gerar `id`
- [ ] Gerar `codigo` no formato `SER-000001`, a partir de sequência global sem reset
- [ ] Registrar data e hora de criação
- [ ] Registrar usuário de criação

**Repositório**

- [ ] Criar `ServicoRepository`
- [ ] Criar método para verificar duplicidade por nome normalizado entre ativos
- [ ] Criar o índice parcial `UNIQUE (nome_normalizado) WHERE ativo = true` na migration
- [ ] Criar método para salvar novo serviço

**Handler HTTP**

- [ ] Implementar `POST /servicos`
- [ ] Criar DTO/request
- [ ] Criar DTO/response
- [ ] Aplicar autenticação JWT
- [ ] Aplicar autorização para o escopo `servicos:escrever`
- [ ] Retornar `201` para serviço cadastrado
- [ ] Retornar `409` para serviço duplicado
- [ ] Retornar `400` para dados inválidos

**Validações**

- [ ] Validar nome obrigatório
- [ ] Validar valor maior ou igual a zero
- [ ] Validar `tempoEstimadoMinutos` obrigatório, maior ou igual a 1
- [ ] Validar unicidade do nome normalizado entre serviços ativos

**Testes unitários**

- [ ] Cadastro válido
- [ ] Nome obrigatório
- [ ] Preço inválido
- [ ] Duplicidade
- [ ] `ativo = true` no cadastro
- [ ] Tempo estimado ausente ou menor que 1
- [ ] Nome igual ao de serviço inativo aceito
- [ ] Geração de `id` e `codigo`

**Testes de integração**

- [ ] Endpoint cadastra serviço válido e retorna `201`
- [ ] Serviço é persistido corretamente
- [ ] Endpoint retorna `400` para dados inválidos
- [ ] Endpoint retorna `409` para duplicidade
- [ ] Endpoint bloqueia usuário sem autorização

**Documentação**

- [ ] Documentar no Swagger/OpenAPI
- [ ] Revisar linguagem ubíqua

**Review**

- [ ] Executar testes automatizados
- [ ] Validar critérios de aceite
- [ ] Code Review aprovado

---
