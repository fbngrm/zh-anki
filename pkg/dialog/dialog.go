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
	"github.com/fbngrm/nprc/pkg/sentence"
	"github.com/fbngrm/nprc/pkg/translate"
)

type Dialog struct {
	Deck      string              `yaml:"deck"`
	Tags      []string            `yaml:"tags"`
	Chinese   string              `yaml:"chinese"`
	Pinyin    string              `yaml:"pinyin"`
	English   string              `yaml:"english"`
	Audio     string              `yaml:"audio"`
	Sentences []sentence.Sentence `yaml:"sentences"`
}

type Processor struct {
	Client    *openai.Client
	Sentences sentence.Processor
	Audio     audio.Downloader
	Exporter  anki.Exporter
}

func (p *Processor) Decompose(path, outdir, deckname string, i ignore.Ignored, pinyinDict pinyin.Dict, t translate.Translations) []*Dialog {
	dialogues := load(path)

	var results []*Dialog
	for y, dialog := range dialogues {
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
		p.fetchAudio(d)
		results = append(results, d)
	}
	return results
}

func (p *Processor) fetchAudio(d *Dialog) {
	ctx := context.Background()

	filename, err := p.Audio.Fetch(ctx, d.Chinese, hash.Sha1(d.Chinese))
	if err != nil {
		fmt.Println(err)
	}
	d.Audio = filename
}

func (p *Processor) ExportCards(dialogues []*Dialog, outdir string) {
	os.Mkdir(outdir, os.FileMode(0522))
	outpath := filepath.Join(outdir, "cards.md")
	_ = os.Remove(outpath)
	p.Exporter.CreateOrAppendAnkiCards(dialogues, "dialog.tmpl", outpath)
}

func (p *Processor) ExportDialog(dialog *Dialog, outdir, filename string) {
	os.Mkdir(outdir, os.FileMode(0522))
	outpath := filepath.Join(outdir, filename)
	_ = os.Remove(outpath)
	p.Exporter.WriteYAMLFile(dialog, outpath)
}
