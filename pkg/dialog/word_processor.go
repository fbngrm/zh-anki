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
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh-freq/pkg/card"
)

type WordProcessor struct {
	Chars       char.Processor
	Audio       *audio.AzureClient
	IgnoreChars []string
	Client      *openai.Client
	WordIndex   *frequency.WordIndex
	CardBuilder *card.Builder
}

func (p *WordProcessor) Decompose(path, outdir string, i ignore.Ignored, t translate.Translations) []Word {
	words := loadWords(path)

	var newWords []Word
	for _, word := range words {
		if word == "" {
			continue
		}

		if contains(p.IgnoreChars, word) {
			continue
		}
		if _, ok := i[word]; ok {
			fmt.Println("word exists in ignore list: ", word)
			continue
		}

		i.Update(word)
		allChars := p.Chars.GetAll(word, t)

		example := ""
		isSingleRune := utf8.RuneCountInString(word) == 1
		if isSingleRune {
			example = removeRedundant(p.WordIndex.GetExamplesForHanzi(word, 5))
		}

		cc := p.CardBuilder.GetWordCard(word)

		newWords = append(newWords, Word{
			Chinese:      word,
			Cedict:       card.GetCedictEntries(cc),
			HSK:          card.GetHSKEntries(cc),
			AllChars:     allChars,
			NewChars:     p.Chars.GetNew(i, allChars),
			IsSingleRune: isSingleRune,
			Components:   cc.Components,
			Traditional:  cc.TraditionalChinese,
			Example:      example,
			MnemonicBase: cc.MnemonicBase,
			Mnemonic:     cc.Mnemonic,
		})
	}
	return p.getAudio(newWords)
}

// used for openai data that contains the translation and pinyin; currently we still use hsk and cedict only.
// TODO: add fallback with openai in case hsk and cedict don't know the word.
func (p *WordProcessor) Get(words []openai.Word, i ignore.Ignored, t translate.Translations) ([]Word, []Word) {
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

		cc := p.CardBuilder.GetWordCard(word.Ch)

		allWords = append(allWords, Word{
			Chinese:      word.Ch,
			English:      word.En, // this comes from openai and is only used in the components of a sentence, which itself is translated by openai
			Cedict:       card.GetCedictEntries(cc),
			HSK:          card.GetHSKEntries(cc),
			AllChars:     p.Chars.GetAll(word.Ch, t),
			IsSingleRune: isSingleRune,
			Components:   cc.Components,
			Traditional:  cc.TraditionalChinese,
			Example:      example,
			MnemonicBase: cc.MnemonicBase,
			Mnemonic:     cc.Mnemonic,
		})
	}
	allWords = p.getAudio(allWords)

	var newWords []Word
	for _, word := range allWords {
		if _, ok := i[word.Chinese]; ok {
			fmt.Println("word exists in ignore list: ", word.Chinese)
			continue
		}
		i.Update(word.Chinese)

		// set new chars after word has been added to ignore list,
		// we want to add words first, then chars
		word.NewChars = p.Chars.GetNew(i, word.AllChars)
		newWords = append(newWords, word)
	}
	return allWords, newWords
}

func (p *WordProcessor) getAudio(words []Word) []Word {
	for y, word := range words {
		filename := hash.Sha1(strings.ReplaceAll(word.Chinese, " ", "")) + ".mp3"
		text := ""
		for _, c := range word.Chinese {
			text += string(c)
			text += " "
		}
		query := p.Audio.PrepareQueryWithRandomVoice(text, true)
		if err := p.Audio.Fetch(context.Background(), query, filename, false); err != nil {
			fmt.Println(err)
		}
		words[y].Audio = filename
	}
	return words
}

func (p *WordProcessor) Export(words []Word, outDir, deckname string) {
	p.ExportCards(deckname, words)
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

func (p *WordProcessor) ExportCards(deckname string, words []Word) {
	for _, w := range words {
		if err := ExportWord(deckname, w); err != nil {
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
