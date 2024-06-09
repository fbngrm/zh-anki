package char

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh-freq/pkg/card"
)

type Processor struct {
	IgnoreChars []string
	Audio       *audio.GCPClient
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
			example = removeRedundant(p.WordIndex.GetExamplesForHanzi(word, 5))
		}

		cc := p.CardBuilder.GetHanziCard(word, c)

		allChars = append(allChars, Char{
			Chinese:      cc.SimplifiedChinese,
			Cedict:       card.GetCedictEntries(cc),
			HSK:          card.GetHSKEntries(cc),
			IsSingleRune: true,
			Components:   cc.Components,
			Traditional:  cc.TraditionalChinese,
			Example:      example,
			MnemonicBase: cc.MnemonicBase,
			Mnemonic:     cc.Mnemonic,
		})
	}
	return p.getAudio(allChars)
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
