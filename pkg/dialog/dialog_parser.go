package dialog

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type RawDialog struct {
	Text string
}

func loadDialogues(path string) []RawDialog {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open dialogues file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var dialogues []RawDialog
	var dialog string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			dialogues = append(
				dialogues,
				RawDialog{Text: strings.TrimSpace(dialog)},
			)
			dialog = ""
			continue
		}
		dialog += " "
		dialog += line
	}
	dialogues = append(
		dialogues,
		RawDialog{Text: strings.TrimSpace(dialog)},
	)
	return dialogues
}
