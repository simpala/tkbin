package main

import (
	"fmt"
	"log"
	"os"

	"github.com/simpala/tkbin"
)

func main() {
	// 1. Pack a sample library with metadata
	binPath, jsonPath := "retrieve_demo.bin", "retrieve_demo.json"
	packer, _ := tkbin.NewPacker()

	packer.AddFile("main.go", []byte(`package main
import "fmt"
func main() {
    fmt.Println("Hello, BM25!")
}`), map[string]string{
		"language": "Go",
		"category": "source_code",
	})

	packer.AddFile("script.py", []byte(`print("Hello, BM25 from Python!")`), map[string]string{
		"language": "Python",
		"category": "source_code",
	})

	packer.AddFile("README.md", []byte(`# BM25 Guide
This document explains how to use BM25 for RAG apps.
BM25 is better than simple keyword search because it ranks results by relevance.`), map[string]string{
		"category": "documentation",
	})

	if err := packer.Save(binPath, jsonPath); err != nil {
		log.Fatal(err)
	}
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	// 2. Open the library
	lib, err := tkbin.Open(binPath, jsonPath)
	if err != nil {
		log.Fatal(err)
	}
	defer lib.Close()

	query := "how to use BM25"

	// 3. Simple BM25 Retrieve
	fmt.Printf("--- Simple BM25 for '%s' ---\n", query)
	results := lib.Retrieve(query, tkbin.RetrieveOptions{
		Limit:        3,
		ContextChars: 50,
	})
	for i, r := range results {
		fmt.Printf("%d. %s (score: %.3f)\n   Snippet: %s\n", i+1, r.Key, r.Score, r.Snippet)
	}

	// 4. Metadata-Filtered Retrieve
	fmt.Printf("\n--- Filtered (category=source_code) ---\n")
	resultsFiltered := lib.Retrieve("Hello", tkbin.RetrieveOptions{
		MetaFilter: map[string]any{"category": "source_code"},
		Limit:      3,
	})
	for i, r := range resultsFiltered {
		fmt.Printf("%d. %s (score: %.3f)\n", i+1, r.Key, r.Score)
	}

	// 5. Retrieve with Custom Metadata Boosting
	fmt.Printf("\n--- Boosted (Python gets +2.0) ---\n")
	resultsBoosted := lib.Retrieve("Hello", tkbin.RetrieveOptions{
		Limit: 3,
		Boost: func(key string, metadata map[string]string) float64 {
			if metadata["language"] == "Python" {
				return 2.0
			}
			return 0.0
		},
	})
	for i, r := range resultsBoosted {
		lang := lib.Index[r.Key].Metadata["language"]
		fmt.Printf("%d. %s [%s] (score: %.3f)\n", i+1, r.Key, lang, r.Score)
	}

	// 6. Full File Retrieval
	fmt.Printf("\n--- Full File Retrieval ---\n")
	resultsFull := lib.Retrieve("script.py", tkbin.RetrieveOptions{
		Limit:        1,
		ContextChars: -1, // Use -1 for full file content
	})
	if len(resultsFull) > 0 {
		fmt.Printf("File: %s\nContent: %s\n", resultsFull[0].Key, resultsFull[0].Snippet)
	} else {
		fmt.Println("No results found for full file retrieval")
	}
}
