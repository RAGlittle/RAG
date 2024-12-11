package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/chat"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/gateway"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/llm"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/similarity"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

const (
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorReset  = "\033[0m"
)

func methodToColor(method string) string {
	switch method {
	case http.MethodDelete:
		return fmt.Sprintf("%s%s%s", colorRed, method, colorReset)
	case http.MethodGet:
		return fmt.Sprintf("%s%s%s", colorGreen, method, colorReset)
	case http.MethodPost:
		return fmt.Sprintf("%s%s%s", colorBlue, method, colorReset)
	default:
		return method
	}

}

func registerApis(gw *gateway.GatewayServer) {
	logrus.Info("Setting up LLM server...")
	llmServer := llm.NewLLMServer(llm.LLmServerConfig{
		ModelName:       gw.Config().ModelName,
		Version:         version.FriendlyVersion(),
		GenerateTimeout: 2 * time.Minute,
	})
	llmServer.ConfigureRoutes(gw.Router(), gw.RuntimeConfig())
	logrus.Info("LLM server configured")

	logrus.Info("Setting up similarity server...")
	simServer := similarity.NewSimilarityServer(context.Background())
	simServer.ConfigureRoutes(gw.Router(), gw.RuntimeConfig())
	logrus.Info("Similarity server configured")

	logrus.Info("Setting up RAG server...")
	ragServer := llm.NewRAGServer(llm.RAGServerConfig{
		GenerateTimeout: time.Minute * 5,
		GatewayAddress:  gw.Config().GatewayAddress,
	})
	ragServer.ConfigureRoutes(gw.Router(), gw.RuntimeConfig())
	logrus.Info("RAG server configured")

	chatServer := chat.NewChatServer()
	chatServer.ConfigureRoutes(gw.Router(), gw.RuntimeConfig())
}

func startGateway(gw *gateway.GatewayServer) error {
	for _, route := range gw.Router().Routes() {

		logrus.Infof("%s %s", methodToColor(route.Method), route.Path)
	}

	if err := gw.ListenAndServe(context.Background(), "0.0.0.0:5555"); err != nil {
		return err
	}
	return nil
}

func BuildGatewayCmd() *cobra.Command {
	var listenAddr string
	var configFile string
	cmd := &cobra.Command{
		Use:     "ragger",
		Version: version.FriendlyVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			gin.SetMode(gin.ReleaseMode)
			version := version.FriendlyVersion()
			logrus.Infof("Rag Gateway %s starting...", version)
			logrus.Infof("loading config file : %s ...", configFile)
			if _, err := os.Stat(configFile); err != nil {
				logrus.Errorf("config file not found: %s", configFile)
				return fmt.Errorf("config file not found: %s", configFile)
			}

			data, err := os.ReadFile(configFile)
			if err != nil {
				logrus.Errorf("error reading config file: %s", err)
			}
			cfg := gateway.GatewayConfig{}
			if err := yaml.Unmarshal(data, &cfg); err != nil {
				logrus.Errorf("error parsing config file: %s", err)
				return fmt.Errorf("error parsing config file: %s", err)
			}
			cfg.GatewayAddress = listenAddr
			cfg.Version = cmd.Version
			logrus.Info("config loaded")

			gw := gateway.NewGatewayServer(cmd.Context(), cfg)
			registerApis(gw)
			return startGateway(gw)
		},
	}
	cmd.Flags().StringVarP(&listenAddr, "listen-addr", "l", "0.0.0.0:5555", "gateway listen address")
	cmd.Flags().StringVarP(&configFile, "config", "f", "/var/opt/ragger/config.yaml", "config file for gateway")
	return cmd
}

func main() {
	cmd := BuildGatewayCmd()
	if err := cmd.Execute(); err != nil {
		logrus.Errorf("error starting gateway: %s", err)
	}
}
