package dialog

import (
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/translate"
)

type SimpleGrammarProcessor struct {
	Sentences SentenceProcessor
	Exporter  anki.Exporter
}

func (g *SimpleGrammarProcessor) Decompose(path string, outdir, deckname string, i ignore.Ignored, t translate.Translations) {
	grammar, err := loadSimpleGrammar(path)
	if err != nil {
		log.Fatal(err)
	}
	g.ExportCards(grammar, outdir)
}

func (g *SimpleGrammarProcessor) ExportCards(grammar SimpleGrammar, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "cards.md")
	g.Exporter.CreateOrAppendAnkiCards(grammar, "simple_grammar.tmpl", outPath)
}
