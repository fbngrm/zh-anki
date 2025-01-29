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
	"github.com/fbngrm/zh-anki/pkg/segment"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"golang.org/x/exp/slog"
)

const wordFrequencySrc = "./pkg/frequency/global_wordfreq.release_UTF-8.txt"
const mnemonicsSrc = "/home/f/Dropbox/notes/chinese/mnemonics/mnemonics.txt"
const segmenterCmd = "/home/f/work/src/github.com/fbngrm/stanford-segmenter/segment.sh"
const segmenterModel = "pku"

var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。", "", " ", "、"}

var lesson string
var source string
var target string
var tags string
var tagList []string
var deckname string
var dryrun bool

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
	flag.BoolVar(&dryrun, "dryrun", false, "perform a dry run (no actual export, only JSON export)")
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

	segmenter := &segment.Segmenter{
		Cmd:   segmenterCmd,
		Model: segmenterModel,
	}
	openAIClient, err := openai.NewClient(openAIApiKey, cache, segmenter)
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
	grammarProcessor := dialog.GrammarProcessor{
		Client: openAIClient,
		Audio:  azureClient,
	}

	outdir := filepath.Join(cwd, "data", source, lesson, "output")

	// load sentences from file
	sentencePath := filepath.Join(cwd, "data", source, lesson, "input", "sentences")
	if _, err := os.Stat(sentencePath); err == nil {
		sentences := sentenceProcessor.DecomposeFromFile(sentencePath, outdir, translations, dryrun)
		if dryrun {
			sentenceProcessor.ExportJSON(sentences, outdir)
		} else {
			sentenceProcessor.Export(sentences, outdir, deckname, ignored)
		}
	}
	// load clozes from file
	clozePath := filepath.Join(cwd, "data", source, lesson, "input", "clozes")
	if _, err := os.Stat(clozePath); err == nil {
		clozes, err := clozeProcessor.DecomposeFromFile(clozePath, outdir, translations, dryrun)
		if err != nil {
			slog.Error("decompose cloze", "error", err)
			os.Exit(1)
		}
		if dryrun {
			clozeProcessor.ExportJSON(clozes, outdir)
		} else {
			clozeProcessor.Export(clozes, outdir, deckname, ignored)
		}

	}
	wordPath := filepath.Join(cwd, "data", source, lesson, "input", "words")
	if _, err := os.Stat(wordPath); err == nil {
		words := wordProcessor.DecomposeFromFile(wordPath, outdir, translations, dryrun)
		if dryrun {
			wordProcessor.ExportJSON(words, outdir)
		} else {
			wordProcessor.Export(words, outdir, deckname, ignored)
		}
	}
	// load grammar from file
	grammarPath := filepath.Join(cwd, "data", source, lesson, "input", "grammar")
	if _, err := os.Stat(grammarPath); err == nil {
		grammar, err := grammarProcessor.DecomposeFromFile(grammarPath, outdir, deckname)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		grammarProcessor.Export(grammar, outdir, deckname)
	}
	// write newly ignored words
	ignored.Write(ignorePath)
}
