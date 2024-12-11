package similarity

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/fileutil"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/retriever"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/gateway"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/vectorstores/milvus"
)

const (
	defaultSimilarityK = 60
	defaultReRankTopN  = 5
)

type SimilarityServerConfig struct{}

type SimilarityServerRuntimeConfig struct {
	Ctx       context.Context
	Retriever retriever.Retriever
}

type SimilarityServer struct {
	util.Initializer
	runtimeCfg SimilarityServerRuntimeConfig
}

func NewSimilarityServer(ctx context.Context) gateway.HTTPApiExtension {
	return &SimilarityServer{}
}

func (s *SimilarityServer) Initialize(cfg gateway.GatewayRuntimeConfig) {
	s.InitOnce(func() {
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

		baseRetriever := retriever.NewSimilarityRetriever(vStore, retriever.WithK(60))
		rerankRetriever := retriever.NewReRankerRetriever(baseRetriever, cfg.ReRankerClient)

		s.runtimeCfg = SimilarityServerRuntimeConfig{
			Ctx:       cfg.GwContext,
			Retriever: rerankRetriever,
		}
	})
}

func (s *SimilarityServer) ConfigureRoutes(router *gin.Engine, cfg gateway.GatewayRuntimeConfig) {
	s.Initialize(cfg)
	router.POST(s.apiRoute("/search"), s.similarity)
}

func (s *SimilarityServer) apiRoute(route string) string {
	return path.Join("/similarity/api/v1alpha1", route)
}

type SimilarityRequest struct {
	Query string `json:"query"`
	K     int    `json:"k,omitempty"`
	TopN  int    `json:"topN,omitempty"`
}

func (s *SimilarityRequest) Validate() error {
	if s.Query == "" {
		return fmt.Errorf("missing field query")
	}
	if s.K < 0 {
		return fmt.Errorf("k must be positive")
	}
	if s.TopN < 0 {
		return fmt.Errorf("topN must be positive")
	}

	if s.K > defaultSimilarityK {
		return fmt.Errorf("k must be less than or equal to %d", defaultSimilarityK)
	}

	if s.K != 0 && s.TopN != 0 && s.TopN > s.K {
		return fmt.Errorf("topN must be less than or equal to k")
	}
	return nil
}

func (s *SimilarityRequest) Sanitize() {
	if s.K == 0 {
		s.K = defaultSimilarityK
	}
	if s.TopN == 0 {
		s.TopN = defaultReRankTopN
	}
}

func (s *SimilarityServer) handleSimilarityRequest(c *gin.Context) (*SimilarityRequest, error) {
	req := &SimilarityRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	}
	if err := req.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return nil, err
	}
	req.Sanitize()
	return req, nil
}

func (s *SimilarityServer) similarity(c *gin.Context) {
	if !s.Initialized() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "server not initialized"})
		return
	}

	req, err := s.handleSimilarityRequest(c)
	if err != nil {
		logrus.Errorf("failed to decode similarity request: %s", err)
		return
	}
	ctxT, caT := context.WithTimeout(c.Request.Context(), 2*time.Minute)
	defer caT()

	docs, err := s.runtimeCfg.Retriever.AsRetriever().GetRelevantDocuments(ctxT, req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		logrus.Errorf("failed to get relevant documents: %s", err)
		return
	}
	logrus.Infof("found %d relevant documents", len(docs))

	reranked := lo.Map(docs, func(d schema.Document, _ int) fileutil.DocumentChunk {
		return fileutil.FromSchemaDoc(d)
	})

	c.JSON(http.StatusOK, gin.H{
		"res": reranked[:req.TopN],
	})
}
