package dialog

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/decomposition"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/pinyin"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

type WordProcessor struct {
	Cedict      map[string][]cedict.Entry
	Chars       char.Processor
	Audio       audio.Downloader
	IgnoreChars []string
	Client      *openai.Client
	Exporter    anki.Exporter
	Decomposer  *decomposition.Decomposer
}

func (p *WordProcessor) Decompose(path, outdir, deckname string, i ignore.Ignored, pinyinDict pinyin.Dict, t translate.Translations) []Word {
	words := loadWords(path)

	var newWords []Word
	for _, word := range words {
		if word == "" {
			continue
		}

		p.Decomposer.Decompose(word)

		if contains(p.IgnoreChars, word) {
			continue
		}
		if _, ok := i[word]; ok {
			fmt.Println("word exists: ", word)
			continue
		}

		i.Update(word)

		definitions := []string{}
		for _, entry := range p.Cedict[word] {
			definitions = append(definitions, entry.Definitions...)
		}

		allChars := p.Chars.GetAll(word, t)
		newWords = append(newWords, Word{
			Chinese:      word,
			Pinyin:       p.getPinyin(word),
			English:      strings.ReplaceAll(strings.Join(definitions, ", "), "&#39;", "'"),
			Audio:        hash.Sha1(word),
			AllChars:     allChars,
			NewChars:     p.Chars.GetNew(i, allChars),
			IsSingleRune: utf8.RuneCountInString(word) == 1,
		})
	}
	return p.getAudio(newWords)
}

func (p *WordProcessor) getPinyin(ch string) string {
	entries, _ := p.Cedict[string(ch)]
	readings := make(map[string]struct{})
	for _, entry := range entries {
		for _, reading := range entry.Readings {
			readings[reading] = struct{}{}
		}
	}
	pinyin := make([]string, 0)
	for reading := range readings {
		pinyin = append(pinyin, reading)
	}
	return strings.Join(pinyin, ", ")
}

func (p *WordProcessor) Get(words []openai.Word, i ignore.Ignored, t translate.Translations) ([]Word, []Word) {
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
			English:      strings.ReplaceAll(strings.Join(definitions, ", "), "&#39;", "'"),
			Audio:        hash.Sha1(word.Ch),
			AllChars:     p.Chars.GetAll(word.Ch, t),
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

func (p *WordProcessor) getAudio(words []Word) []Word {
	for y, word := range words {
		filename, err := p.Audio.Fetch(context.Background(), word.Chinese, hash.Sha1(word.Chinese))
		if err != nil {
			fmt.Println(err)
		}
		words[y].Audio = filename
	}
	return words
}

func (p *WordProcessor) ExportCards(words []Word, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "cards.md")
	p.Exporter.CreateOrAppendAnkiCards(words, "words.tmpl", outPath)
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
