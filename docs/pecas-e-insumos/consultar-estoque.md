---
documento: Refinamento de Requisitos — Consultar Peças
dono: Desconhecido
versao: 0.2
atualizado_em: 2026-08-22
status: rascunho
---

# Refinamento de Requisitos — Consultar Peças

Este documento detalha a tarefa Consultar Peças do contexto de Peças & Insumos.

## 11 · Consultar Peças

### 11.1 Refinamento de Produto

**Persona**

Mecânico.

**Objetivo**

Consultar peças do estoque por código, descrição ou categoria, identificando se existe saldo disponível para utilização no serviço.

**Problema**

Durante o diagnóstico, o mecânico precisa saber se a peça necessária está disponível antes de registrá-la na Ordem de Serviço. A consulta deve informar não apenas se a peça está cadastrada, mas também seu saldo efetivamente disponível para uso.

**Pré-condições**

- Deve existir cadastro de peças disponível para consulta.
- O usuário deve estar autorizado a consultar o estoque.
- Os dados de estoque das peças devem estar disponíveis para consulta.

**Requisitos Funcionais**

| ID | Requisito |
|---|---|
| RF-PEC-01 | Permitir consultar peça por código exato. |
| RF-PEC-02 | Permitir consultar peça por descrição parcial. |
| RF-PEC-03 | Permitir filtrar por categoria e fabricante. |
| RF-PEC-04 | Permitir filtrar apenas peças com saldo disponível. |
| RF-PEC-05 | Exibir saldo físico, saldo reservado e saldo disponível da peça. |
| RF-PEC-06 | Indicar se a peça está disponível para uso. |
| RF-PEC-07 | Permitir informar quantidade desejada e indicar se o saldo atende à necessidade. |
| RF-PEC-08 | Indicar se a peça está abaixo do estoque mínimo. |
| RF-PEC-09 | Indicar se existe pedido de compra em aberto para a peça. |
| RF-PEC-10 | Excluir peças inativas do resultado por padrão, salvo quando solicitado. |
| RF-PEC-11 | Informar quando nenhuma peça corresponde aos parâmetros informados. |

**Requisitos Não Funcionais**

| ID | Requisito |
|---|---|
| RNF-PEC-01 | A consulta deve ser realizada por API RESTful. |
| RNF-PEC-02 | A operação deve ser acessível somente por usuário autorizado. |
| RNF-PEC-03 | A consulta não deve alterar o saldo, gerar reserva ou solicitação de compra. |
| RNF-PEC-04 | A listagem deve ser paginada. |
| RNF-PEC-05 | O saldo apresentado deve refletir o estado do estoque no momento da consulta. |
| RNF-PEC-06 | A disponibilidade deve ser calculada pelo saldo disponível, e não apenas pelo saldo físico. |

**Fluxo Principal**

1. O mecânico acessa a consulta de peças.
2. O mecânico informa os parâmetros de busca.
3. O mecânico pode informar a quantidade desejada.
4. O sistema valida os parâmetros e a autorização do usuário.
5. O sistema consulta as peças correspondentes aos critérios informados.
6. O sistema calcula o saldo disponível de cada peça.
7. O sistema compara o saldo disponível com a quantidade desejada, quando informada.
8. O sistema identifica estoque mínimo e pedido de compra em aberto.
9. O sistema retorna as peças encontradas com suas informações de estoque.
10. O mecânico visualiza a disponibilidade da peça para utilização na OS.

**Fluxos Alternativos / Exceções**

| # | Situação | Comportamento do sistema |
|---|---|---|
| A1 | Nenhuma peça encontrada | O sistema informa que nenhuma peça corresponde aos parâmetros informados. |
| A2 | Peça sem saldo disponível | O sistema retorna a peça com `disponivel: false`. |
| A3 | Quantidade desejada maior que o saldo | O sistema informa que a quantidade solicitada não está totalmente disponível. |
| A4 | Peça abaixo do estoque mínimo | O sistema retorna a peça disponível, sinalizada como abaixo do estoque mínimo. |
| A5 | Peça com pedido de compra em aberto | O sistema informa que existe pedido de compra em aberto. |
| A6 | Peça inativa | O sistema não a retorna na consulta padrão, salvo quando explicitamente solicitado. |
| A7 | Parâmetro inválido | O sistema informa qual parâmetro está incorreto. |
| A8 | Usuário sem autorização | O sistema impede a consulta. |
| A9 | Falha na consulta | O sistema informa a indisponibilidade da operação sem alterar o estoque. |

**Saída**

- Relação de peças correspondentes aos parâmetros informados, com saldo físico, reservado e disponível, quantidade desejada, indicação de disponibilidade, estoque mínimo e pedido de compra, quando aplicável; ou
- Indicação de que nenhuma peça foi encontrada.

**Pós-condições**

- Nenhum dado do estoque é alterado.
- Nenhuma peça é reservada.
- Nenhum pedido de compra é criado.
- O mecânico possui as informações necessárias para registrar a peça na OS ou seguir para uma eventual solicitação de compra.

### 11.2 Refinamento Técnico

**Endpoint**

- `GET /estoque/pecas`

**Autenticação / Autorização**

- Requer autenticação JWT.
- Permitido para usuário com perfil/permissão de Mecânico ou Gestor.
- Requer escopo `estoque:ler`.

**Entrada**

Todos os campos são query params.

| Parâmetro | Tipo | Descrição |
|---|---|---|
| `codigo` | string | Código exato da peça. |
| `descricao` | string | Busca parcial pela descrição. |
| `categoria` | string | Filtro por categoria. |
| `fabricante` | string | Filtro por fabricante. |
| `somenteDisponiveis` | boolean | Retorna somente peças com saldo disponível. |
| `incluirInativos` | boolean | Inclui peças inativas. Default: `false`. |
| `quantidadeDesejada` | number | Quantidade que o mecânico deseja utilizar. |
| `page` | integer | Página da consulta. Default: `0`. |
| `size` | integer | Tamanho da página. Default: `20`; máximo: `100`. |

Exemplos:

```http
GET /estoque/pecas?codigo=PC-0142
GET /estoque/pecas?descricao=pastilha&categoria=Freios
GET /estoque/pecas?codigo=PC-0142&quantidadeDesejada=4
GET /estoque/pecas?descricao=pastilha&somenteDisponiveis=true
```

**Validações**

- Validar autenticação, perfil e escopo do usuário.
- Exigir ao menos um filtro de busca: `codigo` ou `descricao`.
- Validar que `descricao` possui no mínimo 2 caracteres, quando informada.
- Validar que `quantidadeDesejada` é maior que zero, quando informada.
- Validar que `page` é maior ou igual a zero.
- Validar que `size` não excede 100.
- Validar o formato dos filtros informados.
- Garantir que a consulta não altera o estoque, não cria reserva e não cria pedido de compra.

**Processamento**

1. Receber e normalizar os parâmetros da requisição.
2. Identificar o usuário autenticado e validar sua autorização.
3. Montar a consulta com os filtros informados e peças ativas por padrão.
4. Consultar as peças no `ItemEstoqueRepository`.
5. Calcular `saldoDisponivel = saldoFisico - saldoReservado`.
6. Derivar `disponivel = saldoDisponivel > 0`.
7. Quando houver quantidade desejada, calcular `quantidadeDisponivel = saldoDisponivel >= quantidadeDesejada`.
8. Derivar `abaixoDoMinimo = saldoDisponivel < estoqueMinimo`.
9. Consultar pedidos de compra em aberto para as peças encontradas.
10. Aplicar o filtro `somenteDisponiveis`, quando informado.
11. Montar e retornar a resposta paginada.

**Persistência**

- Consultar `item_estoque`.
- Consultar `pedido_compra_item`.
- A operação é somente leitura e não deve executar transação de escrita.

**Saída da API**

```json
{
  "conteudo": [
    {
      "codigo": "PC-0142",
      "tipo": "PECA",
      "descricao": "Pastilha de freio dianteira",
      "categoria": "Freios",
      "fabricante": "Bosch",
      "unidadeMedida": "UN",
      "precoVenda": 189.90,
      "saldoFisico": 6,
      "saldoReservado": 2,
      "saldoDisponivel": 4,
      "quantidadeDesejada": 3,
      "quantidadeDisponivel": true,
      "estoqueMinimo": 4,
      "disponivel": true,
      "abaixoDoMinimo": false,
      "possuiPedidoEmAberto": false,
      "ativo": true
    }
  ],
  "pagina": 0,
  "tamanho": 20,
  "totalElementos": 1,
  "totalPaginas": 1
}
```

> Peça sem saldo não significa peça inexistente: ela retorna `200 OK` com `disponivel: false`. Peça inexistente retorna `404 Not Found`.

**Códigos HTTP / Erros**

| Código | Situação |
|---|---|
| `200 OK` | Uma ou mais peças encontradas, inclusive sem saldo suficiente. |
| `400 Bad Request` | Parâmetro inválido, quantidade inválida, paginação inválida ou filtro de busca ausente. |
| `401 Unauthorized` | Token ausente ou expirado. |
| `403 Forbidden` | Usuário sem o escopo `estoque:ler`. |
| `404 Not Found` | Nenhuma peça corresponde aos parâmetros informados. |
| `500 Internal Server Error` | Falha inesperada. |

**Dependências**

- `ItemEstoqueRepository`.
- `PedidoCompraRepository`.
- Módulo de autenticação JWT.
- Módulo de autorização.
- Caso de uso Registrar Peça Necessária.
- Caso de uso Solicitar Compra de Peças, quando aplicável.
- Banco de dados.

**Testes**

- Deve calcular `saldoDisponivel` com saldo reservado zero, parcial e total.
- Deve indicar `disponivel: true` apenas quando houver saldo disponível.
- Deve comparar saldo disponível e quantidade desejada.
- Deve indicar peça abaixo do estoque mínimo.
- Deve buscar por código exato e descrição parcial.
- Deve aplicar os filtros de categoria, fabricante, disponibilidade e peças inativas.
- Deve retornar `404 Not Found` quando não houver peça correspondente.
- Deve retornar `400 Bad Request` sem filtro de busca, com descrição curta, quantidade inválida ou paginação inválida.
- Deve retornar `401 Unauthorized` sem token e `403 Forbidden` sem escopo.
- Deve garantir que a consulta não cria reserva, não altera saldo e não cria pedido de compra.

### 11.3 Check-list de Implementação

**Domínio**

- [ ] Implementar o cálculo de `saldoDisponivel`.
- [ ] Implementar as flags `disponivel`, `quantidadeDisponivel` e `abaixoDoMinimo`.
- [ ] Garantir que a disponibilidade seja calculada pelo saldo disponível, e não pelo saldo físico isolado.

**Caso de Uso e Repositório**

- [ ] Implementar o caso de uso `ConsultarPecas`.
- [ ] Implementar filtros de código, descrição, categoria, fabricante, disponibilidade e peças ativas.
- [ ] Implementar paginação.
- [ ] Implementar a comparação com quantidade desejada.
- [ ] Implementar `ItemEstoqueRepository.buscarPorFiltro`.
- [ ] Implementar busca por código exato e descrição parcial.
- [ ] Consultar pedidos de compra em aberto.
- [ ] Garantir que a consulta não execute alterações no banco.

**API e Segurança**

- [ ] Criar handler para `GET /estoque/pecas`.
- [ ] Implementar recebimento e validação dos query params.
- [ ] Implementar resposta paginada.
- [ ] Retornar `404` quando nenhuma peça for encontrada.
- [ ] Aplicar autenticação JWT, perfis permitidos e escopo `estoque:ler`.
- [ ] Mapear erros para `400`, `401`, `403`, `404` e `500`.
- [ ] Documentar o endpoint no Swagger/OpenAPI.

**Testes e Qualidade**

- [ ] Criar testes unitários para saldo disponível, disponibilidade, quantidade desejada, estoque mínimo e filtros inválidos.
- [ ] Criar testes de integração para busca por código, descrição, filtros, peça sem saldo, quantidade insuficiente e peça inexistente.
- [ ] Criar testes de autenticação, autorização e paginação.
- [ ] Criar testes de regressão para garantir que a consulta não altera o estoque, cria reserva ou cria pedido de compra.
- [ ] Revisar nomes conforme a Linguagem Ubíqua do projeto.
- [ ] Executar testes automatizados, realizar code review e validar critérios de aceite da task.
