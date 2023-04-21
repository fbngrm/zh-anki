package dialog

import "github.com/fbngrm/nprc/pkg/char"

type Word struct {
	Chinese      string      `yaml:"chinese"`
	Traditional  string      `yaml:"traditional"`
	Pinyin       string      `yaml:"pinyin"`
	English      string      `yaml:"english"`
	Audio        string      `yaml:"audio"`
	NewChars     []char.Char `yaml:"newChars"`
	AllChars     []char.Char `yaml:"allChars"`
	IsSingleRune bool        `yaml:"isSingleRune"`
}
