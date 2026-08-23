# 🔧 POS FIAP — Oficina Mecânica

Backend do Tech Challenge da FIAP para gestão de uma oficina mecânica,
organizado em arquitetura em camadas e DDD.

## 🛠️ Skills necessárias

Para executar o projeto, é necessário ter instalado:

- Git;
- Docker Desktop ou Docker Engine;
- Docker Compose;
- GNU Make (necessário para executar comandos `make`).

O Go e o PostgreSQL não precisam ser instalados localmente quando o projeto é
executado pelos containers.

## 🗃️ Modelagem de dados

O modelo lógico do banco de dados, com as entidades, os relacionamentos e as
decisões de modelagem, está disponível no documento
[Modelo Entidade-Relacionamento](docs/04-modelo-entidade-relacionamento.md).

A estrutura do banco é criada pelas migrations localizadas em `db/migrations`,
e os dados iniciais são carregados pelos scripts de seed em `db/seeds`.

## 🔗 URLs importantes

| Recurso | URL |
| --- | --- |
| 💚 Health check | <http://localhost:8080/health> |

## 🚀 Primeiros passos

1. 📥 Clone o repositório e acesse a pasta do projeto:

   ```bash
   git clone https://github.com/lazaro-contato/pos-fiap-oficina-mecanica.git
   cd pos-fiap-oficina-mecanica
   ```

2. 🏗️ Na primeira execução, prepare o banco e suba toda a aplicação:

   ```bash
   make setup
   ```

   Esse comando inicia o banco de dados, aplica as migrations, executa o seed e
   constrói e inicia todos os containers do projeto.

3. ✅ Confirme que a aplicação está disponível:

   ```bash
   curl http://localhost:8080/health
   ```

4. ▶️ Nas próximas execuções, quando o banco já estiver preparado, suba somente
   os containers:

   ```bash
   make up
   ```

🛑 Para encerrar os containers sem apagar os dados persistidos:

```bash
make db-down
```
