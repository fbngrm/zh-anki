package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"golang.org/x/exp/slog"
)

type ClozeProcessor struct {
	Client *openai.Client
	Words  WordProcessor
	Audio  *audio.AzureClient
}

func (p *ClozeProcessor) DecomposeFromFile(path, outdir string, i ignore.Ignored, t *translate.Translations) ([]Cloze, error) {
	clozes, err := loadClozes(path)
	if err != nil {
		return nil, err
	}
	return p.Decompose(clozes, outdir, i, t), nil
}

func (p *ClozeProcessor) Decompose(clozes []cloze, outdir string, i ignore.Ignored, t *translate.Translations) []Cloze {
	var results []Cloze
	for _, cl := range clozes {
		slog.Info("=================================")
		slog.Info("decompose", "cloze", cl.withoutParenthesis)

		s, err := p.Client.DecomposeSentence(cl.withoutParenthesis)
		if err != nil {
			slog.Error("decompose cloze sentence", "error", err.Error())
			continue
		}

		w, err := p.Words.Decompose(Word{Chinese: cl.word}, i, t)
		if err != nil {
			slog.Error("decompose cloze word", "word", cl.word, "error", err.Error())
			continue
		}

		fmt.Println(w.Components)

		results = append(results, Cloze{
			SentenceFront: cl.withUnderscores,
			SentenceBack:  cl.withoutParenthesis,
			English:       s.English,
			Pinyin:        s.Pinyin,
			// Words:         p.Words.Get(s.Words, i, t),
			Grammar: cl.grammar, // this only works when supplied in the sentences file
			Note:    cl.note,    // this only works when supplied in the sentences file
			Word:    *w,
		})
	}
	return p.getAudio(results)
}

func (p *ClozeProcessor) getAudio(clozes []Cloze) []Cloze {
	for x, c := range clozes {
		filename := strings.ReplaceAll(c.SentenceBack, " ", "") + ".mp3"
		query := p.Audio.PrepareQueryWithRandomVoice(c.SentenceBack, true)
		if err := p.Audio.Fetch(context.Background(), query, filename); err != nil {
			slog.Error("fetching audio from azure", "error", err.Error())
		}
		clozes[x].Audio = filename
	}
	return clozes
}

func (p *ClozeProcessor) Export(clozes []Cloze, outDir, deckname string, i ignore.Ignored) {
	p.ExportCards(deckname, clozes, i)
	p.ExportJSON(clozes, outDir)
}

func (p *ClozeProcessor) ExportJSON(clozes []Cloze, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "clozes.json")
	b, err := json.Marshal(clozes)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (p *ClozeProcessor) ExportCards(deckname string, clozes []Cloze, i ignore.Ignored) {
	for _, c := range clozes {
		if err := ExportCloze(deckname, c, i); err != nil {
			slog.Error("add note", "cloze", c.SentenceBack, "error", err)
		}
	}
}
