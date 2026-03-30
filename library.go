package tkbin

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// TokenPixel is kept for logical structure but we handle it via byte offsets
// type TokenPixel [4]uint16 (Legacy)

type FileEntry struct {
	PixelStart  int               `json:"pixel_start"`
	PixelLength int               `json:"pixel_length"`
	TokenCount  int               `json:"token_count"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type LibraryIndex struct {
	Tokenizer string               `json:"tokenizer"`
	Files     map[string]FileEntry `json:"files"`
}

type Library struct {
	Index     map[string]FileEntry
	BinFile   *os.File
	Tokenizer Tokenizer
	binPath   string
	jsonPath  string
}

type SearchResult struct {
	Key     string
	Snippet string
	Index   int
}

// Open loads the metadata and opens the binary file for reading
func Open(binPath, jsonPath string) (*Library, error) {
	metaBytes, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, err
	}

	var libIndex LibraryIndex
	if err := json.Unmarshal(metaBytes, &libIndex); err != nil {
		// Attempt to parse legacy format where the top level was the map
		var legacyIndex map[string]FileEntry
		if errLegacy := json.Unmarshal(metaBytes, &legacyIndex); errLegacy == nil {
			libIndex = LibraryIndex{
				Tokenizer: "r50k_base",
				Files:     legacyIndex,
			}
		} else {
			return nil, err
		}
	}

	tkm, err := getTokenizer(libIndex.Tokenizer)
	if err != nil {
		return nil, err
	}

	f, err := os.Open(binPath)
	if err != nil {
		return nil, err
	}

	return &Library{
		Index:     libIndex.Files,
		BinFile:   f,
		Tokenizer: tkm,
		binPath:   binPath,
		jsonPath:  jsonPath,
	}, nil
}

//Close 
func (l *Library) Close() error {
    if l.BinFile != nil {
        return l.BinFile.Close()
    }
    return nil
}

// AddFile appends a new file to the existing library on disk and updates the current library instance.
func (l *Library) AddFile(name string, content []byte, metadata ...map[string]string) error {
	if l.binPath == "" || l.jsonPath == "" {
		return fmt.Errorf("library paths not set; cannot append")
	}

	packer, err := OpenPacker(l.binPath, l.jsonPath)
	if err != nil {
		return err
	}
	defer packer.Library.Close()

	packer.AddFile(name, content, metadata...)
	err = packer.Save(l.binPath, l.jsonPath)
	if err != nil {
		return err
	}

	// Update the local library's index
	l.Index[name] = packer.Library.Index[name]
	return nil
}

// GetContent retrieves and decodes the full text for a given key using ReadAt
func (l *Library) GetContent(key string) (string, error) {
	entry, ok := l.Index[key]
	if !ok {
		return "", fmt.Errorf("key not found: %s", key)
	}

	tokenSize := l.Tokenizer.TokenSize()
	pixelSize := int64(tokenSize * 4)
	offset := int64(entry.PixelStart) * pixelSize
	length := int64(entry.PixelLength) * pixelSize

	pixelData := make([]byte, length)

	// ReadAt is thread-safe and doesn't require a Seek call
	_, err := l.BinFile.ReadAt(pixelData, offset)
	if err != nil {
		return "", err
	}

	tokens := make([]int, 0, entry.TokenCount)
	for i := 0; i < len(pixelData); i += tokenSize {
		var val int
		if tokenSize == 4 {
			val = int(binary.LittleEndian.Uint32(pixelData[i : i+4]))
		} else {
			val = int(binary.LittleEndian.Uint16(pixelData[i : i+2]))
		}

		if len(tokens) < entry.TokenCount {
			tokens = append(tokens, val)
		}
	}

	return l.Tokenizer.Decode(tokens), nil
}

// GetTokens retrieves the token IDs for a given key directly from the binary file.
func (l *Library) GetTokens(key string) ([]int, error) {
	entry, ok := l.Index[key]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", key)
	}

	tokenSize := l.Tokenizer.TokenSize()
	pixelSize := int64(tokenSize * 4)
	offset := int64(entry.PixelStart) * pixelSize
	length := int64(entry.PixelLength) * pixelSize

	pixelData := make([]byte, length)

	_, err := l.BinFile.ReadAt(pixelData, offset)
	if err != nil {
		return nil, err
	}

	tokens := make([]int, 0, entry.TokenCount)
	for i := 0; i < len(pixelData); i += tokenSize {
		var val int
		if tokenSize == 4 {
			val = int(binary.LittleEndian.Uint32(pixelData[i : i+4]))
		} else {
			val = int(binary.LittleEndian.Uint16(pixelData[i : i+2]))
		}

		if len(tokens) < entry.TokenCount {
			tokens = append(tokens, val)
		}
	}
	return tokens, nil
}

func (l *Library) Search(query string, contextChars int) []SearchResult {
	variants := []string{
		query,       // "sample"
		" " + query, // " sample"
		query + " ", // "sample "
	}

	var allResults []SearchResult
	seenFiles := make(map[string]bool)
	tokenSize := l.Tokenizer.TokenSize()
	pixelSize := int64(tokenSize * 4)

	for _, v := range variants {
		queryTokens := l.Tokenizer.Encode(v)
		if len(queryTokens) == 0 {
			continue
		}

		// Convert tokens to bytes for binary matching
		queryBytes := make([]byte, len(queryTokens)*tokenSize)
		for i, t := range queryTokens {
			if tokenSize == 4 {
				binary.LittleEndian.PutUint32(queryBytes[i*4:], uint32(t))
			} else {
				binary.LittleEndian.PutUint16(queryBytes[i*2:], uint16(t))
			}
		}

		for key, entry := range l.Index {
			if seenFiles[key] {
				continue
			}

			pixelData := make([]byte, int64(entry.PixelLength)*pixelSize)
			_, err := l.BinFile.ReadAt(pixelData, int64(entry.PixelStart)*pixelSize)
			if err != nil {
				continue
			}

			if idx := bytes.Index(pixelData, queryBytes); idx != -1 {
				content, _ := l.GetContent(key)
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
