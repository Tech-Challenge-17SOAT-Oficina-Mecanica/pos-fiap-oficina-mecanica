# Automacao AWS: conta escolhida na execucao

A conta definitiva pode ser escolhida depois do merge. As configuracoes nao usam
IDs pessoais de conta, VPC, sub-redes ou funcoes IAM. A regiao do laboratorio e
us-east-1. Cada conta tem bucket e estados separados.

## Configuracao no GitHub

Secrets obrigatorios:
- AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY e AWS_SESSION_TOKEN: sessao Academy.
- AWS_DB_PASSWORD: senha estavel do RDS (16+ caracteres, letras/numeros/_/-).
- AWS_JWT_SECRET: chave estavel da aplicacao (32+ caracteres).

CI_DB_PASSWORD e JWT_SECRET continuam destinados aos testes da aplicacao.
AWS_REGION e ECR_URI antigos nao sao usados: a regiao e padronizada e o URI do
ECR e consultado na conta autenticada. Nunca enviar valores desses Secrets por PR.
Para um ambiente existente, usar a senha/chave que ele ja utiliza; a automacao
bloqueia alteracoes implicitas de credenciais. Renovar sessao Academy normalmente
exige trocar apenas os tres Secrets de acesso, nao as credenciais do banco.

Variables opcionais (nao sao Secrets): AWS_CLUSTER_NAME (padrao oficina-academy),
ECR_REPOSITORY (padrao oficina-mecanica), AWS_VPC_ID, AWS_EKS_CLUSTER_ROLE e
AWS_EKS_NODE_ROLE. Sem as tres ultimas, a pipeline consulta a VPC padrao, duas
sub-redes publicas em zonas diferentes com rota Internet Gateway, e as funcoes
LabEksClusterRole/LabEksNodeRole. Descoberta ambigua ou rede inadequada interrompe.

## Fluxos manuais depois do merge

Nenhum push ou PR cria infraestrutura nem executa deploy AWS. As verificacoes
continuam automaticas. O provisionamento e a entrega sao iniciados por botao:

1. **Prepare AWS Terraform state**: prepara somente o bucket e seu estado. Informar
   expected_account_id, correspondente as credenciais. Pode ser usado separadamente
   para a transferencia dos estados existentes.
2. **Prepare AWS infrastructure**: informar expected_account_id, admin_cidr (seu IP
   administrativo /32) e marcar confirm_provision para autorizar consumo de creditos.
   Prepara o bucket automaticamente, consulta os IDs, executa plan/apply de cluster
   e platform. Cria EKS, um no, RDS, ECR, Metrics Server, namespace, Secret e ConfigMap.
   Nao publica a API. Reexecucoes usam o estado remoto e nao recriam tudo.
3. **Docker Build and Deploy**: informar expected_account_id e selecionar a revisao.
   Exige sucesso de Build e Validate deployment no mesmo SHA. Publica API e Flyway
   no ECR, executa migrations e somente entao atualiza a API e verifica /health e
   /ready. Nao executa Terraform nem recria banco.

Planos com remocao/substituicao ou rotacao de senha/chave sao bloqueados. Uma
mudanca desse tipo exige revisao especifica, fora desse fluxo automatico.
Nao publicamos planos ou estados como artifacts, porque podem conter segredos.
A aprovacao manual aqui e a autorizacao no disparo do workflow; nao ha uma segunda
pausa de aprovacao entre plan e apply. Conferir os logs e a conta antes de iniciar.

## Acesso do runner ao EKS

O endpoint mantem o IP administrativo configurado no Terraform. Durante platform
ou deploy, o script adiciona somente o IPv4 /32 do runner, aguarda a atualizacao,
executa a tarefa e remove a entrada temporaria, preservando os demais CIDRs.
Nao abre 0.0.0.0/0 e nao aplica a excecao kubelet-insecure-tls do Kind.
A identidade AWS do runner tambem precisa ter acesso Kubernetes: a identidade
criadora recebe acesso inicial. Trocar de funcao IAM na mesma conta pode exigir
uma autorizacao EKS adicional; credenciais renovadas da mesma funcao nao sao uma
identidade IAM nova.

Os workflows sao serializados por conta. Falhas normais tentam limpar o CIDR;
revogacao de credenciais, cancelamento forcado ou falha da API podem impedir a
limpeza. O script sinaliza esse caso quando consegue finalizar: conferir endpoint
se a execucao for interrompida. Nao ha garantia de cleanup apos kill forcado.

## Recursos existentes e transferencia unica dos estados

A pipeline NAO assume EKS/RDS existentes como recursos novos. Se encontrar o nome
alvo sem o estado remoto correspondente, bloqueia. Nao importar automaticamente
banco ou cluster desconhecidos. O ECR existente pode ser importado na mesma conta.

Na conta em que ja criamos a infraestrutura pelo terminal:
1. Executar somente Prepare AWS Terraform state com as credenciais dessa conta.
2. Guardar backup protegido dos estados locais de cluster e platform.
3. Transferir cada estado com terraform init -migrate-state, usando o bucket dessa
   conta e as chaves cluster/terraform.tfstate e platform/terraform.tfstate.
4. Conferir os estados e executar a preparacao completa com as mesmas credenciais
   do banco. A adicao de ECR/metricas e demais alteracoes aparecem no plan.

Nao rodar init -reconfigure localmente para transferir: ele nao copia o estado.
Nao copiar estados de uma conta para outra nem usar estado vazio para substituir
recursos existentes. Se a outra conta ja tem um cluster criado manualmente com
outro nome (por exemplo, oficina-cluster), ele nao e removido nem adotado por este
fluxo. Revisar nomes, importacao e custos antes de iniciar outro ambiente.

## Conta nova

Atualizar os Secrets AWS e os Secrets da aplicacao e informar o ID da nova conta.
A preparacao usa um bucket proprio, comeca com estado vazio e cria o ambiente.
Isso NAO migra registros do banco nem encerra os recursos da conta anterior.
Os limites e permissoes do Academy precisam permitir as operacoes. O provisionamento
completo ainda precisa ser validado na AWS; testes simulados nao comprovam permissoes.

## Testes sem AWS

```bash
bash scripts/test-aws-automation.sh
bash scripts/test-prepare-aws-state.sh
```

Usam comandos AWS/Terraform simulados, mas jq real, para verificar descoberta,
conta incorreta, rede, estados ausentes, planos perigosos e cleanup do runner.
Os contratos Terraform e a renderizacao Kustomize tambem rodam na PR.
