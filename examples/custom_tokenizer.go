package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/simpala/tkbin"
)

// MyCustomTokenizer is a simple example of a custom tokenizer implementation.
// In a real scenario, this would load a vocab from a JSON file or use an external library.
type MyCustomTokenizer struct{}

func (t *MyCustomTokenizer) Encode(text string) []int {
	// A very simple space-based tokenization for demonstration.
	words := strings.Fields(text)
	ids := make([]int, len(words))
	for i, word := range words {
		// Mock token IDs (e.g., hash the word)
		ids[i] = len(word)
	}
	return ids
}

func (t *MyCustomTokenizer) Decode(tokens []int) string {
	// Simple mock decoding
	res := make([]string, len(tokens))
	for i, id := range tokens {
		res[i] = strings.Repeat("a", id)
	}
	return strings.Join(res, " ")
}

func (t *MyCustomTokenizer) ID() string {
	return "my_custom_v1"
}

func (t *MyCustomTokenizer) TokenSize() int {
	// For small vocabularies, we use 2 bytes (uint16).
	// For > 65k tokens, we would return 4 (uint32).
	return 2
}

func main() {
	binPath, jsonPath := "custom_lib.bin", "custom_lib.json"
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	// 1. REGISTER THE CUSTOM TOKENIZER
	// This is important so that tkbin.Open() knows how to handle the "my_custom_v1" ID
	// when it reads it from the library's JSON index.
	custom := &MyCustomTokenizer{}
	tkbin.RegisterTokenizer(custom)

	// 2. PACK A LIBRARY WITH THE CUSTOM TOKENIZER
	packer := tkbin.NewPackerWithTokenizer(custom)
	packer.AddFile("hello.txt", []byte("this is a test"))

	err := packer.Save(binPath, jsonPath)
	if err != nil {
		log.Fatalf("Failed to save: %v", err)
	}
	fmt.Println("Packed library with custom tokenizer.")

	// 3. OPEN THE LIBRARY
	// tkbin.Open() will read the tokenizer ID from the JSON and
	// look it up in the registry.
	lib, err := tkbin.Open(binPath, jsonPath)
	if err != nil {
		log.Fatalf("Failed to open: %v", err)
	}
	defer lib.BinFile.Close()

	fmt.Printf("Library is using tokenizer: %s\n", lib.Tokenizer.ID())

	content, err := lib.GetContent("hello.txt")
	if err != nil {
		log.Fatalf("Failed to get content: %v", err)
	}
	fmt.Printf("Decoded content: %s\n", content)
}
