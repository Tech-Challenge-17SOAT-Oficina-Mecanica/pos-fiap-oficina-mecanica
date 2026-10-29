# AWS Academy: EKS + PostgreSQL no RDS

O fluxo atual esta em [AUTOMACAO.md](AUTOMACAO.md): Secrets, escolha da conta,
preparacao, deploy, estados existentes, protecoes e acesso temporario do runner.
Merge ou PR nao cria infraestrutura nem faz deploy. Os workflows AWS sao manuais.

## Arquitetura

```text
02 · 🏗️ Prepare AWS infrastructure
  -> Bucket S3 privado e estados separados por conta
  -> Terraform cluster: EKS, um no EC2, RDS, ECR e Metrics Server
  -> Terraform platform: namespace, Secret e ConfigMap com endpoint RDS

03 · 🐳 Docker Build and Deploy
  -> exige CI aprovado no SHA selecionado
  -> publica API/Flyway no ECR da conta autenticada
  -> migrations no RDS a partir de um Job no EKS
  -> rollout da API e verificacao HTTP
```

EKS usa um no t3.medium. RDS usa PostgreSQL 16, db.t3.micro, Single-AZ, 20 GiB gp2,
criptografia, backup automatico por um dia, protecao de exclusao e snapshot final.
O banco nao e publico e so aceita porta 5432 do grupo das maquinas EKS.
Nao criamos funcoes IAM, NAT Gateway, balanceador nem PostgreSQL em pod na AWS.
Um unico no e banco Single-AZ nao oferecem alta disponibilidade.

## Pastas

- state/: preparacao inicial do bucket, com estado em bootstrap/terraform.tfstate.
- cluster/: recursos AWS, com estado remoto em cluster/terraform.tfstate.
- platform/: recursos Kubernetes, com estado remoto em platform/terraform.tfstate.
- k8s/aws: overlay da API e migrations, com imagem preenchida pelo script de deploy.
- k8s/base/config: valores compartilhados usados pelo Terraform para montar o
  ConfigMap AWS com endpoint RDS e verify-full. Nao aplicar esse arquivo direto no EKS.

O Kind e o Docker Compose continuam disponiveis. Os manifestos do PostgreSQL local
ficam em k8s/local/database e nao sao aplicados na AWS.

## Testes sem recursos reais

```bash
terraform -chdir=infra/aws/cluster init -backend=false
terraform -chdir=infra/aws/cluster validate
terraform -chdir=infra/aws/cluster test
terraform -chdir=infra/aws/platform init -backend=false
terraform -chdir=infra/aws/platform validate
bash scripts/test-aws-automation.sh
```

Os testes usam mocks. Nao comprovam as permissoes da conta Academy ou a execucao
real dos workflows. A conta esperada e conferida antes de alterar AWS.

## Estados locais existentes

Os estados que ja criamos pelo terminal continuam no computador ate uma
transferencia unica acompanhada, com backup. Nao foram enviados ao GitHub.
Antes de usar a pipeline sobre esse MESMO ambiente, executar somente o bootstrap
e migrar os estados usando init -migrate-state e o backend da mesma conta.
A migracao sera revisada separadamente: nao executar init -reconfigure sobre os
estados locais como forma de transferi-los. A pipeline bloqueia nomes de EKS/RDS
existentes sem estado remoto correspondente em vez de tentar recria-los.

Para executar Terraform manualmente, agora e necessario fornecer VPC e nomes das
funcoes daquela conta, alem dos valores do exemplo. A pipeline descobre esses IDs.
Nunca guardar senha no exemplo versionado nem copiar tfstate entre contas.

## Acesso e demonstracao

A imagem inclui o certificado publico RDS; a API usa DB_SSLMODE=verify-full e
DB_SSLROOTCERT. Localmente, sem essas variaveis, continua sslmode=disable.
DATABASE_URL prevalece sobre DB_* e precisa de sua propria configuracao TLS.

Depois de implantar, o acesso inicial de demonstracao e por port-forward:

```bash
kubectl --context oficina-aws -n oficina port-forward service/oficina-api 8080:80
```

Isso nao cria URL publica. Exposicao HTTPS compartilhada sera uma decisao adicional.
O HPA aumenta pods da API, nao maquinas. Metrics Server e instalado pelo Terraform
como add-on EKS; nao usamos a excecao TLS do Kind no EKS.

## Custos e remocao

Somar EKS, EC2, RDS, discos, IPv4, S3 e trafego; interromper sessao Academy nao
remove esses recursos. Revisar estimativa antes de autorizar provisionamento.
Uma conta diferente recebe recursos independentes e nao encerra a conta anterior.

Para encerrar apos backup, remover a aplicacao, depois platform e cluster usando
seus backends corretos. Desativar explicitamente db_deletion_protection antes da
remocao do RDS. O snapshot final exige nome unico e pode continuar cobrando.
ECR com imagens nao e apagado a forca: decidir como preservar/remover as imagens.
O bucket de estados e o ultimo recurso a considerar, depois de arquivar estados
com seguranca. Estados e backups contem segredos; nunca publicar como artifacts.

Referencias: https://aws.amazon.com/rds/postgresql/pricing/,
https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL.Concepts.General.SSL.html,
https://docs.aws.amazon.com/eks/latest/userguide/metrics-server.html.
