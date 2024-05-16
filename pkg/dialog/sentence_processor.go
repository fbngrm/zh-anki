package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
)

type SentenceProcessor struct {
	Client *openai.Client
	Words  WordProcessor
	Audio  *audio.Client
}

func (p *SentenceProcessor) DecomposeFromFile(path, outdir string, i ignore.Ignored, t translate.Translations) []Sentence {
	return p.Decompose(loadSentences(path), outdir, i, t)
}

func (p *SentenceProcessor) Decompose(sentences []sentence, outdir string, i ignore.Ignored, t translate.Translations) []Sentence {
	var results []Sentence
	for _, sen := range sentences {
		fmt.Printf("decompose sentence: %s\n", sen.text)

		s, err := p.Client.DecomposeSentence(sen.text)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		allWords, newWords := p.Words.Get(s.Words, i, t)
		sentence := &Sentence{
			Chinese:      sen.text,
			English:      s.English,
			Pinyin:       s.Pinyin,
			AllWords:     allWords,
			NewWords:     newWords,
			IsSingleRune: utf8.RuneCountInString(s.Chinese) == 1,
			UniqueChars:  getUniqueChars(s.Chinese),
			Grammar:      sen.grammar, // this only works when supplied in the sentences file
			Note:         sen.note,    // this only works when supplied in the sentences file
		}
		results = append(results, *sentence)

	}
	return p.getAudio(results)
}

func (p *SentenceProcessor) Get(sentences []openai.Sentence, i ignore.Ignored, t translate.Translations) []Sentence {
	var results []Sentence
	for _, s := range sentences {
		allWords, newWords := p.Words.Get(s.Words, i, t)
		results = append(results, Sentence{
			Chinese:      s.Chinese,
			English:      s.English,
			Pinyin:       s.Pinyin,
			AllWords:     allWords,
			NewWords:     newWords,
			IsSingleRune: utf8.RuneCountInString(s.Chinese) == 1,
		})
	}
	return p.getAudio(results)
}

func (p *SentenceProcessor) getAudio(sentences []Sentence) []Sentence {
	for x, sentence := range sentences {
		filename := hash.Sha1(strings.ReplaceAll(sentence.Chinese, " ", "")) + ".mp3"
		query := p.Audio.PrepareQueryWithRandomVoice(sentence.Chinese, true)
		if err := p.Audio.Fetch(context.Background(), query, filename, true); err != nil {
			fmt.Println(err)
		}
		sentences[x].Audio = filename
	}
	return sentences
}

func (p *SentenceProcessor) Export(sentences []Sentence, outDir, deckname string) {
	p.ExportCards(deckname, sentences)
	p.ExportJSON(sentences, outDir)
}

func (p *SentenceProcessor) ExportJSON(sentences []Sentence, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "sentences.json")
	b, err := json.Marshal(sentences)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (p *SentenceProcessor) ExportCards(deckname string, sentences []Sentence) {
	for _, s := range sentences {
		if err := ExportSentence(deckname, s); err != nil {
			fmt.Println(err)
		}
	}
}

func getAllChars(s string) []string {
	unique := make(map[string]struct{})
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			unique[string(r)] = struct{}{}
		}
	}
	var i int
	chars := make([]string, len(unique))
	for c := range unique {
		chars[i] = c
		i++
	}
	return chars
}
