package char

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"golang.org/x/exp/slog"
)

// FIXME: this is redundant with Word, move to same pkg and remove one.
func Export(deckName string, c Char, i ignore.Ignored) error {
	defer func() {
		i.Update(c.Chinese)
	}()
	if _, ok := i[c.Chinese]; ok {
		slog.Debug("exists in ignore list", "char", c.Chinese)
		return nil
	}
	cedictHeader := ""
	cedictEn1, cedictPinyin1 := "", ""
	cedictEn2, cedictPinyin2 := "", ""
	cedictEn3, cedictPinyin3 := "", ""
	if len(c.Cedict) >= 1 {
		cedictHeader = "Cedict<br>"
		cedictEn1 = c.Cedict[0].CedictEnglish + "<br>" + "<br>"
		cedictPinyin1 = c.Cedict[0].CedictPinyin + "<br>"
	}
	if len(c.Cedict) >= 2 {
		cedictEn2 = c.Cedict[1].CedictEnglish + "<br>" + "<br>"
		cedictPinyin2 = c.Cedict[1].CedictPinyin + "<br>"
	}
	if len(c.Cedict) >= 3 {
		cedictEn3 = c.Cedict[2].CedictEnglish + "<br>" + "<br>"
		cedictPinyin3 = c.Cedict[2].CedictPinyin + "<br>"
	}

	hskHeader, hskEn, hskPinyin := "", "", ""
	if len(c.HSK) >= 1 {
		hskHeader = "HSK 3.0<br>"
		hskEn = c.HSK[0].HSKEnglish + "<br>" + "<br>"
		hskPinyin = c.HSK[0].HSKPinyin + "<br>"
	}

	transHeader, trans := "", ""
	if len(c.HSK) >= 1 {
		transHeader = "Translation" + "<br>"
		trans = c.Translation + "<br>" + "<br>"
	}

	noteFields := map[string]string{
		"Chinese":           c.Chinese,
		"CedictHeader":      cedictHeader,
		"CedictPinyin1":     cedictPinyin1,
		"CedictEnglish1":    cedictEn1,
		"CedictPinyin2":     cedictPinyin2,
		"CedictEnglish2":    cedictEn2,
		"CedictPinyin3":     cedictPinyin3,
		"CedictEnglish3":    cedictEn3,
		"HSKHeader":         hskHeader,
		"HSKPinyin":         hskPinyin,
		"HSKEnglish":        hskEn,
		"Audio":             anki.GetAudioPath(c.Audio),
		"Components":        componentsToString(c.Components),
		"Traditional":       c.Traditional,
		"Examples":          c.Example,
		"MnemonicBase":      c.MnemonicBase,
		"Mnemonic":          c.Mnemonic,
		"Pronounciation":    c.Pronounciation,
		"TranslationHeader": transHeader,
		"Translation":       trans,
	}
	_, err := anki.AddNoteToDeck(deckName, "char_cedict3", noteFields)
	if err != nil {
		return fmt.Errorf("add char note [%s]: %w", c.Chinese, err)
	}
	slog.Info("added successfully", "char", c.Chinese)
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
