package dialog

import (
	"context"
	"fmt"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"golang.org/x/exp/slog"
)

type GrammarProcessor struct {
	Client *openai.Client
	Audio  *audio.AzureClient
}

func (g *GrammarProcessor) Decompose(path string, outdir, deckname string) ([]Grammar, error) {
	grammars, err := loadGrammar(path)
	if err != nil {
		return nil, fmt.Errorf("parse grammars: %w", err)
	}

	results := []Grammar{}
	for _, grammar := range grammars {
		audioFilename := hash.Sha1(grammar.Head) + ".mp3"
		grammar.Audio = audioFilename

		query := g.Audio.PrepareQueryWithRandomVoice(grammar.Examples, false)
		if err := g.Audio.Fetch(context.Background(), query, audioFilename, true); err != nil {
			slog.Error("fetching audio from azure", "error", err.Error())
		}

		decompositon, err := g.Client.Decompose(grammar.Examples)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		examples := ""
		for _, s := range decompositon.Sentences {
			examples += s.Chinese
			examples += "\n"
			examples += s.Pinyin
			examples += "\n"
			examples += s.English
			examples += "\n\n"
		}
		grammar.Examples = examples

		results = append(results, grammar)
	}

	return results, nil
}

func (g *GrammarProcessor) ExportCards(deckname string, grammars []Grammar) {
	for _, g := range grammars {
		if err := ExportGrammar(deckname, g); err != nil {
			slog.Error("add note", "grammar", g.Head, "error", err)
		}
	}
}
