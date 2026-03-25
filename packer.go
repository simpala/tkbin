package tkbin

import (
	"encoding/binary"
	"encoding/json"
	"os"

	"github.com/pkoukk/tiktoken-go"
)

// Packer handles the creation of the library files
type Packer struct {
	Library    *Library
	ImageArray []TokenPixel
}

func NewPacker() (*Packer, error) {
	tkm, err := tiktoken.GetEncoding("r50k_base")
	if err != nil {
		return nil, err
	}
	return &Packer{
		Library: &Library{
			Index:   make(map[string]FileEntry),
			Encoder: tkm,
		},
	}, nil
}

// AddFile tokenizes a file and appends it to the internal buffer
// The "..." makes metadata optional
// AddFile tokenizes a file and appends it to the internal buffer
func (p *Packer) AddFile(name string, content []byte, metadata ...map[string]string) {
	ids := p.Library.Encoder.Encode(string(content), nil, nil)
	
	var filePixels []TokenPixel
	for i := 0; i < len(ids); i += 4 {
		var pix TokenPixel
		for j := 0; j < 4; j++ {
			if i+j < len(ids) {
				pix[j] = uint16(ids[i+j])
			} else {
				pix[j] = 0 // Padding
			}
		}
		filePixels = append(filePixels, pix)
	}

	startIdx := len(p.ImageArray)
	p.ImageArray = append(p.ImageArray, filePixels...)

	// Extract the map from the variadic slice
	var meta map[string]string
	if len(metadata) > 0 {
		meta = metadata[0]
	}

	p.Library.Index[name] = FileEntry{
		PixelStart:  startIdx,
		PixelLength: len(filePixels),
		TokenCount:  len(ids),
		Metadata:    meta, // Correctly use the extracted map
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
	
	err = binary.Write(f, binary.LittleEndian, p.ImageArray)
	if err != nil {
		return err
	}

	// Write JSON
	meta, err := json.MarshalIndent(p.Library.Index, "", "  ")
	if err != nil {
		return err
	}
	
	return os.WriteFile(jsonPath, meta, 0644)
}
