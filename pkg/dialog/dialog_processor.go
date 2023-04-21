package dialog

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/fbngrm/nprc/pkg/anki"
	"github.com/fbngrm/nprc/pkg/audio"
	"github.com/fbngrm/nprc/pkg/hash"
	"github.com/fbngrm/nprc/pkg/ignore"
	"github.com/fbngrm/nprc/pkg/openai"
	"github.com/fbngrm/nprc/pkg/pinyin"
	"github.com/fbngrm/nprc/pkg/translate"
)

type DialogProcessor struct {
	Client    *openai.Client
	Sentences SentenceProcessor
	Audio     audio.Downloader
	Exporter  anki.Exporter
	Cache     *Cache
}

func (p *DialogProcessor) Decompose(path, outdir, deckname string, i ignore.Ignored, pinyinDict pinyin.Dict, t translate.Translations) []*Dialog {
	dialogues := loadDialogues(path)

	var results []*Dialog
	for y, dialog := range dialogues {
		if d, ok := p.Cache.lookupDialog(dialog); ok {
			p.fetchAudio(d)
			results = append(results, d)
			continue
		}

		decompositon := p.Client.Decompose(dialog)
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
			Chinese:   dialog,
			English:   english,
			Pinyin:    pinyin,
			Audio:     hash.Sha1(dialog),
			Sentences: p.Sentences.Get(decompositon.Sentences, i, t),
		}
		p.ExportDialog(d, outdir, fmt.Sprintf("dialog_%02d.yaml", y+1))
		p.Cache.AddDialog(d)
		p.fetchAudio(d)
		results = append(results, d)
	}
	return results
}

func (p *DialogProcessor) fetchAudio(d *Dialog) {
	ctx := context.Background()

	filename, err := p.Audio.Fetch(ctx, d.Chinese, hash.Sha1(d.Chinese))
	if err != nil {
		fmt.Println(err)
	}
	d.Audio = filename
}

func (p *DialogProcessor) ExportCards(dialogues []*Dialog, outdir string) {
	os.Mkdir(outdir, os.FileMode(0522))
	outpath := filepath.Join(outdir, "cards.md")
	_ = os.Remove(outpath)
	p.Exporter.CreateOrAppendAnkiCards(dialogues, "dialog.tmpl", outpath)
}

func (p *DialogProcessor) ExportDialog(dialog *Dialog, outdir, filename string) {
	os.Mkdir(outdir, os.FileMode(0522))
	outpath := filepath.Join(outdir, filename)
	_ = os.Remove(outpath)
	p.Exporter.WriteYAMLFile(dialog, outpath)
}
