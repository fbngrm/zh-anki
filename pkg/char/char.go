package char

type Char struct {
	Chinese      string   `yaml:"chinese"`
	Traditional  string   `yaml:"traditional"`
	Pinyin       string   `yaml:"pinyin"`
	English      string   `yaml:"english"`
	Audio        string   `yaml:"audio"`
	IsSingleRune bool     `yaml:"isSingleRune"`
	Components   []string `yaml:"components"`
	Kangxi       []string `yaml:"kangxi"`
	Equivalents  string   `yaml:"equivalents"`
	Example      string   `yaml:"example"`
	MnemonicBase string   `yaml:"mnemonic_base"`
	Mnemonic     string   `yaml:"mnemonic"`
}
