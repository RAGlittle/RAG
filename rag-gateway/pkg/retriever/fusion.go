package retriever

import (
	"sort"

	"github.com/Synaptic-Lynx/rag-gateway/pkg/fileutil"
	"github.com/samber/lo"
	"github.com/tmc/langchaingo/schema"
)

// TODO : implement reciprocal rank fusion

const (
	DefaultRRFK = 60
)

type scoreItem struct {
	key   string
	score float64
}

type Fusion interface {
	Merge(docs [][]schema.Document, topK int) []schema.Document
}

type rrfFusion struct {
	rrfK int
}

func NewRRFFusion(rrfK int) Fusion {
	if rrfK == 0 {
		rrfK = DefaultRRFK
	}
	return &rrfFusion{rrfK: rrfK}
}

func (r *rrfFusion) Merge(docs [][]schema.Document, topK int) []schema.Document {
	scores := make(map[string]float64)
	docMappings := make(map[string]schema.Document)
	for _, docSet := range docs {
		for _, doc := range docSet {
			chunk := fileutil.FromSchemaDoc(doc)
			if _, ok := docMappings[chunk.UID()]; !ok {
				docMappings[chunk.UID()] = doc
			}
			if _, ok := scores[chunk.UID()]; !ok {
				scores[chunk.UID()] = rrfScore(r.rrfK, chunk.Score)
			} else {
				scores[chunk.UID()] += rrfScore(r.rrfK, chunk.Score)
			}
		}
	}

	scoreList := lo.MapToSlice(scores, func(key string, value float64) scoreItem {
		return scoreItem{key, value}
	})

	sort.Slice(scoreList, func(i, j int) bool {
		return scoreList[j].score > scoreList[i].score
	})

	// Extract the topK keys
	var keys []string
	for i := 0; i < topK && i < len(scoreList); i++ {
		keys = append(keys, scoreList[i].key)
	}

	var topDocs []schema.Document
	for _, key := range keys {
		doc := docMappings[key]
		// FIXME: check clamping is safe here
		doc.Score = float32(scores[key])
		topDocs = append(topDocs, doc)
	}
	return topDocs
}

func rrfScore(rrfK int, score float64) float64 {
	return 1 / (float64(rrfK) + score)
}
