# Testes de desempenho na pipeline

O workflow `1.2 · ⏱️ API performance` executa o Grafana k6 contra uma API e um
PostgreSQL temporários. A imagem da API é construída a partir do código da execução.
O teste aplica as migrations e o seed em um banco vazio, aguarda `/ready` e inicia
a carga. Não utiliza Secrets do GitHub, banco compartilhado, ECR ou cluster AWS.

## Quando executa

| Disparo | Perfil | Usuários virtuais | Duração da carga |
| --- | --- | ---: | --- |
| PR com alterações na aplicação, migrations ou arquivos do teste | `smoke` | 3 | 30 segundos |
| Push na `main` nos mesmos caminhos | `smoke` | 3 | 30 segundos |
| Manual, em Actions → API performance → Run workflow | `load` (padrão) ou `smoke` | 20 ou 3 | 120 ou 30 segundos |

Os caminhos estão declarados em `.github/workflows/performance.yml`. Alterações
somente de documentação não iniciam esse teste. Não configurar este check como
obrigatório indiscriminadamente para todas as PRs, pois os filtros podem impedir
sua execução. A escolha de checks obrigatórios é feita nas regras do repositório.

A duração acima não inclui preparação das imagens, banco e dados, nem a espera
final de até 15 segundos para concluir iterações em andamento. Cada usuário
virtual executa seu fluxo sequencialmente e espera 0,5 segundo entre iterações;
a quantidade de requisições é medida pelo k6 e não é fixa.

## Fluxo exercitado

A preparação autentica o mecânico do seed e cria um cliente e veículo próprios
para cada usuário virtual. Durante a carga, cada usuário executa repetidamente:

1. `POST /autenticacao/login`: validar HTTP 200 e token recebido.
2. `POST /ordens-servico`: abrir uma OS com seu cliente/veículo, um serviço e uma
   peça do catálogo do seed. Validar HTTP 201, identificador e status `RECEBIDA`.
3. `GET /ordens-servico/{osId}`: consultar a OS recém-criada e validar identificação,
   status e orçamento contendo serviço e peça.

As credenciais do mecânico são exclusivamente as de demonstração do seed local.
As senhas de PostgreSQL e JWT são geradas a cada execução, não são versionadas e
não são publicadas nos relatórios. Os serviços não publicam portas do computador.
O destino do teste é fixo (`http://api:8080` na rede temporária do Compose).

## Critérios de aprovação

| Medida | Critério |
| --- | --- |
| Verificações funcionais, incluindo preparação | 100% aprovadas |
| Falhas HTTP, incluindo preparação | Menos de 1% |
| Tempo de resposta p95 de cada uma das três operações durante a carga | Menor que 2.000 ms |
| Requisições de cada operação durante a carga | Pelo menos uma |

P95 indica o tempo abaixo do qual ficaram 95% das respostas daquela operação.
Os limites ficam em `tests/performance/api.js`. O orçamento inicial de 2 segundos
é um critério de regressão para runner compartilhado, não uma promessa de desempenho
em produção. Deve ser acompanhado pelas medições da CI; comparar execuções com o
mesmo perfil e considerar a variação do runner. Não foi definido a partir dos
números do teste de login realizado no Kubernetes local, que tem outro ambiente
e outra composição de requisições.

O k6 retorna erro quando algum threshold não é atendido. A pipeline preserva esse
erro e reprova o job. Falhas no build, migrations, readiness ou preparação dos
clientes também reprovam a execução. Não há `continue-on-error`.

## Resultados

As etapas do workflow são identificadas por emojis. O resumo exibe ✅ aprovado ou
❌ reprovado para o teste e para cada operação, considerando latência, amostras e
verificações funcionais. Falha antes da geração de métricas aparece no resumo
com indicação para consultar os logs.

Na página da execução, o resumo apresenta perfil, critérios, quantidade de
requisições e p95 por operação. Os artifacts `performance-<run>-<attempt>` ficam
disponíveis por 14 dias e contêm:

- `summary.json`: métricas agregadas e resultado dos thresholds.
- `summary.md`: resumo legível do k6.
- `api.log`, `postgres.log` e `migrate.log`: diagnóstico do ambiente temporário.

Se a preparação falhar antes do k6, pode não existir resumo de métricas; consultar
os logs. A publicação do diagnóstico roda mesmo quando o job falha. Não são
exportados tokens, respostas individuais nem dump do banco. Ao terminar, o script
remove containers, rede e volumes do seu próprio projeto Compose. Cancelamento
forçado pode impedir o trap; o runner hospedado é descartado pelo GitHub.

## Executar no computador

Pré-requisitos: Docker com Compose v2 compatível com `up --wait`, Bash e OpenSSL.
Da raiz do projeto, sem precisar iniciar o Compose principal:

```bash
bash scripts/test-performance.sh smoke
bash scripts/test-performance.sh load
```

Os resultados ficam em `reports/performance/`, ignorado pelo Git. Cada execução
substitui os relatórios anteriores; não executar dois testes simultaneamente na
mesma pasta de checkout. Os nomes dos projetos Compose são exclusivos, e a limpeza
não remove containers ou dados do ambiente principal.

## Relação com Kubernetes e HPA

Este workflow mede os fluxos HTTP com uma instância da API. Não instala Kubernetes,
não escala pods e não valida alta disponibilidade nem capacidade de produção.
O teste `k8s/tests/carga-k6.yaml` continua disponível para observar o HPA em um
cluster, conforme [documentação Kubernetes](05-kubernetes.md).

Referências: [thresholds do k6](https://grafana.com/docs/k6/latest/using-k6/thresholds/)
e [resumos personalizados](https://grafana.com/docs/k6/latest/results-output/end-of-test/custom-summary/).
