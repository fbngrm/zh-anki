package char

import (
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	"github.com/fbngrm/zh-anki/pkg/translate"
)

type Processor struct {
	IgnoreChars []string
	Audio       *audio.GCPClient
	WordIndex   *frequency.WordIndex
	CardBuilder *card.Builder
}

func (p *Processor) GetAll(word string, getAudio bool, t *translate.Translations) []Char {
	allChars := make([]Char, 0)
	for _, ch := range word {
		c := string(ch)

		example := ""
		isSingleRune := utf8.RuneCountInString(c) == 1
		if isSingleRune {
			example = removeRedundant(p.WordIndex.GetExamplesForHanzi(word, 5))
		}

		cc := p.CardBuilder.GetHanziCard(c, t)

		allChars = append(allChars, Char{
			Chinese:        cc.SimplifiedChinese,
			Cedict:         card.GetCedictEntries(cc),
			HSK:            card.GetHSKEntries(cc),
			IsSingleRune:   true,
			Components:     cc.Components,
			Traditional:    cc.TraditionalChinese,
			Example:        example,
			MnemonicBase:   cc.MnemonicBase,
			Mnemonic:       cc.Mnemonic,
			Pronounciation: cc.Pronounciation,
			Translation:    cc.Translation,
		})
	}
	if !getAudio {
		return allChars
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
		filename := char.Chinese + ".mp3"
		// currently we don't want audio for chars
		// if err := p.Audio.Fetch(context.Background(), char.Chinese, filename, false); err != nil {
		// 	fmt.Println(err)
		// }
		chars[y].Audio = filename
	}
	return chars
}
