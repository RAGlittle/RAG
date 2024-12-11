package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/collection"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	dir := "../data/westernblot/pdf"
	files, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	for _, file := range files {
		fmt.Println(file.Name())
	}

	minioClient, err := minio.New("localhost:8333", &minio.Options{
		Creds: credentials.NewStaticV4("admin", "admin", "admin"),
	})
	if err != nil {
		panic(err)
	}

	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		panic(err)
	}
	for _, bucket := range buckets {
		fmt.Println(bucket.Name)
	}

	manager := collection.NewSectionManager(context.TODO(), minioClient, client.Config{
		Address:  "localhost:19530",
		Password: "minioadmin",
		APIKey:   "minioadmin",
	},
		nil,
		nil,
	)

	docs := []collection.SectionDocument{}

	for _, file := range files {
		filePath := path.Join(dir, file.Name())
		f, err := os.Open(filePath)
		if err != nil {
			panic(err)
		}
		defer f.Close()
		data, err := io.ReadAll(f)
		if err != nil {
			panic(err)
		}
		docs = append(docs, collection.SectionDocument{
			DocID:   file.Name(),
			Content: data,
			Type:    "application/pdf",
		})
	}

	err = manager.UploadDocuments(context.Background(), "section1", docs)
	if err != nil {
		panic(err)
	}

	objs, err := manager.ListSectionObjects(context.Background(), "section1")
	if err != nil {
		panic(err)
	}
	fmt.Println("files uploaded to remote storage : ")
	fmt.Println(objs)
}
