package collection

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SectionMetadataStore interface {
	CreateSectionMetadata(ctx context.Context, sectionID string, md SectionMetadata) error
	UpdateSectionMetadata(ctx context.Context, sectionID string, md SectionMetadata) (changed bool, err error)
	GetSectionMetadata(ctx context.Context, sectionID string) (SectionMetadata, error)
	DeleteSectionMetadata(ctx context.Context, sectionID string) error
	BatchSectionMetadata(ctx context.Context) ([]SectionMetadata, error)
	ListSectionMetadata(ctx context.Context) (ids []string, err error)
}

var _ SectionMetadataStore = &sectionManager{}

type SectionMetadata struct {
	SectionID       string          `json:"section_id"`
	UserDescription string          `json:"user_description"`
	ChunkSpec       ChunkSpec       `json:"chunk_spec"`
	Spec            []EmbeddingSpec `json:"spec"`
}

type ChunkSpec struct {
	ChunkSize    int  `json:"chunk_size"`
	ChunkOverlap int  `json:"chunk_overlap"`
	CodeBlocks   bool `json:"code_blocks"`
}

func (cs ChunkSpec) Validate() error {
	if cs.ChunkSize <= 0 {
		return fmt.Errorf("invalid chunk size")
	}
	if cs.ChunkOverlap <= 0 {
		return fmt.Errorf("invalid chunk overlap")
	}
	return nil
}

func (s SectionMetadata) Validate() error {
	if s.SectionID == "" {
		return fmt.Errorf("invalid section id")
	}
	if len(s.Spec) == 0 {
		return fmt.Errorf("no embedding specs set")
	}
	if err := s.ChunkSpec.Validate(); err != nil {
		return fmt.Errorf("invalid chunk spec: %w", err)
	}
	for _, es := range s.Spec {
		if err := es.Validate(); err != nil {
			return fmt.Errorf("invalid spec: %w", err)
		}
	}
	return nil
}

type EmbeddingSpec struct {
	// Embedding specific
	EmbeddingID string `json:"embedding_id"`

	// Vector store specific
	IndexId      string `json:"index_id"`
	MetricType   string `json:"metric_type"`
	CollectionId string `json:"collection_id"`
	PartitionId  string `json:"partition_id"`
}

func (es EmbeddingSpec) Validate() error {
	if es.EmbeddingID == "" {
		return fmt.Errorf("embedding id required")
	}
	if es.IndexId == "" {
		return fmt.Errorf("vector store index id required")
	}
	if es.MetricType == "" {
		return fmt.Errorf("vector store metric type required")
	}
	if es.CollectionId == "" {
		return fmt.Errorf("collection id required")
	}
	return nil
}

func (s *sectionManager) CreateSectionMetadata(ctx context.Context, sectionID string, md SectionMetadata) error {
	if err := md.Validate(); err != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	_, err := s.docMetadataKv.Get(ctx, sectionID)
	if err != nil && util.StatusCode(err) != codes.NotFound {
		return status.Error(codes.FailedPrecondition, fmt.Sprintf("fetching section error : %s", err))
	} else if err == nil {
		return status.Error(codes.AlreadyExists, fmt.Sprintf("section already exists : %s", sectionID))
	}
	data, err := json.Marshal(md)

	if err != nil {
		return err
	}
	return s.docMetadataKv.Put(ctx, sectionID, data, storage.WithRevision(0))
}

func (s *sectionManager) UpdateSectionMetadata(
	ctx context.Context,
	sectionID string,
	md SectionMetadata,
) (changed bool, err error) {
	// TODO
	if err := md.Validate(); err != nil {
		return false, status.Error(codes.InvalidArgument, err.Error())
	}
	return false, nil
}

func (s *sectionManager) GetSectionMetadata(ctx context.Context, sectionID string) (SectionMetadata, error) {
	var md SectionMetadata
	data, err := s.docMetadataKv.Get(ctx, sectionID)
	if err != nil {
		return SectionMetadata{}, err
	}
	err = json.Unmarshal(data, &md)
	return md, err
}

func (s *sectionManager) BatchSectionMetadata(ctx context.Context) ([]SectionMetadata, error) {
	keys, err := s.docMetadataKv.ListKeys(ctx, "")
	if err != nil {
		return nil, err
	}
	slices.Sort(keys)
	mds := []SectionMetadata{}

	for _, key := range keys {
		md, err := s.GetSectionMetadata(ctx, key)
		if err != nil {
			return nil, err
		}
		mds = append(mds, md)
	}
	return mds, nil
}

func (s *sectionManager) ListSectionMetadata(ctx context.Context) ([]string, error) {
	keys, err := s.docMetadataKv.ListKeys(ctx, "")
	if err != nil {
		return nil, err
	}
	slices.Sort(keys)
	return keys, nil
}

func (s *sectionManager) DeleteSectionMetadata(ctx context.Context, sectionID string) error {
	return s.docMetadataKv.Delete(ctx, sectionID)
}
