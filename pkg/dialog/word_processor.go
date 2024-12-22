package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"golang.org/x/exp/slog"
)

type WordProcessor struct {
	Chars       char.Processor
	GCPAudio    *audio.GCPClient
	AzureAudio  *audio.AzureClient
	IgnoreChars []string
	Client      *openai.Client
	WordIndex   *frequency.WordIndex
	CardBuilder *card.Builder
}

func (p *WordProcessor) DecomposeFromFile(path, outdir string, t *translate.Translations) []Word {
	words := loadWords(path)

	var newWords []Word
	for _, word := range words {
		if word.Chinese == "" {
			continue
		}
		if contains(p.IgnoreChars, word.Chinese) {
			continue
		}

		w, err := p.Decompose(word, t)
		if err != nil {
			slog.Error("decompose", "word", word, "err", err)
			continue
		}

		newWords = append(newWords, *w)
	}
	return newWords
}

func (p *WordProcessor) Decompose(w Word, t *translate.Translations) (*Word, error) {
	allChars := p.Chars.GetAll(w.Chinese, false, t)

	var cc *card.Card
	var err error
	isSingleRune := utf8.RuneCountInString(w.Chinese) == 1
	exampleWords := ""
	if isSingleRune {
		exampleWords = removeRedundant(p.WordIndex.GetExamplesForHanzi(w.Chinese, 5))
		cc = p.CardBuilder.GetHanziCard(w.Chinese, t)
	} else {
		cc, err = p.CardBuilder.GetWordCard(w.Chinese, t)
		if err != nil {
			return nil, err
		}
	}

	examples, err := p.Client.GetExamplesForWord(w.Chinese)
	if err != nil {
		slog.Error("fetch example sentences", "word", w.Chinese, "err", err)
	}

	trad := ""
	if cc.TraditionalChinese != w.Chinese {
		trad = cc.TraditionalChinese
	}

	newWord := Word{
		Chinese:      w.Chinese,
		Cedict:       card.GetCedictEntries(cc),
		HSK:          card.GetHSKEntries(cc),
		Chars:        allChars,
		IsSingleRune: isSingleRune,
		Components:   cc.Components,
		Traditional:  trad,
		Example:      exampleWords,
		Examples:     p.getExampleSentences(examples.Examples),
		MnemonicBase: cc.MnemonicBase,
		Mnemonic:     cc.Mnemonic,
		Note:         p.getNote(w.Note, examples.Note),
		Translation:  cc.Translation,
		Audio:        p.getAudio(w.Chinese),
	}
	return &newWord, nil
}

// We get a note on usage of the word from ChatGPT and add it to the user defined note (if any).
func (p *WordProcessor) getNote(userNote, examplesNote string) string {
	if userNote != "" {
		userNote = userNote + "<br><br>"
	}
	return userNote + examplesNote
}

func (p *WordProcessor) getExampleSentences(examples []openai.Word) []card.Example {
	results := make([]card.Example, len(examples))
	for i, e := range examples {
		results[i] = card.Example{
			Chinese: strings.ReplaceAll(e.Ch, " ", ""),
			Pinyin:  e.Pi,
			English: e.En,
			Audio:   p.getExampleSentenceAudio(e.Ch),
		}
	}
	return results
}

func (p *WordProcessor) getExampleSentenceAudio(s string) string {
	filename := strings.ReplaceAll(s, " ", "") + ".mp3"
	query := p.AzureAudio.PrepareQueryWithRandomVoice(s, true)
	if err := p.AzureAudio.Fetch(context.Background(), query, filename, 3); err != nil {
		slog.Error("fetching audio from azure", "error", err.Error())
	}
	return filename
}

// used for openai data that contains the translation and pinyin; currently we still use hsk and cedict only.
// TODO: add fallback with openai in case hsk and cedict don't know the word.
func (p *WordProcessor) Get(words []openai.Word, t *translate.Translations) []Word {
	var allWords []Word
	for _, word := range words {
		if word.Ch == "" {
			continue
		}
		if contains(p.IgnoreChars, word.Ch) {
			continue
		}

		example := ""
		isSingleRune := utf8.RuneCountInString(word.Ch) == 1
		if isSingleRune {
			example = strings.Join(p.WordIndex.GetExamplesForHanzi(word.Ch, 5), ", ")
		}

		cc, err := p.CardBuilder.GetWordCard(word.Ch, t)
		if err != nil {
			slog.Error("decompose", "word", word.Ch, "err", err)
			continue
		}

		w := Word{
			Chinese:      word.Ch,
			English:      word.En, // this comes from openai and is only used in the components of a sentence, which itself is translated by openai
			Cedict:       card.GetCedictEntries(cc),
			HSK:          card.GetHSKEntries(cc),
			Chars:        p.Chars.GetAll(word.Ch, false, t),
			IsSingleRune: isSingleRune,
			Components:   cc.Components,
			Traditional:  cc.TraditionalChinese,
			Example:      example,
			MnemonicBase: cc.MnemonicBase,
			Mnemonic:     cc.Mnemonic,
			Translation:  cc.Translation,
			Audio:        p.getAudio(word.Ch),
		}
		allWords = append(allWords, w)
	}
	return allWords
}

func (p *WordProcessor) getAudio(s string) string {
	filename := strings.ReplaceAll(s, " ", "") + ".mp3"
	if err := p.GCPAudio.Fetch(context.Background(), s, filename, false); err != nil {
		slog.Error("download GCP audio", "error", err, "word", s)
	}
	return filename
}

func (p *WordProcessor) Export(words []Word, outDir, deckname string, i ignore.Ignored) {
	p.ExportCards(deckname, words, i)
	p.ExportJSON(words, outDir)
}

func (p *WordProcessor) ExportJSON(words []Word, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "words.json")
	b, err := json.Marshal(words)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (p *WordProcessor) ExportCards(deckname string, words []Word, i ignore.Ignored) {
	for _, w := range words {
		if err := ExportWord(deckname, w, i); err != nil {
			fmt.Println(err)
		}
	}
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

// remove redundant
func removeRedundant(in []string) string {
	set := make(map[string]struct{})
	for _, elem := range in {
		set[elem] = struct{}{}
	}
	out := make([]string, 0)
	for elem := range set {
		out = append(out, elem)
	}
	return strings.Join(out, ", ")
}
