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
| 📖 Swagger UI | <http://localhost:8081> |

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

5. ♻️ Após alterar arquivos Go, reconstrua e reinicie a API:

   ```bash
   make restart
   ```

   O Go gera um binário durante a imagem Docker; por isso apenas reiniciar o
   container não incorpora mudanças no código-fonte.

🛑 Para encerrar os containers sem apagar os dados persistidos:

```bash
make db-down
```

## 🏗️ Arquitetura

O projeto segue uma **arquitetura em camadas**, alinhada aos princípios de **Domain-Driven
Design (DDD)**, com o código organizado por contexto de negócio (`cliente`, `veiculo`,
`servico`, `estoque`, `orcamento`, `ordemservico`) repetido em cada camada:

- **`presentation`** — camada de entrada da aplicação: handlers HTTP, rotas e conversão
  entre o corpo das requisições/respostas e os modelos internos.
- **`application`** — casos de uso: orquestra as regras de negócio, coordenando chamadas ao
  domínio e à infraestrutura.
- **`domain`** — o núcleo do sistema: entidades, regras de negócio e contratos (interfaces)
  independentes de framework ou de banco de dados.
- **`infrastructure`** — implementações concretas dos contratos do domínio, como repositórios
  que acessam o PostgreSQL e integrações externas.
- **`shared`** — recursos técnicos compartilhados entre os contextos: conexão/configuração do
  banco de dados (`database`), respostas/erros/middlewares HTTP comuns (`http`) e configuração
  da aplicação (`config`).

Essa separação mantém as regras de negócio isoladas de detalhes técnicos (HTTP, banco de
dados), facilitando testes e a evolução independente de cada camada.

## 📁 Estrutura de pastas

```text
pos-fiap-oficina-mecanica/
│
├── cmd/
│   └── oficina-mecanica/
│       └── main.go               # Ponto de entrada da aplicação
│
├── internal/
│   │
│   ├── presentation/              # Handlers HTTP e rotas
│   │   ├── cliente/
│   │   ├── veiculo/
│   │   ├── servico/
│   │   ├── estoque/
│   │   ├── orcamento/
│   │   └── ordemservico/
│   │
│   ├── application/                # Casos de uso
│   │   ├── cliente/
│   │   ├── veiculo/
│   │   ├── servico/
│   │   ├── estoque/
│   │   ├── orcamento/
│   │   └── ordemservico/
│   │
│   ├── domain/                     # Entidades e regras de negócio
│   │   ├── cliente/
│   │   ├── veiculo/
│   │   ├── servico/
│   │   ├── estoque/
│   │   ├── orcamento/
│   │   └── ordemservico/
│   │
│   ├── infrastructure/              # Repositórios e integrações externas
│   │   ├── cliente/
│   │   ├── veiculo/
│   │   ├── servico/
│   │   ├── estoque/
│   │   ├── orcamento/
│   │   ├── ordemservico/
│   │   └── seguranca/
│   │
│   └── shared/                      # Recursos técnicos compartilhados
│       ├── database/                # Conexão/configuração do PostgreSQL
│       ├── http/                    # Respostas, erros e middlewares comuns
│       └── config/                  # Configurações da aplicação
│
├── migrations/                       # Scripts de evolução do banco
│
├── tests/
│   └── integration/                  # Testes de integração (unitários ficam junto ao código)
│
├── docs/                              # Documentação do projeto
│
├── Dockerfile
├── docker-compose.yml
├── go.mod                             # Definição do módulo/dependências Go
├── go.sum                             # Checksums das dependências
└── README.md
```

## 🗄️ Por que PostgreSQL

O banco de dados principal da API é o **PostgreSQL**, um banco **SQL relacional**. A decisão
foi tomada porque o domínio da oficina é fortemente relacional e transacional: clientes,
veículos, ordens de serviço, orçamentos, peças/insumos e movimentações de estoque se
relacionam entre si e precisam permanecer consistentes mesmo quando várias entidades são
alteradas na mesma operação (por exemplo, dar baixa em um item de estoque ao mesmo tempo em
que se registra o consumo em uma ordem de serviço).

Principais motivos:

- **Relacionamentos nativos** — chaves estrangeiras e `JOIN`s representam naturalmente as
  relações um-para-muitos e muitos-para-muitos do domínio (um cliente com vários veículos, uma
  ordem de serviço com vários itens de estoque, etc.).
- **Integridade referencial no próprio banco** — restrições como `FOREIGN KEY`, `NOT NULL`,
  `UNIQUE` e `CHECK` evitam estados inválidos (ex.: ordem de serviço para veículo inexistente)
  independentemente de qual código acessa o banco.
- **Transações ACID** — operações que precisam ser "tudo ou nada", como aprovar um orçamento,
  reservar itens de estoque e atualizar a ordem de serviço, são executadas de forma atômica.
- **Consultas e relatórios cruzados** — SQL é adequado para os relatórios operacionais do
  domínio (tempo médio de execução, fila de atendimento, itens abaixo do mínimo, histórico por
  cliente/veículo), que cruzam várias entidades.
- **Flexibilidade quando necessário** — o tipo `JSONB` permite armazenar atributos variáveis
  (ex.: dados específicos de diagnóstico) sem abrir mão do modelo relacional no núcleo do
  sistema.

Alternativas NoSQL (como MongoDB) foram avaliadas e descartadas como banco principal: elas são
tecnicamente capazes, mas deslocariam para a aplicação parte do trabalho de integridade e
consistência que o PostgreSQL já resolve nativamente, sem trazer vantagem clara para este
domínio (não há requisito de dados extremamente heterogêneos, agregados independentes ou
escala horizontal extrema que justifique essa troca). A análise completa está registrada em
[docs/02-decisoes-arquiteturais.md](docs/02-decisoes-arquiteturais.md) e no
[Modelo Entidade-Relacionamento](docs/04-modelo-entidade-relacionamento.md).

## 👥 Integrantes

##### João Victor — RM376616 - jhonnysket - [LinkedIn](https://www.linkedin.com/in/joão-victor-de-oliveira-63a883143/)

##### José Lázaro — RM371406 - joselzro - [LinkedIn](https://www.linkedin.com/in/lazaro-contato/)

##### Hemily Nara — RM373944 - hemilyynara - [LinkedIn](https://www.linkedin.com/in/hemilynara/)

##### Carlos Henrique — RM376201 - caohz_55485 - [LinkedIn](https://www.linkedin.com/in/carlos1br/)

##### Helena Miranda — RM376202 - helenamiranda6974 - [LinkedIn](https://www.linkedin.com/in/hmirandas/)
