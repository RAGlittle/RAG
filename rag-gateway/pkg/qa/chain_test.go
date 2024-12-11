package qa_test

import (
	"bytes"
	"context"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/qa"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/retriever"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/history"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/inmemory"
	mock_llms "github.com/Synaptic-Lynx/rag-gateway/pkg/test/mock/llms"
	mock_vectorstores "github.com/Synaptic-Lynx/rag-gateway/pkg/test/mock/vectorstores"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tmc/langchaingo/chains"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/memory"
	"github.com/tmc/langchaingo/schema"
	"go.uber.org/mock/gomock"
)

var _ = Describe("QA chain", Ordered, Label("unit"), func() {

	var c chains.Chain

	BeforeAll(func() {

		ctrl := gomock.NewController(GinkgoT())

		vStore := mock_vectorstores.NewMockVectorStore(ctrl)

		vStore.EXPECT().SimilaritySearch(
			gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		).Return([]schema.Document{
			{
				PageContent: "hello world",
				Metadata:    map[string]any{"page": 1},
				Score:       0.5,
			},
		}, nil).AnyTimes()

		llm := mock_llms.NewMockModel(ctrl)

		llm.EXPECT().Call(gomock.Any(), gomock.Any()).Return(
			"Hello from an LLM Call", nil,
		).AnyTimes()

		llm.EXPECT().GenerateContent(gomock.Any(), gomock.Any(), gomock.Any()).Return(
			&llms.ContentResponse{
				Choices: []*llms.ContentChoice{
					{
						Content: "Hello from an LLM GenerateContent",
					},
				},
			}, nil,
		).AnyTimes()

		chatkv := inmemory.NewKeyValueStore(bytes.Clone)
		ret := retriever.NewSimilarityRetriever(vStore)
		mem := memory.NewConversationBuffer(
			memory.WithChatHistory(
				history.NewKVChatHistory(chatkv, "chat1"),
			),
			memory.WithReturnMessages(true),
		)
		c = qa.NewQA(llm, ret, mem)
	})

	When("when we use the RAG QA chain", func() {
		It("should successively run the retrieval QA chain", func() {
			Expect(true).To(BeTrue())

			initialQuestion := "what is the meaning of life?"
			initialResponse, err := chains.Run(context.TODO(), c, initialQuestion)
			Expect(err).To(Succeed())
			Expect(initialResponse).NotTo(BeEmpty())

			followUpQuestion := "Is the number 42 connected to that?"
			followUpResponse, err := chains.Run(context.TODO(), c, followUpQuestion)
			Expect(err).To(Succeed())
			Expect(followUpResponse).NotTo(BeEmpty())
		})
	})
})
