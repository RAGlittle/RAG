package gateway

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/Synaptic-Lynx/rag-gateway/api/tei"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/embeddings"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/etcd"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	ollama_api "github.com/ollama/ollama/api"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms/ollama"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// FIXME: hack for in memory broker
type hackBroker struct {
	client *clientv3.Client
}

func (h hackBroker) KeyValueStore(prefix string) storage.KeyValueStore {
	return etcd.NewKeyValueStore(h.client, prefix)
}

func setup(ctx context.Context, config GatewayConfig) (GatewayRuntimeConfig, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: config.EtcdEndpoints,
	})
	if err != nil {
		return GatewayRuntimeConfig{}, fmt.Errorf("failed to create etcd client: %s", err)
	}

	logrus.Info("Setting up embedding client...")
	eEndpoint := config.EmbeddingConfig.Embedders[0].Endpoint

	logrus.Infof("Embedding endpoint is %s", eEndpoint)

	ccE, err := grpc.NewClient(eEndpoint, grpc.WithTransportCredentials(
		insecure.NewCredentials(),
	))
	if err != nil {
		return GatewayRuntimeConfig{}, fmt.Errorf("failed to create grpc client for embedder: %s", err)
	}

	embedInfoClient := tei.NewInfoClient(ccE)
	embedEmbedClient := tei.NewEmbedClient(ccE)

	baseEmbedder := embeddings.NewGrpcProxyEmbedder(embedInfoClient, embedEmbedClient, false)

	emClient := embeddings.NewPoolEmbedder(eEndpoint, embeddings.WithDebug(), embeddings.WithEmbedder(baseEmbedder))
	logrus.Info("Set up embedding client")

	logrus.Info("Setting up reranker client...")

	ccR, err := grpc.NewClient(config.ReRankerEndpoint, grpc.WithTransportCredentials(
		insecure.NewCredentials(),
	))
	if err != nil {
		return GatewayRuntimeConfig{}, fmt.Errorf("failed to create grpc client for reranker: %s", err)
	}

	reRankInfoClient := tei.NewInfoClient(ccR)
	reRankRankClient := tei.NewRerankClient(ccR)

	baseReRanker := embeddings.NewGrpcProxyReRanker(reRankInfoClient, reRankRankClient)

	reClient := embeddings.NewPoolReRanker(config.ReRankerEndpoint, embeddings.WithReRanker(baseReRanker))

	logrus.Info("Set up reranker client")
	milvusConfig := client.Config{
		Address:  config.MilvusEndpoint,
		Password: config.MilvusPassword,
		APIKey:   config.MilvusKey,
	}
	logrus.Info("Setting up ollama client...")
	ollamaURL, err := url.Parse(config.OllamaURL)
	if err != nil {
		return GatewayRuntimeConfig{}, fmt.Errorf("failed to parse ollama URL")
	}
	ollamaClient := ollama_api.NewClient(ollamaURL, http.DefaultClient)
	logrus.Info("Set up ollama client")

	logrus.Info("Setting up LLM model...")
	llm, err := ollama.New(ollama.WithModel(config.ModelName))
	if err != nil {
		logrus.Fatalf("failed to create LLM %s: %v", config.ModelName, err)
	}
	logrus.Info("Set up LLM model")

	broker := hackBroker{
		client: etcdClient,
	}
	chatKV := broker.KeyValueStore("chats")
	return GatewayRuntimeConfig{
		VectorConfig:    milvusConfig,
		EmbeddingClient: emClient,
		ReRankerClient:  reClient,
		GwContext:       ctx,
		OllamaClient:    ollamaClient,
		LLMModel:        llm,
		ChatKV:          chatKV,
	}, nil
}
