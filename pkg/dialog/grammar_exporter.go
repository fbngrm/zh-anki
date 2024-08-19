package dialog

import (
	"strings"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"golang.org/x/exp/slog"
)

func ExportGrammar(deckName string, g Grammar) error {
	noteHeader, note := "", ""
	if len(g.ExampleDescription) >= 1 {
		noteHeader = "Note<br>"
		note = g.ExampleDescription + "<br>" + "<br>"
	}
	examplesHeader, examples := "", ""
	if len(g.Examples) >= 1 {
		examplesHeader = "Examples<br>"
		examples = strings.Join(strings.Split(g.Examples, "\n"), "<br>") + "<br>" + "<br>"
	}
	syntaxHeader, syntax := "", ""
	if len(g.Structure) >= 1 {
		syntaxHeader = "Syntax<br>"
		syntax = g.Structure + "<br>" + "<br>"
	}
	explanationHeader, explanation := "", ""
	if len(g.Description) >= 1 {
		explanationHeader = "Explanation<br>"
		explanation = strings.Join(strings.Split(g.Description, "\n"), "<br>") + "<br>" + "<br>"
	}

	noteFields := map[string]string{
		"Header":            g.Head,
		"Explanation":       explanation,
		"ExplanationHeader": explanationHeader,
		"Syntax":            syntax,
		"SyntaxHeader":      syntaxHeader,
		"ExamplesHeader":    examplesHeader,
		"Examples":          examples,
		"Note":              note,
		"NoteHeader":        noteHeader,
		"Audio":             anki.GetAudioPath(g.Audio),
	}

	_, err := anki.AddNoteToDeck(deckName, "zh-grammar", noteFields)
	if err != nil {
		return err
	}
	slog.Info("note added successfully", "grammar", g.Structure)
	return nil
}
