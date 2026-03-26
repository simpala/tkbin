package tkbin

import (
	"fmt"
	"os"
)

// ExampleNewPacker demonstrates the full initialization and 
// validation of a new Packer instance.
func ExampleNewPacker() {
	packer, err := NewPacker()
	if err != nil {
		return
	}

	fmt.Printf("Packer initialized: %v\n", packer != nil)
	fmt.Printf("Library initialized: %v\n", packer.Library != nil)
	fmt.Printf("Encoder initialized: %v\n", packer.Library.Encoder != nil)
	fmt.Printf("Initial ImageArray length: %d\n", len(packer.ImageArray))

	// Output:
	// Packer initialized: true
	// Library initialized: true
	// Encoder initialized: true
	// Initial ImageArray length: 0
}

// ExamplePacker_AddFile demonstrates adding multiple files with 
// extensive metadata and verifying the internal index.
func ExamplePacker_AddFile() {
	packer, _ := NewPacker()
	
	// Adding multiple files
	files := []string{"file1.txt", "file2.txt"}
	content := []byte("Test content")
	for _, name := range files {
		packer.AddFile(name, content, map[string]string{"type": "test"})
	}

	// Adding a file with extensive metadata
	meta := map[string]string{
		"author":   "John Doe",
		"category": "documents",
		"version":  "1.0",
	}
	packer.AddFile("metadata_test.txt", []byte("Content"), meta)

	// Verify entries
	entry := packer.Library.Index["metadata_test.txt"]
	fmt.Printf("Metadata Author: %s\n", entry.Metadata["author"])
	fmt.Printf("File1 exists: %v\n", packer.Library.Index["file1.txt"].PixelStart == 0)
	
	// Output:
	// Metadata Author: John Doe
	// File1 exists: true
}

// ExamplePacker_Save demonstrates a full round-trip: 
// Packing, Saving, and Re-opening.
func ExamplePacker_Save() {
	// Use temporary paths for the example
	binPath := "roundtrip.bin"
	jsonPath := "roundtrip.json"
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	packer, _ := NewPacker()
	packer.AddFile("roundtrip.txt", []byte("Round-trip content"), map[string]string{"test": "value"})
	
	packer.Save(binPath, jsonPath)

	// Open and verify
	lib, _ := Open(binPath, jsonPath)
	defer lib.BinFile.Close()

	entry, _ := lib.Index["roundtrip.txt"]
	fmt.Printf("Retrieved Metadata: %s\n", entry.Metadata["test"])

	// Output:
	// Retrieved Metadata: value
}
