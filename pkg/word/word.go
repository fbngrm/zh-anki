package word

import (
	"context"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/char"
	"github.com/fbngrm/nprc/pkg/hash"
	"github.com/fbngrm/nprc/pkg/ignore"
	"github.com/fbngrm/nprc/pkg/openai"
	"github.com/fbngrm/nprc/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

type Word struct {
	Chinese      string      `yaml:"chinese"`
	Traditional  string      `yaml:"traditional"`
	Pinyin       string      `yaml:"pinyin"`
	English      string      `yaml:"english"`
	Audio        string      `yaml:"audio"`
	NewChars     []char.Char `yaml:"newChars"`
	AllChars     []char.Char `yaml:"allChars"`
	IsSingleRune bool        `yaml:"isSingleRune"`
}

type Processor struct {
	Cedict      map[string][]cedict.Entry
	Chars       char.Processor
	Audio       audio.Downloader
	IgnoreChars []string
}

func (p *Processor) GetWords(words []openai.Word, i ignore.Ignored, t translate.Translations) ([]Word, []Word) {
	var allWords []Word
	for _, word := range words {
		if word.Ch == "" {
			continue
		}
		if contains(p.IgnoreChars, word.Ch) {
			continue
		}
		definitions := []string{word.En}
		for _, entry := range p.Cedict[word.Ch] {
			definitions = append(definitions, entry.Definitions...)
		}

		allWords = append(allWords, Word{
			Chinese:      word.Ch,
			Pinyin:       word.Pi,
			English:      strings.Join(definitions, ", "),
			Audio:        hash.Sha1(word.Ch),
			AllChars:     p.Chars.GetAll(word.Ch, word.En, t),
			IsSingleRune: utf8.RuneCountInString(word.Ch) == 1,
		})
	}
	allWords = p.getAudio(allWords)

	var newWords []Word
	for _, word := range allWords {
		if _, ok := i[word.Chinese]; ok {
			fmt.Println("word exists: ", word.Chinese)
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

func (p *Processor) getAudio(words []Word) []Word {
	for y, word := range words {
		filename, err := p.Audio.Fetch(context.Background(), word.Chinese, hash.Sha1(word.Chinese))
		if err != nil {
			fmt.Println(err)
		}
		words[y].Audio = filename
	}
	return words
}

func translateWords(words []Word, t translate.Translations) []Word {
	var translated []Word
	for _, word := range words {
		translation, ok := t[word.Chinese]
		if !ok {
			var err error
			translation, err = translate.Translate("en-US", word.Chinese)
			if err != nil {
				log.Fatalf("could not translate word \"%s\": %v", word.Chinese, err)
			}
		}
		word.English = translation
		t.Update(word.Chinese, word.English)

		translated = append(translated, word)
	}
	return translated
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
