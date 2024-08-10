package dialog

import (
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-freq/pkg/card"
)

type Word struct {
	Chinese      string             `yaml:"chinese"`
	English      string             `yaml:"english"`
	Cedict       []card.CedictEntry `yaml:"cedict"`
	HSK          []card.HSKEntry    `yaml:"hsk"`
	Traditional  string             `yaml:"traditional"`
	Audio        string             `yaml:"audio"`
	Chars        []char.Char        `yaml:"allChars"`
	IsSingleRune bool               `yaml:"isSingleRune"`
	Components   []card.Component   `yaml:"components"`
	Kangxi       []string           `yaml:"kangxi"`
	Equivalents  string             `yaml:"equivalents"`
	Example      string             `yaml:"example"`
	MnemonicBase string             `yaml:"mnemonic_base"`
	Mnemonic     string             `yaml:"mnemonic"`
	Note         string             `yaml:"note"`
}
