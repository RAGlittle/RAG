package collection_test

import (
	"bytes"
	"context"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/collection"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/storage/inmemory"
	"github.com/Synaptic-Lynx/rag-gateway/pkg/util"
	"github.com/google/uuid"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"google.golang.org/grpc/codes"
)

var (
	section1 = collection.SectionMetadata{
		SectionID:       "section1",
		UserDescription: "Contains documents related to XXX",
		Spec: []collection.EmbeddingSpec{
			{
				EmbeddingID:  "embedding1",
				IndexId:      "ivf_flat",
				MetricType:   "l2",
				CollectionId: "collection1",
				PartitionId:  "partition1",
			},
		},
		ChunkSpec: collection.ChunkSpec{
			ChunkSize:    100,
			ChunkOverlap: 50,
			CodeBlocks:   true,
		},
	}

	section2 = collection.SectionMetadata{
		SectionID:       "section2",
		UserDescription: "Contains documents related to XXX",
		Spec: []collection.EmbeddingSpec{
			{
				EmbeddingID:  "embedding1",
				IndexId:      "ivf_flat",
				MetricType:   "l2",
				CollectionId: "collection1",
				PartitionId:  "partition1",
			},
		},
		ChunkSpec: collection.ChunkSpec{
			ChunkSize:    100,
			ChunkOverlap: 50,
			CodeBlocks:   true,
		},
	}
)

var _ = Describe("Section Metadata Store", Ordered, Label("unit"), func() {
	var sd collection.SectionMetadataStore
	BeforeAll(func() {
		sd = collection.NewSectionManager(
			context.TODO(), nil,
			client.Config{},
			inmemory.NewKeyValueStore(bytes.Clone),
			nil,
		)
	})

	When("we use the manager", func() {
		It("should initially list empty section metadata", func() {
			keys, err := sd.ListSectionMetadata(context.TODO())
			Expect(err).To(Succeed())
			Expect(keys).To(BeEmpty())

			mds, err := sd.BatchSectionMetadata(context.TODO())
			Expect(err).To(Succeed())
			Expect(mds).To(BeEmpty())
		})

		It("should persist section metadata", func() {
			err := sd.CreateSectionMetadata(context.TODO(), "section1", section1)
			Expect(err).To(Succeed())

			By("expect the metadata to be fetchable after being persisted")

			res, err := sd.GetSectionMetadata(context.TODO(), "section1")
			Expect(err).To(Succeed())
			Expect(res).To(Equal(section1))
		})

		It("should prevent creating duplicate section metadata", func() {
			err := sd.CreateSectionMetadata(context.TODO(), "section1", section2)
			Expect(err).To(HaveOccurred())
			Expect(util.StatusCode(err)).To(Equal(codes.AlreadyExists))
		})

		It("should update section metadata", func() {
			// TODO
		})

		It("should list/batch section metadata", func() {

			err := sd.CreateSectionMetadata(context.TODO(), "section2", section2)

			Expect(err).To(Succeed())

			sectionIds, err := sd.ListSectionMetadata(context.TODO())
			Expect(err).To(Succeed())
			Expect(sectionIds).To(ConsistOf("section1", "section2"))

			sections, err := sd.BatchSectionMetadata(context.TODO())
			Expect(err).To(Succeed())

			Expect(sections).To(ConsistOf(
				section1,
				section2,
			))
		})

		It("should delete section metadata", func() {
			err := sd.DeleteSectionMetadata(context.TODO(), "section1")
			Expect(err).To(Succeed())

			By("expecting the metadata to be deleted")
			_, err = sd.GetSectionMetadata(context.TODO(), "section1")
			Expect(err).To(HaveOccurred())
			Expect(util.StatusCode(err)).To(Equal(codes.NotFound))

			keys, err := sd.ListSectionMetadata(context.TODO())
			Expect(err).To(Succeed())
			Expect(keys).To(ConsistOf("section2"))

			mds, err := sd.BatchSectionMetadata(context.TODO())
			Expect(err).To(Succeed())
			Expect(mds).To(ConsistOf(section2))
		})
	})

	When("we pass in invalid data to section metadata service", func() {
		It("should fail to create section metadata", func() {
			for _, s := range invalidSections {
				err := sd.CreateSectionMetadata(context.TODO(), s.SectionID, s)
				Expect(err).To(HaveOccurred())
				Expect(util.StatusCode(err)).To(Equal(codes.InvalidArgument))

				_, err = sd.GetSectionMetadata(context.TODO(), s.SectionID)
				Expect(err).To(HaveOccurred())

				var expectedCode codes.Code
				if s.SectionID == "" {
					expectedCode = codes.InvalidArgument
				} else {
					expectedCode = codes.NotFound
				}
				Expect(util.StatusCode(err)).To(Equal(expectedCode))

				changed, err := sd.UpdateSectionMetadata(context.TODO(), s.SectionID, s)
				Expect(changed).To(BeFalse())
				Expect(err).To(HaveOccurred())
				Expect(util.StatusCode(err)).To(Equal(codes.InvalidArgument))
			}
		})
	})
})

var invalidSections = []collection.SectionMetadata{
	{
		SectionID: "",
	},
	{
		SectionID: uuid.New().String(),
		Spec:      []collection.EmbeddingSpec{},
	},
	{
		SectionID: uuid.New().String(),
		Spec: []collection.EmbeddingSpec{
			{
				EmbeddingID: "",
			},
		},
	},
	{
		SectionID: uuid.New().String(),
		Spec: []collection.EmbeddingSpec{
			{
				EmbeddingID: "embedding1",
				IndexId:     "",
			},
		},
	},
	{
		SectionID: uuid.New().String(),
		Spec: []collection.EmbeddingSpec{
			{
				EmbeddingID: "embedding1",
				IndexId:     "ivf_flat",
				MetricType:  "",
			},
		},
	},
	{
		SectionID: uuid.New().String(),
		Spec: []collection.EmbeddingSpec{
			{
				EmbeddingID:  "embedding1",
				IndexId:      "ivf_flat",
				MetricType:   "l2",
				CollectionId: "",
			},
		},
	},
}
