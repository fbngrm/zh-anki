package dialog

import (
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/ignore"
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

	transHeader, trans := "", ""
	if len(w.Translation) >= 1 {
		transHeader = "Translation" + "<br>"
		trans = w.Translation + "<br>" + "<br>"
	}

	trad := ""
	if w.Traditional != w.Chinese {
		trad = w.Traditional
	}

	examplesHeader := ""
	if len(w.Example) >= 1 {
		examplesHeader = "Examples" + "<br>"
	}

	examplesSentencesHeader := ""
	exSentence1, exSentencePi1, exSentenceEn1, exSentenceAudio1 := "", "", "", ""
	exSentence2, exSentencePi2, exSentenceEn2, exSentenceAudio2 := "", "", "", ""
	if len(w.Examples) >= 1 {
		examplesSentencesHeader = "Example Sentences<br>"
		exSentence1 = w.Examples[0].Chinese + "<br>"
		exSentencePi1 = w.Examples[0].Pinyin + "<br>"
		exSentenceEn1 = w.Examples[0].English + "<br>"
		exSentenceAudio1 = w.Examples[0].Audio
	}
	if len(w.Examples) >= 2 {
		exSentence2 = w.Examples[1].Chinese + "<br>"
		exSentencePi2 = w.Examples[1].Pinyin + "<br>"
		exSentenceEn2 = w.Examples[1].English + "<br>"
		exSentenceAudio2 = w.Examples[1].Audio
	}

	noteFields := map[string]string{
		"Chinese":                w.Chinese,
		"CedictHeader":           cedictHeader,
		"CedictPinyin1":          cedictPinyin1,
		"CedictEnglish1":         cedictEn1,
		"CedictPinyin2":          cedictPinyin2,
		"CedictEnglish2":         cedictEn2,
		"CedictPinyin3":          cedictPinyin3,
		"CedictEnglish3":         cedictEn3,
		"HSKHeader":              hskHeader,
		"HSKPinyin":              hskPinyin,
		"HSKEnglish":             hskEn,
		"Audio":                  anki.GetAudioPath(w.Audio),
		"Components":             componentsToString(w.Components),
		"Traditional":            trad,
		"ExamplesHeader":         examplesHeader,
		"Examples":               w.Example,
		"MnemonicBase":           w.MnemonicBase,
		"Mnemonic":               w.Mnemonic,
		"NoteHeader":             noteHeader,
		"Note":                   note,
		"TranslationHeader":      transHeader,
		"Translation":            trans,
		"ExampleSentencesHeader": examplesSentencesHeader,
		"ExampleSentenceCh1":     exSentence1,
		"ExampleSentencePi1":     exSentencePi1,
		"ExampleSentenceEn1":     exSentenceEn1,
		"ExampleSentenceAudio1":  anki.GetAudioPath(exSentenceAudio1) + "<br>",
		"ExampleSentenceCh2":     exSentence2,
		"ExampleSentencePi2":     exSentencePi2,
		"ExampleSentenceEn2":     exSentenceEn2,
		"ExampleSentenceAudio2":  anki.GetAudioPath(exSentenceAudio2),
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
