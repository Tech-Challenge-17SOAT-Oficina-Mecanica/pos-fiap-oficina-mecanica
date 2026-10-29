# Estado Terraform: preparacao pela pipeline

O workflow **02 · 🏗️ Prepare AWS infrastructure** chama internamente o script
scripts/prepare-aws-state.sh para preparar o armazenamento em us-east-1.
Nao existe workflow separado para esta etapa. A conta vem da Variable AWS_ACCOUNT_ID;
os tres Secrets AWS precisam pertencer a essa conta.
## O que ele faz

1. Testa a logica com AWS/Terraform simulados.
2. Confere se as credenciais sao da conta informada.
3. Identifica o bucket `oficina-tfstate-CONTA-us-east-1`.
4. Se nao existir, cria o bucket com AWS CLI. Essa preparacao minima e necessaria
   antes de o Terraform poder gravar seu proprio estado no S3.
5. Ativa bloqueio de acesso publico, versionamento, criptografia AES256 e
   propriedade BucketOwnerEnforced antes de gravar qualquer estado.
6. Inicializa o backend S3 do Terraform state, com chave
   `bootstrap/terraform.tfstate`, criptografia e lock S3 nativo.
7. Importa as quatro configuracoes S3 ainda ausentes do estado e executa plan/apply.
   A existencia do bucket fica sob responsabilidade do script; suas protecoes
   continuam administradas pelo Terraform.

O recurso generico `aws_s3_bucket` consulta Object Lock, uma operacao bloqueada
pela politica do Academy. Por isso ele nao e importado. Se estiver em um estado
anterior, o bloco `removed` o desvincula sem apagar o bucket nem os objetos.
O script nao executa exclusao de bucket ou objetos.

Reexecucoes reconhecem o bucket e os recursos importados. Erro 403 interrompe o
fluxo: nao e interpretado como bucket inexistente. O backend e a identidade sao
especificos da conta. O workflow nao publica arquivos de estado como artifacts.
Bucket e configuracoes podem consumir creditos conforme armazenamento/operacoes.
As permissoes S3 do Academy ainda precisam ser comprovadas na execucao real.

## O que ainda NAO faz

Nao cria nem altera EKS, EC2, RDS, ECR ou configuracoes Kubernetes. Nao transfere
estados existentes do computador e nao executa deploy. Nenhuma dessas etapas
pode assumir estado vazio quando ja existem recursos gerenciados localmente.

O bootstrap guarda seu estado no proprio bucket. Cluster e platform usam chaves
separadas: cluster/terraform.tfstate e platform/terraform.tfstate.
Cada conta precisa de seus estados; nunca reaproveitar um tfstate de outra conta.
O workflow 02 · 🏗️ Prepare AWS infrastructure descobre funcoes/rede na conta autenticada
e chama esse bootstrap automaticamente. Ver [automacao completa](../AUTOMACAO.md).

## Estados atuais do EKS e RDS

Na pipeline, cluster e platform usam os estados remotos no S3. Se houver recursos
criados anteriormente com Terraform local, fazer backup e migrar uma unica vez
com init -migrate-state antes de provisionar pela pipeline. Ver o guia de automacao.
Recursos criados sem Terraform exigem importacao planejada ou remocao explicita.
## Testes locais sem criar recursos

```bash
bash scripts/test-prepare-aws-state.sh
terraform -chdir=infra/aws/state init -backend=false
terraform -chdir=infra/aws/state validate
terraform -chdir=infra/aws/state test
```

Os testes de shell usam simulacoes; o teste Terraform tambem usa AWS simulada.
Nao executar plan/apply manualmente nesta pasta sem configurar seu backend.
A pipeline usa Terraform >= 1.10, necessario para use_lockfile.

Os arquivos principais sao `scripts/prepare-aws-state.sh` e
`.github/workflows/provision-aws.yml`. O script tambem exige
EXPECTED_ACCOUNT_ID para evitar preparar armazenamento em uma conta inesperada.
Nao executar o script localmente como substituto da transferencia dos estados.

## Remocao

O script nao exclui bucket nem objetos; a existencia do bucket nao e gerenciada pelo recurso generico Terraform. Nao remover enquanto estados
ou ambientes precisarem ser administrados. Depois de encerrar os ambientes,
arquivar/migrar os estados com seguranca e planejar explicitamente a remocao de
objetos, versoes e bucket. Encerrar a sessao Academy nao apaga esses recursos.
