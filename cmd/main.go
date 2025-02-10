package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

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

const mnemonicsSrc = "/home/f/Dropbox/zh/mnemonics/mnemonics.txt"

const segmenterCmd = "/home/f/work/src/github.com/fbngrm/stanford-segmenter/segment.sh"
const segmenterModel = "pku"

// here we store responses from openai, the tmp output dir will be copied here in the Make target
const openaiCacheDir = "/home/f/Dropbox/zh/cache/openai"

// here we store generated audio, the tmp output dir will be copied here in the Make target
const audioCacheDir = "/home/f/Dropbox/zh/cache/audio"

var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。", "", " ", "、"}

var deckname string
var dryrun bool

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

	flag.StringVar(&deckname, "src", "", "deckname folder name (and anki deck name if target is empty)")
	flag.BoolVar(&dryrun, "dryrun", false, "perform a dry run (no actual export, only JSON export)")
	flag.Parse()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	targetdeck := "chinese::" + deckname

	ignorePath := filepath.Join(cwd, "data", "ignore")
	ignored := ignore_dict.Load(ignorePath)

	translationsPath := filepath.Join(cwd, "data", "translations")
	translations, err := translate.New(translationsPath, ignoreChars)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// here we store generated audio files, that are then copied to anki media dir and the audio cache
	tmpAudioDir := filepath.Join(cwd, "data", deckname, "audio")
	audioCache := &audio.Cache{
		SrcDir: audioCacheDir,
		DstDir: tmpAudioDir,
	}
	azureClient := audio.NewAzureClient(azureApiKey, tmpAudioDir, ignoreChars, audioCache)
	gcpClient := &audio.GCPClient{
		Cache:       audioCache,
		IgnoreChars: ignoreChars,
		AudioDir:    tmpAudioDir,
	}

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

	// we cache responses from openai api
	openaiCache := openai.NewCache(openaiCacheDir)
	openAIClient, err := openai.NewClient(openAIApiKey, openaiCache, segmenter)
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

	tmpOutdir := filepath.Join(cwd, "data", deckname, "output")

	// load sentences from file
	sentencePath := filepath.Join(cwd, "data", deckname, "sentences")
	if _, err := os.Stat(sentencePath); err == nil {
		sentences := sentenceProcessor.DecomposeFromFile(sentencePath, tmpOutdir, translations, dryrun)
		if dryrun {
			sentenceProcessor.ExportJSON(sentences, tmpOutdir)
		} else {
			sentenceProcessor.Export(sentences, tmpOutdir, targetdeck, ignored)
		}
	}
	// load clozes from file
	clozePath := filepath.Join(cwd, "data", deckname, "clozes")
	if _, err := os.Stat(clozePath); err == nil {
		clozes, err := clozeProcessor.DecomposeFromFile(clozePath, tmpOutdir, translations, dryrun)
		if err != nil {
			slog.Error("decompose cloze", "error", err)
			os.Exit(1)
		}
		if dryrun {
			clozeProcessor.ExportJSON(clozes, tmpOutdir)
		} else {
			clozeProcessor.Export(clozes, tmpOutdir, targetdeck, ignored)
		}

	}
	wordPath := filepath.Join(cwd, "data", deckname, "words")
	if _, err := os.Stat(wordPath); err == nil {
		words := wordProcessor.DecomposeFromFile(wordPath, tmpOutdir, translations, dryrun)
		if dryrun {
			wordProcessor.ExportJSON(words, tmpOutdir)
		} else {
			wordProcessor.Export(words, tmpOutdir, targetdeck, ignored)
		}
	}
	// load grammar from file
	grammarPath := filepath.Join(cwd, "data", deckname, "grammar")
	if _, err := os.Stat(grammarPath); err == nil {
		grammar, err := grammarProcessor.DecomposeFromFile(grammarPath, tmpOutdir, deckname)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		grammarProcessor.Export(grammar, tmpOutdir, targetdeck)
	}
	// write newly ignored words
	ignored.Write(ignorePath)
}
