---
documento: Kubernetes
dono: Hemily Nara
versao: 0.4
atualizado_em: 2026-10-23
status: em revisao
---

# Kubernetes

A API possui manifestos Kubernetes para criação do namespace, gerenciamento dos
pods e acesso interno pelo Service. A aplicação também disponibiliza um endpoint
de prontidão que verifica a conexão com o PostgreSQL.

## 1. Recursos implementados

| Arquivo | Recurso | Nome | Namespace |
| --- | --- | --- | --- |
| `k8s/namespace.yaml` | Namespace | `oficina` | — |
| `k8s/deployment.yaml` | Deployment | `oficina-api` | `oficina` |
| `k8s/service.yaml` | Service | `oficina-api` | `oficina` |
| `k8s/configmap.yaml` | ConfigMap | `oficina-api-config` | `oficina` |
| `k8s/secret.yaml.example` | Secret (exemplo) | `oficina-api-secret` | `oficina` |

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

## 4. ConfigMaps e Secrets

As variáveis são separadas pela sensibilidade: o ConfigMap guarda o que pode ser
versionado e o Secret guarda credenciais. Os nomes das chaves são os lidos pelo
código; renomear uma chave no manifesto não altera o nome lido pela aplicação.
Não repetir a mesma chave nos dois recursos.

### ConfigMap `oficina-api-config`

Versionado em `k8s/configmap.yaml`. Todos os valores são strings.

| Variável | Finalidade | Valor versionado |
| --- | --- | --- |
| `DB_HOST` | Host do PostgreSQL visto pelo pod | `postgres` |
| `DB_PORT` | Porta do PostgreSQL | `5432` |
| `DB_NAME` | Banco utilizado | `oficina` |
| `SMTP_HOST` | Servidor SMTP; vazio desliga o canal | vazio |
| `SMTP_PORT` | Porta SMTP | `1025` |
| `SMTP_USUARIO` | Usuário SMTP, quando exigido | vazio |
| `EMAIL_REMETENTE` | Remetente das notificações | `Oficina Mecanica <onboarding@resend.dev>` |
| `NOTIFICACAO_INTERVALO_SEGUNDOS` | Intervalo da fila de notificações | `30` |
| `NOTIFICACAO_LOTE` | Notificações por rodada | `20` |
| `CAPACIDADE_DIARIA_OS` | Capacidade usada na estimativa de entrega | `0` |

`DB_HOST` deve apontar para o Service ou endpoint do banco no ambiente; `postgres`
é o nome de referência. `NOTIFICACAO_INTERVALO_SEGUNDOS` precisa ser positivo:
com `0` a API encerra ao iniciar o processamento da fila. `CAPACIDADE_DIARIA_OS`
igual a `0` deixa a fila fora da estimativa de entrega.

A porta HTTP é fixa em `:8080` no código e a aplicação não possui nível de log
configurável; por isso essas configurações não fazem parte do ConfigMap.

### Secret `oficina-api-secret`

Somente o exemplo `k8s/secret.yaml.example` é versionado, com marcadores no lugar
das credenciais. A extensão `.example` impede que `kubectl apply -f k8s/` aplique
os marcadores sobre o Secret real. O arquivo preenchido `k8s/secret.yaml` está
no `.gitignore` e no `.dockerignore`.

| Variável | Obrigatória | Finalidade |
| --- | --- | --- |
| `DB_USER` | Sim, sem `DATABASE_URL` | Usuário do PostgreSQL |
| `DB_PASSWORD` | Sim, sem `DATABASE_URL` | Senha do PostgreSQL |
| `JWT_SECRET` | Sim | Chave de assinatura dos tokens; igual entre réplicas |
| `SMTP_SENHA` | Não | Senha SMTP, quando exigida pelo servidor |
| `RESEND_API_KEY` | Não | Token do Resend; quando preenchido, tem prioridade sobre o SMTP |
| `DATABASE_URL` | Não | Conexão completa; quando preenchida, substitui as variáveis `DB_*` |

A URL de conexão é montada com `DB_*` sem codificação e com `sslmode=disable`.
Para senha com caracteres reservados de URL (`@ : / ? #`) ou banco que exige TLS,
utilizar `DATABASE_URL` com a senha codificada. Uma URL com senha não deve ficar
no ConfigMap. `SONAR_TOKEN` não é usado pela API e não entra no Secret.

O canal de notificação segue esta ordem: `RESEND_API_KEY` preenchida usa o
Resend; sem ela, `SMTP_HOST` preenchido usa SMTP; sem os dois, as notificações
vão apenas para o log.

### Criação do Secret

Em ambiente local, copiar o exemplo, preencher os marcadores e aplicar:

```bash
cp k8s/secret.yaml.example k8s/secret.yaml
kubectl apply -f k8s/secret.yaml
```

Na pipeline, o Secret é gerado a partir dos secrets do repositório, sem arquivo
em disco:

```bash
kubectl -n oficina create secret generic oficina-api-secret \
  --from-literal=DB_USER="$DB_USER" \
  --from-literal=DB_PASSWORD="$DB_PASSWORD" \
  --from-literal=JWT_SECRET="$JWT_SECRET" \
  --from-literal=RESEND_API_KEY="$RESEND_API_KEY" \
  --dry-run=client -o yaml | kubectl apply -f -
```

O `--dry-run=client` com `apply` cria ou atualiza o recurso no mesmo comando.
Base64 não é criptografia: a proteção do Secret depende das permissões do cluster.

### Atualização da configuração

As variáveis são injetadas por `envFrom` na criação do container. Pods em execução
não recebem novos valores; depois de alterar o ConfigMap ou o Secret, recriar os pods:

```bash
kubectl apply -f k8s/configmap.yaml -f k8s/secret.yaml
kubectl -n oficina rollout restart deployment/oficina-api
kubectl -n oficina rollout status deployment/oficina-api --timeout=300s
```

O `rollout undo` reverte o template dos pods, mas não restaura valores anteriores
do ConfigMap ou do Secret. Alterar `JWT_SECRET` invalida os tokens já emitidos.
Se um valor novo impedir a conexão com o banco, o pod novo não fica pronto e,
com `maxUnavailable: 0`, os pods anteriores continuam atendendo.

## 5. Horizontal Pod Autoscaler (HPA)

<!-- Acrescentar a documentação do recurso quando o manifesto for incorporado. -->

## 6. Execução

### Pré-requisitos

- Docker e cluster Kubernetes disponíveis;
- `kubectl` configurado para o cluster correto;
- imagem acessível pelo runtime do cluster;
- `k8s/secret.yaml` preenchido a partir do exemplo (seção 4);
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
kubectl apply -f k8s/configmap.yaml -f k8s/secret.yaml
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
- ConfigMap e Secret aplicados em cluster Kind com o Deployment: API pronta,
  `/ready` com o banco, login e consulta autenticada com o `JWT_SECRET` do Secret.
- `SMTP_HOST` alterado no ConfigMap: o pod em execução manteve o valor anterior e,
  após `rollout restart`, a notificação de diagnóstico chegou ao Mailpit.
- Senha incorreta no Secret: o pod novo não ficou pronto e o anterior continuou
  atendendo.
- Nenhum valor do Secret encontrado nos logs da API, nos eventos do namespace ou
  no `describe` do Deployment.
