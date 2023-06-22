package dialog

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/translate"
)

type GrammarProcessor struct {
	Sentences SentenceProcessor
	Exporter  anki.Exporter
}

func (g *GrammarProcessor) Decompose(path string, outdir, deckname string, i ignore.Ignored, t translate.Translations) {

	grammars, err := loadGrammar(path)
	if err != nil {
		log.Fatal(err)
	}

	for _, grammar := range grammars {
		g.ExportCards(grammars, outdir)
		// we create cards for the example sentences
		sentences := g.Sentences.Decompose(grammar.Examples, outdir, deckname, i, t)
		g.Sentences.ExportCards(sentences, outdir)
	}
}

func (g *GrammarProcessor) ExportCards(grammars []Grammar, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "cards.md")
	p.Exporter.CreateOrAppendAnkiCards(grammars, "grammar.tmpl", outPath)
}
