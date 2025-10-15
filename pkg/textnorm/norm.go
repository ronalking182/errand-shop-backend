package textnorm

import (
	"regexp"
	"strings"
)

// Normalize applies text normalization rules:
// - lowercase
// - strip punctuation
// - collapse spaces
// - expand aliases
func Normalize(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	
	// Strip punctuation (keep only letters, numbers, and spaces)
	punctuationRegex := regexp.MustCompile(`[^a-z0-9\s]`)
	text = punctuationRegex.ReplaceAllString(text, " ")
	
	// Collapse multiple spaces into single spaces
	spaceRegex := regexp.MustCompile(`\s+`)
	text = spaceRegex.ReplaceAllString(text, " ")
	
	// Expand aliases
	aliases := map[string]string{
		" ii ": " 2 ",
		" iii ": " 3 ",
		" rd ": " road ",
		" ave ": " avenue ",
		" f c t ": " fct ",
	}
	
	// Add spaces around text to ensure boundary matching
	text = " " + strings.TrimSpace(text) + " "
	
	// Apply alias replacements
	for alias, replacement := range aliases {
		text = strings.ReplaceAll(text, alias, replacement)
	}
	
	// Clean up and return
	return strings.TrimSpace(text)
}

// NormalizeKeyword normalizes a single keyword for matching
func NormalizeKeyword(keyword string) string {
	return Normalize(keyword)
}