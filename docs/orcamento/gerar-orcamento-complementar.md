---
documento: Refinamento de Requisitos — Gerar Orçamento Complementar
dono: A definir
versao: 0.1
atualizado_em: 2026-08-19
status: rascunho
---

# Refinamento de Requisitos — Gerar Orçamento Complementar

Este documento detalha a tarefa Gerar Orçamento Complementar do contexto de Orçamento.

## 5 · Gerar Orçamento Complementar

### 5.1 Refinamento de Produto

**Persona**

Mecânico, responsável pelo atendimento.

**Objetivo**

Consolidar os serviços, peças e insumos adicionais identificados durante o atendimento e já
registrados na Ordem de Serviço, gerando um orçamento complementar para análise e aprovação do
cliente.

**Problema**

Durante a execução de uma Ordem de Serviço podem ser identificadas necessidades adicionais que
não estavam previstas no orçamento original. A oficina precisa consolidar esses itens, calcular
seus valores e apresentar um novo orçamento ao cliente antes de executar os itens adicionais.

**Pré-condições**

- Deve existir uma Ordem de Serviço, com cliente vinculado.
- A OS deve estar em uma etapa que permita a geração de orçamento complementar.
- Deve existir pelo menos um serviço, peça ou insumo adicional previamente registrado na OS.
- Os itens registrados devem possuir quantidade válida.
- Os itens registrados devem possuir valor válido no momento em que foram associados à OS.
- Não deve existir outro orçamento complementar pendente de aprovação para a mesma OS.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-ORC-29 | Permitir gerar um orçamento complementar a partir dos itens adicionais já registrados na OS. |
| RF-ORC-30 | Consolidar os serviços adicionais registrados na OS. |
| RF-ORC-31 | Consolidar as peças adicionais registradas na OS. |
| RF-ORC-32 | Consolidar os insumos adicionais registrados na OS. |
| RF-ORC-33 | Calcular automaticamente o valor individual de cada item. |
| RF-ORC-34 | Calcular automaticamente o valor total do orçamento complementar. |
| RF-ORC-35 | Utilizar os valores registrados nos itens da OS, preservando o histórico apresentado ao cliente. |
| RF-ORC-36 | Registrar uma observação ou justificativa relacionada ao orçamento, quando aplicável. |
| RF-ORC-37 | Associar o orçamento complementar à Ordem de Serviço. |
| RF-ORC-38 | Registrar o orçamento complementar como pendente de aprovação. |
| RF-ORC-39 | Disponibilizar o orçamento complementar para o cliente. |
| RF-ORC-40 | Disparar a notificação ao cliente sobre a necessidade de aprovação. |
| RF-ORC-41 | Impedir a geração de um novo orçamento complementar pendente para a mesma OS. |
| RF-ORC-42 | Não autorizar automaticamente a execução dos itens adicionais. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-ORC-15 | A geração do orçamento e de seus itens deve ser persistida de forma consistente. |
| RNF-ORC-16 | O cálculo dos valores deve usar representação monetária adequada, evitando inconsistências de arredondamento. |
| RNF-ORC-17 | O orçamento complementar deve manter vínculo com a Ordem de Serviço original. |
| RNF-ORC-18 | Os valores apresentados ao cliente devem permanecer rastreáveis mesmo que os valores do catálogo mudem depois. |
| RNF-ORC-19 | Somente usuários autorizados devem poder solicitar a geração do orçamento complementar. |
| RNF-ORC-20 | A criação do orçamento e a atualização das informações da OS devem ocorrer de forma consistente. |
| RNF-ORC-21 | Uma falha no envio da notificação não deve causar perda do orçamento já gerado. |
| RNF-ORC-22 | A operação deve manter a rastreabilidade dos itens, valores, data e responsável pela geração. |

**Fluxo Principal**

1. O responsável pelo atendimento acessa uma OS com itens adicionais registrados.
2. O responsável solicita a geração do orçamento complementar.
3. O sistema verifica se a OS existe e se está em etapa que permita o orçamento complementar.
4. O sistema verifica se existem itens adicionais registrados na OS.
5. O sistema recupera os serviços, peças e insumos adicionais, com seus valores e quantidades.
6. O sistema calcula o subtotal de cada item e o valor total do orçamento.
7. O sistema cria o orçamento complementar e o associa à OS.
8. O sistema registra o orçamento como pendente de aprovação, com data, hora e responsável.
9. O sistema disponibiliza o orçamento para o cliente e dispara a notificação.
10. O sistema confirma a geração do orçamento complementar.

**Fluxos Alternativos / Exceções**

| ID | Situação | Comportamento do sistema |
|---|---|---|
| A1 | OS não encontrada | Informa que a Ordem de Serviço não existe. |
| A2 | OS em status inválido | Impede a geração quando a OS não está em etapa que permita orçamento complementar. |
| A3 | Nenhum item adicional registrado | Impede a geração e informa que não existem itens para composição. |
| A4 | Item com quantidade inválida | Impede a geração e informa o item pendente. |
| A5 | Item sem valor registrado | Impede a geração até que o item possua valor válido. |
| A6 | Orçamento complementar pendente existente | Impede a criação de outro orçamento pendente para a mesma OS. |
| A7 | Itens indisponíveis | Respeita a regra de negócio definida para serviços, peças e insumos já registrados. |
| A8 | Erro no cálculo | Não cria o orçamento caso os valores não possam ser calculados corretamente. |
| A9 | Falha na persistência | Nenhuma alteração parcial permanece registrada. |
| A10 | Falha na notificação | O orçamento permanece criado e pendente; a falha é registrada e a notificação pode ser reenviada. |
| A11 | Usuário sem autorização | Impede a operação. |

**Saída**

- Orçamento complementar criado e associado à OS, com os itens consolidados, o valor total
  calculado, pendente de aprovação e notificação disparada ao cliente.

**Pós-condições**

- O orçamento complementar e seus itens ficam registrados, com valores e quantidades preservados.
- Os itens complementares ainda não são considerados autorizados para execução.
- A OS mantém seu histórico e estado atual.
- O cliente fica apto a aprovar ou recusar o orçamento complementar.
- Nenhuma movimentação de estoque é realizada apenas pela geração do orçamento.

---

### 5.2 Refinamento Técnico

**Endpoint**

```http
POST /ordens-servico/{osId}/orcamentos-complementares
```

> **Decisão de projeto.** Este caso de uso **não cadastra itens**: os serviços, peças e insumos
> adicionais precisam já existir na OS, registrados pelos casos de uso do contexto de Ordem de
> Serviço. Por isso a requisição não tem corpo. A alternativa seria receber os itens no body, o
> que duplicaria o cadastro de item e abriria espaço para orçamento divergente da OS.

**Autenticação / Autorização**

- `Bearer <JWT>` obrigatório.
- Perfis: `MECANICO`, `GESTOR`.
- Escopo: `orcamentos:escrever`.
- O identificador do responsável é obtido do token.

**Entrada**

| Local | Parâmetro | Tipo | Descrição |
|---|---|---|---|
| Path | `osId` | uuid | Identificador da Ordem de Serviço. |

Não há corpo na requisição: o sistema usa a OS informada para localizar os itens adicionais.

**Validações**

*Técnicas*

- `osId` em formato UUID válido.

*Negócio*

- A OS deve existir.
- A OS deve estar em etapa que permita a geração do orçamento complementar; não pode estar
  `FINALIZADA` nem `ENTREGUE`.
- Deve existir pelo menos um item adicional registrado na OS.
- Os itens devem possuir quantidade e valor válidos.
- Não pode existir outro orçamento complementar pendente de aprovação para a mesma OS.

**Regra de domínio**

```
OS elegível → itens adicionais registrados → gerar orçamento complementar → pendente de aprovação
```

O orçamento complementar é vinculado ao orçamento principal da mesma OS pelo campo
`orcamentoOriginalId`, com `tipo = COMPLEMENTAR`.

**Processamento**

1. Receber o identificador da OS e identificar o usuário autenticado.
2. Buscar a Ordem de Serviço e validar existência, autorização e status.
3. Verificar se já existe orçamento complementar pendente.
4. Buscar os serviços, peças e insumos adicionais registrados na OS.
5. Validar que existe pelo menos um item e que quantidade e valor são válidos.
6. Calcular o subtotal de cada item, usando os valores registrados no contexto da OS.
7. Calcular o valor total do orçamento.
8. Criar o orçamento complementar e seus itens, associados à OS e ao orçamento principal.
9. Registrar data, hora e usuário responsável.
10. Persistir o orçamento e os itens de forma transacional.
11. Disparar a notificação ao cliente e registrar o resultado do envio.
12. Retornar o orçamento criado.

**Persistência**

- Consulta: `ordem_servico` e seus itens adicionais, `orcamento` (principal da OS).
- Altera: `orcamento` (insert com `tipo = COMPLEMENTAR`), `orcamento_item` (insert).
- Tudo em uma única transação; a notificação acontece depois do commit.

**Saída da API**

```json
{
  "orcamentoId": "b1d47c60-92fe-4a38-8c15-73e0a6b5d284",
  "ordemServicoId": "5d8f2a30-61c4-4e79-b3d2-9a7e4f10c586",
  "tipo": "COMPLEMENTAR",
  "orcamentoOriginalId": "9c2a71f8-4e35-4d19-b8a6-27f0e5c4a913",
  "itens": [
    {
      "tipo": "SERVICO",
      "itemId": "7b4e08d5-3c61-4f92-a0d7-51e83b62c40f",
      "descricao": "Troca de pastilha",
      "quantidade": 1,
      "valorUnitario": 200.0,
      "valorTotal": 200.0
    },
    {
      "tipo": "PECA",
      "itemId": "3f1a9c2e-4b7d-4f56-9a10-0c8e5d21b7a4",
      "descricao": "Pastilha de freio dianteira",
      "quantidade": 2,
      "valorUnitario": 180.0,
      "valorTotal": 360.0
    }
  ],
  "valorTotal": 560.0,
  "dataGeracao": "2026-08-18T21:30:00-03:00"
}
```

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `201` | Orçamento complementar criado. |
| `400` | Requisição inválida. |
| `401` | Token ausente ou expirado. |
| `403` | Perfil sem o escopo `orcamentos:escrever`. |
| `404` | OS não encontrada. |
| `409` | OS em estado incompatível; já existe orçamento complementar pendente; não existem itens adicionais para compor o orçamento. |

**Notificação**

A geração do orçamento e a notificação são responsabilidades distintas. Depois da persistência
bem-sucedida, o evento `OrcamentoComplementarGerado` dispara a notificação ao cliente. Se a
notificação falhar, o orçamento permanece criado e pendente de aprovação, a falha é registrada e
o envio pode ser refeito. Falha de notificação nunca provoca rollback do orçamento.

**Dependências**

- `OrdemDeServicoRepository`.
- `OrcamentoRepository`.
- `OrcamentoItemRepository`.
- Consulta dos itens adicionais registrados na OS.
- Serviço de notificação ao cliente.
- Middleware de autenticação/autorização.
- Mecanismo de logs e auditoria.

**Testes**

*Unitários*

- Cálculo correto dos subtotais e do valor total.
- Preservação dos valores registrados na OS.
- Vínculo com o orçamento principal pelo `orcamentoOriginalId`.
- Rejeita OS sem itens adicionais.
- Rejeita item sem valor ou com quantidade inválida.
- Rejeita segundo orçamento complementar pendente para a mesma OS.

*Integração*

- OS elegível gera o orçamento e retorna `201`.
- O orçamento usa os itens adicionais já registrados na OS e é associado à OS correta.
- OS inexistente retorna `404`.
- OS `FINALIZADA` ou `ENTREGUE` retorna `409`.
- OS sem itens adicionais retorna `409`.
- Orçamento e itens são persistidos de forma transacional.
- A notificação é disparada após a criação e a falha dela não desfaz o orçamento.
- Perfil sem escopo retorna `403`.

---

### 5.3 Checklist de Implementação

**Domínio**

- [ ] Reaproveitar o agregado `Orcamento` com `tipo = COMPLEMENTAR`
- [ ] Garantir o vínculo com o orçamento principal pelo `orcamentoOriginalId`
- [ ] Definir os tipos de item: `SERVICO`, `PECA` e `INSUMO`
- [ ] Confirmar que os itens adicionais são registrados previamente na OS
- [ ] Implementar o cálculo do subtotal de cada item e do valor total
- [ ] Garantir o uso dos valores registrados no contexto da OS

**Caso de uso**

- [ ] Implementar `GerarOrcamentoComplementar`
- [ ] Validar que a OS existe e está em status que permita o complemento
- [ ] Impedir geração para OS `FINALIZADA` ou `ENTREGUE`
- [ ] Validar a existência de itens adicionais, com quantidade e valor válidos
- [ ] Impedir orçamento complementar pendente duplicado
- [ ] Registrar data, hora e usuário responsável

**Repositório**

- [ ] Criar ou ajustar `OrcamentoRepository` e `OrcamentoItemRepository`
- [ ] Implementar a recuperação dos serviços, peças e insumos adicionais da OS

**Integrações**

- [ ] Consultar o módulo de Ordem de Serviço para status e itens adicionais
- [ ] Implementar o envio da notificação ao cliente
- [ ] Registrar sucesso ou falha da notificação

**Handler HTTP**

- [ ] Implementar `POST /ordens-servico/{osId}/orcamentos-complementares`
- [ ] Validar o path param `osId`
- [ ] Criar DTO/response de saída
- [ ] Aplicar autenticação JWT e autorização por escopo na rota
- [ ] Mapear erros de domínio para os códigos HTTP documentados

**Transação**

- [ ] Persistir orçamento e itens em uma única transação
- [ ] Garantir que a falha da notificação não faça rollback do orçamento

**Eventos**

- [ ] Publicar `OrcamentoComplementarGerado`

**Testes unitários**

- [ ] Cálculo dos subtotais e do valor total
- [ ] Ausência de itens adicionais
- [ ] Quantidade inválida
- [ ] Valor inválido
- [ ] Orçamento complementar pendente duplicado

**Testes de integração**

- [ ] Geração válida com itens previamente registrados
- [ ] `404` para OS inexistente e `409` para OS em status inválido
- [ ] Persistência transacional do orçamento e dos itens
- [ ] Falha da notificação sem rollback
- [ ] `403` para usuário sem autorização

**Documentação**

- [ ] Documentar o endpoint no Swagger/OpenAPI
- [ ] Documentar o evento de geração

**Review**

- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto
- [ ] Executar testes automatizados
- [ ] Code Review aprovado

---
