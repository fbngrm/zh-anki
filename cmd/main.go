package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/anki"
	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/decomposition"
	"github.com/fbngrm/zh-anki/pkg/dialog"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	ignore_dict "github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/template"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh/lib/cedict"
)

const cedictSrc = "/home/f/work/src/github.com/fbngrm/zh/lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const wordFrequencySrc = "./lib/global_wordfreq.release_UTF-8.txt"

var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。", "", " "}

var lesson string
var deck string
var tags string
var tagList []string
var deckname string
var path string

// by default, skip rendering separate cards for all sentences in a dialog
var renderSentences bool

var cedictDict map[string][]cedict.Entry

func main() {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Fatal("Environment variable OPENAI_API_KEY is not set")
	}

	flag.StringVar(&lesson, "l", "", "lesson name")
	flag.BoolVar(&renderSentences, "s", false, "render sentences")
	flag.StringVar(&deck, "d", "", "anki deck name")
	flag.StringVar(&tags, "t", "", "comma separated list of anki tags")
	flag.Parse()

	deckname = deck
	path = deck
	if strings.Contains(tags, ",") {
		tagList = strings.Split(tags, ",")
	} else if len(tags) > 0 {
		tagList = append(tagList, tags)
	}
	tagList = append(tagList, deckname)

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
			tagList,
		),
	}

	// we cache responses from openai api and google text-to-speech
	cacheDir := filepath.Join(cwd, "data", "cache")
	cache := openai.NewCache(cacheDir, &ankiExporter)

	decomposer := decomposition.NewDecomposer()

	wordIndex, err := frequency.NewWordIndex(wordFrequencySrc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	charProcessor := char.Processor{
		IgnoreChars: ignoreChars,
		Cedict:      cedictDict,
		Audio:       audioDownloader,
		Decomposer:  decomposer,
		WordIndex:   wordIndex,
	}
	wordProcessor := dialog.WordProcessor{
		Cedict:      cedictDict,
		Chars:       charProcessor,
		Audio:       audioDownloader,
		IgnoreChars: ignoreChars,
		Exporter:    ankiExporter,
		Decomposer:  decomposer,
		WordIndex:   wordIndex,
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
	grammarProcessor := dialog.GrammarProcessor{
		Sentences: sentenceProcessor,
		Exporter:  ankiExporter,
	}
	simpleGrammarProcessor := dialog.SimpleGrammarProcessor{
		Sentences: sentenceProcessor,
		Exporter:  ankiExporter,
	}

	outDir := filepath.Join(cwd, "data", deck, lesson, "output")
	_ = os.Remove(outDir)

	// load grammar from file
	grammarPath := filepath.Join(cwd, "data", deck, lesson, "input", "grammar.yaml")
	if _, err := os.Stat(grammarPath); err == nil {
		grammarProcessor.Decompose(grammarPath, outDir, deckname, ignored, translations)
	}
	simpleGrammarPath := filepath.Join(cwd, "data", deck, lesson, "input", "simple_grammar")
	if _, err := os.Stat(simpleGrammarPath); err == nil {
		simpleGrammarProcessor.Decompose(simpleGrammarPath, outDir, deckname, ignored, translations)
	}

	// load dialogues from file
	dialogPath := filepath.Join(cwd, "data", deck, lesson, "input", "dialogues")
	if _, err := os.Stat(dialogPath); err == nil {
		dialogues := dialogProcessor.Decompose(dialogPath, outDir, deckname, ignored, translations)
		dialogProcessor.ExportCards(dialogues, renderSentences, outDir)
	}

	// load sentences from file
	sentencePath := filepath.Join(cwd, "data", deck, lesson, "input", "sentences")
	if _, err := os.Stat(sentencePath); err == nil {
		sentences := sentenceProcessor.DecomposeFromFile(sentencePath, outDir, deckname, ignored, translations)
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
