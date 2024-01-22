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

type GrammarProcessor struct {
	Sentences SentenceProcessor
	Exporter  anki.Exporter
}

func (g *GrammarProcessor) Decompose(path string, outdir string, i ignore.Ignored, t translate.Translations) {
	grammar, err := loadGrammarYAML(path)
	if err != nil {
		log.Fatal(err)
	}

	g.ExportCards(grammar, outdir)

	for _, section := range grammar.Sections {
		for _, structure := range section.Structures {
			// we create cards for the example sentences
			sentences := g.Sentences.Decompose(structure.Examples, outdir, i, t)
			g.Sentences.ExportCards(sentences, outdir)
		}
	}
}

func (g *GrammarProcessor) Export(grammar Grammar, outDir string) {
	g.ExportCards(grammar, outDir)
	g.ExportJSON(grammar, outDir)
}

func (g *GrammarProcessor) ExportJSON(grammar Grammar, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "grammar.json")
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

func (g *GrammarProcessor) ExportCards(grammar Grammar, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "cards.md")
	g.Exporter.CreateOrAppendAnkiCards(grammar, "grammar2.tmpl", outPath)
}
