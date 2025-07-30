package docs

import (
	"embed"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
)

//go:embed *.md
//go:embed getting-started/*.md
//go:embed guides/*.md
//go:embed reference/*.md
var content embed.FS

// GetDoc returns the content of a documentation file
func GetDoc(path string) (string, error) {
	// Clean the path
	path = strings.TrimPrefix(path, "/")
	
	// Try different variations
	variations := []string{
		path,
		strings.TrimSuffix(path, ".md") + ".md",
	}
	
	// Special case for index.md
	if path == "index.md" || path == "index" {
		variations = []string{"index.md"}
	}
	
	// Try each variation
	for _, v := range variations {
		data, err := content.ReadFile(v)
		if err == nil {
			return string(data), nil
		}
	}
	
	// List available files for debugging
	available, _ := ListDocs()
	return "", fmt.Errorf("documentation not found: %s (available: %v)", path, available)
}

// ListDocs returns all available documentation files
func ListDocs() ([]string, error) {
	var docs []string
	
	err := fs.WalkDir(content, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			docs = append(docs, path)
		}
		
		return nil
	})
	
	return docs, err
}