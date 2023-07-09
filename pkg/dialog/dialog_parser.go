package dialog

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type DialogLine struct {
	Speaker string
	Text    string
}

type RawDialog struct {
	Speakers           map[string]struct{}
	Lines              []DialogLine
	Text               string // one line without speaker prefixes
	TextWithSpeaker    string
	TextWithOutSpeaker string
}

func loadDialogues(path string) []RawDialog {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open dialogues file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var dialogues []RawDialog
	speakers := make(map[string]struct{})
	var lines []DialogLine
	var textWithSpeaker, textWithOutSpeaker, text string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawLine := scanner.Text()
		if rawLine == "---" {
			dialogues = append(
				dialogues,
				RawDialog{
					Speakers:           speakers,
					Lines:              lines,
					TextWithSpeaker:    strings.TrimSpace(textWithSpeaker),
					TextWithOutSpeaker: strings.TrimSpace(textWithOutSpeaker),
					Text:               text,
				},
			)
			textWithSpeaker = ""
			textWithOutSpeaker = ""
			lines = []DialogLine{}
			speakers = make(map[string]struct{})
			text = ""
			continue
		}
		line := splitSpeakerAndText(rawLine)
		lines = append(lines, line)
		speakers[line.Speaker] = struct{}{}

		textWithSpeaker += rawLine
		textWithSpeaker += "<br>"
		textWithOutSpeaker += line.Text
		textWithOutSpeaker += "<br>"

		text += line.Text
		text += " "
	}
	dialogues = append(
		dialogues,
		RawDialog{
			Speakers:           speakers,
			Lines:              lines,
			TextWithSpeaker:    strings.TrimSpace(textWithSpeaker),
			TextWithOutSpeaker: strings.TrimSpace(textWithOutSpeaker),
			Text:               text,
		},
	)
	return dialogues
}

// note, dialogues with a speaker must use `：` (unicode) to separate speaker and text. This is not the same as `:` (ascii)!
func splitSpeakerAndText(line string) DialogLine {
	parts := strings.Split(line, "：")
	if len(parts) == 1 {
		return DialogLine{
			"",
			parts[0],
		}
	}
	return DialogLine{
		parts[0],
		parts[1],
	}
}
