package etcd_test

import (
	"context"
	"testing"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/etcd"
	. "github.com/Synaptic-Lynx/rag-gateway/pkg/test/conformance/storage"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util/future"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func TestEtcd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Etcd Suite")
}

var broker = future.New[testBroker]()

type testBroker struct {
	client *clientv3.Client
}

func (t testBroker) KeyValueStore(prefix string) storage.KeyValueStore {
	return etcd.NewKeyValueStore(t.client, prefix)
}

var _ = BeforeSuite(func() {
	ctx, ca := context.WithCancel(context.Background())
	DeferCleanup(func() {
		ca()
	})
	etcdC, err := StartEtcdContainer(ctx)
	Expect(err).NotTo(HaveOccurred())

	DeferCleanup(func() {
		etcdC.Container.Terminate(context.TODO())
	})

	client, err := clientv3.New(clientv3.Config{
		Endpoints: []string{etcdC.URI},
	})
	Expect(err).NotTo(HaveOccurred())

	t := testBroker{
		client: client,
	}
	broker.Set(t)

})

var _ = Describe("Etcd KV Store", Ordered, Label("integration"), KeyValueStoreTestSuite(broker, NewBytes, Equal))
var _ = Describe("Etcd Chat history", Label("unit"), Ordered, KeyValueHistoryTestSuite(broker))
