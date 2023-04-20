package sentence

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/fbngrm/nprc/pkg/anki"
	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/hash"
	"github.com/fbngrm/nprc/pkg/ignore"
	"github.com/fbngrm/nprc/pkg/openai"
	"github.com/fbngrm/nprc/pkg/pinyin"
	"github.com/fbngrm/nprc/pkg/translate"
	"github.com/fbngrm/nprc/pkg/word"
)

type Sentence struct {
	Chinese      string      `yaml:"chinese"`
	Pinyin       string      `yaml:"pinyin"`
	English      string      `yaml:"english"`
	Audio        string      `yaml:"audio"`
	NewWords     []word.Word `yaml:"newWords"`
	AllWords     []word.Word `yaml:"allWords"`
	IsSingleRune bool        `yaml:"isSingleRune"`
}

type Processor struct {
	Client   *openai.Client
	Words    word.Processor
	Audio    audio.Downloader
	Exporter anki.Exporter
}

func (p *Processor) Decompose(path, outdir, deckname string, i ignore.Ignored, pinyinDict pinyin.Dict, t translate.Translations) []Sentence {
	sentences := load(path)

	var results []Sentence
	for y, sentence := range sentences {
		fmt.Println("decompose sentence:")
		fmt.Println(sentence)

		s := p.Client.DecomposeSentence(sentence)

		allWords, newWords := p.Words.GetWords(s.Words, i, t)
		sentence := Sentence{
			Chinese:      s.Chinese,
			English:      s.English,
			Pinyin:       s.Pinyin,
			Audio:        hash.Sha1(s.Chinese),
			AllWords:     allWords,
			NewWords:     newWords,
			IsSingleRune: utf8.RuneCountInString(s.Chinese) == 1,
		}
		results = append(results, sentence)
		p.ExportSentence(sentence, outdir, fmt.Sprintf("sentence_%02d.yaml", y+1))
	}
	return p.getAudio(results)
}

func (p *Processor) Get(sentences []openai.Sentence, i ignore.Ignored, t translate.Translations) []Sentence {
	var results []Sentence
	for _, s := range sentences {
		allWords, newWords := p.Words.GetWords(s.Words, i, t)
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

func (p *Processor) getAudio(sentences []Sentence) []Sentence {
	for x, sentence := range sentences {
		filename, err := p.Audio.Fetch(context.Background(), sentence.Chinese, hash.Sha1(sentence.Chinese))
		if err != nil {
			fmt.Println(err)
		}
		sentences[x].Audio = filename
	}
	return sentences
}

func (p *Processor) ExportCards(sentences []Sentence, outDir string) {
	os.Mkdir(outDir, os.FileMode(0522))
	outPath := filepath.Join(outDir, "sentence_cards.md")
	_ = os.Remove(outPath)
	p.Exporter.CreateOrAppendAnkiCards(sentences, "sentences.tmpl", outPath)
}

func (p *Processor) ExportSentence(sentence Sentence, outdir, filename string) {
	os.Mkdir(outdir, os.FileMode(0522))
	outpath := filepath.Join(outdir, filename)
	_ = os.Remove(outpath)
	p.Exporter.WriteYAMLFile(sentence, outpath)
}
