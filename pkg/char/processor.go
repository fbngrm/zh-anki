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
	"github.com/fbngrm/zh-freq/pkg/card"
	"github.com/fbngrm/zh/lib/cedict"
)

type Processor struct {
	IgnoreChars []string
	Cedict      map[string][]cedict.Entry
	Audio       audio.Downloader
	Decomposer  *decomposition.Decomposer
	WordIndex   *frequency.WordIndex
	CardBuilder *card.Builder
}

func (p *Processor) GetAll(word string, t translate.Translations) []Char {
	allChars := make([]Char, 0)
	for _, ch := range word {
		c := string(ch)

		example := ""
		isSingleRune := utf8.RuneCountInString(c) == 1
		if isSingleRune {
			examples := p.WordIndex.GetExamplesForHanzi(c, 4)
			examples = append(examples, word)
			example = strings.Join(examples, ", ")
		}

		card := p.CardBuilder.GetHanziCard(word, c)

		hskEntries := make([]HSKEntry, 0)
		if hsk, ok := card.DictEntries["hsk"]; ok {
			for _, entry := range hsk {
				hskEntries = append(hskEntries, HSKEntry{
					HSKPinyin:  entry.Pinyin,
					HSKEnglish: p.translateChar(c, entry.English, t),
				})
			}
		}
		cedictEntries := make([]CedictEntry, 0)
		if cedict, ok := card.DictEntries["cedict"]; ok {
			for _, entry := range cedict {
				cedictEntries = append(cedictEntries, CedictEntry{
					CedictPinyin:  entry.Pinyin,
					CedictEnglish: p.translateChar(c, entry.English, t),
				})
			}
		}
		allChars = append(allChars, Char{
			Chinese:      card.SimplifiedChinese,
			Cedict:       cedictEntries,
			HSK:          hskEntries,
			IsSingleRune: true,
			Components:   card.Components,
			Traditional:  card.TraditionalChinese,
			Example:      example,
			MnemonicBase: card.MnemonicBase,
			Mnemonic:     card.Mnemonic,
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

func (p *Processor) getAudio(chars []Char) []Char {
	for y, char := range chars {
		filename := hash.Sha1(char.Chinese) + ".mp3"
		if err := p.Audio.Fetch(context.Background(), char.Chinese, filename, false); err != nil {
			fmt.Println(err)
		}
		chars[y].Audio = filename
	}
	return chars
}

func (p *Processor) translateChar(ch, en string, t translate.Translations) string {
	translation, ok := t[ch]
	if ok {
		return translation
	}
	return en
}
