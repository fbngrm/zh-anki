package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/dialog"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	ignore_dict "github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"github.com/fbngrm/zh-freq/pkg/card"
	"github.com/fbngrm/zh/lib/cedict"
)

const cedictSrc = "./pkg/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const wordFrequencySrc = "./pkg/frequency/global_wordfreq.release_UTF-8.txt"
const mnemonicsSrc = "/home/f/Dropbox/notes/chinese/mnemonics/mnemonics.txt"

var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。", "", " ", "、"}

var lesson string
var source string
var target string
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
	flag.BoolVar(&renderSentences, "s", true, "render sentences")
	flag.StringVar(&source, "src", "", "source folder name (and anki deck name if target is empty)")
	flag.StringVar(&target, "tgt", "", "anki target deck name (if empty, use source)")
	flag.StringVar(&tags, "t", "", "comma separated list of anki tags")
	flag.Parse()

	deckname = source
	if target != "" {
		deckname = target
	}
	path = source
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

	audioDir := filepath.Join(cwd, "data", source, lesson, "audio")
	audioDownloader := audio.Downloader{
		IgnoreChars: ignoreChars,
		AudioDir:    audioDir,
	}

	// we cache responses from openai api and google text-to-speech
	cacheDir := filepath.Join(cwd, "data", "cache")
	cache := openai.NewCache(cacheDir)

	wordIndex, err := frequency.NewWordIndex(wordFrequencySrc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	builder, err := card.NewBuilder(audioDir, mnemonicsSrc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	charProcessor := char.Processor{
		IgnoreChars: ignoreChars,
		Cedict:      cedictDict,
		WordIndex:   wordIndex,
		CardBuilder: builder,
	}
	wordProcessor := dialog.WordProcessor{
		Chars:       charProcessor,
		Audio:       audioDownloader,
		IgnoreChars: ignoreChars,
		WordIndex:   wordIndex,
		CardBuilder: builder,
	}
	sentenceProcessor := dialog.SentenceProcessor{
		Client: openai.NewClient(apiKey, cache),
		Words:  wordProcessor,
		Audio:  audioDownloader,
	}
	dialogProcessor := dialog.DialogProcessor{
		Client:    openai.NewClient(apiKey, cache),
		Sentences: sentenceProcessor,
		Audio:     audioDownloader,
	}
	simpleGrammarProcessor := dialog.SimpleGrammarProcessor{
		Sentences: sentenceProcessor,
	}

	outDir := filepath.Join(cwd, "data", source, lesson, "output")
	_ = os.Remove(outDir)

	// load grammar from file
	simpleGrammarPath := filepath.Join(cwd, "data", source, lesson, "input", "grammar")
	if _, err := os.Stat(simpleGrammarPath); err == nil {
		simpleGrammarProcessor.Decompose(simpleGrammarPath, outDir, deckname, ignored, translations)
	}
	// load sentences from file
	sentencePath := filepath.Join(cwd, "data", source, lesson, "input", "sentences")
	if _, err := os.Stat(sentencePath); err == nil {
		sentences := sentenceProcessor.DecomposeFromFile(sentencePath, outDir, ignored, translations)
		sentenceProcessor.ExportCards(deckname, sentences)
	}
	wordPath := filepath.Join(cwd, "data", source, lesson, "input", "words")
	if _, err := os.Stat(wordPath); err == nil {
		words := wordProcessor.Decompose(wordPath, outDir, ignored, translations)
		wordProcessor.ExportCards(deckname, words)
	}
	// load dialogues from file
	dialogPath := filepath.Join(cwd, "data", source, lesson, "input", "dialogues")
	if _, err := os.Stat(dialogPath); err == nil {
		dialogues := dialogProcessor.Decompose(dialogPath, outDir, deckname, ignored, translations)
		dialogProcessor.ExportCards(deckname, dialogues, renderSentences)
	}

	// write newly ignored words
	ignored.Write(ignorePath)
	// write translations(translations, translationsPath)
}
