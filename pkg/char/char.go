package char

import (
	"context"
	"fmt"
	"strings"

	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/hash"
	"github.com/fbngrm/nprc/pkg/ignore"
	"github.com/fbngrm/nprc/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

type Char struct {
	Chinese      string   `yaml:"chinese"`
	Pinyin       string   `yaml:"pinyin"`
	English      string   `yaml:"english"`
	Audio        string   `yaml:"audio"`
	IsSingleRune bool     `yaml:"isSingleRune"`
	Components   []string `yaml:"components"`
}

type Processor struct {
	IgnoreChars []string
	Cedict      map[string][]cedict.Entry
	Audio       audio.Downloader
}

func (p *Processor) GetAll(ch string, t translate.Translations) []Char {
	allChars := make([]Char, 0)
	for _, char := range ch {
		allChars = append(allChars, Char{
			Chinese:      string(char),
			English:      p.translateChar(string(char), t),
			Pinyin:       p.getPinyin(string(char)),
			IsSingleRune: true,
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

// func (p *Processor) getComponents(ch string) string {
// 	entries, _ := p.Cedict[string(ch)]
// 	readings := make(map[string]struct{})
// 	for _, entry := range entries {
// 		for _, reading := range entry.Readings {
// 			readings[reading] = struct{}{}
// 		}
// 	}
// 	pinyin := make([]string, 0)
// 	for reading := range readings {
// 		pinyin = append(pinyin, reading)
// 	}
// 	return strings.Join(pinyin, ", ")
// }

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
