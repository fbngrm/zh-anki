package dialog

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
)

func ExportGrammar(deckName string, g SimpleGrammar) error {
	noteFields := map[string]string{
		"Front": g.Head,
		"Back":  g.HTML,
	}
	noteID, err := anki.AddNoteToDeck(deckName, "ch-grammar2", noteFields)
	if err != nil {
		return fmt.Errorf("add simple grammar note [%s]: %w", g.Head, err)
	}
	fmt.Println("simple grammar note added successfully! ID:", noteID)
	return nil
}
