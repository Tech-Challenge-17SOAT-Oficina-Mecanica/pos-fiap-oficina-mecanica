---
documento: Refinamento de Requisitos — Enviar Orçamento
dono: Desconhecido
versao: 0.1
atualizado_em: 2026-08-20
status: rascunho
---

# Refinamento de Requisitos — Enviar Orçamento

Este documento detalha a tarefa Enviar Orçamento do contexto de Orçamento.

## 1 · Enviar Orçamento

### 1.1 Refinamento de Produto

**Persona**

Sistema.

**Objetivo**

Enviar o orçamento ao cliente para análise e aprovação.

**Problema**

O cliente precisa receber o orçamento para decidir se autoriza a execução dos serviços.

**Pré-condições**

- Deve existir uma OS.
- Deve existir um orçamento associado à OS.
- A OS deve estar com status `AGUARDANDO_APROVACAO`.
- O cliente deve possuir um canal de comunicação válido.
- O orçamento deve possuir itens e valor total.

**Requisitos Funcionais**

- Identificar o orçamento da OS aguardando aprovação.
- Obter os dados de contato do cliente.
- Enviar o orçamento pelo canal de comunicação definido.
- Registrar data/hora, canal e resultado do envio.
- Associar o envio ao orçamento.
- Manter a OS em `AGUARDANDO_APROVACAO`.

**Requisitos Não Funcionais**

- O envio deve utilizar canal de comunicação seguro.
- Os dados do cliente e do orçamento devem ser protegidos.
- O sistema deve registrar o histórico do envio.
- O envio deve ser rastreável.

**Fluxo Principal**

1. O sistema identifica a OS em `AGUARDANDO_APROVACAO`.
2. O sistema consulta o orçamento e os dados de contato do cliente.
3. O sistema valida o canal de comunicação.
4. O sistema prepara as informações do orçamento.
5. O sistema envia o orçamento ao cliente.
6. O sistema registra o envio com status `ENVIADO`.
7. A OS permanece em `AGUARDANDO_APROVACAO`.

**Fluxos Alternativos / Exceções**

- Cliente sem contato válido: o sistema não realiza o envio e registra a pendência.
- Orçamento já enviado: o sistema impede novo envio, conforme regra de negócio.
- Falha no envio: o sistema registra o envio como `FALHA`.
- OS fora de `AGUARDANDO_APROVACAO`: o sistema não realiza o envio.

**Saída**

- Orçamento enviado ao cliente.

**Pós-condições**

- Envio registrado.
- Cliente recebe o orçamento para análise.
- OS permanece em `AGUARDANDO_APROVACAO`.

### 1.2 Refinamento Técnico

**Gatilho**

- Após a geração do orçamento e a atualização da OS para `AGUARDANDO_APROVACAO`, o sistema executa internamente o envio ao cliente.

**Autorização**

- O processo interno deve ter acesso ao orçamento, à OS e aos dados de contato do cliente.
- O acesso aos dados pessoais do cliente deve ser restrito.

**Entrada**

- `orcamentoId`: identificador do orçamento a ser enviado.

**Validações**

- Validar se o orçamento existe.
- Validar se o orçamento está associado a uma OS.
- Validar se a OS está em `AGUARDANDO_APROVACAO`.
- Validar se a OS possui cliente vinculado.
- Validar se o cliente possui canal de comunicação válido.
- Validar se o orçamento possui itens e valor total.
- Validar se já existe envio realizado, caso reenvio não seja permitido.

**Processamento**

1. Receber o identificador do orçamento.
2. Consultar o orçamento e a OS associada.
3. Consultar o cliente e seus dados de contato.
4. Validar o canal de comunicação.
5. Montar o conteúdo do orçamento.
6. Enviar o orçamento pelo serviço de comunicação.
7. Registrar data/hora, canal e resultado do envio.
8. Persistir o histórico do envio.
9. Manter a OS em `AGUARDANDO_APROVACAO`.
10. Registrar a operação em log.

**Persistência**

- Criar registro de envio do orçamento:
  - `id`;
  - `orcamentoId`;
  - `canal`;
  - `dataEnvio`;
  - `status`: `ENVIADO` ou `FALHA`;
  - `motivoFalha`, quando aplicável.

**Dependências**

- `OrcamentoRepository`.
- `OrdemDeServicoRepository`.
- `ClienteRepository`.
- `EnvioOrcamentoRepository`.
- Serviço de comunicação, como e-mail ou WhatsApp.
- Banco de dados.

**Testes**

- Deve enviar o orçamento quando a OS estiver em `AGUARDANDO_APROVACAO`.
- Deve registrar data/hora, canal e status `ENVIADO`.
- Deve manter a OS em `AGUARDANDO_APROVACAO`.
- Não deve enviar quando o cliente não possuir contato válido.
- Não deve enviar quando a OS não estiver aguardando aprovação.
- Deve registrar status `FALHA` quando o serviço de comunicação falhar.
- Deve impedir reenvio quando já existir envio realizado, se esta for a regra definida.
- Deve garantir que dados sensíveis não sejam registrados em logs.

### 1.3 Check-list de Implementação

- [ ] Criar/ajustar a entidade `EnvioOrcamento`.
- [ ] Definir os campos de canal, data de envio, status e motivo de falha.
- [ ] Criar `EnvioOrcamentoRepository`.
- [ ] Implementar o caso de uso `EnviarOrcamento`.
- [ ] Consultar orçamento, OS e cliente vinculados.
- [ ] Validar se a OS está em `AGUARDANDO_APROVACAO`.
- [ ] Validar o canal de comunicação do cliente.
- [ ] Validar itens e valor total do orçamento.
- [ ] Implementar a montagem do conteúdo do orçamento.
- [ ] Integrar o serviço de comunicação definido para o MVP.
- [ ] Registrar envio com status `ENVIADO`.
- [ ] Registrar falha com status `FALHA` e motivo.
- [ ] Impedir reenvio duplicado, se essa for a regra definida.
- [ ] Garantir que a OS permaneça em `AGUARDANDO_APROVACAO`.
- [ ] Proteger dados pessoais do cliente e não registrá-los em logs.
- [ ] Criar testes unitários para envio válido, contato inválido e falha de envio.
- [ ] Criar teste de integração com o serviço de comunicação.
- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto.
- [ ] Executar testes automatizados, code review e validar critérios de aceite.
