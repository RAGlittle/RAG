package collection_test

import (
	"context"
	"strings"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/collection"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	miniotest "github.com/testcontainers/testcontainers-go/modules/minio"
)

var _ = Describe("We we use the section manager object store", Ordered, Label("integration"), func() {
	var o collection.SectionObjectStore
	BeforeAll(func() {
		ctx := context.Background()

		minioContainer, err := miniotest.Run(ctx, "minio/minio:RELEASE.2024-01-16T16-07-38Z")
		Expect(err).To(Succeed())

		DeferCleanup(func() {
			minioContainer.Terminate(ctx)
		})

		endpoint, err := minioContainer.Endpoint(ctx, "http")
		Expect(err).To(Succeed())

		endpoint = strings.TrimPrefix(endpoint, "http://")
		GinkgoWriter.Write([]byte("Minio endpoint: " + endpoint + "\n"))
		minioClient, err := minio.New(endpoint, &minio.Options{
			Creds: credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		})

		Expect(err).To(Succeed())

		_, err = minioClient.ListBuckets(ctx)
		Expect(err).To(Succeed())

		s := collection.NewSectionManager(
			context.TODO(),
			minioClient,
			client.Config{},
			nil,
			nil,
		)

		o = s

	})

	When("we use the object store", func() {
		It("should initially have no sections", func() {
			objs, err := o.ListSectionObjects(context.Background(), "section1")
			Expect(err).To(Succeed())

			Expect(objs).To(BeEmpty())
		})

		It("should upload documents", func() {
			err := o.UploadDocuments(context.Background(), "section1", []collection.SectionDocument{
				{
					DocID:   "doc1",
					Content: []byte("hello world"),
					Type:    "application/octet-stream",
				},
				{
					DocID:   "doc2",
					Content: []byte("foo bar"),
					Type:    "application/octet-stream",
				},
			})
			Expect(err).To(Succeed())

			objs, err := o.ListSectionObjects(context.Background(), "section1")
			Expect(err).To(Succeed())

			Expect(objs).To(ConsistOf("doc1", "doc2"))

			data, err := o.GetSectionObject(context.Background(), "section1", "doc1")
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte("hello world")))

			data, err = o.GetSectionObject(context.Background(), "section1", "doc2")
			Expect(err).To(Succeed())
			Expect(data).To(Equal([]byte("foo bar")))
		})

		It("should delete section objects", func() {
			err := o.DeleteSectionObject(context.Background(), "section1", "doc1")
			Expect(err).To(Succeed())

			objs, err := o.ListSectionObjects(context.Background(), "section1")
			Expect(err).To(Succeed())

			Expect(objs).To(ConsistOf("doc2"))
		})

		It("should delete all section objects", func() {
			err := o.DeleteAllSectionObjects(context.Background(), "section1")
			Expect(err).To(Succeed())

			objs, err := o.ListSectionObjects(context.Background(), "section1")
			Expect(err).To(Succeed())

			Expect(objs).To(BeEmpty())
		})
	})
})
