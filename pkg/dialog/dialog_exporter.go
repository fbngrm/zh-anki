package dialog

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
)

func ExportDialog(renderSentences bool, d *Dialog) error {
	if renderSentences {
		for _, s := range d.Sentences {
			if err := ExportSentence(d.Deck, s); err != nil {
				fmt.Printf("error sentence when exporting dialog [%s]: %v\n", s.Chinese, err)
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
	noteID, err := anki.AddNoteToDeck(d.Deck, "dialog", noteFields)
	if err != nil {
		return fmt.Errorf("add dialog note [%s]: %w", d.Chinese[:25], err)
	}
	fmt.Println("dialog note added successfully! ID:", noteID)
	return nil
}

func sentencesToString(sentences []Sentence) string {
	s := ""
	for _, sentence := range sentences {
		for _, w := range sentence.AllWords {
			s = fmt.Sprintf(`%s
<a href="https://hanzicraft.com/character/%s">%s</a> %s
<br/>`, s, w.Chinese, w.Chinese, w.English)
		}
	}
	return s
}
