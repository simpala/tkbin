package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"github.com/simpala/tkbin"
)

func main() {
	// 1. Open the library
	library, err := tkbin.Open("library.bin", "library.json")
	if err != nil {
		log.Fatalf("Could not open library: %v", err)
	}
	defer library.BinFile.Close()

	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "list":
		// List all files and their metadata
		fmt.Printf("%-30s | %-12s | %-10s\n", "PATH", "LANGUAGE", "SIZE")
		fmt.Println(strings.Repeat("-", 60))
		
		// Sort keys for consistent output
		keys := make([]string, 0, len(library.Index))
		for k := range library.Index {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			entry := library.Index[k]
			lang := entry.Metadata["language"]
			size := entry.Metadata["size_bytes"]
			fmt.Printf("%-30s | %-12s | %-10s bytes\n", k, lang, size)
		}

	case "view":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a filename to view")
		}
		fileName := os.Args[2]
		
		content, err := library.GetContent(fileName)
		if err != nil {
			log.Fatalf("Error retrieving file: %v", err)
		}

		fmt.Printf("--- Content of %s ---\n", fileName)
		fmt.Println(content)
		fmt.Println("---------------------------")

	case "info":
		if len(os.Args) < 3 {
			log.Fatal("Please provide a filename for info")
		}
		fileName := os.Args[2]
		entry, ok := library.Index[fileName]
		if !ok {
			log.Fatal("File not found in index")
		}

		fmt.Printf("File: %s\n", fileName)
		fmt.Printf("Binary Offset: %d pixels\n", entry.PixelStart)
		fmt.Printf("Token Count:   %d\n", entry.TokenCount)
		fmt.Println("Metadata:")
		for k, v := range entry.Metadata {
			fmt.Printf("  %s: %s\n", k, v)
		}

	default:
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run examples/viewer.go list          - List all files in library")
	fmt.Println("  go run examples/viewer.go view <path>   - Show file content")
	fmt.Println("  go run examples/viewer.go info <path>   - Show file metadata/stats")
}
