package fileutil

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/tmc/langchaingo/documentloaders"
	"github.com/tmc/langchaingo/schema"
	"github.com/tmc/langchaingo/textsplitter"
)

type DocumentChunk struct {
	Metadata   DocumentMetadata `json:"metadata"`
	Contents   string           `json:"contents"`
	ChunkIndex int              `json:"chunk_index"`
	PDF        *PDFMetadata     `json:"pdf"`
	Score      float64          `json:"score"`
}

func (d DocumentChunk) AsSchemaDoc() schema.Document {
	doc := schema.Document{
		PageContent: d.Contents,
		Metadata: map[string]interface{}{
			"doc_id":      d.Metadata.DocID,
			"hash":        d.Metadata.Hash,
			"chunk_index": float64(d.ChunkIndex),
		},
		// FIXME: check score clamping is safe here
		Score: float32(d.Score),
	}
	if d.PDF != nil {
		doc.Metadata["cur_page"] = d.PDF.CurPage
		doc.Metadata["last_page"] = d.PDF.LastPage
	}
	return doc
}

func FromSchemaDoc(doc schema.Document) DocumentChunk {
	var pdf *PDFMetadata
	if doc.Metadata["cur_page"] != nil {
		pdf = &PDFMetadata{
			CurPage:  int(doc.Metadata["cur_page"].(float64)),
			LastPage: int(doc.Metadata["last_page"].(float64)),
		}
	}
	return DocumentChunk{
		Metadata: DocumentMetadata{
			DocID:    doc.Metadata["doc_id"].(string),
			Hash:     doc.Metadata["hash"].(string),
			Mimetype: "application/pdf",
		},
		Contents:   doc.PageContent,
		ChunkIndex: int(doc.Metadata["chunk_index"].(float64)),
		PDF:        pdf,
		Score:      float64(doc.Score),
	}
}

func (d DocumentChunk) UID() string {
	return strings.Join(
		[]string{
			d.Metadata.DocID,
			d.Metadata.Hash,
			fmt.Sprintf("%d", d.ChunkIndex),
		},
		"-",
	)
}

type PDFMetadata struct {
	CurPage  int
	LastPage int
}

type LoadOptions struct {
	textsplitter.TextSplitter
}

type LoadOption = func(o *LoadOptions)

func WithTextSplitter(spl textsplitter.TextSplitter) LoadOption {
	return func(o *LoadOptions) {
		o.TextSplitter = spl
	}
}

func LoadDocs(doc DocumentMetadata, opts ...LoadOption) ([]DocumentChunk, error) {
	opt := &LoadOptions{}
	for _, o := range opts {
		o(opt)
	}
	logrus.Tracef(
		"loading document from %s",
		doc.Path,
	)
	f, err := os.Open(doc.Path)
	if err != nil {
		return nil, err

	}
	switch doc.Mimetype {
	case "application/pdf":
		return loadPDF(f, doc, opt)
	default:
		logrus.Errorf("unsupported mimetype %s", doc.Mimetype)
	}

	return nil, nil
}

func loadPDF(f io.ReaderAt, doc DocumentMetadata, options *LoadOptions) ([]DocumentChunk, error) {
	pdfLoader := documentloaders.NewPDF(f, doc.Size)
	var schemas []schema.Document
	var err error
	if options.TextSplitter == nil {
		schemas, err = pdfLoader.Load(context.TODO())
	} else {
		schemas, err = pdfLoader.LoadAndSplit(context.TODO(), options.TextSplitter)
	}
	if err != nil {
		return nil, err
	}
	chunks := []DocumentChunk{}
	for i, schema := range schemas {
		chunks = append(chunks, DocumentChunk{
			Metadata:   doc,
			Contents:   schema.PageContent,
			ChunkIndex: i,
			PDF: &PDFMetadata{
				CurPage:  schema.Metadata["page"].(int),
				LastPage: schema.Metadata["total_pages"].(int),
			},
		})

	}
	return chunks, nil
}
