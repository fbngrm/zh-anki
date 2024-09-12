package dialog

import (
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/char"
)

type Word struct {
	Chinese       string             `yaml:"chinese"`
	English       string             `yaml:"english"`
	Cedict        []card.CedictEntry `yaml:"cedict"`
	HSK           []card.HSKEntry    `yaml:"hsk"`
	Traditional   string             `yaml:"traditional"`
	Audio         string             `yaml:"audio"`
	Chars         []char.Char        `yaml:"allChars"`
	IsSingleRune  bool               `yaml:"isSingleRune"`
	Components    []card.Component   `yaml:"components"`
	Kangxi        []string           `yaml:"kangxi"`
	Equivalents   string             `yaml:"equivalents"`
	Example       string             `yaml:"example"`
	ExamplesAudio string             `yaml:"examples_audio"`
	MnemonicBase  string             `yaml:"mnemonic_base"`
	Mnemonic      string             `yaml:"mnemonic"`
	Note          string             `yaml:"note"`
	Translation   string             `yaml:"translation"` // this is coming from data/translations file
}
