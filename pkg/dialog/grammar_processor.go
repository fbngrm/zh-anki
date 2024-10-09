package dialog

import (
	"context"
	"fmt"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"golang.org/x/exp/slog"
)

type GrammarProcessor struct {
	Words  WordProcessor
	Client *openai.Client
	Audio  *audio.AzureClient
}

func (g *GrammarProcessor) DecomposeFromFile(path string, outdir, deckname string) (Grammar, error) {
	grammar, err := loadGrammar(path)
	if err != nil {
		return Grammar{}, fmt.Errorf("parse grammars: %w", err)
	}
	slog.Info("=================================")
	slog.Info("decompose", "grammar", grammar.Cloze)

	s, err := g.Client.DecomposeSentence(grammar.SentenceBack)
	if err != nil {
		slog.Error("decompose grammar sentence", "error", err.Error())
	}

	var e []card.Example
	if grammar.Examples != "" {
		decompositon, err := g.Client.Decompose(grammar.Examples)
		if err != nil {
			return Grammar{}, err
		}

		examples := make([]card.Example, len(decompositon.Sentences))
		for i, s := range decompositon.Sentences {
			examples[i] = card.Example{
				Chinese: s.Chinese,
				English: s.English,
				Pinyin:  s.Pinyin,
				Audio:   g.getAudio(s.Chinese),
			}
		}
		e = examples
	} else {
		examples, err := g.Client.GetExamplesForPattern(grammar.Pattern)
		if err != nil {
			slog.Error("fetch example sentences", "word", grammar.Cloze, "err", err)
		}
		e = g.getExampleSentences(examples.Examples)
	}

	return Grammar{
		Cloze:           grammar.Cloze,
		SentenceFront:   grammar.SentenceFront,
		SentenceBack:    grammar.SentenceBack,
		SentencePinyin:  s.Pinyin,
		SentenceEnglish: s.English,
		SentenceAudio:   g.getAudio(grammar.SentenceBack),
		Pattern:         grammar.Pattern,
		Note:            grammar.Note,
		Structure:       grammar.Structure,
		Examples:        e,
		Summary:         grammar.Summary,
	}, nil
}

func (g *GrammarProcessor) getExampleSentences(examples []openai.Word) []card.Example {
	results := make([]card.Example, len(examples))
	for i, e := range examples {
		results[i] = card.Example{
			Chinese: e.Ch,
			Pinyin:  e.Pi,
			English: e.En,
			Audio:   g.getAudio(e.Ch),
		}
	}
	return results
}

func (g *GrammarProcessor) getAudio(s string) string {
	w := strings.ReplaceAll(s, " ", "")
	filename := hash.Sha1(w) + ".mp3"
	if err := g.Audio.Fetch(context.Background(), s, filename); err != nil {
		slog.Error("fetch example sentences audio", "sentence", s, "err", err)
	}
	return filename
}

func (g *GrammarProcessor) ExportCards(deckname string, gr Grammar) {
	if err := ExportGrammar(deckname, gr); err != nil {
		slog.Error("add note", "grammar", gr.Cloze, "error", err)
	}
}
