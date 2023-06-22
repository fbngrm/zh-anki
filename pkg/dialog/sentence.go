package dialog

type Sentence struct {
	Chinese      string   `yaml:"chinese"`
	Pinyin       string   `yaml:"pinyin"`
	English      string   `yaml:"english"`
	Audio        string   `yaml:"audio"`
	NewWords     []Word   `yaml:"newWords"`
	AllWords     []Word   `yaml:"allWords"`
	IsSingleRune bool     `yaml:"isSingleRune"`
	UniqueChars  []string `yaml:"unique_chars"`
}
