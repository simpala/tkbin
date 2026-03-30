package main

import (
	"fmt"
	"log"
	"os"

	"github.com/simpala/tkbin"
)

func main() {
	binPath := "continuous.bin"
	jsonPath := "continuous.json"

	// Cleanup any old files from previous runs
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	// --- Session 1: Create a new library ---
	fmt.Println("--- Session 1: Creating initial library ---")
	packer, err := tkbin.NewPacker()
	if err != nil {
		log.Fatalf("Failed to create packer: %v", err)
	}

	packer.AddFile("file1.txt", []byte("Content for file 1: This is the first entry."), map[string]string{"version": "1"})
	packer.AddFile("file2.txt", []byte("Content for file 2: Hello from the beginning!"), map[string]string{"version": "1"})

	if err := packer.Save(binPath, jsonPath); err != nil {
		log.Fatalf("Failed to save initial library: %v", err)
	}
	fmt.Println("Initial library saved with 2 files.")

	// --- Session 2: Append in a new session using OpenPacker ---
	fmt.Println("\n--- Session 2: Appending to existing library ---")
	// Open the existing library for appending
	packer2, err := tkbin.OpenPacker(binPath, jsonPath)
	if err != nil {
		log.Fatalf("Failed to open packer for appending: %v", err)
	}

	packer2.AddFile("file3.txt", []byte("Content for file 3: This was appended later."), map[string]string{"version": "2"})

	if err := packer2.Save(binPath, jsonPath); err != nil {
		log.Fatalf("Failed to save appended library: %v", err)
	}
	fmt.Println("Library updated with 1 more file.")

	// --- Session 3: Direct append via Library.AddFile ---
	fmt.Println("\n--- Session 3: Direct append via Library instance ---")
	lib, err := tkbin.Open(binPath, jsonPath)
	if err != nil {
		log.Fatalf("Failed to open library: %v", err)
	}
	defer lib.Close()

	// AddFile on the Library instance handles the packer internally
	err = lib.AddFile("file4.txt", []byte("Content for file 4: Appended directly through the library instance."), map[string]string{"version": "3"})
	if err != nil {
		log.Fatalf("Library.AddFile failed: %v", err)
	}
	fmt.Println("Library updated directly via AddFile method.")

	// --- Verification: Read everything back ---
	fmt.Println("\n--- Verification: Reading all content ---")
	for i := 1; i <= 4; i++ {
		filename := fmt.Sprintf("file%d.txt", i)
		content, err := lib.GetContent(filename)
		if err != nil {
			log.Fatalf("Failed to get content for %s: %v", filename, err)
		}
		fmt.Printf("Read %s: %s (Metadata version: %s)\n",
			filename, content, lib.Index[filename].Metadata["version"])
	}
}
