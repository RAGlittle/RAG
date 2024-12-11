package gateway

import (
	"context"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/embeddings"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/embed"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/ollama/ollama/api"
	"github.com/tmc/langchaingo/llms/ollama"
)

// TODO : validate method
// TODO : parse from json/yaml
type GatewayConfig struct {
	GatewayAddress string
	Version        string

	ModelName string `yaml:"model_name"`

	// TODO : extend to multiple endpoints
	EmbeddingConfig  embed.EmbedderConfig `yaml:"embedding_config"`
	ReRankerEndpoint string               `yaml:"reranker_endpoint"`
	OllamaURL        string               `yaml:"ollama_url"`

	// TODO : refactor into struct
	MilvusEndpoint string `yaml:"milvus_endpoint"`
	MilvusPassword string `yaml:"milvus_password"`
	MilvusKey      string `yaml:"milvus_key"`

	EtcdEndpoints []string `yaml:"etcd_endpoints"`
}

// TODO : validate method
type GatewayRuntimeConfig struct {
	LLMModel        *ollama.LLM
	OllamaClient    *api.Client
	VectorConfig    client.Config
	EmbeddingClient embeddings.Embedder[float32]
	ReRankerClient  embeddings.ReRanker
	GwContext       context.Context
	ChatKV          storage.KeyValueStore
}
