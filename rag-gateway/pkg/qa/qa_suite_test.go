package qa_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestQa(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Qa Suite")
}
