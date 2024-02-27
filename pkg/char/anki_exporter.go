package char

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
)

func Export(deckName string, c Char) error {
	noteFields := map[string]string{
		"Chinese":      c.Chinese,
		"Pinyin":       c.Pinyin,
		"English":      c.English,
		"Audio":        anki.GetAudioPath(c.Audio),
		"Components":   componentsToString(c.Components),
		"Kangxi":       "",
		"Equivalents":  componentsToString([]string{c.Equivalents}),
		"Traditional":  componentsToString([]string{c.Traditional}),
		"Examples":     c.Example,
		"MnemonicBase": c.MnemonicBase,
		"Mnemonic":     c.Mnemonic,
	}
	noteID, err := anki.AddNoteToDeck(deckName, "char", noteFields)
	if err != nil {
		return fmt.Errorf("add char note [%s]: %w", c.Chinese, err)
	}
	fmt.Println("char note added successfully! ID:", noteID)
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
