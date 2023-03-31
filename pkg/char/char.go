package char

import (
	"context"
	"fmt"
	"log"

	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/hash"
	"github.com/fbngrm/nprc/pkg/ignore"
	"github.com/fbngrm/nprc/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

type Char struct {
	Chinese      string   `yaml:"chinese"`
	Pinyin       []string `yaml:"pinyin"`
	English      string   `yaml:"english"`
	Audio        string   `yaml:"audio"`
	IsSingleRune bool     `yaml:"isSingleRune"`
}

type Processor struct {
	Cedict map[string][]cedict.Entry
	Audio  audio.Downloader
}

func (p *Processor) GetAll(word string, t translate.Translations) []Char {
	allChars := make([]Char, 0)
	for _, char := range word {
		entries, _ := p.Cedict[string(char)]
		readings := make(map[string]struct{})
		for _, entry := range entries {
			for _, reading := range entry.Readings {
				readings[reading] = struct{}{}
			}
		}
		// FIXME: filter redundant readings
		pinyin := make([]string, 0)
		for reading := range readings {
			pinyin = append(pinyin, reading)
		}
		allChars = append(allChars, Char{
			Chinese:      string(char),
			Pinyin:       pinyin,
			IsSingleRune: true,
		})
	}
	return p.getAudio(translateChars(allChars, t))
}

func (p *Processor) GetNew(i ignore.Ignored, allChars []Char) []Char {
	newChars := make([]Char, 0)
	for _, char := range allChars {
		if _, ok := i[char.Chinese]; ok {
			continue
		}
		newChars = append(newChars, char)
		i.Update(char.Chinese)
	}
	return newChars
}

func (p *Processor) getAudio(chars []Char) []Char {
	for y, char := range chars {
		filename, err := p.Audio.Fetch(context.Background(), char.Chinese, hash.Sha1(char.Chinese))
		if err != nil {
			fmt.Println(err)
		}
		chars[y].Audio = filename
	}
	return chars
}

func translateChars(chars []Char, t translate.Translations) []Char {
	for i, char := range chars {
		chars[i] = translateChar(char, t)
	}
	return chars
}

func translateChar(char Char, t translate.Translations) Char {
	translation, ok := t[char.Chinese]
	if !ok {
		var err error
		translation, err = translate.Translate("en-US", char.Chinese)
		if err != nil {
			log.Fatalf("could not translate word \"%s\": %v", char.Chinese, err)
		}
	}
	char.English = translation
	t.Update(char.Chinese, char.English)
	return char
}
