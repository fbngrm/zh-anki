package dialog

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-freq/pkg/card"
	"golang.org/x/exp/slog"
)

func ExportWord(deckName string, w Word, i ignore.Ignored) error {
	defer func() {
		i.Update(w.Chinese)
	}()
	for _, c := range w.Chars {
		if err := char.Export(deckName, c, i); err != nil {
			slog.Error("export char for word", "word", w.Chinese, "char", c.Chinese, "error", err)
		}
	}
	if _, ok := i[w.Chinese]; ok {
		slog.Debug("export word, exists in ignore list", "word", w.Chinese)
		return nil
	}
	cedictHeader := ""
	cedictEn1, cedictPinyin1 := "", ""
	cedictEn2, cedictPinyin2 := "", ""
	cedictEn3, cedictPinyin3 := "", ""
	if len(w.Cedict) >= 1 {
		cedictHeader = "Cedict<br>"
		cedictEn1 = w.Cedict[0].CedictEnglish + "<br>" + "<br>"
		cedictPinyin1 = w.Cedict[0].CedictPinyin + "<br>"
	}
	if len(w.Cedict) >= 2 {
		cedictEn2 = w.Cedict[1].CedictEnglish + "<br>" + "<br>"
		cedictPinyin2 = w.Cedict[1].CedictPinyin + "<br>"
	}
	if len(w.Cedict) >= 3 {
		cedictEn3 = w.Cedict[2].CedictEnglish + "<br>" + "<br>"
		cedictPinyin3 = w.Cedict[2].CedictPinyin + "<br>"
	}

	hskHeader, hskEn, hskPinyin := "", "", ""
	if len(w.HSK) >= 1 {
		hskHeader = "HSK 3.0<br>"
		hskEn = w.HSK[0].HSKEnglish + "<br>" + "<br>"
		hskPinyin = w.HSK[0].HSKPinyin + "<br>"
	}
	if hskEn == "" && cedictEn1 == "" {
		return fmt.Errorf("no translation for word: %s", w.Chinese)
	}

	noteHeader, note := "", ""
	if len(w.Note) >= 1 {
		noteHeader = "Note<br>"
		note = w.Note + "<br>" + "<br>"
	}

	noteFields := map[string]string{
		"Chinese":        w.Chinese,
		"CedictHeader":   cedictHeader,
		"CedictPinyin1":  cedictPinyin1,
		"CedictEnglish1": cedictEn1,
		"CedictPinyin2":  cedictPinyin2,
		"CedictEnglish2": cedictEn2,
		"CedictPinyin3":  cedictPinyin3,
		"CedictEnglish3": cedictEn3,
		"HSKHeader":      hskHeader,
		"HSKPinyin":      hskPinyin,
		"HSKEnglish":     hskEn,
		"Audio":          anki.GetAudioPath(w.Audio),
		"Components":     componentsToString(w.Components),
		"Traditional":    w.Traditional,
		"Examples":       w.Example,
		"MnemonicBase":   w.MnemonicBase,
		"Mnemonic":       w.Mnemonic,
		"NoteHeader":     noteHeader,
		"Note":           note,
	}
	_, err := anki.AddNoteToDeck(deckName, "word_cedict3", noteFields)
	if err != nil {
		return fmt.Errorf("add word note [%s]: %w", w.Chinese, err)
	}
	slog.Info("added successfully", "word", w.Chinese)
	return nil
}

func componentsToString(components []card.Component) string {
	s := ""
	for _, c := range components {
		s = fmt.Sprintf(`%s
<a href="https://hanzicraft.com/character/%s">%s</a> = %s
<br/>`, s, c.SimplifiedChinese, c.SimplifiedChinese, c.English)
	}
	return s
}
