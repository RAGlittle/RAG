package embeddings

import "github.com/alitto/pond"

const (
	maxWorkers  = 16
	maxCapacity = 1000
)

var (
	strategy = pond.Balanced()
)

type ReRankerOptions struct {
	reranker ReRanker
	*PoolOptions
}

type ReRankerOption func(*ReRankerOptions)

func WithReRanker(r ReRanker) ReRankerOption {
	return func(o *ReRankerOptions) {
		o.reranker = r
	}
}

func defaultReRankerOptions() *ReRankerOptions {
	return &ReRankerOptions{
		PoolOptions: defaultPoolOptions(),
	}
}

type EmbeddingClientOptions struct {
	*PoolOptions
	embedder Embedder[float32]
	Debug    bool
}

type EmbeddingClientOption func(*EmbeddingClientOptions)

func defaultEmbeddingClientOptions() *EmbeddingClientOptions {
	return &EmbeddingClientOptions{
		PoolOptions: defaultPoolOptions(),
		Debug:       false,
	}

}

func WithEmbedder(e Embedder[float32]) EmbeddingClientOption {
	return func(o *EmbeddingClientOptions) {
		o.embedder = e
	}
}

func WithDebug() EmbeddingClientOption {
	return func(o *EmbeddingClientOptions) {
		o.Debug = true
	}
}

type PoolOptions struct {
	maxWorkers  int
	maxCapacity int
	strategy    pond.ResizingStrategy
}

type PoolOption func(*PoolOptions)

func defaultPoolOptions() *PoolOptions {
	return &PoolOptions{
		maxWorkers:  maxWorkers,
		maxCapacity: maxCapacity,
		strategy:    strategy,
	}
}

func WithMaxWorkers(maxWorkers int) PoolOption {
	return func(o *PoolOptions) {
		o.maxWorkers = maxWorkers
	}
}

func WithMaxCapacity(maxCapacity int) PoolOption {
	return func(o *PoolOptions) {
		o.maxCapacity = maxCapacity
	}
}

func WithStrategy(strategy pond.ResizingStrategy) PoolOption {
	return func(o *PoolOptions) {
		o.strategy = strategy
	}
}

func (p *PoolOptions) newPool() *pond.WorkerPool {
	return pond.New(p.maxWorkers, p.maxCapacity, pond.Strategy(p.strategy))
}
