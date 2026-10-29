# Infraestrutura

Escolha o ambiente que deseja preparar:

- [Local: Kind e PostgreSQL](local/README.md): arquivos Terraform em
  `infra/local/cluster` e `infra/local/platform`.
- [AWS Academy: EKS e RDS](aws/README.md): arquivos Terraform em
  `infra/aws/cluster` e `infra/aws/platform`.

Cada ambiente tem estados Terraform e dados independentes. Executar um ambiente
nao exige iniciar o outro. Nunca versionar estados, planos, credenciais ou kubeconfig.
