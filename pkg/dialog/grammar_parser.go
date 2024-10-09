package dialog

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type grammar struct {
	Cloze         string
	SentenceFront string
	SentenceBack  string
	Pattern       string
	Note          string
	Structure     string
	Examples      string
	Summary       []string
}

func loadGrammar(path string) (*grammar, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("error reading grammar file: %w", err)
	}

	splits := strings.Split(string(content), "---\n")
	if len(splits) != 5 {
		return nil, errors.New("Invalid file format")
	}

	cloze := strings.TrimSpace(splits[0])
	withoutParenthesis, withUnderscores, pattern, err := processPattern(cloze)
	if err != nil {
		return nil, err
	}

	summary := make([]string, 0)
	if strings.Contains(splits[4], "\n") {
		points := strings.Split(splits[4], "\n")
		for _, p := range points {
			summary = append(summary, strings.TrimSpace(p))
		}
	} else {
		summary = append(summary, splits[4])
	}

	return &grammar{
		Cloze:         cloze,
		SentenceFront: withUnderscores,
		SentenceBack:  withoutParenthesis,
		Pattern:       pattern,
		Note:          strings.TrimSpace(splits[1]),
		Structure:     strings.TrimSpace(splits[2]),
		Examples:      strings.ReplaceAll(splits[3], " ", ""),
		Summary:       summary,
	}, nil

}

// processPattern processes a Chinese sentence, extracts words in parentheses,
// and returns three variables based on the criteria.
func processPattern(cloze string) (string, string, string, error) {
	// Regular expression to match content inside parentheses
	re := regexp.MustCompile(`\(([^)]+)\)`)

	// Find all matches of the words in parentheses
	matches := re.FindAllStringSubmatch(cloze, -1)

	// If no matches found, return an error
	if len(matches) == 0 {
		return "", "", "", fmt.Errorf("no words found in parentheses")
	}

	// Remove parentheses and create the first output pattern
	patternWithoutParentheses := re.ReplaceAllString(cloze, "$1")

	// Replace parentheses content with underscores for the second output
	patternWithUnderscores := re.ReplaceAllString(cloze, "___")

	// Create the third output with extracted words joined by ellipsis
	var extractedWords []string
	for _, match := range matches {
		extractedWords = append(extractedWords, match[1])
	}
	wordsWithEllipsis := strings.Join(extractedWords, "...")

	return patternWithoutParentheses, patternWithUnderscores, wordsWithEllipsis, nil
}
