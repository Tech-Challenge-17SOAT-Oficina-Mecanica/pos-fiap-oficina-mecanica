# Terraform, PostgreSQL e HPA locais

Para a alternativa AWS Academy com EKS + RDS, consulte [infra/aws](../aws/README.md).
Os comandos abaixo continuam exclusivos do ambiente local Kind.

## Organizacao local

As pastas antigas `infra/cluster` e `infra/platform` foram movidas para
`infra/local/cluster` e `infra/local/platform`, incluindo estados, variaveis e
kubeconfig existentes. Os arquivos sensiveis auxiliares ficam em
`infra/local/.local`, ignorados pelo Git.

Depois da reorganizacao, gere novos planos; nao reutilize planos salvos antes da
mudanca de caminho. Nao e necessario recriar os recursos por causa dessa mudanca.
O provider Kind marca kubeconfig_path como ForceNew; ignoramos alteracoes nesse
atributo para preservar clusters existentes. O arquivo e o output de acesso usam
o novo caminho. Estados antigos podem continuar registrando o caminho anterior
internamente; nao editar o estado manualmente para ajustar esse campo.

## Preparacao automatica no Windows (recomendado)

Com Docker Desktop em modo Linux, Terraform e kubectl instalados, execute da raiz
do repositorio, pelo Git Bash ou PowerShell:

```bash
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/provisionar-infra.ps1
```

O comando prepara armazenamento, inicializa/valida o Terraform, aplica primeiro
`cluster` e depois `platform`, aguarda PostgreSQL e mostra pods/PVC. Nao precisa
criar a pasta ou digitar credenciais manualmente em uma instalacao nova.

- **Armazenamento:** preserva `data_path` de `infra/local/cluster/terraform.tfvars` ou
  do estado existente. Em instalacao nova, usa `%LOCALAPPDATA%/oficina-mecanica/postgres`,
  fora do OneDrive, e cria o diretorio automaticamente. Continua sendo pasta
  persistente montada no Kind; nao houve troca para volume Docker nem migracao de dados.
- **Credenciais:** reutiliza `infra/local/platform/terraform.tfvars` ou `.tfvars.json`
  existentes ou salva valores aleatorios em `infra/local/.local/platform.tfvars.json`.
  Se o estado local de platform contem as credenciais anteriores, recupera esses
  valores. Reexecucoes usam o mesmo arquivo. Variaveis `TF_VAR_db_password` e
  `TF_VAR_jwt_secret` podem fornecer os valores iniciais.
- **Protecao de dados:** com dados existentes e sem credenciais recuperaveis,
  exige os valores anteriores. Alteracao do caminho de um cluster existente,
  planos com exclusao/substituicao e troca de senha do banco sao bloqueados.
- **Execucao:** planos permitidos sao aplicados pelo script. Reutiliza o ambiente
  existente; nao e necessario destrui-lo para executar novamente.
- **Segredos:** `infra/local/.local/` guarda credenciais e planos sensiveis, ignorados
  pelo Git e pelo build Docker. Preserve e proteja esses arquivos e o estado.
  Nao os envie como artifacts ou anexos.

Em instalacao nova, opcionalmente escolha um caminho absoluto com `-DataPath`.
O script nao migra um cluster existente para outro diretorio.

Resultado esperado: `postgres-0` pronto e `postgres-data` Bound. **Migrations,
API e HPA continuam nas etapas seguintes:** o comando prepara a infraestrutura
e o servidor PostgreSQL, sem criar tabelas da aplicacao nem executar deploy da API.

Testes isolados do automatizador, sem criar infraestrutura:

```bash
powershell.exe -NoProfile -ExecutionPolicy Bypass -File scripts/test-provisionar-infra.ps1
```

Se Terraform concluir mas o pod nao ficar pronto, o comando falha e preserva os
recursos para diagnostico. Consulte eventos do namespace; `FailedCreatePodSandBox`
indica um problema na criacao do ambiente do pod pelo runtime do no. Nao apague
volumes ou troque credenciais para tentar corrigir um erro de runtime.

Os comandos manuais abaixo sao alternativa didatica. Se usou o automatizador,
nao gere novas credenciais: continue na preparacao do schema.

## Arquitetura e responsabilidades

`cluster/` cria Kind `oficina` com Kubernetes 1.32.2 e um no. Um diretorio do
host e montado no no para persistir dados mesmo apos recriar o cluster.
`platform/` cria namespace, Secret compartilhado pela API e banco, Service
`postgres:5432`, StatefulSet PostgreSQL 16 e PV/PVC. Os YAML ficam em
`k8s/local/database/`, mas **somente Terraform os gerencia**.

API e HPA continuam gerenciados por kubectl. Metrics Server e instalado pelo
Kustomize em `k8s/local/metrics-server`. Nao ha dependencias de conta de nuvem.
Os estados sao locais e separados: primeiro cluster, depois platform; destrua
na ordem inversa. Nao aplique namespace, Secret ou banco novamente com kubectl.

## Pre-requisitos

- Docker Desktop iniciado em modo Linux; Terraform >= 1.6 e < 2.
- kubectl compativel com Kubernetes 1.32; CLI Kind para carregar a imagem da API.
- Internet para providers, imagens e manifesto do Metrics Server.
- Recursos para PostgreSQL, Metrics Server e ate cinco replicas da API,
  mais uma durante RollingUpdate. Um no nao oferece alta disponibilidade.
- Diretorio exclusivo de dados compartilhado com Docker, fora do OneDrive.
  Se o bind mount Windows apresentar problemas de permissao do PostgreSQL,
  execute todo o fluxo em WSL2, usando diretorio no filesystem Linux.

Comandos PowerShell abaixo partem da raiz do repositorio. Em WSL/Linux adapte
atribuicoes para `export` e caminhos para o filesystem Linux; nao misture os ambientes.

## Provisionar cluster

Crie um diretorio vazio para este ambiente e ajuste `data_path` no tfvars.

```powershell
New-Item -ItemType Directory -Force C:/oficina-data/postgres
Copy-Item infra/local/cluster/terraform.tfvars.example infra/local/cluster/terraform.tfvars
terraform -chdir=infra/local/cluster init
terraform -chdir=infra/local/cluster validate
terraform -chdir=infra/local/cluster test
terraform -chdir=infra/local/cluster plan -out=cluster.tfplan
terraform -chdir=infra/local/cluster apply cluster.tfplan
$env:KUBECONFIG = (Resolve-Path infra/local/cluster/kubeconfig).Path
kubectl --context kind-oficina get nodes
```

O kubeconfig e separado. Revise o plano: alterar a configuracao do Kind pode
substituir o cluster. Os dados do host nao sao removidos pelo Terraform.

## Provisionar banco

Se usou `provisionar-infra.ps1`, o banco ja foi provisionado. Para plan manual
posterior, forneca `-var-file=../.local/platform.tfvars.json` caso tenha sido esse
o arquivo gerado. Nao gere outra senha nem sobrescreva esse arquivo.

Gere credenciais apenas na primeira execucao e guarde-as com seguranca. Nas
execucoes seguintes forneca os mesmos valores pelas variaveis de ambiente.

```powershell
$env:TF_VAR_db_password = [guid]::NewGuid().ToString('N')
$env:TF_VAR_jwt_secret = [guid]::NewGuid().ToString('N') + [guid]::NewGuid().ToString('N')
terraform -chdir=infra/local/platform init
terraform -chdir=infra/local/platform validate
terraform -chdir=infra/local/platform plan -out=platform.tfplan
terraform -chdir=infra/local/platform apply platform.tfplan
kubectl --context kind-oficina -n oficina rollout status statefulset/postgres --timeout=300s
kubectl --context kind-oficina -n oficina get pvc
```

O PVC deve ficar `Bound`. Banco e usuario sao `oficina`, em acordo com a API.
O contrato Terraform aceita senhas alfanumericas, `_` e `-`, com pelo menos 16
caracteres. O Secret ja inclui DB_USER, DB_PASSWORD e JWT_SECRET:
nao aplique o exemplo `k8s/local/secret.yaml` sobre ele.

**Estado e segredos:** `sensitive` nao criptografa o tfstate. Estados, backups,
planos e kubeconfig podem conter credenciais/certificados: nao publique em Git
nem como artifact do CI; proteja e mantenha backup local. `infra` e excluido do
contexto Docker. Alterar o Secret nao altera a senha de um PostgreSQL ja
inicializado; rotacao exige alteracao coordenada no banco e Secret.

## Inicializar schema (somente banco novo)

Terraform provisiona o servidor. O mecanismo SQL existente nao tem historico de
migrations; nao reexecute esta inicializacao sobre banco persistido. Em evolucoes,
aplique apenas migrations novas. Automacao de historico/checksums cabe ao fluxo
de integracao da aplicacao; nao e criada por este provisionamento.

```powershell
kubectl --context kind-oficina -n oficina cp db postgres-0:/tmp/oficina-db
if ($LASTEXITCODE -ne 0) { throw 'Falha ao copiar SQL' }
kubectl --context kind-oficina -n oficina exec postgres-0 -- sh -ec 'for f in /tmp/oficina-db/migrations/*.sql; do psql -v ON_ERROR_STOP=1 -U oficina -d oficina -f "$f"; done; psql -v ON_ERROR_STOP=1 -U oficina -d oficina -f /tmp/oficina-db/seeds/V900__dados_iniciais.sql'
if ($LASTEXITCODE -ne 0) { throw 'Falha na inicializacao do schema' }
```

O seed inclui usuario demonstrativo conhecido: use apenas neste ambiente local.

## Instalar Metrics Server e disponibilizar API/HPA

```powershell
kubectl --context kind-oficina apply -k k8s/local/metrics-server
kubectl --context kind-oficina -n kube-system rollout status deployment/metrics-server --timeout=180s
kubectl --context kind-oficina wait --for=condition=Available apiservice/v1beta1.metrics.k8s.io --timeout=180s
docker build -t oficina-api:local .
kind load docker-image oficina-api:local --name oficina
kubectl --context kind-oficina apply -k k8s/local
kubectl --context kind-oficina -n oficina rollout status deployment/oficina-api --timeout=300s
kubectl --context kind-oficina top pods -n oficina
kubectl --context kind-oficina -n oficina get hpa
kubectl --context kind-oficina -n oficina port-forward service/oficina-api 8080:80
```

Em outro terminal, consulte `/health` e `/ready` com `curl.exe --fail`.
Teste tambem login/consulta autenticada: readiness testa conexao, nao schema.
No Linux utilize `curl`. O Metrics Server v0.7.2 suporta Kubernetes 1.32;
o patch `--kubelet-insecure-tls` e exclusivo do Kind local, para aceitar certificados
dos kubelets. Nao aplique essa excecao em producao.

## Verificar HPA e gerar carga

O HPA existente mantem 2 a 5 replicas, com metas CPU 70% e memoria 80% dos
requests (70m e aproximadamente 102Mi por pod). A maior recomendacao prevalece:
nao e necessario ultrapassar ambas as metas. Aguarde a coleta inicial.

```powershell
kubectl --context kind-oficina -n oficina run hpa-load --image=fortio/fortio:1.69.0 --restart=Never -- load -c 32 -qps 0 -t 180s http://oficina-api/health
kubectl --context kind-oficina -n oficina get hpa -w
# Em outro terminal:
kubectl --context kind-oficina -n oficina top pods
kubectl --context kind-oficina -n oficina get deployment oficina-api
kubectl --context kind-oficina -n oficina describe hpa oficina-api-hpa
```

Registre metricas e aumento das replicas. A carga depende do hardware; aumente
concorrencia/duracao se CPU ficar abaixo da meta e observe a capacidade do no e
do gerador. Memoria deve ter metricas disponiveis, mas este teste busca escala
por CPU. Verifique tambem o teste autenticado ja descrito em `docs/05-kubernetes.md`.
`ScalingActive=True` indica uma metrica valida; confira as duas metas sem
`<unknown>` e os eventos. Para erros, consulte os logs de Metrics Server.

```powershell
kubectl --context kind-oficina -n oficina logs hpa-load
kubectl --context kind-oficina -n oficina delete pod hpa-load
```

A reducao pode levar mais de 300s pela estabilizacao. A API possui consumidor
de notificacoes por replica: valide sua coordenacao antes de demonstrar fluxos
de negocio concorrentes; o teste `/health` nao cria notificacoes.

## Validar persistencia e remover recursos

Apos aplicar, execute `plan` em cada diretorio: nao deve haver alteracoes.
Para comprovar persistencia, insira um registro de teste, exclua somente
`postgres-0`, aguarde o StatefulSet recria-lo e consulte o registro novamente.
O volume usa `Retain`, mas nao substitui backup. HostPath nao impoe a quota
de 2Gi declarada no PV e esta configuracao suporta somente o no Kind definido.

Quando encerrar o ambiente, remova primeiro a aplicacao e depois a infraestrutura:

```powershell
kubectl --context kind-oficina delete -k k8s/local --ignore-not-found
kubectl --context kind-oficina delete -k k8s/local/metrics-server --ignore-not-found
terraform -chdir=infra/local/platform destroy
terraform -chdir=infra/local/cluster destroy
```

O diretorio de dados fica preservado. Apaga-lo exige decisao explicita e backup.
Para restaurar, use o mesmo caminho, credenciais e versao principal PostgreSQL;
nao reexecute a inicializacao do schema. Em CI, use diretorio temporario e
credenciais descartaveis na mesma maquina runner; destrua platform antes de
cluster inclusive em falha. Esta entrega nao modifica a pipeline existente.

Referencias: [Kind provider](https://registry.terraform.io/providers/tehcyx/kind/0.11.0/docs),
[Kubernetes provider](https://registry.terraform.io/providers/hashicorp/kubernetes/2.38.0/docs),
[Metrics Server](https://github.com/kubernetes-sigs/metrics-server/tree/v0.7.2).
