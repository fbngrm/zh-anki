package char

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/decomposition"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh-mnemonics/mnemonic"
	"github.com/fbngrm/zh/lib/cedict"
)

type Processor struct {
	IgnoreChars     []string
	Cedict          map[string][]cedict.Entry
	Audio           audio.Downloader
	Decomposer      *decomposition.Decomposer
	WordIndex       *frequency.WordIndex
	MnemonicBuilder *mnemonic.Builder
}

func (p *Processor) GetAll(word string, t translate.Translations) []Char {
	allChars := make([]Char, 0)
	for _, ch := range word {
		c := string(ch)
		hanzi := p.Decomposer.Decompose(c)

		example := ""
		isSingleRune := utf8.RuneCountInString(c) == 1
		if isSingleRune {
			examples := p.WordIndex.GetExamplesForHanzi(c, 4)
			examples = append(examples, word)
			example = strings.Join(examples, ", ")
		}

		allChars = append(allChars, Char{
			Chinese:      c,
			English:      p.translateChar(c, t),
			Pinyin:       p.getPinyin(c),
			IsSingleRune: true,
			Components:   decomposition.GetComponents(hanzi),
			Kangxi:       decomposition.GetKangxi(hanzi),
			Equivalents:  removeRedundant(hanzi.Equivalents),
			Traditional:  removeRedundant(hanzi.IdeographsTraditional),
			Example:      example,
			MnemonicBase: p.getMnemonicBase(c),
			Mnemonic:     p.MnemonicBuilder.Lookup(c),
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

func (p *Processor) getReadings(ch string) []string {
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
	return pinyin
}

func (p *Processor) getPinyin(ch string) string {
	return strings.Join(p.getReadings(ch), ", ")
}

func (p *Processor) getMnemonicBase(ch string) string {
	mnemonicBase := ""
	for _, pinyin := range p.getReadings(ch) {
		m, err := p.MnemonicBuilder.GetBase(pinyin)
		if err != nil {
			fmt.Printf("could not get mnemonic base for word: %s", pinyin)
		}
		mnemonicBase = fmt.Sprintf("%s%s<br>%s<br>", mnemonicBase, pinyin, m)
	}
	return mnemonicBase
}

func (p *Processor) getAudio(chars []Char) []Char {
	for y, char := range chars {
		filename, err := p.Audio.Fetch(context.Background(), char.Chinese, hash.Sha1(char.Chinese), false)
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
