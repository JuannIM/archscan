package analyzer

import (
	"io"
	"os"
	"strings"
)

// shortenPath strips the repo root prefix from a path for cleaner display.
func shortenPath(root, path string) string {
	rel := strings.TrimPrefix(path, root)
	rel = strings.TrimPrefix(rel, "/")
	if rel == "" {
		return path
	}
	return rel
}

// readFileSafe reads up to maxBytes from a file.
func readFileSafe(path string, maxBytes int) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := make([]byte, maxBytes)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return buf[:n], nil
}
