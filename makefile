.PHONY: migrate-up infra-up infra-down

migrate-up:
	scripts/migrate-up.sh

migrate-down:
	scripts/migrate-down.sh

infra-up:
	docker compose up -d

infra-down:
	docker compose down

run:
	scripts/run.sh