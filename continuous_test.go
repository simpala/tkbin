package tkbin

import (
	"os"
	"testing"
)

func TestContinuousPacking(t *testing.T) {
	binPath := "continuous.bin"
	jsonPath := "continuous.json"
	defer os.Remove(binPath)
	defer os.Remove(jsonPath)

	// 1. Initial Pack
	packer, err := NewPacker()
	if err != nil {
		t.Fatalf("Failed to create packer: %v", err)
	}

	packer.AddFile("file1.txt", []byte("Hello world from file 1"))
	err = packer.Save(binPath, jsonPath)
	if err != nil {
		t.Fatalf("Failed to save initial: %v", err)
	}

	// 2. Open and verify
	lib, err := Open(binPath, jsonPath)
	if err != nil {
		t.Fatalf("Failed to open library: %v", err)
	}
	defer lib.Close()

	content, err := lib.GetContent("file1.txt")
	if err != nil {
		t.Fatalf("Failed to get content from file1: %v", err)
	}
	if content != "Hello world from file 1" {
		t.Errorf("Unexpected content: %s", content)
	}

	// 3. Continuous append (same session)
	packer.AddFile("file2.txt", []byte("More content in file 2"))
	err = packer.Save(binPath, jsonPath)
	if err != nil {
		t.Fatalf("Failed to save append in same session: %v", err)
	}

	// Re-open to verify
	lib2, err := Open(binPath, jsonPath)
	if err != nil {
		t.Fatalf("Failed to re-open library: %v", err)
	}
	defer lib2.Close()

	if _, ok := lib2.Index["file1.txt"]; !ok {
		t.Error("file1.txt missing after same-session append")
	}
	content2, _ := lib2.GetContent("file2.txt")
	if content2 != "More content in file 2" {
		t.Errorf("Unexpected content for file2: %s", content2)
	}

	// 4. Library.AddFile (new session/direct method)
	err = lib2.AddFile("file3.txt", []byte("Final content from file 3"))
	if err != nil {
		t.Fatalf("Library.AddFile failed: %v", err)
	}

	// Verify locally updated index
	if _, ok := lib2.Index["file3.txt"]; !ok {
		t.Error("file3.txt missing from lib2 index after AddFile")
	}

	// Verify we can read the new content from the same library instance
	content3_lib2, err := lib2.GetContent("file3.txt")
	if err != nil {
		t.Errorf("Failed to get file3 content from lib2: %v", err)
	}
	if content3_lib2 != "Final content from file 3" {
		t.Errorf("Unexpected content for file3 from lib2: %s", content3_lib2)
	}

	// Re-re-open to verify disk persistence
	lib3, err := Open(binPath, jsonPath)
	if err != nil {
		t.Fatalf("Failed to re-re-open library: %v", err)
	}
	defer lib3.Close()

	if _, ok := lib3.Index["file1.txt"]; !ok {
		t.Error("file1.txt missing after lib.AddFile append")
	}
	if _, ok := lib3.Index["file2.txt"]; !ok {
		t.Error("file2.txt missing after lib.AddFile append")
	}
	content3, _ := lib3.GetContent("file3.txt")
	if content3 != "Final content from file 3" {
		t.Errorf("Unexpected content for file3: %s", content3)
	}
}
