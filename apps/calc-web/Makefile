.DEFAULT_GOAL := help

.PHONY: help env generate build test lint format debug

help:
	@echo "env       Copy .env.example to .env (fails if it exists)"
	@echo "generate  Not implemented; no code generation is configured"
	@echo "build     Type-check and build the frontend for production"
	@echo "test      Not implemented; no test runner is configured"
	@echo "lint      Lint frontend code with ESLint"
	@echo "format    Format frontend code with Prettier"
	@echo "debug     Run the Vite development server"

env:
	@test ! -e .env && test ! -L .env || { echo ".env already exists"; exit 1; }
	cp .env.example .env

generate:
	@echo "Not implemented: no code generation is configured"

build:
	pnpm run build

test:
	@echo "Not implemented: no test runner is configured"

lint:
	pnpm run lint

format:
	pnpm run format

debug:
	pnpm run dev --host 0.0.0.0
