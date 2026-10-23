---
documento: Clean Code
dono: Hemily Nara
versao: 1.0
atualizado_em: 2026-10-23
status: implementado
---

# Clean Code

A implementação aplica nomes claros, simplicidade e coesão à apresentação HTTP,
à inicialização da API e à organização da persistência de ordens de serviço.
As alterações preservam as regras de negócio, os contratos HTTP e as transações
existentes.

## 1. Nomes que expressam a intenção

Nos handlers de veículo, os nomes identificam diretamente os dados e as operações
representados pelo código:

| Identificador anterior | Identificador implementado | Significado |
| --- | --- | --- |
| `pl` | `placa` | Placa normalizada para consulta |
| `v` | `veiculo` | Veículo retornado pelo caso de uso |
| `incluir` | `incluirInativos` | Filtro de inclusão de veículos inativos |
| `result` | `resultado` | Resultado da operação executada |
| `NewHandler` | `NewCadastrarHandler` | Construção do handler de cadastro de veículo |

Na composição da aplicação, os casos de uso de veículo também possuem nomes
explícitos: `cadastrarVeiculo`, `consultarVeiculo`, `atualizarVeiculo`,
`inativarVeiculo` e `reativarVeiculo`.

Esses identificadores permitem entender a intenção da operação sem depender de
comentários explicativos. Os chamadores e testes utilizam o construtor
`NewCadastrarHandler`.

Referência: [handlers de veículo](../internal/presentation/veiculo/handler.go).

## 2. Constantes com significado

Os handlers de veículo utilizam as constantes de `net/http` para representar os
status das respostas:

| Código HTTP | Constante utilizada |
| --- | --- |
| `201` | `http.StatusCreated` |
| `400` | `http.StatusBadRequest` |
| `404` | `http.StatusNotFound` |
| `409` | `http.StatusConflict` |
| `412` | `http.StatusPreconditionFailed` |
| `428` | `http.StatusPreconditionRequired` |
| `500` | `http.StatusInternalServerError` |

A leitura do código informa o significado da resposta. Os valores enviados ao
consumidor da API permanecem os mesmos.

## 3. Inicialização e rotas com responsabilidades definidas

O [main.go](../cmd/oficina-mecanica/main.go) concentra a inicialização dos recursos,
a configuração do servidor HTTP e o gerenciamento do seu ciclo de vida.

A função `montarAplicacao`, em [routes.go](../cmd/oficina-mecanica/routes.go),
compõe as dependências e o handler HTTP da aplicação. O registro das rotas é
organizado em funções por contexto:

- `registrarRotasSeguranca`;
- `registrarRotasMecanicos`;
- `registrarRotasClientes`;
- `registrarRotasVeiculos`;
- `registrarRotasFornecedores`;
- `registrarRotasServicos`;
- `registrarRotasOrdensServico`;
- `registrarRotasOrcamentos`;
- `registrarRotasEstoque`.

Cada função reúne o registro de rotas e a composição dos respectivos casos de uso.
O repositório de veículo é instanciado uma vez nessa composição e compartilhado
entre os casos de uso do contexto.

Essa organização reduziu o `main.go` de 304 para 120 linhas e tornou explícita a
responsabilidade de cada função. Os 56 registros de rotas de negócio e os dois
endpoints de saúde mantêm seus handlers e controles de acesso.

## 4. Coesão na persistência de ordens de serviço

Os métodos do repositório PostgreSQL de OS foram agrupados em arquivos conforme
a responsabilidade:

| Arquivo | Responsabilidade |
| --- | --- |
| [postgres_repository.go](../internal/infrastructure/ordemservico/postgres_repository.go) | Tipo do repositório e construtor |
| [postgres_consultas.go](../internal/infrastructure/ordemservico/postgres_consultas.go) | Consultas, listagens, fila e tempos de execução |
| [postgres_registro.go](../internal/infrastructure/ordemservico/postgres_registro.go) | Criação de OS e registro de problemas, serviços e itens |
| [postgres_execucao.go](../internal/infrastructure/ordemservico/postgres_execucao.go) | Início da execução, baixa de reservas, finalização e entrega |
| [postgres_devolucao.go](../internal/infrastructure/ordemservico/postgres_devolucao.go) | Devolução de itens, liberação de reservas e desvinculação de pedidos |

Os arquivos permanecem no mesmo pacote `ordemservico` e utilizam o mesmo tipo
`PostgresRepository`. A organização preserva os métodos, as consultas SQL e os
limites das transações, facilitando a localização das operações.

## 5. Testes e preservação do comportamento

Os testes acompanham a organização implementada:

- O teste de duplicidade de rotas verifica os registros em `routes.go`.
- Os testes do roteador real verificam ausência de token, token inválido e falta
  de permissão nas 55 rotas protegidas: 165 verificações por execução.
- Saúde e CORS são verificados através do handler montado pela aplicação.
- Os testes de veículo utilizam `NewCadastrarHandler`.
- Nos testes de integração de OS, `t.Cleanup(db.Close)` mantém o pool disponível
  até a conclusão da limpeza dos registros.

### Validações executadas

| Validação | Resultado |
| --- | --- |
| Suíte completa, sem cache, com detector de concorrência | 62 pacotes aprovados |
| Cobertura das camadas internas | 84,4% |
| Cinco repetições dos pacotes críticos com ordem aleatória | Aprovadas |
| Análise estática com `go vet` | Aprovada |
| Comparação de 30 respostas HTTP entre as versões anterior e refatorada | Equivalentes, normalizando IDs, horários e conteúdo do JWT gerados |
| Vinte cadastros simultâneos com o mesmo CPF, por versão | Um cadastro criado; demais requisições retornaram conflito |
| Vinte atualizações simultâneas com o mesmo `If-Match`, por versão | Uma atualização aplicada; demais requisições recusadas por versão divergente |
| Lotes de serviços e peças com item inexistente | Rollback sem gravação parcial |

Os testes de escrita utilizaram bancos PostgreSQL isolados com as migrations e
o seed do projeto. As consultas realizadas no banco original preservaram os
registros existentes.

Para executar a validação automatizada com um PostgreSQL de testes configurado
em `DATABASE_URL`, com migrations e seed aplicados:

```bash
go test -count=1 -race ./...
go vet ./...
go test -race -shuffle=on -count=5 ./cmd/oficina-mecanica ./internal/presentation/veiculo ./tests/integration/veiculo ./tests/integration/ordemservico ./tests/integration/notificacao
```
