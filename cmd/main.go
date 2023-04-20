package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/fbngrm/nprc/pkg/anki"
	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/char"
	"github.com/fbngrm/nprc/pkg/dialog"
	ignore_dict "github.com/fbngrm/nprc/pkg/ignore"
	"github.com/fbngrm/nprc/pkg/openai"
	pinyin_dict "github.com/fbngrm/nprc/pkg/pinyin"
	"github.com/fbngrm/nprc/pkg/sentence"
	"github.com/fbngrm/nprc/pkg/template"
	"github.com/fbngrm/nprc/pkg/translate"
	"github.com/fbngrm/nprc/pkg/word"
	"github.com/fbngrm/zh/lib/cedict"
)

const idsSrc = "../../lib/cjkvi/ids.txt"
const cedictSrc = "/home/f/work/src/github.com/fbngrm/zh/lib/cedict/cedict_1_0_ts_utf-8_mdbg.txt"
const wordFrequencySrc = "../../lib/word_frequencies/global_wordfreq.release_UTF-8.txt"

var ignoreChars = []string{"!", "！", "？", "?", "，", ",", ".", "。", "", " "}

var meta = map[string]struct {
	deckname string
	tags     []string
	path     string
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

	pinyinPath := filepath.Join(cwd, "data", "pinyin")
	pinyin := pinyin_dict.Load(pinyinPath)

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

	charProcessor := char.Processor{
		IgnoreChars: ignoreChars,
		Cedict:      cedictDict,
		Audio:       audioDownloader,
	}
	wordProcessor := word.Processor{
		Cedict:      cedictDict,
		Chars:       charProcessor,
		Audio:       audioDownloader,
		IgnoreChars: ignoreChars,
	}
	sentenceProcessor := sentence.Processor{
		Client:   openai.NewClient(apiKey),
		Words:    wordProcessor,
		Audio:    audioDownloader,
		Exporter: ankiExporter,
	}
	dialogProcessor := dialog.Processor{
		Client:    openai.NewClient(apiKey),
		Sentences: sentenceProcessor,
		Audio:     audioDownloader,
		Exporter:  ankiExporter,
	}

	outDir := filepath.Join(cwd, "data", deck, lesson, "output")

	// load dialogues from file
	dialogPath := filepath.Join(cwd, "data", deck, lesson, "input", "dialogues")
	dialogues := dialogProcessor.Decompose(dialogPath, outDir, deckname, ignored, pinyin, translations)

	// load sentences from file
	sentencePath := filepath.Join(cwd, "data", deck, lesson, "input", "sentences")
	sentences := sentenceProcessor.Decompose(sentencePath, outDir, deckname, ignored, pinyin, translations)

	// write cards
	dialogProcessor.ExportCards(dialogues, outDir)
	sentenceProcessor.ExportCards(sentences, outDir)

	// write newly ignored words
	ignored.Write(ignorePath)
	// write translations(translations, translationsPath)
	pinyin.Write(pinyinPath)
}

func loadFile(path string) []string {
	file, err := os.Open(path)
	if err != nil {
		fmt.Printf("could not open vocab file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines

}
