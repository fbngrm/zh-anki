package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/decomposition"
	"github.com/fbngrm/zh-anki/pkg/dialog"
	ignore_dict "github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/template"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

const cedictSrc = "/home/f/work/src/github.com/fbngrm/zh/lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"

var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。", "", " "}

var meta = map[string]struct {
	deckname string
	tags     []string
	path     string // relative to `zh-anki/data` dir
}{
	"npcr_01": {
		deckname: "new-practical-chinese-reader-01",
		tags:     []string{"hsk1"},
		path:     "npcr_01",
	},
	"sc_02": {
		deckname: "super-chinese-02",
		tags:     []string{"hsk2"},
		path:     "super-chinese_02",
	},
	"var": {
		deckname: "var",
		tags:     []string{"daily-life"},
		path:     "var",
	},
	"most-frequent": {
		deckname: "10k-most-frequent-words",
		tags:     []string{"10k-most-frequent-words"},
		path:     "most-frequent",
	},
}

var lesson string
var deck string
var tags []string
var deckname string
var path string

var cedictDict map[string][]cedict.Entry

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Environment variable OPENAI_API_KEY is not set")
	}

	flag.StringVar(&lesson, "l", "", "lesson name")
	flag.StringVar(&deck, "d", "", "dec name [npcr|sc]")
	flag.Parse()

	m, ok := meta[deck]
	if !ok {
		fmt.Println("need deck name as parameter [npcr|sc]")
		os.Exit(1)
	}
	deckname = m.deckname
	tags = append(m.tags, "lesson-"+lesson)
	tags = append(tags, deckname)
	path = m.path

	var err error
	cedictDict, err = cedict.BySimplifiedHanzi(cedictSrc)
	if err != nil {
		fmt.Printf("could not init cedict: %v\n", err)
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ignorePath := filepath.Join(cwd, "data", "ignore")
	ignored := ignore_dict.Load(ignorePath)

	translationsPath := filepath.Join(cwd, "data", "translations")
	translations := translate.Load(translationsPath)

	audioDownloader := audio.Downloader{
		IgnoreChars: ignoreChars,
		AudioDir:    filepath.Join(cwd, "data", deck, lesson, "audio"),
	}

	ankiExporter := anki.Exporter{
		Deckname: deckname,
		TmplProcessor: template.NewProcessor(
			deckname,
			filepath.Join(cwd, "tmpl"),
			tags,
		),
	}

	// we cache responses from openai api and google text-to-speech
	cacheDir := filepath.Join(cwd, "data", "cache")
	cache := openai.NewCache(cacheDir, &ankiExporter)
	decomposer := decomposition.NewDecomposer()

	charProcessor := char.Processor{
		IgnoreChars: ignoreChars,
		Cedict:      cedictDict,
		Audio:       audioDownloader,
		Decomposer:  decomposer,
	}
	wordProcessor := dialog.WordProcessor{
		Cedict:      cedictDict,
		Chars:       charProcessor,
		Audio:       audioDownloader,
		IgnoreChars: ignoreChars,
		Exporter:    ankiExporter,
		Decomposer:  decomposer,
	}
	sentenceProcessor := dialog.SentenceProcessor{
		Client:   openai.NewClient(apiKey, cache),
		Words:    wordProcessor,
		Audio:    audioDownloader,
		Exporter: ankiExporter,
	}
	dialogProcessor := dialog.DialogProcessor{
		Client:    openai.NewClient(apiKey, cache),
		Sentences: sentenceProcessor,
		Audio:     audioDownloader,
		Exporter:  ankiExporter,
	}

	outDir := filepath.Join(cwd, "data", deck, lesson, "output")
	_ = os.Remove(outDir)

	// load dialogues from file
	dialogPath := filepath.Join(cwd, "data", deck, lesson, "input", "dialogues")
	if _, err := os.Stat(dialogPath); err == nil {
		dialogues := dialogProcessor.Decompose(dialogPath, outDir, deckname, ignored, translations)
		dialogProcessor.ExportCards(dialogues, outDir)
	}

	// load sentences from file
	sentencePath := filepath.Join(cwd, "data", deck, lesson, "input", "sentences")
	if _, err := os.Stat(sentencePath); err == nil {
		sentences := sentenceProcessor.Decompose(sentencePath, outDir, deckname, ignored, translations)
		sentenceProcessor.ExportCards(sentences, outDir)
	}

	wordPath := filepath.Join(cwd, "data", deck, lesson, "input", "words")
	if _, err := os.Stat(wordPath); err == nil {
		words := wordProcessor.Decompose(wordPath, outDir, deckname, ignored, translations)
		wordProcessor.ExportCards(words, outDir)
	}

	// write newly ignored words
	ignored.Write(ignorePath)
	// write translations(translations, translationsPath)
}
