package embeddings_test

import (
	"github.com/Synaptic-Lynx/rag-gateway/pkg/embeddings"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Embedding helpers", Label("unit"), func() {
	When("we split texts into payload limits", func() {
		It("should split empty texts", func() {
			texts := []string{}
			limit := 20000
			res := embeddings.SplitTexts(texts, limit)
			Expect(res).To(HaveLen(0))
		})

		It("should split texts into payload limits", func() {
			texts := []string{"aaaaa", "bbbbb", "ccccc", "ddddd", "eeeee", "fffff"}
			limit := 10
			res := embeddings.SplitTexts(texts, limit)
			Expect(res).To(HaveLen(3))
			Expect(res[0]).To(HaveLen(2))
			Expect(res[1]).To(HaveLen(2))
			Expect(res[2]).To(HaveLen(2))
			Expect(res[0]).To(ConsistOf("aaaaa", "bbbbb"))
			Expect(res[1]).To(ConsistOf("ccccc", "ddddd"))
			Expect(res[2]).To(ConsistOf("eeeee", "fffff"))
		})
	})
})

var _ = Describe("Embedding client", func() {
	When("we create a new rerank client", func() {
		It("should create a new rerank client", func() {
			endpoint := "http://localhost:8080"
			res := embeddings.NewPoolReRanker(endpoint)
			Expect(res).NotTo(BeNil())

		})
	})
})

var _ = Describe("ReRanker client", func() {
	When("we create a new rerank client", func() {
		It("should create a new rerank client", func() {
			endpoint := "http://localhost:8080"
			res := embeddings.NewPoolReRanker(endpoint)
			Expect(res).NotTo(BeNil())
		})
	})
})
