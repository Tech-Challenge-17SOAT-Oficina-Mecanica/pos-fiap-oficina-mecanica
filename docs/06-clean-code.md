---
documento: Clean Code
dono: Hemily Nara
versao: 1.1
atualizado_em: 2026-10-24
status: implementado
---

# Clean Code

O projeto aplica boas práticas de Clean Code para tornar a intenção do código
explícita, organizar responsabilidades e manter os fluxos de negócio fáceis de
acompanhar. A implementação reúne nomes claros, funções focadas, coesão,
reutilização de validações, contratos específicos e testes de comportamento.

## 1. Nomes que expressam a intenção

Os nomes dos identificadores descrevem os dados e as operações representados.
Nos handlers e na composição do contexto de veículos, são utilizados:

| Identificador | Intenção |
| --- | --- |
| `placa` | Placa normalizada utilizada na consulta |
| `veiculo` | Veículo retornado pelo caso de uso |
| `incluirInativos` | Filtro que determina a inclusão de veículos inativos |
| `resultado` | Resultado da operação executada |
| `NewCadastrarHandler` | Construção do handler de cadastro de veículo |
| `cadastrarVeiculo`, `consultarVeiculo`, `atualizarVeiculo` | Casos de uso correspondentes a cada operação |
| `inativarVeiculo`, `reativarVeiculo` | Alterações explícitas da situação do veículo |

Nas regras de OS, nomes como `ValidarAutorizacaoParaExecucao`,
`ValidarRecursosParaExecucao` e `ValidarOrcamentosParaEntrega` identificam
as condições de negócio verificadas.

Referências: [handlers de veículo](../internal/presentation/veiculo/handler.go),
[composição das rotas](../cmd/oficina-mecanica/routes.go) e
[regras de execução](../internal/domain/ordemservico/iniciar_execucao.go).

## 2. Funções focadas e responsabilidades explícitas

A inicialização da aplicação está organizada conforme a responsabilidade:

| Elemento | Responsabilidade |
| --- | --- |
| [main.go](../cmd/oficina-mecanica/main.go) | Configuração dos recursos e do servidor, inicialização e acionamento do ciclo de vida |
| [routes.go](../cmd/oficina-mecanica/routes.go) | Composição dos repositórios, adaptadores, casos de uso e handlers |
| [lifecycle.go](../cmd/oficina-mecanica/lifecycle.go) | Drenagem e encerramento do servidor e do processamento de notificações |
| [ready.go](../cmd/oficina-mecanica/ready.go) | Verificação da conexão com o banco para a readiness |

O registro das rotas é dividido nas funções `registrarRotasSeguranca`,
`registrarRotasMecanicos`, `registrarRotasClientes`, `registrarRotasVeiculos`,
`registrarRotasFornecedores`, `registrarRotasServicos`,
`registrarRotasOrdensServico`, `registrarRotasOrcamentos` e
`registrarRotasEstoque`.

Cada função reúne a montagem de um contexto de negócio. O repositório de veículo,
por exemplo, é instanciado nessa composição e compartilhado pelos respectivos
casos de uso. Os handlers interpretam a requisição, chamam o caso de uso e
constroem a resposta HTTP.

## 3. Coesão na organização dos arquivos

A persistência de ordens de serviço utiliza arquivos agrupados por operação,
no mesmo pacote e com o mesmo tipo `PostgresRepository`:

| Arquivo | Responsabilidade |
| --- | --- |
| [postgres_repository.go](../internal/infrastructure/ordemservico/postgres_repository.go) | Tipo do repositório e construtor |
| [postgres_consultas.go](../internal/infrastructure/ordemservico/postgres_consultas.go) | Consultas, listagens, fila e tempos de execução |
| [postgres_registro.go](../internal/infrastructure/ordemservico/postgres_registro.go) | Criação de OS e registro de problemas, serviços e itens |
| [postgres_execucao.go](../internal/infrastructure/ordemservico/postgres_execucao.go) | Execução, baixa de reservas, finalização e entrega |
| [postgres_devolucao.go](../internal/infrastructure/ordemservico/postgres_devolucao.go) | Devolução de itens, liberação de reservas e desvinculação de pedidos |

Essa divisão permite localizar uma operação pelo seu contexto e mantém juntas
as consultas e gravações pertencentes à mesma responsabilidade. Os limites das
transações e os bloqueios utilizados pelas operações são preservados.

## 4. Fluxo direto com retornos antecipados

Validações e erros encerram o fluxo assim que a condição é identificada.
Nos handlers de veículo, a validação do identificador, a leitura de `If-Match`
e a interpretação do JSON ocorrem antes da execução do caso de uso.

Nas regras de domínio, a mesma prática deixa explícita a sequência de decisões:

```go
func ValidarAutorizacaoParaExecucao(orcamentoAprovado, servicosAutorizados bool) error {
    if !orcamentoAprovado {
        return ErrOrcamentoNaoAprovado
    }
    if !servicosAutorizados {
        return ErrServicosNaoAutorizados
    }
    return nil
}
```

A leitura acompanha a ordem das regras e identifica o caminho de sucesso sem
acumular blocos condicionais aninhados. A prioridade dos erros é preservada.

Referências: [atualização de veículo](../internal/presentation/veiculo/handler.go)
e [autorização de execução](../internal/domain/ordemservico/iniciar_execucao.go).

## 5. Constantes e erros com significado

Os handlers de veículo representam os status HTTP com constantes de `net/http`,
como `http.StatusCreated`, `http.StatusBadRequest`, `http.StatusConflict`,
`http.StatusPreconditionFailed` e `http.StatusPreconditionRequired`.
O identificador informa a finalidade da resposta durante a leitura do código.

Os erros de aplicação e domínio também possuem nomes que expressam a condição:
`ErrClienteDuplicado`, `ErrVersaoDivergente`, `ErrOrcamentoNaoAprovado` e
`ErrRecursosIndisponiveis`, por exemplo. Os handlers utilizam esses erros para
selecionar as respostas correspondentes.

O helper [WriteProblem](../internal/shared/http/problem.go) centraliza a escrita
de erros no formato `application/problem+json`, com campos como `type`, `title`,
`status`, `detail` e os erros de campos quando informados. Isso mantém uma
estrutura comum de resposta nos contextos que utilizam o helper.

## 6. Reutilização de validações e rotinas compartilhadas

O pacote [shared/validation](../internal/shared/validation) reúne validações de
valores utilizadas pelos contextos: documento, e-mail, UUID, texto e decimais.
O domínio de cliente utiliza `validation.IsEmail`, compartilhando a mesma
interpretação de sintaxe de e-mail.

O pacote [shared/http](../internal/shared/http) reúne rotinas próprias do protocolo,
como escrita de problemas, paginação, respostas de listagem, leitura de
`If-Match` e CORS. Nos handlers de peças, por exemplo, `LerIfMatch` interpreta o
cabeçalho e distingue ausência da pré-condição e valor inválido.

A validação de placa permanece no domínio de veículo, em `NormalizarPlaca`,
pois representa um valor específico desse contexto. A localização dos helpers
acompanha a responsabilidade da regra que eles implementam.

Referências: [e-mail no domínio de cliente](../internal/domain/cliente/cliente.go),
[validação compartilhada de e-mail](../internal/shared/validation/email.go) e
[leitura de If-Match](../internal/shared/http/ifmatch.go).

## 7. Contratos pequenos e dependências explícitas

As interfaces descrevem as operações necessárias para cada consumidor.
A consulta de cliente depende de `BuscarPorDocumento`; o cadastro depende da
verificação de duplicidade e de `Salvar`. Cadastro e atualização de mecânico,
listagem e busca de fornecedor, alterações de situação de serviço e consultas
de tempos de execução também possuem contratos específicos.

Os construtores recebem suas dependências como argumentos. Os 56 construtores
de handlers HTTP recebem interfaces locais dos casos de uso; autenticação e
cadastro de mecânico recebem contratos de senha. A leitura da assinatura permite
identificar as dependências de cada operação.

Essa organização aplica responsabilidade única, segregação de interfaces e
inversão de dependências. Os detalhes das camadas e dos adaptadores estão em
[Clean Architecture](07-clean-architecture.md).

Referências: [contratos de cliente](../internal/application/cliente/ports.go),
[contratos de mecânico](../internal/application/mecanico/ports.go) e
[contratos HTTP de OS](../internal/presentation/ordemservico/usecases.go).

## 8. Testes legíveis e preservação do comportamento

Os testes acompanham as responsabilidades do código:

- Testes de domínio descrevem condições de negócio e os erros esperados,
  incluindo tabelas de cenários para autorização da execução de OS.
- Testes de casos de uso recebem implementações substitutas dos contratos,
  permitindo verificar validação, chamadas de persistência e propagação de erros.
- Testes do adaptador bcrypt verificam o algoritmo real e a compatibilidade
  entre a geração do hash e a verificação da senha.
- Testes de integração exercitam persistência e comportamento com PostgreSQL.
- Testes do roteador verificam autenticação e autorização nas 55 rotas
  protegidas, além de saúde e CORS.
- Testes de arquitetura verificam dependências entre camadas e os contratos
  recebidos pelos construtores HTTP.

Nos testes de integração de OS que realizam limpeza de registros,
`t.Cleanup(db.Close)` mantém o pool disponível até a conclusão dessa limpeza.

A validação da implementação registrou 65 pacotes aprovados, nenhum teste pulado,
detector de concorrência ativo e 84,4% de cobertura das camadas internas.
A comparação funcional antes/depois exercitou as 58 rotas, com respostas
equivalentes, além de duplicidade, versionamento concorrente e rollback.

Referências: [testes de autorização das rotas](../cmd/oficina-mecanica/routes_security_test.go),
[testes das regras de OS](../internal/domain/ordemservico/politicas_execucao_test.go)
e [testes de arquitetura](../tests/architecture).

Com `DATABASE_URL` apontando para um PostgreSQL de testes com migrations e seed
aplicados, a validação automatizada é executada com:

```bash
go test -count=1 -race ./...
go vet ./...
```
