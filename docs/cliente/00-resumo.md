---
documento: Resumo do Contexto — Cliente
dono: A definir
versao: 0.1
atualizado_em: 2026-08-22
status: em construcao
---

# Resumo do Contexto — Cliente

## O que é este documento

Um retrato do que existe hoje neste diretório: as tarefas refinadas, as rotas que elas expõem, os
tipos e enums do contexto e as convenções que valem aqui. Serve para quem chega saber o que já
está decidido sem abrir os cinco documentos, e para quem for implementar conferir se o que vai
escrever bate com o que já existe.

O que **não** está resolvido não fica aqui: vai para [`pontos-em-aberto.md`](pontos-em-aberto.md).

## O que este contexto cobre

O cadastro do cliente da oficina, sua identificação por CPF/CNPJ e o vínculo com os veículos que
ele traz para atendimento. É o ponto de entrada do fluxo: sem cliente identificado não se abre
Ordem de Serviço.

## Tarefas documentadas

| # | Tarefa | Rota | Escopo | Arquivo |
|---|---|---|---|---|
| 1 | Consultar Cliente | `GET /clientes` | `clientes:ler` | [consultar-cliente.md](consultar-cliente.md) |
| 2 | Cadastrar Cliente | `POST /clientes` | `clientes:escrever` | [cadastrar-cliente.md](cadastrar-cliente.md) |
| 3 | Atualizar Cliente | `PUT /clientes/{clienteId}` | `clientes:escrever` | [atualizar-cliente.md](atualizar-cliente.md) |
| 4 | Vincular Veículo ao Cliente | `POST /clientes/{clienteId}/veiculos/{veiculoId}` | `clientes:escrever` | [vincular-veiculo-ao-cliente.md](vincular-veiculo-ao-cliente.md) |
| 5 | Deletar Cliente | `DELETE /clientes/{clienteId}` e `POST /clientes/{clienteId}/reativacao` | `clientes:escrever` | [deletar-cliente.md](deletar-cliente.md) |

## Tipos do contexto

**Cliente**

| Campo | Tipo | Observação |
|---|---|---|
| `id` | uuid | Identificador técnico, gerado pelo sistema. |
| `nome` | string | Obrigatório. |
| `documento` | string | CPF ou CNPJ, validado. Identificador de negócio. |
| `tipoDocumento` | enum | `CPF` \| `CNPJ`. |
| `ativo` | boolean | `false` após a exclusão lógica. |
| `inativadoEm` | datetime | Preenchido na inativação. |
| `inativadoPor` | uuid | Usuário responsável pela inativação. |
| `motivoInativacao` | string | Até 200 caracteres, opcional. |

**Vínculo cliente-veículo**

O cliente conhece a lista de seus veículos. O vínculo é criado por rota sob `/clientes` e é o
cadastro do cliente que é atualizado na operação.

## Convenções em vigor neste contexto

- Rotas sem prefixo de versão; recurso no plural; identificadores em rota sempre UUID.
- Autenticação `Bearer <JWT>`; perfis `MECANICO` e `GESTOR`; escopos `clientes:ler` e `clientes:escrever`.
- Exclusão é **lógica**: o `DELETE` inativa e preserva o histórico das Ordens de Serviço.
- A unicidade de CPF/CNPJ vale **apenas entre clientes ativos**, por índice parcial
  `UNIQUE (cpf_cnpj) WHERE ativo = true`.
- A inativação do cliente inativa os veículos vinculados, por política; a reativação **não**
  reativa os veículos em cascata.
- Códigos de erro usados: `400`, `401`, `403`, `404`, `409` e `204` para operação idempotente.

## O que este contexto não faz

- Não guarda contato do cliente (telefone ou e-mail), embora o fluxo do negócio preveja avisar o
  cliente sobre orçamento, atraso e conclusão.
- Não trata anonimização de dados pessoais.
- Não decide sobre a Ordem de Serviço: apenas é consultado por ela.
