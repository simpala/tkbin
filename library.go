package tkbin

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"bytes"

	"github.com/pkoukk/tiktoken-go"
)

type TokenPixel [4]uint16

type FileEntry struct {
	PixelStart  int               `json:"pixel_start"`
	PixelLength int               `json:"pixel_length"`
	TokenCount  int               `json:"token_count"`
	Metadata    map[string]string `json:"metadata,omitempty"` // New field
}

type Library struct {
	Index   map[string]FileEntry
	BinFile *os.File
	Encoder *tiktoken.Tiktoken
}

type SearchResult struct {
	Key     string
	Snippet string
	Index   int
}

// Open loads the metadata and opens the binary file for reading
func Open(binPath, jsonPath string) (*Library, error) {
	tkm, err := tiktoken.GetEncoding("r50k_base")
	if err != nil {
		return nil, err
	}

	metaBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var index map[string]FileEntry
	if err := json.Unmarshal(metaBytes, &index); err != nil {
		return nil, err
	}

	f, err := os.Open(binPath)
	if err != nil {
		return nil, err
	}

	return &Library{Index: index, BinFile: f, Encoder: tkm}, nil
}

//Close 
func (l *Library) Close() error {
    if l.BinFile != nil {
        return l.BinFile.Close()
    }
    return nil
}

// GetContent retrieves and decodes the full text for a given key using ReadAt
func (l *Library) GetContent(key string) (string, error) {
	entry, ok := l.Index[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}

	const pixelSize = 8
	offset := int64(entry.PixelStart) * pixelSize
	length := entry.PixelLength * pixelSize
	
	pixelData := make([]byte, length)
	
	// ReadAt is thread-safe and doesn't require a Seek call
	_, err := l.BinFile.ReadAt(pixelData, offset)
	if err != nil {
		return "", err
	}

	tokens := make([]int, 0, entry.TokenCount)
	for i := 0; i < len(pixelData); i += 2 {
		val := binary.LittleEndian.Uint16(pixelData[i : i+2])
		// We only append if the value is non-zero or within our token count
		// to respect the padding you added in Packer
		if len(tokens) < entry.TokenCount {
			tokens = append(tokens, int(val))
		}
	}

	return l.Encoder.Decode(tokens), nil
}

func (l *Library) Search(query string, contextChars int) []SearchResult {
	// 1. Define the variants to test
	variants := []string{
		query,           // "sample"
		" " + query,     // " sample"
		query + " ",     // "sample "
	}

	var allResults []SearchResult
	seenFiles := make(map[string]bool)

	for _, v := range variants {
		queryTokens := l.Encoder.Encode(v, nil, nil)
		if len(queryTokens) == 0 {
			continue
		}

		// Convert tokens to bytes for binary matching
		queryBytes := make([]byte, len(queryTokens)*2)
		for i, t := range queryTokens {
			binary.LittleEndian.PutUint16(queryBytes[i*2:], uint16(t))
		}

		for key, entry := range l.Index {
			// Skip if we already found a match for this file in a previous variant
			if seenFiles[key] {
				continue
			}

			pixelData := make([]byte, entry.PixelLength*8)
			_, err := l.BinFile.ReadAt(pixelData, int64(entry.PixelStart)*8)
			if err != nil {
				continue
			}

			if idx := bytes.Index(pixelData, queryBytes); idx != -1 {
				content, _ := l.GetContent(key)
				// Use case-insensitive string search for the snippet location
				charIdx := strings.Index(strings.ToLower(content), strings.ToLower(query))
				
				if charIdx != -1 {
					start := max(0, charIdx-contextChars)
					end := min(len(content), charIdx+len(query)+contextChars)
					allResults = append(allResults, SearchResult{
						Key:     key,
						Snippet: strings.ReplaceAll(content[start:end], "\n", " "),
						Index:   charIdx,
					})
					seenFiles[key] = true
				}
			}
		}
	}
	return allResults
}
func max(a, b int) int { if a > b { return a }; return b }
func min(a, b int) int { if a < b { return a }; return b }
