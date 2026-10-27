.DEFAULT_GOAL := help

APP ?= calc-service

DEV_UID ?= $(shell id -u)
DEV_GID ?= $(shell id -g)

export DEV_UID DEV_GID

.PHONY: help env dev-up dev-down generate build test lint format

help:
	@echo "env        Copy each app's .env.example to .env (fails if it exists)"
	@echo "dev-up     Start all services locally"
	@echo "dev-down   Stop local development"
	@echo "generate/build/test/lint/format  Run a service's Make target in Docker (APP=calc-service)"

env:
	@$(MAKE) --no-print-directory -C apps/calc-service env
	@$(MAKE) --no-print-directory -C apps/calc-web env

dev-up:
	@test -f apps/calc-service/.env || { echo "Copy apps/calc-service/.env.example to apps/calc-service/.env first"; exit 1; }
	@test -f apps/calc-web/.env || { echo "Copy apps/calc-web/.env.example to apps/calc-web/.env first"; exit 1; }

	docker compose up -d

dev-down:
	docker compose down

generate build test lint format:
	docker compose run --rm --no-deps -T "$(APP)" make $@
