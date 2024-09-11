package dialog

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type cloze struct {
	withUnderscores    string
	withoutParenthesis string
	word               string
	grammar            string
	note               string
}

func loadClozes(path string) ([]cloze, error) {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open sentences file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var clozes []cloze
	var s string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s = scanner.Text()
		if s == "" {
			continue
		}

		// note
		if strings.Contains(s, "|") {
			parts := strings.Split(s, "|")
			if len(parts) > 2 {
				fmt.Printf("could not split cloze, too many '|' : %s\n", s)
				continue
			}

			withoutParenthesis, withUnderscores, word, err := processSentence(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, err
			}

			clozes = append(clozes, cloze{
				withUnderscores:    withUnderscores,
				withoutParenthesis: withoutParenthesis,
				word:               word,
				note:               parts[1],
			})
			continue
		}

		// grammar
		if strings.Contains(s, ">") || strings.Contains(s, "》") {
			parts := strings.Split(s, ">")
			if len(parts) > 2 {
				fmt.Printf("could not split cloze, too many '>' : %s\n", s)
				continue
			}
			withoutParenthesis, withUnderscores, word, err := processSentence(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, err
			}

			clozes = append(clozes, cloze{
				withUnderscores:    withUnderscores,
				withoutParenthesis: withoutParenthesis,
				word:               word,
				note:               parts[1],
			})
			continue
		}
		if strings.Contains(s, "》") {
			parts := strings.Split(s, "》")
			if len(parts) > 2 {
				fmt.Printf("could not split cloze, too many '》' : %s\n", s)
				continue
			}
			withoutParenthesis, withUnderscores, word, err := processSentence(strings.TrimSpace(parts[0]))
			if err != nil {
				return nil, err
			}

			clozes = append(clozes, cloze{
				withUnderscores:    withUnderscores,
				withoutParenthesis: withoutParenthesis,
				word:               word,
				note:               parts[1],
			})
			continue
		}
		withoutParenthesis, withUnderscores, word, err := processSentence(strings.TrimSpace(s))
		if err != nil {
			return nil, err
		}

		clozes = append(clozes, cloze{
			withUnderscores:    withUnderscores,
			withoutParenthesis: withoutParenthesis,
			word:               word,
		})
	}
	return clozes, nil
}

// processSentence processes a Chinese sentence, extracts words in parentheses,
// and returns three variables based on the criteria.
func processSentence(cloze string) (string, string, string, error) {
	// Regular expression to find the word inside parentheses
	re := regexp.MustCompile(`\(([^)]+)\)`)

	// Check if there's a match
	matches := re.FindStringSubmatch(cloze)
	if len(matches) < 2 {
		return "", "", "", fmt.Errorf("no word wrapped in parentheses found: %s", cloze)
	}

	// Extract the word between parentheses
	word := matches[1]

	// Create the sentence without the parentheses and the word
	withoutParenthesis := strings.Replace(cloze, "("+word+")", word, 1)

	// Create the sentence with the word in parentheses replaced by underscores
	withUnderscores := strings.Replace(cloze, "("+word+")", "___", 1)

	// Return the results
	return withoutParenthesis, withUnderscores, strings.TrimSpace(word), nil
}
