package collection

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SectionObjectStore interface {
	ListSectionObjects(ctx context.Context, sectionId string) (objectIds []string, err error)
	GetSectionObject(ctx context.Context, sectionId string, objectKey string) ([]byte, error)
	DeleteSectionObject(ctx context.Context, sectionId string, objectKey string) (err error)
	DeleteAllSectionObjects(ctx context.Context, sectionId string) (err error)
	UploadDocuments(ctx context.Context, sectionId string, docs []SectionDocument) error
}

func (s *sectionManager) ListSectionObjects(ctx context.Context, sectionId string) (ids []string, err error) {
	if sectionId == "" {
		return nil, status.Error(codes.InvalidArgument, "section id required")
	}

	ok, err := s.objectClient.BucketExists(ctx, sectionId)
	if err != nil {
		return nil, err
	}
	if !ok {
		logrus.Warnf("section bucket not found, no documents associated with give section ID: %s", sectionId)
		return []string{}, nil
	}

	objects := s.objectClient.ListObjects(ctx, sectionId, minio.ListObjectsOptions{})
	keys := []string{}
	for obj := range objects {
		if obj.Err != nil {
			return nil, obj.Err
		}
		keys = append(keys, obj.Key)
	}
	return keys, nil
}

func (s *sectionManager) uploadOne(
	ctx context.Context,
	doc io.Reader,
	sectionId string,
	documentKey string,
	docLen int64,
	docType string,
) error {
	_, err := s.objectClient.PutObject(
		ctx,
		sectionId,
		documentKey,
		doc,
		docLen,
		minio.PutObjectOptions{
			UserMetadata: map[string]string{
				"DocType": docType,
			},
		},
	)
	return err
}

type SectionDocument struct {
	DocID   string `json:"doc_id"`
	Content []byte `json:"doc"`
	Type    string `json:"type"`
}

func (s *SectionDocument) Validate() error {
	if s.DocID == "" {
		return fmt.Errorf("doc id required")
	}
	if len(s.Content) == 0 {
		return fmt.Errorf("no document content provided")
	}
	if s.Type == "" {
		return fmt.Errorf("doc type required")
	}
	return nil
}

func (s *sectionManager) UploadDocuments(ctx context.Context, sectionId string, docs []SectionDocument) error {
	if sectionId == "" {
		return status.Error(codes.InvalidArgument, "section id required")
	}

	if len(docs) == 0 {
		return status.Error(codes.InvalidArgument, "no documents provided")
	}
	for _, doc := range docs {
		if err := doc.Validate(); err != nil {
			return status.Error(codes.InvalidArgument, err.Error())
		}
	}

	ok, err := s.objectClient.BucketExists(ctx, sectionId)
	if err != nil {
		return err
	}
	if !ok {
		err = s.objectClient.MakeBucket(ctx, sectionId, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	// TODO : can parallelize/ pool this
	for _, doc := range docs {
		err = s.uploadOne(ctx, bytes.NewReader(doc.Content), sectionId, doc.DocID, int64(len(doc.Content)), doc.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *sectionManager) GetSectionObject(ctx context.Context, sectionId string, objectKey string) ([]byte, error) {
	obj, err := s.objectClient.GetObject(ctx, sectionId, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *sectionManager) DeleteSectionObject(ctx context.Context, sectionId string, objectKey string) (err error) {
	if sectionId == "" {
		return status.Error(codes.InvalidArgument, "section id required")
	}
	if objectKey == "" {
		return status.Error(codes.InvalidArgument, "object key required")
	}
	return s.objectClient.RemoveObject(ctx, sectionId, objectKey, minio.RemoveObjectOptions{})
}

func (s *sectionManager) DeleteAllSectionObjects(ctx context.Context, sectionId string) (err error) {
	if sectionId == "" {
		return status.Error(codes.InvalidArgument, "section id required")
	}

	ok, err := s.objectClient.BucketExists(ctx, sectionId)
	if err != nil {
		return err
	}
	if !ok {
		logrus.Warnf("requested delete with no documents associated with give section ID: %s", sectionId)
		return nil
	}

	objects := s.objectClient.ListObjects(ctx, sectionId, minio.ListObjectsOptions{})
	for obj := range objects {
		if obj.Err != nil {
			return obj.Err
		}
		err = s.objectClient.RemoveObject(ctx, sectionId, obj.Key, minio.RemoveObjectOptions{})
		if err != nil {
			return err
		}
	}

	return s.objectClient.RemoveBucket(ctx, sectionId)
}
