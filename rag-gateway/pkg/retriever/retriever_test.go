package retriever_test

import (
	"context"
	"fmt"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/retriever"
	mock_retriever "github.com/Synaptic-Lynx/rag-gateway/pkg/test/mock/retriever"
	mock_schema "github.com/Synaptic-Lynx/rag-gateway/pkg/test/mock/schema"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tmc/langchaingo/schema"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Retriever", Label("unit"), Ordered, func() {
	var baseRetrievers []retriever.Retriever
	BeforeAll(func() {
		ctrl := gomock.NewController(GinkgoT())

		retriever1 := mock_retriever.NewMockRetriever(ctrl)
		retriever2 := mock_retriever.NewMockRetriever(ctrl)

		schema1 := mock_schema.NewMockRetriever(ctrl)
		schema2 := mock_schema.NewMockRetriever(ctrl)

		schema1.EXPECT().GetRelevantDocuments(gomock.Any(), "query1").Return([]schema.Document{
			newSD("doc1", "hash1", 0, 0.9),
			newSD("doc2", "hash2", 0, 0.2),
			newSD("doc2", "hash2", 1, 0.4),
		}, nil).AnyTimes()
		schema2.EXPECT().GetRelevantDocuments(gomock.Any(), gomock.Any()).Return([]schema.Document{
			newSD("doc1", "hash1", 0, 0.5),
			newSD("doc2", "hash2", 0, 0.9),
			newSD("doc2", "hash2", 1, 1.0),
		}, nil).AnyTimes()

		schema1.EXPECT().GetRelevantDocuments(gomock.Any(), "query2").Return(
			[]schema.Document{},
			fmt.Errorf("failed to get documents"),
		).AnyTimes()

		retriever1.EXPECT().AsRetriever().Return(schema1).AnyTimes()
		retriever2.EXPECT().AsRetriever().Return(schema2).AnyTimes()

		baseRetrievers = []retriever.Retriever{retriever1, retriever2}
	})

	When("we use the hybrid document retriever", func() {
		It("should return the top documents", func() {
			fuser := retriever.NewRRFFusion(retriever.DefaultRRFK)

			hybridRet := retriever.NewHybridRetriever(baseRetrievers, 2, fuser)
			reranked, err := hybridRet.GetRelevantDocuments(context.TODO(), "query1")
			Expect(err).NotTo(HaveOccurred())
			Expect(reranked).To(HaveLen(2))
			Expect(reranked[0].PageContent).To(Equal("doc1-0"))
			Expect(reranked[0].Score).NotTo(BeNumerically("<=", 0))
			Expect(reranked[1].PageContent).To(Equal("doc2-1"))
			Expect(reranked[1].Score).NotTo(BeNumerically("<=", 0))
		})

		It("should surface errors if one of the retrievers fails", func() {
			fuser := retriever.NewRRFFusion(retriever.DefaultRRFK)

			hybridRet := retriever.NewHybridRetriever(baseRetrievers, 2, fuser)
			reranked, err := hybridRet.GetRelevantDocuments(context.TODO(), "query2")
			Expect(err).To(HaveOccurred())
			Expect(reranked).To(BeNil())
		})
	})
})
