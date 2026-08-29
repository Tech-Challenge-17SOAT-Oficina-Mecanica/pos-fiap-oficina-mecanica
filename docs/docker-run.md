# Execucao local com Docker

Este guia explica como iniciar e testar a API Go e o PostgreSQL localmente.

## Servicos iniciados

- `app`: API Go na porta `8080`.
- `postgres`: banco PostgreSQL na porta `5432`.

## Pre-requisito

- Docker Desktop instalado e em execucao.

## Configuracao local

Os comandos do `Makefile` ja possuem valores locais padrao para banco e JWT. Para iniciar sem
customizacao, basta usar:

```bash
make setup
```

Crie um `.env` somente quando precisar sobrescrever algum valor local ou salvar o token do Sonar:

```bash
cp .env.example .env
```

No Windows PowerShell:

```powershell
Copy-Item .env.example .env
```

O `.env` nao deve ser versionado. Quando existir, ele sobrescreve os valores padrao usados pelo `Makefile`.
Se executar `docker compose` diretamente, sem `make`, informe as variaveis no terminal ou crie o `.env`.

## Como iniciar

Na raiz do repositorio, execute:

```bash
make up
```

O Docker cria a imagem da API, inicia o PostgreSQL e aguarda o banco ficar saudavel antes de iniciar a aplicacao.

## Banco com Make

O `Makefile` organiza a criacao do schema e a carga dos dados iniciais. O Docker Compose nao
executa a migration ou o seed automaticamente; use os alvos abaixo quando quiser popular o banco.

```bash
make db-init
```

`db-init` deve ser usado em um banco vazio. Para recriar o banco local e carregar novamente os
dados de exemplo:

```bash
make db-reset
```

Outros comandos disponiveis:

```bash
make help
make db-migrate
make db-seed
make db-verify
make db-down
```

No Windows, instale GNU Make ou execute os comandos por Git Bash/WSL antes de usar o `Makefile`.

## SonarQube local

Suba o SonarQube:

```bash
make sonar-up
```

Execute coverage e analise:

```bash
make sonar
```

Gere o relatorio local em `reports/sonar-report.md` e `reports/sonar-report.html`:

```bash
make sonar-report
```

Para `make sonar` e `make sonar-report`, informe `SONAR_TOKEN` no terminal ou salve no `.env`.

## Como testar a API

Em outro terminal, execute:

```bash
curl http://localhost:8080/health
```

Resposta esperada:

```json
{"status":"ok"}
```

## Como ver os containers

```bash
docker compose ps
```

## Como ver os logs da API

```bash
docker compose logs -f app
```

Use `Ctrl + C` para sair da visualizacao dos logs.

## Como acessar o PostgreSQL

```bash
docker compose exec postgres psql -U oficina -d oficina
```

Dentro do PostgreSQL, use:

```sql
\dt
```

Para sair, use:

```sql
\q
```

## Configuracao local do banco

| Variavel | Valor |
| --- | --- |
| `DB_HOST` | `postgres` |
| `DB_PORT` | `5432` |
| `DB_NAME` | `oficina` |
| `DB_USER` | `oficina` |
| `DB_PASSWORD` | `oficina` |
| `JWT_SECRET` | `segredo-local` |
| `SONAR_PROJECT_KEY` | `oficina-mecanica` |
| `SONAR_TOKEN` | token local do SonarQube |

As variaveis sao entregues ao container da API pelo `docker-compose.yml`.

## Como parar

```bash
docker compose down
```

Para remover tambem os dados locais do banco:

```bash
docker compose down -v
```

> Atencao: o ultimo comando remove o volume local do PostgreSQL e todos os dados armazenados nele.
