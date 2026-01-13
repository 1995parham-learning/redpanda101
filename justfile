default:
    @just --list

# build koochooloo binary
build:
    @echo '{{ BOLD + CYAN }}Building Redpanda 101!{{ NORMAL }}'
    go build -o redpanda101 ./cmd/redpanda101/main.go

# update go packages
update:
    @cd ./cmd/redpanda101 && go get -u

# run golang-migrate
migrate *flags:
    go tool github.com/golang-migrate/migrate/v4/cmd/migrate {{ flags }}

# create new migration
migrate-new name:
    @just migrate create -dir migrations -ext sql {{ name }}

# set up the dev environment with docker-compose
dev cmd *flags:
    #!/usr/bin/env bash
    echo '{{ BOLD + YELLOW }}Development environment based on docker-compose{{ NORMAL }}'
    set -euxo pipefail
    if [ {{ cmd }} = 'down' ]; then
      docker compose -f ./deployments/docker-compose.yml down
      docker compose -f ./deployments/docker-compose.yml rm
    elif [ {{ cmd }} = 'up' ]; then
      docker compose -f ./deployments/docker-compose.yml up --wait -d {{ flags }}
    else
      docker compose -f ./deployments/docker-compose.yml {{ cmd }} {{ flags }}
    fi

# run golangci-lint
lint:
    golangci-lint run -c .golangci.yml --fix

# run producer (HTTP API + Kafka producer)
produce:
    go run ./cmd/redpanda101 -c configs/producer.toml produce

# run consumer (Kafka consumer + PostgreSQL)
consume:
    go run ./cmd/redpanda101 -c configs/consumer.toml consume

# run database migrations
db-migrate:
    go run ./cmd/redpanda101 migrate

# create redpanda topic
redpanda-migrate:
    docker exec -it deployments-redpanda-1 rpk topic create orders --brokers localhost:9092

# run k6 load test
loadtest:
    k6 run api/k6/script.js

# setup and run everything (infrastructure + migrations)
setup:
    @echo '{{ BOLD + GREEN }}Setting up Redpanda101 environment{{ NORMAL }}'
    just dev up
    just redpanda-migrate
    just db-migrate
    @echo '{{ BOLD + GREEN }}Setup complete! Run "just produce" and "just consume" in separate terminals{{ NORMAL }}'
