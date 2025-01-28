package dialog

import "github.com/fbngrm/zh-anki/pkg/card"

type Grammar struct {
	Cloze           string         `json:"cloze"`
	SentenceFront   string         `json:"sentenceFront"`
	SentenceBack    string         `json:"sentenceBack"`
	SentencePinyin  string         `json:"sentencePinyin"`
	SentenceEnglish string         `json:"sentenceEnglish"`
	SentenceAudio   string         `json:"sentenceAudio"`
	Pattern         string         `json:"pattern"`
	Note            string         `json:"note"`
	Structure       string         `json:"structure"`
	Examples        []card.Example `json:"examples"`
	Summary         []string       `json:"summary"`
}
