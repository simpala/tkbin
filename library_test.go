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
