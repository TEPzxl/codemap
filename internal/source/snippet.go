package source

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Location struct {
	NodeID    string
	File      string
	StartLine int
	EndLine   int
}

type Snippet struct {
	NodeID    string `json:"node_id"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Source    string `json:"source"`
	Language  string `json:"language"`
}

func ReadSnippet(root string, location Location) (Snippet, error) {
	if location.NodeID == "" {
		return Snippet{}, fmt.Errorf("node id is required")
	}
	if location.File == "" {
		return Snippet{}, fmt.Errorf("source file is required")
	}
	if location.StartLine <= 0 || location.EndLine < location.StartLine {
		return Snippet{}, fmt.Errorf("invalid line range %d-%d", location.StartLine, location.EndLine)
	}
	if filepath.IsAbs(location.File) {
		return Snippet{}, fmt.Errorf("absolute source paths are not allowed")
	}

	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return Snippet{}, fmt.Errorf("resolve root path: %w", err)
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return Snippet{}, fmt.Errorf("resolve project root symlinks: %w", err)
	}
	fileAbs := filepath.Clean(filepath.Join(rootAbs, filepath.FromSlash(location.File)))
	if !isWithinRoot(rootReal, fileAbs) {
		return Snippet{}, fmt.Errorf("source file escapes project root")
	}
	fileReal, err := filepath.EvalSymlinks(fileAbs)
	if err != nil {
		return Snippet{}, fmt.Errorf("resolve source file symlinks: %w", err)
	}
	if !isWithinRoot(rootReal, fileReal) {
		return Snippet{}, fmt.Errorf("source file escapes project root")
	}

	file, err := os.Open(fileReal)
	if err != nil {
		return Snippet{}, fmt.Errorf("open source file: %w", err)
	}
	defer file.Close()

	lines := make([]string, 0, location.EndLine-location.StartLine+1)
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo < location.StartLine {
			continue
		}
		if lineNo > location.EndLine {
			break
		}
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return Snippet{}, fmt.Errorf("read source file: %w", err)
	}
	if len(lines) == 0 {
		return Snippet{}, fmt.Errorf("source line range is outside file")
	}

	return Snippet{
		NodeID:    location.NodeID,
		File:      filepath.ToSlash(location.File),
		StartLine: location.StartLine,
		EndLine:   location.EndLine,
		Source:    strings.Join(lines, "\n"),
		Language:  "go",
	}, nil
}

func isWithinRoot(rootAbs string, fileAbs string) bool {
	rootClean := filepath.Clean(rootAbs)
	if fileAbs == rootClean {
		return true
	}
	rootPrefix := rootClean + string(os.PathSeparator)
	return strings.HasPrefix(fileAbs, rootPrefix)
}
