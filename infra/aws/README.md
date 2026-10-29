# AWS Academy: EKS + PostgreSQL no RDS

Preparacao da automacao: [bucket privado para os estados Terraform](state/README.md).
Essa etapa ainda nao transfere os estados existentes nem executa deploy.

Esta alternativa preserva Docker Compose e Terraform Kind locais. Nenhum recurso
AWS foi criado durante a preparacao dos arquivos. Execute os passos acompanhado,
um de cada vez; nao execute apply antes de revisar plano, permissoes e custo.

## Arquitetura

```text
Terraform aws/cluster
  + EKS e uma maquina EC2 t3.medium
  + PostgreSQL 16 no RDS (db.t3.micro, Single-AZ, 20 GiB gp2)
  + rede de acesso ao banco, subnet group e parametros TLS

Terraform aws/platform
  + namespace oficina, Secret e ConfigMap com endpoint RDS

Deploy da aplicacao (integracao pendente)
  + imagem da API -> EKS -> RDS pela rede interna
  + Metrics Server e HPA no EKS
```

Reutilizamos VPC, sub-redes publicas e funcoes IAM do Academy. Nao criamos funcoes
IAM, NAT Gateway, balanceador, driver EBS ou PostgreSQL dentro do Kubernetes.
Mesmo nas sub-redes existentes, o RDS nao tem endereco publico. O security group
aceita porta 5432 apenas do grupo das maquinas EKS. O computador do administrador
nao conecta diretamente ao banco; migrations e consultas deverao partir do cluster.

O EKS permanece com um unico no e nao oferece alta disponibilidade. HPA aumenta
pods, nao maquinas; novas replicas podem ficar Pending se faltar capacidade.
O banco Single-AZ nao oferece redundancia Multi-AZ.

## Arquivos

- `cluster/database.tf`: RDS, armazenamento, acesso de rede e TLS obrigatorio.
- `cluster/main.tf`: EKS e maquinas; usa funcoes IAM existentes.
- `cluster/outputs.tf`: nome do cluster, comando kubectl e endpoint do banco.
- `platform/main.tf`: credenciais e configuracao da API para acessar o RDS.
- `internal/shared/database/database.go`: DB_SSLMODE e DB_SSLROOTCERT opcionais.
- `Dockerfile`: inclui o certificado publico da AWS para verificar o RDS.

Os manifestos `k8s/local/database` continuam exclusivos do Kind. As configuracoes
compartilhadas ficam em `k8s/base/config/configmap.yaml`; na AWS, Terraform platform
adapta esses valores para RDS. Nao aplicar esse ConfigMap diretamente no EKS.
Use `k8s/aws` para o deploy da API com Kustomize e selecione o SHA publicado no
ECR antes de aplicar. Nao aplicar `kubectl apply -R -f k8s`.

## Pre-requisitos

AWS CLI v2, Terraform >= 1.7 para testes com mocks, kubectl compativel com EKS,
Docker para construir imagem e sessao ativa do Academy. Credenciais temporarias
ficam em `~/.aws/credentials` ou variaveis AWS, fora do repositorio.

As consultas feitas pelo usuario confirmaram duas sub-redes em AZs distintas,
IPv4 publico automatico e rota ativa para Internet Gateway. As funcoes de cluster
e no existem; politicas basicas do no foram conferidas. Criacao real ainda depende
das permissoes do Academy. A disponibilidade de PostgreSQL 16/db.t3.micro/gp2 foi
consultada, mas nao comprova autorizacao para criar a instancia.

## 1. Validar os arquivos (sem recursos reais)

```bash
terraform -chdir=infra/aws/cluster init
terraform -chdir=infra/aws/cluster validate
terraform -chdir=infra/aws/cluster test
terraform -chdir=infra/aws/platform init
terraform -chdir=infra/aws/platform validate
```

Os testes usam AWS simulada e verificam rede restrita, banco privado, criptografia,
protecao de dados, acesso administrativo e numero de maquinas.

## 2. Preparar configuracoes

Copiar `infra/aws/cluster/terraform.tfvars.example` para `terraform.tfvars` na mesma
pasta, preservando as sub-redes conferidas e substituindo o IP administrativo.
Validar que a versao EKS selecionada esta em suporte padrao em us-east-1.

No Git Bash, digitar uma senha forte sem coloca-la no historico do terminal:

```bash
read -rsp 'Senha do banco (16+ caracteres, letras/numeros/_/-): ' TF_VAR_db_password
printf '\n'
export TF_VAR_db_password
```

Guardar essa senha com seguranca e reutilizar o MESMO valor em cluster e platform.
Nao regenerar a cada deploy. A senha e sensivel na saida, mas consta no estado
Terraform: estados e planos ficam ignorados no Git e precisam de armazenamento
protegido. A pipeline futura precisara de estado duravel e controle de concorrencia.

## 3. Revisar plano e provisionar (com autorizacao)

```bash
terraform -chdir=infra/aws/cluster plan -out=aws.tfplan
# Revisar recursos e custo antes do proximo comando.
terraform -chdir=infra/aws/cluster apply aws.tfplan
aws eks update-kubeconfig --region us-east-1 --name oficina-academy --alias oficina-aws
kubectl --context oficina-aws get nodes
aws rds describe-db-instances --region us-east-1 --db-instance-identifier oficina-academy-postgres --query 'DBInstances[0].{Estado:DBInstanceStatus,Publico:PubliclyAccessible}'
```

Resultado esperado: no Ready, RDS available e Publico false. Isso ainda nao foi
comprovado em AWS. O principal IAM criador recebe acesso inicial; outra funcao,
incluindo a da pipeline, precisa ser autorizada no EKS explicitamente.

## 4. Preparar configuracao da API no Kubernetes

```bash
export TF_VAR_db_host=$(terraform -chdir=infra/aws/cluster output -raw db_host)
read -rsp 'Chave JWT (32+ caracteres): ' TF_VAR_jwt_secret
printf '\n'
export TF_VAR_jwt_secret
terraform -chdir=infra/aws/platform plan -out=platform.tfplan
terraform -chdir=infra/aws/platform apply platform.tfplan
kubectl --context oficina-aws -n oficina get configmap,secret
```

Platform cria apenas namespace, Secret e ConfigMap. Nao cria tabelas nem instala
API. Configura DB_SSLMODE=verify-full e o certificado RDS incluido na imagem.
O endereco precisa ser o DNS original do RDS para a verificacao de identidade.
DATABASE_URL, quando definida, prevalece sobre DB_* e deve conter seu proprio TLS.
Conexoes locais sem DB_SSLMODE continuam com sslmode=disable.

## 5. Deploy recebido da main, aguardando validacao real

O workflow Build executa build/testes/cobertura. Depois de sucesso em push na main,
docker-build-push publica API e migrations no ECR oficina-mecanica em us-east-1,
com tags do mesmo SHA, e executa deploy. PR e workflow_dispatch do Build nao
publicam. O workflow Docker Build and Deploy tambem permite acionamento manual,
exigindo Build aprovado no mesmo commit. Ver [guia de deploy](../../docs/06-deploy-eks.md).
Fluxo: verificacoes -> imagens -> Flyway no RDS a partir do cluster -> API -> HPA -> HTTP.

Faltam acesso da pipeline ao endpoint restrito por IP, autenticacao temporaria
Academy e estado compartilhado Terraform. Flyway e historico de checksums vieram
da main, com renomeacao das migrations repetidas para subversoes unicas.
Nao reaplicar SQL inteiro nem criar baseline automatico em banco ja inicializado.
O repositorio ECR precisa existir. A publicacao usa os tres Secrets AWS temporarios
do Academy. Os nos EKS tem AmazonEC2ContainerRegistryReadOnly; validar leitura da
imagem na mesma conta. O deploy deve consumir o SHA publicado, sem novo build.

Depois de implantar a API, acesso inicial de demonstracao sera por:

```bash
kubectl --context oficina-aws -n oficina port-forward service/oficina-api 8080:80
```

Isso nao cria URL publica. Exposicao compartilhada com HTTPS sera uma decisao
adicional. O usuario do RDS e administrador neste laboratorio; em producao separar
usuario de migrations e usuario de aplicacao com privilegios minimos.

## Custos e remocao

RDS acrescenta cobranca de instancia, 20 GiB e backups/snapshots aos custos EKS,
EC2, disco do no, IPv4 publico e trafego. A estimativa anterior so do EKS nao vale
como total. Consultar calculadora AWS antes do apply. Encerrar sessao Academy nao
remove recursos nem garante fim das cobrancas; RDS pode continuar executando.

O banco tem backup automatico por 1 dia, protecao de exclusao e snapshot final.
Para encerrar, apos guardar dados necessarios:

1. Remover API/HPA e recursos criados fora do Terraform.
2. `terraform -chdir=infra/aws/platform destroy`.
3. Definir `db_deletion_protection = false` no tfvars de cluster, revisar plan e
   aplicar a mudanca ANTES de destroy.
4. `terraform -chdir=infra/aws/cluster destroy` cria snapshot final antes da remocao
   do RDS. Nao executa rollback dos dados locais; sao ambientes independentes.
5. Conferir snapshot final e backups retidos: continuam podendo gerar custos.
   O nome do snapshot deve ser unico; se ja existir de uma execucao anterior,
   alterar final_snapshot_identifier conscientemente antes de novo destroy.
6. Conferir console e custos. VPC e funcoes preexistentes nao sao removidas.

Referencias: https://aws.amazon.com/rds/postgresql/pricing/,
https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/PostgreSQL.Concepts.General.SSL.html,
https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.SSL.html.
