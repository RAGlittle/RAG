package retriever_test

import (
	"fmt"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/fileutil"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/retriever"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tmc/langchaingo/schema"
)

func newSD(id, hash string, chunkIndex int, score float64) schema.Document {
	return (&fileutil.DocumentChunk{
		Contents: fmt.Sprintf("%s-%d", id, chunkIndex),
		Score:    score,
		Metadata: fileutil.DocumentMetadata{
			DocID: id,
			Hash:  hash,
		},
		ChunkIndex: chunkIndex,
	}).AsSchemaDoc()
}

var _ = Describe("Reciprocal Rank Fusion", func() {
	When("we have a list of documents", func() {
		It("should return the top documents", func() {

			docSet := [][]schema.Document{
				// doc set 1
				{
					newSD("doc1", "hash1", 0, 0.9),
					newSD("doc2", "hash2", 0, 0.2),
					newSD("doc2", "hash2", 1, 0.4),
				},
				// doc set 2
				{
					newSD("doc1", "hash1", 0, 0.5),
					newSD("doc2", "hash2", 0, 0.9),
					newSD("doc2", "hash2", 1, 1.0),
				},
			}

			fuser := retriever.NewRRFFusion(retriever.DefaultRRFK)
			reranked := fuser.Merge(docSet, 2)
			Expect(reranked).To(HaveLen(2))
			Expect(reranked[0].PageContent).To(Equal("doc1-0"))
			Expect(reranked[0].Score).NotTo(BeNumerically("<=", 0))
			Expect(reranked[1].PageContent).To(Equal("doc2-1"))
			Expect(reranked[1].Score).NotTo(BeNumerically("<=", 0))
		})
	})
})
