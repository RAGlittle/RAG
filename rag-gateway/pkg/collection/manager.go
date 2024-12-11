package collection

import (
	"context"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/embed"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/minio/minio-go/v7"
)

type SectionManager interface {
	VectorStoreFactory
	SectionObjectStore
	SectionMetadataStore
}

type sectionManager struct {
	parentCtx    context.Context
	objectClient *minio.Client

	docMetadataKv     storage.KeyValueStore
	vectorStoreConfig client.Config
	embeddingGw       *embed.EmbedGatewayServer
}

func NewSectionManager(
	ctx context.Context,
	objectClient *minio.Client,
	vectorStoreConfig client.Config,
	metadataKV storage.KeyValueStore,
	embedGw *embed.EmbedGatewayServer,
) SectionManager {
	return &sectionManager{
		parentCtx:         ctx,
		objectClient:      objectClient,
		vectorStoreConfig: vectorStoreConfig,
		docMetadataKv:     metadataKV,
		embeddingGw:       embedGw,
	}
}
