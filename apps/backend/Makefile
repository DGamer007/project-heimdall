WEB_ENTRYPOINT="./cmd/www.go"
WEB_OUTPUT="./bin/www"

ENV_FILE ?= .env
include $(ENV_FILE)
export

build:
	go build -o ${WEB_OUTPUT} ${WEB_ENTRYPOINT}

run: build
	${WEB_OUTPUT}

debug:
	dlv debug ${WEB_ENTRYPOINT} --continue --headless --listen=:2345 --api-version=2 --accept-multiclient --output ./tmp/debug_bin

pg-goose:
	GOOSE_DRIVER="postgres" \
	GOOSE_DBSTRING="postgres://${DATABASE_POSTGRES_USER}:${DATABASE_POSTGRES_PASSWORD}@${DATABASE_POSTGRES_HOST}:${DATABASE_POSTGRES_PORT}/${DATABASE_POSTGRES_DB_NAME}" \
	GOOSE_MIGRATION_DIR="./internal/infra/postgres/migrations" \
	GOOSE_TABLE="public.goose_migrations" \
	goose $(filter-out $@,$(MAKECMDGOALS))

web-e2e-test:
	./scripts/web-e2e-test.sh

clean:
	go clean
	rm -f ${WEB_OUTPUT}

# Prevent make from trying to run the arguments as targets
%:
	@:
