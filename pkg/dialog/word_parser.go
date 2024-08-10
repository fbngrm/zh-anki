package dialog

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadWords(path string) []Word {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open words file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var words []Word
	var word string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word = scanner.Text()
		if word == "" {
			continue
		}
		if strings.Contains(word, "|") {
			parts := strings.Split(word, "|")
			if len(parts) > 2 {
				fmt.Printf("could not split word, too many '|' : %s\n", word)
				continue
			}
			words = append(words, Word{
				Chinese: strings.TrimSpace(parts[0]),
				Note:    parts[1],
			})
			continue
		}
		words = append(words, Word{
			Chinese: strings.TrimSpace(word),
		})
	}
	words = append(words, Word{})
	return words
}
