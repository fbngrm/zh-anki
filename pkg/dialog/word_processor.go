package dialog

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"golang.org/x/exp/slog"
)

type WordProcessor struct {
	Chars       char.Processor
	Audio       *audio.GCPClient
	IgnoreChars []string
	Client      *openai.Client
	WordIndex   *frequency.WordIndex
	CardBuilder *card.Builder
}

func (p *WordProcessor) DecomposeFromFile(path, outdir string, i ignore.Ignored, t *translate.Translations) []Word {
	words := loadWords(path)

	var newWords []Word
	for _, word := range words {
		if word.Chinese == "" {
			continue
		}
		if contains(p.IgnoreChars, word.Chinese) {
			continue
		}

		w, err := p.Decompose(word, i, t)
		if err != nil {
			slog.Error("decompose", "word", word, "err", err)
			continue
		}

		newWords = append(newWords, *w)
	}
	return newWords
}

func (p *WordProcessor) Decompose(w Word, i ignore.Ignored, t *translate.Translations) (*Word, error) {
	if _, ok := i[strings.ReplaceAll(w.Chinese, " ", "")]; ok {
		return nil, errors.New("exists in ignore list")
	}

	allChars := p.Chars.GetAll(w.Chinese, t)

	example := ""
	isSingleRune := utf8.RuneCountInString(w.Chinese) == 1
	if isSingleRune {
		example = removeRedundant(p.WordIndex.GetExamplesForHanzi(w.Chinese, 5))
	}

	cc, err := p.CardBuilder.GetWordCard(w.Chinese, t)
	if err != nil {
		return nil, err
	}

	newWord := Word{
		Chinese:      w.Chinese,
		Cedict:       card.GetCedictEntries(cc),
		HSK:          card.GetHSKEntries(cc),
		Chars:        allChars,
		IsSingleRune: isSingleRune,
		Components:   cc.Components,
		Traditional:  cc.TraditionalChinese,
		Example:      example,
		MnemonicBase: cc.MnemonicBase,
		Mnemonic:     cc.Mnemonic,
		Note:         w.Note,
		Translation:  cc.Translation,
	}

	newWord = p.getAudio(newWord)
	return &newWord, nil
}

// used for openai data that contains the translation and pinyin; currently we still use hsk and cedict only.
// TODO: add fallback with openai in case hsk and cedict don't know the word.
func (p *WordProcessor) Get(words []openai.Word, i ignore.Ignored, t *translate.Translations) []Word {
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

		allWords = append(allWords, Word{
			Chinese:      word.Ch,
			English:      word.En, // this comes from openai and is only used in the components of a sentence, which itself is translated by openai
			Cedict:       card.GetCedictEntries(cc),
			HSK:          card.GetHSKEntries(cc),
			Chars:        p.Chars.GetAll(word.Ch, t),
			IsSingleRune: isSingleRune,
			Components:   cc.Components,
			Traditional:  cc.TraditionalChinese,
			Example:      example,
			MnemonicBase: cc.MnemonicBase,
			Mnemonic:     cc.Mnemonic,
			Translation:  cc.Translation,
		})
	}
	return allWords
}

func (p *WordProcessor) getAudio(word Word) Word {
	w := strings.ReplaceAll(word.Chinese, " ", "")
	filename := hash.Sha1(w) + ".mp3"
	text := ""
	for _, c := range word.Chinese {
		text += string(c)
		text += " "
	}
	if err := p.Audio.Fetch(context.Background(), text, filename, false); err != nil {
		fmt.Println(err)
	}
	word.Audio = filename
	return word
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
