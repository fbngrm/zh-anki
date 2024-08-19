package dialog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/translate"
)

type SimpleGrammarProcessor struct {
	Sentences SentenceProcessor
}

func (g *SimpleGrammarProcessor) Decompose(path, outdir, deckname string, i ignore.Ignored, t *translate.Translations) {
	grammar, err := loadSimpleGrammar(path)
	if err != nil {
		log.Fatal(err)
	}
	g.Export(grammar, outdir, deckname)
}

func (g *SimpleGrammarProcessor) Export(grammar SimpleGrammar, outDir, deckname string) {
	g.ExportCards(deckname, grammar)
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

func (g *SimpleGrammarProcessor) ExportCards(deckname string, grammar SimpleGrammar) {
	if err := ExportSimpleGrammar(deckname, grammar); err != nil {
		fmt.Printf("export simple grammar: %s\n", err.Error())
	}
}
