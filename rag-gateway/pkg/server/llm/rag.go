package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/embeddings"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/fileutil"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/prompts"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/qa"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/retriever"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/gateway"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/similarity"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/history"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/ollama/ollama/api"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/vectorstores"
	"github.com/tmc/langchaingo/vectorstores/milvus"
	"golang.org/x/sync/errgroup"
)

const (
	defaultScoreThreshold = 0.7
	defaultK              = 60
	defaultTopN           = 5
)

type RAGServerConfig struct {
	GenerateTimeout time.Duration
	GatewayAddress  string
}

type RAGServerRuntimeConfig struct {
	LLMModel     *ollama.LLM
	OllamaClient *api.Client

	VectorStore vectorstores.VectorStore
	Index       *entity.IndexIvfFlat
	Reranker    embeddings.ReRanker
	ChatKV      storage.KeyValueStore
}

type ragServer struct {
	util.Initializer
	runCfg RAGServerRuntimeConfig
	cfg    RAGServerConfig
}

func NewRAGServer(config RAGServerConfig) gateway.HTTPApiExtension {
	return &ragServer{
		cfg: config,
	}
}

var _ gateway.HTTPApiExtension = (*ragServer)(nil)

func (r *ragServer) ConfigureRoutes(router *gin.Engine, cfg gateway.GatewayRuntimeConfig) {
	r.Initialize(cfg)
	router.POST(r.apiRoute("answer"), r.answer)
	router.POST(r.apiRoute("stream"), r.deprecatedStream)
}

func (r *ragServer) Initialize(cfg gateway.GatewayRuntimeConfig) {
	r.InitOnce(func() {
		// FIXME: hardcoded
		index, err := entity.NewIndexIvfFlat(entity.L2, 2)
		if err != nil {
			logrus.Fatalf("error setting up vectorstore index : %s", err)
		}
		vStore, err := milvus.New(
			cfg.GwContext,
			cfg.VectorConfig,
			// FIXME: hardcoded
			milvus.WithCollectionName("westernblot"),
			milvus.WithEmbedder(cfg.EmbeddingClient),
			milvus.WithIndex(index),
		)
		if err != nil {
			logrus.Fatalf("error setting up vector store client : %s", err)
		}

		r.runCfg = RAGServerRuntimeConfig{
			LLMModel:     cfg.LLMModel,
			OllamaClient: cfg.OllamaClient,
			VectorStore:  vStore,
			Index:        index,
			Reranker:     cfg.ReRankerClient,
			ChatKV:       cfg.ChatKV,
		}
	})
}

func (r *ragServer) Timeout() time.Duration {
	return r.cfg.GenerateTimeout
}

type RAGRequest struct {
	ChatId          string  `json:"chatId,omitempty"`
	Question        string  `json:"question"`
	ScoreThreshhold float64 `json:"score_threshhold"`
	K               int     `json:"k,omitempty"`
	TopN            int     `json:"topN,omitempty"`
}

func (r *RAGRequest) Validate() error {
	if r.Question == "" {
		return fmt.Errorf("question must be set")
	}

	if r.ScoreThreshhold < 0 {
		return fmt.Errorf("score_threshhold must be positive")
	}

	if r.ScoreThreshhold >= 0.8 {
		return fmt.Errorf("score_threshhold must be less than 0.8")
	}

	if r.K < 0 {
		return fmt.Errorf("k must be positive")
	}

	if r.K > defaultK {
		return fmt.Errorf("k must be less than or equal to %d", defaultK)
	}

	if r.TopN < 0 {
		return fmt.Errorf("topN must be positive")
	}

	if r.TopN > r.K {
		return fmt.Errorf("topN must be less than or equal to k")
	}

	return nil
}

func (r *RAGRequest) Sanitize() {
	if r.ScoreThreshhold == 0 {
		r.ScoreThreshhold = 0.7
	}
	if r.K == 0 {
		r.K = defaultK
	}
	if r.TopN == 0 {
		r.TopN = defaultTopN
	}
}

func (r *ragServer) newChain(k int, chatId string) chains.Chain {
	// TODO : not all request hyper parameters are handled here, only K is

	baseRetriever := retriever.NewSimilarityRetriever(r.runCfg.VectorStore, retriever.WithK(k))
	rerankRetriever := retriever.NewReRankerRetriever(baseRetriever, r.runCfg.Reranker)

	mem := memory.NewConversationBuffer(
		memory.WithChatHistory(
			history.NewKVChatHistory(r.runCfg.ChatKV, chatId),
		),
		memory.WithReturnMessages(true),
	)
	chain := qa.NewQA(
		r.runCfg.LLMModel,
		rerankRetriever,
		mem,
	)
	return chain
}

func (r *ragServer) answer(c *gin.Context) {
	if !r.Initialized() {
		c.JSON(500, gin.H{"error": "rag server not initialized"})
		return
	}
	req := &RAGRequest{}
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("failed to bind request: %s", err)})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid request: %s", err)})
		return
	}

	req.Sanitize()
	var chatId string
	if req.ChatId == "" {
		// TODO : should encapsulate assumptions about keys being ordered by timestamp
		chatId = fmt.Sprintf("%d", time.Now().UnixNano())
	} else {
		chatId = req.ChatId
	}
	c.Header("Content-Type", "application/json")

	ctxT, caT := context.WithTimeout(c.Request.Context(), r.Timeout())
	defer caT()

	qaChain := r.newChain(req.K, chatId)

	response, err := chains.Run(
		ctxT,
		qaChain,
		req.Question,
	)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to generate response: %s", err)})
		return
	}

	c.JSON(200, gin.H{"response": response, "chatId": chatId})
}

func (r *ragServer) deprecatedStream(c *gin.Context) {
	if !r.Initialized() {
		c.JSON(500, gin.H{"error": "rag server not initialized"})
		return
	}

	req := &RAGRequest{}
	if err := c.BindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("failed to bind request: %s", err)})
		return
	}

	if err := req.Validate(); err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("invalid request: %s", err)})
		return
	}

	req.Sanitize()
	c.Header("Content-Type", "application/json")

	qaTmpl := prompts.NewQATemplate()
	ctxT, caT := context.WithTimeout(c.Request.Context(), r.Timeout())
	defer caT()
	chunks, err := r.similarity(ctxT, &similarity.SimilarityRequest{
		Query: req.Question,
		K:     req.K,
		TopN:  req.TopN,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get similarity: %s", err)})
		return
	}

	goodChunks := lo.Filter(chunks, func(c fileutil.DocumentChunk, _ int) bool {
		return c.Score >= req.ScoreThreshhold
	})

	if len(goodChunks) == 0 {
		c.JSON(200, gin.H{"response": "Unfortunately no matching were documents found"})
		return
	}

	logrus.Infof("found %d good chunks", len(goodChunks))

	sources := lo.Map(goodChunks, func(c fileutil.DocumentChunk, _ int) string {
		return c.Metadata.DocID
	})
	logrus.Infof("relevant documents : %v", strings.Join(sources, ", "))

	docs := lo.Map(goodChunks, func(c fileutil.DocumentChunk, _ int) string {
		return c.Contents
	})
	llm := r.runCfg.LLMModel

	ctxT2, caT2 := context.WithTimeout(c.Request.Context(), r.Timeout())
	defer caT2()

	// TODO : requires openAPI chatbot langchaingo implementation
	if _, err := qaTmpl.Format(map[string]any{
		"question": req.Question,
		"context":  strings.Join(docs, "\n----\n"),
	}); err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("failed to format question: %s", err)})
		return
	}
	//FIXME: hack
	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, fmt.Sprintf(
			"Answer the question based only on the following context : %s",
			strings.Join(docs, "\n----\n"),
		)),
		llms.TextParts(llms.ChatMessageTypeHuman, fmt.Sprintf(
			"Answer the question based on the above context : %s",
			req.Question,
		)),
	}

	chanResp := make(chan []byte, 16)
	eg, ctx := errgroup.WithContext(ctxT2)
	eg.Go(func() error {
		defer close(chanResp)
		_, err = llm.GenerateContent(ctx, content, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			chanResp <- chunk
			fmt.Print(string(chunk))
			return nil
		}))
		return nil
	})

	enc := json.NewEncoder(c.Writer)
	c.Writer.WriteHeader(http.StatusOK)
	eg.Go(func() error {
		for {
			select {
			case chunk, ok := <-chanResp:
				if !ok {
					return nil
				}
				if err := enc.Encode(string(chunk)); err != nil {
					logrus.Errorf("failed to encode chunk: %v", err)
					return err
				}
				c.Writer.Flush()
			case <-ctx.Done():
				logrus.Errorf("context exceeded: %v", ctx.Err())
				c.Writer.Flush()
				return ctx.Err()
			}
		}
	})

	if err := eg.Wait(); err != nil {
		logrus.Errorf("failed to generate content: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf(
			"failed to generate content : %s",
			err.Error(),
		)})
		c.Set("trailer", fmt.Sprintf("failed to generate content : %s", err.Error()))
		return
	}

	c.Set("trailer", "success")
}

type SimilarityResponse struct {
	Documents []fileutil.DocumentChunk `json:"res"`
}

// FIXME: hack
func (r *ragServer) similarity(ctx context.Context, req *similarity.SimilarityRequest) ([]fileutil.DocumentChunk, error) {
	b := bytes.NewBuffer([]byte{})
	if err := json.NewEncoder(b).Encode(req); err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		// FIXME: hardcoded
		fmt.Sprintf("http://%s/similarity/api/v1alpha1/search", r.cfg.GatewayAddress),
		b,
	)
	if err != nil {
		return nil, err
	}

	client := http.DefaultClient
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("similarity server error: %s", resp.Status)
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	res := &SimilarityResponse{}

	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&res); err != nil {
		return nil, err
	}

	return res.Documents, nil
}

func (r *ragServer) apiRoute(route string) string {
	return path.Join("/rag/api/v1alpha1", route)
}
