package tkbin

import (
	"sync"
)

// Tokenizer defines the interface for different tokenization strategies.
type Tokenizer interface {
	// Encode converts text into a slice of token IDs.
	Encode(text string) []int
	// Decode converts a slice of token IDs back into text.
	Decode(tokens []int) string
	// ID returns the identifier for this tokenizer (e.g., "r50k_base").
	ID() string
	// TokenSize returns the number of bytes used to store each token (2 or 4).
	TokenSize() int
}

var (
	registryMu sync.RWMutex
	registry   = make(map[string]Tokenizer)
)

// RegisterTokenizer adds a custom tokenizer to the registry so it can be
// automatically used by Open() when encountered in a library index.
func RegisterTokenizer(t Tokenizer) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[t.ID()] = t
}

func getTokenizer(id string) (Tokenizer, error) {
	registryMu.RLock()
	t, ok := registry[id]
	registryMu.RUnlock()
	if ok {
		return t, nil
	}

	// Fallback to built-in tiktoken
	return NewTiktokenAdapter(id)
}
