package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"github.com/simpala/tkbin"
)

func getLanguageName(ext string) string {
	switch ext {
	case ".go": return "Go"
	case ".py": return "Python"
	case ".c":  return "C"
	case ".h":  return "C Header"
	case ".cpp", ".cc", ".cxx": return "C++"
	case ".hpp": return "C++ Header"
	case ".rs":  return "Rust"
	case ".js":  return "JavaScript"
	case ".ts":  return "TypeScript"
	case ".md":  return "Markdown"
	case ".txt": return "Plain Text"
	case ".json": return "JSON"
	case ".yml", ".yaml": return "YAML"
	case ".sh":  return "Shell Script"
	default:     return "Unknown"
	}
}

func main() {
	packer, err := tkbin.NewPacker()
	if err != nil {
		log.Fatal(err)
	}

	dataRoot := "./mydata"

	// Define a map of supported code extensions
	// var supportedExts = map[string]bool{
	//     ".txt": true, ".md":  true, ".go":  true,
	// 	".py":  true, ".c":   true, ".h":   true,
	// 	".cpp": true, ".rs":  true, ".js":  true,
	// 	".ts":  true, ".json": true, ".yml": true,
	// }	
	
	filepath.Walk(dataRoot, func(path string, info os.FileInfo, err error) error {
	    if !info.IsDir() {
	        ext := filepath.Ext(path)
	        relPath, _ := filepath.Rel(dataRoot, path)
	        content, _ := os.ReadFile(path)
	
	        meta := map[string]string{
	            "extension":  ext,
	            "language":   getLanguageName(ext),
	            "size_bytes": fmt.Sprintf("%d", info.Size()),
	        }
	
	        // Add a "logic" vs "docs" classification
	        if ext == ".md" || ext == ".txt" {
	            meta["category"] = "documentation"
	        } else {
	            meta["category"] = "source_code"
	        }
	
	        packer.AddFile(relPath, content, meta)
	        log.Printf("Packed %s as %s", relPath, meta["language"])
	    }
	    return nil
	})
	err = packer.Save("library.bin", "library.json")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Library packed successfully!")

	// Load the library to verify
	library, err := tkbin.Open("library.bin", "library.json")
	if err != nil {
		log.Fatal(err)
	}

	// You can now access metadata directly from the index
	if entry, ok := library.Index["eino_docs/intro.txt"]; ok {
		fmt.Printf("\nMetadata for intro.txt: %v\n", entry.Metadata)
	}

	// Test Search (Corrected to search for 'sample' as per the log)
	log.Println("\nSearching for 'sample'...")
	results := library.Search("sample", 20)
	if len(results) > 0 {
		fmt.Println("Search Results:")
		for _, r := range results {
			fmt.Printf("File: %s, Index: %d, Snippet: %s\n", r.Key, r.Index, r.Snippet)
		}
	} else {
		log.Println("No results found.")
	}
}
