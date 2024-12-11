package gateway

import (
	"context"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/api"
	"github.com/sirupsen/logrus"
)

type GatewayServer struct {
	ctx           context.Context
	router        *gin.Engine
	config        GatewayConfig
	runtimeConfig GatewayRuntimeConfig
}

func NewGatewayServer(ctx context.Context, cfg GatewayConfig) *GatewayServer {
	rcfg, err := setup(ctx, cfg)
	if err != nil {
		logrus.Fatalf("failed to setup gateway server : %s", err)
	}

	r := gin.Default()
	// https://github.com/gin-gonic/gin/issues/2047
	r.UseRawPath = true
	r.UnescapePathValues = false
	gw := &GatewayServer{
		ctx:           ctx,
		router:        r,
		config:        cfg,
		runtimeConfig: rcfg,
	}
	gw.configureManagementRoutes()
	return gw
}

func (g *GatewayServer) configureManagementRoutes() {
	g.router.GET(g.apiRoute("/info"), g.info)
	g.router.GET(g.apiRoute("/health"), g.health)
}

func (g *GatewayServer) apiRoute(route string) string {
	return path.Join("/management/api/v1alpha1", route)
}

func (g *GatewayServer) info(c *gin.Context) {
	info := GenericInfo{
		Info: map[string]InfoResponse{},
	}

	llmInfo, err := g.llmInfo(c.Request.Context())
	if err != nil {
		info.Info["llm"] = InfoResponse{
			Health: "unhealthy",
			Info:   err.Error(),
		}
	} else {
		info.Info["llm"] = InfoResponse{
			Health: "healthy",
			Info:   llmInfo,
		}
	}

	embeddingHealth, _ := g.runtimeConfig.EmbeddingClient.Health(c.Request.Context())
	if !embeddingHealth {
		info.Info["embedding"] = InfoResponse{
			Health: "unhealthy",
			Info:   "embedding client unhealthy",
		}
	} else {
		embeddingInfo, err := g.runtimeConfig.EmbeddingClient.Info(c.Request.Context())
		if err != nil {
			info.Info["embedding"] = InfoResponse{
				Health: "healthy",
				Info:   err.Error(),
			}
		} else {
			info.Info["embedding"] = InfoResponse{
				Health: "healthy",
				Info:   string(embeddingInfo),
			}
		}
	}

	reRankerHealth, _ := g.runtimeConfig.ReRankerClient.Health(c.Request.Context())
	if !reRankerHealth {
		info.Info["reranker"] = InfoResponse{
			Health: "unhealthy",
			Info:   "reranker client unhealthy",
		}
	} else {
		reRankerInfo, err := g.runtimeConfig.ReRankerClient.Info(c.Request.Context())
		if err != nil {
			info.Info["reranker"] = InfoResponse{
				Health: "healthy",
				Info:   err.Error(),
			}
		} else {
			info.Info["reranker"] = InfoResponse{
				Health: "healthy",
				Info:   string(reRankerInfo),
			}
		}
	}

	c.Header("Content-Type", "application/json")
	c.JSON(200, info)
}

type GenericInfo struct {
	Info map[string]InfoResponse `json:",inline"`
}

type InfoResponse struct {
	Health string `json:"health"`
	Info   any    `json:"info"`
}

func (g *GatewayServer) llmInfo(ctx context.Context) ([]api.ProcessModelResponse, error) {
	resp, err := g.runtimeConfig.OllamaClient.ListRunning(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Models, nil
}

func (g *GatewayServer) health(c *gin.Context) {
	embeddingHealth, _ := g.runtimeConfig.EmbeddingClient.Health(c.Request.Context())

	reRankerHealth, _ := g.runtimeConfig.ReRankerClient.Health(c.Request.Context())

	llmErr := g.runtimeConfig.OllamaClient.Heartbeat(c.Request.Context())
	llmHealth := llmErr == nil

	code := http.StatusOK

	if !embeddingHealth {
		code = http.StatusServiceUnavailable
	}

	if !reRankerHealth {
		code = http.StatusServiceUnavailable
	}

	if !llmHealth {
		code = http.StatusServiceUnavailable
	}

	c.JSON(code, gin.H{
		"embedding": embeddingHealth,
		"reranker":  reRankerHealth,
		"llm":       llmHealth,
	})
}

func (g *GatewayServer) Router() *gin.Engine {
	return g.router
}

func (g *GatewayServer) RuntimeConfig() GatewayRuntimeConfig {
	return g.runtimeConfig
}

func (g *GatewayServer) Config() GatewayConfig {
	return g.config
}

// Blocks until the server exits
func (g *GatewayServer) ListenAndServe(ctx context.Context, addr string) error {
	logrus.Infof("setting up RAG gateway server...")

	logrus.Infof("RAG gateway running on : %s", addr)
	return g.router.Run(addr)
}
