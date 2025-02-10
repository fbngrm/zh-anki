package dialog

type Cloze struct {
	SentenceFront string `json:"cloze"`
	SentenceBack  string `json:"chinese"`
	FileName      string `json:"filename"`
	Pinyin        string `json:"pinyin"`
	English       string `json:"english"`
	Audio         string `json:"audio"`
	Words         []Word `json:"allWords"`
	Grammar       string `json:"grammar"`
	Note          string `json:"note"`
	Word          Word   `json:"word"`
}
