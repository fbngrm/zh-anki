package dialog

import (
	"fmt"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"golang.org/x/exp/slog"
)

func ExportSentence(deckName string, s Sentence, i ignore.Ignored) error {
	for _, w := range s.Words {
		if err := ExportWord(deckName, w, i); err != nil {
			slog.Error("exporting word when exporting sentence", "sentence", w.Chinese, "error", err)
		}
	}
	noteFields := map[string]string{
		"Chinese":    strings.ReplaceAll(s.Chinese, " ", ""),
		"Pinyin":     s.Pinyin,
		"English":    s.English,
		"Audio":      anki.GetAudioPath(s.Audio),
		"Components": wordsToString(s.Words),
		"Note":       s.Note,
		"Grammar":    s.Grammar,
	}
	_, err := anki.AddNoteToDeck(deckName, "sentence", noteFields)
	if err != nil {
		return err
	}
	slog.Info("note added successfully", "sentence", s.Chinese)
	return nil
}

func wordsToString(words []Word) string {
	s := ""
	for _, word := range words {
		s = fmt.Sprintf(`%s
<a href="https://hanzicraft.com/character/%s">%s</a> %s
<br/>`, s, word.Chinese, word.Chinese, word.English)
	}
	return s
}
