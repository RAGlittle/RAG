.PHONY: prereqs
.PHONY: load
.PHONY : up
.PHONY: down

prereqs: check-ollama

check-ollama:
	@command -v ollama --version >/dev/null 2>&1 || { echo >&2 "ollama is not installed. Aborting, run `curl -fsSL https://ollama.com/install.sh | sh`"; exit 1; }

install-requirements: prereqs install-deps install-models


install-models:
	ollama pull nomic-embed-text
	ollama pull mistral


up:
	docker compose -f ./services/vectorstore/docker-compose.yml up -d
	docker compose -f ./services/embedding/docker-compose.yml up -d
	docker compose -f ./services/storage/docker-compose.yml up -d

down:
	docker compose -f ./services/vectorstore/docker-compose.yml down
	docker compose -f ./services/embedding/docker-compose.yml down
	docker compose -f ./services/storage/docker-compose.yml down

images:
	docker build . -f ./package/gateway/Dockerfile -t rag-gateway:dev

snapshot:
	goreleaser release --snapshot --clean --config=./rag-gateway/.goreleaser.yaml