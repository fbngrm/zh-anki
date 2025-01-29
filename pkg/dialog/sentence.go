package dialog

type Sentence struct {
	Chinese      string `yaml:"chinese"`
	Pinyin       string `yaml:"pinyin"`
	English      string `yaml:"english"`
	Audio        string `yaml:"audio"`
	Words        []Word `yaml:"allWords"`
	IsSingleRune bool   `yaml:"isSingleRune"`
	Grammar      string `yaml:"grammar"`
	Note         string `yaml:"note"`
}
