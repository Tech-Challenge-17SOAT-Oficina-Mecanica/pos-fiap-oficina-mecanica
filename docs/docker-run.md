# Execucao local com Docker

Este guia explica como iniciar e testar a API Go e o PostgreSQL localmente.

## Servicos iniciados

- `app`: API Go na porta `8080`.
- `postgres`: banco PostgreSQL na porta `5432`.

## Pre-requisito

- Docker Desktop instalado e em execucao.

## Como iniciar

Na raiz do repositorio, execute:

```bash
docker compose up --build
```

O Docker cria a imagem da API, inicia o PostgreSQL e aguarda o banco ficar saudavel antes de iniciar a aplicacao.

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
