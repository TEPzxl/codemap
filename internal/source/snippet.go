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

type CallsiteLocation struct {
	EdgeID        string
	File          string
	Line          int
	Column        int
	ContextBefore int
	ContextAfter  int
}

type Snippet struct {
	NodeID    string `json:"node_id"`
	File      string `json:"file"`
	StartLine int    `json:"start_line"`
	EndLine   int    `json:"end_line"`
	Source    string `json:"source"`
	Language  string `json:"language"`
}

type CallsiteSnippet struct {
	EdgeID        string `json:"edge_id"`
	File          string `json:"file"`
	Line          int    `json:"line"`
	Column        int    `json:"column"`
	StartLine     int    `json:"start_line"`
	EndLine       int    `json:"end_line"`
	Source        string `json:"source"`
	HighlightLine int    `json:"highlight_line"`
	Language      string `json:"language"`
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

	lines, err := readLines(root, location.File, location.StartLine, location.EndLine)
	if err != nil {
		return Snippet{}, err
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

func ReadCallsite(root string, location CallsiteLocation) (CallsiteSnippet, error) {
	if location.EdgeID == "" {
		return CallsiteSnippet{}, fmt.Errorf("edge id is required")
	}
	if location.File == "" {
		return CallsiteSnippet{}, fmt.Errorf("source file is required")
	}
	if location.Line <= 0 || location.Column <= 0 {
		return CallsiteSnippet{}, fmt.Errorf("invalid callsite position %d:%d", location.Line, location.Column)
	}
	if filepath.IsAbs(location.File) {
		return CallsiteSnippet{}, fmt.Errorf("absolute source paths are not allowed")
	}

	before := location.ContextBefore
	if before < 0 {
		before = 0
	}
	after := location.ContextAfter
	if after < 0 {
		after = 0
	}
	startLine := location.Line - before
	if startLine < 1 {
		startLine = 1
	}
	endLine := location.Line + after

	lines, err := readLines(root, location.File, startLine, endLine)
	if err != nil {
		return CallsiteSnippet{}, err
	}
	endLine = startLine + len(lines) - 1
	if location.Line < startLine || location.Line > endLine {
		return CallsiteSnippet{}, fmt.Errorf("source line range is outside file")
	}

	return CallsiteSnippet{
		EdgeID:        location.EdgeID,
		File:          filepath.ToSlash(location.File),
		Line:          location.Line,
		Column:        location.Column,
		StartLine:     startLine,
		EndLine:       endLine,
		Source:        strings.Join(lines, "\n"),
		HighlightLine: location.Line,
		Language:      "go",
	}, nil
}

func readLines(root string, filePath string, startLine int, endLine int) ([]string, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve root path: %w", err)
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("resolve project root symlinks: %w", err)
	}
	fileAbs := filepath.Clean(filepath.Join(rootAbs, filepath.FromSlash(filePath)))
	if !isWithinRoot(rootReal, fileAbs) {
		return nil, fmt.Errorf("source file escapes project root")
	}
	fileReal, err := filepath.EvalSymlinks(fileAbs)
	if err != nil {
		return nil, fmt.Errorf("resolve source file symlinks: %w", err)
	}
	if !isWithinRoot(rootReal, fileReal) {
		return nil, fmt.Errorf("source file escapes project root")
	}

	file, err := os.Open(fileReal)
	if err != nil {
		return nil, fmt.Errorf("open source file: %w", err)
	}
	defer file.Close()

	lines := make([]string, 0, endLine-startLine+1)
	scanner := bufio.NewScanner(file)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		if lineNo < startLine {
			continue
		}
		if lineNo > endLine {
			break
		}
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read source file: %w", err)
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("source line range is outside file")
	}

	return lines, nil
}

func isWithinRoot(rootAbs string, fileAbs string) bool {
	rootClean := filepath.Clean(rootAbs)
	if fileAbs == rootClean {
		return true
	}
	rootPrefix := rootClean + string(os.PathSeparator)
	return strings.HasPrefix(fileAbs, rootPrefix)
}
