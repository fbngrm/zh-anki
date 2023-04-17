package sentence

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func load(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open sentences file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var sentences []string
	var sentence string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		sentences = append(sentences, strings.TrimSpace(sentence))
	}
	sentences = append(sentences, sentence)
	return sentences
}
