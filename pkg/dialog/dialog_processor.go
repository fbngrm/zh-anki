package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/hash"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
)

type DialogProcessor struct {
	Client    *openai.Client
	Sentences SentenceProcessor
	Audio     audio.Downloader
}

func (p *DialogProcessor) Decompose(path, outdir, deckname string, i ignore.Ignored, t translate.Translations) []*Dialog {
	// note, dialogues with a speaker must use `：` (unicode) to separate speaker and text.
	// this is not the same as `:` (ascii)!
	dialogues := loadDialogues(path)

	var results []*Dialog
	for _, dialog := range dialogues {
		decompositon := make([]openai.Sentence, 0)
		for _, sentence := range dialog.Lines {
			decomp, err := p.Client.DecomposeSentence(sentence.Text)
			if err != nil {
				fmt.Println(err.Error())
				continue
			}
			if decomp != nil {
				decompositon = append(decompositon, *decomp)
			}
		}

		pinyin := ""
		english := ""
		for _, s := range decompositon {
			pinyin += s.Pinyin
			pinyin += " "
			english += s.English
			english += " "
		}
		chinese := getChineseText(dialog.Lines, getColorsForSpeakers(dialog.Speakers))
		d := &Dialog{
			Deck: deckname,
			// this determines the audio filename. it is used in the template to set the audio file name.
			Chinese:     chinese,
			English:     english,
			Audio:       hash.Sha1(chinese),
			Pinyin:      pinyin,
			Sentences:   p.Sentences.Get(decompositon, i, t),
			UniqueChars: getUniqueChars(chinese),
		}
		results = append(results, d)

		// note, we support 4 different voices only!
		if len(dialog.Speakers) != 0 {
			if err := p.fetchDialogAudio(dialog, chinese); err != nil {
				fmt.Println(err)
			}
		} else {
			p.fetchAudio(dialog, chinese)
		}
	}
	return results
}

func (p *DialogProcessor) fetchAudio(d RawDialog, text string) string {
	filename, err := p.Audio.Fetch(context.Background(), d.Text, hash.Sha1(text), true)
	if err != nil {
		fmt.Println(err)
	}
	return filename
}

// if there are several speakers in a dialog, we fetch each line separately
// using a different voice for each speaker. we then merge them into a single file.
func (p *DialogProcessor) fetchDialogAudio(dialog RawDialog, text string) error {
	voices := p.Audio.GetVoices(dialog.Speakers)
	var paths []string
	for _, line := range dialog.Lines {
		voice, ok := voices[line.Speaker]
		if !ok {
			fmt.Printf("could not find voice for speaker: %s\n", line.Speaker)
		}
		fmt.Println("fetch line for dialog: ", line.Text)
		fmt.Println("use voice: ", voice.Name, voice.SsmlGender)
		path, err := p.Audio.FetchTmpAudioWithVoice(
			context.Background(),
			line.Text,
			hash.Sha1(line.Text),
			voice,
		)
		if err != nil {
			fmt.Println(err)
		}
		paths = append(paths, path)
	}

	return p.Audio.JoinAndSaveDialogAudio(hash.Sha1(text), paths)
}

func (p *DialogProcessor) Export(dialogues []*Dialog, renderSentences bool, outDir, deckname string) {
	p.ExportCards(deckname, dialogues, renderSentences)
	p.ExportJSON(dialogues, outDir)
}

func (p *DialogProcessor) ExportJSON(dialogues []*Dialog, outDir string) {
	os.Mkdir(outDir, os.ModePerm)
	outPath := filepath.Join(outDir, "dialogues.json")
	b, err := json.Marshal(dialogues)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if err := os.WriteFile(outPath, b, 0644); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func (p *DialogProcessor) ExportCards(deckname string, dialogues []*Dialog, renderSentences bool) {
	for _, d := range dialogues {
		if err := ExportDialog(renderSentences, d); err != nil {
			fmt.Println(err)
		}
	}
}

func getColorsForSpeakers(speakers map[string]struct{}) map[string]string {
	colorBySpeaker := make(map[string]string)
	colors := []string{
		"color1",
		"color2",
		"color3",
		"color4",
		"color5",
		"color6",
	}
	var i int
	for speaker := range speakers {
		colorBySpeaker[speaker] = colors[i%len(colors)]
		i++

	}
	return colorBySpeaker
}

// if color == true, wrap text in a span tag with color
func getChineseText(lines []DialogLine, colors map[string]string) string {
	var result string
	for _, line := range lines {
		if colors != nil {
			c := colors[line.Speaker]
			result += "<span class=\""
			result += c
			result += "\">"
			result += strings.ReplaceAll(line.Text, " ", "")
			result += "</span>"
		} else {
			result += "<span class=\"color6\">"
			result += strings.ReplaceAll(line.Text, " ", "")
			result += "</span>"
		}
		result += "<br>"
	}
	return result
}

func getUniqueChars(s string) []string {
	unique := make(map[string]struct{})
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			unique[string(r)] = struct{}{}
		}
	}
	var i int
	chars := make([]string, len(unique))
	for c := range unique {
		chars[i] = c
		i++
	}
	return chars
}
