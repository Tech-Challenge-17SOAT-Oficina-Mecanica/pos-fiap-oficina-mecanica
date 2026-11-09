# CI reutilizável nos repositórios da oficina

Cada repositório define seus gatilhos e chama o workflow central com `uses` no
nível do job. O checkout usa o código do repositório chamador. Substitua
`SHA_PUBLICADO` pelo SHA completo do commit que contém os workflows, depois
que a alteração estiver publicada no repositório central.

## Lambda de autenticação em Go

Se `app-lambda-oficina-auth` usar Go, crie `.github/workflows/ci.yml` com o
exemplo abaixo. O `go.mod` deve ficar na raiz. O padrão `./...` compila os
pacotes do módulo. Para compilar só o pacote de entrada, informe seu caminho
em `with.build-package`. A cobertura mínima é de 80%.

```yaml
name: CI
on:
  pull_request:
  push:
    branches: [main]
  workflow_dispatch:
permissions:
  contents: read
jobs:
  ci:
    uses: Tech-Challenge-17SOAT-Oficina-Mecanica/pos-fiap-oficina-mecanica/.github/workflows/build.yml@SHA_PUBLICADO
```

Por padrão, a chamada não inicia PostgreSQL. A opção `database: postgres` é
para projetos que usam a estrutura de migrations e seed da oficina. Nesse caso
passe `CI_DB_PASSWORD` e `JWT_SECRET` de teste explicitamente em `secrets`.
Use `coverage-packages` e `coverage-scope` para ajustar os pacotes medidos.

## Infraestrutura

Em `infra-oficina-db` e `infra-oficina-k8s`, use o mesmo arquivo de gatilhos
e substitua o bloco `jobs` pelo seguinte.

```yaml
jobs:
  terraform:
    uses: Tech-Challenge-17SOAT-Oficina-Mecanica/pos-fiap-oficina-mecanica/.github/workflows/terraform-ci.yml@SHA_PUBLICADO
    with:
      directory: .
```

Informe o diretório do módulo se ele ficar fora da raiz. Para vários módulos,
use uma matriz com um diretório por execução. Escreva testes Terraform com
providers simulados para executar sem credenciais de nuvem.

Se houver regras de proteção que exijam os checks individuais de Terraform no
repositório central, confira seus nomes após a primeira execução. A chamada
reutilizável acrescenta o nome do job interno ao check.
