# POS FIAP Oficina Mecanica

Backend do Tech Challenge da FIAP para gestao de uma oficina mecanica,
organizado em arquitetura em camadas e DDD.

## Ambiente local

O projeto utiliza Go e PostgreSQL. O Docker Compose inicia os dois servicos.

### Pre-requisitos

- Docker Desktop em execucao.

### Como iniciar

Na raiz do repositorio, execute:

```bash
docker compose up --build
```

A API sera iniciada em `http://localhost:8080` e o PostgreSQL em
`localhost:5432`.

### Como testar

Em outro terminal:

```bash
curl http://localhost:8080/health
```

Resposta esperada:

```json
{"status":"ok"}
```

Para verificar o PostgreSQL:

```bash
docker compose exec postgres pg_isready -U oficina -d oficina
```

### Configuracao do banco

| Variavel | Valor local |
| --- | --- |
| `DB_HOST` | `postgres` |
| `DB_PORT` | `5432` |
| `DB_NAME` | `oficina` |
| `DB_USER` | `oficina` |
| `DB_PASSWORD` | `oficina` |

### Como parar

```bash
docker compose down
```

Para remover tambem os dados locais do banco:

```bash
docker compose down -v
```
