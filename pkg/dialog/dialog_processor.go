package dialog

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/fbngrm/zh-anki/pkg/audio"
	"github.com/fbngrm/zh-anki/pkg/ignore"
	"github.com/fbngrm/zh-anki/pkg/openai"
	"github.com/fbngrm/zh-anki/pkg/translate"
	"golang.org/x/exp/slog"
)

type DialogProcessor struct {
	Client    *openai.Client
	Sentences SentenceProcessor
	Audio     *audio.AzureClient
}

func (p *DialogProcessor) Decompose(path, outdir, deckname string, i ignore.Ignored, t *translate.Translations) []*Dialog {
	// note, dialogues with a speaker must use `：` (unicode) to separate speaker and text.
	// this is not the same as `:` (ascii)!
	dialogues := loadDialogues(path)

	var results []*Dialog

	for _, dialog := range dialogues {
		decompositon, err := p.Client.Decompose(dialog.Text)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}

		pinyin := ""
		english := ""
		for _, s := range decompositon.Sentences {
			pinyin += s.Pinyin
			pinyin += " "
			english += s.English
			english += " "
		}

		chinese := getChineseText(dialog.Lines, getColorsForSpeakers(dialog.Speakers))
		audioFilename := chinese + ".mp3"
		d := &Dialog{
			Deck: deckname,
			// this determines the audio filename. it is used in the template to set the audio file name.
			Chinese:     chinese,
			English:     english,
			Audio:       audioFilename,
			Pinyin:      pinyin,
			Sentences:   p.Sentences.Get(decompositon.Sentences, i, t),
			UniqueChars: getUniqueChars(chinese),
			Note:        dialog.Note,
		}
		results = append(results, d)

		// if an audio file called dialog.mp3 already exists, we use instead of generating one
		if err := p.checkForOriginalAudio(audioFilename); err == nil {
			return results
		}

		query := ""
		if len(dialog.Speakers) != 0 {
			query = p.prepareQuery(dialog)
		} else {
			query = p.Audio.PrepareQueryWithRandomVoice(dialog.Text, false)
		}
		if err := p.Audio.Fetch(context.Background(), query, audioFilename); err != nil {
			slog.Error("fetching audio from azure", "error", err.Error())
		}
	}
	return results
}

func (p *DialogProcessor) checkForOriginalAudio(filename string) error {
	existingAudioPath := filepath.Join(p.Audio.AudioDir, "dialog.mp3")
	if _, err := os.Stat(existingAudioPath); err != nil {
		return err
	}
	dst := filepath.Join(p.Audio.AudioDir, filename)
	if err := copy(existingAudioPath, dst); err != nil {
		fmt.Printf("copy original audio: %v\n", err)
		return err
	}
	return nil
}

func (p *DialogProcessor) prepareQuery(dialog RawDialog) string {
	query := ""
	voices := p.Audio.GetVoices(dialog.Speakers)
	for _, line := range dialog.Lines {
		voice, ok := voices[line.Speaker]
		if !ok {
			fmt.Printf("could not find voice for speaker: %s\n", line.Speaker)
		}
		query += p.Audio.PrepareQuery(line.Text, voice, false)
	}
	return query
}

func (p *DialogProcessor) Export(dialogues []*Dialog, renderSentences bool, outDir, deckname string, i ignore.Ignored) {
	p.ExportCards(deckname, dialogues, renderSentences, i)
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

func (p *DialogProcessor) ExportCards(deckname string, dialogues []*Dialog, renderSentences bool, i ignore.Ignored) {
	for _, d := range dialogues {
		if err := ExportDialog(renderSentences, d, i); err != nil {
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

// Copy copies the contents of the file at srcpath to a regular file
// at dstpath. If the file named by dstpath already exists, it is
// truncated. The function does not copy the file mode, file
// permission bits, or file attributes.
func copy(src, dst string) (err error) {
	r, err := os.Open(src)
	if err != nil {
		return err
	}
	defer r.Close() // ignore error: file was opened read-only.

	w, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		// Report the error, if any, from Close, but do so
		// only if there isn't already an outgoing error.
		if c := w.Close(); err == nil {
			err = c
		}
	}()

	_, err = io.Copy(w, r)
	return err
}
