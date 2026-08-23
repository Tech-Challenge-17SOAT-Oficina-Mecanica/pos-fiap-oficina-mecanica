.DEFAULT_GOAL := help

POSTGRES_SERVICE := postgres
POSTGRES_CONTAINER := oficina-postgres
POSTGRES_USER := oficina
POSTGRES_DB := oficina
MIGRATION := db/migrations/V001__schema_inicial.sql
SEED := db/seeds/V900__dados_iniciais.sql

ifeq ($(OS),Windows_NT)
DOCKER ?= docker.exe
else
DOCKER ?= docker
endif

.PHONY: help setup up restart db-up db-down db-reset db-migrate db-seed db-init db-verify

help: ## Lista os comandos disponiveis
	@echo "Uso: make <alvo>"
	@echo "  setup          Prepara e sobe todo o projeto apos o primeiro clone"
	@echo "  up             Sobe os containers para o uso diario"
	@echo "  restart        Reconstroi e reinicia somente a API"
	@echo "  db-up          Inicia o PostgreSQL local"
	@echo "  db-down        Para os containers sem remover os dados"
	@echo "  db-reset       Remove o banco local, cria o schema e carrega o seed"
	@echo "  db-migrate     Aplica a migration no banco vazio"
	@echo "  db-seed        Carrega os dados iniciais apos aplicar a migration"
	@echo "  db-init        Cria um banco local vazio com schema e seed"
	@echo "  db-verify      Exibe contagens essenciais para verificar a carga"

setup: db-init ## Prepara e sobe todo o projeto apos o primeiro clone
	$(DOCKER) compose up -d --build

up: ## Sobe os containers para o uso diario
	$(DOCKER) compose up -d

restart: ## Reconstroi e reinicia somente a API
	$(DOCKER) compose up -d --build app

db-up: ## Inicia o PostgreSQL local
	$(DOCKER) compose up -d --wait $(POSTGRES_SERVICE)

db-down: ## Para os containers sem remover os dados
	$(DOCKER) compose down

db-reset: ## Remove o banco local, cria o schema e carrega o seed
	$(DOCKER) compose down -v
	$(MAKE) db-init

db-migrate: db-up ## Aplica a migration no banco vazio
	$(DOCKER) compose exec -T $(POSTGRES_SERVICE) psql -v ON_ERROR_STOP=1 -U $(POSTGRES_USER) -d $(POSTGRES_DB) < $(MIGRATION)

db-seed: db-up ## Carrega os dados iniciais apos aplicar a migration
	$(DOCKER) compose exec -T $(POSTGRES_SERVICE) psql -v ON_ERROR_STOP=1 -U $(POSTGRES_USER) -d $(POSTGRES_DB) < $(SEED)

db-init: db-migrate db-seed ## Cria um banco local vazio com schema e seed

db-verify: db-up ## Exibe contagens essenciais para verificar a carga
	$(DOCKER) compose exec -T $(POSTGRES_SERVICE) psql -U $(POSTGRES_USER) -d $(POSTGRES_DB) -P pager=off -c "SELECT 'clientes' AS entidade, COUNT(*) AS quantidade FROM cliente UNION ALL SELECT 'veiculos', COUNT(*) FROM veiculo UNION ALL SELECT 'servicos', COUNT(*) FROM servico UNION ALL SELECT 'itens_estoque', COUNT(*) FROM item_estoque UNION ALL SELECT 'ordens_servico', COUNT(*) FROM ordem_servico UNION ALL SELECT 'orcamentos', COUNT(*) FROM orcamento UNION ALL SELECT 'reservas_estoque', COUNT(*) FROM reserva_estoque UNION ALL SELECT 'pedidos_compra', COUNT(*) FROM pedido_compra ORDER BY entidade;"
