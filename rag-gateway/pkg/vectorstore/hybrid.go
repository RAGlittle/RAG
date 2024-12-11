package vectorstore

import (
	"context"

	"github.com/samber/lo"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
)

// TODO : this a naive implementation
type hybridVectorstore struct {
	baseVectorStores []vectorstores.VectorStore
}

func NewHybridVectorstore(baseVectorStores ...vectorstores.VectorStore) *hybridVectorstore {
	return &hybridVectorstore{
		baseVectorStores: baseVectorStores,
	}
}

var _ vectorstores.VectorStore = &hybridVectorstore{}

func (h *hybridVectorstore) AddDocuments(ctx context.Context, docs []schema.Document, options ...vectorstores.Option) ([]string, error) {
	seenDocs := map[string]struct{}{}
	for _, vecS := range h.baseVectorStores {
		docs, err := vecS.AddDocuments(ctx, docs, options...)
		if err != nil {
			return nil, err
		}
		for _, doc := range docs {
			seenDocs[doc] = struct{}{}
		}
	}
	return lo.Keys(seenDocs), nil
}

func (h *hybridVectorstore) SimilaritySearch(ctx context.Context, query string, numDocuments int, options ...vectorstores.Option) ([]schema.Document, error) {
	schemaDocs := []schema.Document{}
	for _, vecS := range h.baseVectorStores {
		docs, err := vecS.SimilaritySearch(ctx, query, numDocuments, options...)
		if err != nil {
			return nil, err
		}
		schemaDocs = append(schemaDocs, docs...)
	}
	return schemaDocs, nil
}
