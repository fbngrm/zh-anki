package dialog

type Dialog struct {
	Deck      string     `yaml:"deck"`
	Tags      []string   `yaml:"tags"`
	Chinese   string     `yaml:"chinese"`
	Pinyin    string     `yaml:"pinyin"`
	English   string     `yaml:"english"`
	Audio     string     `yaml:"audio"`
	Sentences []Sentence `yaml:"sentences"`
}
