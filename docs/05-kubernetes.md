---
documento: Kubernetes — Configuração e execução da API
dono: Hemily Nara
versao: 0.2
atualizado_em: 2026-10-19
status: em revisao
---

# Kubernetes — Configuração e execução da API

A API possui manifestos Kubernetes para criação do namespace, gerenciamento dos
pods e acesso interno pelo Service. A aplicação também disponibiliza um endpoint
de prontidão que verifica a conexão com o PostgreSQL. Este guia reúne os parâmetros
de implantação, as variáveis da aplicação e as referências de configuração dos
recursos utilizados em conjunto.

## 1. Recursos implementados

| Arquivo | Recurso | Nome | Namespace |
| --- | --- | --- | --- |
| `k8s/namespace.yaml` | Namespace | `oficina` | — |
| `k8s/deployment.yaml` | Deployment | `oficina-api` | `oficina` |
| `k8s/service.yaml` | Service | `oficina-api` | `oficina` |

O label `app: oficina-api` identifica os pods e é utilizado pelos selectors do
Deployment e do Service. O container da aplicação se chama `api`.

## 2. Deployment

O Deployment executa a imagem `oficina-api:local`, com `imagePullPolicy: IfNotPresent`.
A API escuta na porta `8080`, identificada no manifesto pelo nome `http`.
O token da ServiceAccount não é montado no pod, pois a aplicação não precisa
acessar a API do Kubernetes.

### Configuração da aplicação

As variáveis de ambiente são carregadas por `envFrom`, a partir destes recursos
no namespace `oficina`:

- ConfigMap `oficina-api-config`;
- Secret `oficina-api-secret`.

Esses recursos precisam existir para o container iniciar. A conexão utiliza
`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER` e `DB_PASSWORD`, ou `DATABASE_URL`,
quando definida. A autenticação utiliza `JWT_SECRET`.
O endereço do banco deve ser acessível a partir do pod; `localhost` se refere
ao próprio pod.

### Recursos computacionais

| Recurso | Request | Limit |
| --- | --- | --- |
| CPU | `100m` | `500m` |
| Memória | `128Mi` | `256Mi` |

`spec.replicas` não está definido no manifesto. Na criação sem controlador de
escala, o Deployment utiliza o padrão de uma réplica. O manifesto não impõe uma
quantidade fixa a cada aplicação.

### Verificações de saúde

| Probe | Endpoint | Intervalo | Timeout | Falhas consecutivas |
| --- | --- | --- | --- | --- |
| Startup | `/health` | 2 s | 2 s | 30 |
| Liveness | `/health` | 10 s | 2 s | 3 |
| Readiness | `/ready` | 5 s | 2 s | 3 |

`GET /health` retorna HTTP `200` com `{"status":"ok"}` quando o servidor HTTP
responde. Esse endpoint não consulta o banco.

`GET /ready` executa `Ping` no PostgreSQL com timeout de um segundo:

- HTTP `200` com `{"status":"ready"}` quando a conexão está disponível;
- HTTP `503` com `{"status":"not_ready"}` quando a verificação falha.

A resposta não expõe detalhes da conexão. O contexto da requisição é respeitado.
Uma falha de readiness retira o pod do tráfego normal do Service sem acionar a
liveness por indisponibilidade do banco. O Ping verifica conectividade; não
valida o schema nem substitui a execução das migrations.

### Estratégia de atualização

O Deployment utiliza `RollingUpdate`, com:

| Parâmetro | Valor |
| --- | --- |
| `maxUnavailable` | `0` |
| `maxSurge` | `1` |
| `minReadySeconds` | `5` |
| `progressDeadlineSeconds` | `300` |
| `revisionHistoryLimit` | `3` |

A atualização permite um pod adicional e aguarda a disponibilidade das novas
réplicas antes de reduzir as anteriores. O histórico permite consultar revisões
e executar rollback. Esses parâmetros não garantem, isoladamente, a conclusão
de todas as requisições em andamento durante a terminação.

## 3. Service

O Service é do tipo `ClusterIP` e seleciona os pods com `app: oficina-api`.

| Campo | Valor |
| --- | --- |
| Protocolo | `TCP` |
| Porta do Service | `80` |
| `targetPort` | `http` — porta `8080` do container |
| DNS interno | `oficina-api.oficina.svc.cluster.local` |

O endereço do Service permanece estável quando o Deployment substitui os pods.

## 4. ConfigMap, Secret e variáveis de ambiente

O Deployment já referencia os recursos `oficina-api-config` e `oficina-api-secret`.
Os exemplos desta seção documentam como configurar esses recursos; não representam
manifestos de ConfigMap ou Secret já adicionados a `k8s/`. Ambos devem ser criados
no namespace `oficina`, com as chaves esperadas pelo código.

### Configurações não sensíveis — ConfigMap

| Variável | Finalidade | Valor de referência / comportamento |
| --- | --- | --- |
| `DB_HOST` | Host do PostgreSQL | DNS do Service do banco ou endpoint externo acessível pelo pod |
| `DB_PORT` | Porta do PostgreSQL | `"5432"` |
| `DB_NAME` | Banco utilizado | `"oficina"` |
| `SMTP_HOST` | Servidor de e-mail | `"mailpit"` quando esse Service existir; vazio desabilita SMTP |
| `SMTP_PORT` | Porta SMTP | `"1025"` para Mailpit; ajustar ao servidor escolhido |
| `SMTP_USUARIO` | Usuário de autenticação SMTP | Vazio para Mailpit local; tratar como Secret se o ambiente exigir confidencialidade |
| `EMAIL_REMETENTE` | Nome e endereço do remetente | Padrão do código: `Oficina Mecanica <onboarding@resend.dev>` |
| `NOTIFICACAO_INTERVALO_SEGUNDOS` | Intervalo de processamento da fila | `"30"`; utilizar valor positivo, pois zero não é válido para o ticker atual |
| `NOTIFICACAO_LOTE` | Quantidade por rodada de processamento | `"20"`; utilizar valor positivo |
| `CAPACIDADE_DIARIA_OS` | Capacidade usada na estimativa de entrega | `"0"` desconsidera a fila nesse cálculo; ajustar ao negócio |

Exemplo de referência para `k8s/configmap.yaml`:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: oficina-api-config
  namespace: oficina
data:
  DB_HOST: "postgres"
  DB_PORT: "5432"
  DB_NAME: "oficina"
  SMTP_HOST: "mailpit"
  SMTP_PORT: "1025"
  SMTP_USUARIO: ""
  EMAIL_REMETENTE: "Oficina Mecanica <oficina@example.com>"
  NOTIFICACAO_INTERVALO_SEGUNDOS: "30"
  NOTIFICACAO_LOTE: "20"
  CAPACIDADE_DIARIA_OS: "0"
```

`postgres` e `mailpit` são nomes de referência, não recursos criados pelos manifestos
atuais. Substituir pelos Services existentes ou pelos endpoints corretos. Todos os
valores de `data` devem ser strings. Para um teste sem e-mail externo, deixar
`SMTP_HOST` vazio e não configurar `RESEND_API_KEY`; as notificações usam o log.

O Compose não é aplicado automaticamente pelo Kubernetes: seus serviços e o arquivo
`.env` não são importados pelo Deployment. A porta HTTP está fixa em `:8080` no
código atual; definir uma variável `PORT` não altera esse comportamento.

### Credenciais — Secret

| Variável | Quando utilizar |
| --- | --- |
| `DB_USER` | Usuário PostgreSQL, ao utilizar conexão por `DB_*` |
| `DB_PASSWORD` | Senha PostgreSQL, ao utilizar conexão por `DB_*` |
| `JWT_SECRET` | Chave de assinatura e validação dos tokens; deve ser igual entre réplicas |
| `SMTP_SENHA` | Senha SMTP, quando exigida pelo servidor |
| `RESEND_API_KEY` | Token de envio pelo Resend, somente quando esse canal for utilizado |
| `DATABASE_URL` | Alternativa às variáveis `DB_*`, com a conexão completa e opções de TLS |

Não manter a mesma chave no ConfigMap e no Secret. `SONAR_TOKEN` não é uma variável
necessária ao runtime da API e não deve ser incluída nesse Secret.

Para criar ou atualizar o Secret, preparar um arquivo de variáveis **fora do
repositório**, com acesso restrito, contendo somente as credenciais necessárias.
Exemplo de estrutura, com marcadores que precisam ser substituídos:

```dotenv
DB_USER=PREENCHER_USUARIO
DB_PASSWORD=PREENCHER_SENHA
JWT_SECRET=PREENCHER_CHAVE_ALEATORIA
```

Substituir o caminho pelo arquivo real, sem publicar seu conteúdo:

```bash
kubectl -n oficina create secret generic oficina-api-secret \
  --from-env-file=/caminho-seguro/oficina-secret.env \
  --dry-run=client -o yaml | kubectl apply -f -
```

O Secret usa o tipo `Opaque`. Não versionar o arquivo com credenciais nem o YAML
real gerado. Base64 não é criptografia; o acesso ao Secret depende das permissões
do cluster e a criptografia em repouso depende de sua configuração.

### Banco e canais de notificação

A aplicação dá prioridade a `DATABASE_URL` quando ela não está vazia. Nessa opção,
configurar a URL completa, incluindo as opções de TLS exigidas pelo banco. Sem
`DATABASE_URL`, o código monta a conexão com `DB_*` e `sslmode=disable`; para um
banco externo que exige TLS, utilizar uma URL apropriada. Não colocar uma URL com
senha no ConfigMap. Caracteres especiais em usuário/senha precisam de codificação
correta na URI.

A ordem de escolha do envio é:

1. `RESEND_API_KEY` preenchida: Resend;
2. sem Resend e com `SMTP_HOST` preenchido: SMTP;
3. sem ambos: log.

Os nomes das variáveis devem ser preservados. Uma troca de nome no manifesto não
altera o nome lido pela aplicação.

### Aplicar uma alteração de configuração

Depois de criar o arquivo de ConfigMap e o Secret:

```bash
kubectl apply -f k8s/configmap.yaml
kubectl -n oficina get configmap oficina-api-config
kubectl -n oficina get secret oficina-api-secret
kubectl -n oficina rollout restart deployment/oficina-api
kubectl -n oficina rollout status deployment/oficina-api --timeout=300s
```

Como a injeção ocorre por variáveis de ambiente, pods existentes não recebem novos
valores automaticamente. Atualizar primeiro os recursos e depois recriar os pods.
O rollback do Deployment reverte o template dos pods, mas não restaura os valores
anteriores de ConfigMap ou Secret; uma reversão de configuração exige restaurar
os respectivos valores. Alterar `JWT_SECRET` invalida tokens assinados com a chave
anterior, pois o código atual utiliza uma única chave.

## 5. HPA e parâmetros de escala

O Deployment já define requests de CPU e memória e não fixa `spec.replicas`.
O HPA ainda não possui manifesto versionado nesta implementação. O exemplo abaixo
é uma referência para adicionar `k8s/hpa.yaml`, e não um recurso já entregue.

| Parâmetro | Referência inicial |
| --- | --- |
| Nome e namespace | `oficina-api-hpa` / `oficina` |
| Recurso alvo | Deployment `oficina-api` |
| API | `autoscaling/v2` |
| Mínimo / máximo | `2` / `5` réplicas |
| CPU | Média de `70%` do request |
| Memória | Média de `80%` do request |
| Janela de estabilização para redução | `300` segundos |

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: oficina-api-hpa
  namespace: oficina
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: oficina-api
  minReplicas: 2
  maxReplicas: 5
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
```

As porcentagens são relativas aos **requests**, não aos limits. Com request de
CPU `100m`, a meta de `70%` corresponde a `70m` em média por pod; com request de
memória `128Mi`, `80%` corresponde a aproximadamente `102,4Mi`. Alterar requests
muda o significado dessas metas. Com mais de uma métrica, o HPA utiliza a maior
recomendação de réplicas, respeitando o máximo configurado.

O cluster precisa fornecer a API de métricas de recursos, normalmente via Metrics
Server. A criação do cluster, por si só, não comprova que essas métricas estão disponíveis:

```bash
kubectl top nodes
kubectl -n oficina top pods
```

Após adicionar o manifesto, aplicar e acompanhar:

```bash
kubectl apply -f k8s/hpa.yaml
kubectl -n oficina get hpa oficina-api-hpa -w
kubectl -n oficina describe hpa oficina-api-hpa
```

Os valores são iniciais e precisam ser medidos sob carga. O teste de concorrência
de notificações identificou duplicidade com dois consumidores; essa condição deve
ser corrigida antes de ativar múltiplas réplicas para processar notificações.
O teste de terminação também identificou perda de uma requisição em andamento;
RollingUpdate não substitui o encerramento gracioso na aplicação.

### Onde alterar cada configuração

| Alteração | Local |
| --- | --- |
| Imagem / tag | `k8s/deployment.yaml`: `spec.template.spec.containers[].image` |
| Nome do ConfigMap / Secret | Metadados dos recursos e referências `envFrom` no Deployment |
| Namespace | `namespace.yaml`, metadados dos demais recursos e comandos `kubectl -n` |
| Host do banco e parâmetros não sensíveis | `data` do ConfigMap |
| Senhas e tokens | Secret ou mecanismo de provisionamento de segredos |
| CPU / memória | `resources.requests` e `resources.limits` do container no Deployment |
| Metas / quantidade de réplicas | `spec.metrics`, `minReplicas` e `maxReplicas` do HPA |
| Porta HTTP da aplicação | Código do servidor, porta do container e probes |
| Porta de acesso do Service | `spec.ports[].port` em `k8s/service.yaml` |
| Seleção de pods | Labels do template e selectors do Deployment e Service |
| Intervalos / limites das probes | `startupProbe`, `livenessProbe` e `readinessProbe` no Deployment |

O selector do Deployment é imutável após a criação. Não alterar seus labels de
seleção sem planejar a recriação do recurso; o label comum `app: oficina-api` deve
permanecer consistente entre Deployment, pods e Service.

## 6. Execução

### Pré-requisitos

- Docker e cluster Kubernetes disponíveis;
- `kubectl` configurado para o cluster correto;
- imagem acessível pelo runtime do cluster;
- ConfigMap e Secret referenciados pelo Deployment;
- PostgreSQL acessível, com as migrations aplicadas.

Executar os comandos a partir da raiz do repositório.

### Construção da imagem

```bash
docker build -t oficina-api:local .
```

Em clusters com runtime separado do Docker local, carregar a imagem ou utilizar
uma imagem publicada em registry acessível. Para Kind e Minikube, respectivamente:

```bash
kind load docker-image oficina-api:local --name NOME_DO_CLUSTER
minikube image load oficina-api:local
```

Executar somente o comando correspondente ao ambiente utilizado.

### Aplicação dos manifestos

```bash
kubectl config current-context
kubectl apply -f k8s/namespace.yaml
kubectl -n oficina get configmap oficina-api-config
kubectl -n oficina get secret oficina-api-secret
kubectl apply -f k8s/deployment.yaml -f k8s/service.yaml
kubectl -n oficina rollout status deployment/oficina-api --timeout=300s
```

A consulta ao Secret acima exibe metadados, sem imprimir os valores.

### Acesso local

```bash
kubectl -n oficina port-forward service/oficina-api 8080:80
```

Em outro terminal:

```bash
curl --fail http://localhost:8080/health
curl --fail http://localhost:8080/ready
```

O acesso por port-forward depende da permanência desse processo. Se o pod
selecionado for substituído, pode ser necessário executar o comando novamente.

### Atualização e rollback

Substituir `NOVA_IMAGEM:TAG` por uma imagem disponível no cluster:

```bash
kubectl -n oficina set image deployment/oficina-api api=NOVA_IMAGEM:TAG
kubectl -n oficina rollout status deployment/oficina-api --timeout=300s
kubectl -n oficina rollout history deployment/oficina-api
kubectl -n oficina rollout undo deployment/oficina-api
kubectl -n oficina rollout status deployment/oficina-api --timeout=300s
```

O rollback exige uma revisão anterior. Atualizar também a referência da imagem no
manifesto antes de reaplicá-lo, para manter a configuração declarada consistente.
Variáveis injetadas por `envFrom` são recarregadas quando os pods são recriados:

```bash
kubectl -n oficina rollout restart deployment/oficina-api
```

## 7. Verificação e diagnóstico

```bash
kubectl -n oficina get pods -l app=oficina-api
kubectl -n oficina get service oficina-api
kubectl -n oficina get endpointslices -l kubernetes.io/service-name=oficina-api
kubectl -n oficina describe deployment oficina-api
kubectl -n oficina logs deployment/oficina-api --tail=100
kubectl -n oficina get events --sort-by=.metadata.creationTimestamp
```

Para observar a substituição de um pod no ambiente de teste:

```bash
kubectl -n oficina delete pod NOME_DO_POD
kubectl -n oficina get pods -l app=oficina-api -w
```

Com apenas uma réplica, a exclusão pode causar indisponibilidade temporária até
o substituto ficar pronto.

## 8. Validações executadas

- Manifestos validados por schema e por dry-run no servidor Kubernetes.
- Testes automatizados de `/ready`: banco disponível, indisponível e contexto cancelado.
- API executada no Kubernetes local com acesso pelo DNS do Service e por port-forward.
- Login e consulta autenticada ao banco com dados de teste.
- Exclusão de pod e criação automática de substituto pronto para receber tráfego.
- Atualização para imagem distinta, identificada na resposta de saúde, e rollback
  para a imagem original.
- Amostra local de 435 chamadas curtas pelo Service durante as trocas, todas com
  resposta HTTP `200`; esse resultado não representa garantia geral de disponibilidade.


## 9. Referências de configuração

- [ConfigMaps — Kubernetes](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [Secrets — Kubernetes](https://kubernetes.io/docs/concepts/configuration/secret/)
- [Horizontal Pod Autoscaling — Kubernetes](https://kubernetes.io/docs/concepts/workloads/autoscaling/horizontal-pod-autoscale/)
- Variáveis e valores padrão do projeto: [`.env.example`](../.env.example),
  [`main.go`](../cmd/oficina-mecanica/main.go) e
  [`database.go`](../internal/shared/database/database.go).
