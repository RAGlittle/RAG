package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	"github.com/Synaptic-Lynx/rag-gateway/api/tei"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/embeddings"
	"google.golang.org/protobuf/encoding/protojson"
)

type floatType interface {
	float32 | float64
}

// Implements langchaingo.embedding.Embedder interface
type Embedder[T floatType] interface {
	Health(ctx context.Context) (bool, error)
	Info(ctx context.Context) ([]byte, error)
	// EmbedDocuments returns a vector for each text.
	EmbedDocuments(ctx context.Context, texts []string) ([][]T, error)
	// EmbedQuery embeds a single text.
	EmbedQuery(ctx context.Context, text string) ([]T, error)
}

// PoolEmbedder pools and aggregates embedding requests to an embedder
type PoolEmbedder struct {
	endpoint   string
	httpClient *http.Client
	*EmbeddingClientOptions
}

var _ embeddings.Embedder = (*PoolEmbedder)(nil)

func NewPoolEmbedder(endpoint string, opts ...EmbeddingClientOption) Embedder[float32] {

	opt := defaultEmbeddingClientOptions()
	for _, o := range opts {
		o(opt)
	}

	if opt.embedder == nil {
		opt.embedder = NewProxyEmbedder(endpoint, http.DefaultClient, opt.Debug)
	}

	return &PoolEmbedder{
		endpoint:               endpoint,
		httpClient:             http.DefaultClient,
		EmbeddingClientOptions: opt,
	}
}

func (e *PoolEmbedder) EmbedQuery(
	ctx context.Context,
	content string,
) ([]float32, error) {
	embeddings, err := e.embedder.EmbedQuery(ctx, content)
	if err != nil {
		logrus.Errorf("error embedding content: %s", err)
		return nil, err
	}
	return embeddings, nil
}

func (e *PoolEmbedder) EmbedDocuments(
	ctx context.Context,
	contents []string,
) ([][]float32, error) {
	logrus.Infof("submitting %d documents for embedding", len(contents))

	// due to how langchain go works, we need to handle the base case where len(contents) = 1
	if len(contents) == 1 {
		embeddings, err := e.embedder.EmbedQuery(ctx, contents[0])
		if err != nil {
			return nil, err
		}
		return [][]float32{embeddings}, nil
	}

	pool := e.newPool()
	yieldedVectors := make(chan lo.Tuple2[int, [][]float32], len(contents))
	go func() {
		for i, content := range contents {
			i := i
			content := content
			pool.Submit(func() {
				embeddings, err := e.embedder.EmbedQuery(ctx, content)
				if err != nil {
					logrus.Errorf("error embedding content: %s", err)
					return
				}
				yieldedVectors <- lo.T2(i, [][]float32{embeddings})
			})
		}

		pool.StopAndWait()
		close(yieldedVectors)
	}()
	allEmbeddings := []lo.Tuple2[int, [][]float32]{}
	for {
		select {
		case v, ok := <-yieldedVectors:
			if !ok {
				logrus.Infof("Submitted %d documents for embedding ", len(allEmbeddings))
				return e.flatten(allEmbeddings), nil
			}
			allEmbeddings = append(allEmbeddings, v)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

}

func (e *PoolEmbedder) flatten(allEmbeddings []lo.Tuple2[int, [][]float32]) [][]float32 {
	// sort by index
	sort.Slice(allEmbeddings, func(i, j int) bool {
		return allEmbeddings[i].A < allEmbeddings[j].A
	})

	flattenedEmbeddings := [][]float32{}
	for _, v := range allEmbeddings {
		flattenedEmbeddings = append(flattenedEmbeddings, v.B...)
	}
	return flattenedEmbeddings
}

func (e *PoolEmbedder) Info(ctx context.Context) ([]byte, error) {
	return e.embedder.Info(ctx)
}

func (e *PoolEmbedder) Health(ctx context.Context) (bool, error) {
	return e.embedder.Health(ctx)
}

type ProxyEmbedder struct {
	endpoint   string
	httpClient *http.Client
	Debug      bool
}

func NewProxyEmbedder(endpoint string, client *http.Client, debug bool) Embedder[float32] {
	return &ProxyEmbedder{
		httpClient: client,
		endpoint:   endpoint,
	}
}

func (p *ProxyEmbedder) Health(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		p.endpoint+"/health",
		nil,
	)
	if err != nil {
		return false, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, nil
	}
	defer resp.Body.Close()
	return true, nil
}

func (p *ProxyEmbedder) Info(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		p.endpoint+"/info",
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New(resp.Status)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return data, nil

}

func (p *ProxyEmbedder) EmbedDocuments(
	ctx context.Context,
	contents []string,
) ([][]float32, error) {
	panic("unimplemented by design")

}

func (p *ProxyEmbedder) EmbedQuery(
	ctx context.Context,
	content string,
) ([]float32, error) {
	req, err := p.newEmbedRequest(ctx, content)
	if err != nil {
		logrus.Errorf("error creating request: %s", err)
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		logrus.Errorf("request to embed content failed : %s", err)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("request to embed content failed : %s", resp.Status)
	}
	defer resp.Body.Close()

	var embeddings [][]float32

	if err := json.NewDecoder(resp.Body).Decode(&embeddings); err != nil {
		logrus.Errorf("error decoding response: %s", err)
		return nil, err
	}

	ret := embeddings[0]
	if p.Debug {
		checkEmpty(ret)
	}
	return ret, nil
}

func checkEmpty[T floatType](vector []T) {
	for _, v := range vector {
		if v != 0 {
			return
		}
	}
	logrus.Warn("got empty vector from embedding response")
}

func (p *ProxyEmbedder) newEmbedRequest(
	ctx context.Context,
	content string,
) (*http.Request, error) {
	type EmbedRequest struct {
		Inputs string `json:"inputs"`
	}

	body := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(body).Encode(EmbedRequest{
		Inputs: content,
	}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		p.endpoint+"/embed",
		body,
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	return req, nil
}

type GrpcProxyEmbedder struct {
	tei.InfoClient
	tei.EmbedClient

	WithSparse bool
}

var _ Embedder[float32] = (*GrpcProxyEmbedder)(nil)

func NewGrpcProxyEmbedder(infoClient tei.InfoClient, embedClient tei.EmbedClient, sparse bool) Embedder[float32] {
	return &GrpcProxyEmbedder{
		InfoClient:  infoClient,
		EmbedClient: embedClient,
	}
}

func (g *GrpcProxyEmbedder) Health(ctx context.Context) (bool, error) {
	_, err := g.InfoClient.Info(ctx, &tei.InfoRequest{})
	if err != nil {
		logrus.Errorf("error getting health from embedding server: %s", err)
		return false, err
	}
	return true, nil
}

func (g *GrpcProxyEmbedder) Info(ctx context.Context) ([]byte, error) {
	resp, err := g.InfoClient.Info(ctx, &tei.InfoRequest{})
	if err != nil {
		logrus.Errorf("error getting info from embedding server: %s", err)
		return nil, err
	}

	data, err := protojson.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (g *GrpcProxyEmbedder) EmbedDocuments(
	ctx context.Context,
	contents []string,
) ([][]float32, error) {
	panic("unimplemented by design")

}

func (g *GrpcProxyEmbedder) embed(
	ctx context.Context,
	content string,
) ([]float32, error) {
	resp, err := g.EmbedClient.Embed(ctx, &tei.EmbedRequest{
		Inputs: content,
	})
	if err != nil {
		return nil, err
	}
	return resp.Embeddings, nil
}

func (g *GrpcProxyEmbedder) embedSparse(
	ctx context.Context,
	content string,
) ([]float32, error) {
	resp, err := g.EmbedClient.EmbedSparse(ctx, &tei.EmbedSparseRequest{
		Inputs: content,
	})
	if err != nil {
		return nil, err
	}

	return lo.Map(resp.SparseEmbeddings, func(se *tei.SparseValue, _ int) float32 {
		return se.Value
	}), nil

}

func (g *GrpcProxyEmbedder) EmbedQuery(
	ctx context.Context,
	content string,
) ([]float32, error) {
	if !g.WithSparse {
		return g.embed(ctx, content)
	}
	return g.embedSparse(ctx, content)
}
