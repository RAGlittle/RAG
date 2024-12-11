package collection

import "github.com/milvus-io/milvus-sdk-go/v2/entity"

// !! Changing the values of this map will result in breaking changes
var (
	IndexMappings = map[string]func(entity.MetricType) (entity.Index, error){
		"ivf_flat": func(mt entity.MetricType) (entity.Index, error) {
			return entity.NewIndexIvfFlat(mt, 2)
		},
	}

	MetricMappings = map[string]func() entity.MetricType{
		"L2": func() entity.MetricType { return entity.L2 },
	}
)
