package inmemory_test

import (
	"bytes"
	"testing"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/inmemory"
	. "github.com/Synaptic-Lynx/rag-gateway/pkg/test/conformance/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util/future"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInmemory(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Inmemory Suite")
}

type testBroker struct{}

func (t testBroker) KeyValueStore(string) storage.KeyValueStore {
	return inmemory.NewKeyValueStore(bytes.Clone)
}

var _ = Describe("In-memory KV Store", Ordered, Label("unit"), KeyValueStoreTestSuite(future.Instant(testBroker{}), NewBytes, Equal))
var _ = Describe("In-memory Chat history", Label("unit"), Ordered, KeyValueHistoryTestSuite(future.Instant(testBroker{})))
