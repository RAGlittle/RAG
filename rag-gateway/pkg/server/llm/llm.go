package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"time"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/gateway"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/ollama/ollama/api"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"golang.org/x/sync/errgroup"
)

type LLmServerConfig struct {
	ModelName       string
	Version         string
	GenerateTimeout time.Duration
}

type LLmServerRuntimeConfig struct {
	LLMModel     *ollama.LLM
	OllamaClient *api.Client
}

type llmServer struct {
	util.Initializer
	cfg        LLmServerConfig
	runtimeCfg LLmServerRuntimeConfig
}

type LLMServerInfo struct {
	Models []api.ProcessModelResponse `json:"inline"`
}

func NewLLMServer(config LLmServerConfig) gateway.HTTPApiExtension {
	return &llmServer{
		cfg: config,
	}
}

func (l *llmServer) apiRoute(route string) string {
	return path.Join("/llm/api/v1alpha1", route)
}

func (l *llmServer) Timeout() time.Duration {
	return l.cfg.GenerateTimeout
}

func (l *llmServer) Initialize(config gateway.GatewayRuntimeConfig) {
	l.InitOnce(func() {
		l.runtimeCfg = LLmServerRuntimeConfig{
			LLMModel:     config.LLMModel,
			OllamaClient: config.OllamaClient,
		}
	})
}

func (l *llmServer) ConfigureRoutes(router *gin.Engine, config gateway.GatewayRuntimeConfig) {
	l.Initialize(config)
	router.GET(
		l.apiRoute("info"),
		l.getLLM,
	)
	router.POST(
		l.apiRoute("prompt/full"),
		l.postFullPrompt,
	)

	router.POST(
		l.apiRoute("prompt/full/stream"),
		l.streamFullPrompt,
	)
}

func (l *llmServer) getLLM(c *gin.Context) {
	if !l.Initialized() {
		c.JSON(500, gin.H{"error": "llm server not initialized"})
		return
	}
	resp, err := l.runtimeCfg.OllamaClient.ListRunning(c.Request.Context())
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf(
			"failed to list running models : %s",
			err.Error(),
		)})
		return
	}
	c.Header("Content-Type", "application/json")
	c.JSON(200, gin.H{
		"models": LLMServerInfo{
			Models: resp.Models,
		},
		"version": l.cfg.Version,
	})
}

type FullPrompt struct {
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
}

func (l *llmServer) streamFullPrompt(c *gin.Context) {
	if !l.Initialized() {
		c.JSON(500, gin.H{"error": "llm server not initialized"})
		return
	}
	ctx, ca := context.WithTimeout(c.Request.Context(), l.Timeout())
	defer ca()

	contents := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "You are a company branding design wizard."),
		llms.TextParts(llms.ChatMessageTypeHuman, "What is today's date?"),
	}
	chanResp := make(chan []byte, 16)
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer close(chanResp)
		completion, err := l.runtimeCfg.LLMModel.GenerateContent(c.Request.Context(), contents, llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
			chanResp <- chunk
			return nil
		}))
		if err != nil {
			return err
		}
		_ = completion
		return nil
	})

	c.Header("Content-Type", "application/json")
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

func (l *llmServer) postFullPrompt(c *gin.Context) {
	if !l.Initialized() {
		c.JSON(500, gin.H{"error": "llm server not initialized"})
		return
	}
	ctx, ca := context.WithTimeout(c.Request.Context(), l.Timeout())
	defer ca()
	var fullPrompt FullPrompt
	if err := c.BindJSON(&fullPrompt); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	content := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, fullPrompt.SystemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, fullPrompt.UserPrompt),
	}

	completion, err := l.runtimeCfg.LLMModel.GenerateContent(ctx, content)
	if err != nil {
		logrus.Errorf("failed to generate content: %v", err)
		c.JSON(500, gin.H{"error": fmt.Sprintf(
			"failed to generate content : %s",
			err.Error(),
		)})
		return
	}
	c.Header("Content-Type", "application/json")
	c.JSON(200, gin.H{
		"completion": completion,
	})
}
