package processor

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// ExtractText dispatches to the appropriate text extractor based on MIME type
func ExtractText(ctx context.Context, mimeType string, src io.Reader) (string, int, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return "", 0, fmt.Errorf("failed to read file data: %w", err)
	}

	switch mimeType {
	case "text/plain", "text/markdown":
		return extractPlainText(data)
	case "application/pdf":
		return extractPDF(data)
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/msword":
		return extractDOCX(data)
	default:
		// Best-effort: attempt plain text extraction
		return extractPlainText(data)
	}
}

// extractPlainText returns raw text, page count estimate, and word count from plain text files
func extractPlainText(data []byte) (string, int, error) {
	text := string(data)
	wordCount := countWords(text)
	// Estimate pages: ~500 words per page
	pageCount := max(1, wordCount/500)
	return text, pageCount, nil
}

// extractPDF extracts readable text from a PDF byte stream
// For the MVP we perform a best-effort byte-scan to pull ASCII/UTF-8 text
// between BT (Begin Text) and ET (End Text) markers in the PDF stream.
// NOTE: Add github.com/ledongthuc/pdf for production-grade extraction.
func extractPDF(data []byte) (string, int, error) {
	var sb strings.Builder
	content := string(data)

	pageCount := strings.Count(content, "/Page ")
	if pageCount == 0 {
		pageCount = 1
	}

	// Simple heuristic: extract printable text runs > 4 chars from PDF byte stream
	// For production quality, swap this with a proper PDF parser
	inText := false
	var curr strings.Builder

	for _, b := range data {
		r := rune(b)
		if unicode.IsPrint(r) && r != '(' && r != ')' {
			curr.WriteRune(r)
			inText = true
		} else {
			if inText && curr.Len() > 4 {
				word := strings.TrimSpace(curr.String())
				if len(word) > 0 {
					sb.WriteString(word)
					sb.WriteRune(' ')
				}
			}
			curr.Reset()
			inText = false
		}
	}

	extracted := strings.TrimSpace(sb.String())
	if extracted == "" {
		extracted = "[PDF text extraction produced no readable content - file may be scanned/image-based]"
	}

	return extracted, pageCount, nil
}

// extractDOCX extracts readable text from DOCX (ZIP+XML) byte stream
// Searches for w:t XML text nodes in the word/document.xml part.
func extractDOCX(data []byte) (string, int, error) {
	reader := bytes.NewReader(data)
	_ = reader // used if we add a real ZIP reader

	// Best-effort: scan for XML text content between <w:t> tags
	var sb strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		// Find all <w:t ...>content</w:t> patterns
		start := 0
		for {
			beginTag := strings.Index(line[start:], "<w:t")
			if beginTag < 0 {
				break
			}
			beginTag += start
			closeAngle := strings.Index(line[beginTag:], ">")
			if closeAngle < 0 {
				break
			}
			contentStart := beginTag + closeAngle + 1
			endTag := strings.Index(line[contentStart:], "</w:t>")
			if endTag < 0 {
				break
			}
			text := line[contentStart : contentStart+endTag]
			if len(strings.TrimSpace(text)) > 0 {
				sb.WriteString(text)
				sb.WriteRune(' ')
			}
			start = contentStart + endTag + 6
		}
	}

	extracted := strings.TrimSpace(sb.String())
	if extracted == "" {
		extracted = "[DOCX text extraction produced no readable content]"
	}

	wordCount := countWords(extracted)
	pageCount := max(1, wordCount/500)
	return extracted, pageCount, nil
}

// ChunkText splits a raw text body into overlapping chunks for vector search
func ChunkText(text string, chunkSize, overlap int) []string {
	if chunkSize <= 0 {
		chunkSize = 800
	}
	if overlap < 0 {
		overlap = 100
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	var chunks []string
	step := chunkSize - overlap
	if step <= 0 {
		step = chunkSize
	}

	for i := 0; i < len(words); i += step {
		end := i + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[i:end], " ")
		if strings.TrimSpace(chunk) != "" {
			chunks = append(chunks, chunk)
		}
		if end == len(words) {
			break
		}
	}

	return chunks
}

// EstimateTokenCount approximates token count using the ~4 chars per token heuristic
func EstimateTokenCount(text string) int {
	return max(1, len(text)/4)
}

func countWords(text string) int {
	count := 0
	inWord := false
	for _, r := range text {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			count++
		}
	}
	return count
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
