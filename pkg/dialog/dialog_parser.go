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
	Note               string
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
	var textWithSpeaker, textWithOutSpeaker, text, note string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		rawLine := scanner.Text()
		if strings.HasPrefix(rawLine, "|") {
			note = strings.TrimSpace(strings.Trim(rawLine, "|"))
			continue
		}
		if rawLine == "---" {
			dialogues = append(
				dialogues,
				RawDialog{
					Speakers:           speakers,
					Lines:              lines,
					TextWithSpeaker:    strings.TrimSpace(textWithSpeaker),
					TextWithOutSpeaker: strings.TrimSpace(textWithOutSpeaker),
					Text:               text,
					Note:               note,
				},
			)
			textWithSpeaker = ""
			textWithOutSpeaker = ""
			lines = []DialogLine{}
			speakers = make(map[string]struct{})
			text = ""
			note = ""
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
			Note:               note,
		},
	)
	return dialogues
}

func splitSpeakerAndText(line string) DialogLine {
	parts := []string{}
	if strings.Contains(line, ":") {
		parts = strings.Split(line, ":")
	} else if strings.Contains(line, "：") {
		parts = strings.Split(line, "：")
	}
	if len(parts) == 0 {
		return DialogLine{
			"",
			line,
		}
	}
	return DialogLine{
		parts[0],
		parts[1],
	}
}
