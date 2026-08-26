---
documento: Refinamento de Requisitos — Registrar Serviços Necessários
dono: A definir
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Registrar Serviços Necessários

Este documento detalha a tarefa Registrar Serviços Necessários do contexto de Ordem de Serviço.

## 4 · Registrar Serviços Necessários

### 4.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Registrar os serviços necessários para o reparo do veículo, adicionando-os ao orçamento
correspondente à etapa atual da Ordem de Serviço.

**Problema**

Depois de identificar os problemas do veículo, o mecânico precisa informar quais serviços serão
necessários para realizar o reparo e compor o orçamento que será apresentado ao cliente.

**Pré-condições**

- A Ordem de Serviço deve existir.
- A OS deve estar `EM_DIAGNOSTICO` ou `EM_EXECUCAO`.
- Deve existir um orçamento com status `CRIADO` correspondente à etapa atual da OS.
- Os serviços devem estar previamente cadastrados e ativos.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-OS-29 | Permitir registrar um ou mais serviços necessários. |
| RF-OS-30 | Permitir informar uma observação para cada serviço. |
| RF-OS-31 | Vincular os serviços diretamente ao orçamento correspondente. |
| RF-OS-32 | Identificar automaticamente o orçamento a partir da Ordem de Serviço. |
| RF-OS-33 | Utilizar o orçamento `PRINCIPAL` quando a OS estiver `EM_DIAGNOSTICO`. |
| RF-OS-34 | Utilizar o orçamento `COMPLEMENTAR` com status `CRIADO` quando a OS estiver `EM_EXECUCAO`. |
| RF-OS-35 | Utilizar o valor cadastrado do serviço na composição do orçamento. |
| RF-OS-36 | Atualizar o valor total do orçamento após o registro. |
| RF-OS-37 | Impedir a inclusão do mesmo serviço mais de uma vez no mesmo orçamento. |
| RF-OS-38 | Impedir alterações em orçamento `APROVADO` ou `RECUSADO`. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-OS-13 | Apenas usuários autorizados devem realizar o registro. |
| RNF-OS-14 | O tipo do orçamento não deve ser informado pelo usuário. |
| RNF-OS-15 | O serviço deve ser vinculado ao orçamento, sem necessidade de vínculo direto com um problema. |
| RNF-OS-16 | O registro dos serviços e a atualização do orçamento devem ocorrer de forma consistente. |

**Fluxo Principal**

1. O mecânico seleciona a Ordem de Serviço.
2. O mecânico informa os serviços necessários.
3. O sistema identifica o status atual da OS.
4. O sistema identifica o orçamento correspondente: `EM_DIAGNOSTICO` leva ao `PRINCIPAL` com
   status `CRIADO`; `EM_EXECUCAO` leva ao `COMPLEMENTAR` com status `CRIADO`.
5. O sistema valida os serviços informados.
6. O sistema adiciona os serviços ao orçamento.
7. O sistema atualiza o valor total do orçamento.
8. O sistema confirma o registro.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Ordem de Serviço não encontrada | Informa que a OS não existe. |
| A2 | OS não está em uma etapa permitida | Impede o registro. |
| A3 | Orçamento correspondente não encontrado | Impede o registro. |
| A4 | Orçamento já está `APROVADO` ou `RECUSADO` | Impede a alteração. |
| A5 | Serviço não encontrado | Impede o registro do item. |
| A6 | Serviço inativo | Impede o registro do item. |
| A7 | Serviço já registrado no mesmo orçamento | Impede a duplicidade. |
| A8 | Usuário sem permissão | Impede a operação. |

**Saída**

- Serviços necessários registrados no orçamento correspondente e valor do orçamento atualizado.

**Pós-condições**

- Os serviços estão vinculados ao orçamento.
- O orçamento permanece com status `CRIADO`.
- O valor total do orçamento reflete os serviços registrados.

---

### 4.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/servicos
```

> **Decisão de projeto.** Esta tarefa **não cria** problema nem orçamento, e **não recebe** o tipo
> do orçamento no corpo da requisição: o tipo é deduzido do status da OS. O orçamento já existe
> porque foi criado no registro do problema encontrado. O serviço é vinculado diretamente ao
> orçamento, e não ao problema — o problema explica o porquê, o orçamento é o que vai ao cliente.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`.
- Escopo: `os:escrever`.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |
| Body | `servicos[]` | array | Obrigatório; ao menos um item. |
| Body | `servicos[].servicoId` | uuid | Obrigatório; serviço do catálogo, ativo. |
| Body | `servicos[].observacao` | string | Opcional. |

```json
{
  "servicos": [
    {
      "servicoId": "7b4e08d5-3c61-4f92-a0d7-51e83b62c40f",
      "observacao": "Necessário para o reparo"
    }
  ]
}
```

**Validações**

*Técnicas*

- `osId` em formato UUID válido.
- `servicos` deve conter pelo menos um item.
- `servicoId` obrigatório; `observacao` opcional.

*Negócio*

- A OS deve existir e estar `EM_DIAGNOSTICO` ou `EM_EXECUCAO`.
- Deve existir orçamento aplicável com status `CRIADO`.
- O serviço deve existir e estar ativo.
- Não é permitido duplicar o mesmo serviço no mesmo orçamento.
- Orçamentos `APROVADO` ou `RECUSADO` não podem ser alterados.

**Regra de domínio**

```
OS EM_DIAGNOSTICO → orçamento PRINCIPAL / CRIADO
OS EM_EXECUCAO    → orçamento COMPLEMENTAR / CRIADO
```

Não existindo o orçamento aplicável, a operação retorna conflito. A criação do orçamento acontece
antes, no registro do problema encontrado.

**Processamento**

1. Buscar a OS pelo identificador e validar sua existência.
2. Validar se está `EM_DIAGNOSTICO` ou `EM_EXECUCAO`.
3. Determinar o tipo de orçamento pelo status da OS.
4. Buscar o orçamento correspondente com status `CRIADO` e validar sua existência.
5. Para cada serviço recebido: buscar o serviço, validar existência e situação ativa, validar que
   ainda não está registrado no orçamento, recuperar o valor atual e registrar no orçamento.
6. Recalcular o valor total do orçamento.
7. Persistir as alterações na mesma transação.
8. Retornar os serviços registrados e o orçamento atualizado.

**Persistência**

- Consulta: `ordem_servico` (status e vínculo), `servico` (catálogo).
- Altera: `orcamento_item` (insert dos serviços vinculados) e `orcamento` (`data_atualizacao`).

Campos usados de `orcamento_item`: `id`, `orcamento_id`, `servico_id`, `tipo_item`, `descricao`,
`quantidade`, `valor_unitario`, `valor_total` e `observacao`.

Regras de persistência: o serviço deve pertencer ao orçamento identificado para a OS; não é
permitida duplicidade de `servico_id` dentro do mesmo orçamento; a inclusão dos serviços e a
atualização da data do orçamento ocorrem na mesma transação; o valor total é calculado pela soma
dos itens; em caso de erro, nenhuma alteração parcial
permanece salva.

**Saída da API**

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "orcamento": {
    "id": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
    "tipo": "PRINCIPAL",
    "status": "CRIADO",
    "valorTotal": 250.0
  },
  "servicos": [
    {
      "servicoId": "7b4e08d5-3c61-4f92-a0d7-51e83b62c40f",
      "descricao": "Troca de pastilhas de freio",
      "valorUnitario": 250.0,
      "observacao": "Necessário para o reparo"
    }
  ]
}
```

Durante a execução, o mesmo contrato responde com o orçamento complementar:

```json
{
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "orcamento": {
    "id": "b1d47c60-92fe-4a38-8c15-73e0a6b5d284",
    "tipo": "COMPLEMENTAR",
    "status": "CRIADO",
    "valorTotal": 180.0
  },
  "servicos": [
    {
      "servicoId": "c05d9a37-1e48-4b26-90f5-72a3e6c81b04",
      "descricao": "Substituição da bomba d'água",
      "valorUnitario": 180.0,
      "observacao": "Serviço necessário durante a execução"
    }
  ]
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Serviços registrados com sucesso. |
| `400` | Payload inválido. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `os:escrever`. |
| `404` | OS ou serviço não encontrado. |
| `409` | OS em status não permitido; orçamento aplicável não encontrado; orçamento não está `CRIADO`; serviço já registrado no orçamento. |

**Dependências**

- `OrdemDeServicoRepository`.
- `OrcamentoRepository`.
- `ServicoRepository`.
- Middleware de autenticação/autorização.

**Testes**

*Unitários*

- Determina o tipo do orçamento pelo status da OS.
- Recalcula o valor total do orçamento.
- Rejeita serviço inexistente, inativo ou já registrado no orçamento.
- Rejeita alteração de orçamento `APROVADO` ou `RECUSADO`.

*Integração*

- Registro em orçamento `PRINCIPAL` com a OS `EM_DIAGNOSTICO`.
- Registro em orçamento `COMPLEMENTAR` com a OS `EM_EXECUCAO`.
- Vários serviços na mesma requisição.
- OS inexistente retorna `404` e serviço inexistente retorna `404`.
- OS em status não permitido retorna `409`.
- Orçamento aplicável inexistente retorna `409`.
- Serviço duplicado retorna `409`.
- Rollback completo em caso de falha de persistência.

---

### 4.3 Checklist de Implementação

**Domínio**

- [ ] Determinar automaticamente o tipo do orçamento pelo status da OS
- [ ] Registrar os serviços diretamente no orçamento
- [ ] Recalcular o valor total do orçamento
- [ ] Impedir serviço duplicado no mesmo orçamento
- [ ] Impedir alteração de orçamento `APROVADO` ou `RECUSADO`
- [ ] Garantir que esta tarefa não crie um novo orçamento

**Caso de uso**

- [ ] Implementar `RegistrarServicosNecessarios`
- [ ] Validar que a OS exista e esteja `EM_DIAGNOSTICO` ou `EM_EXECUCAO`
- [ ] Buscar `PRINCIPAL` / `CRIADO` quando a OS estiver `EM_DIAGNOSTICO`
- [ ] Buscar `COMPLEMENTAR` / `CRIADO` quando a OS estiver `EM_EXECUCAO`
- [ ] Validar que o orçamento correspondente exista
- [ ] Validar que a lista de serviços não esteja vazia
- [ ] Validar que cada serviço exista e esteja ativo
- [ ] Recuperar o valor atual cadastrado de cada serviço

**Repositório**

- [ ] Persistir os registros em `orcamento_item`
- [ ] Calcular o novo valor total do orçamento

**Transação**

- [ ] Executar a inclusão dos serviços e a atualização do orçamento na mesma transação
- [ ] Garantir rollback completo em caso de erro

**Handler HTTP**

- [ ] Criar o handler para `POST /ordens-servico/{osId}/servicos`
- [ ] Criar DTO/request de entrada e DTO/response de saída
- [ ] Aplicar autenticação e autorização na rota
- [ ] Mapear os erros para os códigos HTTP definidos

**Testes unitários**

- [ ] Registro em orçamento `PRINCIPAL` / `CRIADO`
- [ ] Registro em orçamento `COMPLEMENTAR` / `CRIADO`
- [ ] Múltiplos serviços
- [ ] Serviço duplicado
- [ ] Serviço inexistente
- [ ] Serviço inativo
- [ ] OS inexistente
- [ ] OS em status não permitido
- [ ] Orçamento correspondente inexistente
- [ ] Orçamento `APROVADO` e orçamento `RECUSADO`
- [ ] Recálculo do valor total

**Testes de integração**

- [ ] Endpoint registrando os serviços e devolvendo o orçamento atualizado
- [ ] Rollback em caso de falha de persistência

**Documentação**

- [ ] Documentar o endpoint no OpenAPI/Swagger

**Review**

- [ ] Executar testes automatizados
- [ ] Code Review aprovado
- [ ] Validar os critérios de aceite da task

---
