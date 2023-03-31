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
	"github.com/fbngrm/nprc/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

type Word struct {
	Chinese      string      `yaml:"chinese"`
	Pinyin       []string    `yaml:"pinyin"`
	English      string      `yaml:"english"`
	Audio        string      `yaml:"audio"`
	NewChars     []char.Char `yaml:"newChars"`
	AllChars     []char.Char `yaml:"allChars"`
	IsSingleRune bool        `yaml:"isSingleRune"`
}

type Processor struct {
	Cedict map[string][]cedict.Entry
	Chars  char.Processor
	Audio  audio.Downloader
}

func (p *Processor) Get(sentence string, i ignore.Ignored, t translate.Translations) ([]Word, []Word) {
	var allWords []Word
	for _, word := range cleanAndSplit(sentence) {
		if word == "" {
			continue
		}

		// FIXME: pinyin order needs to be preserved
		entries, _ := p.Cedict[word]
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

		allWords = append(allWords, Word{
			Chinese:      word,
			Pinyin:       pinyin,
			Audio:        hash.Sha1(word),
			AllChars:     p.Chars.GetAll(word, t),
			IsSingleRune: utf8.RuneCountInString(word) == 1,
		})
	}
	allWords = translateWords(allWords, t)
	allWords = p.getAudio(allWords)

	var newWords []Word
	for _, word := range allWords {
		if _, ok := i[word.Chinese]; ok {
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

		// for z, char := range word.AllChars {
		// 	translation, ok := t[char.Chinese]
		// 	if !ok {
		// 		var err error
		// 		translation, err = translateText("en-US", char.Chinese)
		// 		if err != nil {
		// 			log.Fatalf("could not translate char \"%s\": %v", char.Chinese, err)
		// 		}
		// 	}
		// 	char.English = translation
		// 	t.update(char.Chinese, char.English)
		// 	word.AllChars[z] = char
		// }
		// for z, char := range word.NewChars {
		// 	translation, ok := t[char.Chinese]
		// 	if !ok {
		// 		var err error
		// 		translation, err = translateText("en-US", char.Chinese)
		// 		if err != nil {
		// 			log.Fatalf("could not translate char \"%s\": %v", char.Chinese, err)
		// 		}
		// 	}
		// 	char.English = translation
		// 	t.update(char.Chinese, char.English)
		// 	word.NewChars[z] = char
		// }
		translated = append(translated, word)
	}
	return translated
}

func cleanAndSplit(sentence string) []string {
	sentence = strings.ReplaceAll(sentence, "。", " ")
	sentence = strings.ReplaceAll(sentence, ".", " ")
	sentence = strings.ReplaceAll(sentence, "，", " ")
	sentence = strings.ReplaceAll(sentence, ",", " ")
	sentence = strings.ReplaceAll(sentence, "?", " ")
	sentence = strings.ReplaceAll(sentence, "？", " ")
	sentence = strings.ReplaceAll(sentence, "！", " ")
	sentence = strings.ReplaceAll(sentence, "!", " ")

	var words []string
	if strings.Contains(sentence, " ") {
		words = strings.Split(sentence, " ")
	} else if strings.Contains(sentence, " ") {
		words = strings.Split(sentence, " ")
	} else {
		words = []string{sentence}
	}
	return words
}
