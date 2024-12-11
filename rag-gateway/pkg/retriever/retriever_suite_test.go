package retriever_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRetriever(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Retriever Suite")
}
