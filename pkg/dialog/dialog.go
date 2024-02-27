package dialog

type Dialog struct {
	Deck        string     `yaml:"deck"`
	Tags        []string   `yaml:"tags"`
	Chinese     string     `yaml:"chinese"`
	Pinyin      string     `yaml:"pinyin"`
	Audio       string     `yaml:"audio"`
	English     string     `yaml:"english"`
	Sentences   []Sentence `yaml:"sentences"`
	UniqueChars []string   `yaml:"unique_chars"`
}
