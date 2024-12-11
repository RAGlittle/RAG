package collection

import (
	"fmt"

	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/milvus"
)

var _ VectorStoreFactory = &sectionManager{}

type VectorStoreFactory interface {
	ToVectorStore(section EmbeddingSpec) (vectorstores.VectorStore, error)
}

func (s *sectionManager) ToVectorStore(spec EmbeddingSpec) (vectorstores.VectorStore, error) {
	embedder, err := s.embeddingGw.AsEmbedder(spec.EmbeddingID)
	if err != nil {
		return nil, err
	}

	metricTypeF, ok := MetricMappings[spec.MetricType]
	if !ok {
		return nil, fmt.Errorf("invalid metric type: %s", spec.MetricType)
	}

	indexF, ok := IndexMappings[spec.IndexId]
	if !ok {
		return nil, fmt.Errorf("invalid index id: %s", spec.IndexId)
	}

	index, err := indexF(metricTypeF())
	if err != nil {
		return nil, fmt.Errorf("error creating index: %s", err)
	}
	return milvus.New(
		s.parentCtx,
		s.vectorStoreConfig,
		milvus.WithCollectionName(spec.CollectionId),
		milvus.WithIndex(index),
		milvus.WithPartitionName(spec.PartitionId),
		milvus.WithEmbedder(embedder),
	)
}
