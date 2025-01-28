package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"golang.org/x/exp/slog"
)

type SentenceProcessor struct {
	Client *openai.Client
	Words  WordProcessor
	Audio  *audio.AzureClient
}

func (p *SentenceProcessor) DecomposeFromFile(path, outdir string, t *translate.Translations, dry bool) []Sentence {
	return p.Decompose(loadSentences(path), outdir, t, dry)
}

func (p *SentenceProcessor) Decompose(sentences []sentence, outdir string, t *translate.Translations, dry bool) []Sentence {
	var results []Sentence
	for _, sen := range sentences {
		slog.Info("=================================")
		slog.Info("decompose", "sentence", sen.text)

		s, err := p.Client.DecomposeSentence(sen.text)
		if err != nil {
			slog.Error("decompose sentence", "error", err.Error())
			continue
		}

		sentence := &Sentence{
			Chinese:      sen.text,
			English:      s.English,
			Pinyin:       s.Pinyin,
			Words:        p.Words.Get(s.Words, t),
			IsSingleRune: utf8.RuneCountInString(s.Chinese) == 1,
			UniqueChars:  getUniqueChars(s.Chinese),
			Grammar:      sen.grammar, // this only works when supplied in the sentences file
			Note:         sen.note,    // this only works when supplied in the sentences file
		}
		results = append(results, *sentence)

	}
	return p.getAudio(results, dry)
}

func (p *SentenceProcessor) Get(sentences []openai.Sentence, t *translate.Translations, dry bool) []Sentence {
	var results []Sentence
	for _, s := range sentences {
		results = append(results, Sentence{
			Chinese:      s.Chinese,
			English:      s.English,
			Pinyin:       s.Pinyin,
			Words:        p.Words.Get(s.Words, t),
			IsSingleRune: utf8.RuneCountInString(s.Chinese) == 1,
		})
	}
	return p.getAudio(results, dry)
}

func (p *SentenceProcessor) getAudio(sentences []Sentence, dry bool) []Sentence {
	for x, sentence := range sentences {
		filename := strings.ReplaceAll(sentence.Chinese, " ", "") + ".mp3"
		if !dry {
			query := p.Audio.PrepareQueryWithRandomVoice(sentence.Chinese, true)
			if err := p.Audio.Fetch(context.Background(), query, filename, 3); err != nil {
				slog.Error("fetching audio from azure", "error", err.Error())
			}
		}
		sentences[x].Audio = filename
	}
	return sentences
}

func (p *SentenceProcessor) Export(sentences []Sentence, outDir, deckname string, i ignore.Ignored) {
	p.ExportCards(deckname, sentences, i)
	p.ExportJSON(sentences, outDir)
}

func (p *SentenceProcessor) ExportJSON(sentences []Sentence, outDir string) {
	outDir = path.Join(outDir, "sentences")
	if err := os.MkdirAll(outDir, os.ModePerm); err != nil {
		fmt.Println("create sentence export dir: ", err.Error())
	}
	for _, s := range sentences {
		b, err := json.MarshalIndent(s, "", "    ")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		outPath := path.Join(outDir, s.Chinese+".json")
		if err := os.WriteFile(outPath, b, 0644); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}

func (p *SentenceProcessor) ExportCards(deckname string, sentences []Sentence, i ignore.Ignored) {
	for _, s := range sentences {
		if err := ExportSentence(deckname, s, i); err != nil {
			slog.Error("add note", "sentence", s.Chinese, "error", err)
		}
	}
}
