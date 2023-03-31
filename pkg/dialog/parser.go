package dialog

import (
	"bufio"
	"fmt"
	"os"
)

func load(path string) [][]string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open dialogues file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	var dialogues [][]string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			dialogues = append(dialogues, lines)
			lines = []string{}
			continue
		}
		lines = append(lines, line)
	}
	return dialogues
}
