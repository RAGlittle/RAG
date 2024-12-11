package embed

import (
	"context"
	"fmt"
	"strings"

	"github.com/Synaptic-Lynx/rag-gateway/api/embedgw"
	"github.com/Synaptic-Lynx/rag-gateway/api/tei"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/embeddings"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type clientWrapper struct {
	embedder tei.EmbedClient
	info     tei.InfoClient
	sparse   bool
}

type EmbedGatewayServer struct {
	cfg     EmbedderConfig
	clients map[string]clientWrapper
}

var _ embedgw.EmbedGatewayServer = (*EmbedGatewayServer)(nil)

// TODO : wrap setup logic in an intiializer pattern
func NewEmbedGatewayServer(cfg EmbedderConfig) (*EmbedGatewayServer, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	clients := make(map[string]clientWrapper)
	for _, embedder := range cfg.Embedders {
		cc, err := grpc.NewClient(
			embedder.Endpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if err != nil {
			return nil, err
		}

		infoClient := tei.NewInfoClient(cc)
		embedClient := tei.NewEmbedClient(cc)

		// TODO : add exponential backoff logic to this
		info, err := infoClient.Info(context.Background(), &tei.InfoRequest{})
		if err != nil {
			return nil, err
		}

		if info.ModelId != embedder.ModelId {
			return nil, fmt.Errorf("model id mismatch: %s != %s", info.ModelId, embedder.ModelId)
		}

		if info.ModelType.String() != embedder.ModelType {
			return nil, fmt.Errorf("model type mismatch: %s != %s", info.ModelType, embedder.ModelType)
		}
		clients[info.ModelId] = clientWrapper{
			embedder: embedClient,
			info:     infoClient,
			sparse:   embedder.Sparse,
		}
	}

	if len(clients) == 0 {
		return nil, fmt.Errorf("no embedders configured")
	}

	registered := lo.Keys(clients)

	logrus.Infof("Registered embedders: %s", strings.Join(registered, ", "))

	s := &EmbedGatewayServer{
		cfg:     cfg,
		clients: clients,
	}

	logrus.Infof("Embed gateway server initialized with %d embedders", len(registered))

	return s, nil
}

func (s *EmbedGatewayServer) Embed(ctx context.Context, req *embedgw.EmbedSpecificRequest) (*tei.EmbedResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	content := req.Inputs

	embedder, err := s.AsEmbedder(req.EmbeddingID)
	if err != nil {
		return nil, err
	}

	vec, err := embedder.EmbedQuery(ctx, content)
	if err != nil {
		return nil, err
	}

	return &tei.EmbedResponse{
		Embeddings: vec,
	}, nil
}

func (s *EmbedGatewayServer) Info(ctx context.Context, req *tei.InfoRequest) (*embedgw.InfoMapResponse, error) {
	res := &embedgw.InfoMapResponse{}
	for embedderId, client := range s.clients {
		infoResp, err := client.info.Info(ctx, req)
		if err != nil {
			return nil, err
		}
		res.Info[embedderId] = infoResp
	}
	return res, nil
}

func (s *EmbedGatewayServer) AsEmbedder(embeddingID string) (embeddings.Embedder[float32], error) {
	logrus.Info(lo.Keys(s.clients))
	client, ok := s.clients[embeddingID]
	if !ok {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("embedder not configured as part of RAG gateway: %s", embeddingID))
	}
	baseEmbedder := embeddings.NewGrpcProxyEmbedder(client.info, client.embedder, client.sparse)
	return embeddings.NewPoolEmbedder("", embeddings.WithDebug(), embeddings.WithEmbedder(baseEmbedder)), nil
}
