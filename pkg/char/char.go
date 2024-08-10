package char

import "github.com/fbngrm/zh-anki/pkg/card"

type CedictEntry struct {
	CedictPinyin  string `yaml:"cedict_pinyin"`
	CedictEnglish string `yaml:"cedict_en"`
}

type HSKEntry struct {
	HSKPinyin  string `yaml:"hsk_pinyin"`
	HSKEnglish string `yaml:"hsk_en"`
}

type Char struct {
	Chinese        string             `yaml:"chinese"`
	Cedict         []card.CedictEntry `yaml:"cedict"`
	HSK            []card.HSKEntry    `yaml:"hsk"`
	Traditional    string             `yaml:"traditional"`
	Audio          string             `yaml:"audio"`
	IsSingleRune   bool               `yaml:"isSingleRune"`
	Components     []card.Component   `yaml:"components"`
	Kangxi         []string           `yaml:"kangxi"`
	Equivalents    string             `yaml:"equivalents"`
	Example        string             `yaml:"example"`
	MnemonicBase   string             `yaml:"mnemonic_base"`
	Mnemonic       string             `yaml:"mnemonic"`
	Pronounciation string             `yaml:"pronounciation"`
	Translation    string             `yaml:"translation"` // this is coming from data/translations file
}
