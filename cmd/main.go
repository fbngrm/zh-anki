package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/card"
	"github.com/fbngrm/zh-anki/pkg/char"
	"github.com/fbngrm/zh-anki/pkg/dialog"
	"github.com/fbngrm/zh-anki/pkg/frequency"
	ignore_dict "github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"golang.org/x/exp/slog"
)

const wordFrequencySrc = "./pkg/frequency/global_wordfreq.release_UTF-8.txt"
const mnemonicsSrc = "/home/f/Dropbox/notes/chinese/mnemonics/mnemonics.txt"

var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。", "", " ", "、"}

var lesson string
var source string
var target string
var tags string
var tagList []string
var deckname string

// by default, skip rendering separate cards for all sentences in a dialog
var renderSentences bool

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	openAIApiKey := os.Getenv("OPENAI_API_KEY")
	if openAIApiKey == "" {
		log.Fatal("Environment variable OPENAI_API_KEY is not set")
	}
	azureApiKey := os.Getenv("SPEECH_KEY")
	if azureApiKey == "" {
		log.Fatal("Environment variable SPEECH_KEY is not set")
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
	if strings.Contains(tags, ",") {
		tagList = strings.Split(tags, ",")
	} else if len(tags) > 0 {
		tagList = append(tagList, tags)
	}
	tagList = append(tagList, deckname)

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	ignorePath := filepath.Join(cwd, "data", "ignore")
	ignored := ignore_dict.Load(ignorePath)

	translationsPath := filepath.Join(cwd, "data", "translations")
	translations, err := translate.New(translationsPath, ignoreChars)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	audioDir := filepath.Join(cwd, "data", source, lesson, "audio")
	azureClient := audio.NewAzureClient(azureApiKey, audioDir, ignoreChars)
	gcpClient := &audio.GCPClient{
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

	builder, err := card.NewBuilder(mnemonicsSrc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	openAIClient, err := openai.NewClient(openAIApiKey, cache)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	charProcessor := char.Processor{
		IgnoreChars: ignoreChars,
		Audio:       gcpClient,
		WordIndex:   wordIndex,
		CardBuilder: builder,
	}
	wordProcessor := dialog.WordProcessor{
		Chars:       charProcessor,
		GCPAudio:    gcpClient,
		AzureAudio:  azureClient,
		IgnoreChars: ignoreChars,
		WordIndex:   wordIndex,
		CardBuilder: builder,
		Client:      openAIClient,
	}
	sentenceProcessor := dialog.SentenceProcessor{
		Client: openAIClient,
		Words:  wordProcessor,
		Audio:  azureClient,
	}
	clozeProcessor := dialog.ClozeProcessor{
		Client: openAIClient,
		Words:  wordProcessor,
		Audio:  azureClient,
	}
	dialogProcessor := dialog.DialogProcessor{
		Client:    openAIClient,
		Sentences: sentenceProcessor,
		Audio:     azureClient,
	}
	simpleGrammarProcessor := dialog.SimpleGrammarProcessor{
		Sentences: sentenceProcessor,
	}
	grammarProcessor := dialog.GrammarProcessor{
		Client: openAIClient,
		Audio:  azureClient,
	}

	outDir := filepath.Join(cwd, "data", source, lesson, "output")

	// load grammar from file
	simpleGrammarPath := filepath.Join(cwd, "data", source, lesson, "input", "simple_grammar")
	if _, err := os.Stat(simpleGrammarPath); err == nil {
		simpleGrammarProcessor.Decompose(simpleGrammarPath, outDir, deckname, ignored, translations)
	}
	// load sentences from file
	sentencePath := filepath.Join(cwd, "data", source, lesson, "input", "sentences")
	if _, err := os.Stat(sentencePath); err == nil {
		sentences := sentenceProcessor.DecomposeFromFile(sentencePath, outDir, translations)
		sentenceProcessor.ExportCards(deckname, sentences, ignored)
	}
	// load clozes from file
	clozePath := filepath.Join(cwd, "data", source, lesson, "input", "clozes")
	if _, err := os.Stat(clozePath); err == nil {
		clozes, err := clozeProcessor.DecomposeFromFile(clozePath, outDir, translations)
		if err != nil {
			slog.Error("decompose cloze", "error", err)
			os.Exit(1)
		}
		clozeProcessor.Export(clozes, outDir, deckname, ignored)
	}
	wordPath := filepath.Join(cwd, "data", source, lesson, "input", "words")
	if _, err := os.Stat(wordPath); err == nil {
		words := wordProcessor.DecomposeFromFile(wordPath, outDir, translations)
		wordProcessor.ExportCards(deckname, words, ignored)
	}
	// load dialogues from file
	dialogPath := filepath.Join(cwd, "data", source, lesson, "input", "dialogues")
	if _, err := os.Stat(dialogPath); err == nil {
		dialogues := dialogProcessor.Decompose(dialogPath, outDir, deckname, translations)
		dialogProcessor.ExportCards(deckname, dialogues, renderSentences, ignored)
	}
	// load grammar from file
	grammarPath := filepath.Join(cwd, "data", source, lesson, "input", "grammar")
	if _, err := os.Stat(grammarPath); err == nil {
		grammar, err := grammarProcessor.DecomposeFromFile(grammarPath, outDir, deckname)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		grammarProcessor.ExportCards(deckname, grammar)
	}

	// write newly ignored words
	ignored.Write(ignorePath)
}
