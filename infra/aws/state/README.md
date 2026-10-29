# Estado Terraform: preparacao pela pipeline

O workflow **Prepare AWS Terraform state** prepara o armazenamento compartilhado
em us-east-1. E manual e nao roda em PR ou push. Depois de publicado no GitHub,
abrir Actions, selecionar esse workflow e informar o ID de 12 digitos da conta
Academy esperada. Os tres Secrets AWS precisam pertencer a essa mesma conta.

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
7. Importa os cinco recursos S3 ainda ausentes do estado e executa plan/apply,
   passando a administrar o bucket pelo Terraform.

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
O workflow Prepare AWS infrastructure descobre funcoes/rede na conta autenticada
e chama esse bootstrap automaticamente. Ver [automacao completa](../AUTOMACAO.md).

## Estados atuais do EKS e RDS

Os estados de `infra/aws/cluster` e `infra/aws/platform` continuam LOCAIS.
Depois de validar o bootstrap na mesma conta dos recursos atuais, faremos backup
protegido e uma transferencia unica com init -migrate-state. Nao apagar estados
locais, nao enviar por chat e nao commitar. Nao iniciar um novo apply de cluster
na pipeline antes dessa transferencia: ela precisa reconhecer os recursos atuais.

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
`.github/workflows/prepare-aws-state.yml`. O script tambem exige
EXPECTED_ACCOUNT_ID para evitar preparar armazenamento em uma conta inesperada.
Nao executar o script localmente como substituto da transferencia dos estados.

## Remocao

O bucket tem prevent_destroy e force_destroy=false. Nao remover enquanto estados
ou ambientes precisarem ser administrados. Depois de encerrar os ambientes,
arquivar/migrar os estados com seguranca e planejar explicitamente a remocao de
objetos, versoes e bucket. Encerrar a sessao Academy nao apaga esses recursos.
