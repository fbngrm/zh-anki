package dialog

import (
	"encoding/json"
	"fmt"
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

func (g *SimpleGrammarProcessor) Decompose(path string, outdir string, i ignore.Ignored, t translate.Translations) {
	grammar, err := loadSimpleGrammar(path)
	if err != nil {
		log.Fatal(err)
	}
	g.ExportCards(grammar, outdir)
}

func (g *SimpleGrammarProcessor) Export(grammar SimpleGrammar, outDir string) {
	g.ExportCards(grammar, outDir)
	g.ExportJSON(grammar, outDir)
}

func (g *SimpleGrammarProcessor) ExportJSON(grammar SimpleGrammar, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "simple_grammar.json")
	b, err := json.Marshal(grammar)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (g *SimpleGrammarProcessor) ExportCards(grammar SimpleGrammar, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "cards.md")
	g.Exporter.CreateOrAppendAnkiCards(grammar, "simple_grammar.tmpl", outPath)
}
