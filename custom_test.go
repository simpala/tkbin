package tkbin

import (
	"fmt"
	"os"
	"strings"
)

// SimpleTokenizer is a mock tokenizer for testing the CustomTokenizer interface.
type SimpleTokenizer struct {
	vocab map[string]int
	rev   map[int]string
}

func NewSimpleTokenizer() *SimpleTokenizer {
	v := map[string]int{"hello": 1, "world": 2, " ": 3}
	r := map[int]string{1: "hello", 2: "world", 3: " "}
	return &SimpleTokenizer{vocab: v, rev: r}
}

func (t *SimpleTokenizer) Encode(text string) []int {
	var ids []int
	words := strings.Split(text, "") // Very simple char-based for example
	// Actually let's just do a few words
	if text == "hello world" {
		return []int{1, 3, 2}
	}
	for _, char := range words {
		if id, ok := t.vocab[char]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func (t *SimpleTokenizer) Decode(tokens []int) string {
	var res []string
	for _, id := range tokens {
		if s, ok := t.rev[id]; ok {
			res = append(res, s)
		}
	}
	return strings.Join(res, "")
}

func (t *SimpleTokenizer) ID() string { return "simple_mock" }
func (t *SimpleTokenizer) TokenSize() int { return 2 }

func ExampleRegisterTokenizer() {
	binPath, jsonPath := "custom.bin", "custom.json"
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	// 1. Create and register custom tokenizer
	myTokenizer := NewSimpleTokenizer()
	RegisterTokenizer(myTokenizer)

	// 2. Pack with custom tokenizer
	packer := NewPackerWithTokenizer(myTokenizer)
	packer.AddFile("test.txt", []byte("hello world"))
	packer.Save(binPath, jsonPath)

	// 3. Open - should automatically use the registered custom tokenizer
	lib, _ := Open(binPath, jsonPath)
	defer lib.BinFile.Close()

	content, _ := lib.GetContent("test.txt")
	fmt.Printf("Tokenizer ID: %s\n", lib.Tokenizer.ID())
	fmt.Printf("Content: %s\n", content)

	// Output:
	// Tokenizer ID: simple_mock
	// Content: hello world
}

func ExampleNewPacker_cl100k() {
	binPath, jsonPath := "cl100k.bin", "cl100k.json"
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	// Use cl100k_base which uses uint32 (4 bytes per token)
	packer, _ := NewPacker("cl100k_base")
	packer.AddFile("large.txt", []byte("Testing cl100k_base tokenizer."))
	packer.Save(binPath, jsonPath)

	lib, _ := Open(binPath, jsonPath)
	defer lib.BinFile.Close()

	content, _ := lib.GetContent("large.txt")
	fmt.Printf("Tokenizer ID: %s\n", lib.Tokenizer.ID())
	fmt.Printf("Token Size: %d\n", lib.Tokenizer.TokenSize())
	fmt.Printf("Content: %s\n", content)

	// Output:
	// Tokenizer ID: cl100k_base
	// Token Size: 4
	// Content: Testing cl100k_base tokenizer.
}
