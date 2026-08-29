.DEFAULT_GOAL := help

POSTGRES_SERVICE := postgres
POSTGRES_CONTAINER := oficina-postgres
POSTGRES_USER := oficina
POSTGRES_DB := oficina
MIGRATIONS := $(sort $(wildcard db/migrations/*.sql))
SEED := db/seeds/V900__dados_iniciais.sql
SONAR_SCANNER_IMAGE ?= sonarsource/sonar-scanner-cli:latest

ifeq ($(OS),Windows_NT)
DOCKER ?= docker.exe
else
DOCKER ?= docker
endif

.PHONY: help setup up recreate db-up db-down db-reset db-migrate db-seed db-init db-verify test coverage sonar-up sonar

help: ## Lista os comandos disponiveis
	@echo "Uso: make <alvo>"
	@echo "  setup          Prepara e sobe todo o projeto apos o primeiro clone"
	@echo "  up             Sobe os containers para o uso diario"
	@echo "  recreate       Recria app e banco, aplica migrations e carrega o seed"
	@echo "  db-up          Inicia o PostgreSQL local"
	@echo "  db-down        Para os containers sem remover os dados"
	@echo "  db-reset       Remove o banco local, cria o schema e carrega o seed"
	@echo "  db-migrate     Aplica a migration no banco vazio"
	@echo "  db-seed        Carrega os dados iniciais apos aplicar a migration"
	@echo "  db-init        Cria um banco local vazio com schema e seed"
	@echo "  db-verify      Exibe contagens essenciais para verificar a carga"
	@echo "  test           Executa todos os testes no container Go"
	@echo "  coverage       Gera coverage.out para o SonarQube"
	@echo "  sonar-up       Sobe o SonarQube local"
	@echo "  sonar          Executa a analise SonarQube via Docker"

setup: db-init ## Prepara e sobe todo o projeto apos o primeiro clone
	$(DOCKER) compose up -d --build

up: ## Sobe os containers para o uso diario
	$(DOCKER) compose up -d

recreate: db-reset ## Recria app e banco, aplica migrations e carrega o seed
	$(DOCKER) compose up -d --build

db-up: ## Inicia o PostgreSQL local
	$(DOCKER) compose up -d --wait $(POSTGRES_SERVICE)

db-down: ## Para os containers sem remover os dados
	$(DOCKER) compose down

db-reset: ## Remove o banco local, cria o schema e carrega o seed
	$(DOCKER) compose down -v
	"$(MAKE)" db-init

db-migrate: db-up ## Aplica as migrations no banco vazio
	@for migration in $(MIGRATIONS); do $(DOCKER) compose exec -T $(POSTGRES_SERVICE) psql -v ON_ERROR_STOP=1 -U $(POSTGRES_USER) -d $(POSTGRES_DB) < $$migration || exit $$?; done

test: ## Executa testes com cobertura no container Go
	$(DOCKER) compose run --rm --build test

coverage: ## Gera coverage.out para o SonarQube
	$(DOCKER) compose run --rm --build test sh -c "go test -covermode=atomic -coverpkg=./internal/... -coverprofile=coverage.out ./internal/... ./tests/integration/..."

sonar-up: ## Sobe o SonarQube local
	$(DOCKER) compose up -d sonarqube

sonar: coverage ## Executa a analise SonarQube via Docker
	$(DOCKER) compose run --rm sonar-scanner

db-seed: db-up ## Carrega os dados iniciais apos aplicar a migration
	$(DOCKER) compose exec -T $(POSTGRES_SERVICE) psql -v ON_ERROR_STOP=1 -U $(POSTGRES_USER) -d $(POSTGRES_DB) < $(SEED)

db-init: db-migrate db-seed ## Cria um banco local vazio com schema e seed

db-verify: db-up ## Exibe contagens essenciais para verificar a carga
	$(DOCKER) compose exec -T $(POSTGRES_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -P pager=off -c "SELECT 'clientes' AS entidade, COUNT(*) AS quantidade FROM cliente UNION ALL SELECT 'veiculos', COUNT(*) FROM veiculo UNION ALL SELECT 'servicos', COUNT(*) FROM servico UNION ALL SELECT 'itens_estoque', COUNT(*) FROM item_estoque UNION ALL SELECT 'ordens_servico', COUNT(*) FROM ordem_servico UNION ALL SELECT 'orcamentos', COUNT(*) FROM orcamento UNION ALL SELECT 'reservas_estoque', COUNT(*) FROM reserva_estoque UNION ALL SELECT 'pedidos_compra', COUNT(*) FROM pedido_compra ORDER BY entidade;"
