# Manifestos por ambiente

```text
k8s/
  base/
    api/                 Deployment, Service e HPA compartilhados
    config/              Valores comuns, usados localmente e como modelo pelo Terraform AWS
    migrations/          Job Flyway compartilhado
  local/
    database/            PostgreSQL e armazenamento do Kind (gerenciados pelo Terraform)
    metrics-server/      Instalacao com excecao TLS exclusiva do Kind
    kustomization.yaml   API + ConfigMap local
  aws/
    kustomization.yaml   API com imagem ECR, sem banco ou ConfigMap local
    verificar-rds.yaml   Diagnostico opcional da conexao ao RDS
  tests/                 Carga k6 opcional, nao executada pelos overlays
```

Kustomize combina a base com os ajustes de cada ambiente e acompanha o kubectl.
Mover arquivos nao altera o nome ou namespace dos recursos existentes.

## Local

Primeiro provisionar namespace, Secret e PostgreSQL com `infra/local` e preparar
as tabelas. Construir/carregar a imagem no Kind. Depois:

```bash
kubectl kustomize k8s/local
kubectl --context kind-oficina apply -k k8s/local
```

Metrics Server e instalado separadamente com
`kubectl --context kind-oficina apply -k k8s/local/metrics-server`.
O overlay nao inclui banco, namespace ou Secret, para nao disputar gestao com
Terraform. `namespace.yaml` e `secret.yaml.example` sao referencias para uso manual
sem Terraform; nao aplicar exemplos com marcadores sobre credenciais reais.

## AWS

`infra/aws` cria EKS, RDS, namespace, Secret e ConfigMap com endpoint RDS e TLS.
Nao existe PostgreSQL em pod nesse ambiente.

A API usa Service LoadBalancer com NLB publico pelo controlador integrado do EKS
tradicional. Requer sub-redes publicas descobertas pelo controlador (tag
`kubernetes.io/role/elb=1`) e permissoes AWS para provisionar o balanceador.
O balanceador tem custo enquanto existir. O ambiente local permanece ClusterIP.
Esta configuracao inicial usa HTTP na porta 80; nao enviar senhas ou tokens por
essa conexao sem configurar HTTPS. Para consultar o endereco apos o deploy:

```bash
kubectl -n oficina get service oficina-api -w
# Testar http://<EXTERNAL-IP>/health e /ready.
# Se permanecer Pending: kubectl -n oficina describe service oficina-api
```

Antes do deploy, substituir `DEFINIR_SHA_DO_COMMIT` em `aws/kustomization.yaml`
pelo SHA completo publicado no ECR. Conferir conta e regiao do registry. O script
da pipeline preenche a imagem a partir do mesmo commit aprovado pelo CI.

```bash
kubectl kustomize k8s/aws
# Somente depois de selecionar imagem, executar migrations e preparar metricas:
kubectl --context oficina-aws apply -k k8s/aws
```

Esse overlay NAO instala Metrics Server. O Terraform cluster o prepara como add-on
EKS; nao aplicar a excecao `--kubelet-insecure-tls` do Kind na AWS.
Diagnostico RDS: `kubectl --context oficina-aws apply -f k8s/aws/verificar-rds.yaml`.

Nao aplicar recursivamente toda a pasta k8s: isso mistura ambientes, exemplos e
tarefas opcionais. `kubectl kustomize` apenas renderiza; nao altera o cluster.

Migrations sao renderizadas separadamente em `k8s/local/migrations` e
`k8s/aws/migrations`. AWS adiciona verificacao TLS do RDS. O script deploy-eks.sh
atualiza imagens e sufixo do Job no workspace do runner, executa migrations antes
da API e preserva namespace, Secret e ConfigMap gerenciados pelo Terraform.

Teste estatico dos overlays, sem acesso a clusters:

```powershell
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/test-k8s-overlays.ps1
```
