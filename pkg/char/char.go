package char

import (
	"context"
	"fmt"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/decomposition"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

type Char struct {
	Chinese      string   `yaml:"chinese"`
	Traditional  string   `yaml:"traditional"`
	Pinyin       string   `yaml:"pinyin"`
	English      string   `yaml:"english"`
	Audio        string   `yaml:"audio"`
	IsSingleRune bool     `yaml:"isSingleRune"`
	Components   []string `yaml:"components"`
	Kangxi       []string `yaml:"kangxi"`
	Equivalents  string   `yaml:"equivalents"`
	Example      string   `yaml:"example"`
}

type Processor struct {
	IgnoreChars []string
	Cedict      map[string][]cedict.Entry
	Audio       audio.Downloader
	Decomposer  *decomposition.Decomposer
}

func (p *Processor) GetAll(word string, t translate.Translations) []Char {
	allChars := make([]Char, 0)
	for _, ch := range word {
		hanzi := p.Decomposer.Decompose(string(ch))
		allChars = append(allChars, Char{
			Chinese:      string(ch),
			English:      p.translateChar(string(ch), t),
			Pinyin:       p.getPinyin(string(ch)),
			IsSingleRune: true,
			Components:   decomposition.GetComponents(hanzi),
			Kangxi:       decomposition.GetKangxi(hanzi),
			// FIXME: remove redundant
			Equivalents: strings.Join(hanzi.Equivalents, ", "),
			Traditional: strings.Join(hanzi.IdeographsTraditional, ", "),
			// FIXME: more most-frequent words
			Example: word,
		})
	}
	return p.getAudio(allChars)
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

func (p *Processor) getPinyin(ch string) string {
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

func (p *Processor) translateChar(ch string, t translate.Translations) string {
	translation, ok := t[ch]
	if !ok {
		definitions := []string{}
		for _, entry := range p.Cedict[ch] {
			definitions = append(definitions, entry.Definitions...)
		}
		translation = strings.Join(definitions, ", ")
	}
	t.Update(ch, translation)
	return translation
}
