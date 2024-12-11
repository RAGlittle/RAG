package retriever

import (
	"context"
	"path"
	"slices"
	"sync"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/embeddings"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores"
	"golang.org/x/sync/errgroup"
)

type Retriever interface {
	Name() string
	// Weight is the factor used in reciprocal rank fusion
	Weight() float64
	// Priority decides tie breaks in reciprocal rank fusion
	Priority() int
	AsRetriever() schema.Retriever
}

// SimilarityRetriever uses dense vector similarity serach to retrieve documents
// for semantic search across documents
type SimilarityRetriever struct {
	*Options
	vectorStore vectorstores.VectorStore
}

func NewSimilarityRetriever(
	vectorStore vectorstores.VectorStore,
	opts ...Option,
) *SimilarityRetriever {
	opt := &Options{}
	opt.Apply(opts...)
	return &SimilarityRetriever{
		Options:     opt,
		vectorStore: vectorStore,
	}
}

func (s *SimilarityRetriever) Name() string {
	return "similarity"
}

func (s *SimilarityRetriever) AsRetriever() schema.Retriever {
	return s
}

func (s *SimilarityRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	k := s.K
	res, err := s.vectorStore.SimilaritySearch(ctx, query, k)
	if err != nil {
		return nil, err
	}
	sortDocsByScore(res)
	return res, nil
}

var _ schema.Retriever = (*SimilarityRetriever)(nil)
var _ Retriever = (*SimilarityRetriever)(nil)

// RerankRetriever uses a semantic reranker to rerank documents retrieved by another retriever
type RerankerRetriever struct {
	*Options
	baseRetriever Retriever
	reranker      embeddings.ReRanker
}

func NewReRankerRetriever(
	baseRetriever Retriever,
	reranker embeddings.ReRanker,
	opts ...Option,
) *RerankerRetriever {
	opt := &Options{}
	opt.Apply(opts...)
	return &RerankerRetriever{
		Options:       opt,
		baseRetriever: baseRetriever,
		reranker:      reranker,
	}
}

func (r *RerankerRetriever) Name() string {
	return path.Join(r.baseRetriever.Name(), "reranker")
}

func (r *RerankerRetriever) AsRetriever() schema.Retriever {
	return r
}

func (r *RerankerRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	docs, err := r.baseRetriever.AsRetriever().GetRelevantDocuments(ctx, query)
	if err != nil {
		return nil, err
	}
	texts := lo.Map(
		docs,
		func(c schema.Document, _ int) string {
			return c.PageContent
		},
	)

	textLen := lo.Reduce(texts, func(agg int, t string, _ int) int {
		return agg + len(t)
	}, 0)

	logrus.Infof("submitting %d documents for reranking, total text length : %d", len(texts), textLen)
	resp, err := r.reranker.RankDocuments(ctx, query, texts)
	if err != nil {
		return nil, err
	}

	for _, r := range resp {
		// FIXME: check if clamping is safe here
		docs[r.Index].Score = float32(r.Score)
	}
	sortDocsByScore(docs)
	return docs, nil
}

// SparseRetriever uses sparse vector similarity search to retrieve documents
type SparseRetriever struct {
	*Options
}

func sortDocsByScore(docs []schema.Document) {
	slices.SortFunc(docs, func(i, j schema.Document) int {
		return int(j.Score*1000 - i.Score*1000)
	})
}

// HybridRetriever fetches documents using multiple retrievers and merges the results
// using a fusion algorithm
type HybridRetriever struct {
	baseRetrievers []Retriever
	k              int
	fuser          Fusion
	maxWorkers     int
}

func NewHybridRetriever(
	baseRetrievers []Retriever,
	k int,
	fuser Fusion,
) *HybridRetriever {
	return &HybridRetriever{
		baseRetrievers: baseRetrievers,
		k:              k,
		fuser:          fuser,
		maxWorkers:     12,
	}
}

func (h *HybridRetriever) Name() string {
	return "hybrid"
}

func (h *HybridRetriever) AsRetriever() schema.Retriever {
	return h
}

func (h *HybridRetriever) GetRelevantDocuments(ctx context.Context, query string) ([]schema.Document, error) {
	mergeDocs := [][]schema.Document{}
	eg, ctx := errgroup.WithContext(ctx)

	eg.SetLimit(h.maxWorkers)
	var mu sync.Mutex

	for _, retriever := range h.baseRetrievers {
		retriever := retriever
		eg.Go(func() error {
			docs, err := retriever.AsRetriever().GetRelevantDocuments(ctx, query)
			if err != nil {
				return err
			}
			mu.Lock()
			defer mu.Unlock()
			mergeDocs = append(mergeDocs, docs)
			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return h.fuser.Merge(mergeDocs, h.k), nil
}
