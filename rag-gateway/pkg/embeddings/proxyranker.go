package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/Synaptic-Lynx/rag-gateway/api/tei"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	payloadLimit = 20000
)

type ReRanker interface {
	Health(ctx context.Context) (bool, error)
	Info(ctx context.Context) ([]byte, error)
	RankDocuments(ctx context.Context, query string, texts []string) ([]*ReRankResponse, error)
}

type ReRankResponse struct {
	Score float64 `json:"score"`
	Index int     `json:"index"`
}

type reRankRequest struct {
	Query string   `json:"query"`
	Texts []string `json:"texts"`
}

type PoolReranker struct {
	endpoint   string
	httpClient *http.Client
	*ReRankerOptions
}

var _ ReRanker = (*PoolReranker)(nil)

func NewPoolReRanker(endpoint string, opts ...ReRankerOption) ReRanker {

	opt := defaultReRankerOptions()
	for _, o := range opts {
		o(opt)
	}
	if opt.reranker == nil {
		opt.reranker = NewProxyReranker(endpoint, http.DefaultClient)
	}
	return &PoolReranker{
		endpoint:        endpoint,
		httpClient:      http.DefaultClient,
		ReRankerOptions: opt,
	}
}

func (r *PoolReranker) Health(ctx context.Context) (bool, error) {
	return r.reranker.Health(ctx)
}

func (r *PoolReranker) Info(ctx context.Context) ([]byte, error) {
	return r.reranker.Info(ctx)
}

type rerankPipeResponse struct {
	Offset int
	Score  float64
	Index  int
}

func (r *PoolReranker) RankDocuments(ctx context.Context, query string, texts []string) ([]*ReRankResponse, error) {
	logrus.Infof("Reranking %d documents", len(texts))

	pool := r.newPool()
	batchTexts := SplitTexts(texts, payloadLimit/2)
	yieldedScores := make(chan []rerankPipeResponse, len(batchTexts))
	batchedScores := [][]rerankPipeResponse{}
	logrus.Infof("length of batches : %d", len(batchTexts))
	go func() {
		curOffset := 0
		for _, textChunk := range batchTexts {
			textChunk := textChunk
			logrus.Infof("Submitting %d documents for reranking ", len(textChunk))
			offset := curOffset
			pool.Submit(func() {
				reranks, err := r.reranker.RankDocuments(ctx, query, textChunk)
				if err != nil {
					logrus.Errorf("error reranking documents: %s", err)
					return
				}

				res := []rerankPipeResponse{}
				for _, rerank := range reranks {
					res = append(res, rerankPipeResponse{
						Offset: offset,
						Score:  rerank.Score,
						Index:  rerank.Index,
					})
				}
				yieldedScores <- res
			})
			curOffset += len(textChunk)
		}
		pool.StopAndWait()
		close(yieldedScores)
	}()
	for {
		select {
		case scoreBatch, ok := <-yieldedScores:
			if !ok {
				res := ReduceScoreBatch(batchedScores)
				logrus.Infof("Reranked %d documents", len(res))
				return res, nil
			}
			batchedScores = append(batchedScores, scoreBatch)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// ReduceScoreBatch reduces the score batches into a single slice of ReRankResponse, sorted by scores
func ReduceScoreBatch(scoreBatches [][]rerankPipeResponse) []*ReRankResponse {
	res := []*ReRankResponse{}
	for _, scoreBatch := range scoreBatches {
		for _, score := range scoreBatch {
			res = append(res, &ReRankResponse{
				Score: score.Score,
				Index: score.Index + score.Offset,
			})
		}
	}
	return res
}

// Split texts once the total length of texts exceeds the limit.
func SplitTexts(texts []string, limit int) [][]string {
	cur := 0
	res := [][]string{}
	for _, text := range texts {
		cur += len(text)
		if cur > limit {
			res = append(res, []string{text})
			cur = len(text)
		} else {
			if len(res) == 0 {
				res = append(res, []string{text})
			} else {
				res[len(res)-1] = append(res[len(res)-1], text)
			}
		}
	}
	return res
}

type ProxyReRanker struct {
	endpoint   string
	httpClient *http.Client
}

func NewProxyReranker(endpoint string, client *http.Client) ReRanker {
	return &ProxyReRanker{
		endpoint:   endpoint,
		httpClient: client,
	}
}

var _ ReRanker = (*ProxyReRanker)(nil)

func (p *ProxyReRanker) RankDocuments(ctx context.Context, query string, texts []string) ([]*ReRankResponse, error) {
	payload := &reRankRequest{
		Query: query,
		Texts: texts,
	}
	b := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(b).Encode(payload); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.endpoint+"/rerank", b)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		logrus.Errorf("error reranking documents: %s", err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		logrus.Errorf("error reranking documents: %s", resp.Status)
		return nil, errors.New(resp.Status)
	}

	defer resp.Body.Close()
	res := []*ReRankResponse{}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, err
	}
	return res, nil
}

func (p *ProxyReRanker) Health(ctx context.Context) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/health", nil)
	if err != nil {
		return false, err
	}
	resp, err := p.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		return false, errors.New(resp.Status)
	}
	defer resp.Body.Close()
	return true, nil
}

func (p *ProxyReRanker) Info(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.endpoint+"/info", nil)
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

type GrpcProxyReRanker struct {
	tei.InfoClient
	tei.RerankClient
}

func NewGrpcProxyReRanker(infoClient tei.InfoClient, rerankClient tei.RerankClient) ReRanker {
	return &GrpcProxyReRanker{
		InfoClient:   infoClient,
		RerankClient: rerankClient,
	}
}

var _ ReRanker = (*GrpcProxyReRanker)(nil)

func (g *GrpcProxyReRanker) Health(ctx context.Context) (bool, error) {
	_, err := g.InfoClient.Info(ctx, &tei.InfoRequest{})
	if err != nil {
		logrus.Errorf("error fetching reranker health: %s", err)
		return false, err
	}
	return true, nil
}

func (g *GrpcProxyReRanker) Info(ctx context.Context) ([]byte, error) {
	resp, err := g.InfoClient.Info(ctx, &tei.InfoRequest{})
	if err != nil {
		logrus.Errorf("error fetching reranker info: %s", err)
		return nil, err
	}
	data, err := protojson.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return data, err
}

func (g *GrpcProxyReRanker) RankDocuments(ctx context.Context, query string, texts []string) ([]*ReRankResponse, error) {
	resp, err := g.RerankClient.Rerank(ctx, &tei.RerankRequest{
		Query: query,
		Texts: texts,
	})
	if err != nil {
		return nil, err
	}
	return lo.Map(resp.Ranks, func(r *tei.Rank, _ int) *ReRankResponse {
		return &ReRankResponse{
			Score: float64(r.Score),
			Index: int(r.Index),
		}
	}), nil
}
