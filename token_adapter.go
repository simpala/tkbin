package tkbin

import (
	"fmt"
	"github.com/pkoukk/tiktoken-go"
)

// TiktokenAdapter implements the Tokenizer interface for tiktoken encodings.
type TiktokenAdapter struct {
	encoder *tiktoken.Tiktoken
	id      string
}

// NewTiktokenAdapter creates a new adapter for a given tiktoken encoding ID.
func NewTiktokenAdapter(id string) (*TiktokenAdapter, error) {
	tkm, err := tiktoken.GetEncoding(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get tiktoken encoding %s: %v", id, err)
	}
	return &TiktokenAdapter{
		encoder: tkm,
		id:      id,
	}, nil
}

func (a *TiktokenAdapter) Encode(text string) []int {
	return a.encoder.Encode(text, nil, nil)
}

func (a *TiktokenAdapter) Decode(tokens []int) string {
	return a.encoder.Decode(tokens)
}

func (a *TiktokenAdapter) ID() string {
	return a.id
}

func (a *TiktokenAdapter) TokenSize() int {
	// cl100k_base and newer encodings use more than 65k tokens.
	if a.id == "cl100k_base" || a.id == "o200k_base" {
		return 4
	}
	return 2
}
