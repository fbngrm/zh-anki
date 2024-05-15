package dialog

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type sentence struct {
	text    string
	grammar string
	note    string
}

func loadSentences(path string) []sentence {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open sentences file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var sentences []sentence
	var s string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s = scanner.Text()
		if s == "" {
			continue
		}
		if strings.Contains(s, "|") {
			parts := strings.Split(s, "|")
			if len(parts) > 2 {
				fmt.Printf("could not split sentence, too many '|' : %s\n", s)
				continue
			}
			sentences = append(sentences, sentence{
				text: strings.TrimSpace(parts[0]),
				note: parts[1],
			})
			continue
		}
		// chinese "|"
		if strings.Contains(s, "|") {
			parts := strings.Split(s, "|")
			if len(parts) > 2 {
				fmt.Printf("could not split sentence, too many '|' : %s\n", s)
				continue
			}
			sentences = append(sentences, sentence{
				text: strings.TrimSpace(parts[0]),
				note: parts[1],
			})
			continue
		}
		if strings.Contains(s, ">") || strings.Contains(s, "》") {
			parts := strings.Split(s, ">")
			if len(parts) > 2 {
				fmt.Printf("could not split sentence, too many '>' : %s\n", s)
				continue
			}
			sentences = append(sentences, sentence{
				text:    strings.TrimSpace(parts[0]),
				grammar: parts[1],
			})
			continue
		}
		if strings.Contains(s, "》") {
			parts := strings.Split(s, "》")
			if len(parts) > 2 {
				fmt.Printf("could not split sentence, too many '》' : %s\n", s)
				continue
			}
			sentences = append(sentences, sentence{
				text:    strings.TrimSpace(parts[0]),
				grammar: parts[1],
			})
			continue
		}
		sentences = append(sentences, sentence{text: strings.TrimSpace(s)})
	}
	return sentences
}
