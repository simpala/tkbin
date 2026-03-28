package main

import (
	"fmt"
	"os"
	"strings"
	"github.com/simpala/tkbin"
)

// SimpleCharTokenizer - inline version for standalone demo
// Uses pre-computed mappings for O(1) lookup performance
type SimpleCharTokenizer struct {
	encoder map[rune]int
	decoder map[int]rune
}

func NewSimpleCharTokenizer() *SimpleCharTokenizer {
	charSet := []rune{' ', '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', ':', ';', '<', '=', '>', '?', '@',
		'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z',
		'[', '\\', ']', '^', '_', '`',
		'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z',
		'{', '|', '}', '~'}

	encoder := make(map[rune]int)
	decoder := make(map[int]rune)
	for i, char := range charSet {
		encoder[char] = i
		decoder[i] = char
	}

	return &SimpleCharTokenizer{
		encoder: encoder,
		decoder: decoder,
	}
}

func (s *SimpleCharTokenizer) ID() string {
	return "simple_char_vocab"
}

func (s *SimpleCharTokenizer) TokenSize() int {
	return 2
}

func (s *SimpleCharTokenizer) Encode(text string) []int {
	if text == "" {
		return []int{}
	}

	tokenIDs := make([]int, 0, len(text))
	for _, char := range text {
		if id, exists := s.encoder[char]; exists {
			tokenIDs = append(tokenIDs, id)
		} else {
			tokenIDs = append(tokenIDs, 127)
		}
	}

	return tokenIDs
}

func (s *SimpleCharTokenizer) Decode(tokens []int) string {
	if len(tokens) == 0 {
		return ""
	}

	var result strings.Builder
	for _, tokenID := range tokens {
		if char, exists := s.decoder[tokenID]; exists {
			result.WriteRune(char)
		}
	}

	return result.String()
}

func main() {
	fmt.Println("=== TKBIN Custom Vocab Integration Demo (Optimized) ===")
	fmt.Println()

	// Create tokenizer with pre-computed mappings
	customTk := NewSimpleCharTokenizer()

	// Register the tokenizer with tkbin
	tkbin.RegisterTokenizer(customTk)

	fmt.Printf("✓ Registered tokenizer: %s\n", customTk.ID())
	fmt.Printf("✓ Token size: %d bytes\n", customTk.TokenSize())
	fmt.Println()

	// Test the tokenizer
	testText := "Hello, World!"
	tokenIDs := customTk.Encode(testText)
	decodedText := customTk.Decode(tokenIDs)

	fmt.Printf("Original: %s\n", testText)
	fmt.Printf("Tokens:   %v\n", tokenIDs)
	fmt.Printf("Decoded:  %s\n", decodedText)
	fmt.Printf("✓ Round-trip successful: %v\n\n", testText == decodedText)

	// Create a Packer with our custom tokenizer
	fmt.Println("Creating Packer with custom tokenizer...")
	packer := tkbin.NewPackerWithTokenizer(customTk)
	if packer == nil {
		fmt.Println("✗ Failed to create packer")
		os.Exit(1)
	}
	fmt.Println("✓ Packer created successfully")

	// Add a sample file
	fmt.Println("\nAdding sample files...")
	sample1 := "This is a sample text file for tkbin."
	sample2 := "Another sample with numbers 12345 and symbols!"

	packer.AddFile("sample1.txt", []byte(sample1))
	packer.AddFile("sample2.txt", []byte(sample2))
	
	// Add some binary data
	binaryData := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE}
	packer.AddFile("binary.dat", binaryData)
	
	fmt.Println("✓ Files added successfully")

	// Save the library
	fmt.Println("\nSaving library to binary files...")
	binPath := "tkbin_custom_vocab_optimized.bin"
	jsonPath := "tkbin_custom_vocab_optimized.json"
	
	err := packer.Save(binPath, jsonPath)
	if err != nil {
		fmt.Printf("✗ Error saving packer: %v\n", err)
		os.Exit(1)
	}
	
	fmt.Printf("✓ Library saved to:\n")
	fmt.Printf("  - Binary: %s\n", binPath)
	fmt.Printf("  - JSON: %s\n", jsonPath)
	fmt.Println()

	// Load and verify the library
	fmt.Println("Loading and verifying the library...")
	library, err := tkbin.Open(binPath, jsonPath)
	if err != nil {
		fmt.Printf("✗ Error loading library: %v\n", err)
		os.Exit(1)
	}
	defer library.Close()
	
	fmt.Printf("✓ Library loaded successfully\n")
	fmt.Printf("  Files in library: %d\n", len(library.Index))
	fmt.Println()

	// Print file details
	fmt.Println("File details:")
	for key, file := range library.Index {
		fmt.Printf("  - %s\n", key)
		fmt.Printf("    Token count: %d\n", file.TokenCount)
		fmt.Printf("    Pixel length: %d\n", file.PixelLength)
		if file.Metadata != nil {
			fmt.Printf("    Metadata: %v\n", file.Metadata)
		}
	}
	fmt.Println()

	// Demonstrate tokenization of file content using GetContent
	fmt.Println("Tokenizing sample content:")
	for key := range library.Index {
		content, err := library.GetContent(key)
		if err != nil {
			fmt.Printf("  %s: Error getting content - %v\n", key, err)
			continue
		}
		
		tokens := customTk.Encode(content)
		decoded := customTk.Decode(tokens)
		
		fmt.Printf("  %s:\n", key)
		fmt.Printf("    Original: %s\n", content)
		fmt.Printf("    Tokens:   %v\n", tokens)
		fmt.Printf("    Decoded:  %s\n", decoded)
		fmt.Printf("    Match:    %v\n\n", content == decoded)
	}

	fmt.Println("=== Demo Complete ===")
	fmt.Println("The custom vocab is now integrated with tkbin!")
}