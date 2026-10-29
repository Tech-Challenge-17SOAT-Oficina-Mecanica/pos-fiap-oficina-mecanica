# Deploy no EKS com Kustomize e migrations

O deploy usa o cluster `oficina-academy`, na regiao `us-east-1`, e o repositorio
ECR `oficina-mecanica`. O deploy e iniciado manualmente; o merge na main nao o executa. A execucao manual
permite testar a branch selecionada no mesmo cluster de laboratorio, antes do merge.

## Fluxo

1. O workflow `Build` valida compilacao, testes e cobertura minima de 80%.
2. O workflow `Docker Build and Deploy` publica a API com a tag do commit aprovado
   e o Flyway com a tag `<commit>-migrations`, no mesmo repositorio ECR.
3. No mesmo arquivo `.github/workflows/docker-build-push.yml`, o job `deploy`
   depende de `build-and-push` por `needs`. Ambos usam `github.sha` da revisao selecionada
   do CI aprovado para checkout e imagens; nao precisam transferir artifacts.
4. O workflow configura AWS, kubectl e Kustomize. Credenciais AWS ficam nos secrets
   `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` e `AWS_SESSION_TOKEN`.
5. Terraform prepara EKS, RDS, namespace, Secret e ConfigMap. O script verifica os recursos existentes, sem sobrescrever as configuracoes do RDS.
6. Um Job Kubernetes executa Flyway dentro do cluster, com acesso privado ao RDS.
   O JDBC usa verify-full e o certificado RDS incluido na imagem. O banco nao e exposto ao runner.
7. Somente depois do Job concluir, aplica a API com Kustomize, aguarda o rollout
   e consulta `/health` e `/ready` por um port-forward temporario.

Deploys sao serializados, sem cancelar migrations em andamento. Execucoes antigas
sao ignoradas antes de aplicar recursos se a `main` ja avancou. Se a migration
falhar, o script exibe o diagnostico e nao aplica o Deployment da API. Nao ha
rollback automatico de schema ou de aplicacao.

A conta e configurada na Variable AWS_ACCOUNT_ID do GitHub; ela deve corresponder
aos Secrets AWS. O provisionamento e o deploy conferem a identidade automaticamente.

## Teste manual antes do merge

O arquivo de workflow ja existe na branch padrao. A versao da PR adiciona
`workflow_dispatch`, que pode ser acionado pelo CLI com a referencia da branch:

```bash
gh workflow run docker-build-push.yml \
  --repo Tech-Challenge-17SOAT-Oficina-Mecanica/pos-fiap-oficina-mecanica \
  --ref main
```

A execucao exige sucesso de Build e Validate deployment para o mesmo commit
(incluindo testes e cobertura). Usa o SHA da branch selecionada para as duas
imagens, e verifica se essa branch nao avancou antes de aplicar recursos.
O teste manual publica imagens e altera o namespace `oficina` no EKS: nao e
dry-run nem ambiente separado. Nao carrega seed. Nenhum merge na main e necessario.

## Recursos Kustomize

- `k8s/aws/kustomization.yaml`: Service, Deployment da API e HPA compartilhados.
  Nao inclui banco, namespace, ConfigMap, Secret ou teste k6; Terraform gerencia esses recursos de infraestrutura.
- `k8s/aws/migrations/kustomization.yaml`: Job Flyway, executado separadamente antes
  da aplicacao. Nome e imagem sao ajustados no workspace do runner.
- As imagens locais sao substituidas por referencias ECR com tags por commit;
  essas alteracoes nao sao commitadas de volta ao repositorio.

Para somente visualizar os manifests, sem acessar o cluster:

```bash
kubectl kustomize k8s/aws
kubectl kustomize k8s/aws/migrations
```

Os renders locais ainda utilizam imagens locais. O workflow preenche as imagens
ECR antes da aplicacao; `kubectl apply -k k8s/aws` sozinho nao executa as migrations.

## Pre-requisitos no ambiente

- Node EKS pronto, acesso ao ECR e capacidade para Flyway, API e
  componentes do Kubernetes; o rollout da API pode precisar de um Pod adicional.
- RDS disponivel e acessivel pela porta 5432 a partir do grupo das maquinas EKS; nao requer driver EBS no Kubernetes.
- Namespace `oficina` e Secret `oficina-api-secret`, com `DB_USER`, `DB_PASSWORD`
  e `JWT_SECRET`. O workflow verifica a existencia e nao imprime nem sobrescreve
  o Secret. Se a API usa `DATABASE_URL`, ela deve apontar para o mesmo banco que
  as variaveis `DB_*` utilizadas pelo Job.
- Identidade AWS do Actions autorizada no EKS para ler o Secret por metadados,
  aplicar os manifests, criar e acompanhar Jobs, consultar logs e fazer port-forward.
  Os secrets AWS do GitHub nao concedem RBAC Kubernetes automaticamente.
- Metrics Server instalado pelo workflow de infraestrutura como add-on EKS. O HPA nao aumenta o numero de nodes.

Nao se aplicam credenciais do banco de testes do CI neste ambiente. As credenciais
temporarias do Academy precisam estar validas no momento da publicacao e do deploy.

## Historico e primeira aplicacao

Flyway registra versoes e checksums na tabela `flyway_schema_history`. Na primeira
execucao em um banco vazio, aplica os 22 SQLs existentes. Nas proximas, valida o
historico e executa somente versoes pendentes. O seed V900 nao entra na imagem.

Durante o build, somente nas copias incorporadas a imagem, removem-se os wrappers
externos BEGIN/COMMIT. Flyway controla a transacao e registra o historico junto
com o SQL. Os arquivos originais mantem esses wrappers para os fluxos com psql.

Os prefixos repetidos V005, V006, V008 e V009 receberam subversoes, como V005.1,
V005.2 e V005.3. Apenas os nomes mudaram; a ordem anterior e o conteudo dos arquivos
foram preservados. Os comandos existentes do Makefile e CI continuam ordenando
os arquivos por nome. Novas migrations devem usar versoes unicas e crescentes.

Nao editar SQL ja aplicado com Flyway. Banco com schema criado manualmente e sem
historico nao e assumido como vazio: Flyway deve bloquear a inicializacao e a
adocao desse banco precisa ser planejada. Nao habilitar baseline automatico.

O Job tem prazo de 600 segundos, nao repete automaticamente uma falha e e removido
apos um dia. Logs devem ser consultados antes da remocao:

```bash
kubectl get jobs -n oficina
kubectl logs job/NOME_DO_JOB -n oficina
kubectl get pods -n oficina
```

## Validacao automatizada

O workflow `Validate deployment` renderiza Kustomize, valida schemas Kubernetes
e workflows, e executa `scripts/test-migrations.sh` em Docker. O teste cria seu
proprio banco descartavel e verifica primeira aplicacao, reexecucao sem duplicacao,
rejeicao de checksum alterado e rollback de uma migration invalida.

Para repetir localmente com Docker ativo:

```bash
bash scripts/test-migrations.sh
```

Esse teste nao acessa o EKS. Backup e recuperacao do banco do cluster continuam sendo responsabilidades separadas; conferir backups e snapshot final do RDS.

O endpoint EKS atual aceita somente os IPs administrativos configurados no Terraform.
O runner GitHub precisa de acesso de rede autorizado e da identidade correta no EKS.
Esses pontos e o deploy real ainda precisam ser validados; testes simulados nao os comprovam.

A preparacao completa e a escolha da conta estao em [Automacao AWS](../infra/aws/AUTOMACAO.md).
O runner recebe acesso /32 temporario, com tentativa de remocao ao terminar; nao e necessario abrir o endpoint para toda a internet.
