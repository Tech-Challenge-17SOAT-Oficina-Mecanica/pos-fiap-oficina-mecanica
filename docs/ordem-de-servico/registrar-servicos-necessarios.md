---
documento: Refinamento de Requisitos — Registrar Serviços Necessários
dono: Helena Miranda
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Registrar Serviços Necessários

Este documento detalha a tarefa Registrar Serviços Necessários do contexto de Ordem de Serviço.

## 2 · Registrar Serviços Necessários

### 2.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar na Ordem de Serviço os serviços identificados como necessários durante o diagnóstico
do veículo, para que possam compor o orçamento e sejam executados somente após a autorização do
cliente.

**Problema**

Após avaliar o veículo, a oficina precisa registrar de forma estruturada os serviços necessários
para corrigir os problemas encontrados. Sem esse registro, o orçamento pode ficar incompleto, o
histórico do diagnóstico perde rastreabilidade e existe o risco de executar algo que não foi
autorizado pelo cliente.

**Pré-condições**

- A Ordem de Serviço deve existir.
- A Ordem de Serviço deve estar na situação `EM_DIAGNOSTICO`.
- O diagnóstico deve ter sido iniciado.
- Deve existir pelo menos um problema identificado no veículo.
- Os serviços informados devem existir e estar ativos no catálogo de serviços.
- O mecânico deve estar autorizado a alterar os serviços necessários da Ordem de Serviço.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-06 | Permitir ao mecânico consultar os serviços disponíveis no catálogo. |
| RF-OS-07 | Permitir selecionar e registrar um ou mais serviços necessários na Ordem de Serviço. |
| RF-OS-08 | Permitir informar uma observação para cada serviço, quando aplicável. |
| RF-OS-09 | Validar se cada serviço existe e está ativo no catálogo. |
| RF-OS-10 | Impedir que o mesmo serviço seja associado mais de uma vez à Ordem de Serviço. |
| RF-OS-11 | Manter os serviços registrados disponíveis para a composição do orçamento. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-05 | Persistir todos os serviços informados de forma consistente. |
| RNF-OS-06 | Restringir a alteração dos serviços da Ordem de Serviço a usuário autenticado e autorizado. |
| RNF-OS-07 | Preservar a rastreabilidade entre o diagnóstico, a Ordem de Serviço e o orçamento. |
| RNF-OS-08 | Não alterar dados do catálogo durante a associação dos serviços à Ordem de Serviço. |
| RNF-OS-09 | Impedir que alterações concorrentes sobrescrevam os serviços registrados por outro usuário. |

**Fluxo Principal**

1. O mecânico acessa uma Ordem de Serviço em diagnóstico.
2. O sistema apresenta os problemas identificados e permite consultar o catálogo de serviços
   disponíveis.
3. O mecânico seleciona um ou mais serviços necessários.
4. O mecânico informa uma observação para cada serviço, quando necessário.
5. O sistema valida se os serviços existem e estão ativos.
6. O sistema verifica se algum serviço já está associado à Ordem de Serviço.
7. O sistema associa todos os serviços válidos à Ordem de Serviço.
8. O sistema registra os serviços como necessários para o diagnóstico.
9. O sistema confirma o registro.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | Informa que a Ordem de Serviço não existe e não registra nenhum serviço. |
| A2 | Ordem de Serviço fora da etapa de diagnóstico | Impede a inclusão e mantém a Ordem de Serviço inalterada. |
| A3 | Serviço não encontrado | Informa que o serviço não existe e não registra nenhum item da solicitação. |
| A4 | Serviço inativo | Impede a associação e não registra nenhum item da solicitação. |
| A5 | Serviço já associado à Ordem de Serviço | Impede a duplicidade e não registra nenhum item da solicitação. |
| A6 | Lista de serviços vazia | Solicita pelo menos um serviço e não altera a Ordem de Serviço. |
| A7 | Usuário não autenticado | Impede a operação e solicita autenticação. |
| A8 | Usuário sem autorização | Impede a operação e mantém a Ordem de Serviço inalterada. |
| A9 | A Ordem de Serviço foi alterada por outro usuário | Impede a sobrescrita da versão mais recente e solicita nova consulta. |

**Saída**

- Serviços necessários registrados e associados à Ordem de Serviço; ou
- Indicação do motivo pelo qual os serviços não puderam ser registrados.

**Pós-condições**

- A Ordem de Serviço permanece na situação `EM_DIAGNOSTICO`.
- Os serviços passam a integrar a relação de serviços necessários da Ordem de Serviço.
- Os serviços registrados ficam disponíveis para a composição do orçamento.
- O catálogo de serviços permanece inalterado.
- O mecânico pode continuar registrando serviços ou seguir para Registrar Peças e Insumos
  Necessários.

---

### 2.2 Refinamento Técnico

**Endpoint**

```http
POST /api/v1/ordens-servico/{id}/servicos
```

O endpoint cria uma ou mais associações entre serviços do catálogo e uma Ordem de Serviço em
diagnóstico.

> **Decisão de projeto.** Foi adotado `POST` com resposta `201` porque cada item incluído é um
> novo recurso subordinado à Ordem de Serviço. A alternativa seria tratar a operação como uma
> atualização genérica da Ordem de Serviço e retornar `200`, mas isso esconderia a criação das
> associações e tornaria menos claro o endereço do recurso criado.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfil esperado: `MECANICO`.
- Escopo: `os:escrever`.
- O identificador do usuário é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `id` | string | Identificador obrigatório da Ordem de Serviço. |
| Header | `If-Match` | inteiro | Versão atual da Ordem de Serviço, obrigatória para controle de concorrência. |
| Body | `servicos` | array | Lista obrigatória e não vazia de serviços necessários. |
| Body | `servicos[].servicoId` | string | Identificador obrigatório e não repetido de um serviço do catálogo. |
| Body | `servicos[].observacao` | string | Informação complementar opcional sobre a necessidade do serviço. |

```json
{
  "servicos": [
    {
      "servicoId": "SRV-001",
      "observacao": "Troca necessária devido ao desgaste identificado"
    },
    {
      "servicoId": "SRV-002",
      "observacao": "Alinhamento recomendado após a substituição da peça"
    }
  ]
}
```

O preço não é aceito na entrada. Os dados comerciais usados posteriormente pelo orçamento são
obtidos do catálogo de serviços.

**Validações**

*Técnicas*

- `id` e `If-Match` devem possuir formatos válidos.
- `servicos` deve conter pelo menos um item.
- Cada `servicoId` é obrigatório e não pode se repetir no mesmo corpo.
- `observacao` é opcional.

*Negócio*

- A Ordem de Serviço deve existir e estar na situação `EM_DIAGNOSTICO`.
- O diagnóstico deve ter uma data de início registrada.
- Deve existir pelo menos um problema identificado no diagnóstico.
- Cada serviço deve existir e estar ativo no catálogo.
- Um serviço ainda não pode estar associado à mesma Ordem de Serviço.
- O usuário deve possuir autorização para executar a operação.

**Processamento**

1. Validar a entrada e identificar o usuário autenticado.
2. Buscar a Ordem de Serviço pelo identificador.
3. Verificar a autorização do usuário e a versão informada em `If-Match`.
4. Validar se a Ordem de Serviço está em diagnóstico e possui problema identificado.
5. Buscar no catálogo todos os serviços informados.
6. Validar a existência e a situação ativa de cada serviço.
7. Verificar duplicidades no corpo e nas associações existentes da Ordem de Serviço.
8. Solicitar ao agregado a inclusão dos serviços necessários.
9. Persistir todas as associações e a nova versão da Ordem de Serviço na mesma transação.
10. Retornar os serviços registrados.

A regra de associação pertence ao agregado de Ordem de Serviço. O catálogo é somente consultado
e não é alterado por este caso de uso.

**Persistência**

- Consulta: `ordem_servico`, problemas do diagnóstico e catálogo de `servico`.
- Altera: associações de serviço da Ordem de Serviço e `ordem_servico.version`.
- Não altera: catálogo de serviços, situação da Ordem de Serviço, cliente, veículo, estoque ou
  orçamento.
- A inclusão do lote é atômica: se um item for inválido, nenhum serviço da solicitação é
  associado.

**Saída da API**

```json
{
  "ordemServicoId": "OS-2026-0123",
  "servicos": [
    {
      "servicoId": "SRV-001",
      "nome": "Troca de pastilhas",
      "observacao": "Troca necessária devido ao desgaste identificado"
    },
    {
      "servicoId": "SRV-002",
      "nome": "Alinhamento",
      "observacao": "Alinhamento recomendado após a substituição da peça"
    }
  ],
  "version": 3
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Serviços associados à Ordem de Serviço. |
| `400` | Corpo inválido, lista vazia ou identificador repetido na solicitação. |
| `401` | Token ausente ou expirado. |
| `403` | Usuário sem o escopo `os:escrever`. |
| `404` | Ordem de Serviço ou serviço do catálogo não encontrado. |
| `409` | Ordem de Serviço fora de diagnóstico, diagnóstico não iniciado, ausência de problema identificado, serviço inativo ou serviço já associado. |
| `412` | A versão informada em `If-Match` diverge da versão atual. |

**Dependências**

- `OrdemDeServicoRepository` para consultar e persistir a Ordem de Serviço.
- `ServicoRepository` para consultar o catálogo de serviços.
- Middleware de autenticação JWT.
- Middleware ou política de autorização para o escopo `os:escrever`.
- Gerenciador de transações.
- Caso de uso de consulta do catálogo pertencente ao contexto de Serviços.

**Testes**

*Unitários*

- Registra um serviço válido em uma Ordem de Serviço em diagnóstico.
- Registra mais de um serviço na mesma operação.
- Rejeita serviço duplicado no corpo ou já associado à Ordem de Serviço.
- Rejeita associação quando a Ordem de Serviço não está em diagnóstico.
- Rejeita associação sem diagnóstico iniciado ou sem problema identificado.
- Mantém a situação `EM_DIAGNOSTICO` após a associação.

*Integração*

- Solicitação válida retorna `201` com os serviços registrados e a nova versão.
- Ordem de Serviço ou serviço inexistente retorna `404`.
- Serviço inativo retorna `409`.
- Serviço duplicado retorna `409`.
- Ordem de Serviço em situação incompatível retorna `409`.
- Lista vazia retorna `400`.
- Token ausente ou expirado retorna `401`.
- Usuário sem autorização retorna `403`.
- Versão divergente retorna `412`.
- Falha em qualquer item não persiste parcialmente o lote.
- O catálogo de serviços permanece inalterado.

---

### 2.3 Checklist de Implementação

**Domínio**

- [ ] Modelar o serviço necessário como parte da Ordem de Serviço
- [ ] Implementar a associação de serviços necessários no agregado `OrdemDeServico`
- [ ] Impedir serviço duplicado na mesma Ordem de Serviço
- [ ] Restringir a associação à situação `EM_DIAGNOSTICO`
- [ ] Preservar a situação da Ordem de Serviço após a associação

**Caso de uso**

- [ ] Implementar `RegistrarServicosNecessarios`
- [ ] Consultar todos os serviços informados no catálogo
- [ ] Orquestrar a inclusão atômica do lote de serviços

**Repositório**

- [ ] Implementar a persistência das associações de serviço da Ordem de Serviço
- [ ] Consultar serviços por lote no `ServicoRepository`
- [ ] Incrementar a versão da Ordem de Serviço na persistência

**Integrações**

- [ ] Integrar a consulta ao catálogo do contexto de Serviços sem alterá-lo

**Handler HTTP**

- [ ] Implementar `POST /api/v1/ordens-servico/{id}/servicos`
- [ ] Obter o usuário autenticado a partir do JWT
- [ ] Retornar os serviços registrados e a nova versão com status `201`
- [ ] Mapear erros para os códigos HTTP documentados

**Validações**

- [ ] Validar lista obrigatória e não vazia
- [ ] Validar `servicoId` obrigatório e sem repetição no corpo
- [ ] Validar existência e situação ativa de todos os serviços
- [ ] Validar diagnóstico iniciado e existência de problema identificado
- [ ] Validar o escopo `os:escrever`

**Concorrência**

- [ ] Comparar `If-Match` com `ordem_servico.version`
- [ ] Retornar `412` sem sobrescrever associações concorrentes quando a versão divergir

**Transação e idempotência**

- [ ] Persistir todo o lote e a nova versão em uma única transação
- [ ] Garantir rollback integral quando qualquer serviço for inválido ou a persistência falhar

**Testes unitários**

- [ ] Registro de um serviço válido
- [ ] Registro de múltiplos serviços
- [ ] Rejeição de serviço duplicado
- [ ] Rejeição de Ordem de Serviço fora de diagnóstico
- [ ] Rejeição sem diagnóstico iniciado ou problema identificado
- [ ] Preservação da situação `EM_DIAGNOSTICO`

**Testes de integração**

- [ ] Resposta `201` com os serviços registrados e a nova versão
- [ ] Resposta `404` para Ordem de Serviço ou serviço inexistente
- [ ] Resposta `409` para serviço inativo, duplicado ou situação incompatível
- [ ] Resposta `400` para lista vazia
- [ ] Respostas `401` e `403` para falhas de acesso
- [ ] Resposta `412` para versão divergente
- [ ] Rollback integral quando um item do lote for inválido
- [ ] Catálogo de serviços inalterado após a operação

**Testes de concorrência**

- [ ] Duas solicitações com a mesma versão não sobrescrevem nem duplicam serviços

**Documentação**

- [ ] Documentar o endpoint, a entrada e os erros no Swagger/OpenAPI

**Review**

- [ ] Code Review aprovado

---
