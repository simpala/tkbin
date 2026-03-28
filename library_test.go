package tkbin

import (
	"fmt"
	"os"
)

// ExampleLibrary_GetContent demonstrates retrieving decoded 
// text from a saved binary library.
func ExampleLibrary_GetContent() {
	binPath, jsonPath := "get_content.bin", "get_content.json"
	packer, _ := NewPacker()
	packer.AddFile("document.txt", []byte("This is test content."))
	packer.Save(binPath, jsonPath)
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	lib, _ := Open(binPath, jsonPath)
	defer lib.BinFile.Close()

	content, _ := lib.GetContent("document.txt")
	fmt.Println(content)

	// Output:
	// This is test content.
}

// ExampleLibrary_Search demonstrates the search functionality 
// including snippets and key matching.
func ExampleLibrary_Search() {
	binPath, jsonPath := "search.bin", "search.json"
	packer, _ := NewPacker()
	content := []byte("The quick brown fox jumps over the lazy dog.")
	packer.AddFile("fox.txt", content)
	packer.Save(binPath, jsonPath)
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	lib, _ := Open(binPath, jsonPath)
	defer lib.BinFile.Close()

	// Using a context of 5 chars around the match
	results := lib.Search("brown fox", 5) 
	if len(results) > 0 {
		fmt.Printf("Found in: %s\n", results[0].Key)
		fmt.Printf("Snippet: %s\n", results[0].Snippet)
	}

	// Output:
	// Found in: fox.txt
	// Snippet: uick brown fox jump
}

// ExampleLibrary_Retrieve demonstrates the BM25 search functionality
// with metadata filtering and custom boosting.
func ExampleLibrary_Retrieve() {
	binPath, jsonPath := "retrieve.bin", "retrieve.json"
	packer, _ := NewPacker()

	packer.AddFile("go_code.go", []byte("package main\nfunc main() { fmt.Println(\"hello\") }"), map[string]string{
		"language": "Go",
		"category": "source_code",
	})

	packer.AddFile("python_code.py", []byte("print(\"hello\")"), map[string]string{
		"language": "Python",
	})

	packer.AddFile("readme.txt", []byte("This is a hello world guide."), map[string]string{
		"category": "documentation",
	})

	packer.Save(binPath, jsonPath)
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	lib, _ := Open(binPath, jsonPath)
	defer lib.BinFile.Close()

	// Search for "hello" in Go source code
	opts := RetrieveOptions{
		MetaFilter: map[string]any{
			"language": "Go",
		},
		Limit:        5,
		ContextChars: 10,
	}

	results := lib.Retrieve("hello", opts)
	for _, r := range results {
		if r.Score > 0 {
			fmt.Printf("Key: %s, Found\n", r.Key)
		}
	}

	// Search for "hello" with a boost for documentation
	optsBoost := RetrieveOptions{
		Limit: 5,
		Boost: func(key string, metadata map[string]string) float64 {
			if metadata["category"] == "documentation" {
				return 1.0
			}
			return 0.0
		},
	}

	resultsBoost := lib.Retrieve("hello", optsBoost)
	fmt.Println("Top result with boost:", resultsBoost[0].Key)

	// Output:
	// Key: go_code.go, Found
	// Top result with boost: readme.txt
}
