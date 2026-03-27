package tkbin

import (
	"encoding/binary"
	"encoding/json"
	"os"
)

// Packer handles the creation of the library files
type Packer struct {
	Library   *Library
	ImageData []byte
}

func NewPacker(tokenizerID ...string) (*Packer, error) {
	id := "r50k_base"
	if len(tokenizerID) > 0 {
		id = tokenizerID[0]
	}

	adapter, err := NewTiktokenAdapter(id)
	if err != nil {
		return nil, err
	}

	return &Packer{
		Library: &Library{
			Index:     make(map[string]FileEntry),
			Tokenizer: adapter,
		},
	}, nil
}

// NewPackerWithTokenizer allows using a custom tokenizer implementation.
func NewPackerWithTokenizer(t Tokenizer) *Packer {
	return &Packer{
		Library: &Library{
			Index:     make(map[string]FileEntry),
			Tokenizer: t,
		},
	}
}

// AddFile tokenizes a file and appends it to the internal buffer
func (p *Packer) AddFile(name string, content []byte, metadata ...map[string]string) {
	ids := p.Library.Tokenizer.Encode(string(content))
	tokenSize := p.Library.Tokenizer.TokenSize()
	pixelSize := tokenSize * 4

	numPixels := (len(ids) + 3) / 4
	fileData := make([]byte, numPixels*pixelSize)

	for i := 0; i < len(ids); i++ {
		offset := i * tokenSize
		if tokenSize == 4 {
			binary.LittleEndian.PutUint32(fileData[offset:], uint32(ids[i]))
		} else {
			binary.LittleEndian.PutUint16(fileData[offset:], uint16(ids[i]))
		}
	}

	startPixel := len(p.ImageData) / pixelSize
	p.ImageData = append(p.ImageData, fileData...)

	var meta map[string]string
	if len(metadata) > 0 {
		meta = metadata[0]
	}

	p.Library.Index[name] = FileEntry{
		PixelStart:  startPixel,
		PixelLength: numPixels,
		TokenCount:  len(ids),
		Metadata:    meta,
	}
}

// Save writes the binary and JSON to disk
func (p *Packer) Save(binPath, jsonPath string) error {
	// Write Binary
	f, err := os.Create(binPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(p.ImageData)
	if err != nil {
		return err
	}

	// Write JSON
	libIndex := LibraryIndex{
		Tokenizer: p.Library.Tokenizer.ID(),
		Files:     p.Library.Index,
	}
	meta, err := json.MarshalIndent(libIndex, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(jsonPath, meta, 0644)
}
