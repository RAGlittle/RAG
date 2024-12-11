package embedgw

import "fmt"

func (e *EmbedSpecificRequest) Validate() error {
	if e.EmbeddingID == "" {
		return fmt.Errorf("embedding id required")
	}
	return nil
}
