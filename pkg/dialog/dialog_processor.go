package dialog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fbngrm/zh-anki/pkg/anki"
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
	Exporter  anki.Exporter
}

func (p *DialogProcessor) Decompose(
	path, outdir, deckname string,
	i ignore.Ignored,
	t translate.Translations,
) []*Dialog {

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
		d := &Dialog{
			Deck:      deckname,
			Chinese:   dialog.Text,
			English:   english,
			Pinyin:    pinyin,
			Audio:     hash.Sha1(dialog.Text),
			Sentences: p.Sentences.Get(decompositon.Sentences, i, t),
		}
		p.fetchAudio(d)
		results = append(results, d)
	}
	return results
}

func (p *DialogProcessor) fetchAudio(d *Dialog) {
	ctx := context.Background()

	filename, err := p.Audio.Fetch(ctx, d.Chinese, hash.Sha1(d.Chinese), true)
	if err != nil {
		fmt.Println(err)
	}
	d.Audio = filename
}

func (p *DialogProcessor) ExportCards(dialogues []*Dialog, renderSentences bool, outdir string) {
	os.Mkdir(outdir, os.ModePerm)
	outpath := filepath.Join(outdir, "cards.md")
	data := map[string]interface{}{
		"Dialogues":       dialogues,
		"RenderSentences": renderSentences,
	}
	p.Exporter.CreateOrAppendAnkiCards(data, "dialog.tmpl", outpath)
}
