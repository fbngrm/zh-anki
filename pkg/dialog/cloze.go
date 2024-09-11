package dialog

type Cloze struct {
	SentenceFront string `yaml:"chinese"`
	SentenceBack  string `yaml:"chinese"`
	Pinyin        string `yaml:"pinyin"`
	English       string `yaml:"english"`
	Audio         string `yaml:"audio"`
	Words         []Word `yaml:"allWords"`
	Grammar       string `yaml:"grammar"`
	Note          string `yaml:"note"`
	Word          Word   `yaml:"word"`
}
