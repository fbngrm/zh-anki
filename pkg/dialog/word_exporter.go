package dialog

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/char"
)

func ExportWord(deckName string, w Word) error {
	for _, c := range w.NewChars {
		if err := char.Export(deckName, c); err != nil {
			return err
		}
	}
	noteFields := map[string]string{
		"Chinese":      w.Chinese,
		"Pinyin":       w.Pinyin,
		"English":      w.English,
		"Audio":        anki.GetAudioPath(w.Audio),
		"Components":   componentsToString(w.Components),
		"Traditional":  componentsToString([]string{w.Traditional}),
		"Mnemonic":     w.Mnemonic,
		"MnemonicBase": w.MnemonicBase,
	}
	noteID, err := anki.AddNoteToDeck(deckName, "word", noteFields)
	if err != nil {
		return fmt.Errorf("add word note [%s]: %w", w.Chinese, err)
	}
	fmt.Println("word note added successfully! ID:", noteID)
	return nil
}

func componentsToString(comps []string) string {
	s := ""
	for _, c := range comps {
		s = fmt.Sprintf(`%s
<a href="https://hanzicraft.com/character/%s">%s</a>
<br/>`, s, c, c)
	}
	return s
}
