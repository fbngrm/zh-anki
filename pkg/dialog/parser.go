package dialog

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func load(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open dialogues file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var dialogues []string
	var dialog string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			dialogues = append(dialogues, strings.TrimSpace(dialog))
			dialog = ""
			continue
		}
		dialog += " "
		dialog += line
	}
	dialogues = append(dialogues, dialog)
	return dialogues
}
