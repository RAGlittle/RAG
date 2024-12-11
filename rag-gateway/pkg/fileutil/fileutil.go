package fileutil

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
)

type DocumentMetadata struct {
	// Path is the local file path for this file
	Path string `json:"path"`
	// DocID is the title of the document
	DocID string `json:"doc_id"`
	// Hash is the hash uniquely identifying the contents
	Hash     string `json:"hash"`
	Mimetype string `json:"mimetype"`
	Size     int64  `json:"size"`
}

func FileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

func DetectMimeType(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", err
	}

	// Detect MIME type
	mimeType := http.DetectContentType(buffer)
	return mimeType, nil
}

func GetDocumentMetadata(dir string) ([]DocumentMetadata, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return []DocumentMetadata{}, err
	}
	ret := []DocumentMetadata{}
	var errs []error
	for _, file := range files {
		filePath := path.Join(dir, file.Name())

		fileInfo, err := os.Stat(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		hash, err := FileHash(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		mt, err := DetectMimeType(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		d := DocumentMetadata{
			Path:     filePath,
			DocID:    file.Name(),
			Hash:     hash,
			Mimetype: mt,
			Size:     fileInfo.Size(),
		}
		ret = append(ret, d)
	}
	return ret, errors.Join(errs...)
}
