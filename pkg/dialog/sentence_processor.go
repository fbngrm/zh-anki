package dialog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
)

type SentenceProcessor struct {
	Client   *openai.Client
	Words    WordProcessor
	Audio    audio.Downloader
	Exporter anki.Exporter
}

func (p *SentenceProcessor) Decompose(path, outdir, deckname string, i ignore.Ignored, t translate.Translations) []Sentence {
	sentences := loadSentences(path)

	var results []Sentence
	for _, sentence := range sentences {
		sentence = strings.ReplaceAll(sentence, " ", "")
		fmt.Println("decompose sentence:")
		fmt.Println(sentence)

		s, err := p.Client.DecomposeSentence(sentence)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		allWords, newWords := p.Words.Get(s.Words, i, t)
		sentence := &Sentence{
			Chinese:      s.Chinese,
			English:      s.English,
			Pinyin:       s.Pinyin,
			Audio:        hash.Sha1(s.Chinese),
			AllWords:     allWords,
			NewWords:     newWords,
			IsSingleRune: utf8.RuneCountInString(s.Chinese) == 1,
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
			Audio:        hash.Sha1(s.Chinese),
			AllWords:     allWords,
			NewWords:     newWords,
			IsSingleRune: utf8.RuneCountInString(s.Chinese) == 1,
		})
	}
	return p.getAudio(results)
}

func (p *SentenceProcessor) getAudio(sentences []Sentence) []Sentence {
	for x, sentence := range sentences {
		filename, err := p.Audio.Fetch(context.Background(), sentence.Chinese, hash.Sha1(sentence.Chinese), true)
		if err != nil {
			fmt.Println(err)
		}
		sentences[x].Audio = filename
	}
	return sentences
}

func (p *SentenceProcessor) ExportCards(sentences []Sentence, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "cards.md")
	p.Exporter.CreateOrAppendAnkiCards(sentences, "sentences.tmpl", outPath)
}
