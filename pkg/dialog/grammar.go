package dialog

import "github.com/fbngrm/zh-anki/pkg/card"

type Grammar struct {
	Cloze           string
	SentenceFront   string
	SentenceBack    string
	SentencePinyin  string
	SentenceEnglish string
	SentenceAudio   string
	Pattern         string
	Note            string
	Structure       string
	Examples        []card.Example
	Summary         []string
}
