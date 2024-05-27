package dialog

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"golang.org/x/exp/slog"
)

func ExportDialog(renderSentences bool, d *Dialog, i ignore.Ignored) error {
	if renderSentences {
		for _, s := range d.Sentences {
			if err := ExportSentence(d.Deck, s, i); err != nil {
				slog.Error("add note", "sentence", s.Chinese, "error", err)
			}
		}
	}
	noteFields := map[string]string{
		"Chinese":    d.Chinese,
		"Pinyin":     d.Pinyin,
		"English":    d.English,
		"Audio":      anki.GetAudioPath(d.Audio),
		"Components": sentencesToString(d.Sentences),
	}
	_, err := anki.AddNoteToDeck(d.Deck, "dialog", noteFields)
	if err != nil {
		return fmt.Errorf("add dialog note [%s]: %w", d.Chinese[:25], err)
	}
	slog.Info("note added successfully", "dialog", d.Chinese[:25])
	return nil
}

func sentencesToString(sentences []Sentence) string {
	s := ""
	for _, sentence := range sentences {
		for _, w := range sentence.Words {
			s = fmt.Sprintf(`%s
<a href="https://hanzicraft.com/character/%s">%s</a> %s
<br/>`, s, w.Chinese, w.Chinese, w.English)
		}
	}
	return s
}
