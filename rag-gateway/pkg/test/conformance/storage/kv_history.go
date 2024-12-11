package conformance_storage

import (
	"context"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/history"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util/future"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/tmc/langchaingo/memory"
)

func KeyValueHistoryTestSuite[B storage.KeyValueStoreTBroker[[]byte]](
	tsF future.Future[B],
) func() {
	return func() {
		var kv storage.KeyValueStore
		BeforeAll(func() {
			kv = tsF.Get().KeyValueStore("chats")
		})

		When("We use KV stores as QA chain memory", func() {
			It("should have an empty history by default", func() {
				mem := memory.NewConversationBuffer(
					memory.WithChatHistory(
						history.NewKVChatHistory(kv, "chat1"),
					),
				)
				res1, err := mem.LoadMemoryVariables(context.TODO(), map[string]any{})
				Expect(err).NotTo(HaveOccurred())
				Expect(res1).To(Equal(map[string]any{
					"history": "",
				}))
			})

			It("should save and load memory variables", func() {
				mem := memory.NewConversationBuffer(
					memory.WithChatHistory(
						history.NewKVChatHistory(kv, "chat1"),
					),
				)
				err := mem.SaveContext(context.TODO(), map[string]any{
					"foo": "bar",
				}, map[string]any{
					"bar": "foo",
				})
				Expect(err).NotTo(HaveOccurred())
				res1, err := mem.LoadMemoryVariables(context.TODO(), map[string]any{})
				Expect(err).NotTo(HaveOccurred())
				Expect(res1).To(Equal(map[string]any{
					"history": "Human: bar\nAI: foo",
				}))
			})

			It("should load older memory variables", func() {
				mem := memory.NewConversationBuffer(
					memory.WithChatHistory(
						history.NewKVChatHistory(kv, "chat1"),
					),
				)
				result, err := mem.LoadMemoryVariables(context.TODO(), map[string]any{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(map[string]any{
					"history": "Human: bar\nAI: foo",
				}))
			})

			It("should add to an existing conversation", func() {
				mem := memory.NewConversationBuffer(
					memory.WithChatHistory(
						history.NewKVChatHistory(kv, "chat1"),
					),
				)
				result, err := mem.LoadMemoryVariables(context.TODO(), map[string]any{})
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(map[string]any{
					"history": "Human: bar\nAI: foo",
				}))

				err = mem.SaveContext(context.TODO(), map[string]any{
					"foo": "bar2",
				}, map[string]any{
					"bar": "foo2",
				})

				Expect(err).NotTo(HaveOccurred())

				res1, err := mem.LoadMemoryVariables(context.TODO(), map[string]any{})
				Expect(err).NotTo(HaveOccurred())
				Expect(res1).To(Equal(map[string]any{
					"history": "Human: bar\nAI: foo\nHuman: bar2\nAI: foo2",
				}))

			})
		})
	}
}
