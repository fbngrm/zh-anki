package dialog

import (
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/char"
)

type Word struct {
	Chinese      string             `json:"chinese"`
	English      string             `json:"english"`
	Cedict       []card.CedictEntry `json:"cedict"`
	HSK          []card.HSKEntry    `json:"hsk"`
	Traditional  string             `json:"traditional"`
	Audio        string             `json:"audio"`
	Chars        []char.Char        `json:"allChars"`
	IsSingleRune bool               `json:"isSingleRune"`
	Components   []card.Component   `json:"components"`
	Kangxi       []string           `json:"kangxi"`
	Equivalents  string             `json:"equivalents"`
	Example      string             `json:"example"`
	Examples     []card.Example     `json:"examples"`
	MnemonicBase string             `json:"mnemonic_base"`
	Mnemonic     string             `json:"mnemonic"`
	Note         string             `json:"note"`
	Translation  string             `json:"translation"` // this is coming from data/translations file
	Tones        []string           `json:"tones"`
}
