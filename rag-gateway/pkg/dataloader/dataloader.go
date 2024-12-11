package dataloader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/collection"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/fileutil"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/render"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/server/embed"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/samber/lo"
	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
	"github.com/tmc/langchaingo/vectorstores"
	"google.golang.org/grpc/codes"
)

type DataLoader interface {
	Load(ctx context.Context, dataDir string) error
}

type localDataLoader struct {
	ctx     context.Context
	embedGw *embed.EmbedGatewayServer
	sm      collection.SectionManager
	md      collection.SectionMetadata
}

func NewDataloader(ctx context.Context,
	sm collection.SectionManager,
	gw *embed.EmbedGatewayServer,
	md collection.SectionMetadata,
) DataLoader {
	return &localDataLoader{
		ctx:     ctx,
		sm:      sm,
		embedGw: gw,
		md:      md,
	}
}

func (l *localDataLoader) readOne(filepath string) ([]byte, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return io.ReadAll(file)
}

func (l *localDataLoader) Load(ctx context.Context, dataDir string) error {
	logrus.Info("creating section metadata data...")
	err := l.sm.CreateSectionMetadata(ctx, l.md.SectionID, l.md)

	if err != nil && util.StatusCode(err) != codes.AlreadyExists {
		// FIXME: only for this hacky implementation we check can continue if it already exists
		return err
	}

	logrus.Infof("loading data from %s", dataDir)
	files, err := l.fetchDocumentInfo(dataDir)
	if err != nil {
		return err
	}

	logrus.Infof("uploading documents to object store...")

	sectionDocs := []collection.SectionDocument{}

	for _, files := range files {
		data, err := l.readOne(files.Path)
		if err != nil {
			return err
		}

		sectionDocs = append(sectionDocs, collection.SectionDocument{
			DocID:   path.Base(files.Path),
			Content: data,
			Type:    files.Mimetype,
		})
	}

	logrus.Infof("uploading %d documents to object store...", len(sectionDocs))
	if err := l.sm.UploadDocuments(ctx, l.md.SectionID, sectionDocs); err != nil {
		return err
	}

	chunks, err := l.chunkDocuments(files)
	if err != nil {
		return err
	}
	logrus.Infof("got %d chunks to process", len(chunks))
	schemas := l.ReduceChunksToSchema(chunks)

	// To add support for

	logrus.Infof("adding %d documents to vector store(s)...", len(schemas))
	// TODO : implement a deduplicator fn to pass into option here

	vectorstores := []vectorstores.VectorStore{}

	for _, e := range l.md.Spec {
		logrus.Infof("creating vector store for %v...", e)
		v, err := l.sm.ToVectorStore(e)
		if err != nil {
			return err
		}
		vectorstores = append(vectorstores, v)
	}

	for _, v := range vectorstores {
		docs, err := v.AddDocuments(ctx, schemas)
		if err != nil {
			return err
		}
		// Note : this returns 0 due to a missing implementation detail in langchaingo
		logrus.Infof("added %d document embeddings to vector store", len(docs))
	}

	return nil
}

func (l *localDataLoader) fetchDocumentInfo(dataDir string) ([]fileutil.DocumentMetadata, error) {
	files, err := fileutil.GetDocumentMetadata(dataDir)
	if err != nil {
		return nil, err
	}
	w := table.NewWriter()
	// w.SetStyle(table.StyleColoredDark)
	w.AppendRow(table.Row{"DocID", "Hash", "Mimetype"})
	for _, f := range files {
		w.AppendRow(table.Row{
			render.ClampString(f.DocID, 50, "..."),
			f.Hash,
			f.Mimetype})
	}
	fmt.Println(w.Render())
	return files, nil
}

func (l *localDataLoader) chunkDocuments(files []fileutil.DocumentMetadata) ([]fileutil.DocumentChunk, error) {
	allChunks := []fileutil.DocumentChunk{}
	logrus.Infof("Chunking documents with chunk size : %d, overlap : %d", l.md.ChunkSpec.ChunkSize, l.md.ChunkSpec.ChunkOverlap)
	for _, f := range files {
		chunk, err := fileutil.LoadDocs(f, fileutil.WithTextSplitter(textsplitter.NewRecursiveCharacter(
			textsplitter.WithChunkSize(l.md.ChunkSpec.ChunkSize),
			textsplitter.WithChunkOverlap(l.md.ChunkSpec.ChunkOverlap),
		)))
		if err != nil {
			panic(err)
		}
		allChunks = append(allChunks, chunk...)
	}

	return allChunks, nil
}

func (l *localDataLoader) ReduceChunksToSchema(chunks []fileutil.DocumentChunk) []schema.Document {
	return lo.Map(chunks, func(c fileutil.DocumentChunk, _ int) schema.Document {
		return c.AsSchemaDoc()
	})
}
