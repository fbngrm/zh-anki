package dialog

import (
	"strings"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"golang.org/x/exp/slog"
)

func ExportGrammar(deckName string, g Grammar) error {
	syntaxHeader, syntax := "", ""
	if len(g.Structure) >= 1 {
		syntaxHeader = "Syntax<br>"
		syntax = g.Structure + "<br><br>"
	}

	noteHeader, note := "", ""
	if len(g.Note) >= 1 {
		noteHeader = "Note<br>"
		note = strings.Join(strings.Split(g.Note, "\n"), "<br>") + "<br><br>"
	}

	examplesSentencesHeader := ""
	exSentence1, exSentencePi1, exSentenceEn1, exSentenceAudio1 := "", "", "", ""
	exSentence2, exSentencePi2, exSentenceEn2, exSentenceAudio2 := "", "", "", ""
	exSentence3, exSentencePi3, exSentenceEn3, exSentenceAudio3 := "", "", "", ""
	if len(g.Examples) >= 1 {
		examplesSentencesHeader = "Example Sentences<br>"
		exSentence1 = g.Examples[0].Chinese + "<br>"
		exSentencePi1 = g.Examples[0].Pinyin + "<br>"
		exSentenceEn1 = g.Examples[0].English + "<br>"
		exSentenceAudio1 = anki.GetAudioPath(g.Examples[0].Audio) + "<br>"
	}
	if len(g.Examples) >= 2 {
		exSentence2 = g.Examples[1].Chinese + "<br>"
		exSentencePi2 = g.Examples[1].Pinyin + "<br>"
		exSentenceEn2 = g.Examples[1].English + "<br>"
		exSentenceAudio2 = anki.GetAudioPath(g.Examples[1].Audio) + "<br>"
	}
	if len(g.Examples) >= 3 {
		exSentence3 = g.Examples[2].Chinese + "<br>"
		exSentencePi3 = g.Examples[2].Pinyin + "<br>"
		exSentenceEn3 = g.Examples[2].English + "<br>"
		exSentenceAudio3 = anki.GetAudioPath(g.Examples[2].Audio) + "<br>"
	}

	summaryHeader, summary := "", ""
	if len(g.Summary) > 1 {
		summary = "Key points to remember:"
		summary += "<ul>"
		summaryHeader = "Summary<br>"
		for _, p := range g.Summary {
			if len(p) == 0 {
				continue
			}
			summary += "<li>"
			summary += p
			summary += "</li>"
		}
		summary += "</ui>"
		summary += "<br><br>"
	} else if len(g.Summary) == 0 {
		summary = g.Summary[0]
	}

	noteFields := map[string]string{
		"SentenceFront":         g.SentenceFront,
		"SentenceBack":          g.SentenceBack,
		"SentencePinyin":        g.SentencePinyin,
		"SentenceEnglish":       g.SentenceEnglish,
		"SentenceAudio":         anki.GetAudioPath(g.SentenceAudio),
		"Pattern":               g.Pattern,
		"NoteHeader":            noteHeader,
		"Note":                  note,
		"SyntaxHeader":          syntaxHeader,
		"Syntax":                syntax,
		"ExamplesHeader":        examplesSentencesHeader,
		"ExampleSentenceCh1":    exSentence1,
		"ExampleSentencePi1":    exSentencePi1,
		"ExampleSentenceEn1":    exSentenceEn1,
		"ExampleSentenceAudio1": exSentenceAudio1,
		"ExampleSentenceCh2":    exSentence2,
		"ExampleSentencePi2":    exSentencePi2,
		"ExampleSentenceEn2":    exSentenceEn2,
		"ExampleSentenceAudio2": exSentenceAudio2,
		"ExampleSentenceCh3":    exSentence3,
		"ExampleSentencePi3":    exSentencePi3,
		"ExampleSentenceEn3":    exSentenceEn3,
		"ExampleSentenceAudio3": exSentenceAudio3,
		"SummaryHeader":         summaryHeader,
		"Summary":               summary,
	}

	_, err := anki.AddNoteToDeck(deckName, "pattern", noteFields)
	if err != nil {
		return err
	}
	slog.Info("note added successfully", "grammar", g.Structure)
	return nil
}
