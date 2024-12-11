.PHONY : build
.PHONY : mockgen

GO_BUILD_FLAGS :-v 
VERSION ?=	$(shell git describe --tags --abbrev=0 || echo "v0.0.0-dev") 
COMMIT ?= $(shell git rev-parse --short HEAD)

LINK_FLAGS := "-extldflags -static -s -w -X github.com/Synaptic-Lynx/rag-gateway/pkg/version.Version=$(VERSION) -X github.com/Synaptic-Lynx/rag-gateway/pkg/version.GitCommit=$(COMMIT)"

build:
	echo VERSION=$(VERSION) COMMIT=$(COMMIT)
	go build -ldflags $(LINK_FLAGS) $(GO_BUILD_FLAGS) -o rdl ./cmd/dataloader
	go build -ldflags $(LINK_FLAGS) $(GO_BUILD_FLAGS) -o ragger ./cmd/gateway

# requires go install go.uber.org/mock/mockgen@latest
mockgen:
	mockgen -destination=pkg/test/mock/vectorstores/mock_vectorstore.go github.com/tmc/langchaingo/vectorstores VectorStore
	mockgen -destination=pkg/test/mock/llms/mock_model.go github.com/tmc/langchaingo/llms Model
	mockgen -source=./pkg/retriever/retriever.go -destination=pkg/test/mock/retriever/mock_retriever.go  Retriever
	mockgen -destination=pkg/test/mock/schema/mock_retriever.go github.com/tmc/langchaingo/schema Retriever