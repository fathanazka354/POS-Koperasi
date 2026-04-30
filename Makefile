SHELL := /bin/bash

.PHONY: migrate seed fresh reset run

migrate:
	@bash scripts/db.sh migrate

seed:
	@bash scripts/db.sh seed

fresh:
	@bash scripts/db.sh fresh

reset:
	@bash scripts/db.sh reset

run:
	@go run ./cmd/api
