package embed

import (
	"fmt"
	"strings"

	"github.com/Synaptic-Lynx/rag-gateway/api/tei"
	"github.com/samber/lo"
)

type EmbedderConfig struct {
	Embedders []EmbedderSpec `yaml:"embedders"`
}

type EmbedderSpec struct {
	Endpoint  string `yaml:"endpoint"`
	ModelId   string `yaml:"model_id"`
	ModelType string `yaml:"model_type"`
	Sparse    bool   `yaml:"sparse"`
}

func (e *EmbedderConfig) Validate() error {
	if len(e.Embedders) == 0 {
		return fmt.Errorf("no embedders defined")
	}
	seenIds := map[string]struct{}{}
	for _, embedder := range e.Embedders {
		if _, ok := seenIds[embedder.ModelId]; ok {
			return fmt.Errorf("duplicate model id: %s", embedder.ModelId)
		}
		if err := embedder.Validate(); err != nil {
			return err
		}
		seenIds[embedder.ModelId] = struct{}{}
	}
	return nil
}

func (e *EmbedderSpec) Validate() error {
	if e.Endpoint == "" {
		return fmt.Errorf("missing endpoint")
	}
	if e.ModelId == "" {
		return fmt.Errorf("missing model id")
	}
	if e.ModelType == "" {
		return fmt.Errorf("missing model type")
	}

	if _, ok := tei.ModelType_value[e.ModelType]; !ok {
		return fmt.Errorf("invalid model type: %s, available : %s", e.ModelType, strings.Join(lo.Keys(tei.ModelType_value), ","))
	}
	return nil
}
