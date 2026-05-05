SHELL := /bin/bash

GOPATH_BIN := $(shell go env GOPATH)/bin
AIR := $(shell command -v air 2>/dev/null)
ifeq ($(AIR),)
  AIR := $(GOPATH_BIN)/air
endif

.PHONY: migrate seed fresh reset run run-once test test-race test-cover stress-smoke stress-k6

migrate:
	@bash scripts/db.sh migrate

seed:
	@bash scripts/db.sh seed

fresh:
	@bash scripts/db.sh fresh

reset:
	@bash scripts/db.sh reset

# Live reload saat file .go/.html berubah (perlu: go install github.com/air-verse/air@latest)
run:
	@test -x "$(AIR)" && exec "$(AIR)" || \
		( printf '%s\n' 'air tidak terpasang. Pasang: go install github.com/air-verse/air@latest' >&2; \
		  printf '%s\n' 'Menjalankan sekali tanpa watch…' >&2; \
		  exec go run ./cmd/api )

# Tanpa watch (untuk CI / debugging)
run-once:
	@go run ./cmd/api

# Unit test (jalankan di root modul; tanpa .env lebih deterministik di CI)
test:
	@go test ./... -count=1

test-race:
	@go test ./... -race -count=1

test-cover:
	@go test ./... -count=1 -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -8

# Smoke: butuh server jalan (make run-once). Hanya curl, tanpa alat eksternal.
stress-smoke:
	@bash scripts/load-smoke.sh

# k6: beban terkendali (scripts/k6/load.js).
# Urutan: binary k6 di PATH → jika tidak ada, pakai Docker (grafana/k6).
# Contoh: K6_PROFILE=smoke make stress-k6
# Checkout (token otomatis): K6_MEMBER_CODE=MBR001 K6_MEMBER_PHONE=081234567890 K6_CHECKOUT_RATIO=0.05 make stress-k6
# API di mesin host + Docker: default BASE_URL=http://host.docker.internal:8080 (override dengan BASE_URL=...)
stress-k6:
	@mkdir -p tmp
	@if command -v k6 >/dev/null 2>&1; then \
		k6 run --summary-export=tmp/k6-summary.json scripts/k6/load.js; \
	elif command -v docker >/dev/null 2>&1; then \
		K6_URL="$${BASE_URL:-http://host.docker.internal:8080}"; \
		printf '%s\n' "k6 tidak di PATH — memakai Docker. BASE_URL=$$K6_URL"; \
		docker run --rm \
			--add-host=host.docker.internal:host-gateway \
			-e BASE_URL="$$K6_URL" \
			-e K6_PROFILE \
			-e K6_SHOP_TOKEN \
			-e K6_MEMBER_CODE \
			-e K6_MEMBER_PHONE \
			-e K6_CHECKOUT_RATIO \
			-e K6_CHECKOUT_PRODUCT_ID \
			-e K6_CHECKOUT_QTY \
			-e K6_CHECKOUT_PAY_METHOD \
			-e K6_CHECKOUT_ADDRESS_ID \
			-e K6_CHECKOUT_VOUCHER_CODE \
			-v "$$(pwd):/work" -w /work \
			grafana/k6:latest run \
			--summary-export=/work/tmp/k6-summary.json \
			/work/scripts/k6/load.js; \
	else \
		printf '%s\n' 'Pasang k6: brew install k6   atau pasang Docker lalu jalankan ulang make stress-k6' >&2; \
		exit 1; \
	fi
