package vectorstore

// !! Deprecated in favor of langchaingo.VectorStore implementation

import (
	"context"
	"fmt"
	"log"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/fileutil"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/sirupsen/logrus"
)

type VectorStore interface {
	Health(ctx context.Context) (bool, error)
	Client() client.Client
}

func NewVectorStore() (VectorStore, error) {
	client, err := setupClient()
	if err != nil {
		return nil, err
	}
	logrus.Info("setting up vector store defaults")
	setupDefaults(client)
	logrus.Info("vector store setup complete")
	return &vectorStore{
		client: client,
	}, nil
}

type vectorStore struct {
	client client.Client
}

func (v *vectorStore) Health(ctx context.Context) (bool, error) {
	state, err := v.client.CheckHealth(ctx)
	if err != nil {
		return false, err
	}
	return state.IsHealthy, nil
}

func (v *vectorStore) Client() client.Client {
	return v.client
}

func setupClient() (client.Client, error) {
	client, err := client.NewClient(context.Background(), client.Config{
		Address:  "localhost:19530",
		Password: "minioadmin",
		APIKey:   "minioadmin",
	})
	if err != nil {
		// handle error
		return nil, err
	}
	return client, nil
}

// ================= HACK ==================

var westernBlotSchema = &entity.Schema{
	CollectionName: "westernblot",
	Description:    "Embeddings for westernblot documents",
	Fields: []*entity.Field{
		entity.NewField().WithName("ID").WithDataType(entity.FieldTypeVarChar).WithMaxLength(256).
			WithIsPrimaryKey(true),
		entity.NewField().WithName("type").WithDataType(entity.FieldTypeVarChar).WithMaxLength(64),
		entity.NewField().WithName("embeddingID").WithDataType(entity.FieldTypeVarChar).WithMaxLength(64),
		entity.NewField().WithName("embedding").WithDataType(entity.FieldTypeFloatVector).WithDim(1024),
	},
}

func InsertDocumentChunk(client client.Client, doc []fileutil.DocumentChunk, embeddings [][][]float32) {
	ids := []string{}
	types := []string{}
	embeddingIDs := []string{}
	embeddingsFlat := [][]float32{}
	for i, chunk := range doc {
		ids = append(ids, chunk.UID())
		types = append(types, chunk.Metadata.Mimetype)
		// FIXME: hardcoded
		embeddingIDs = append(embeddingIDs, "Alibaba-NLP/gte-large-en-v1.5")
		embeddingsFlat = append(embeddingsFlat, embeddings[i]...)
	}
	idColumn := entity.NewColumnVarChar("ID", ids)
	typeColumn := entity.NewColumnVarChar("type", types)
	embeddingIDColumn := entity.NewColumnVarChar("embeddingID", embeddingIDs)
	embeddingColumn := entity.NewColumnFloatVector("embedding", 1024, embeddingsFlat)
	column, err := client.Insert(context.TODO(), westernBlotSchema.CollectionName, "", idColumn, typeColumn, embeddingIDColumn, embeddingColumn)
	if err != nil {
		panic(err)
	}
	// Now add index
	idx, err := entity.NewIndexIvfFlat(entity.L2, 2)
	if err != nil {
		logrus.Fatal("fail to create ivf flat index:", err.Error())
	}
	err = client.CreateIndex(context.TODO(), westernBlotSchema.CollectionName, "embedding", idx, false)
	if err != nil {
		log.Fatal("fail to create index:", err.Error())
	}

	err = client.LoadCollection(context.TODO(), westernBlotSchema.CollectionName, false)
	if err != nil {
		log.Fatal("failed to load collection:", err.Error())
	}

	fmt.Println(column)
}

// HACK: set some default databases and collections
func setupDefaults(client client.Client) error {
	exists, err := client.HasCollection(context.TODO(), westernBlotSchema.CollectionName)
	if err != nil {
		panic(err)
	}
	if !exists {
		logrus.Info("Westernblot collection doesn't exist, creating...")
		if err := client.CreateCollection(
			context.TODO(),
			westernBlotSchema,
			entity.DefaultShardNumber,
		); err != nil {
			panic(err)
		}
	} else {
		logrus.Info("Westernblot collection already exists")
	}

	return nil
}
