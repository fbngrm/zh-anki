package dialog

import (
	"fmt"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"golang.org/x/exp/slog"
)

func ExportCloze(deckName string, cl Cloze, i ignore.Ignored) error {
	defer func() {
		i.Update(cl.Word.Chinese)
	}()
	// add cards for all chars in the cloze's word
	for _, c := range cl.Word.Chars {
		if err := char.Export(deckName, c, i); err != nil {
			slog.Error("export char for word", "word", cl.Word.Chinese, "char", c.Chinese, "error", err)
		}
	}
	cedictHeader := ""
	cedictEn1, cedictPinyin1 := "", ""
	cedictEn2, cedictPinyin2 := "", ""
	cedictEn3, cedictPinyin3 := "", ""
	if len(cl.Word.Cedict) >= 1 {
		cedictHeader = "Cedict<br>"
		cedictEn1 = cl.Word.Cedict[0].CedictEnglish + "<br>" + "<br>"
		cedictPinyin1 = cl.Word.Cedict[0].CedictPinyin + "<br>"
	}
	if len(cl.Word.Cedict) >= 2 {
		cedictEn2 = cl.Word.Cedict[1].CedictEnglish + "<br>" + "<br>"
		cedictPinyin2 = cl.Word.Cedict[1].CedictPinyin + "<br>"
	}
	if len(cl.Word.Cedict) >= 3 {
		cedictEn3 = cl.Word.Cedict[2].CedictEnglish + "<br>" + "<br>"
		cedictPinyin3 = cl.Word.Cedict[2].CedictPinyin + "<br>"
	}

	hskHeader, hskEn, hskPinyin := "", "", ""
	if len(cl.Word.HSK) >= 1 {
		hskHeader = "HSK 3.0<br>"
		hskEn = cl.Word.HSK[0].HSKEnglish + "<br>" + "<br>"
		hskPinyin = cl.Word.HSK[0].HSKPinyin + "<br>"
	}
	if hskEn == "" && cedictEn1 == "" {
		return fmt.Errorf("no translation for word: %s", cl.Word.Chinese)
	}

	noteHeader, note := "", ""
	if len(cl.Word.Note) >= 1 {
		noteHeader = "Note<br>"
		note = cl.Word.Note + "<br>" + "<br>"
	}

	transHeader, trans := "", ""
	if len(cl.Word.Translation) >= 1 {
		transHeader = "Translation" + "<br>"
		trans = cl.Word.Translation + "<br>" + "<br>"
	}

	exampleWordsHeader, exampleWords := "", ""
	if len(cl.Word.Example) >= 1 {
		exampleWordsHeader = "Example Words" + "<br>"
		exampleWords = cl.Word.Example + "<br>" + "<br>"
	}

	examplesSentencesHeader := ""
	exSentence1, exSentencePi1, exSentenceEn1, exSentenceAudio1 := "", "", "", ""
	exSentence2, exSentencePi2, exSentenceEn2, exSentenceAudio2 := "", "", "", ""
	if len(cl.Word.Examples) >= 1 {
		examplesSentencesHeader = "Example Sentences<br>"
		exSentence1 = strings.ReplaceAll(cl.Word.Examples[0].Chinese, " ", "") + "<br>"
		exSentencePi1 = cl.Word.Examples[0].Pinyin + "<br>"
		exSentenceEn1 = cl.Word.Examples[0].English + "<br>"
		exSentenceAudio1 = cl.Word.Examples[0].Audio
	}
	if len(cl.Word.Examples) >= 2 {
		exSentence2 = strings.ReplaceAll(cl.Word.Examples[1].Chinese, " ", "") + "<br>"
		exSentencePi2 = cl.Word.Examples[1].Pinyin + "<br>"
		exSentenceEn2 = cl.Word.Examples[1].English + "<br>"
		exSentenceAudio2 = cl.Word.Examples[1].Audio
	}

	noteFields := map[string]string{
		"Chinese":                cl.Word.Chinese,
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
		"Audio":                  anki.GetAudioPath(cl.Word.Audio),
		"Components":             componentsToString(cl.Word.Components),
		"Traditional":            cl.Word.Traditional,
		"ExampleWordsHeader":     exampleWordsHeader,
		"Examples":               exampleWords,
		"ExampleSentencesHeader": examplesSentencesHeader,
		"ExampleSentenceCh1":     exSentence1,
		"ExampleSentencePi1":     exSentencePi1,
		"ExampleSentenceEn1":     exSentenceEn1,
		"ExampleSentenceAudio1":  anki.GetAudioPath(exSentenceAudio1) + "<br>",
		"ExampleSentenceCh2":     exSentence2,
		"ExampleSentencePi2":     exSentencePi2,
		"ExampleSentenceEn2":     exSentenceEn2,
		"ExampleSentenceAudio2":  anki.GetAudioPath(exSentenceAudio2),
		"MnemonicBase":           cl.Word.MnemonicBase,
		"Mnemonic":               cl.Word.Mnemonic,
		"NoteHeader":             noteHeader,
		"Note":                   note,
		"TranslationHeader":      transHeader,
		"Translation":            trans,
		// cloze sentence fields
		"SentenceFront":   strings.ReplaceAll(cl.SentenceFront, " ", ""),
		"SentenceBack":    strings.ReplaceAll(cl.SentenceBack, " ", ""),
		"SentencePinyin":  cl.Pinyin,
		"SentenceEnglish": cl.English,
		"SentenceAudio":   anki.GetAudioPath(cl.Audio),
	}
	_, err := anki.AddNoteToDeck(deckName, "cloze", noteFields)
	if err != nil {
		return fmt.Errorf("add cloze note [%s]: %w", cl.SentenceBack, err)
	}
	slog.Info("note added successfully", "cloze", cl.SentenceBack)
	return nil
}
