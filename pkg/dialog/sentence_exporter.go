package dialog

import (
	"fmt"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/anki"
)

func ExportSentence(deckName string, s Sentence) error {
	for _, w := range s.NewWords {
		if err := ExportWord(deckName, w); err != nil {
			fmt.Printf("error exporting word when exporting sentence [%s]: %v", w.Chinese, err)
		}
	}
	noteFields := map[string]string{
		"Chinese":    strings.ReplaceAll(s.Chinese, " ", ""),
		"Pinyin":     s.Pinyin,
		"English":    s.English,
		"Audio":      anki.GetAudioPath(s.Audio),
		"Components": wordsToString(s.AllWords),
		"Note":       s.Note,
		"Grammar":    s.Grammar,
	}
	noteID, err := anki.AddNoteToDeck(deckName, "sentence", noteFields)
	if err != nil {
		return fmt.Errorf("add sentence note [%s]: %w", s.Chinese, err)
	}
	fmt.Println("sentence note added successfully! ID:", noteID)
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
