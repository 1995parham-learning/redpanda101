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
      docker compose -f ./docker-compose.yml down
      docker compose -f ./docker-compose.yml rm
    elif [ {{ cmd }} = 'up' ]; then
      docker compose -f ./docker-compose.yml up --wait -d {{ flags }}
    else
      docker compose -f ./docker-compose.yml {{ cmd }} {{ flags }}
    fi

# run golangci-lint
lint:
    golangci-lint run -c .golangci.yml --fix
