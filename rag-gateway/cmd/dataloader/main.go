package main

import (
	"context"
	"io"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/collection"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/dataloader"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/embed"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/etcd"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/version"
	"github.com/google/uuid"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	clientv3 "go.etcd.io/etcd/client/v3"
	"gopkg.in/yaml.v3"
)

// FIXME: hack for in memory broker
type hackBroker struct {
	client *clientv3.Client
}

func (h hackBroker) KeyValueStore(prefix string) storage.KeyValueStore {
	return etcd.NewKeyValueStore(h.client, prefix)
}

type PartialSectionMetadata struct {
	SectionId        string `yaml:"section_id"`
	user_description string `yaml:"user_description"`
	ChunkSpec        struct {
		ChunkSize    int  `yaml:"chunk_size"`
		ChunkOverlap int  `yaml:"chunk_overlap"`
		CodeBlocks   bool `yaml:"code_blocks"`
	} `yaml:"chunk_spec"`
}

type DataloaderConfig struct {
	DataDir                string                 `yaml:"data_dir"`
	EmbedderConfig         embed.EmbedderConfig   `yaml:"embedding_config"`
	EtcdEndpoints          []string               `yaml:"etcd_endpoints"`
	MetricType             string                 `yaml:"metric_type"`
	Index                  string                 `yaml:"index"`
	PartialSectionMetadata PartialSectionMetadata `yaml:"section_metadata"`
	MilvusEndpoint         string                 `yaml:"milvus_endpoint"`
	MilvusPassword         string                 `yaml:"milvus_password"`
	MilvusKey              string                 `yaml:"milvus_key"`
}

func BuildDataloaderCmd() *cobra.Command {
	var configFile string
	cmd := &cobra.Command{
		Use:     "dataloader",
		Version: version.FriendlyVersion(),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			f, err := os.Open(configFile)
			if err != nil {
				logrus.Errorf("error opening config file : %s", err)
				return err
			}
			defer f.Close()

			data, err := io.ReadAll(f)
			if err != nil {
				logrus.Errorf("error reading config file : %s", err)
			}
			cfg := &DataloaderConfig{}
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return err
			}

			etcdClient, err := clientv3.New(clientv3.Config{
				Endpoints: cfg.EtcdEndpoints,
			})
			if err != nil {
				return err
			}

			broker := hackBroker{client: etcdClient}

			metadataKv := broker.KeyValueStore("metadata")
			minioClient, err := minio.New("localhost:8333", &minio.Options{
				Creds: credentials.NewStaticV4("admin", "admin", "admin"),
			})
			if err != nil {
				panic(err)
			}

			embedGw, err := embed.NewEmbedGatewayServer(cfg.EmbedderConfig)
			if err != nil {
				return err
			}
			manager := collection.NewSectionManager(context.TODO(), minioClient, client.Config{
				Address:  cfg.MilvusEndpoint,
				Password: cfg.MilvusPassword,
				APIKey:   cfg.MilvusKey,
			}, metadataKv, embedGw)

			specs := lo.Map(cfg.EmbedderConfig.Embedders, func(e embed.EmbedderSpec, _ int) collection.EmbeddingSpec {
				return collection.EmbeddingSpec{
					EmbeddingID:  e.ModelId,
					IndexId:      cfg.Index,
					MetricType:   cfg.MetricType,
					CollectionId: strings.ReplaceAll(cfg.PartialSectionMetadata.SectionId+"-"+uuid.New().String(), "-", "_"),
					PartitionId:  "",
				}
			})

			sectionMd := collection.SectionMetadata{
				SectionID:       cfg.PartialSectionMetadata.SectionId,
				UserDescription: cfg.PartialSectionMetadata.user_description,
				ChunkSpec: collection.ChunkSpec{
					ChunkSize:    cfg.PartialSectionMetadata.ChunkSpec.ChunkSize,
					ChunkOverlap: cfg.PartialSectionMetadata.ChunkSpec.ChunkOverlap,
					CodeBlocks:   cfg.PartialSectionMetadata.ChunkSpec.CodeBlocks,
				},
				Spec: specs,
			}

			dataloader := dataloader.NewDataloader(ctx, manager, embedGw, sectionMd)
			if err := dataloader.Load(ctx, cfg.DataDir); err != nil {
				logrus.Errorf("error loading data : %s", err)
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVarP(&configFile, "config", "f", "./dataloader.yaml", "path to dataloader config file")
	return cmd
}

func main() {
	BuildDataloaderCmd().Execute()
}
