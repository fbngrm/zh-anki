package dialog

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func loadWords(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open words file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var words []string
	var word string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word = scanner.Text()
		if word == "" {
			continue
		}
		words = append(words, strings.TrimSpace(word))
	}
	words = append(words, word)
	return words
}
