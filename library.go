package tkbin

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"
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

type corpusStats struct {
	docFreq   map[int]int
	avgDocLen float64
	N         int
}

type Library struct {
	Index     map[string]FileEntry
	BinFile   *os.File
	Tokenizer Tokenizer
	stats     *corpusStats
	statsOnce sync.Once
}

type SearchResult struct {
	Key     string
	Snippet string
	Index   int
	Score   float64
}

type RetrieveOptions struct {
	MetaFilter   map[string]any
	Limit        int
	ContextChars int
	// Boost is an optional callback to provide a score bonus for a file.
	Boost func(key string, metadata map[string]string) float64
	// Optional BM25 parameters, defaults used if zero.
	K1 float64
	B  float64
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

	return &Library{Index: libIndex.Files, BinFile: f, Tokenizer: tkm}, nil
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

func (l *Library) ensureCorpusStats() {
	l.statsOnce.Do(func() {
		docFreq := make(map[int]int)
		totalTokens := 0
		N := len(l.Index)

		for key := range l.Index {
			tokens, err := l.GetTokens(key)
			if err != nil {
				continue
			}
			totalTokens += len(tokens)
			seen := make(map[int]bool)
			for _, t := range tokens {
				if !seen[t] {
					seen[t] = true
					docFreq[t]++
				}
			}
		}

		avgDocLen := 0.0
		if N > 0 {
			avgDocLen = float64(totalTokens) / float64(N)
		}

		l.stats = &corpusStats{
			docFreq:   docFreq,
			avgDocLen: avgDocLen,
			N:         N,
		}
	})
}

func (l *Library) bm25Score(queryTokens []int, docTokens []int, k1, b float64) float64 {
	if len(queryTokens) == 0 || len(docTokens) == 0 {
		return 0.0
	}

	l.ensureCorpusStats()
	score := 0.0

	tf := make(map[int]int)
	for _, t := range docTokens {
		tf[t]++
	}

	for _, term := range queryTokens {
		freq, exists := tf[term]
		if !exists || freq == 0 {
			continue
		}

		df := l.stats.docFreq[term]
		if df == 0 {
			df = 1
		}
		idf := math.Log(1 + (float64(l.stats.N-df)+0.5)/(float64(df)+0.5))

		numerator := float64(freq) * (k1 + 1)
		denominator := float64(freq) + k1*(1-b+b*float64(len(docTokens))/l.stats.avgDocLen)

		score += idf * (numerator / denominator)
	}
	return score
}

func matchesMetadata(meta map[string]string, filter map[string]any) bool {
	if filter == nil || len(filter) == 0 {
		return true
	}
	for k, v := range filter {
		if metaVal, ok := meta[k]; !ok || fmt.Sprintf("%v", v) != metaVal {
			return false
		}
	}
	return true
}

// Retrieve performs a BM25 search with metadata filtering and optional boosting.
func (l *Library) Retrieve(query string, opts RetrieveOptions) []SearchResult {
	if opts.Limit <= 0 {
		opts.Limit = 10
	}
	if opts.ContextChars == 0 {
		opts.ContextChars = 300
	}
	if opts.K1 == 0 {
		opts.K1 = 1.2
	}
	if opts.B == 0 {
		opts.B = 0.75
	}

	queryTokens := l.Tokenizer.Encode(query)
	if len(queryTokens) == 0 {
		return nil
	}

	var candidates []string
	for key, entry := range l.Index {
		if matchesMetadata(entry.Metadata, opts.MetaFilter) {
			candidates = append(candidates, key)
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	type scored struct {
		key   string
		score float64
	}
	scores := make([]scored, 0, len(candidates))

	for _, key := range candidates {
		tokens, err := l.GetTokens(key)
		if err != nil || len(tokens) == 0 {
			continue
		}

		score := l.bm25Score(queryTokens, tokens, opts.K1, opts.B)
		if opts.Boost != nil {
			score += opts.Boost(key, l.Index[key].Metadata)
		}

		if score <= 0 && len(opts.MetaFilter) == 0 {
			// If no terms match and no meta filter was used to narrow down, maybe skip?
			// But if meta filter WAS used, we might still want the result even with score 0.
			// Standard BM25 score is 0 if no terms match.
			if score == 0 {
				continue
			}
		}

		scores = append(scores, scored{key, score})
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].score == scores[j].score {
			return scores[i].key < scores[j].key
		}
		return scores[i].score > scores[j].score
	})

	limit := opts.Limit
	if len(scores) < limit {
		limit = len(scores)
	}

	results := make([]SearchResult, 0, limit)
	for i := 0; i < limit; i++ {
		content, _ := l.GetContent(scores[i].key)
		snippet, charIdx := l.generateSnippet(content, query, opts.ContextChars)

		results = append(results, SearchResult{
			Key:     scores[i].key,
			Snippet: snippet,
			Index:   charIdx,
			Score:   scores[i].score,
		})
	}

	return results
}

func (l *Library) generateSnippet(content string, query string, contextChars int) (string, int) {
	if contextChars < 0 {
		return strings.ReplaceAll(content, "\n", " "), 0
	}

	charIdx := strings.Index(strings.ToLower(content), strings.ToLower(query))
	if charIdx == -1 {
		// Fallback: first part of the file
		limit := contextChars * 2
		if len(content) < limit {
			limit = len(content)
		}
		return strings.ReplaceAll(content[:limit], "\n", " "), 0
	}

	start := max(0, charIdx-contextChars)
	end := min(len(content), charIdx+len(query)+contextChars)
	return strings.ReplaceAll(content[start:end], "\n", " "), charIdx
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
					snippet, _ := l.generateSnippet(content, query, contextChars)

					allResults = append(allResults, SearchResult{
						Key:     key,
						Snippet: snippet,
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
