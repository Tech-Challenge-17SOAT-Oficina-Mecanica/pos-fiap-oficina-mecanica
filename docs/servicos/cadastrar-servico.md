---
documento: Refinamento de Requisitos — Cadastrar Serviço
dono: João Victor Silva de Oliveira
versao: 0.1
atualizado_em: 2026-08-20
status: rascunho
---

# Refinamento de Requisitos — Cadastrar Serviço

Este documento detalha a tarefa Cadastrar Serviço do contexto de Serviços.

## 2 · Cadastrar Serviço

### 2.1 Refinamento de Produto

**Persona**
Administrador/Gestor da oficina.

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
| RF-SRV-07 | Permitir cadastrar um serviço. |
| RF-SRV-08 | Registrar nome ou descrição do serviço. |
| RF-SRV-09 | Registrar valor do serviço. |
| RF-SRV-10 | Registrar demais informações necessárias para sua utilização no sistema. |
| RF-SRV-11 | Validar os dados informados. |
| RF-SRV-12 | Impedir o cadastro de serviços duplicados conforme a regra de negócio. |
| RF-SRV-13 | Disponibilizar o serviço cadastrado para utilização em novas OS e orçamentos. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-SRV-06 | A operação deve ser persistida de forma consistente. |
| RNF-SRV-07 | Os valores monetários devem ser armazenados corretamente. |
| RNF-SRV-08 | A operação deve exigir autenticação e autorização administrativa. |
| RNF-SRV-09 | Os dados devem ser validados antes da persistência. |
| RNF-SRV-10 | O cadastro não deve alterar serviços já existentes. |

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
| A3 | Serviço duplicado | O sistema informa que já existe um serviço equivalente. |
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

### 2.2 Refinamento Técnico

**Endpoint**

```http
POST /servicos
```

O endpoint cadastra um novo serviço no catálogo da oficina, gerando seu identificador técnico e
código funcional.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `GESTOR`.
- Escopo: `servicos:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Body | `nome` | string | Nome do serviço, obrigatório. |
| Body | `descricao` | string | Descrição do serviço. |
| Body | `valor` | decimal | Valor do serviço, obrigatório, maior ou igual a zero. |
| Body | `tempoEstimadoMinutos` | int | Tempo estimado de execução, obrigatório caso adotado pelo time. |

```json
{
  "nome": "Troca de óleo",
  "descricao": "Troca de óleo e filtro",
  "valor": 150.0,
  "tempoEstimadoMinutos": 60
}
```

O cliente da API não informa `id`, `codigo`, `status` nem `dataCriacao`; esses dados são gerados
pelo sistema.

**Validações**

- `nome` deve ser informado.
- `valor` deve ser maior ou igual a zero.
- `tempoEstimadoMinutos`, caso adotado, deve ser maior que zero.
- Não deve existir outro serviço com o mesmo critério de unicidade definido.
- O usuário deve possuir autorização para cadastrar serviços.

**Processamento**

1. Receber o payload.
2. Identificar o usuário autenticado.
3. Validar autorização.
4. Validar os dados informados.
5. Verificar duplicidade.
6. Gerar o `id`.
7. Gerar o `codigo`.
8. Criar a entidade `Servico`.
9. Definir status inicial `ATIVO`.
10. Registrar data e hora de criação.
11. Persistir o serviço.
12. Retornar o serviço criado.

**Persistência**

- Consulta: `ServicoRepository` para verificar duplicidade.
- Altera: `Servico` com novo registro.
- Persiste: `id`, `codigo`, `nome`, `descricao`, `valor`, `tempo_estimado_minutos`, `status`,
  `data_criacao` e `usuario_criacao`.

**Saída da API**

```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "codigo": "SER-000001",
  "nome": "Troca de óleo",
  "descricao": "Troca de óleo e filtro",
  "valor": 150.0,
  "tempoEstimadoMinutos": 60,
  "status": "ATIVO",
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
| `409` | Serviço duplicado. |
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
- Inicia com status `ATIVO`.
- Rejeita nome vazio.
- Rejeita valor negativo.
- Rejeita duplicidade.
- Não altera serviços já existentes.

*Integração*

- `POST /servicos` válido retorna `201`.
- Serviço é persistido corretamente no banco.
- Requisição sem autenticação retorna `401`.
- Usuário sem permissão retorna `403`.
- Serviço duplicado retorna `409`.
- Dados inválidos retornam `400`.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Criar entidade/agregado `Servico`
- [ ] Definir atributos obrigatórios
- [ ] Criar método de domínio para criação
- [ ] Validar nome obrigatório
- [ ] Validar descrição, caso aplicável
- [ ] Validar valor do serviço
- [ ] Impedir valor negativo
- [ ] Definir tempo estimado, caso aplicável
- [ ] Definir status inicial `ATIVO`
- [ ] Validar duplicidade de serviço

**Caso de uso**

- [ ] Criar caso de uso `CadastrarServico`
- [ ] Gerar `id`
- [ ] Gerar `codigo`
- [ ] Registrar data e hora de criação
- [ ] Registrar usuário de criação

**Repositório**

- [ ] Criar `ServicoRepository`
- [ ] Criar método para verificar duplicidade
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
- [ ] Validar tempo estimado maior que zero, caso adotado
- [ ] Validar critério de unicidade do serviço

**Testes unitários**

- [ ] Cadastro válido
- [ ] Nome obrigatório
- [ ] Preço inválido
- [ ] Duplicidade
- [ ] Status inicial `ATIVO`
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
